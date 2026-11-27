import Dexie from 'dexie';
import { ProcessingState } from 'src/constant/global';

export interface CollectLogRunRecord {
	runID: string;
	state: ProcessingState;
	path: string;
	error: string;
	createdAt: number;
	updatedAt: number;
}

const MAX_RECORDS = 5;

export class LogRunDatabase extends Dexie {
	runs: Dexie.Table<CollectLogRunRecord, string>;

	constructor() {
		super('LogRunDatabase');
		this.version(1).stores({
			runs: '&runID,createdAt'
		});
		this.runs = this.table('runs');
	}
}

let logRunDB: LogRunDatabase | undefined;

function getLogRunDB(): LogRunDatabase {
	if (!logRunDB) {
		logRunDB = new LogRunDatabase();
	}
	return logRunDB;
}

async function pruneCollectLogRuns(): Promise<void> {
	try {
		const records = await getLogRunDB().runs.toArray();
		if (records.length <= MAX_RECORDS) {
			return;
		}
		records.sort((a, b) => a.createdAt - b.createdAt);
		const removed = records
			.slice(0, records.length - MAX_RECORDS)
			.map((record) => record.runID);
		await getLogRunDB().runs.bulkDelete(removed);
	} catch (error) {
		console.warn('pruneCollectLogRuns failed', error);
	}
}

export async function addCollectLogRun(runID: string): Promise<void> {
	if (!runID) {
		return;
	}
	try {
		const now = Date.now();
		await getLogRunDB().runs.put({
			runID,
			state: ProcessingState.InProgress,
			path: '',
			error: '',
			createdAt: now,
			updatedAt: now
		});
		await pruneCollectLogRuns();
	} catch (error) {
		console.warn('addCollectLogRun failed', error);
	}
}

export async function updateCollectLogRun(
	runID: string,
	payload: { state: ProcessingState; path?: string; error?: string }
): Promise<void> {
	if (!runID) {
		return;
	}
	try {
		await getLogRunDB().runs.update(runID, {
			state: payload.state,
			path: payload.path || '',
			error: payload.error || '',
			updatedAt: Date.now()
		});
	} catch (error) {
		console.warn('updateCollectLogRun failed', error);
	}
}

export async function getCollectLogRuns(): Promise<CollectLogRunRecord[]> {
	try {
		const records = await getLogRunDB().runs.toArray();
		return records.sort((a, b) => b.createdAt - a.createdAt);
	} catch (error) {
		console.warn('getCollectLogRuns failed', error);
		return [];
	}
}

export async function getPendingCollectLogRunId(): Promise<string> {
	try {
		const records = await getCollectLogRuns();
		const pending = records.find(
			(record) =>
				record.state === ProcessingState.InProgress ||
				record.state === ProcessingState.Init
		);
		return pending ? pending.runID : '';
	} catch (error) {
		console.warn('getPendingCollectLogRunId failed', error);
		return '';
	}
}
