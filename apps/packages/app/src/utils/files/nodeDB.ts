import Dexie from 'dexie';
import { FileNode } from 'src/stores/files';

// Cached node record tagged with the machine/account scope it belongs to, so
// cache from one Olares device is never shown for another.
export interface CachedFileNode extends FileNode {
	scope?: string;
}

export interface FilesSortRecord {
	id: string;
	by: number;
	asc: boolean;
}

export class FilesNodeDatabase extends Dexie {
	nodes: Dexie.Table<CachedFileNode, string>;
	sorts: Dexie.Table<FilesSortRecord, string>;

	constructor() {
		super('FilesNodeDatabase');
		this.version(2).stores({
			nodes: '&name,master',
			sorts: '&id'
		});
		// v3 adds the `scope` index so cached nodes can be filtered by machine.
		this.version(3).stores({
			nodes: '&name,master,scope',
			sorts: '&id'
		});
		this.nodes = this.table('nodes');
		this.sorts = this.table('sorts');
	}
}
