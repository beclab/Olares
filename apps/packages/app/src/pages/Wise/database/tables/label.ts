import { sendMessageToWorker } from '../sqliteService';
import { Label } from '../../../../utils/rss-types';
import { convertIntegerPropertiesToBoolean } from '../utils';

const booleanProperties = ['deleted'];

function convertLabel(labels: Label[]) {
	return labels.map((filter: any) =>
		convertIntegerPropertiesToBoolean(filter, booleanProperties)
	);
}

export function createLabelTable(db: any): string {
	db.exec(`
		CREATE TABLE IF NOT EXISTS labels (
			id TEXT PRIMARY KEY,
			name TEXT,
			entries TEXT,
			notes TEXT,
			deleted INTEGER,
			updated_at TEXT
		);
	`);
	return 'labels';
}

export async function addOrUpdateLabels(labels: Label[]) {
	if (labels.length === 0) {
		console.log('Label update length 0 return');
		return;
	}
	try {
		const params = labels.map((label) => [
			label.id,
			label.name,
			JSON.stringify(label.entries),
			JSON.stringify(label.notes),
			label.deleted ? 1 : 0,
			label.updated_at
		]);
		await sendMessageToWorker('transaction', {
			sql: `
				INSERT INTO labels (id, name, entries, notes, deleted, updated_at)
				VALUES (?, ?, ?, ?, ?, ?) ON CONFLICT(id) DO
				UPDATE SET
					name = excluded.name,
					entries = excluded.entries,
					notes = excluded.notes,
					deleted = excluded.deleted,
					updated_at = excluded.updated_at;
			`,
			params: params
		});
		console.log('Label added/updated successfully!');
	} catch (error) {
		console.error('Failed to add/update label:', error);
	}
}

export async function getAllLabels(): Promise<Label[]> {
	try {
		const labels: any = await sendMessageToWorker('query', {
			sql: `SELECT * FROM labels;`
		});
		const convertList = convertLabel(labels);
		return convertList.length > 0 ? (convertList as Label[]) : [];
	} catch (error) {
		console.error('Failed to get all labels:', error);
		return [];
	}
}

export async function getLabelById(id: string): Promise<Label | null> {
	try {
		const labels: any = await sendMessageToWorker('query', {
			sql: `SELECT * FROM labels WHERE id = ?;`,
			params: [id]
		});
		const convertList = convertLabel(labels);
		return convertList.length > 0 ? (convertList[0] as Label) : null;
	} catch (error) {
		console.error('Failed to get label:', error);
		return null;
	}
}

export async function removeLabelById(id: string) {
	try {
		await sendMessageToWorker('execute', {
			sql: `DELETE FROM labels WHERE id = ?;`,
			params: [id]
		});
		console.log('Label removed successfully!');
	} catch (error) {
		console.error('Failed to remove label:', error);
	}
}

export async function clearLabels() {
	try {
		await sendMessageToWorker('execute', {
			sql: `DELETE
						FROM labels`
		});
		console.log('Label clear successfully!');
	} catch (error) {
		console.error('Failed to clear label:', error);
	}
}
