// src/lib/stores/gameStore.ts
import { writable, type Writable } from 'svelte/store';
import {
    EntityTypeMonster,
	EntityTypePlayer,
	type S2C_MapData,
	type S2C_PlayerData,
	type S2C_MonsterData,
	type S2C_InitialStatePayload,
	type S2C_PlayerJoinedPayload,
	type S2C_PlayerLeftPayload,
	type S2C_EntityMovedPayload,
	type S2C_CombatInitiatedPayload,
	type S2C_CombatUpdatePayload,
	type S2C_EntityRemovedPayload,
	type S2C_PlayerStatUpdatePayload,
} from '$lib/protocol/messages';
import { websocketService } from '$lib/services/websocketService';
import {
	S2C_MessageTypeInitialState,
	S2C_MessageTypePlayerJoined,
	S2C_MessageTypePlayerLeft,
	S2C_MessageTypeEntityMoved,
	S2C_MessageTypeCombatInitiated,
	S2C_MessageTypeCombatUpdate,
	S2C_MessageTypeEntityRemoved,
	S2C_MessageTypePlayerStatUpdate
} from '$lib/protocol/messages';

export const selfId: Writable<string | null> = writable(null);
export const mapData: Writable<S2C_MapData | null> = writable(null);

export interface ClientPlayerData extends S2C_PlayerData {
    isInCombat?: boolean;
    combatTargetId?: string | null;
    xp?: number;
    xpToNextLevel?: number;
    attack?: number;
    defense?: number;
}

export interface ClientMonsterData extends S2C_MonsterData {
    isInCombat?: boolean;
    combatTargetId?: string | null;
}

export const players: Writable<Map<string, ClientPlayerData>> = writable(new Map());
export const monsters: Writable<Map<string, ClientMonsterData>> = writable(new Map());

export interface ActiveCombatInfo {
    playerId: string;
    monsterId: string;
}

export function initializeGameStoreListeners() {
	// Initial State
	websocketService.onMessage<S2C_InitialStatePayload>(S2C_MessageTypeInitialState, (payload) => {
		console.log('Received Initial State:', payload);
		selfId.set(payload.player_id);
		mapData.set(payload.map);
		
		const newPlayers = new Map<string, ClientPlayerData>();
		payload.players.forEach(p => newPlayers.set(p.id, { ...p, isInCombat: false, combatTargetId: null }));
		players.set(newPlayers);

		const newMonsters = new Map<string, ClientMonsterData>();
		payload.monsters.forEach(m => newMonsters.set(m.id, { ...m, isInCombat: false, combatTargetId: null }));
		monsters.set(newMonsters);
	});

	// Player Joined
	websocketService.onMessage<S2C_PlayerJoinedPayload>(S2C_MessageTypePlayerJoined, (payload) => {
		console.log('Player Joined:', payload);
		players.update(currentPlayers => {
			currentPlayers.set(payload.id, { ...payload, isInCombat: false, combatTargetId: null });
			return new Map(currentPlayers);
		});
	});

	// Player Left
	websocketService.onMessage<S2C_PlayerLeftPayload>(S2C_MessageTypePlayerLeft, (payload) => {
		console.log('Player Left:', payload);
		players.update(currentPlayers => {
			currentPlayers.delete(payload.id);
			return new Map(currentPlayers);
		});
	});

	// Entity Moved
	websocketService.onMessage<S2C_EntityMovedPayload>(S2C_MessageTypeEntityMoved, (payload) => {
		if (payload.entity_type === EntityTypePlayer) {
			players.update(currentPlayers => {
                const p = currentPlayers.get(payload.id);
                if (p) {
                    p.x = payload.x;
                    p.y = payload.y;
                    currentPlayers.set(payload.id, {...p});
                }
                return new Map(currentPlayers);
            });
		} else if (payload.entity_type === EntityTypeMonster) {
			monsters.update(currentMonsters => { 
                const m = currentMonsters.get(payload.id);
                if (m) {
                    m.x = payload.x;
                    m.y = payload.y;
                    currentMonsters.set(payload.id, {...m});
                }
                return new Map(currentMonsters);
            });
		}
	});

	// Combat Initiated
	websocketService.onMessage<S2C_CombatInitiatedPayload>(S2C_MessageTypeCombatInitiated, (payload) => {
		console.log('Combat Initiated:', payload);
		players.update(currentPlayers => {
			const player = currentPlayers.get(payload.player_id);
			if (player) {
				currentPlayers.set(payload.player_id, { ...player, isInCombat: true, combatTargetId: payload.monster_id });
			}
			return new Map(currentPlayers);
		});
		monsters.update(currentMonsters => {
			const monster = currentMonsters.get(payload.monster_id);
			if (monster) {
				currentMonsters.set(payload.monster_id, { ...monster, isInCombat: true, combatTargetId: payload.player_id });
			}
			return new Map(currentMonsters);
		});
	});

	// Combat Update
	websocketService.onMessage<S2C_CombatUpdatePayload>(S2C_MessageTypeCombatUpdate, (payload) => {
		console.log('Combat Update:', payload);
		const defenderIsPlayer = payload.defender_id.startsWith('player-');
		const attackerIsPlayer = payload.attacker_id.startsWith('player-');

		if (defenderIsPlayer) {
			players.update(currentPlayers => {
				const player = currentPlayers.get(payload.defender_id);
				if (player) {
					player.current_hp = payload.defender_current_hp;
					if (payload.is_defender_defeated) {
						console.log(`Player ${player.id} was defeated!`);
						player.isInCombat = false;
						player.combatTargetId = null;
						monsters.update(ms => {
                            const attackerMonster = ms.get(payload.attacker_id);
                            if(attackerMonster && attackerMonster.combatTargetId === player.id){
                                attackerMonster.isInCombat = false;
                                attackerMonster.combatTargetId = null;
                                ms.set(payload.attacker_id, {...attackerMonster});
                            }
                            return new Map(ms);
                        });
					}
					currentPlayers.set(payload.defender_id, { ...player });
				}
				return new Map(currentPlayers);
			});
		} else {
			monsters.update(currentMonsters => {
				const monster = currentMonsters.get(payload.defender_id);
				if (monster) {
					monster.current_hp = payload.defender_current_hp;
					if (payload.is_defender_defeated) {
						console.log(`Monster ${monster.id} was defeated!`);
						players.update(ps => {
                            const attackerPlayer = ps.get(payload.attacker_id);
                            if(attackerPlayer && attackerPlayer.combatTargetId === monster.id){
                                attackerPlayer.isInCombat = false;
                                attackerPlayer.combatTargetId = null;
                                ps.set(payload.attacker_id, {...attackerPlayer});
                            }
                            return new Map(ps);
                        });
					}
					currentMonsters.set(payload.defender_id, { ...monster });
				}
				return new Map(currentMonsters);
			});
		}
	});

	// Entity Removed
	websocketService.onMessage<S2C_EntityRemovedPayload>(S2C_MessageTypeEntityRemoved, (payload) => {
		console.log('Entity Removed:', payload);
		if (payload.entity_type === EntityTypeMonster) {
			const removedMonsterId = payload.id;
			monsters.update(currentMonsters => {
				currentMonsters.delete(removedMonsterId);
				return new Map(currentMonsters);
			});

			players.update(currentPlayers => {
				let changed = false;
				currentPlayers.forEach(p => {
					if (p.combatTargetId === removedMonsterId) {
						p.isInCombat = false;
						p.combatTargetId = null;
						currentPlayers.set(p.id, {...p});
						changed = true;
					}
				});
				return changed ? new Map(currentPlayers) : currentPlayers;
			});
		}
	});
	
	// Player Stat Update
	websocketService.onMessage<S2C_PlayerStatUpdatePayload>(S2C_MessageTypePlayerStatUpdate, (payload) => {
		console.log('Player Stat Update:', payload);
		players.update(currentPlayers => {
			const player = currentPlayers.get(payload.player_id);
			if (player) {
				player.level = payload.level;
				player.xp = payload.xp;
				player.xp_to_next_level = payload.xp_to_next_level;
				player.max_hp = payload.max_hp;
				player.current_hp = payload.current_hp;
				player.attack = payload.attack;
				player.defense = payload.defense;
				
				currentPlayers.set(payload.player_id, { ...player }); 
			}
			return new Map(currentPlayers);
		});
	});
}