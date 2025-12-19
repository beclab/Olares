import sqlite3InitModule from '@sqlite.org/sqlite-wasm';
import { createEntryTable } from '../tables/entry';
import { createFeedTable } from '../tables/feed';
import { createLabelTable } from '../tables/label';
import { createNoteTable } from '../tables/note';
import { createViewTable } from '../tables/view';

export class DBManager {
	private db: any | null = null;
	private poolUtil: any | null = null;
	private isInitializing = false;
	private isClosing = false;
	private tableSet: Set<string>;
	inited = false;
	events = new Map();

	async initDB(): Promise<string> {
		if (this.db) {
			console.warn('Database is already initialized.');
			return 'Database is already initialized.';
		}

		if (this.isInitializing) {
			console.warn('Database is initializing, please wait.');
			return 'Database is initializing, please wait.';
		}

		this.isInitializing = true;

		try {
			const sqlite3 = await sqlite3InitModule();
			console.log('Running SQLite3 version', sqlite3.version.libVersion);

			const opfsVfs = sqlite3.capi.sqlite3_vfs_find('opfs');
			if (!opfsVfs) {
				console.log('OPFS VFS is not available in this environment.');
			}

			this.poolUtil = await sqlite3.installOpfsSAHPoolVfs({
				clearOnInit: false,
				directory: '/sqlite3'
			});

			this.db = new this.poolUtil.OpfsSAHPoolDb(`/lpWise.db`);
			console.log(
				opfsVfs
					? `Database OPFS is available, created persisted database at ${this.db.filename}`
					: `Database OPFS is not available, created transient database ${this.db.filename}`
			);

			this.createTables();
			this.inited = true;
			this.emit('init');

			return 'Database initialized with OPFS.';
		} catch (error) {
			console.error('Database initialization failed:', error);
			throw new Error('Database initialization failed.');
		} finally {
			this.isInitializing = false;
		}
	}

	async closeDB(): Promise<void> {
		if (!this.db) {
			console.warn('Database is not initialized. Skipping close.');
			return;
		}

		if (this.isClosing) {
			console.warn('Database is already closing. Please wait.');
			return;
		}

		this.isClosing = true;

		try {
			this.db.exec('VACUUM;');
			this.db.close();
			this.db = null;
			this.inited = false;

			if (this.poolUtil) {
				await this.poolUtil.removeVfs();
				this.poolUtil = null;
			}

			console.log('Database closed successfully.');
		} catch (error) {
			console.error('Database closing failed:', error);
			throw new Error('Database closing failed.');
		} finally {
			this.isClosing = false;
		}
	}

	on(eventName: string, callback: any) {
		if (!this.events.has(eventName)) {
			this.events.set(eventName, new Set());
		}
		this.events.get(eventName).add(callback);
	}

	emit(eventName: string, ...args: any) {
		const callbacks = this.events.get(eventName);
		if (callbacks) {
			callbacks.forEach((callback) => callback(...args));
		}
	}

	execute(sql: string, params: any[] = []): void {
		if (!this.db) throw new Error('Database is not initialized.');
		this.db.exec({ sql, bind: params });
	}

	executeTransaction(sql: string, params: any[]): void {
		if (!this.db) throw new Error('Database is not initialized.');
		this.db.exec('BEGIN TRANSACTION');
		let stmt;
		try {
			stmt = this.db.prepare(sql);
			for (const row of params) {
				stmt.bind(row);
				stmt.step();
				stmt.reset();
			}

			this.db.exec('COMMIT');
			console.log('Database batch insert with conflict handling completed!');
		} catch (err) {
			console.error(
				'Database error during batch insert with conflict handling:',
				err
			);
			this.db.exec('ROLLBACK');
		} finally {
			stmt?.finalize();
		}
	}

	getTablesFromQuery(sql: string) {
		if (!this.db) throw new Error('is not initialized.');
		let stmt;
		const matchTables = new Set();

		try {
			stmt = this.db.prepare(`EXPLAIN QUERY PLAN ${sql}`);
			while (stmt.step()) {
				const row = stmt.get({});
				const detail = row.detail;

				console.log('getTables details ', detail);
				const match = detail.match(/(?:SCAN|SEARCH|LOOP ON)\s+(\w+)/i);
				if (match) {
					console.log('getTables match ', match);
					const tableName = match[1];
					if (this.tableSet.has(tableName)) {
						matchTables.add(tableName);
					}
				}
			}
		} catch (e) {
			console.error('getTables error ', e);
		} finally {
			stmt.finalize();
		}
		const tablesArray = Array.from(matchTables);
		console.log('getTables sql ', sql);
		console.log('getTables tables ', tablesArray);
		return tablesArray;
	}

	query(sql: string, params: any = []): any[] {
		if (!this.db) throw new Error('Database is not initialized.');
		let stmt;
		const results: any = [];
		try {
			stmt = this.db.prepare(sql);

			if (params.length > 0) {
				stmt.bind(params);
			}

			const columnNames = stmt.getColumnNames();

			while (stmt.step()) {
				const row = stmt.get([]);
				const rowObject: any = {};
				columnNames.forEach((col: string, index: number) => {
					const jsonColumns = [
						'sources',
						'algorithms',
						'extra',
						'entries',
						'notes'
					];
					if (jsonColumns.includes(col)) {
						try {
							rowObject[col] = JSON.parse(row[index]);
						} catch (e) {
							console.warn(
								`columns ${col} value don't parse to JSON:`,
								row[index]
							);
							rowObject[col] = row[index];
						}
					} else {
						rowObject[col] = row[index];
					}
				});
				results.push(rowObject);
			}
		} catch (e) {
			console.error('Database query error ', e);
		} finally {
			stmt?.finalize();
		}
		return results;
	}

	private createTables(): void {
		this.tableSet = new Set();
		this.tableSet.add(createEntryTable(this.db));
		this.tableSet.add(createFeedTable(this.db));
		this.tableSet.add(createNoteTable(this.db));
		this.tableSet.add(createLabelTable(this.db));
		this.tableSet.add(createViewTable(this.db));
		console.log('Database tables created successfully.');
	}
}
