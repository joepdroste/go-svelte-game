<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { websocketService } from '$lib/services/websocketService';
	import { selfId, mapData, players, monsters, initializeGameStoreListeners } from '$lib/stores/gameStore';
	import { C2S_MessageTypeMove } from '$lib/protocol/messages'; // For sending move command
	import { derived } from 'svelte/store';

	import MapGrid from '$lib/components/MapGrid.svelte';
	import PlayerToken from '$lib/components/PlayerToken.svelte';
	import MonsterToken from '$lib/components/MonsterToken.svelte';

	let connectionStatus = '';
	let errorStatus = '';
	let unsubscribers: (() => void)[] = [];

	onMount(() => {
		const unsubIsConnected = websocketService.isConnected.subscribe(value => {
			connectionStatus = value ? 'Connected' : 'Disconnected';
		});
		const unsubLastError = websocketService.lastError.subscribe(value => {
			errorStatus = value || '';
		});
		unsubscribers.push(unsubIsConnected, unsubLastError);
		
		initializeGameStoreListeners();
		websocketService.connect();

		function handleKeyDown(event: KeyboardEvent) {
			let dx = 0;
			let dy = 0;
			switch (event.key) {
				case 'ArrowUp': case 'w': dy = -1; break;
				case 'ArrowDown': case 's': dy = 1; break;
				case 'ArrowLeft': case 'a': dx = -1; break;
				case 'ArrowRight': case 'd': dx = 1; break;
				default: return;
			}
			event.preventDefault();
			if (dx !== 0 || dy !== 0) {
				sendMoveCommand(dx, dy);
			}
		}
		window.addEventListener('keydown', handleKeyDown);
		unsubscribers.push(() => window.removeEventListener('keydown', handleKeyDown));


		return () => {
			console.log('Page unmounting, disconnecting WebSocket.');
			websocketService.disconnect();
			unsubscribers.forEach(unsub => unsub());
		};
	});

	const currentPlayer = derived(
        [selfId, players],
        ([$sId, $p]) => $sId ? $p.get($sId) : null
    );

    function sendMoveCommand(dx: number, dy: number) {
        if ($currentPlayer?.isInCombat) {
            console.log('Cannot move: player is in combat.');
            return;
        }
        websocketService.sendMessage(C2S_MessageTypeMove, { dx, dy });
    }

	function handleAttack() {
        if ($currentPlayer?.isInCombat && $currentPlayer.combatTargetId) {
            console.log(`Player ${$selfId} attacks monster ${$currentPlayer.combatTargetId}! (Not implemented yet)`);
        }
    }
</script>

<main>
	<h1>Game Client</h1>
	<p>Status: {connectionStatus}</p>
	{#if errorStatus}
		<p style="color: red;">Error: {errorStatus}</p>
	{/if}

    <div class="controls">
        {#if !$currentPlayer?.isInCombat}
            <button on:click={() => sendMoveCommand(0, -1)} disabled={$currentPlayer?.isInCombat}>Up (W)</button>
            <div>
                <button on:click={() => sendMoveCommand(-1, 0)} disabled={$currentPlayer?.isInCombat}>Left (A)</button>
                <button on:click={() => sendMoveCommand(1, 0)} disabled={$currentPlayer?.isInCombat}>Right (D)</button>
            </div>
            <button on:click={() => sendMoveCommand(0, 1)} disabled={$currentPlayer?.isInCombat}>Down (S)</button>
        {:else if $currentPlayer?.combatTargetId}
            <p><strong>IN COMBAT!</strong> Target: {$currentPlayer.combatTargetId}</p>
            <button on:click={handleAttack}>Attack {$currentPlayer.combatTargetId}</button>
        {/if}
    </div>

	<h2>Your Player ID: {$selfId || 'N/A'}</h2>

    <div class="game-area">
        <MapGrid map={$mapData}>
            {#if $players && $mapData}
                {#each Array.from($players.values()) as player (player.id)}
                    <PlayerToken player={player} isSelf={player.id === $selfId} />
                {/each}
            {/if}
            {#if $monsters && $mapData}
                {#each Array.from($monsters.values()) as monster (monster.id)}
                    <MonsterToken monster={monster} />
                {/each}
            {/if}
        </MapGrid>
    </div>
</main>

<style>
	main {
		display: flex;
		flex-direction: column;
		align-items: center;
		padding: 1em;
	}
	.controls {
		margin-bottom: 1em;
		display: flex;
		flex-direction: column;
		align-items: center;
	}
	.controls button { margin: 3px; padding: 5px 10px; }
	.controls div { display: flex; justify-content: center; }

    .game-area {
        margin-top: 20px;
    }
</style>