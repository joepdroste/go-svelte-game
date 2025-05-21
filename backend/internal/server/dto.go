package server

import (
	"game-server/internal/game"
	"game-server/internal/protocol"
)

func NewS2C_MapData(world *game.World) protocol.S2C_MapData {
	s2cTiles := make([][]protocol.S2C_TileData, world.Height)
	for y := 0; y < world.Height; y++ {
		s2cTiles[y] = make([]protocol.S2C_TileData, world.Width)
		for x := 0; x < world.Width; x++ {
			s2cTiles[y][x] = protocol.S2C_TileData{Type: world.Tiles[y][x].Type}
		}
	}
	return protocol.S2C_MapData{
		Width:  world.Width,
		Height: world.Height,
		Tiles:  s2cTiles,
	}
}

func NewS2C_PlayerData(p *game.Player) protocol.S2C_PlayerData {
	return protocol.S2C_PlayerData{
		ID:        p.GetID(),
		X:         p.GetX(),
		Y:         p.GetY(),
		Level:     p.Level,
		MaxHP:     p.MaxHP,
		CurrentHP: p.CurrentHP,
	}
}

func NewS2C_MonsterData(m *game.Monster) protocol.S2C_MonsterData {
	return protocol.S2C_MonsterData{
		ID:        m.GetID(),
		X:         m.GetX(),
		Y:         m.GetY(),
		Type:      m.Type,
		Name:      m.Name,
		MaxHP:     m.MaxHP,
		CurrentHP: m.CurrentHP,
	}
}
