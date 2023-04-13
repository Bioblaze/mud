package Map_test

import (
	"testing"

	"github.com/Bioblaze/mud/Player"
	"github.com/Bioblaze/mud/Map"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewMap(t *testing.T) {
	m := Map.NewMap("Test Map", 10, 10)
	assert.NotNil(t, m)
	assert.Equal(t, "Test Map", m.GetName())
	assert.Equal(t, 10, m.GetWidth())
	assert.Equal(t, 10, m.GetHeight())
}

func TestSetPlayer(t *testing.T) {
	m := Map.NewMap("Test Map", 10, 10)
	player := Player.NewPlayer(uuid.New(), "testplayer")

	err := m.SetPlayer(player, 5, 5)
	assert.NoError(t, err)
	x, y := m.GetPlayerLocation()
	assert.Equal(t, 5, x)
	assert.Equal(t, 5, y)
}

func TestRemovePlayer(t *testing.T) {
	m := Map.NewMap("Test Map", 10, 10)
	player := Player.NewPlayer(uuid.New(), "testplayer")

	err := m.SetPlayer(player, 5, 5)
	assert.NoError(t, err)

	m.RemovePlayer(player)
	assert.Nil(t, m.Player)
}

func TestMovePlayer(t *testing.T) {
	m := Map.NewMap("Test Map", 10, 10)
	player := Player.NewPlayer(uuid.New(), "testplayer")

	err := m.SetPlayer(player, 5, 5)
	assert.NoError(t, err)

	err = m.MovePlayer(1, 0)
	assert.NoError(t, err)

	x, y := m.GetPlayerLocation()
	assert.Equal(t, 6, x)
	assert.Equal(t, 5, y)
}

func TestOutOfBounds(t *testing.T) {
	m := Map.NewMap("Test Map", 10, 10)

	_, err := m.GetTile(15, 15)
	assert.Error(t, err)
}

func TestTileDescription(t *testing.T) {
	tile := Map.Tile{
		TerrainType: Map.Forest,
	}

	assert.Equal(t, "a dense forest", tile.Description())
}

func TestTileIsWalkable(t *testing.T) {
	tile := Map.Tile{
		TerrainType: Map.Forest,
		Obstacle:    Map.NoObstacle,
	}

	assert.True(t, tile.IsWalkable())
}

func TestTileNotWalkable(t *testing.T) {
	tile := Map.Tile{
		TerrainType: Map.Water,
		Obstacle:    Map.NoObstacle,
	}

	assert.False(t, tile.IsWalkable())
}
