import {
	TransferFront,
	TransferItem,
	TransferItemInMemory,
	TransferStatus
} from 'src/utils/interface/transfer';
import {
	convertTransferItemToMemeryItem,
	TransferBaseType
} from '../abstractions/transferManager/interface';
import { TransferDatabase } from 'src/utils/interface/transferDB';
import TransferClient from '../transfer';
import { useTransfer2Store } from 'src/stores/transfer2';
import { notifyFailed } from 'src/utils/notifyRedefinedUtil';
import Taskmanager from '../olaresTask';
import { OlaresTaskStatus } from '../abstractions/olaresTask/interface';

class CopyTransferType implements TransferBaseType {
	db: TransferDatabase | undefined = undefined;

	async formatTranferItemToMemeryType(item: TransferItem): Promise<{
		memeryItem: TransferItemInMemory;
		cancelItems: TransferItem[];
	}> {
		const memeryItem = convertTransferItemToMemeryItem(item);
		const cancelItems: TransferItem[] = [];
		const store = useTransfer2Store();

		if (memeryItem.phaseTaskId) {
			store.filesCopyTransferMap[memeryItem.phaseTaskId] = memeryItem.id!;

			if (
				memeryItem.status != TransferStatus.Canceled &&
				memeryItem.status != TransferStatus.Error &&
				memeryItem.status != TransferStatus.Completed
			) {
				Taskmanager.tasks[memeryItem.phaseTaskId] = {
					id: memeryItem.phaseTaskId,
					status: OlaresTaskStatus.PENDING,
					node: item.node || '',
					transfer_id: item.id,
					pending: false,
					task_type: memeryItem.front == TransferFront.move ? 'move' : 'copy',
					retryCount: 0,
					pause_able:
						memeryItem.pauseDisable != undefined
							? !memeryItem.pauseDisable
							: true
				};
			}
		}

		return {
			memeryItem: memeryItem,
			cancelItems: cancelItems
		};
	}
	async addPrecheck() {
		return undefined;
	}

	async pause(item: TransferItem) {
		if (!item.id) {
			return;
		}
		const result = await Taskmanager.pauseOrResumeTask(item, 'pause');
		const store = useTransfer2Store();
		if (result) {
			await store.pausedOrResumeTaskStatus(item.id, true);
			await store.updateTaskStatus(item.id, TransferStatus.Pending);
		}
	}

	async resume(item: TransferItem) {
		if (!item.id) {
			return;
		}

		const store = useTransfer2Store();
		await store.updateTaskStatus(item.id, TransferStatus.Resuming);

		const result = await Taskmanager.pauseOrResumeTask(item, 'resume');

		if (result) {
			await store.pausedOrResumeTaskStatus(item.id, false);
			await store.updateTaskStatus(item.id, TransferStatus.Pending);
		}
	}

	async bulkPause(ids: number[]): Promise<void> {}

	async bulkResume(ids: number[]) {}

	async cancel(item: TransferItem) {
		if (!item.id) {
			return;
		}
		const store = useTransfer2Store();
		await store.updateTaskStatus(item.id, TransferStatus.Canceling);
		const result = await Taskmanager.cancelTask(item);
		if (result) {
			await this.removeById(item.id);
		}
	}

	async bulkCancel(ids: number[]) {
		await Taskmanager.cancelAllTask();
		await this.bulkRemove(ids);
	}

	async remove(item: TransferItem): Promise<void> {
		if (item.id) await this.removeById(item.id);
	}

	async bulkRemove(ids: number[]): Promise<void> {
		for (const key of ids) {
			await this.removeById(key, false);
		}
		if (this.db) await this.db.transferData.bulkDelete(ids);
	}

	async removeById(id: number, operateDB = true) {
		if (operateDB && this.db) {
			await this.db.transferData.delete(id);
		}
		const store = useTransfer2Store();
		if (store.transferMap[id]) {
			delete store.transferMap[id];
			store.filesInDialogMap[id] && delete store.filesInDialogMap[id];
			store.transfers = store.transfers.filter((item) => item !== id);
			store.filesInDialogMap[id] && delete store.filesInDialogMap[id];
			store.filesInDialog = store.filesInDialog.filter((item) => item !== id);
		}
	}

	async onFileCompleted(id: number) {
		if (!id) {
			return;
		}
		const store = useTransfer2Store();
		await store.updateTaskStatus(id, TransferStatus.Completed);
	}

	async onFileError(id: number, message?: string) {
		let item: TransferItemInMemory | undefined = undefined;
		const store = useTransfer2Store();
		if (id in store.transferMap) {
			item = store.transferMap[id];
		}

		if (item && (await TransferClient.autoRetry(item))) {
			return false;
		}

		if (message) {
			notifyFailed(message);
		}

		if (item && message) {
			item.message = message;
			await store.update(id, {
				message: message
			});
		}

		await store.updateTaskStatus(id, TransferStatus.Error);

		return true;
	}

	startRunEnable(item: TransferItem) {
		return true;
	}
}

export default new CopyTransferType();
