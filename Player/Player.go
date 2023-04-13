package Player

import (
    "fmt"
    "math"
    "math/rand"
    "time"
    "github.com/google/uuid"
)


type Player struct {
    id            uuid.UUID
    name          string
    hp            int
    maxHP         int
    armorRating   float64
    x             int
    y             int
    mapId         int
    level         int
    exp           int
    expCurve      float64
    expModifier   float64
    regenRate     int
    lastRegenTime time.Time
    strength      int
    constitution  int
    regenInterval int // new field for regenInterval
}


func NewPlayer(name string, maxHP, regenRate, strength, constitution int) *Player {
    player := &Player{
        id:            uuid.New(),
        name:          name,
        armorRating:   0.0,
        x:             0,
        y:             0,
        mapId:         0,
        level:         1,
        exp:           0,
        expCurve:      1.2,
        expModifier:   1.0,
        regenRate:     regenRate,
        lastRegenTime: time.Now(),
        strength:      strength,
        constitution:  constitution,
        maxHP:         10 + (1-1)*(constitution+5),
        hp:            10 + (1-1)*(constitution+5),
        regenInterval: 10, // set regenInterval to 10 seconds
    }

    return player
}

func (p *Player) calculateMaxHP() int {
    return 10 + (p.level-1) * (p.constitution + 5)
}


func (p *Player) SetName(name string) {
    p.name = name
}

func (p *Player) SetLocation(x, y, mapId int) {
    p.x = x
    p.y = y
    p.mapId = mapId
}

func (p *Player) SetLevel(level int) {
    p.level = level
}

func (p *Player) AddExp(exp int) {
    p.exp += int(float64(exp) * p.expModifier)
    for p.exp >= int(float64(p.level)*p.expCurve*100) {
    p.exp -= int(float64(p.level)*p.expCurve*100)
    p.level++
}
}

func (p *Player) String() string {
    return fmt.Sprintf("Player %s (%s): Level %d, Exp %d, Location (%d,%d) on Map %d", p.id.String(), p.name, p.level, p.exp, p.x, p.y, p.mapId)
}

func (p *Player) Move(direction string, distance int) {
    switch direction {
    case "North":
        p.y += distance
    case "East":
        p.x += distance
    case "South":
        p.y -= distance
    case "West":
        p.x -= distance
    }
}

func (p *Player) GetID() uuid.UUID {
    return p.id
}

func (p *Player) GetLocation() (int, int, int) {
    return p.x, p.y, p.mapId
}

func (p *Player) GetLevel() int {
    return p.level
}

func (p *Player) GetExp() int {
    return p.exp
}

func (p *Player) SetExpCurve(curve float64) {
    p.expCurve = curve
}

func (p *Player) SetExpModifier(modifier float64) {
    p.expModifier = modifier
}

func (p *Player) SetMapId(mapId int) {
    p.mapId = mapId
}

func (p *Player) GetExpCurve() float64 {
    return p.expCurve
}

func (p *Player) GetExpModifier() float64 {
    return p.expModifier
}

func (p *Player) GetDistanceTo(x, y int) float64 {
    deltaX := float64(p.x - x)
    deltaY := float64(p.y - y)
    return math.Sqrt(deltaX*deltaX + deltaY*deltaY)
}

func (p *Player) SetHP(hp int) {
    p.hp = hp
}

func (p *Player) GetHP() int {
    return p.hp
}

func (p *Player) GetMaxHP() int {
    return p.maxHP
}

func (p *Player) SetMaxHP(maxHP int) {
    p.maxHP = maxHP
}


func (p *Player) SetArmorRating(armorRating float64) {
    p.armorRating = armorRating
}

func (p *Player) GetArmorRating() float64 {
    return p.armorRating
}


func (p *Player) Heal(healing int) {
    p.hp += healing
    if p.hp > p.maxHP {
        p.hp = p.maxHP
    }
}

func (p *Player) IsAlive() bool {
    return p.hp > 0
}

func (p *Player) TakeHealing(healing int) {
    missingHP := p.maxHP - p.hp
    if missingHP == 0 {
        fmt.Println("Player is already at full HP")
        return
    }

    if healing > missingHP {
        p.hp = p.maxHP
    } else {
        p.hp += healing
    }

    fmt.Printf("Player %s received %d healing. HP: %d/%d\n", p.id.String(), healing, p.hp, p.maxHP)
}

func (p *Player) Attack(target *Player, attackPower int, criticalChance float64) {
    if !p.IsAlive() {
        fmt.Println("Attacking player is dead")
        return
    }

    if !target.IsAlive() {
        fmt.Println("Target player is already dead")
        return
    }

    if rand.Float64() < criticalChance {
        attackPower *= 2
        fmt.Printf("Critical hit! Attack power doubled to %d.\n", attackPower)
    }

    // include strength in attack power calculation
    attackPower *= p.strength

    damageDealt := int(math.Round(float64(attackPower) * (1 - target.armorRating)))
    target.SetHP(target.GetHP() - damageDealt)

    fmt.Printf("Player %s attacked player %s for %d damage (%d -> %d).\n", p.id.String(), target.GetID().String(), damageDealt, target.GetHP()+damageDealt, target.GetHP())

    if !target.IsAlive() {
        fmt.Printf("Player %s has been defeated!\n", target.GetName())
        p.AddExp(target.GetLevel() * 100)
    }
}


func (p *Player) TakeDamage(damage int) {
    if !p.IsAlive() {
        fmt.Println("Defending player is already dead")
        return
    }

    actualDamage := int(float64(damage) * (1.0 - p.armorRating))
    p.hp -= actualDamage
    if p.hp < 0 {
        p.hp = 0
    }

    if !p.IsAlive() {
        fmt.Printf("Player %s has been defeated\n", p.id.String())
    }
}



func (p *Player) RegenerateHealth() {
    now := time.Now()
    timeSinceLastRegen := now.Sub(p.lastRegenTime).Seconds()

    if timeSinceLastRegen >= float64(p.regenInterval) {
        p.lastRegenTime = now
        regenAmount := int(math.Round(float64(p.constitution) / 5.0))
        p.Heal(regenAmount)
    }
}

func (p *Player) Regenerate() {
    now := time.Now()
    if now.Sub(p.lastRegenTime).Seconds() >= float64(p.regenRate) {
        p.Heal(p.constitution / 2)
        p.lastRegenTime = now
    }
}


