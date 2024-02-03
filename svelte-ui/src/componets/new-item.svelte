<script lang="ts">
	import { itemsStore } from '$lib/store';
	import type { Item } from '../models/interfaces';

	const newItem: Item = {
		name: '',
		quantity: 0
	};

	async function addItem() {
		console.log(newItem);
		const response = await fetch('http://localhost:3000/new-item', {
			method: 'POST',
			headers: {
				'Content-Type': 'application/json'
			},
			body: JSON.stringify(newItem)
		});
		return response.json();
	}

	function fakeAddItem() {
		const newItem = { id: 0, name: '', quantity: 0 };

		newItem.id = $itemsStore.length + 1;
		newItem.name = `Item ${newItem.id}`;
		newItem.quantity = Math.floor(Math.random() * 100);
		//itemsStore.update((items) => [...items, newItem]);

		itemsStore.update((items) => {
			items.unshift(newItem);
			return items;
		});
	}
</script>

<h1>New item form</h1>
<form on:submit|preventDefault={addItem}>
	<label for="name">Name</label>
	<input type="text" id="name" required bind:value={newItem.name} />
	<label for="quantity">Quantity</label>
	<input type="number" id="quantity" required bind:value={newItem.quantity} />
	<button type="submit">Add</button>
</form>

<hr>


<p> dev zone </p>

<button on:click={fakeAddItem}>Fake add item</button>
