import {
	TransferItem,
	TransferItemInMemory
} from 'src/utils/interface/transfer';

export interface OlaresTaskInterface {
	setQueryResult(item: OlaresTaskBaseItem): Promise<OlaresTaskBaseItem>;
}
export type OlaresCopyTask = OlaresTaskInterface;

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
	task_type: 'copy' | 'upload' | 'move';
	retryCount: number;
	pause_able: boolean;
}
