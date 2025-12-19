import {
	TransferFront,
	TransferItem,
	TransferItemInMemory,
	TransferStatus
} from 'src/utils/interface/transfer';
import {
	calculateFolderItemCompletedCount,
	calculateFolderItemTotalCount,
	convertTransferItemToMemeryItem,
	TransferBaseType,
	updateFolderSpeed,
	updateMemoryItemInfo
} from '../abstractions/transferManager/interface';

import TransferClient from '../transfer';
import { TransferDatabase } from 'src/utils/interface/transferDB';
import { TransferType, useTransfer2Store } from 'src/stores/transfer2';
import { notifyFailed, notifySuccess } from 'src/utils/notifyRedefinedUtil';
import { i18n } from 'src/boot/i18n';
import { useDeviceStore } from 'src/stores/device';

class DownloadTransferType implements TransferBaseType {
	maxConnections = 1;
	folderMaxConnections = 1;

	db: TransferDatabase | undefined = undefined;

	async formatTranferItemToMemeryType(
		item: TransferItem,
		db: TransferDatabase
	): Promise<{
		memeryItem: TransferItemInMemory;
		cancelItems: TransferItem[];
	}> {
		const memeryItem = convertTransferItemToMemeryItem(item);

		const restartEnable = await TransferClient.client.downloader.restartEnable(
			item
		);

		const restartAutoResume =
			TransferClient.client.downloader.restartAutoResume;

		let cancelItems: TransferItem[] = [];

		if (!restartEnable) {
			if (item.status != TransferStatus.Completed) {
				memeryItem.status = TransferStatus.Canceled;
				cancelItems.push(item);
			}
		} else {
			if (!restartAutoResume && item.status == TransferStatus.Pending) {
				memeryItem.isPaused = true;
			}
		}

		if (item.isFolder) {
			const currentTasks = await db.transferData
				.where('task')
				.equals(item.id!)
				.toArray();

			const hasCompleted = currentTasks.filter(
				(item) => item.status === TransferStatus.Completed
			);

			if (!restartEnable) {
				cancelItems = cancelItems.concat(
					currentTasks.filter(
						(e) =>
							e.status != TransferStatus.Completed &&
							e.status != TransferStatus.Canceled
					)
				);
			}

			memeryItem.folderCompletedCount = hasCompleted.length;
			memeryItem.folderTotalCount = currentTasks.length;
			if (item.status != TransferStatus.Completed) {
				const size = hasCompleted.reduce((accumulator, item) => {
					return accumulator + item.size;
				}, 0);
				memeryItem.bytes = size;
			}
		} else {
			if (
				restartEnable &&
				item.id &&
				item.status != TransferStatus.Completed &&
				item.status != TransferStatus.Canceled
			) {
				const transferInfo =
					await TransferClient.client.uploader.getTransferInfo(item);
				if (transferInfo && transferInfo.bytes > 0) {
					updateMemoryItemInfo(memeryItem, transferInfo.bytes);
				}
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

		let result = false;
		if (item.isFolder) {
			await this.pauseOrResumeFolderTask(item);
			result = true;
		} else {
			result = await TransferClient.client.uploader.pause(item);
		}

		const store = useTransfer2Store();
		if (result) {
			await store.pausedOrResumeTaskStatus(item.id, true);
			await store.updateTaskStatus(item.id, TransferStatus.Pending);
		}

		if (item.task && item.task > 0) {
			await store.runNextFileInTask(item.task, item.front);
		}
	}

	async bulkPause(ids: number[]): Promise<void> {
		const store = useTransfer2Store();
		ids.forEach((id) => {
			if (id in store.filesInFolderMap) {
				this.pause(store.filesInFolderMap[id]);
			} else {
				this.pause(store.transferMap[id]);
			}
		});
	}

	async resume(item: TransferItem) {
		if (!item.id) {
			return;
		}

		const store = useTransfer2Store();
		await store.updateTaskStatus(item.id, TransferStatus.Resuming);
		await store.pausedOrResumeTaskStatus(item.id, false);

		if (item.isFolder) {
			await this.pauseOrResumeFolderTask(item, false);
		}

		await store.updateTaskStatus(item.id, TransferStatus.Pending);

		if (item.task && item.task > 0) {
			if (store.transferMap[item.task].isPaused) {
				store.pausedOrResumeTaskStatus(item.task, false);
			}
			if (
				!store.taskCurrentSingleFiles[item.front][item.task] ||
				store.taskCurrentSingleFiles[item.front][item.task].length == 0
			) {
				await store.runNextFileInTask(item.task, item.front);
			}
		}
	}

	async bulkResume(ids: number[]) {
		const store = useTransfer2Store();
		ids.forEach((id) => {
			if (id in store.filesInFolderMap) {
				this.resume(store.filesInFolderMap[id]);
			} else {
				this.resume(store.transferMap[id]);
			}
		});
	}

	async cancel(item: TransferItem, notify = true) {
		if (!item.id) {
			return;
		}
		const store = useTransfer2Store();
		await store.updateTaskStatus(item.id, TransferStatus.Canceling);
		if (item.isFolder) {
			await this.cancelAndRemoveFolder(item);
		} else {
			await TransferClient.doAction(item, 'cancel');
			await this.removeById(item.id);
		}
		if (notify)
			notifySuccess(
				i18n.global.t('files.remove_file', {
					fileName: item.name
				})
			);

		if (item.task && item.task > 0) {
			const task = store.transferMap[item.task];
			if (!task) {
				return;
			}

			calculateFolderItemTotalCount(task, 'sub', 1);

			if (
				item.status == TransferStatus.Completed ||
				item.status == TransferStatus.Canceled
			) {
				calculateFolderItemCompletedCount(task, 'sub', 1);
			}

			if (task.status != TransferStatus.Canceling) {
				if (
					store.taskCurrentSingleFiles[item.front] &&
					store.taskCurrentSingleFiles[item.front][item.task] &&
					store.taskCurrentSingleFiles[item.front][item.task].find(
						(e) => e.id == item.id
					)
				) {
					store.calFolderTask(item as any, 'sub');
					await store.runNextFileInTask(item.task!, item.front);
				}
			}
		}
	}

	async bulkCancel(ids: number[]) {
		const store = useTransfer2Store();
		ids.forEach((id) => {
			if (id in store.filesInFolderMap) {
				this.cancel(store.filesInFolderMap[id], false);
			} else {
				this.cancel(store.transferMap[id], false);
			}
		});
	}

	async remove(item: TransferItem): Promise<void> {
		if (item.id) await this.removeById(item.id);
	}

	async bulkRemove(ids: number[]): Promise<void> {
		const filter_ids: number[] = [];
		const store = useTransfer2Store();

		ids.forEach((id) => {
			if (id in store.filesInFolderMap) {
				filter_ids.push(id);
			} else {
				if (store.transferMap[id] && !store.transferMap[id].isFolder) {
					filter_ids.push(id);
				}
			}
		});

		for (const key of ids) {
			await this.removeById(key, false);
		}
		if (this.db) await this.db.transferData.bulkDelete(filter_ids);
	}

	startRunEnable(item: TransferItem) {
		const deviceStore = useDeviceStore();
		return deviceStore.transferEnable();
	}

	async removeById(id: number, operateDB = true) {
		if (operateDB && this.db) {
			await this.db.transferData.delete(id);
		}
		const store = useTransfer2Store();
		if (id in store.filesInFolderMap) {
			delete store.filesInFolderMap[id];
			store.filesInFolder = store.filesInFolder.filter((e) => e !== id);
		} else {
			if (store.transferMap[id] && store.transferMap[id].isFolder) {
				this.cancelAndRemoveFolder(store.transferMap[id]);
			} else {
				if (store.transferMap[id]) {
					delete store.transferMap[id];
					store.filesInDialogMap[id] && delete store.filesInDialogMap[id];
					store.transfers = store.transfers.filter((item) => item !== id);
					store.filesInDialogMap[id] && delete store.filesInDialogMap[id];
					store.filesInDialog = store.filesInDialog.filter(
						(item) => item !== id
					);
				}
			}
		}
	}

	async onFileCompleted(id: number) {
		if (!id) {
			return;
		}
		const store = useTransfer2Store();
		await store.updateTaskStatus(id, TransferStatus.Completed);

		if (id in store.transferMap) {
			/* empty */
		} else {
			store.replaceFileInFolder(id);

			const currentSingleFile = store.getSubTransferItem(
				TransferFront.download,
				id
			);

			if (!currentSingleFile) return;

			const folderTask = store.transferMap[currentSingleFile.task as number];

			if (folderTask) {
				folderTask.bytes =
					(folderTask.bytes || 0) + (currentSingleFile.size || 0);
				updateFolderSpeed(folderTask, folderTask.bytes);

				store.calFolderTask(currentSingleFile, 'sub');
				await store.runNextFileInTask(folderTask.id!, currentSingleFile.front);
			}
		}
	}

	async onFileError(id: number, message?: string) {
		let item: TransferItemInMemory | undefined = undefined;
		const store = useTransfer2Store();
		const currentSingleFile = store.getSubTransferItem(
			TransferFront.download,
			id
		);
		if (id in store.transferMap) {
			item = store.transferMap[id];
		} else if (currentSingleFile && currentSingleFile.id == id) {
			item = currentSingleFile;
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

		if (item && item.task) {
			store.calFolderTask(item, 'sub');
			store.runNextFileInTask(item.task, item.front);
		}
		return true;
	}

	private async pauseOrResumeFolderTask(item: TransferItem, isPaused = true) {
		if (!item.id || !this.db) {
			return;
		}

		const items = await this.db.transferData
			.where('task')
			.equals(item.id)
			.and(
				(item) =>
					item.status !== TransferStatus.Completed &&
					item.status !== TransferStatus.Canceled
			)
			.toArray();

		items.forEach(async (e) => {
			await TransferClient.doAction(e, 'pause');
		});

		const update = items.map((e) => {
			return {
				key: e.id!,
				changes: {
					...e,
					isPaused: isPaused,
					status: TransferStatus.Pending
				}
			};
		});

		await this.db.transferData.bulkUpdate(update);
	}

	private async cancelAndRemoveFolder(item: TransferItem) {
		if (!item.id || !this.db) {
			return;
		}
		if (!item.isFolder) {
			return;
		}

		const items = await this.db.transferData
			.where('task')
			.equals(item.id)
			.toArray();

		let ids: number[] = [];

		const store = useTransfer2Store();

		if (
			store.transferType === TransferType.UPLOADED ||
			store.transferType === TransferType.DOWNLOADED
		) {
			ids = items
				.map((e) => {
					if (
						e.id &&
						(e.status === TransferStatus.Completed ||
							e.status === TransferStatus.Canceled)
					) {
						return e.id;
					}
				})
				.filter((e) => e) as number[];
		} else {
			ids = items
				.map((e) => {
					if (
						e.id &&
						e.status !== TransferStatus.Completed &&
						e.status !== TransferStatus.Canceled
					) {
						return e.id;
					}
				})
				.filter((e) => e) as number[];

			for (let index = 0; index < items.length; index++) {
				if (ids.find((e) => e == items[index].id)) {
					TransferClient.doAction(items[index], 'cancel');
				}
			}
		}

		await this.db.transferData.bulkDelete(ids);

		const now_items = await this.db.transferData
			.where('task')
			.equals(item.id)
			.toArray();

		if (now_items.length <= 0) {
			await this.db.transferData.delete(item.id);
			delete store.transferMap[item.id];
			store.filesInDialogMap[item.id] && delete store.filesInDialogMap[item.id];
			store.transfers = store.transfers.filter((e) => e !== item.id);
			store.filesInDialog = store.filesInDialog.filter((e) => e !== item.id);
		} else {
			//
			const folderCompletedCount = now_items.filter(
				(e) => e.status === TransferStatus.Completed
			).length;

			if (folderCompletedCount == now_items.length) {
				await store.updateTaskStatus(item.id, TransferStatus.Completed);
			}

			calculateFolderItemTotalCount(
				store.transferMap[item.id],
				'set',
				now_items.length
			);
		}
	}
}

export default new DownloadTransferType();
