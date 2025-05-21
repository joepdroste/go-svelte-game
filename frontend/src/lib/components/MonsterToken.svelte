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
</script>

<div
    class="entity monster-token"
    class:in-combat={monster.isInCombat}
    style:left="{leftPosition}px"
    style:top="{topPosition}px"
    title="{monster.name} ({monster.id}) HP: {monster.current_hp}/{monster.max_hp}{monster.isInCombat ? ' (IN COMBAT with ' + monster.combatTargetId + ')' : ''}"
>
    {monsterSymbol}
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

	.monster-token {
		background-color: red;
		color: white;
		border: 1px solid darkred;
	}

	.monster-token.in-combat {
        border: 2px solid orange;
        box-shadow: 0 0 5px orange;
    }
</style>