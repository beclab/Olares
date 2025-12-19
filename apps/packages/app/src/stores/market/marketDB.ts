import Dexie, { Table } from 'dexie';

const pendingUpdates = new Map<string, any>();
let mergeTimer: any = null;
let isProcessing = false;

interface SingleMapRecord {
	id: 'appFullInfoMap';
	mapData: [string, any][];
}

export class MarketDatabase extends Dexie {
	singleMapStore: Table<SingleMapRecord, 'appFullInfoMap'>;

	constructor() {
		super('Market');
		this.version(1).stores({
			singleMapStore: 'id'
		});
		this.singleMapStore = this.table('singleMapStore');
	}

	private static _instance: MarketDatabase;
	public static get instance(): MarketDatabase {
		if (!this._instance) {
			this._instance = new MarketDatabase();
		}
		return this._instance;
	}
}

const db = MarketDatabase.instance;

export async function saveAppFullInfoMap(
	mapData: Map<string, any>
): Promise<void> {
	const mapArray = Array.from(mapData.entries());

	try {
		await db.singleMapStore.put({
			id: 'appFullInfoMap',
			mapData: mapArray
		});
	} catch (error) {
		console.log(mapArray);
		console.error('MarketDB app full info save failure:', error);
	}
}

export async function loadAppFullInfoMap(): Promise<
	Map<string, any> | null | undefined
> {
	try {
		const record = await db.singleMapStore.get('appFullInfoMap');

		if (record) {
			return new Map<string, any>(record.mapData);
		} else {
			return null;
		}
	} catch (error) {
		console.error(
			`MarketDB app full info load failure: ${(error as Error).message}`
		);
	}
}

export async function updateAppFullInfoMapKey(
	key: string,
	value: any
): Promise<void> {
	pendingUpdates.set(key, value);

	if (!mergeTimer) {
		mergeTimer = setTimeout(processPendingUpdates, 100);
	}

	return new Promise((resolve) => {
		mergeTimer = setTimeout(resolve, 500);
	});
}

async function processPendingUpdates() {
	if (isProcessing) return;

	isProcessing = true;
	mergeTimer = null;

	const updates = Array.from(pendingUpdates.entries());
	try {
		pendingUpdates.clear();

		if (updates.length === 0) {
			isProcessing = false;
			return;
		}

		const currentMap = (await loadAppFullInfoMap()) || new Map();

		updates.forEach(([key, value]) => {
			currentMap.set(key, value);
		});

		await saveAppFullInfoMap(currentMap);
	} catch (error) {
		console.error(`MarketDB batch update failure: `, error);
		updates.forEach(([key, value]) => {
			pendingUpdates.set(key, value);
		});
	} finally {
		isProcessing = false;

		if (pendingUpdates.size > 0) {
			processPendingUpdates();
		}
	}
}

export async function deleteAppFullInfoMapKey(key: string): Promise<void> {
	try {
		const currentMap = await loadAppFullInfoMap();

		if (currentMap && currentMap.has(key)) {
			currentMap.delete(key);
			await saveAppFullInfoMap(currentMap);
		} else {
			console.log(`MarketDB app full info delete, key : "${key}" not found`);
		}
	} catch (error) {
		console.error(
			`MarketDB app full info delete failure, key : "${key}" reason : `,
			error
		);
	}
}

export async function clearAppFullInfoMap(): Promise<void> {
	try {
		await db.singleMapStore.delete('appFullInfoMap');
		console.log('MarketDB app full info map has been cleared');
	} catch (error) {
		console.error('MarketDB clear app full info map failure:', error);
	}
}
