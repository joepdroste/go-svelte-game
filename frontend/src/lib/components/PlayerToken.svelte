<script lang="ts">
	import type { ClientPlayerData } from '$lib/stores/gameStore';
	export let player: ClientPlayerData;
	export let isSelf: boolean = false;
	const TILE_SIZE = 20;
	$: leftPosition = player.x * TILE_SIZE;
	$: topPosition = player.y * TILE_SIZE;
    $: hpPercentage = player.max_hp > 0 ? (player.current_hp / player.max_hp) * 100 : 0;
</script>

<div
	class="entity player-token"
	class:self={isSelf}
	class:in-combat={player.isInCombat}
	style:left="{leftPosition}px"
	style:top="{topPosition}px"
	title="Player {player.id} (Lvl {player.level}) HP: {player.current_hp}/{player.max_hp}{player.isInCombat ? ' (VS ' + player.combatTargetId + ')' : ''}"
>
	P
    <div class="hp-bar-token-container">
        <div class="hp-bar-token-filled" style:width="{hpPercentage}%"></div>
    </div>
</div>

<style>
	.entity {
		width: 18px;
		height: 18px;
		display: flex;
		align-items: center;
		justify-content: center;
		font-weight: bold;
		border-radius: 3px;
		position: absolute;
		transition: left 0.1s linear, top 0.1s linear;
        user-select: none;
	}

	.player-token {
		background-color: blue;
		color: white;
		border: 1px solid darkblue;
	}

	.player-token.self {
		background-color: lightblue;
		color: black;
		border: 1px solid blue;
		z-index: 10;
	}

	.player-token.in-combat {
        border: 2px solid orange;
        box-shadow: 0 0 5px orange;
    }

	.hp-bar-token-container {
        position: absolute;
        bottom: -7px;
        left: 0;
        width: 100%;
        height: 5px;
        background-color: rgba(0, 0, 0, 0.2);
        border-radius: 2px;
    }
    .hp-bar-token-filled {
        height: 100%;
        background-color: limegreen;
        border-radius: 2px;
        transition: width 0.3s ease;
    }
    .player-token.self .hp-bar-token-filled {
        background-color: green;
    }
</style>