package game

import (
	"game-server/internal/protocol"
	"log"
	"math/rand"
	"time"
)

type Monster struct {
	ID string
	X  int
	Y  int
	// stats
	Type      protocol.MonsterType
	Name      string
	MaxHP     int
	CurrentHP int
	Attack    int
	Defense   int
	XPValue   int

	// Combat State
	IsInCombat     bool
	CombatTargetID string

	stopAI chan struct{}
}

func NewMonster(id string, mType protocol.MonsterType, x, y int) *Monster {
	monster := &Monster{
		ID:             id,
		Type:           mType,
		X:              x,
		Y:              y,
		IsInCombat:     false,
		CombatTargetID: "",
		stopAI:         make(chan struct{}),
	}

	switch mType {
	case protocol.Goblin:
		monster.Name = "Goblin"
		monster.MaxHP = 30
		monster.CurrentHP = 30
		monster.Attack = 8
		monster.Defense = 3
		monster.XPValue = 10
	case protocol.Orc:
		monster.Name = "Orc"
		monster.MaxHP = 70
		monster.CurrentHP = 70
		monster.Attack = 15
		monster.Defense = 8
		monster.XPValue = 25
	default:
		monster.Name = "Mysterious Creature"
		monster.MaxHP = 50
		monster.CurrentHP = 50
		monster.Attack = 10
		monster.Defense = 5
		monster.XPValue = 15
	}
	return monster
}

func (m *Monster) GetID() string {
	return m.ID
}

func (m *Monster) GetX() int {
	return m.X
}

func (m *Monster) GetY() int {
	return m.Y
}

func (m *Monster) TakeDamage(amount int) bool {
	m.CurrentHP -= amount
	if m.CurrentHP <= 0 {
		m.CurrentHP = 0
		return true // Defeated
	}
	return false
}

func (m *Monster) RunAI(w *World) {
	initialDelay := time.Duration(rand.Intn(750)+250) * time.Millisecond // 0.25s to 1s
	time.Sleep(initialDelay)

	ticker := time.NewTicker(time.Duration(rand.Intn(750)+250) * time.Millisecond) // Random tick between 0.25s and 1s
	defer ticker.Stop()

	for {
		select {
		case <-m.stopAI:
			return
		case <-ticker.C:
			w.Mu.Lock()

			if m.IsInCombat {
				log.Printf("Monster %s (%s) is in combat with %s, not moving.", m.ID, m.Name, m.CombatTargetID)
			} else {
				dx, dy := 0, 0
				r := rand.Intn(4)
				switch r {
				case 0: // Up
					dy = -1
				case 1: // Down
					dy = 1
				case 2: // Left
					dx = -1
				case 3: // Right
					dx = 1
				}

				currentX, currentY := m.X, m.Y
				newX, newY := currentX+dx, currentY+dy

				w.MoveMonster(m, newX, newY)
			}
			w.Mu.Unlock()
		}
	}
}
