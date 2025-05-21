package protocol

type TileType int

const (
	Grass TileType = iota
	Stone
)

type MonsterType string

const (
	Goblin MonsterType = "Goblin"
	Orc    MonsterType = "Orc"
)

// You might also want the EntityType constants here if they are fundamental
const (
	EntityTypePlayer  = "player"
	EntityTypeMonster = "monster"
)
