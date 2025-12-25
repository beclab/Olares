import {
	TransferItem,
	TransferItemInMemory,
	TransferStatus
} from 'src/utils/interface/transfer';

import { TransferDatabase } from 'src/utils/interface/transferDB';

export const convertTransferItemToMemeryItem = (
	item: TransferItem
): TransferItemInMemory => {
	return {
		speed: 0,
		leftTime: 0,
		bytes: 0,
		...item,
		progress: item.status == TransferStatus.Completed ? 1 : 0
	};
};

export const folderRecalculateCount = (
	initCount: number,
	operate: 'add' | 'sub' | 'set',
	step: number
) => {
	if (operate == 'add') {
		initCount = initCount + step;
	} else if (operate == 'sub') {
		initCount = initCount > step ? initCount - step : 0;
	} else if (operate == 'set') {
		initCount = step;
	}
	return initCount;
};

export const updateMemoryItemInfo = (
	item: TransferItemInMemory,
	bytes: number
) => {
	if (!item) return;
	const times = new Date().getTime();
	if (!item.bytes) item.bytes = 0;
	const seconds =
		item.bytes && item.bytes > 0
			? ((times - (item.updateTime ? item.updateTime : item.startTime!)) *
					1.0) /
			  1000
			: 1;
	if (item.size > 0 && seconds < 0.3 && bytes < item.size && item.bytes > 0) {
		return;
	}
	item.bytes = bytes;

	if (!item.speedTime) {
		item.speedTime = times;
		item.speed = bytes / 1;
	} else {
		if (times - item.speedTime > 1500) {
			const speed = (bytes - (item.preBytes || 0)) / 1.5;
			item.speed = speed > 0 ? speed : 0;
			if (speed > 0 && item.size > 0) {
				item.leftTime =
					(item.size * item.totalPhase -
						(bytes + item.size * (item.currentPhase - 1))) /
					item.speed;
			} else {
				item.leftTime = -1;
			}

			item.preBytes = bytes;
			item.speedTime = times;
		}
	}

	item.updateTime = times;
	if (item.size) {
		item.progress =
			(bytes * 1.0 + item.size * (item.currentPhase - 1)) /
			(item.size * item.totalPhase);
	}

	return item;
};

export const updateFolderSpeed = (
	folderItem: TransferItemInMemory,
	bytes: number,
	defaultSpeed = 0
) => {
	if (!folderItem.folderUpdateTime) {
		folderItem.folderUpdateTime = Date.now();
		folderItem.speed = defaultSpeed;
		folderItem.folderCompletedBytes = bytes;
	} else {
		if (Date.now() - 2000 > folderItem.folderUpdateTime) {
			folderItem.folderUpdateTime = Date.now();
			folderItem.speed = (bytes - (folderItem.folderCompletedBytes || 0)) / 2;
			folderItem.folderCompletedBytes = bytes;
		}
	}
};

export const calculateFolderItemTotalCount = (
	task: TransferItemInMemory,
	operate: 'add' | 'sub' | 'set',
	step: number
) => {
	if (!task.id) {
		return;
	}

	const folderTotalCount = task.folderTotalCount || 0;
	task.folderTotalCount = folderRecalculateCount(
		folderTotalCount,
		operate,
		step
	);
	return task;
};

export const calculateFolderItemCompletedCount = (
	task: TransferItemInMemory,
	operate: 'add' | 'sub' | 'set',
	step: number
) => {
	if (!task.id) {
		return;
	}
	const folderCompletedCount = task.folderCompletedCount || 0;
	task.folderCompletedCount = folderRecalculateCount(
		folderCompletedCount,
		operate,
		step
	);
	return task;
};

export const updateFolderMemoryItemInfo = (
	folderItem: TransferItemInMemory,
	subItem: TransferItemInMemory,
	runningSubItems: TransferItemInMemory[]
) => {
	if (!subItem.bytes) subItem.bytes = 0;

	const bytes = folderItem.bytes || 0;
	const folderSize = folderItem.size || 0;

	const { runningBytes, totalBytes } = runningSubItems.reduce(
		(acc, e) => {
			const cbytes = (e.bytes || 0) + (e.currentPhase - 1) * e.size;
			const totalBytes = (e.totalPhase - 1) * e.size;

			acc.runningBytes += cbytes;
			acc.totalBytes += totalBytes;

			return acc;
		},
		{ runningBytes: bytes, totalBytes: folderSize }
	);

	updateFolderSpeed(folderItem, runningBytes, subItem.speed);

	const progress = runningBytes / totalBytes;

	let leftTime = 0;
	if (folderItem.speed > 0) {
		leftTime = (totalBytes - runningBytes) / folderItem.speed;
	} else {
		leftTime = -1;
	}

	folderItem.progress = progress;
	folderItem.leftTime = leftTime;
};

export interface TransferBaseType {
	db: TransferDatabase | undefined;

	formatTranferItemToMemeryType(
		item: TransferItem,
		db: TransferDatabase
	): Promise<{
		memeryItem: TransferItemInMemory;
		cancelItems: TransferItem[];
	}>;

	addPrecheck(
		item: TransferItem,
		db: TransferDatabase
	): Promise<TransferItem | undefined>;

	pause(item: TransferItem): Promise<void>;

	bulkPause(ids: number[]): Promise<void>;

	resume(item: TransferItem): Promise<void>;

	bulkResume(ids: number[]): Promise<void>;

	cancel(item: TransferItem): Promise<void>;

	bulkCancel(ids: number[]): Promise<void>;

	remove(item: TransferItem): Promise<void>;

	bulkRemove(ids: number[]): Promise<void>;

	removeById(id: number): Promise<void>;

	onFileCompleted(
		id: number,
		phase?: number,
		nextPhaseTaskId?: string
	): Promise<void>;

	onFileError(id: number, message?: string): Promise<boolean>;

	startRunEnable(item: TransferItem): boolean;
}
