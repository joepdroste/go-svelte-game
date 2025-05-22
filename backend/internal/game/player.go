package game

import (
	"fmt"
	"log"
)

type Player struct {
	ID        string
	X         int
	Y         int
	Level     int
	XP        int
	MaxHP     int
	CurrentHP int
	Attack    int
	Defense   int

	// Combat State
	XPToNextLevel  int
	IsInCombat     bool
	CombatTargetID string
}

// TODO replace with math
var xpThresholds = map[int]int{
	1: 100,
	2: 250,
	3: 500,
	4: 1000,
}

func NewPlayer(id string, startX, startY int) *Player {
	initialLevel := 1
	return &Player{
		ID:             id,
		X:              startX,
		Y:              startY,
		Level:          initialLevel,
		XP:             0,
		MaxHP:          100,
		CurrentHP:      100,
		Attack:         10,
		Defense:        5,
		XPToNextLevel:  CalculateXPToNextLevel(initialLevel),
		IsInCombat:     false,
		CombatTargetID: "",
	}
}

func CalculateXPToNextLevel(level int) int {
	if nextXP, ok := xpThresholds[level]; ok {
		return nextXP
	}

	return 1000 + (level-4)*500
}

func (p *Player) GetID() string {
	return p.ID
}

func (p *Player) GetX() int {
	return p.X
}

func (p *Player) GetY() int {
	return p.Y
}

func (p *Player) Move(dx, dy int, world *World) (moved bool, engagedMonster *Monster) {
	if p.IsInCombat {
		fmt.Printf("Player %s tried to move while in combat. Move denied.\n", p.GetID())
		return false, nil
	}

	newX, newY := p.X+dx, p.Y+dy

	if monster := world.getMonsterAtInternal(newX, newY); monster != nil {
		if monster.IsInCombat {
			fmt.Printf("Player %s tried to engage Monster %s, but monster is already in combat with %s.\n", p.GetID(), monster.GetID(), monster.CombatTargetID)
			return false, nil
		}
		fmt.Printf("Player %s attempts to engage Monster %s at (%d,%d).\n", p.GetID(), monster.GetID(), newX, newY)
		return false, monster
	}

	if world.IsWalkable(newX, newY) {
		if otherPlayer := world.getPlayerAtInternal(newX, newY); otherPlayer != nil && otherPlayer.GetID() != p.GetID() {
			return false, nil
		}
		p.X = newX
		p.Y = newY
		return true, nil
	}
	return false, nil
}

func (p *Player) GainXP(amount int) (leveledUp bool) {
	if amount <= 0 {
		return false
	}

	p.XP += amount
	log.Printf("Player %s gained %d XP. Total XP: %d. Needed for next level: %d", p.GetID(), amount, p.XP, p.XPToNextLevel)

	leveledUp = false
	for p.XP >= p.XPToNextLevel {
		p.XP -= p.XPToNextLevel
		p.LevelUp()
		leveledUp = true
	}
	return leveledUp
}

func (p *Player) LevelUp() {
	p.Level++
	log.Printf("Player %s LEVELED UP to Level %d!", p.GetID(), p.Level)

	p.MaxHP += 20
	p.CurrentHP = p.MaxHP
	p.Attack += 2
	p.Defense += 1

	p.XPToNextLevel = CalculateXPToNextLevel(p.Level)

	log.Printf("Player %s new stats: Level %d, MaxHP %d, Attack %d, Defense %d, XP for next: %d",
		p.GetID(), p.Level, p.MaxHP, p.Attack, p.Defense, p.XPToNextLevel)
}

// Returns true if player is defeated
func (p *Player) TakeDamage(amount int) bool {
	p.CurrentHP -= amount
	if p.CurrentHP <= 0 {
		p.CurrentHP = 0
		return true
	}
	return false
}

func (p *Player) ResetToLevel1() {
	p.Level = 1
	p.XP = 0
	p.MaxHP = 100
	p.CurrentHP = p.MaxHP
	p.Attack = 10
	p.Defense = 5
	p.XPToNextLevel = CalculateXPToNextLevel(p.Level)
	p.IsInCombat = false
	p.CombatTargetID = ""
}
