package main_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/stretchr/testify/assert"
	"github.com/Bioblaze/mud"
)

func TestRegisterEventHandler(t *testing.T) {
	dummyHandler := func(eventBody json.RawMessage) {
		// Dummy handler
	}

	eventName := "testEvent"
	registerEventHandler(eventName, dummyHandler)

	if _, ok := eventRegistry[eventName]; !ok {
		t.Errorf("Event handler not registered for event: %s", eventName)
	}
}

func TestAuthenticateConnection(t *testing.T) {
	claims := &JwtClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour).Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(jwtSecret)
	if err != nil {
		t.Fatalf("Failed to sign test token: %v", err)
	}

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to create listener: %v", err)
	}
	defer ln.Close()

	go func() {
		conn, err := ln.Accept()
		if err != nil {
			t.Fatalf("Failed to accept connection: %v", err)
		}
		defer conn.Close()

		buf := make([]byte, 4096)
		n, err := conn.Read(buf)
		if err != nil {
			t.Fatalf("Failed to read from connection: %v", err)
		}

		if !strings.EqualFold(strings.TrimSpace(string(buf[:n])), signedToken) {
			t.Errorf("Expected token not received: got %s, want %s", buf[:n], signedToken)
		}
	}()

	conn, err := net.Dial("tcp", ln.Addr().String())
	if err != nil {
		t.Fatalf("Failed to connect to listener: %v", err)
	}
	defer conn.Close()

	conn.Write([]byte(signedToken))
	time.Sleep(100 * time.Millisecond)

	err = authenticateConnection(conn)
	if err != nil {
		t.Errorf("Expected successful authentication, got error: %v", err)
	}
}

func TestHandlePacket(t *testing.T) {
	eventName := "testEvent2"
	eventTriggered := false

	registerEventHandler(eventName, func(eventBody json.RawMessage) {
		eventTriggered = true
	})

	packet := Packet{
		EventName: eventName,
		EventBody: json.RawMessage(`{"key": "value"}`),
	}

	handlePacket(packet)

	if !eventTriggered {
		t.Errorf("Expected event handler to be triggered, but it wasn't")
	}
}

func createMockConnection() (net.Conn, net.Conn) {
	client, server := net.Pipe()
	return client, server
}

func sendPacket(conn net.Conn, packet *Packet) {
	data, _ := json.Marshal(packet)
	conn.Write(data)
}

func readPacket(conn net.Conn, buf []byte) (*Packet, error) {
	n, err := conn.Read(buf)
	if err != nil {
		return nil, err
	}

	var packet Packet
	err = json.Unmarshal(buf[:n], &packet)
	if err != nil {
		return nil, err
	}

	return &packet, nil
}

func TestProcessConnection(t *testing.T) {
	eventName := "testEvent3"
	eventTriggered := false

	registerEventHandler(eventName, func(eventBody json.RawMessage) {
		eventTriggered = true
	})

	client, server := createMockConnection()
	defer client.Close()
	defer server.Close()

	packet := Packet{
		EventName: eventName,
		EventBody: json.RawMessage(`{"key": "value"}`),
	}
	sendPacket(server, &packet)

	go processConnection(client)

	time.Sleep(100 * time.Millisecond)

	if !eventTriggered {
		t.Errorf("Expected event handler to be triggered, but it wasn't")
	}
}

func TestHandleConnection(t *testing.T) {
	eventName := "testEvent4"
	eventTriggered := false

	registerEventHandler(eventName, func(eventBody json.RawMessage) {
		eventTriggered = true
	})

	claims := &JwtClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour).Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(jwtSecret)
	if err != nil {
		t.Fatalf("Failed to sign test token: %v", err)
	}

	client, server := createMockConnection()
	defer client.Close()
	defer server.Close()

	go func() {
		client.Write([]byte(signedToken))
		time.Sleep(100 * time.Millisecond)

		packet := Packet{
			EventName: eventName,
			EventBody: json.RawMessage(`{"key": "value"}`),
		}
		sendPacket(client, &packet)
	}()

	var wg sync.WaitGroup
	wg.Add(1)
	go HandleConnection(server, &wg)
	wg.Wait()

	if !eventTriggered {
		t.Errorf("Expected event handler to be triggered, but it wasn't")
	}
}

func TestAuthenticateConnectionInvalidToken(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to create listener: %v", err)
	}
	defer ln.Close()

	go func() {
		conn, err := ln.Accept()
		if err != nil {
			t.Fatalf("Failed to accept connection: %v", err)
		}
		defer conn.Close()

		conn.Write([]byte("invalid_token"))
	}()

	conn, err := net.Dial("tcp", ln.Addr().String())
	if err != nil {
		t.Fatalf("Failed to connect to listener: %v", err)
	}
	defer conn.Close()

	err = authenticateConnection(conn)
	if err == nil {
		t.Errorf("Expected authentication error, got nil")
	}
}

func TestAuthenticateConnectionExpiredToken(t *testing.T) {
	claims := &JwtClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(-time.Hour).Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(jwtSecret)
	if err != nil {
		t.Fatalf("Failed to sign test token: %v", err)
	}

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to create listener: %v", err)
	}
	defer ln.Close()

	go func() {
		conn, err := ln.Accept()
		if err != nil {
			t.Fatalf("Failed to accept connection: %v", err)
		}
		defer conn.Close()

		conn.Write([]byte(signedToken))
	}()

	conn, err := net.Dial("tcp", ln.Addr().String())
	if err != nil {
		t.Fatalf("Failed to connect to listener: %v", err)
	}
	defer conn.Close()

	err = authenticateConnection(conn)
	if err == nil {
		t.Errorf("Expected authentication error, got nil")
	}
}

func TestProcessConnectionInvalidPacket(t *testing.T) {
	client, server := createMockConnection()
	defer client.Close()
	defer server.Close()

	// Send an invalid packet
	client.Write([]byte("invalid_packet"))

	go processConnection(client)

	time.Sleep(100 * time.Millisecond)

	// Ensure the connection is closed
	if _, err := client.Read(make([]byte, 1)); err == nil {
		t.Errorf("Expected connection to be closed, but it wasn't")
	}
}

func TestProcessConnectionUnknownEvent(t *testing.T) {
	client, server := createMockConnection()
	defer client.Close()
	defer server.Close()

	// Send a packet with an unknown event name
	packet := Packet{
		EventName: "unknown_event",
		EventBody: json.RawMessage(`{"key": "value"}`),
	}
	sendPacket(server, &packet)

	go processConnection(client)

	time.Sleep(100 * time.Millisecond)

	// Ensure the connection is still open
	if _, err := client.Read(make([]byte, 1)); err != io.EOF {
		t.Errorf("Expected connection to be closed, but it wasn't")
	}
}

func TestProcessConnectionMultiplePackets(t *testing.T) {
	eventName1 := "testEvent1"
	eventName2 := "testEvent2"
	eventTriggered1 := false
	eventTriggered2 := false

	registerEventHandler(eventName1, func(eventBody json.RawMessage) {
		eventTriggered1 = true
	})

	registerEventHandler(eventName2, func(eventBody json.RawMessage) {
		eventTriggered2 = true
	})

	client, server := createMockConnection()
	defer client.Close()
	defer server.Close()

	packet1 := Packet{
		EventName: eventName1,
		EventBody: json.RawMessage(`{"key": "value1"}`),
	}

	packet2 := Packet{
		EventName: eventName2,
		EventBody: json.RawMessage(`{"key": "value2"}`),
	}

	sendPacket(server, &packet1)
	sendPacket(server, &packet2)

	go processConnection(client)

	time.Sleep(100 * time.Millisecond)

	if !eventTriggered1 {
		t.Errorf("Expected event1 handler to be triggered, but it wasn't")
	}

	if !eventTriggered2 {
		t.Errorf("Expected event2 handler to be triggered, but it wasn't")
	}
}

func TestServerIntegration(t *testing.T) {
	// Set environment variables for the test
	os.Setenv("SERVER_ADDRESS", "localhost:9090")

	// Start the server
	go main()

	// Allow the server to start
	time.Sleep(1 * time.Second)

	// Generate a JWT token for authentication
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  "1234567890",
		"name": "John Doe",
		"iat":  time.Now().Unix(),
	})

	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		t.Fatalf("Error signing JWT token: %v", err)
	}

	// Create a client connection
	conn, err := net.Dial("tcp", "localhost:9090")
	if err != nil {
		t.Fatalf("Failed to connect to the server: %v", err)
	}
	defer conn.Close()

	// Authenticate the client
	_, err = conn.Write([]byte(tokenString))
	if err != nil {
		t.Fatalf("Failed to send JWT token: %v", err)
	}

	// Send a packet
	testPacket := Packet{
		EventName: "event1",
		EventBody: json.RawMessage(`{"message": "Hello, World!"}`),
	}

	packetBytes, err := json.Marshal(testPacket)
	if err != nil {
		t.Fatalf("Failed to marshal packet: %v", err)
	}

	_, err = conn.Write(packetBytes)
	if err != nil {
		t.Fatalf("Failed to send packet: %v", err)
	}

	// Add some logic here to verify that the server correctly processes the packet
	// For example, you can add a channel in the event handler to receive the processed event data
	resultChannel := make(chan string, 1)

	registerEventHandler("event1", func(eventBody json.RawMessage) {
		// Handle event1
		var data map[string]string
		json.Unmarshal(eventBody, &data)
		resultChannel <- data["message"]
	})

	// Wait for the result
	select {
	case result := <-resultChannel:
		assert.Equal(t, "Hello, World!", result, "Unexpected result from event handler")
	case <-time.After(5 * time.Second):
		t.Fatal("Timeout waiting for event handler to process the packet")
	}
}