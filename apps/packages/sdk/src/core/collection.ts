import { Serializable } from './encoding';
import { VaultItem } from './item';

/**
 * A collection of vault items items, used for consolidating changes made independently
 * across multiple instances through "merging".
 */
export class VaultItemCollection
	extends Serializable
	implements Iterable<VaultItem>
{
	/** Number of items in this VaultItemCollection */
	get size() {
		return this._items.size;
	}

	get hasChanges() {
		return !!this._changes.size;
	}

	/** Aggregated list of tags assigned to the items in this collection */
	get tags(): string[] {
		const tags = new Set<string>();
		for (const r of this) {
			for (const t of r.tags) {
				tags.add(t);
			}
		}
		return [...tags];
	}

	private _items: Map<string, VaultItem>;
	private _changes = new Map<string, Date>();

	constructor(items: VaultItem[] = []) {
		super();
		this._items = new Map(
			items.map((item) => [item.id, item] as [string, VaultItem])
		);
	}

	/** Get an item with a given `id` */
	get(id: string) {
		return this._items.get(id) || null;
	}

	/**
	 * Updates one or more items based on their id. If no item with the same id
	 * exists, the item will be added to the collection
	 */
	update(...items: VaultItem[]) {
		for (const item of items) {
			item.updated = new Date();
			this._items.set(item.id, item);
			this._changes.set(item.id, item.updated);
		}
	}

	/**
	 * Removes one or more items based on their id.
	 */
	remove(...items: VaultItem[]) {
		for (const item of items) {
			this._items.delete(item.id);
			this._changes.set(item.id, new Date());
		}
	}

	/**
	 * Merges in changes from another [[VaultItemCollection]] instance.
	 */
	merge(coll: VaultItemCollection) {
		// Delete any items from this collection that don't
		// exist in the other collection and haven't been changed recently
		for (const item of this) {
			if (!this._changes.has(item.id) && !coll.get(item.id)) {
				this._items.delete(item.id);
			}
		}

		// Get changes items from other collection (but only if they haven't recently changed locally)
		for (const item of coll) {
			if (!this._changes.has(item.id)) {
				this._items.set(item.id, item);
			}
		}
	}

	clearChanges(before?: Date) {
		for (const [id, changed] of this._changes.entries()) {
			if (!before || changed <= before) {
				this._changes.delete(id);
			}
		}
	}

	protected _toRaw(version: string) {
		return {
			items: Array.from(this).map((item) => item.toRaw(version)),
			changes: [...this._changes]
		};
	}

	protected _fromRaw({ items, changes }: any) {
		this._items = new Map(
			items.map(
				(item: any) =>
					[item.id, new VaultItem().fromRaw(item)] as [
						string,
						VaultItem
					]
			)
		);
		this._changes = new Map<string, Date>(
			changes &&
				changes.map(([id, date]: [string, string]) => [
					id,
					new Date(date)
				])
		);
	}

	[Symbol.iterator]() {
		return this._items.values();
	}
}
