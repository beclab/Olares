import { sendMessageToWorker } from '../sqliteService';
import { Note } from '../../../../utils/rss-types';
import { convertIntegerPropertiesToBoolean } from '../utils';

const booleanProperties = ['deleted'];

function convertNote(notes: Note[]) {
	return notes.map((filter: any) =>
		convertIntegerPropertiesToBoolean(filter, booleanProperties)
	);
}

export function createNoteTable(db: any): string {
	db.exec(`
		CREATE TABLE IF NOT EXISTS notes (
			id TEXT PRIMARY KEY,
			entry_id TEXT,
			highlight TEXT,
			content TEXT,
			start INTEGER,
			length INTEGER,
			deleted INTEGER,
			create_at TEXT,
			updated_at TEXT
		);
	`);
	return 'notes';
}

export async function addOrUpdateNotes(notes: Note[]) {
	if (notes.length === 0) {
		console.log('Note update length 0 return');
		return;
	}
	try {
		const params = notes.map((note) => [
			note.id,
			note.entry_id,
			note.highlight,
			note.content,
			note.start,
			note.length,
			note.deleted ? 1 : 0,
			note.create_at,
			note.updated_at
		]);
		await sendMessageToWorker('transaction', {
			sql: `
				INSERT INTO notes (id, entry_id, highlight, content, start, length, deleted, create_at, updated_at)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
				ON CONFLICT(id) DO UPDATE SET
					entry_id = excluded.entry_id,
					highlight = excluded.highlight,
					content = excluded.content,
					start = excluded.start,
					length = excluded.length,
					deleted = excluded.deleted,
					create_at = excluded.create_at,
					updated_at = excluded.updated_at;
			`,
			params: params
		});
		console.log('Note added/updated successfully!');
	} catch (error) {
		console.error('Failed to add/update note:', error);
	}
}

export async function getAllNotes(): Promise<Note[]> {
	try {
		const notes: any = await sendMessageToWorker('query', {
			sql: `SELECT * FROM notes;`
		});
		const convertList = convertNote(notes);
		return convertList.length > 0 ? (convertList as Note[]) : [];
	} catch (error) {
		console.error('Failed to get all notes:', error);
		return [];
	}
}

export async function getNoteById(id: string): Promise<Note | null> {
	try {
		const notes: any = await sendMessageToWorker('query', {
			sql: `SELECT * FROM notes WHERE id = ?;`,
			params: [id]
		});
		const convertList = convertNote(notes);
		return convertList.length > 0 ? (convertList[0] as Note) : null;
	} catch (error) {
		console.error('Failed to get note:', error);
		return null;
	}
}

export async function removeNoteById(id: string) {
	try {
		await sendMessageToWorker('execute', {
			sql: `DELETE FROM notes WHERE id = ?;`,
			params: [id]
		});
		console.log('Note removed successfully!');
	} catch (error) {
		console.error('Failed to remove note:', error);
	}
}

export async function clearNotes() {
	try {
		await sendMessageToWorker('execute', {
			sql: `DELETE
						FROM notes`
		});
		console.log('Note clear successfully!');
	} catch (error) {
		console.error('Failed to clear note:', error);
	}
}
