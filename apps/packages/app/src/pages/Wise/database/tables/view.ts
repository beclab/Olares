import { sendMessageToWorker } from '../sqliteService';
import { FilterInfo } from 'src/utils/rss-types';
import { convertIntegerPropertiesToBoolean } from '../utils';

const booleanProperties = ['showbadge', 'pin', 'system'];

function convertFilterInfo(filters: FilterInfo[]) {
	return filters.map((filter: any) =>
		convertIntegerPropertiesToBoolean(filter, booleanProperties)
	);
}

export function createViewTable(db: any): string {
	db.exec(`
		CREATE TABLE IF NOT EXISTS views (
			id TEXT PRIMARY KEY,
			icon TEXT,
			name TEXT,
			query TEXT,
			description TEXT,
			showbadge INTEGER,
			pin INTEGER,
			system INTEGER,
			serial_no INTEGER,
			created_at TEXT,
			updated_at TEXT,
			sortby TEXT,
			orderby TEXT,
			splitview TEXT
		);
	`);
	return 'views';
}

export async function addOrUpdateFilters(filters: FilterInfo[]) {
	if (filters.length === 0) {
		console.log('View update length 0 return');
		return;
	}
	try {
		const params = filters.map((filter) => [
			filter.id,
			filter.icon,
			filter.name,
			filter.query,
			filter.description,
			filter.showbadge ? 1 : 0,
			filter.pin ? 1 : 0,
			filter.system ? 1 : 0,
			filter.serial_no,
			filter.created_at,
			filter.updated_at,
			filter.sortby,
			filter.orderby,
			filter.splitview
		]);
		await sendMessageToWorker('transaction', {
			sql: `
				INSERT INTO views (id, icon, name, query, description, showbadge, pin, system, serial_no, created_at, updated_at, sortby, orderby, splitview)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
				ON CONFLICT(id) DO UPDATE SET
					icon = excluded.icon,
					name = excluded.name,
					query = excluded.query,
					description = excluded.description,
					showbadge = excluded.showbadge,
					pin = excluded.pin,
					system = excluded.system,
					serial_no = excluded.serial_no,
					created_at = excluded.created_at,
					updated_at = excluded.updated_at,
					sortby = excluded.sortby,
					orderby = excluded.orderby,
					splitview = excluded.splitview;
			`,
			params: params
		});
		console.log('View added/updated successfully!');
	} catch (error) {
		console.error('Failed to add/update filter:', error);
	}
}

export async function getAllFilters(): Promise<FilterInfo[]> {
	try {
		const filters: any = await sendMessageToWorker('query', {
			sql: `SELECT * FROM views;`
		});
		const convertList = convertFilterInfo(filters);
		return convertList.length > 0 ? (convertList as FilterInfo[]) : [];
	} catch (error) {
		console.error('Failed to get all filters:', error);
		return [];
	}
}

export async function getFilterById(id: string): Promise<FilterInfo | null> {
	try {
		const filters: any = await sendMessageToWorker('query', {
			sql: `SELECT * FROM views WHERE id = ?;`,
			params: [id]
		});
		const convertList = convertFilterInfo([filters]);
		return convertList.length > 0 ? (convertList[0] as FilterInfo) : null;
	} catch (error) {
		console.error('Failed to get filter:', error);
		return null;
	}
}

export async function removeFilterById(id: string) {
	try {
		await sendMessageToWorker('execute', {
			sql: `DELETE FROM views WHERE id = ?;`,
			params: [id]
		});
		console.log('View removed successfully!');
	} catch (error) {
		console.error('Failed to remove filter:', error);
	}
}

export async function clearFilters() {
	try {
		await sendMessageToWorker('execute', {
			sql: `DELETE FROM views;`
		});
		console.log('All filters cleared successfully!');
	} catch (error) {
		console.error('Failed to clear filters:', error);
	}
}
