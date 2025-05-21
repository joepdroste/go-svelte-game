package protocol

// --- Generic Message Wrapper ---

// GenericMessage is a wrapper for all messages to include a type.
type GenericMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"` // The actual data for the message
}

// --- Client-to-Server (C2S) Message Payloads ---

// C2S_MovePayload defines the structure for a client's move input.
type C2S_MovePayload struct {
	DX int `json:"dx"`
	DY int `json:"dy"`
}

// (Add more C2S payloads here as needed, e.g., for attacks)
// type C2S_AttackPayload struct {
//    TargetID string `json:"target_id"`
// }

// --- Server-to-Client (S2C) Message Payloads ---

// S2C_TileData represents a single tile for sending to the client.
type S2C_TileData struct {
	Type TileType `json:"type"`
	// X and Y can be inferred by position in array if sent as a grid
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
	// Add other relevant stats as needed by the frontend:
	// XP        int    `json:"xp,omitempty"`
	// Attack    int    `json:"attack,omitempty"`
	// Defense   int    `json:"defense,omitempty"`
}

// S2C_MonsterData represents a monster's state to be sent to clients.
type S2C_MonsterData struct {
	ID        string      `json:"id"`
	X         int         `json:"x"`
	Y         int         `json:"y"`
	Type      MonsterType `json:"type"` // e.g., "Goblin", "Orc"
	Name      string      `json:"name"`
	MaxHP     int         `json:"max_hp"`
	CurrentHP int         `json:"current_hp"`
	// Add other relevant stats as needed:
	// Attack    int    `json:"attack,omitempty"`
	// Defense   int    `json:"defense,omitempty"`
}

// S2C_InitialStatePayload is sent to a client upon successful connection.
type S2C_InitialStatePayload struct {
	PlayerID string            `json:"player_id"` // The ID assigned to this client's player
	Map      S2C_MapData       `json:"map"`
	Players  []S2C_PlayerData  `json:"players"`
	Monsters []S2C_MonsterData `json:"monsters"`
}

// S2C_PlayerJoinedPayload is broadcast when a new player joins.
// It reuses S2C_PlayerData by embedding.
type S2C_PlayerJoinedPayload struct {
	S2C_PlayerData // Embeds ID, X, Y, Level, MaxHP, CurrentHP
}

// S2C_PlayerLeftPayload is broadcast when a player disconnects.
type S2C_PlayerLeftPayload struct {
	ID string `json:"id"` // ID of the player who left
}

// S2C_EntityMovedPayload is broadcast when any entity (player or monster) moves.
type S2C_EntityMovedPayload struct {
	ID         string `json:"id"`          // ID of the entity that moved
	EntityType string `json:"entity_type"` // "player" or "monster"
	X          int    `json:"x"`           // New X coordinate
	Y          int    `json:"y"`           // New Y coordinate
}

// S2C_MonsterSpawnedPayload is broadcast when a new monster appears (if dynamic spawning).
// It reuses S2C_MonsterData by embedding.
type S2C_MonsterSpawnedPayload struct {
	S2C_MonsterData // Embeds ID, X, Y, Type, Name, MaxHP, CurrentHP
}

// S2C_EntityRemovedPayload is broadcast when an entity is removed (e.g., monster defeated).
type S2C_EntityRemovedPayload struct {
	ID         string `json:"id"`          // ID of the entity removed
	EntityType string `json:"entity_type"` // "player" or "monster"
}

type S2C_CombatInitiatedPayload struct {
	PlayerID  string `json:"player_id"`
	MonsterID string `json:"monster_id"`
	// Player's current X,Y (they don't move onto monster tile for initiation)
	PlayerX int `json:"player_x"`
	PlayerY int `json:"player_y"`
	// Monster's current X,Y
	MonsterX int `json:"monster_x"`
	MonsterY int `json:"monster_y"`
}

// (Add more S2C payloads here as needed, e.g., for combat updates, stat changes)
// type S2C_CombatUpdatePayload struct { ... }
// type S2C_PlayerStatUpdatePayload struct { ... }

// --- Message Type Constants ---
// Helps avoid magic strings when creating and parsing messages.

// C2S (Client to Server) Message Types
const (
	C2S_MessageTypeMove = "move"
	// C2S_MessageTypeAttack = "attack" // Example for later
)

// S2C (Server to Client) Message Types
const (
	S2C_MessageTypeInitialState    = "initial_state"
	S2C_MessageTypePlayerJoined    = "player_joined"
	S2C_MessageTypePlayerLeft      = "player_left"
	S2C_MessageTypeEntityMoved     = "entity_moved"
	S2C_MessageTypeMonsterSpawned  = "monster_spawned"
	S2C_MessageTypeEntityRemoved   = "entity_removed"
	S2C_MessageTypeCombatInitiated = "combat_initiated"
	// S2C_MessageTypeCombatUpdate   = "combat_update"   // Example for later
	// S2C_MessageTypeStatUpdate     = "stat_update"     // Example for later
	// S2C_MessageTypeNotification = "notification"    // For general text messages
)

// --- END OF FILE internal/protocol/messages.go ---
