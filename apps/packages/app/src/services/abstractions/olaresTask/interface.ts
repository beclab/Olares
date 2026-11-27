import {
	TransferItem,
	TransferItemInMemory
} from 'src/utils/interface/transfer';

export interface OlaresTaskInterface {
	maxRetries?: number;
	setQueryResult(item: OlaresTaskBaseItem): Promise<OlaresTaskBaseItem>;
}
export type OlaresCopyTask = OlaresTaskInterface;

export type OlaresArchiveTask = OlaresTaskInterface;

export interface OlaresCloudUploadTask extends OlaresTaskInterface {
	addTasks(items: TransferItemInMemory[]): OlaresTaskBaseItem[];
}

export enum OlaresTaskStatus {
	PENDING = 'pending',
	RUNNING = 'running',
	COMPLETED = 'completed',
	CANCELLED = 'cancelled',
	CANCELED = 'canceled',
	FAILED = 'failed',
	PAUSED = 'paused'
}

export interface OlaresTaskBaseItem {
	id: string;
	status: OlaresTaskStatus;
	node: string;
	transfer_id?: number;
	pending?: boolean;
	task_type: 'copy' | 'upload' | 'move' | 'compress' | 'extract';
	retryCount: number;
	pause_able: boolean;
	maxRetries?: number;
	// Files store id of the view that started the task; used to refresh the right
	// listing on completion (e.g. the share view uses FilesIdType.SHARE, not the
	// default PAGEID).
	origin_id?: number;
}
