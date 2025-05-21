// --- Core Game Data Types ---
export enum TileType {
    Grass = 0,
    Stone = 1,
}

export enum MonsterType {
    Goblin = "Goblin",
    Orc = "Orc",
}

export const EntityTypePlayer = "player";
export const EntityTypeMonster = "monster";

export interface GenericMessage<T = any> {
	type: string;
	payload: T;
}

export interface C2S_MovePayload {
	dx: number;
	dy: number;
}

export interface S2C_TileData {
	type: TileType;
}

export interface S2C_MapData {
	width: number;
	height: number;
	tiles: S2C_TileData[][];
}

export interface S2C_PlayerData {
	id: string;
	x: number;
	y: number;
	level: number;
	max_hp: number;
	current_hp: number;
}

export interface S2C_MonsterData {
	id: string;
	x: number;
	y: number;
	type: MonsterType;
	name: string;
	max_hp: number;
	current_hp: number;
}

export interface S2C_InitialStatePayload {
	player_id: string;
	map: S2C_MapData;
	players: S2C_PlayerData[];
	monsters: S2C_MonsterData[];
}

export type S2C_PlayerJoinedPayload = S2C_PlayerData;

export interface S2C_PlayerLeftPayload {
	id: string;
}

export interface S2C_EntityMovedPayload {
	id: string;
	entity_type: typeof EntityTypePlayer | typeof EntityTypeMonster;
	x: number;
	y: number;
}

export type S2C_MonsterSpawnedPayload = S2C_MonsterData;

export interface S2C_EntityRemovedPayload {
	id: string;
	entity_type: typeof EntityTypePlayer | typeof EntityTypeMonster;
}

export interface S2C_CombatInitiatedPayload {
    player_id: string;
    monster_id: string;
    player_x: number;
    player_y: number;
    monster_x: number;
    monster_y: number;
}


// C2S
export const C2S_MessageTypeMove = "move";

// S2C
export const S2C_MessageTypeInitialState = "initial_state";
export const S2C_MessageTypePlayerJoined = "player_joined";
export const S2C_MessageTypePlayerLeft = "player_left";
export const S2C_MessageTypeEntityMoved = "entity_moved";
export const S2C_MessageTypeMonsterSpawned = "monster_spawned";
export const S2C_MessageTypeEntityRemoved = "entity_removed";
export const S2C_MessageTypeCombatInitiated = "combat_initiated";