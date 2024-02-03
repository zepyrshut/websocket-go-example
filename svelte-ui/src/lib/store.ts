import { writable } from "svelte/store";
import type { Item } from "../models/interfaces";

export const itemsStore = writable<Item[]>([]);