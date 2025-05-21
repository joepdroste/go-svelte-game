<script lang="ts">
	import type { S2C_MapData } from '$lib/protocol/messages';
	import Tile from './Tile.svelte';

	export let map: S2C_MapData | null;
</script>

{#if map}
	<div
		class="map-grid-container"
		style:grid-template-columns="repeat({map.width}, auto)"
		style:width="{map.width * 20}px" 
        style:height="{map.height * 20}px"
	>
		{#each map.tiles as row, y}
			{#each row as tileData, x}
				<Tile type={tileData.type} x={x} y={y} />
			{/each}
		{/each}
        <slot></slot>
	</div>
{:else}
	<p>Map data not available...</p>
{/if}

<style>
	.map-grid-container {
		display: grid;
		border: 1px solid #ccc;
		position: relative;
		background-color: #e0e0e0;
        user-select: none; 
        -webkit-user-select: none;
        -moz-user-select: none;
        -ms-user-select: none;
	}
</style>