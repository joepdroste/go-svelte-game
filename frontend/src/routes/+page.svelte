<script lang="ts">
	import { onMount, onDestroy } from "svelte";
	import { websocketService } from "$lib/services/websocketService";
	import {
		selfId,
		mapData,
		players,
		monsters,
		initializeGameStoreListeners,
		type ClientPlayerData,
	} from "$lib/stores/gameStore";
	import {
		C2S_MessageTypeMove,
		C2S_MessageTypeAttack,
	} from "$lib/protocol/messages";

	import MapGrid from "$lib/components/MapGrid.svelte";
	import PlayerToken from "$lib/components/PlayerToken.svelte";
	import MonsterToken from "$lib/components/MonsterToken.svelte";
	import { derived } from "svelte/store";

	let connectionStatus = "";
	let errorStatus = "";
	let unsubscribers: (() => void)[] = [];

	const currentPlayer = derived(
		[selfId, players],
		([$sId, $pMap]): ClientPlayerData | null =>
			$sId ? $pMap.get($sId) || null : null,
	);

	onMount(() => {
		const unsubIsConnected = websocketService.isConnected.subscribe(
			(v) => (connectionStatus = v ? "Connected" : "Disconnected"),
		);
		const unsubLastError = websocketService.lastError.subscribe(
			(v) => (errorStatus = v || ""),
		);
		unsubscribers.push(unsubIsConnected, unsubLastError);

		initializeGameStoreListeners();
		websocketService.connect();

		function handleKeyDown(event: KeyboardEvent) {
			if ($currentPlayer?.isInCombat && event.key !== " ") return;

			let dx = 0;
			let dy = 0;
			let attackAction = false;

			switch (event.key) {
				case "ArrowUp":
				case "w":
					dy = -1;
					break;
				case "ArrowDown":
				case "s":
					dy = 1;
					break;
				case "ArrowLeft":
				case "a":
					dx = -1;
					break;
				case "ArrowRight":
				case "d":
					dx = 1;
					break;
				case " ":
					if ($currentPlayer?.isInCombat) {
						attackAction = true;
					}
					break;
				default:
					return;
			}
			event.preventDefault();

			if (attackAction) {
				handleAttack();
			} else if (dx !== 0 || dy !== 0) {
				sendMoveCommand(dx, dy);
			}
		}
		window.addEventListener("keydown", handleKeyDown);
		unsubscribers.push(() =>
			window.removeEventListener("keydown", handleKeyDown),
		);

		return () => {
			websocketService.disconnect();
			unsubscribers.forEach((unsub) => unsub());
		};
	});

	function sendMoveCommand(dx: number, dy: number) {
		if ($currentPlayer?.isInCombat) {
			console.log("Cannot move: player is in combat.");
			return;
		}
		websocketService.sendMessage(C2S_MessageTypeMove, { dx, dy });
	}

	function handleAttack() {
		if ($currentPlayer?.isInCombat && $currentPlayer.combatTargetId) {
			console.log(
				`Player ${$selfId} sending attack command for monster ${$currentPlayer.combatTargetId}`,
			);
			websocketService.sendMessage(C2S_MessageTypeAttack, {
				target_id: $currentPlayer.combatTargetId,
			});
		} else {
			console.log("Cannot attack: not in combat or no target.");
		}
	}
</script>

<main>
	<h1>Game Client</h1>
	<p>Status: {connectionStatus}</p>
	{#if errorStatus}<p style="color: red;">Error: {errorStatus}</p>{/if}

	{#if $currentPlayer}
		<div class="player-stats">
			<h3>{$currentPlayer.id} (Level: {$currentPlayer.level})</h3>
			<p>HP: {$currentPlayer.current_hp} / {$currentPlayer.max_hp}</p>
			{#if typeof $currentPlayer.xp === "number" && typeof $currentPlayer.xp_to_next_level === "number" && $currentPlayer.xp_to_next_level > 0}
				<p>
					XP: {$currentPlayer.xp} / {$currentPlayer.xp_to_next_level}
				</p>
				<div class="xp-bar-container">
					<div
						class="xp-bar-filled"
						style:width="{($currentPlayer.xp /
							$currentPlayer.xp_to_next_level) *
							100}%"
					></div>
				</div>
			{:else if typeof $currentPlayer.xp === "number"}
				<p>XP: {$currentPlayer.xp}</p>
			{/if}
			{#if typeof $currentPlayer.attack === "number"}
				<p>Attack: {$currentPlayer.attack}</p>
			{/if}
			{#if typeof $currentPlayer.defense === "number"}
				<p>Defense: {$currentPlayer.defense}</p>
			{/if}
		</div>
	{/if}

	<div class="controls">
		{#if !$currentPlayer?.isInCombat}
			<button on:click={() => sendMoveCommand(0, -1)}>Up (W)</button>
			<div>
				<button on:click={() => sendMoveCommand(-1, 0)}>Left (A)</button
				>
				<button on:click={() => sendMoveCommand(1, 0)}>Right (D)</button
				>
			</div>
			<button on:click={() => sendMoveCommand(0, 1)}>Down (S)</button>
		{:else if $currentPlayer?.combatTargetId}
			<p>
				<strong>IN COMBAT!</strong> Target: {$currentPlayer.combatTargetId}
			</p>
			<button on:click={handleAttack}>Attack Target (Space)</button>
		{/if}
	</div>

	<h2>Your Player ID: {$selfId || "N/A"}</h2>

	<div class="game-area">
		<MapGrid map={$mapData}>
			{#if $players && $mapData}
				{#each Array.from($players.values()) as player (player.id)}
					<PlayerToken {player} isSelf={player.id === $selfId} />
				{/each}
			{/if}
			{#if $monsters && $mapData}
				{#each Array.from($monsters.values()) as monster (monster.id)}
					<MonsterToken {monster} />
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
	.controls button {
		margin: 3px;
		padding: 5px 10px;
	}
	.controls div {
		display: flex;
		justify-content: center;
	}
	.game-area {
		margin-top: 20px;
	}
	.player-stats {
		border: 1px solid #ccc;
		padding: 10px;
		margin-bottom: 10px;
		background-color: #f9f9f9;
		min-width: 220px;
		text-align: center;
	}
	.xp-bar-container {
		width: 100%;
		height: 20px;
		background-color: #e0e0e0;
		border-radius: 5px;
		overflow: hidden;
		margin-top: 5px;
	}
	.xp-bar-filled {
		height: 100%;
		background-color: #4caf50;
		transition: width 0.5s ease-out;
		text-align: center;
		color: white;
		font-size: 0.8em;
		line-height: 20px;
	}
</style>
