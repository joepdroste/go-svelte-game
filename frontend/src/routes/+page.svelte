<script lang="ts">
	import { onMount, onDestroy } from "svelte";
	import { derived } from "svelte/store";

	import { websocketService } from "$lib/services/websocketService";
	import {
		selfId,
		mapData,
		players,
		monsters,
		notifications,
		initializeGameStoreListeners,
		type ClientPlayerData,
	} from "$lib/stores/gameStore";

	import {
		C2S_MessageTypeMove,
		C2S_MessageTypeAttack,
		C2S_MessageTypeUsePotion,
	} from "$lib/protocol/messages";

	import MapGrid from "$lib/components/MapGrid.svelte";
	import PlayerToken from "$lib/components/PlayerToken.svelte";
	import MonsterToken from "$lib/components/MonsterToken.svelte";

	let connectionStatus = "Disconnected";
	let errorStatus = "";
	let unsubscribers: (() => void)[] = [];

	const currentPlayer = derived(
		[selfId, players],
		([$sId, $pMap]): ClientPlayerData | null => {
			if ($sId) {
				return $pMap.get($sId) || null;
			}
			return null;
		},
	);

	onMount(() => {
		const unsubIsConnected = websocketService.isConnected.subscribe(
			(value) => {
				connectionStatus = value ? "Connected" : "Disconnected";
			},
		);
		const unsubLastError = websocketService.lastError.subscribe((value) => {
			errorStatus = value || "";
		});
		unsubscribers.push(unsubIsConnected, unsubLastError);

		initializeGameStoreListeners();
		websocketService.connect();

		function handleKeyDown(event: KeyboardEvent) {
			const $cp = $currentPlayer;

			if (!$cp) return;

			if ($cp.isInCombat) {
				if (event.key === " ") {
					event.preventDefault();
					handleAttack();
				} else if (event.key.toLowerCase() === "h") {
					event.preventDefault();
					handleUsePotion();
				}
				return;
			}

			let dx = 0;
			let dy = 0;
			let usePotionAction = false;

			switch (event.key.toLowerCase()) {
				case "arrowup":
				case "w":
					dy = -1;
					break;
				case "arrowdown":
				case "s":
					dy = 1;
					break;
				case "arrowleft":
				case "a":
					dx = -1;
					break;
				case "arrowright":
				case "d":
					dx = 1;
					break;
				case "h":
					usePotionAction = true;
					break;
				default:
					return;
			}
			event.preventDefault();

			if (usePotionAction) {
				handleUsePotion();
			} else if (dx !== 0 || dy !== 0) {
				sendMoveCommand(dx, dy);
			}
		}
		window.addEventListener("keydown", handleKeyDown);
		unsubscribers.push(() =>
			window.removeEventListener("keydown", handleKeyDown),
		);

		return () => {
			console.log("Page unmounting, disconnecting WebSocket.");
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
			websocketService.sendMessage(C2S_MessageTypeAttack, {
				target_id: $currentPlayer.combatTargetId,
			});
		} else {
			console.log("Cannot attack: not in combat or no target.");
		}
	}

	function handleUsePotion() {
		if (!$currentPlayer || $currentPlayer.current_hp <= 0) {
			console.log("Cannot use potion: player is defeated or not loaded.");
			return;
		}
		console.log(`Player ${$selfId} attempting to use potion.`);
		websocketService.sendMessage(C2S_MessageTypeUsePotion, {});
	}
</script>

<main>
	<h1>Game Client</h1>
	<p>Status: {connectionStatus}</p>
	{#if errorStatus}
		<p style="color: red;">Error: {errorStatus}</p>
	{/if}

	{#if $currentPlayer}
		<div class="player-stats">
			<h3>{$currentPlayer.id} (Level: {$currentPlayer.level})</h3>
			<p>
				HP: {$currentPlayer.current_hp} / {$currentPlayer.max_hp}
				<progress
					class="hp-progress"
					value={$currentPlayer.current_hp}
					max={$currentPlayer.max_hp}
				></progress>
			</p>
			{#if typeof $currentPlayer.xp === "number" && typeof $currentPlayer.xp_to_next_level === "number" && $currentPlayer.xp_to_next_level > 0}
				<p>
					XP: {$currentPlayer.xp} / {$currentPlayer.xp_to_next_level}
					<progress
						class="xp-progress"
						value={$currentPlayer.xp}
						max={$currentPlayer.xp_to_next_level}
					></progress>
				</p>
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
			<button on:click={() => sendMoveCommand(0, -1)} title="Move Up (W)"
				>Up (W)</button
			>
			<div>
				<button
					on:click={() => sendMoveCommand(-1, 0)}
					title="Move Left (A)">Left (A)</button
				>
				<button
					on:click={() => sendMoveCommand(1, 0)}
					title="Move Right (D)">Right (D)</button
				>
			</div>
			<button on:click={() => sendMoveCommand(0, 1)} title="Move Down (S)"
				>Down (S)</button
			>
		{:else if $currentPlayer?.combatTargetId}
			<p>
				<strong>IN COMBAT!</strong> Target: {$currentPlayer.combatTargetId}
			</p>
			<button on:click={handleAttack} title="Attack Target (Spacebar)"
				>Attack Target (Space)</button
			>
		{/if}
	</div>

	<div class="actions">
		<button
			on:click={handleUsePotion}
			disabled={!$currentPlayer ||
				$currentPlayer.current_hp <= 0 ||
				$currentPlayer.current_hp >= $currentPlayer.max_hp}
			title="Use Potion (H)"
		>
			Use Potion (H)
		</button>
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

	{#if $notifications.length > 0}
		<div class="notifications-container">
			{#each $notifications as notification (notification.id)}
				<div
					class="notification notification-{notification.level}"
					role="alert"
				>
					{notification.message}
				</div>
			{/each}
		</div>
	{/if}
</main>

<style>
	main {
		display: flex;
		flex-direction: column;
		align-items: center;
		padding: 1em;
		font-family: sans-serif;
	}
	.player-stats {
		border: 1px solid #ccc;
		padding: 10px;
		margin-bottom: 1em;
		background-color: #f9f9f9;
		min-width: 250px;
		text-align: left;
		border-radius: 5px;
	}
	.player-stats h3 {
		margin-top: 0;
		text-align: center;
	}
	.player-stats p {
		margin: 0.3em 0;
	}
	progress.hp-progress {
		width: 100%;
		height: 10px;
	}
	progress.hp-progress::-webkit-progress-bar {
		background-color: #ef5350;
		border-radius: 3px;
	}
	progress.hp-progress::-webkit-progress-value {
		background-color: #4caf50;
		border-radius: 3px;
	}
	progress.hp-progress::-moz-progress-bar {
		background-color: #4caf50;
	}

	progress.xp-progress {
		width: 100%;
		height: 10px;
	}
	progress.xp-progress::-webkit-progress-bar {
		background-color: #64b5f6;
		border-radius: 3px;
	}
	progress.xp-progress::-webkit-progress-value {
		background-color: #1976d2;
		border-radius: 3px;
	}
	progress.xp-progress::-moz-progress-bar {
		background-color: #1976d2;
	}

	.controls,
	.actions {
		margin-bottom: 1em;
		display: flex;
		flex-direction: column;
		align-items: center;
	}
	.controls button,
	.actions button {
		margin: 3px;
		padding: 8px 15px;
		min-width: 100px;
		cursor: pointer;
	}
	.controls div {
		display: flex;
		justify-content: center;
	}

	.game-area {
		margin-top: 20px;
		border: 2px solid black;
	}

	.notifications-container {
		position: fixed;
		bottom: 20px;
		right: 20px;
		width: 300px;
		z-index: 1000;
		display: flex;
		flex-direction: column-reverse;
	}
	.notification {
		padding: 12px;
		margin-top: 8px;
		border-radius: 4px;
		color: white;
		font-size: 0.9em;
		opacity: 0.95;
		box-shadow: 0 2px 5px rgba(0, 0, 0, 0.2);
	}
	.notification-info {
		background-color: #007bff;
	}
	.notification-success {
		background-color: #28a745;
	}
	.notification-error {
		background-color: #dc3545;
	}
</style>
