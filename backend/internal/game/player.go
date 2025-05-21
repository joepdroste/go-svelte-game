package game

import "fmt"

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
	IsInCombat     bool
	CombatTargetID string
}

func NewPlayer(id string, startX, startY int) *Player {
	return &Player{
		ID:             id,
		X:              startX,
		Y:              startY,
		Level:          1,
		XP:             0,
		MaxHP:          100,
		CurrentHP:      100,
		Attack:         10,
		Defense:        5,
		IsInCombat:     false,
		CombatTargetID: "",
	}
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
}
