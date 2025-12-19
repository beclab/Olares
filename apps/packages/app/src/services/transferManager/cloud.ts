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

class CloudTransferType implements TransferBaseType {
	db: TransferDatabase | undefined = undefined;

	async formatTranferItemToMemeryType(item: TransferItem): Promise<{
		memeryItem: TransferItemInMemory;
		cancelItems: TransferItem[];
	}> {
		const memeryItem = convertTransferItemToMemeryItem(item);
		const cancelItems: TransferItem[] = [];
		const store = useTransfer2Store();
		if (item.uniqueIdentifier) {
			store.filesCloudTransferMap[item.uniqueIdentifier] = item.id!;
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
		if (!item.id || !TransferClient.client.clouder) {
			return;
		}
		const result = await TransferClient.client.clouder.pause(item);
		const store = useTransfer2Store();
		if (result) {
			await store.pausedOrResumeTaskStatus(item.id, true);
			await store.updateTaskStatus(item.id, TransferStatus.Pending);
		}
	}

	async bulkPause(ids: number[]): Promise<void> {}

	async resume(item: TransferItem) {
		if (!item.id || !TransferClient.client.clouder) {
			return;
		}

		const store = useTransfer2Store();
		await store.updateTaskStatus(item.id, TransferStatus.Resuming);

		const result = await TransferClient.client.clouder.resume(item);

		if (result) {
			console.log('cloud download resume successs');
		}

		await store.pausedOrResumeTaskStatus(item.id, false);

		await store.updateTaskStatus(item.id, TransferStatus.Pending);
	}

	async bulkResume(ids: number[]) {}

	async cancel(item: TransferItem) {
		if (!item.id) {
			return;
		}
		const store = useTransfer2Store();
		await store.updateTaskStatus(item.id, TransferStatus.Canceling);
		await TransferClient.doAction(item, 'cancel');
		await this.removeById(item.id);
	}

	async bulkCancel(ids: number[]) {
		if (TransferClient.client.clouder) {
			await TransferClient.client.clouder.cancelAllTask();
		}
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

export default new CloudTransferType();
