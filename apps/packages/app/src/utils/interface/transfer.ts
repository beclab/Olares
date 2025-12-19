import { DriveType } from './files';

export enum TransferFront {
	download = 1,
	upload = 2,
	cloud = 3,
	copy = 4,
	move = 5
}

export enum TransferStatus {
	Prepare = 'prepare',
	// Pausing = 'pausing',
	// Paused = 'paused',
	Resuming = 'resuming',
	Checking = 'checking',
	Canceling = 'canceling',
	Canceled = 'canceled',
	Removing = 'removing',

	All = 'all',

	//wise only has the following states.
	Running = 'running',
	Pending = 'pending',
	Removed = 'removed',
	Completed = 'completed',
	Error = 'error'
}

export interface TransferItem {
	id?: number;
	task?: number;
	name: string;

	path: string;
	type: string;
	isFolder: boolean;
	driveType?: DriveType;
	front: TransferFront;
	status: TransferStatus;

	url?: string;

	startTime?: number;
	endTime?: number;
	updateTime?: number;

	from?: string;
	to?: string;

	isPaused: boolean;

	size: number;
	message?: string;
	uniqueIdentifier?: string;
	repo_id?: string;
	params?: string;
	parentPath?: string;

	userId?: string;

	relatePath?: string;
	wiseRecordId?: string;
	cancellable?: boolean;

	pauseDisable?: boolean;

	node?: string;

	currentPhase: number;
	totalPhase: number;
	phaseTaskId?: string;
}

export interface TransferItemInMemory extends TransferItem {
	speed: number;
	progress: number;
	leftTime: number;
	bytes?: number;
	preBytes?: number;
	folderCompletedCount?: number;
	folderTotalCount?: number;
	prepareNextTime?: number;

	networkOfflinePaused?: boolean;
	onlyWifiPaused?: boolean;

	folderUpdateTime?: number;
	folderCompletedBytes?: number;
	retryCount?: number;

	speedTime?: number;
}

export const transferItemIsPaused = (item: TransferItemInMemory) => {
	return item.isPaused || item.networkOfflinePaused || item.onlyWifiPaused;
};
