package Player_test

import (
    "math"
    "testing"
	"github.com/Bioblaze/mud/Player"
)

func TestNewPlayer(t *testing.T) {
    name := "John"
    maxHP := 20
    regenRate := 2
    strength := 10
    constitution := 5

    player := NewPlayer(name, maxHP, regenRate, strength, constitution)

    if player.name != name {
        t.Errorf("Expected name to be %s, but got %s", name, player.name)
    }

    if player.maxHP != maxHP {
        t.Errorf("Expected maxHP to be %d, but got %d", maxHP, player.maxHP)
    }

    if player.regenRate != regenRate {
        t.Errorf("Expected regenRate to be %d, but got %d", regenRate, player.regenRate)
    }

    if player.strength != strength {
        t.Errorf("Expected strength to be %d, but got %d", strength, player.strength)
    }

    if player.constitution != constitution {
        t.Errorf("Expected constitution to be %d, but got %d", constitution, player.constitution)
    }
}

func TestSetName(t *testing.T) {
    player := NewPlayer("John", 20, 2, 10, 5)

    name := "Jane"
    player.SetName(name)

    if player.name != name {
        t.Errorf("Expected name to be %s, but got %s", name, player.name)
    }
}

func TestSetLocation(t *testing.T) {
    player := NewPlayer("John", 20, 2, 10, 5)

    x := 10
    y := 20
    mapId := 30
    player.SetLocation(x, y, mapId)

    newX, newY, newMapId := player.GetLocation()

    if newX != x {
        t.Errorf("Expected x to be %d, but got %d", x, newX)
    }

    if newY != y {
        t.Errorf("Expected y to be %d, but got %d", y, newY)
    }

    if newMapId != mapId {
        t.Errorf("Expected mapId to be %d, but got %d", mapId, newMapId)
    }
}

func TestSetLevel(t *testing.T) {
    player := NewPlayer("John", 20, 2, 10, 5)

    level := 2
    player.SetLevel(level)

    if player.GetLevel() != level {
        t.Errorf("Expected level to be %d, but got %d", level, player.GetLevel())
    }
}

func TestAddExp(t *testing.T) {
    player := NewPlayer("John", 20, 2, 10, 5)

    exp := 100
    player.AddExp(exp)

    if player.GetExp() != exp {
        t.Errorf("Expected exp to be %d, but got %d", exp, player.GetExp())
    }

    level := 2
    player.AddExp(int(float64(level)*player.expCurve*100))

    if player.GetLevel() != level {
        t.Errorf("Expected level to be %d, but got %d", level, player.GetLevel())
    }
}

func TestAddExp_LevelUp(t *testing.T) {
    player := NewPlayer("John", 20, 2, 10, 5)

    exp := 100
    player.AddExp(exp)

    if player.GetExp() != exp {
        t.Errorf("Expected exp to be %d, but got %d", exp, player.GetExp())
    }

    level := 2
    player.AddExp(int(float64(level)*player.expCurve*100))

    if player.GetLevel() != level {
        t.Errorf("Expected level to be %d, but got %d", level, player.GetLevel())
    }

    // Gain enough experience points to level up
    expToLevelUp := int(float64(player.GetLevel())*player.expCurve*100) - player.GetExp()
    player.AddExp(expToLevelUp)

    if player.GetLevel() != level+1 {
        t.Errorf("Expected level to be %d, but got %d", level+1, player.GetLevel())
    }
}

func TestSetStats_LevelUp(t *testing.T) {
    player := NewPlayer("John", 20, 2, 10, 5)

    level := 2
    player.SetLevel(level)

    if player.GetLevel() != level {
        t.Errorf("Expected level to be %d, but got %d", level, player.GetLevel())
    }

    player.SetStats()

    expectedStrength := 12
    expectedConstitution := 6

    if player.strength != expectedStrength {
        t.Errorf("Expected strength to be %d, but got %d", expectedStrength, player.strength)
    }

    if player.constitution != expectedConstitution {
        t.Errorf("Expected constitution to be %d, but got %d", expectedConstitution, player.constitution)
    }
}

func TestDamage(t *testing.T) {
    player := NewPlayer("John", 20, 2, 10, 5)

    damage := 5
    player.Damage(damage)

    expectedHP := 15

    if player.GetCurrentHP() != expectedHP {
        t.Errorf("Expected current HP to be %d, but got %d", expectedHP, player.GetCurrentHP())
    }
}

func TestAddExp_NegativeExp(t *testing.T) {
    player := NewPlayer("John", 20, 2, 10, 5)

    initialExp := player.GetExp()
    initialLevel := player.GetLevel()

    player.AddExp(-100)

    if player.GetExp() != initialExp {
        t.Errorf("Expected exp to be %d, but got %d", initialExp, player.GetExp())
    }

    if player.GetLevel() != initialLevel {
        t.Errorf("Expected level to be %d, but got %d", initialLevel, player.GetLevel())
    }
}

func TestSetStats_InvalidLevel(t *testing.T) {
    player := NewPlayer("John", 20, 2, 10, 5)

    player.SetLevel(0)

    initialStrength := player.strength
    initialConstitution := player.constitution

    player.SetStats()

    if player.strength != initialStrength {
        t.Errorf("Expected strength to be %d, but got %d", initialStrength, player.strength)
    }

    if player.constitution != initialConstitution {
        t.Errorf("Expected constitution to be %d, but got %d", initialConstitution, player.constitution)
    }
}

func TestDamage_NegativeDamage(t *testing.T) {
    player := NewPlayer("John", 20, 2, 10, 5)

    initialHP := player.GetCurrentHP()

    player.Damage(-5)

    if player.GetCurrentHP() != initialHP {
        t.Errorf("Expected current HP to be %d, but got %d", initialHP, player.GetCurrentHP())
    }
}

func TestSetLocation_NegativeCoords(t *testing.T) {
    player := NewPlayer("John", 20, 2, 10, 5)

    initialX, initialY, initialMapId := player.GetLocation()

    player.SetLocation(-10, -20, 30)

    newX, newY, newMapId := player.GetLocation()

    if newX != initialX {
        t.Errorf("Expected x to be %d, but got %d", initialX, newX)
    }

    if newY != initialY {
        t.Errorf("Expected y to be %d, but got %d", initialY, newY)
    }

    if newMapId != initialMapId {
        t.Errorf("Expected mapId to be %d, but got %d", initialMapId, newMapId)
    }
}

func TestNewPlayer_ZeroMaxHp(t *testing.T) {
    name := "John"
    maxHP := 0
    regenRate := 2
    strength := 10
    constitution := 5

    player := NewPlayer(name, maxHP, regenRate, strength, constitution)

    if player.maxHP != 1 {
        t.Errorf("Expected maxHP to be 1, but got %d", player.maxHP)
    }
}

func TestSetLocation_NegativeCoord(t *testing.T) {
    player := NewPlayer("John", 20, 2, 10, 5)

    x := -1
    y := -1
    mapId := 30
    player.SetLocation(x, y, mapId)

    newX, newY, newMapId := player.GetLocation()

    if newX != 0 {
        t.Errorf("Expected x to be 0, but got %d", newX)
    }

    if newY != 0 {
        t.Errorf("Expected y to be 0, but got %d", newY)
    }

    if newMapId != mapId {
        t.Errorf("Expected mapId to be %d, but got %d", mapId, newMapId)
    }
}

func TestAddExp_Overflow(t *testing.T) {
    player := NewPlayer("John", 20, 2, 10, 5)

    exp := 1_000_000_000
    player.AddExp(exp)

    if player.GetExp() != player.MaxExp {
        t.Errorf("Expected exp to be %d, but got %d", player.MaxExp, player.GetExp())
    }
}

func TestDamage_MoreThanCurrentHP(t *testing.T) {
    player := NewPlayer("John", 20, 2, 10, 5)

    damage := 30
    player.Damage(damage)

    expectedHP := 0

    if player.GetCurrentHP() != expectedHP {
        t.Errorf("Expected current HP to be %d, but got %d", expectedHP, player.GetCurrentHP())
    }
}

func TestSetLevel_NegativeLevel(t *testing.T) {
    player := NewPlayer("John", 20, 2, 10, 5)

    level := -1
    player.SetLevel(level)

    if player.GetLevel() != 1 {
        t.Errorf("Expected level to be 1, but got %d", player.GetLevel())
    }
}

func TestAddExp_NegativeAmount(t *testing.T) {
    player := NewPlayer("John", 20, 2, 10, 5)

    exp := -100
    player.AddExp(exp)

    if player.GetExp() != 0 {
        t.Errorf("Expected exp to be 0, but got %d", player.GetExp())
    }
}

func TestNewPlayer_Invalid(t *testing.T) {
    // Test invalid name
    invalidName := ""
    _, err := NewPlayer(invalidName, 20, 2, 10, 5)
    if err == nil {
        t.Errorf("Expected error for invalid name, but got none")
    }

    // Test invalid maxHP
    invalidMaxHP := -10
    _, err = NewPlayer("John", invalidMaxHP, 2, 10, 5)
    if err == nil {
        t.Errorf("Expected error for invalid maxHP, but got none")
    }

    // Test invalid regenRate
    invalidRegenRate := -2
    _, err = NewPlayer("John", 20, invalidRegenRate, 10, 5)
    if err == nil {
        t.Errorf("Expected error for invalid regenRate, but got none")
    }

    // Test invalid strength
    invalidStrength := -10
    _, err = NewPlayer("John", 20, 2, invalidStrength, 5)
    if err == nil {
        t.Errorf("Expected error for invalid strength, but got none")
    }

    // Test invalid constitution
    invalidConstitution := -5
    _, err = NewPlayer("John", 20, 2, 10, invalidConstitution)
    if err == nil {
        t.Errorf("Expected error for invalid constitution, but got none")
    }
}

func TestSetLocation_Invalid(t *testing.T) {
    player := NewPlayer("John", 20, 2, 10, 5)

    // Test invalid x coordinate
    invalidX := -1
    player.SetLocation(invalidX, 20, 30)
    newX, _, _ := player.GetLocation()
    if newX != 0 {
        t.Errorf("Expected x to be 0 for invalid x coordinate, but got %d", newX)
    }

    // Test invalid y coordinate
    invalidY := -1
    player.SetLocation(10, invalidY, 30)
    _, newY, _ := player.GetLocation()
    if newY != 0 {
        t.Errorf("Expected y to be 0 for invalid y coordinate, but got %d", newY)
    }

    // Test invalid map ID
    invalidMapID := -1
    player.SetLocation(10, 20, invalidMapID)
    _, _, newMapID := player.GetLocation()
    if newMapID != 0 {
        t.Errorf("Expected map ID to be 0 for invalid map ID, but got %d", newMapID)
    }
}

func TestSetLevel_Invalid(t *testing.T) {
    player := NewPlayer("John", 20, 2, 10, 5)

    // Test invalid level
    invalidLevel := -1
    player.SetLevel(invalidLevel)
    if player.GetLevel() != 1 {
        t.Errorf("Expected level to be 1 for invalid level, but got %d", player.GetLevel())
    }
}

func TestCalculateMaxHP(t *testing.T) {
    player := &Player{
        level:        5,
        constitution: 10,
    }

    expected := 65
    actual := player.calculateMaxHP()

    if actual != expected {
        t.Errorf("calculateMaxHP() returned %d, expected %d", actual, expected)
    }
}

func TestMove(t *testing.T) {
    player := &Player{
        x: 10,
        y: 10,
    }

    player.Move("North", 2)

    expectedX := 10
    expectedY := 12

    if player.x != expectedX || player.y != expectedY {
        t.Errorf("Move() failed. Expected (%d, %d), got (%d, %d)", expectedX, expectedY, player.x, player.y)
    }
}

func TestGetDistanceTo(t *testing.T) {
    player := &Player{
        x: 10,
        y: 10,
    }

    distance := player.GetDistanceTo(13, 14)

    expected := math.Sqrt(13*13 + 14*14)

    if distance != expected {
        t.Errorf("GetDistanceTo() failed. Expected %f, got %f", expected, distance)
    }
}

func TestHeal(t *testing.T) {
    player := &Player{
        hp:    5,
        maxHP: 10,
    }

    player.Heal(5)

    expected := 10
    actual := player.hp

    if actual != expected {
        t.Errorf("Heal() failed. Expected %d, got %d", expected, actual)
    }
}

func TestIsAlive(t *testing.T) {
    player := &Player{
        hp: 0,
    }

    if player.IsAlive() {
        t.Errorf("IsAlive() failed. Expected false, got true")
    }

    player.hp = 10

    if !player.IsAlive() {
        t.Errorf("IsAlive() failed. Expected true, got false")
    }
}

func TestTakeHealing(t *testing.T) {
    player := &Player{
        hp:    5,
        maxHP: 10,
        id:    uuid.New(),
    }

    player.TakeHealing(3)

    expectedHP := 8

    if player.hp != expectedHP {
        t.Errorf("TakeHealing() failed. Expected HP %d, got %d", expectedHP, player.hp)
    }

    player.TakeHealing(6)

    expectedHP = 10

    if player.hp != expectedHP {
        t.Errorf("TakeHealing() failed. Expected HP %d, got %d", expectedHP, player.hp)
    }

    // Check message output for already full HP
    msg := captureOutput(func() {
        player.TakeHealing(0)
    })

    expectedMsg := fmt.Sprintf("Player %s is already at full HP\n", player.id.String())

    if msg != expectedMsg {
        t.Errorf("TakeHealing() failed. Expected message '%s', got '%s'", expectedMsg, msg)
    }
}