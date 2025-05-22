<script lang="ts">
	import type { ClientMonsterData } from '$lib/stores/gameStore';
	import { MonsterType } from '$lib/protocol/messages';
	export let monster: ClientMonsterData;
	const TILE_SIZE = 20;
	$: leftPosition = monster.x * TILE_SIZE;
	$: topPosition = monster.y * TILE_SIZE;
	$: monsterSymbol = (() => {
		switch (monster.type) {
			case MonsterType.Goblin: return 'g';
			case MonsterType.Orc: return 'O';
			default: return 'M';
		}
	})();
    $: hpPercentage = monster.max_hp > 0 ? (monster.current_hp / monster.max_hp) * 100 : 0;
</script>

<div
	class="entity monster-token"
    class:in-combat={monster.isInCombat}
	style:left="{leftPosition}px"
	style:top="{topPosition}px"
	title="{monster.name} ({monster.id}) HP: {monster.current_hp}/{monster.max_hp}{monster.isInCombat ? ' (VS ' + monster.combatTargetId + ')' : ''}"
>
	{monsterSymbol}
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
        flex-direction: column; 
        justify-content: center;
        padding-top: 2px;
	}

	.monster-token {
		background-color: #E53E3E;
		color: white;
		border: 1px solid #C53030;
	}

    .monster-token.in-combat {
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
        background-color: #c53030;
        border-radius: 2px;
        transition: width 0.3s ease;
    }

    .monster-token {
        font-size: 0.9em;
    }
</style>