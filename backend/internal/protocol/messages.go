package protocol

// --- Generic Message Wrapper ---

// GenericMessage is a wrapper for all messages to include a type.
type GenericMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// --- Client-to-Server (C2S) Message Payloads ---
type C2S_MovePayload struct {
	DX int `json:"dx"`
	DY int `json:"dy"`
}

type C2S_AttackPayload struct {
	TargetID string `json:"target_id"`
}

type C2S_UsePotionPayload struct {
}

// --- Server-to-Client (S2C) Message Payloads ---
type S2C_TileData struct {
	Type TileType `json:"type"`
}

// S2C_MapData represents the entire map structure.
type S2C_MapData struct {
	Width  int              `json:"width"`
	Height int              `json:"height"`
	Tiles  [][]S2C_TileData `json:"tiles"`
}

// S2C_PlayerData represents a player's state to be sent to clients.
type S2C_PlayerData struct {
	ID        string `json:"id"`
	X         int    `json:"x"`
	Y         int    `json:"y"`
	Level     int    `json:"level"`
	MaxHP     int    `json:"max_hp"`
	CurrentHP int    `json:"current_hp"`
}

// S2C_MonsterData represents a monster's state to be sent to clients.
type S2C_MonsterData struct {
	ID        string      `json:"id"`
	X         int         `json:"x"`
	Y         int         `json:"y"`
	Type      MonsterType `json:"type"`
	Name      string      `json:"name"`
	MaxHP     int         `json:"max_hp"`
	CurrentHP int         `json:"current_hp"`
}

// S2C_InitialStatePayload is sent to a client upon successful connection.
type S2C_InitialStatePayload struct {
	PlayerID string            `json:"player_id"`
	Map      S2C_MapData       `json:"map"`
	Players  []S2C_PlayerData  `json:"players"`
	Monsters []S2C_MonsterData `json:"monsters"`
}

// S2C_PlayerJoinedPayload is broadcast when a new player joins.
type S2C_PlayerJoinedPayload struct {
	S2C_PlayerData
}

// S2C_PlayerLeftPayload is broadcast when a player disconnects.
type S2C_PlayerLeftPayload struct {
	ID string `json:"id"`
}

// S2C_EntityMovedPayload is broadcast when any entity (player or monster) moves.
type S2C_EntityMovedPayload struct {
	ID         string `json:"id"`
	EntityType string `json:"entity_type"`
	X          int    `json:"x"`
	Y          int    `json:"y"`
}

// S2C_MonsterSpawnedPayload is broadcast when a new monster appears (if dynamic spawning).
type S2C_MonsterSpawnedPayload struct {
	S2C_MonsterData
}

// S2C_EntityRemovedPayload is broadcast when an entity is removed (e.g., monster defeated).
type S2C_EntityRemovedPayload struct {
	ID         string `json:"id"`
	EntityType string `json:"entity_type"`
}

type S2C_CombatInitiatedPayload struct {
	PlayerID  string `json:"player_id"`
	MonsterID string `json:"monster_id"`

	PlayerX int `json:"player_x"`
	PlayerY int `json:"player_y"`

	MonsterX int `json:"monster_x"`
	MonsterY int `json:"monster_y"`
}

type S2C_CombatUpdatePayload struct {
	AttackerID         string `json:"attacker_id"`
	DefenderID         string `json:"defender_id"`
	DamageDealt        int    `json:"damage_dealt"`
	DefenderCurrentHP  int    `json:"defender_current_hp"`
	IsDefenderDefeated bool   `json:"is_defender_defeated"`
}

type S2C_PlayerStatUpdatePayload struct {
	PlayerID      string `json:"player_id"`
	Level         int    `json:"level"`
	XP            int    `json:"xp"`
	XPToNextLevel int    `json:"xp_to_next_level"`
	MaxHP         int    `json:"max_hp"`
	CurrentHP     int    `json:"current_hp"`
	Attack        int    `json:"attack"`
	Defense       int    `json:"defense"`
}

type S2C_NotificationPayload struct {
	Message string `json:"message"`
	Level   string `json:"level"`
}

// --- Message Type Constants ---
// C2S (Client to Server) Message Types
const (
	C2S_MessageTypeMove   = "move"
	C2S_MessageTypeAttack = "attack"
)

// S2C (Server to Client) Message Types
const (
	S2C_MessageTypeInitialState     = "initial_state"
	S2C_MessageTypePlayerJoined     = "player_joined"
	S2C_MessageTypePlayerLeft       = "player_left"
	S2C_MessageTypeEntityMoved      = "entity_moved"
	S2C_MessageTypeMonsterSpawned   = "monster_spawned"
	S2C_MessageTypeEntityRemoved    = "entity_removed"
	S2C_MessageTypeCombatInitiated  = "combat_initiated"
	S2C_MessageTypeCombatUpdate     = "combat_update"
	S2C_MessageTypePlayerStatUpdate = "player_stat_update"
	C2S_MessageTypeUsePotion        = "use_potion"
	S2C_MessageTypeNotification     = "notification"
)
