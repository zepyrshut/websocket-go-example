<script lang="ts">
	import { itemsStore } from '$lib/store';
	import { onMount } from 'svelte';

	let isLoading = true;

	onMount(async () => {
		await getItems();
		isLoading = false;
	});

	let items: { id?: number; name: string; quantity: number }[];

	async function getItems() {
		const response = await fetch('http://localhost:3000/items');
		items = await response.json();
		itemsStore.set(items);
	}



    // websockets getLastItem









	$: items = $itemsStore;
</script>

<div class="items-list" id="items-list">
	<h1>Last items added</h1>

	{#if !isLoading}
		<table>
			<thead>
				<tr>
					<th>ID</th>
					<th>Name</th>
					<th>Quantity</th>
				</tr>
			</thead>
			<tbody>
				{#each items as item}
					<tr>
						<td>{item.id}</td>
						<td>{item.name}</td>
						<td>{item.quantity}</td>
					</tr>
				{/each}
			</tbody>
		</table>
	{:else}
		<p>Loading</p>
	{/if}
</div>
