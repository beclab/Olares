import uploadTransferType from './upload';
import downloadTransferType from './download';
import cloudTransferType from './cloud';
import copyTransferType from './copy';
import {
	TransferFront,
	TransferItem,
	TransferItemInMemory,
	TransferStatus
} from 'src/utils/interface/transfer';
import { TransferDatabase } from 'src/utils/interface/transferDB';
import { useTransfer2Store } from 'src/stores/transfer2';
import { useDeviceStore } from 'src/stores/device';
import {
	TransferBaseType,
	calculateFolderItemCompletedCount,
	calculateFolderItemTotalCount
} from '../abstractions/transferManager/interface';
import TransferClient from '../transfer';
import { notifyFailed } from 'src/utils/notifyRedefinedUtil';

class TransferManager {
	upload = uploadTransferType;
	download = downloadTransferType;
	cloud = cloudTransferType;
	copy = copyTransferType;

	constructor() {
		this.upload.db = this.db;
		this.download.db = this.db;
		this.cloud.db = this.db;
		this.copy.db = this.db;
	}

	db = new TransferDatabase();

	async initTranferItemToMemeryType(item: TransferItem) {
		const transfetType = this.getTransferType(item.front);
		if (!transfetType) {
			return undefined;
		}

		if (item.front == TransferFront.copy || item.front == TransferFront.move) {
			return await this.copy.formatTranferItemToMemeryType(item);
		}

		return await transfetType.formatTranferItemToMemeryType(item, this.db);
	}

	resolveUpload() {
		const store = useTransfer2Store();
		const runningTask: TransferItemInMemory[] = [];
		for (const key in store.transferMap) {
			const value = store.transferMap[key];
			if (
				value.front === TransferFront.upload &&
				value.status == TransferStatus.Running &&
				value.currentPhase <= 1 &&
				value.isPaused == false
			) {
				runningTask.push(value);
			}
		}

		if (runningTask.length > this.upload.maxConnections) {
			store.updateTransferItemsSpeed(runningTask, TransferFront.upload);
			return;
		}

		const deviceStore = useDeviceStore();
		if (!deviceStore.transferEnable()) {
			return;
		}

		const pendingTask: TransferItemInMemory[] = [];
		for (const key in store.transferMap) {
			const value = store.transferMap[key];
			if (
				value.front === TransferFront.upload &&
				value.status == TransferStatus.Pending &&
				value.isPaused == false &&
				value.id
			) {
				pendingTask.push(value);
			}
		}

		if (pendingTask.length > 0 && pendingTask[0].id) {
			const leftNumbers = this.upload.maxConnections - runningTask.length;
			const leftPendingTasks =
				pendingTask.length > leftNumbers
					? pendingTask.slice(0, leftNumbers)
					: pendingTask;
			leftPendingTasks.forEach((e) => {
				store.run(e);
			});
		}
	}

	resolveDownload() {
		const store = useTransfer2Store();
		const runningTask: TransferItemInMemory[] = [];
		for (const key in store.transferMap) {
			const value = store.transferMap[key];
			if (
				value.front === TransferFront.download &&
				value.status == TransferStatus.Running &&
				value.isPaused == false
			) {
				runningTask.push(value);
			}
		}
		if (runningTask.length > this.download.maxConnections) {
			store.updateTransferItemsSpeed(runningTask, TransferFront.download);
			return;
		}
		const deviceStore = useDeviceStore();
		if (!deviceStore.transferEnable()) {
			return;
		}
		const pendingTask: TransferItemInMemory[] = [];

		for (const key in store.transferMap) {
			const value = store.transferMap[key];

			if (
				value.front === TransferFront.download &&
				value.status == TransferStatus.Pending &&
				value.isPaused == false
			) {
				pendingTask.push(value);
			}
		}

		if (pendingTask.length > 0 && pendingTask[0].id) {
			const leftNumbers = this.download.maxConnections - runningTask.length;
			const leftPendingTasks =
				pendingTask.length > leftNumbers
					? pendingTask.slice(0, leftNumbers)
					: pendingTask;
			leftPendingTasks.forEach((e) => {
				store.run(e);
			});
		}
	}

	checkStatus(
		tasks: {
			status: TransferStatus;
		}[]
	): TransferStatus {
		const items = tasks.find(
			(item) =>
				item.status === TransferStatus.Pending ||
				item.status === TransferStatus.Running ||
				item.status === TransferStatus.Canceled ||
				item.status === TransferStatus.Error
		);

		if (items) {
			return items.status;
		} else {
			return TransferStatus.Completed;
		}
	}

	async addPrecheck(item: TransferItem) {
		if (item.front == TransferFront.upload) {
			return this.upload.addPrecheck(item, this.db);
		}
		return undefined;
	}

	async add(item: TransferItem) {
		const oItem = await this.addPrecheck(item);

		if (oItem && oItem.id) {
			const transfer2Store = useTransfer2Store();
			if (oItem.isFolder) {
				setTimeout(() => {
					transfer2Store.prepare(oItem.id!);
				}, 100);
				return oItem.id;
			}
			if (oItem.status == TransferStatus.Pending) {
				return oItem.id;
			}
			transfer2Store.cancel(oItem);
		}

		const id = await this.db.transferData.add(item);
		const store = useTransfer2Store();
		store.transfers.unshift(id);

		store.filesInDialog.unshift(id);
		console.log('store.filesInDialog ===>', store.filesInDialog);
		const cur_map = {
			...item,
			id: id,
			speed: 0,
			progress: 0,
			leftTime: 0,
			isPaused: false,
			bytes: 0
		};

		store.transferMap[id] = cur_map;
		store.filesInDialogMap[id] = cur_map;

		if (item.isFolder) {
			await this.calculateFolderItemCompletedCount(cur_map, 'set', 0);
			await this.calculateFolderItemTotalCount(cur_map, 'set', 0);

			setTimeout(() => {
				store.prepare(id);
			}, 100);

			return id;
		}

		if (item.status == TransferStatus.Prepare) {
			await store.updateTaskStatus(id, TransferStatus.Pending);
		}

		return id;
	}

	async pause(item: TransferItem) {
		if (!item.id || !item.front) {
			return;
		}

		const transfetType = this.getTransferType(item.front);
		if (!transfetType) {
			return;
		}
		await transfetType.pause(item);
	}

	async bulkPause(ids: number[], front: TransferFront) {
		const transfetType = this.getTransferType(front);
		if (!transfetType) {
			return;
		}
		await transfetType.bulkPause(ids);
	}

	async resume(item: TransferItem) {
		const transfetType = this.getTransferType(item.front);
		if (!transfetType) {
			return;
		}
		await transfetType.resume(item);
	}

	async bulkResume(ids: number[], front: TransferFront) {
		const transfetType = this.getTransferType(front);
		if (!transfetType) {
			return;
		}
		await transfetType.bulkResume(ids);
	}

	async cancel(item: TransferItem) {
		const transfetType = this.getTransferType(item.front);
		if (!transfetType) {
			return;
		}
		await transfetType.cancel(item);
	}

	async bulkCancel(ids: number[], front: TransferFront) {
		const transfetType = this.getTransferType(front);
		if (!transfetType) {
			return;
		}
		await transfetType.bulkCancel(ids);
	}

	async remove(item: TransferItem) {
		const transfetType = this.getTransferType(item.front);
		if (!transfetType) {
			return;
		}
		await transfetType.remove(item);
	}

	async removeById(id: number, front: TransferFront) {
		const transfetType = this.getTransferType(front);
		if (!transfetType) {
			return;
		}
		await transfetType.removeById(id);
	}

	async bulkRemove(ids: number[], front: TransferFront): Promise<void> {
		const transfetType = this.getTransferType(front);
		if (!transfetType) {
			return;
		}
		await transfetType.bulkRemove(ids);
	}

	async onFileComplete(
		id: number,
		front: TransferFront,
		phase = 1,
		nextPhaseTaskId?: string
	) {
		const transfetType = this.getTransferType(front);
		if (!transfetType) {
			return;
		}
		await transfetType.onFileCompleted(id, phase, nextPhaseTaskId);
	}

	async onFileError(id: number, front: TransferFront, message?: string) {
		const transfetType = this.getTransferType(front);
		if (!transfetType) {
			return;
		}
		return await transfetType.onFileError(id, message);
	}

	async updateTaskStatus(id: number, status: TransferStatus) {
		await this.db.transferData.update(id, { status: status });
		const store = useTransfer2Store();
		if (status == TransferStatus.Completed && store.transferMap[id]) {
			TransferClient.doAction(store.transferMap[id], 'complete');
		}
	}

	async updateTaskPaused(id: number, isPaused: boolean) {
		await this.db.transferData.update(id, {
			isPaused: isPaused
		});
	}

	async calculateFolderItemCompletedCount(
		task: TransferItemInMemory,
		operate: 'add' | 'sub' | 'set',
		step: number
	) {
		return calculateFolderItemCompletedCount(task, operate, step);
	}

	async calculateFolderItemTotalCount(
		task: TransferItemInMemory,
		operate: 'add' | 'sub' | 'set',
		step: number
	) {
		return calculateFolderItemTotalCount(task, operate, step);
	}

	calFolderTask(transferItem: TransferItemInMemory, operate: 'add' | 'sub') {
		if (!transferItem.task || transferItem.task < 0) {
			return;
		}

		const store = useTransfer2Store();

		let items: Record<number, TransferItemInMemory[]> | undefined =
			store.taskCurrentSingleFiles[transferItem.front];

		if (!items) {
			items = {};
		}

		let task = items[transferItem.task];
		if (!task) {
			task = [];
		}
		const index = task.findIndex((e) => e.id == transferItem.id);
		if (operate == 'add') {
			if (index >= 0) {
				task.splice(index, 1, transferItem);
			} else {
				task.push(transferItem);
			}
		} else if (operate == 'sub') {
			if (index >= 0) {
				task.splice(index, 1);
			}
		}
		store.taskCurrentSingleFiles[transferItem.front][transferItem.task] = task;
	}

	preResumeCheck() {
		const deviceStore = useDeviceStore();
		if (!deviceStore.networkOnLine) {
			notifyFailed('please check your network');
			return false;
		}
		if (!deviceStore.transferWifiEnable()) {
			notifyFailed('select only wifi transfer file');
			return false;
		}

		return true;
	}

	startRunEnable(item: TransferItem) {
		const transfetType = this.getTransferType(item.front);
		if (!transfetType) {
			return true;
		}
		return transfetType.startRunEnable(item);
	}

	private getTransferType(front: TransferFront): TransferBaseType | undefined {
		if (front == TransferFront.download) {
			return this.download;
		} else if (front == TransferFront.upload) {
			return this.upload;
		} else if (front == TransferFront.cloud) {
			return this.cloud;
		} else if (front == TransferFront.copy || front == TransferFront.move) {
			return this.copy;
		}
		return undefined;
	}
}

export default new TransferManager();
