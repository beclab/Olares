import { defineStore } from 'pinia';
import { i18n } from '../boot/i18n';
import { common, filesIsV2 } from '../api';
import TransferClient from '../services/transfer';
import Taskmanager from '../services/olaresTask';
import transferManager from '../services/transferManager';
import { busOn } from 'src/utils/bus';
import { useDeviceStore } from './device';
import {
	TransferFront,
	TransferItem,
	TransferItemInMemory,
	TransferStatus
} from 'src/utils/interface/transfer';
import {
	convertTransferItemToMemeryItem,
	updateFolderMemoryItemInfo,
	updateMemoryItemInfo
} from 'src/services/abstractions/transferManager/interface';

export enum TransferType {
	UPLOADING,
	DOWNLOADING,
	UPLOADED,
	DOWNLOADED,
	CLOUDING,
	CLOUDED,
	COPYING,
	COPIED
}

export enum TransferMobileItem {
	DATE,
	PROCESSING,
	COMPLETED,
	ERROR,
	EMPTY
}

export function formatLeftTimes(leftTimes: number): string {
	if (leftTimes < 0) {
		return '--';
	}
	const seconds = leftTimes % 60;
	const minutes = Math.floor(leftTimes / 60);
	const hours = Math.floor(minutes / 60);

	const leftMinute = Math.floor(minutes % 60);

	return (
		hours.toFixed(0) +
		':' +
		leftMinute.toFixed(0).padStart(2, '0') +
		':' +
		seconds.toFixed(0).padStart(2, '0') +
		''
	);
}

let moveTimer: NodeJS.Timeout | null = null;
let isInTickTransfer = false;
const forderMaxLength = 1;

function getCurrentTransferingItems() {
	const store = useTransfer2Store();

	const runningTask: TransferItemInMemory[] = [];
	for (const key in store.transferMap) {
		const value = store.transferMap[key];
		if (
			value.front != TransferFront.cloud &&
			value.status == TransferStatus.Running &&
			value.isPaused == false
		) {
			runningTask.push(value);
		}
	}

	return runningTask;
}

export type DataState = {
	activeItem: TransferFront;
	transferType: TransferType;
	isUploadProgressDialogShow: boolean;
	taskCurrentSingleFiles: {
		[TransferFront.download]: Record<number, TransferItemInMemory[]>;
		[TransferFront.upload]: Record<number, TransferItemInMemory[]>;
	};
	transfers: number[];
	transferMap: Record<number, TransferItemInMemory>;
	filesInFolder: number[];
	filesInFolderMap: Record<number, TransferItemInMemory>;
	filesInDialog: number[];
	filesInDialogMap: Record<number, TransferItemInMemory>;
	filesCloudTransferMap: Record<string, number>;
	filesCopyTransferMap: Record<string, number>;
	isIniting: boolean;
};

export const useTransfer2Store = defineStore('transfer2', {
	state: () => {
		return {
			activeItem: TransferFront.upload,
			transferType: TransferType.UPLOADING,
			isUploadProgressDialogShow: false,
			taskCurrentSingleFiles: {
				[TransferFront.download]: {},
				[TransferFront.upload]: {}
			},
			transfers: [],
			transferMap: {},
			filesInFolder: [],
			filesInFolderMap: {},
			filesInDialog: [],
			filesInDialogMap: {},
			filesCloudTransferMap: {},
			filesCopyTransferMap: {},
			isIniting: true
		} as DataState;
	},

	getters: {
		uploadComplete: (state) => {
			return state.transfers.filter((item: number) => {
				return (
					state.transferMap[item] &&
					((state.transferMap[item].front === TransferFront.upload &&
						!state.transferMap[item].isFolder &&
						(state.transferMap[item].status == TransferStatus.Completed ||
							state.transferMap[item].status == TransferStatus.Canceled)) ||
						(state.transferMap[item].front === TransferFront.upload &&
							state.transferMap[item].isFolder)) &&
					(TransferClient.getUserId()
						? TransferClient.getUserId() == state.transferMap[item].userId
						: true)
				);
			});
		},
		uploading: (state) => {
			return state.transfers.filter((item: number) => {
				return (
					state.transferMap[item] &&
					state.transferMap[item].front === TransferFront.upload &&
					state.transferMap[item].status != TransferStatus.Completed &&
					state.transferMap[item].status != TransferStatus.Canceled &&
					(TransferClient.getUserId()
						? TransferClient.getUserId() == state.transferMap[item].userId
						: true)
				);
			});
		},
		downloadComplete: (state) => {
			return state.transfers.filter((item: number) => {
				return (
					state.transferMap[item] &&
					((state.transferMap[item].front === TransferFront.download &&
						!state.transferMap[item].isFolder &&
						(state.transferMap[item].status === TransferStatus.Completed ||
							state.transferMap[item].status === TransferStatus.Canceled)) ||
						(state.transferMap[item].front === TransferFront.download &&
							state.transferMap[item].isFolder)) &&
					(TransferClient.getUserId()
						? TransferClient.getUserId() == state.transferMap[item].userId
						: true)
				);
			});
		},
		downloading: (state) => {
			return state.transfers.filter((item: number) => {
				return (
					state.transferMap[item] &&
					state.transferMap[item].front === TransferFront.download &&
					state.transferMap[item].status != TransferStatus.Completed &&
					state.transferMap[item].status != TransferStatus.Canceled &&
					(TransferClient.getUserId()
						? TransferClient.getUserId() == state.transferMap[item].userId
						: true)
				);
			});
		},

		upload: (state) => {
			return state.transfers.filter((item: number) => {
				return (
					state.transferMap[item] &&
					state.transferMap[item].front === TransferFront.upload &&
					!state.transferMap[item].isFolder &&
					(TransferClient.getUserId()
						? TransferClient.getUserId() == state.transferMap[item].userId
						: true)
				);
			});
		},

		download: (state) => {
			return state.transfers.filter((item: number) => {
				return (
					state.transferMap[item] &&
					state.transferMap[item].front === TransferFront.download &&
					!state.transferMap[item].isFolder &&
					(TransferClient.getUserId()
						? TransferClient.getUserId() == state.transferMap[item].userId
						: true)
				);
			});
		},
		cloudComplete: (state) => {
			return state.transfers.filter((item: number) => {
				return (
					state.transferMap[item] &&
					state.transferMap[item].front === TransferFront.cloud &&
					(state.transferMap[item].status === TransferStatus.Completed ||
						state.transferMap[item].status === TransferStatus.Canceled) &&
					(TransferClient.getUserId()
						? TransferClient.getUserId() == state.transferMap[item].userId
						: true)
				);
			});
		},
		clouding: (state) => {
			return state.transfers.filter((item: number) => {
				return (
					state.transferMap[item] &&
					state.transferMap[item].front === TransferFront.cloud &&
					state.transferMap[item].status != TransferStatus.Completed &&
					state.transferMap[item].status != TransferStatus.Canceled &&
					(TransferClient.getUserId()
						? TransferClient.getUserId() == state.transferMap[item].userId
						: true)
				);
			});
		},
		cloud: (state) => {
			return state.transfers.filter((item: number) => {
				return (
					state.transferMap[item] &&
					state.transferMap[item].front === TransferFront.cloud &&
					(TransferClient.getUserId()
						? TransferClient.getUserId() == state.transferMap[item].userId
						: true)
				);
			});
		},

		copyComplete: (state) => {
			return state.transfers.filter((item: number) => {
				return (
					state.transferMap[item] &&
					state.transferMap[item].front === TransferFront.copy &&
					(state.transferMap[item].status === TransferStatus.Completed ||
						state.transferMap[item].status === TransferStatus.Canceled) &&
					(TransferClient.getUserId()
						? TransferClient.getUserId() == state.transferMap[item].userId
						: true)
				);
			});
		},
		copying: (state) => {
			return state.transfers.filter((item: number) => {
				return (
					state.transferMap[item] &&
					state.transferMap[item].front === TransferFront.copy &&
					state.transferMap[item].status != TransferStatus.Completed &&
					state.transferMap[item].status != TransferStatus.Canceled &&
					(TransferClient.getUserId()
						? TransferClient.getUserId() == state.transferMap[item].userId
						: true)
				);
			});
		},
		copy: (state) => {
			return state.transfers.filter((item: number) => {
				return (
					state.transferMap[item] &&
					state.transferMap[item].front === TransferFront.copy &&
					(TransferClient.getUserId()
						? TransferClient.getUserId() == state.transferMap[item].userId
						: true)
				);
			});
		},

		shareUploadList: (state) => (id: string) => {
			return state.transfers.filter((item: number) => {
				return (
					state.transferMap[item] &&
					state.transferMap[item].front === TransferFront.upload &&
					state.transferMap[item].path.startsWith('/Share/' + id)
				);
			});
		},
		shareUploadedList: (state) => (id: string) => {
			return state.transfers.filter((item: number) => {
				return (
					state.transferMap[item] &&
					state.transferMap[item].front === TransferFront.upload &&
					(state.transferMap[item].status == TransferStatus.Completed ||
						state.transferMap[item].status == TransferStatus.Canceled) &&
					state.transferMap[item].path.startsWith('/Share/' + id)
				);
			});
		},
		shareUploadingList: (state) => (id: string) => {
			return state.transfers.filter((item: number) => {
				return (
					state.transferMap[item] &&
					state.transferMap[item].front === TransferFront.upload &&
					state.transferMap[item].status != TransferStatus.Completed &&
					state.transferMap[item].status != TransferStatus.Canceled &&
					state.transferMap[item].path.startsWith('/Share/' + id)
				);
			});
		}
	},

	actions: {
		menus() {
			const items = [
				{
					label: i18n.global.t('transmission.title'),
					key: 'transmission',
					icon: '',
					children: [
						{
							label: i18n.global.t('transmission.upload.title'),
							icon: 'sym_r_cloud_upload',
							key: `${TransferFront.upload}`,
							count: 0
						},
						{
							label: i18n.global.t('transmission.download.title'),
							icon: 'sym_r_browser_updated',
							key: `${TransferFront.download}`,
							count: 0
						},
						{
							label: i18n.global.t('transmission.cloud.title'),
							icon: 'sym_r_cloud_download',
							key: `${TransferFront.cloud}`
						}
					]
				}
			] as {
				label: string;
				key: string;
				icon: string;
				children: {
					label: string;
					key: string;
					icon: string;
					count: string | number | undefined;
				}[];
			}[];
			if (filesIsV2()) {
				items[0].children.push({
					label: i18n.global.t('transmission.copy_paste'),
					icon: 'sym_r_content_copy',
					key: `${TransferFront.copy}`,
					count: 0
				});
			}
			return items;
		},

		async init() {
			this.isIniting = true;

			const transferInDB = await transferManager.db.transferData
				.where('task')
				.equals(-1)
				.reverse()
				.limit(1000)
				.toArray();

			let canceledItems: TransferItem[] = [];

			for (const item of transferInDB) {
				this.transfers.push(item.id!);
				const memerys = await transferManager.initTranferItemToMemeryType(item);
				if (!memerys) {
					continue;
				}
				this.transferMap[item.id!] = memerys.memeryItem;
				canceledItems = canceledItems.concat(memerys.cancelItems);
			}

			if (canceledItems.length > 0) {
				const update = canceledItems.map((e) => {
					return {
						key: e.id!,
						changes: {
							...e,
							status: TransferStatus.Canceled
						}
					};
				});
				await transferManager.db.transferData.bulkUpdate(update);
			}

			if (moveTimer) {
				clearInterval(moveTimer);
			}

			if (process.env.APPLICATION == 'FILES') {
				moveTimer = setInterval(async () => {
					if (isInTickTransfer) {
						return;
					}
					isInTickTransfer = true;

					transferManager.resolveUpload();
					transferManager.resolveDownload();
					Taskmanager.query();

					isInTickTransfer = false;
				}, 1000);
			} else {
				busOn('runTask', () => {
					transferManager.resolveUpload();
					transferManager.resolveDownload();
					Taskmanager.query();
				});

				busOn('network_update', () => {
					this.updateRunningItemsStatus();
				});

				busOn('appTransferTypeChanged', () => {
					this.updateRunningItemsStatus();
				});

				busOn('userIsLocalUpdate', () => {
					this.updateRunningItemsBaseUrl();
				});

				setTimeout(() => {
					transferManager.resolveUpload();
					transferManager.resolveDownload();
				}, 1000);
			}

			this.isIniting = false;
		},

		async add(
			item: TransferItem,
			front = TransferFront.upload
		): Promise<number> {
			const transferData: TransferItem = {
				task: -1,
				name: item.name,
				path: item.path,
				parentPath: item.parentPath,
				type: item.type,
				isFolder: item.isFolder,
				driveType: item.driveType
					? item.driveType
					: common().formatUrltoDriveType(item.path),
				front: front,
				status: item.status,
				url: item.url,
				startTime: item.startTime ? item.startTime : new Date().getTime(),
				endTime: 0,
				updateTime: item.startTime ? item.startTime : new Date().getTime(),
				from:
					front == TransferFront.download
						? item.path
						: front == TransferFront.upload
						? item.from
						: item.from,
				to:
					front == TransferFront.download
						? ''
						: front == TransferFront.upload
						? item.path
						: item.to,
				isPaused: item.isPaused,
				size: item.size,
				message: item.message || '',
				uniqueIdentifier: item.uniqueIdentifier,
				repo_id: item.repo_id,
				params: item.params,
				userId: TransferClient.getUserId(),
				cancellable: item.cancellable,
				node: item.node,
				currentPhase: item.currentPhase ? item.currentPhase : 1,
				totalPhase: item.totalPhase ? item.totalPhase : 1,
				phaseTaskId: item.phaseTaskId ? item.phaseTaskId : undefined,
				pauseDisable: item.pauseDisable != undefined ? item.pauseDisable : false
			};

			return await transferManager.add(transferData);
		},

		async update(id: number, item: Partial<TransferItem>): Promise<number> {
			return await transferManager.db.transferData.update(id, item);
		},

		async prepare(taskId: number) {
			if (!this.transferMap[taskId]) {
				return;
			}
			const task = this.transferMap[taskId];

			const childPrepare = await TransferClient.prepare(
				task,
				task.folderTotalCount
			);

			if (!childPrepare) {
				return;
			}

			if (childPrepare.child?.length == 0) {
				if (childPrepare && !childPrepare.finished) {
					setTimeout(() => {
						this.prepare(taskId);
					}, 1000);
					return;
				}
			}

			const oTransferItems: TransferItem[] = childPrepare.child || [];

			const transferItems: TransferItem[] = [];
			for (const item of oTransferItems) {
				const e = await transferManager.addPrecheck(item);
				if (!e) {
					transferItems.push(item);
				} else {
					if (e.status == TransferStatus.Error) {
						if (e.id) await this.cancel(e);
						transferItems.push(item);
					}
				}
			}

			const folderTask = task;

			const folderTotalCount = transferItems.length;

			const size = transferItems.reduce((accumulator, item) => {
				return accumulator + item.size;
			}, 0);

			folderTask.status = transferManager.checkStatus(transferItems);

			folderTask.size = folderTask.size + size;

			this.transferMap[folderTask.id as number] = folderTask;
			this.filesInDialogMap[folderTask.id as number] = folderTask;

			await transferManager.calculateFolderItemTotalCount(
				task,
				'set',
				(folderTask.folderTotalCount || 0) + folderTotalCount
			);

			const subTask: TransferItem[] = [];
			for (const item of transferItems) {
				subTask.push({
					task: task.id,
					name: item.name,
					path: item.path,
					parentPath: item.parentPath ? item.parentPath : task.path,
					type: item.type,
					isFolder: false,
					driveType: item.driveType,
					front: task.front,
					status: TransferStatus.Pending,
					url: item.url,
					startTime: new Date().getTime(),
					endTime: 0,
					updateTime: new Date().getTime(),
					from: task.front == TransferFront.download ? item.path : item.from,
					to: task.front == TransferFront.download ? '' : item.path,
					isPaused: false,
					size: item.size,
					message: '',
					uniqueIdentifier: item.uniqueIdentifier,
					repo_id: item.repo_id,
					userId: task.userId,
					relatePath: item.relatePath,
					params: item.params,
					node: item.node,
					currentPhase: item.currentPhase ? item.currentPhase : 1,
					totalPhase: item.totalPhase ? item.totalPhase : 1,
					pauseDisable:
						item.pauseDisable != undefined ? item.pauseDisable : false
				});
			}

			const ids = (await transferManager.db.transferData.bulkAdd(
				subTask,
				undefined,
				{
					allKeys: true
				}
			)) as any as number[];

			if (childPrepare.callback) {
				const identifys = transferItems.map((e) => e.uniqueIdentifier || '');
				childPrepare.callback(true, ids, identifys);
			}

			if (childPrepare.finished) {
				await this.update(folderTask.id!, {
					size: folderTask.size
				});
				if (this.transferMap[taskId].folderTotalCount == 0) {
					await this.updateTaskStatus(taskId, TransferStatus.Completed);
				} else {
					await this.updateTaskStatus(taskId, TransferStatus.Pending);
					this.transferMap[taskId].isPaused = false;
				}
			} else {
				if (this.transferMap[taskId].status != TransferStatus.Prepare) {
					await this.updateTaskStatus(taskId, TransferStatus.Prepare);
				}
				setTimeout(() => {
					this.prepare(taskId);
				}, 1000);
			}
		},

		async startFileInTask(item: TransferItemInMemory) {
			TransferClient.doAction(item, 'start');
		},

		async runNextFileInTask(task: number, front: TransferFront) {
			const runningTasks = this.getFolderRunningTasks(task, front);

			const nextAddCount =
				forderMaxLength > runningTasks.length
					? forderMaxLength - runningTasks.length
					: 0;

			const nextItems = await transferManager.db.transferData
				.where('task')
				.equals(task)
				.and(
					(item) =>
						item.status === TransferStatus.Pending &&
						item.isPaused === false &&
						!runningTasks
							.filter(
								(e) => e.status == TransferStatus.Running && e.isPaused == false
							)
							.map((e) => e.id)
							.includes(item.id)
				)
				.limit(nextAddCount)
				.toArray();
			if (nextItems.length > 0) {
				for (let index = 0; index < nextItems.length; index++) {
					const t = nextItems[index];
					const item = convertTransferItemToMemeryItem(t);
					if (item.task && item.task > 0) {
						this.filesInFolderMap[item.id as number] = {
							...item,
							status: TransferStatus.Checking
						};
					}
					transferManager.calFolderTask(item, 'add');
					item.status = TransferStatus.Running;
					await this.startFileInTask(item);
				}
			}

			if (nextAddCount > 0 && nextItems.length == 0) {
				const nextItemTask = await transferManager.db.transferData
					.where('task')
					.equals(task)
					.and(
						(item) =>
							item.status !== TransferStatus.Completed &&
							item.status !== TransferStatus.Canceled
					)
					.limit(1)
					.toArray();

				const currentTasks: TransferItemInMemory[] =
					this.taskCurrentSingleFiles[front][task] || [];

				if (currentTasks.length == 0 && nextItemTask.length == 0) {
					this.updateTaskStatus(task, TransferStatus.Completed);
					delete this.taskCurrentSingleFiles[front][task];
				} else {
					const runningTasks = this.getFolderRunningTasks(task, front);

					if (runningTasks.length > 0) {
						return;
					}
					this.updateTaskStatus(task, TransferStatus.Pending);
					this.pausedOrResumeTaskStatus(task, true);
				}
			}
		},

		async run(item: TransferItemInMemory) {
			if (!transferManager.startRunEnable(item)) {
				return;
			}
			if (
				item.status === TransferStatus.Completed ||
				item.status === TransferStatus.Removing ||
				item.status === TransferStatus.Removed ||
				item.status === TransferStatus.Canceling ||
				item.status === TransferStatus.Canceled
			) {
				throw Error('Task already completed');
			}

			if (
				item.status === TransferStatus.Checking ||
				item.status === TransferStatus.Resuming
			) {
				return;
			}

			item.isPaused = false;
			item.status = TransferStatus.Running;
			console.log('run --->', item);

			if (item.isFolder) {
				await this.runNextFileInTask(item.id!, item.front);
			} else {
				await this.startFileInTask(item);
			}
			this.transferMap[item.id!] = item;
			if (this.filesInDialogMap[item.id!]) {
				this.filesInDialogMap[item.id!] = item;
			}
		},

		check(item: TransferItem) {
			console.log('check', item);
		},

		async pause(item: TransferItem) {
			if (!item.id) {
				return;
			}
			await transferManager.pause(item);
		},

		bulkPause(ids: number[]) {
			transferManager.bulkPause(ids, this.activeItem);
		},

		async resume(item: TransferItem) {
			if (!item.id) {
				return;
			}
			await transferManager.resume(item);
		},

		bulkResume(ids: number[]) {
			transferManager.bulkResume(ids, this.activeItem);
		},

		async cancel(item: TransferItem) {
			transferManager.cancel(item);
		},

		bulkCancel(ids: number[]) {
			transferManager.bulkCancel(ids, this.activeItem);
		},

		async remove(id: number) {
			await transferManager.removeById(id, this.activeItem);
		},

		async bulkRemove(ids: number[]) {
			await transferManager.bulkRemove(ids, this.activeItem);
		},

		async onFileProgress(
			id: number | undefined,
			bytes = 0,
			front: TransferFront
		) {
			const currentData = this.getSubTransferItem(front, id);

			if (currentData && id && id !== currentData.id) {
				return false;
			}

			if (!id) return false;

			const transferData = currentData ? currentData : this.transferMap[id];

			const result = updateMemoryItemInfo(transferData, bytes);
			if (!result) return false;
			if (transferData.task && transferData.task > 0) {
				transferManager.calFolderTask(transferData, 'add');
				this.filesInFolderMap[transferData.id as number] = transferData;
				await this.onFolderProgress(transferData);
			} else {
				this.transferMap[id] = transferData;
				if (this.filesInDialogMap[id]) {
					this.filesInDialogMap[id] = transferData;
				}
			}
			return true;
		},

		async onFolderProgress(item: TransferItemInMemory) {
			if (!item.task || !this.transferMap[item.task]) {
				return;
			}
			const runningTasks: TransferItemInMemory[] =
				this.taskCurrentSingleFiles[item.front][item.task as number] || [];

			updateFolderMemoryItemInfo(
				this.transferMap[item.task],
				item,
				runningTasks
			);

			if (this.filesInDialogMap[item.task]) {
				this.filesInDialogMap[item.task] = this.transferMap[item.task];
			}
		},

		async replaceFileInFolder(id: number) {
			if (id in this.filesInFolderMap) {
				const task = this.transferMap[this.filesInFolderMap[id].task as number];
				await transferManager.calculateFolderItemCompletedCount(task, 'add', 1);
			}
		},

		async onFileComplete(
			id: number,
			front: TransferFront,
			phase = 1,
			nextPhaseTaskId?: string
		) {
			transferManager.onFileComplete(id, front, phase, nextPhaseTaskId);
		},

		async recoverErrorTransfer(id: number) {
			this.updateTaskStatus(id, TransferStatus.Pending);
			if (id in this.filesInFolderMap) {
				this.filesInFolderMap[id].retryCount = 0;

				const currentSingleFile = this.getSubTransferItem(
					this.filesInFolderMap[id].front,
					id,
					this.filesInFolderMap[id].task
				);
				if (currentSingleFile) {
					currentSingleFile.retryCount = 0;
				}

				if (this.transferMap[this.filesInFolderMap[id].task!].isPaused) {
					this.transferMap[this.filesInFolderMap[id].task!].isPaused = false;
				}
			} else if (id in this.transferMap) {
				this.transferMap[id].retryCount = 0;
			}
		},

		async onFileError(id: number, front: TransferFront, message?: string) {
			return await transferManager.onFileError(id, front, message);
		},

		async updateTaskStatus(id: number, status: TransferStatus) {
			await transferManager.updateTaskStatus(id, status);
			if (this.transferMap[id]) {
				this.transferMap[id].status = status;
			}
			if (this.filesInDialogMap[id]) {
				this.filesInDialogMap[id].status = status;
			}

			if (this.filesInFolderMap[id]) {
				this.filesInFolderMap[id].status = status;
			}
		},

		async pausedOrResumeTaskStatus(id: number, isPaused: boolean) {
			await transferManager.updateTaskPaused(id, isPaused);
			if (this.transferMap[id]) {
				this.transferMap[id].isPaused = isPaused;
			}
			if (this.filesInDialogMap[id]) {
				this.filesInDialogMap[id].isPaused = isPaused;
			}
			if (this.filesInFolderMap[id]) {
				this.filesInFolderMap[id].isPaused = isPaused;
			}
		},

		async getFilesInTask(taskId: number) {
			const ids: number[] = [];
			const result = await transferManager.db.transferData
				.where('task')
				.equals(taskId)
				.toArray();

			for (const item of result) {
				if (!this.filesInFolder.includes(item.id!)) {
					ids.push(item.id!);
				}

				const currentSingleFile = this.getSubTransferItem(
					item.front,
					item.id,
					taskId
				);

				if (currentSingleFile) {
					this.filesInFolderMap[item.id!] = currentSingleFile;
				} else {
					this.filesInFolderMap[item.id!] =
						convertTransferItemToMemeryItem(item);
				}
			}

			this.filesInFolder = ids;
		},

		updateTransferItemsSpeed(
			runningTask: TransferItemInMemory[],
			front: TransferFront,
			forceRefresh = false
		) {
			for (let index = 0; index < runningTask.length; index++) {
				const task = runningTask[index];

				if (task.isFolder) {
					const tasks: TransferItemInMemory[] =
						this.taskCurrentSingleFiles[front][runningTask[0].id];
					if (tasks) {
						tasks.forEach((singleFile) => {
							const lateUpdateTime =
								singleFile.updateTime ||
								singleFile.startTime ||
								new Date().getTime();
							const times = new Date().getTime() - lateUpdateTime;
							if ((times > 2000 || forceRefresh) && singleFile.id) {
								this.onFileProgress(singleFile.id, singleFile.bytes, front);
							}
						});
					}
				} else {
					const lateUpdateTime =
						runningTask[0].updateTime ||
						runningTask[0].startTime ||
						new Date().getTime();
					const times = new Date().getTime() - lateUpdateTime;
					if (runningTask[0].id && (times > 2000 || forceRefresh)) {
						this.onFileProgress(runningTask[0].id, runningTask[0].bytes, front);
					}
				}
			}
		},

		updateRunningItemsStatus() {
			const deviceStore = useDeviceStore();
			const items = getCurrentTransferingItems();
			if (!deviceStore.transferEnable()) {
				items.forEach((item) => {
					if (!deviceStore.networkOnLine) {
						item.networkOfflinePaused = true;
						item.onlyWifiPaused = false;
					} else {
						item.networkOfflinePaused = false;
						item.onlyWifiPaused = true;
					}

					if (!item.isFolder) {
						TransferClient.doAction(item, 'pause');
						this.updateTransferItemsSpeed([item], item.front);
					} else {
						if (this.taskCurrentSingleFiles[item.front][item.id]) {
							const tasks: TransferItemInMemory[] =
								this.taskCurrentSingleFiles[item.front][item.id];
							if (tasks) {
								tasks.forEach((task) => {
									TransferClient.doAction(task, 'pause');
								});
								this.updateTransferItemsSpeed(tasks, item.front, true);
							}
						}
					}
				});
			} else {
				items.forEach((item) => {
					if (item.networkOfflinePaused || item.onlyWifiPaused) {
						item.networkOfflinePaused = false;
						item.onlyWifiPaused = false;
						if (!item.isFolder) {
							item.retryCount = 0;
							this.onFileError(item.id!, item.front, 'need resume');
						} else {
							if (this.taskCurrentSingleFiles[item.front][item.id]) {
								// this.taskCurrentSingleFiles[item.front][item.id].retryCount = 0;

								const tasks: TransferItemInMemory[] =
									this.taskCurrentSingleFiles[item.front][item.id];
								if (tasks) {
									tasks.forEach((task) => {
										task.retryCount = 0;
										this.onFileError(task.id!, task.front, 'need resume');
									});
								}
							}
						}
					}
				});
			}
		},

		updateRunningItemsBaseUrl() {
			const items = getCurrentTransferingItems();
			items.forEach(async (item) => {
				if (!item.isFolder) {
					await TransferClient.doAction(item, 'pause');
					item.retryCount = 0;
					await this.onFileError(item.id!, item.front, 'base url update error');
				} else {
					if (this.taskCurrentSingleFiles[item.front][item.id]) {
						const tasks: TransferItemInMemory[] =
							this.taskCurrentSingleFiles[item.front][item.id];
						if (tasks) {
							tasks.forEach(async (task) => {
								await TransferClient.doAction(task, 'pause');
								task.retryCount = 0;
								this.onFileError(task.id!, task.front, 'base url update error');
							});
						}
					}
				}
			});
		},

		calFolderTask(transferItem: TransferItemInMemory, operate: 'add' | 'sub') {
			return transferManager.calFolderTask(transferItem, operate);
		},

		async clearTransferData() {
			try {
				await transferManager.db.transferData.clear();
				console.log('table transferData clear');
			} catch (error) {
				console.error('clear transferData error:', error);
			}
		},

		getSubTransferItem(front: TransferFront, id?: number, taskId?: number) {
			const items: Record<number, TransferItemInMemory[]> | undefined =
				this.taskCurrentSingleFiles[front];

			if (!items) {
				return undefined;
			}
			if (taskId) {
				const subTasks = items[taskId];
				if (subTasks) {
					return subTasks.find((e) => e.id == id);
				}
				return undefined;
			}
			const runningTasks = Object.keys(items);
			for (let index = 0; index < runningTasks.length; index++) {
				const element = runningTasks[index];
				const subTasks = items[Number(element)];
				const item = subTasks.find((e) => e.id == id);
				if (item) {
					return item;
				}
			}
		},

		getFolderRunningTasks(taskId: number, front: TransferFront) {
			const items: Record<number, TransferItemInMemory[]> | undefined =
				this.taskCurrentSingleFiles[front];
			if (!items) {
				return [];
			}
			const tasks = items[taskId];

			if (!tasks) {
				return [];
			}
			return tasks.filter(
				(e) => e.status == TransferStatus.Running && e.isPaused == false
			);
		}
	}
});
