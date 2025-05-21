// src/lib/stores/gameStore.ts
import { writable, type Writable } from 'svelte/store';
import {
    EntityTypeMonster,
	type S2C_MapData,
	type S2C_PlayerData,
	type S2C_MonsterData,
	type S2C_InitialStatePayload,
	type S2C_PlayerJoinedPayload,
	type S2C_PlayerLeftPayload,
	type S2C_EntityMovedPayload,
	EntityTypePlayer,
	type S2C_CombatInitiatedPayload,
} from '$lib/protocol/messages';
import { websocketService } from '$lib/services/websocketService';
import {
	S2C_MessageTypeInitialState,
	S2C_MessageTypePlayerJoined,
	S2C_MessageTypePlayerLeft,
	S2C_MessageTypeEntityMoved,
	S2C_MessageTypeCombatInitiated,
} from '$lib/protocol/messages';

export const selfId: Writable<string | null> = writable(null);
export const mapData: Writable<S2C_MapData | null> = writable(null);

export interface ClientPlayerData extends S2C_PlayerData {
    isInCombat?: boolean; // Optional client-side flag
    combatTargetId?: string | null;
}
export interface ClientMonsterData extends S2C_MonsterData {
    isInCombat?: boolean; // Optional client-side flag
    combatTargetId?: string | null;
}

export const players: Writable<Map<string, ClientPlayerData>> = writable(new Map());
export const monsters: Writable<Map<string, ClientMonsterData>> = writable(new Map());

export interface ActiveCombatInfo {
    playerId: string;
    monsterId: string;
}

export function initializeGameStoreListeners() {
	websocketService.onMessage<S2C_InitialStatePayload>(S2C_MessageTypeInitialState, (payload) => {
		console.log('Received Initial State:', payload);
		selfId.set(payload.player_id);
		mapData.set(payload.map);
		
		const newPlayers = new Map<string, S2C_PlayerData>();
		payload.players.forEach(p => newPlayers.set(p.id, p));
		players.set(newPlayers);

		const newMonsters = new Map<string, S2C_MonsterData>();
		payload.monsters.forEach(m => newMonsters.set(m.id, m));
		monsters.set(newMonsters);
	});

	websocketService.onMessage<S2C_PlayerJoinedPayload>(S2C_MessageTypePlayerJoined, (payload) => {
		console.log('Player Joined:', payload);
		players.update(currentPlayers => {
			currentPlayers.set(payload.id, payload);
			return new Map(currentPlayers);
		});
	});

	websocketService.onMessage<S2C_PlayerLeftPayload>(S2C_MessageTypePlayerLeft, (payload) => {
		console.log('Player Left:', payload);
		players.update(currentPlayers => {
			currentPlayers.delete(payload.id);
			return new Map(currentPlayers);
		});
	});

	websocketService.onMessage<S2C_EntityMovedPayload>(S2C_MessageTypeEntityMoved, (payload) => {
		if (payload.entity_type === EntityTypePlayer) {
			players.update(currentPlayers => {
				const player = currentPlayers.get(payload.id);
				if (player) {
					player.x = payload.x;
					player.y = payload.y;
					currentPlayers.set(payload.id, { ...player });
				}
				return new Map(currentPlayers);
			});
		} else if (payload.entity_type === EntityTypeMonster) {
			monsters.update(currentMonsters => {
				const monster = currentMonsters.get(payload.id);
				if (monster) {
					monster.x = payload.x;
					monster.y = payload.y;
					currentMonsters.set(payload.id, { ...monster });
				}
				return new Map(currentMonsters);
			});
		}
	});

	websocketService.onMessage<S2C_CombatInitiatedPayload>(S2C_MessageTypeCombatInitiated, (payload) => {
		console.log('Combat Initiated:', payload);

		// Update player state
		players.update(currentPlayers => {
			const player = currentPlayers.get(payload.player_id);
			if (player) {
				currentPlayers.set(payload.player_id, { 
					...player, 
					isInCombat: true, 
					combatTargetId: payload.monster_id 
				});
			}
			return new Map(currentPlayers); // Ensure reactivity
		});

		// Update monster state
		monsters.update(currentMonsters => {
			const monster = currentMonsters.get(payload.monster_id);
			if (monster) {
				currentMonsters.set(payload.monster_id, { 
					...monster, 
					isInCombat: true, 
					combatTargetId: payload.player_id 
				});
			}
			return new Map(currentMonsters); // Ensure reactivity
		});

		// TODO: Later, handle S2C_CombatEnded message to clear these flags
	});

    // TODO: Add listeners for S2C_MessageTypeMonsterSpawned, S2C_MessageTypeEntityRemoved
}