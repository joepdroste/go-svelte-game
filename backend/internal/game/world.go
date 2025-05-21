package game

import (
	"encoding/json"
	"fmt"
	"game-server/internal/protocol"
	"log"
	"math/rand"
	"sync"
)

type Tile struct {
	Type protocol.TileType
	X, Y int
}

type HubBroadcaster interface {
	Broadcast(message []byte)
}

type World struct {
	Width    int
	Height   int
	Tiles    [][]Tile
	Monsters map[string]*Monster
	Players  map[string]*Player
	Mu       sync.Mutex
	hub      HubBroadcaster
}

func (w *World) SetHubBroadcaster(broadcaster HubBroadcaster) {
	w.hub = broadcaster
}

func NewWorld(width, height int) *World {
	tiles := make([][]Tile, height)
	for y := 0; y < height; y++ {
		tiles[y] = make([]Tile, width)
		for x := 0; x < width; x++ {
			tileType := protocol.Grass

			if rand.Intn(100) < 20 {
				tileType = protocol.Stone
			}

			if x == 0 || y == 0 || x == width-1 || y == height-1 {
				tileType = protocol.Stone
			}
			tiles[y][x] = Tile{Type: tileType, X: x, Y: y}
		}
	}

	world := &World{
		Width:    width,
		Height:   height,
		Tiles:    tiles,
		Monsters: make(map[string]*Monster),
		Players:  make(map[string]*Player),
		hub:      nil,
	}
	return world
}

func (w *World) AddMonster(m *Monster) {
	w.Mu.Lock()
	w.Monsters[m.GetID()] = m
	w.Mu.Unlock()

	go m.RunAI(w)
}

func (w *World) RemoveMonster(MonsterID string) {
	w.Mu.Lock()
	defer w.Mu.Unlock()

	if monster, ok := w.Monsters[MonsterID]; ok {
		if monster.stopAI != nil {
			close(monster.stopAI)
		}
		delete(w.Monsters, MonsterID)
		fmt.Printf("Monster %s removed.\n", MonsterID)
	}
}

func (w *World) SpawnInitialMonsters(count int) {
	for i := 0; i < count; i++ {
		fmt.Print(i)
		id := fmt.Sprintf("monster-%03d", i)
		var mType protocol.MonsterType
		if rand.Intn(2) == 0 {
			mType = protocol.Goblin
		} else {
			mType = protocol.Orc
		}

		var spawnX, spawnY int
		for {
			spawnX = rand.Intn(w.Width)
			spawnY = rand.Intn(w.Height)
			if w.IsWalkable(spawnX, spawnY) && w.GetMonsterAt(spawnX, spawnY) == nil {
				break
			}
		}
		monster := NewMonster(id, mType, spawnX, spawnY)
		w.AddMonster(monster)
	}
}

func (w *World) AddPlayer(p *Player) {
	w.Players[p.GetID()] = p
}

func (w *World) RemovePlayer(playerID string) {
	w.Mu.Lock()
	delete(w.Players, playerID)
	w.Mu.Unlock()
}

func (w *World) GetPlayer(playerID string) *Player {
	w.Mu.Lock()
	defer w.Mu.Unlock()
	return w.Players[playerID]
}

func (w *World) GetMonster(monsterID string) *Monster {
	w.Mu.Lock()
	defer w.Mu.Unlock()
	return w.Monsters[monsterID]
}

func (w *World) getMonsterAtInternal(x, y int) *Monster {
	// Assumes w.Mu is HELD
	for _, m := range w.Monsters {
		if m.GetX() == x && m.GetY() == y {
			return m
		}
	}
	return nil
}

func (w *World) getPlayerAtInternal(x, y int) *Player {
	// Assumes w.Mu is HELD
	for _, p := range w.Players {
		if p.GetX() == x && p.GetY() == y {
			return p
		}
	}
	return nil
}

// IsOccupiedInternal assumes w.Mu is HELD by caller
func (w *World) IsOccupiedInternal(x, y int) bool {
	if w.getMonsterAtInternal(x, y) != nil { // Uses internal getter
		return true
	}
	if w.getPlayerAtInternal(x, y) != nil { // Uses internal getter
		return true
	}
	return false
}

func (w *World) GetMonsterAt(x, y int) *Monster {
	w.Mu.Lock()
	defer w.Mu.Unlock()
	for _, m := range w.Monsters {
		if m.GetX() == x && m.GetY() == y {
			return m
		}
	}
	return nil
}

func (w *World) GetPlayerAt(x, y int) *Player {
	w.Mu.Lock()
	defer w.Mu.Unlock()
	for _, p := range w.Players {
		if p.GetX() == x && p.GetY() == y {
			return p
		}
	}
	return nil
}

func (w *World) IsOccupied(x, y int) bool {
	if w.GetMonsterAt(x, y) != nil {
		return true
	}
	if w.GetPlayerAt(x, y) != nil {
		return true
	}
	return false
}

func (w *World) GetTile(x, y int) *Tile {
	if x < 0 || x >= w.Width || y < 0 || y >= w.Height {
		return nil
	}
	return &w.Tiles[y][x]
}

func (w *World) IsWalkable(x, y int) bool {
	tile := w.GetTile(x, y)
	if tile == nil {
		return false // out of bounds
	}
	return tile.Type == protocol.Grass
}

func (w *World) String() string {
	w.Mu.Lock()
	defer w.Mu.Unlock()

	grid := make([][]rune, w.Height)
	for y := 0; y < w.Height; y++ {
		grid[y] = make([]rune, w.Width)
		for x := 0; x < w.Width; x++ {
			tile := w.Tiles[y][x]
			switch tile.Type {
			case protocol.Grass:
				grid[y][x] = '.'
			case protocol.Stone:
				grid[y][x] = '#'
			default:
				grid[y][x] = '?'
			}
		}
	}

	for _, monster := range w.Monsters {
		mx, my := monster.GetX(), monster.GetY()
		if my >= 0 && my < w.Height && mx >= 0 && mx < w.Width {
			if w.Tiles[my][mx].Type == protocol.Grass {
				switch monster.Type {
				case protocol.Goblin:
					grid[my][mx] = 'g'
				case protocol.Orc:
					grid[my][mx] = 'O'
				default:
					grid[my][mx] = 'M'
				}
			}
		}
	}

	for _, player := range w.Players {
		px, py := player.GetX(), player.GetY()
		if py >= 0 && py < w.Height && px >= 0 && px < w.Width {
			if w.Tiles[py][px].Type == protocol.Grass {
				grid[py][px] = '@'
			}
		}
	}

	var s string
	for y := 0; y < w.Height; y++ {
		s += string(grid[y])
		s += "\n"
	}
	return s
}

func (w *World) MoveMonster(m *Monster, newX, newY int) bool {
	if !w.IsWalkable(newX, newY) {
		return false
	}

	for _, otherMonster := range w.Monsters {
		if otherMonster.GetID() != m.GetID() && otherMonster.GetX() == newX && otherMonster.GetY() == newY {
			return false
		}
	}
	for _, player := range w.Players {
		if player.GetX() == newX && player.GetY() == newY {
			return false
		}
	}

	m.X = newX
	m.Y = newY

	if w.hub != nil {
		movedPayload := protocol.S2C_EntityMovedPayload{
			ID:         m.GetID(),
			EntityType: protocol.EntityTypeMonster,
			X:          newX,
			Y:          newY,
		}
		msg := protocol.GenericMessage{
			Type:    protocol.S2C_MessageTypeEntityMoved,
			Payload: movedPayload,
		}
		jsonMsg, err := json.Marshal(msg)
		if err != nil {
			log.Printf("Error marshaling monster move update for %s: %v", m.GetID(), err)
		} else {
			w.hub.Broadcast(jsonMsg)
		}
	}
	return true
}
