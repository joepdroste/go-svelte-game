<script lang="ts">
	import { TileType } from '$lib/protocol/messages';

	export let type: TileType;
	export let x: number;
	export let y: number;

	$: tileSymbol = (() => {
		switch (type) {
			case TileType.Grass:
				return '.';
			case TileType.Stone:
				return '#';
			default:
				return '?';
		}
	})();

	$: tileClass = (() => {
		switch (type) {
			case TileType.Grass:
				return 'grass';
			case TileType.Stone:
				return 'stone';
			default:
				return 'unknown';
		}
	})();
</script>

<div class="tile {tileClass}" title={`(${x},${y}) - ${TileType[type]}`}>
	{tileSymbol}
</div>

<style>
	.tile {
		width: 20px;
		height: 20px;
		display: flex;
		align-items: center;
		justify-content: center;
		font-family: monospace;
		line-height: 1;
	}
	.grass {
		background-color: #90ee90;
		color: #333;
	}
	.stone {
		background-color: #808080;
		color: #fff;
	}
	.unknown {
		background-color: #ff00ff;
	}
</style>