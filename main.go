package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
)

const (
	ConnTimeout         = 20 * time.Second
	KeepAlivePeriod     = 5 * time.Minute
	ServerAddress       = ":8080"
	MaxConnectionsPerSec = 5
	MaxPacketsPerSec     = 20
)

type Packet struct {
	EventName string          `json:"event_name"`
	EventBody json.RawMessage `json:"event_body"`
}

type JwtClaims struct {
	jwt.StandardClaims
}

type EventHandler func(eventBody json.RawMessage)

var jwtSecret = []byte("your_jwt_secret")
var eventRegistry = make(map[string]EventHandler)
var connectionLimiter = rate.NewLimiter(rate.Limit(MaxConnectionsPerSec), MaxConnectionsPerSec)
var packetLimiter = rate.NewLimiter(rate.Limit(MaxPacketsPerSec), MaxPacketsPerSec)

func registerEventHandler(eventName string, handler EventHandler) {
	eventRegistry[eventName] = handler
}

func HandleConnection(conn net.Conn, wg *sync.WaitGroup) {
	defer wg.Done()

	err := authenticateConnection(conn)
	if err != nil {
		logrus.Error("Authentication error: ", err)
		conn.Close()
		return
	}

	processConnection(conn)
}

func authenticateConnection(conn net.Conn) error {
	conn.SetDeadline(time.Now().Add(ConnTimeout))

	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err != nil {
		return err
	}

	token, err := jwt.ParseWithClaims(string(buf[:n]), &JwtClaims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil {
		return err
	}

	if !token.Valid {
		return fmt.Errorf("invalid JWT")
	}

	return nil
}

// Add a custom error type for packet validation errors
type PacketValidationError struct {
	msg string
}

func (e *PacketValidationError) Error() string {
	return e.msg
}

func validatePacket(packet *Packet) error {
	if packet.EventName == "" {
		return &PacketValidationError{msg: "Event name is missing"}
	}

	handler, ok := eventRegistry[packet.EventName]
	if !ok {
		return &PacketValidationError{msg: fmt.Sprintf("Unknown event: %s", packet.EventName)}
	}

	if packet.EventBody == nil {
		return &PacketValidationError{msg: "Event body is missing"}
	}

	// Add additional validation checks for specific events if needed
	// For example, you can unmarshal EventBody into a specific struct and check if required fields are set

	return nil
}

func processConnection(conn net.Conn) {
    defer conn.Close() // Ensure the connection is closed

    conn.SetDeadline(time.Time{})
    conn.SetKeepAlive(true)
    conn.SetKeepAlivePeriod(KeepAlivePeriod)

    buf := make([]byte, 4096)
    for {
        if !packetLimiter.Allow() {
            logrus.Warn("Packet rate limit exceeded")
            continue
        }

        n, err := conn.Read(buf)
        if err != nil {
            if err == io.EOF {
                logrus.Info("Client disconnected")
                return
            }
            logrus.Error("Error reading packet: ", err)
            return
        }

        var packet Packet
        err = json.Unmarshal(buf[:n], &packet)
        if err != nil {
            logrus.Error("Error parsing packet: ", err)
            return
        }

        err = validatePacket(&packet)
        if err != nil {
            logrus.Warn("Invalid packet: ", err)
            continue
        }

        go handlePacket(packet)
    }
}

func handlePacket(packet Packet) {
	handler, ok := eventRegistry[packet.EventName]
	if !ok {
		logrus.Warn("Unknown event: ", packet.EventName)
		return
	}

	handler(packet.EventBody)
}

// Read environment variables and provide default values
func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}

func getEnvInt(key string, fallback int) int {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	intValue, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return intValue
}

func main() {

		// Set variables from environment variables or use default values
		jwtSecret := []byte(getEnv("JWT_SECRET", "your_jwt_secret"))
		ServerAddress := getEnv("SERVER_ADDRESS", ":8080")
		ConnTimeout := time.Duration(getEnvInt("CONN_TIMEOUT", 20)) * time.Second
		KeepAlivePeriod := time.Duration(getEnvInt("KEEP_ALIVE_PERIOD", 5)) * time.Minute
		maxConnectionsPerSec := getEnvInt("MAX_CONNECTIONS_PER_SEC", 5)
		maxPacketsPerSec := getEnvInt("MAX_PACKETS_PER_SEC", 20)
	
		connectionLimiter := rate.NewLimiter(rate.Limit(maxConnectionsPerSec), maxConnectionsPerSec)
		packetLimiter := rate.NewLimiter(rate.Limit(maxPacketsPerSec), maxPacketsPerSec)
	

	registerEventHandler("event1", func(eventBody json.RawMessage) {
		// Handle event1
	})

	registerEventHandler("event2", func(eventBody json.RawMessage) {
		// Handle event2
	})

	ln, err := net.Listen("tcp", ServerAddress)
	if err != nil {
		logrus.Fatal(err)
	}

	logrus.Info("Server listening on ", ServerAddress)

	// Set up signal handling for graceful shutdown
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)
	done := make(chan struct{})

	var wg sync.WaitGroup

	go func() {
		<-shutdown
		logrus.Info("Shutting down server...")

		// Stop accepting new connections
		ln.Close()

		// Cancel ongoing connections
		wg.Wait()
		close(done)
	}()

	for {
		if !connectionLimiter.Allow() {
			logrus.Warn("Connection rate limit exceeded")
			time.Sleep(100 * time.Millisecond) // Sleep for a short duration before trying again
			continue
		}

		conn, err := ln.Accept()
		if err != nil {
			if opErr, ok := err.(*net.OpError); ok && opErr.Op == "accept" {
				logrus.Info("Server stopped accepting new connections")
				break
			}
			logrus.Error("Error accepting connection: ", err)
			continue
		}

		wg.Add(1)
		go HandleConnection(conn, &wg)
	}

	<-done
	logrus.Info("Server gracefully stopped")
}