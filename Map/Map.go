package Map

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"

	"github.com/Bioblaze/mud/Player"
	"github.com/aquilax/go-perlin"
)

type TerrainType int

const (
	Forest TerrainType = iota
	Mountain
	Water
)

type Tile struct {
    x           int
    y           int
    terrainType TerrainType
    obstacle    ObstacleType
}

type ObstacleType int

const (
    NoObstacle ObstacleType = iota
    Tree
    Boulder
    Building
)


type Map struct {
	id           uuid.UUID
	name         string
	width        int
	height       int
	tiles        [][]Tile
	player       *Player.Player
	adjacentMaps map[uuid.UUID]*Map
}

func init() {
    rand.Seed(time.Now().UnixNano())
}


func NewMap(name string, width, height int) *Map {
    m := &Map{
        id:           uuid.New(),
        name:         name,
        width:        width,
        height:       height,
        tiles:        make([][]Tile, width),
        adjacentMaps: make(map[uuid.UUID]*Map),
    }

    // Generate Perlin noise values for each tile
    alpha := 2.
	beta := 2.
	n := 3
	seed := rand.Int63()
	noise := perlin.NewPerlin(alpha, beta, n, seed)


    for i := 0; i < width; i++ {
        m.tiles[i] = make([]Tile, height)
        for j := 0; j < height; j++ {
            // Calculate the terrain type based on the noise value
            terrainValue := noise.At(i, j)
            terrainType := getTerrainType(terrainValue)

            // Add obstacles based on the terrain type
            obstacle := getObstacleType(terrainType)

            m.tiles[i][j] = Tile{
                x:           i,
                y:           j,
                terrainType: terrainType,
                obstacle:    obstacle,
            }
        }
    }

    return m
}

func getTerrainType(value float64) TerrainType {
    if value < -0.2 {
        return Water
    } else if value < 0.4 {
        return Forest
    } else {
        return Mountain
    }
}

func getObstacleType(terrainType TerrainType) ObstacleType {
    switch terrainType {
    case Forest:
        
        random := rand.Intn(10)
        if random < 3 {
            return Tree
        }
    case Mountain:
        
        random := rand.Intn(10)
        if random < 3 {
            return Boulder
        }
    case Water:
        
        random := rand.Intn(10)
        if random < 1 {
            return Building
        }
    }

    return NoObstacle
}

func (m *Map) GenerateMaze() {
    // Initialize all tiles as unvisited
    visited := make([][]bool, m.width)
    for i := 0; i < m.width; i++ {
        visited[i] = make([]bool, m.height)
    }

    // Start the maze generation at a random tile
    
    startX := rand.Intn(m.width)
    startY := rand.Intn(m.height)
    m.visitTile(startX, startY, visited)

    // Keep track of the current path
    path := []Tile{{x: startX, y: startY}}

    // Keep track of the last visited tile that has unvisited neighbors
    var lastTile *Tile

    for len(path) > 0 {
        // Get the current tile
        currentTile := path[len(path)-1]

        // Get the unvisited neighbors
        neighbors := m.getUnvisitedNeighbors(currentTile.x, currentTile.y, visited)

        if len(neighbors) > 0 {
            // Choose a random unvisited neighbor
            
            nextTile := neighbors[rand.Intn(len(neighbors))]

            // Mark the next tile as visited
            m.visitTile(nextTile.x, nextTile.y, visited)

            // Remove the wall between the current tile and the next tile
            m.removeWall(currentTile, nextTile)

            // Add the next tile to the path
            path = append(path, nextTile)

            // Remember the current tile for backtracking
            lastTile = &currentTile
        } else {
            // Backtrack to the last tile that has unvisited neighbors
            path = path[:len(path)-1]
            lastTile = nil
        }
    }
}

func (m *Map) visitTile(x, y int, visited [][]bool) {
    visited[x][y] = true
    tile := &m.tiles[x][y]
    tile.obstacle = NoObstacle
}

func (m *Map) getUnvisitedNeighbors(x, y int, visited [][]bool) []Tile {
    neighbors := make([]Tile, 0)

    for i := x-1; i <= x+1; i++ {
        for j := y-1; j <= y+1; j++ {
            if i == x && j == y {
                continue
            }

            if i < 0 || i >= m.width || j < 0 || j >= m.height {
                continue
            }

            if visited[i][j] == true {
                continue
            }

            if i == x || j == y {
                // Only add neighbors that are adjacent
                neighbors = append(neighbors, m.tiles[i][j])
            }
        }
    }

    return neighbors
}



func getRandomTerrainType(obstacle ObstacleType) TerrainType {
    
    random := rand.Intn(3)

    switch random {
    case 0:
        if obstacle == Tree {
            return Forest
        }
        return Water
    case 1:
        if obstacle == Boulder {
            return Mountain
        }
        return Forest
    default:
        if obstacle == Building {
            return Forest
        }
        return Mountain
    }
}


func (m *Map) GetTile(x, y int) (Tile, error) {
	if x < 0 || x >= m.width || y < 0 || y >= m.height {
		return Tile{}, fmt.Errorf("coordinates (%d, %d) are out of bounds", x, y) 
	}

	return m.tiles[x][y], nil
}

func (m *Map) SetPlayer(player *Player.Player, x, y int) error {
	if x < 0 || x >= m.width || y < 0 || y >= m.height {
		return fmt.Errorf("coordinates (%d, %d) are out of bounds", x, y) 
	}

	if m.player != nil {
		return errors.New("player already exists on map")
	}

	m.player = player
	player.SetLocation(x, y, m.id)

	return nil
}

func (m *Map) RemovePlayer(player *Player.Player) {
	if m.player == player {
		m.player = nil
	}
}

func (m *Map) AddAdjacentMap(adjacentMap *Map) {
	m.adjacentMaps[adjacentMap.id] = adjacentMap
}

func (m *Map) RemoveAdjacentMap(adjacentMap *Map) {
	delete(m.adjacentMaps, adjacentMap.id)
}

func (t Tile) IsWalkable() bool {
    return t.terrainType != Water && t.obstacle == NoObstacle
}


func (m *Map) GetAdjacentTiles(x, y int) []Tile {
    adjacentTiles := make([]Tile, 0)

    // Check all 8 adjacent tiles
    for i := x-1; i <= x+1; i++ {
        for j := y-1; j <= y+1; j++ {
            if i == x && j == y {
                // Skip the tile itself
                continue
            }

            if i < 0 || i >= m.width || j < 0 || j >= m.height {
                // Skip tiles outside the map
                continue
            }

            adjacentTiles = append(adjacentTiles, m.tiles[i][j])
        }
    }

    return adjacentTiles
}

func (m *Map) GetWalkableAdjacentTiles(x, y int) []Tile {
    adjacentTiles := m.GetAdjacentTiles(x, y)
    walkableTiles := make([]Tile, 0)

    for _, tile := range adjacentTiles {
        if tile.IsWalkable() {
            walkableTiles = append(walkableTiles, tile)
        }
    }

    return walkableTiles
}

func (m *Map) MovePlayer(dx, dy int) error {
    // Get the player's current location
    x, y := m.player.GetLocation()

    // Calculate the new location
    newX, newY := x+dx, y+dy

    // Check if the new location is within bounds
    if newX < 0 || newX >= m.width || newY < 0 || newY >= m.height {
        return fmt.Errorf("coordinates (%d, %d) are out of bounds", x, y) 
    }

    // Get the tile at the new location
    newTile := m.tiles[newX][newY]

    // Check if the new tile is walkable
    if !newTile.IsWalkable() {
        return errors.New("cannot move to non-walkable tile")
    }

    // Update the player's location
    m.player.SetLocation(newX, newY, m.id)

    return nil
}

func (m *Map) GetID() uuid.UUID {
    return m.id
}

func (m *Map) GetName() string {
    return m.name
}

func (m *Map) GetWidth() int {
    return m.width
}

func (m *Map) GetHeight() int {
    return m.height
}

func (m *Map) GetAdjacentMaps() []*Map {
    maps := make([]*Map, 0, len(m.adjacentMaps))
    for _, m := range m.adjacentMaps {
        maps = append(maps, m)
    }
    return maps
}

func (m *Map) GetPlayerLocation() (int, int) {
    return m.player.GetLocation()
}

func (t Tile) String() string {
    switch t.terrainType {
    case Forest:
        return "forest"
    case Mountain:
        return "mountain"
    case Water:
        return "water"
    default:
        return "unknown"
    }
}

func (m *Map) String() string {
    var result string

    for i := 0; i < m.width; i++ {
        for j := 0; j < m.height; j++ {
            result += m.tiles[i][j].String() + " "
        }
        result += "\n"
    }

    return result
}

func generateRandomName() string {
    adjectives := []string{"green", "dark", "peaceful", "hidden", "sunny", "quiet", "ancient", "mysterious", "enchanted", "bloody", "stormy", "haunted"}
    nouns := []string{"forest", "mountains", "valley", "lake", "river", "cave", "temple", "ruins", "castle", "city", "jungle", "desert"}

    
    adjectiveIndex := rand.Intn(len(adjectives))
    nounIndex := rand.Intn(len(nouns))

    return adjectives[adjectiveIndex] + " " + nouns[nounIndex]
}

func (m *Map) GenerateRandomName() {
    m.name = generateRandomName()
}

func (m *Map) SetTileTerrainType(x, y int, terrainType TerrainType) error {
    if x < 0 || x >= m.width || y < 0 || y >= m.height {
        return fmt.Errorf("coordinates (%d, %d) are out of bounds", x, y) 
    }

    m.tiles[x][y].terrainType = terrainType
    return nil
}

func (m *Map) GetTilesOfType(terrainType TerrainType) []Tile {
    var tiles []Tile
    for i := 0; i < m.width; i++ {
        for j := 0; j < m.height; j++ {
            if m.tiles[i][j].terrainType == terrainType {
                tiles = append(tiles, m.tiles[i][j])
            }
        }
    }
    return tiles
}

func (t Tile) Description() string {
    switch t.terrainType {
    case Forest:
        return "a dense forest"
    case Mountain:
        return "a rugged mountain range"
    case Water:
        return "a body of water"
    default:
        return "an unknown terrain type"
    }
}

func (m *Map) removeWall(currentTile, nextTile Tile) {
    // Calculate the x and y coordinates of the wall between the two tiles
    wallX := currentTile.x + (nextTile.x - currentTile.x)/2
    wallY := currentTile.y + (nextTile.y - currentTile.y)/2

    // Remove the wall obstacle
    m.tiles[wallX][wallY].obstacle = NoObstacle
}