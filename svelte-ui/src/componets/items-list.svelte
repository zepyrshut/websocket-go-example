<script lang="ts">
	import { browser } from '$app/environment';
	import { itemsStore } from '$lib/store';
	import { onMount } from 'svelte';

	let isLoading:boolean = true;
    let ws: WebSocket;

    if (browser) {
        ws = new WebSocket('ws://localhost:3000/ws/last-item');

        ws.onopen = () => {
            console.log('ws opened');
        };

        ws.onmessage = (event) => {
            const item = JSON.parse(event.data);
            itemsStore.update((items) => {
                items.unshift(item);
                return items;
            });
        };
    }

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

//    const ws = new WebSocket('ws://localhost:3000/ws/last-item');
//     ws.onopen = () => {
//         console.log('ws opened');
//     };

//     ws.onclose = () => {
//         console.log('ws closed');
//     };

    // ws.onmessage = (event) => {
    //     console.log('ws message', event.data);
    //     const item = JSON.parse(event.data);
    //     itemsStore.update((items) => {
    //         items.unshift(item);
    //         return items;
    //     });
    // };








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
