/* eslint-disable @typescript-eslint/no-unused-vars */
import { useTransfer2Store } from 'src/stores/transfer2';
import {
	TransferClientService,
	ClouderTransferStatus,
	CloudFileInfo
} from '../abstractions/transfer/interface';
import Resumable from '../../utils/resumejs';
import { busOn } from 'src/utils/bus';
import { useUserStore } from 'src/stores/user';
import { DOWNLOAD_OPERATE } from 'src/utils/rss-types';
import { DownloadRecordOperateRequest } from 'src/api/wise';
import { axiosInstanceProxy } from 'src/platform/httpProxy';
import { getAppPlatform } from 'src/application/platform';
import { useWebsocketManager2Store } from 'src/stores/websocketManager2';
import { DownloadRecord } from 'src/utils/interface/rss';
import {
	TransferFront,
	TransferItem,
	TransferStatus
} from 'src/utils/interface/transfer';
import { useFilesCopyStore } from 'src/stores/files-copy';
import { useAppAbilitiesStore } from 'src/stores/appAbilities';
import Taskmanager from '../olaresTask';
import { convertTransferItemToMemeryItem } from 'src/services/abstractions/transferManager/interface';

export const commonDownloader = {
	async start(item: TransferItem) {
		const transferStore = useTransfer2Store();
		const iframe = document.createElement('iframe');
		iframe.style.display = 'none';
		iframe.src = item.url!;
		document.body.appendChild(iframe);
		setTimeout(() => iframe.remove(), 30000);
		transferStore.updateTaskStatus(item.id!, TransferStatus.Completed);
		return true;
	},
	cancel: function (_item: TransferItem): Promise<boolean> {
		throw new Error('Function not implemented.');
	},
	pause: function (_item: TransferItem): Promise<boolean> {
		throw new Error('Function not implemented.');
	},
	resume: function (_item: TransferItem): Promise<boolean> {
		throw new Error('Function not implemented.');
	},
	complete: async function (_item: TransferItem): Promise<boolean> {
		return true;
	},
	getTransferInfo: async function (
		_item: TransferItem
	): Promise<{ id: number; bytes: number } | undefined> {
		return undefined;
	},
	restartEnable: async function (_item: TransferItem): Promise<boolean> {
		return false;
	},
	restartAutoResume: false
};
// resume js
export const commonUploader = {
	async start(uploadingItem: TransferItem) {
		const item = Resumable.getFiles().find((e) => e.id == uploadingItem.id);
		console.log('item ===>', item?.isPaused());

		if (!item || !item.isPaused()) {
			return true;
		}
		return await this.resume(uploadingItem);
	},
	cancel: async function (uploadingItem: TransferItem): Promise<boolean> {
		if (uploadingItem.phaseTaskId !== undefined) {
			return await Taskmanager.doAction(uploadingItem, 'cancel');
		}
		const file = Resumable.getFiles().find((e) => e.id == uploadingItem.id);
		if (file) {
			file.cancel();
		}

		return true;
	},
	pause: async function (uploadingItem: TransferItem): Promise<boolean> {
		if (uploadingItem.phaseTaskId !== undefined) {
			return await Taskmanager.doAction(uploadingItem, 'pause');
		}
		const file = Resumable.getFiles().find((e) => e.id == uploadingItem.id);
		if (!file) {
			return false;
		}
		if (!file.isPaused()) {
			file.pause();
		}
		return true;
	},
	resume: async function (uploadingItem: TransferItem): Promise<boolean> {
		if (uploadingItem.phaseTaskId !== undefined) {
			return await Taskmanager.doAction(uploadingItem, 'resume');
		}
		const file = Resumable.getFiles().find((e) => e.id == uploadingItem.id);
		if (!file) {
			return true;
		}

		if (file.isPaused()) {
			file._pause = false;
			file.upload();
		} else {
			file.retry();
		}
		return true;
	},
	complete: async function (_item: TransferItem): Promise<boolean> {
		return true;
	},
	getTransferInfo: async function (
		_item: TransferItem
	): Promise<{ id: number; bytes: number } | undefined> {
		return undefined;
	},

	addSubtasksSuccess: async function (
		_offset: number,
		taskId: number,
		subtaskIds: number[],
		identifys?: string[]
	) {
		Resumable.updateSubtasksSuccess(taskId, subtaskIds, identifys);
	},

	restartEnable: async function (_item: TransferItem): Promise<boolean> {
		return false;
	},

	restartAutoResume: false
};

const cloudHistoryIdentify = 'cloud-transfer_';

export const commonClouder = {
	init() {
		busOn(
			'CloudTransferUpdate',
			async (data: { item: TransferItem; data: DownloadRecord }) => {
				console.log('item ===>', data.item);

				console.log('item.id', data.data);

				this.updateTransferItemStatus(data.item, data.data);
			}
		);

		busOn('account_update', () => {
			const userStore = useUserStore();
			if (!userStore.current_user?.setup_finished) {
				return;
			}
			this.syncCloudData();
		});
	},

	async start() {
		return true;
	},
	cancel: async function (): Promise<boolean> {
		return true;
	},
	pause: async function (item: TransferItem): Promise<boolean> {
		if (!item.uniqueIdentifier) {
			return false;
		}
		const taskId = this.getTaskIdByIdentify(item.uniqueIdentifier);
		await this.taskOperate(taskId, DOWNLOAD_OPERATE.PAUSE, false);
		return true;
	},
	resume: async function (item: TransferItem): Promise<boolean> {
		if (!item.uniqueIdentifier) {
			return false;
		}
		const taskId = this.getTaskIdByIdentify(item.uniqueIdentifier);
		await this.taskOperate(taskId, DOWNLOAD_OPERATE.RETRY, false);
		return true;
	},
	complete: async function (_item: TransferItem): Promise<boolean> {
		return true;
	},
	getTransferInfo: async function (
		_item: TransferItem
	): Promise<{ id: number; bytes: number } | undefined> {
		return undefined;
	},
	restartEnable: async function (item: TransferItem): Promise<boolean> {
		const transferStore = useTransfer2Store();
		if (
			item.uniqueIdentifier &&
			item.id &&
			!transferStore.filesCloudTransferMap[item.uniqueIdentifier]
		) {
			transferStore.filesCloudTransferMap[item.uniqueIdentifier] = item.id;
		}
		return true;
	},
	async taskOperate(
		taskId: string,
		operate: DOWNLOAD_OPERATE,
		removeFile?: boolean
	) {
		const instance = this.getTaskInstance();
		const req = new DownloadRecordOperateRequest(
			taskId + '',
			operate,
			removeFile
		);
		try {
			const history = await instance.get(
				'/knowledge/download/' + req.toString()
			);
			console.log(history);
			return true;
		} catch (e: any) {
			console.log(e.message);
			return false;
		}
	},
	getTaskInstance() {
		const userStore = useUserStore();
		let baseURL = userStore.getModuleSever('wise');
		if (userStore.current_user?.isLargeVersion12) {
			const abilitiesStore = useAppAbilitiesStore();
			baseURL = userStore.getModuleSever(abilitiesStore.wise.id);
		}

		const instance = axiosInstanceProxy({
			headers: {
				'Content-Type': 'application/json',
				Accept: 'application/json'
			},
			baseURL: process.env.DEV ? '' : baseURL
		});
		return instance;
	},
	async queryUrl(url: string): Promise<CloudFileInfo | undefined> {
		try {
			const instance = this.getTaskInstance();
			const info: any = await instance.get(
				'/knowledge/download/larepass_file_query?url=' + encodeURIComponent(url)
			);
			console.log(info);
			return info.data.data;
		} catch (e: any) {
			console.log(e.message);
			return undefined;
		}
	},
	async downloadFile(
		name: string,
		download_url: string,
		larepass_id: string,
		path: string,
		file_type: string
	): Promise<DownloadRecord | undefined> {
		try {
			console.log(file_type);
			const instance = this.getTaskInstance();
			const result: any = await instance.post(
				'/knowledge/download/larepass_file_download',
				{
					url: download_url,
					name: name,
					larepass_id,
					path
				}
			);

			if (result && result.data && result.data.data.length > 0) {
				return result.data.data[0];
			}

			return undefined;
		} catch (e: any) {
			return undefined;
		}
	},

	async removeTask(item: TransferItem, deleteFile: boolean) {
		if (!item.uniqueIdentifier) {
			return false;
		}
		const taskId = this.getTaskIdByIdentify(item.uniqueIdentifier);
		return await this.taskOperate(taskId, DOWNLOAD_OPERATE.REMOVE, deleteFile);
	},

	async cancelTask(item: TransferItem) {
		if (!item.uniqueIdentifier) {
			return false;
		}
		const taskId = this.getTaskIdByIdentify(item.uniqueIdentifier);
		return await this.taskOperate(taskId, DOWNLOAD_OPERATE.CANCEL, false);
	},

	async cancelAllTask() {
		const instance = this.getTaskInstance();
		const userStore = useUserStore();
		try {
			const result = await instance.put(
				'/knowledge/download/larepass_cancel_all_task',
				{
					larepass_id: userStore.id!
				}
			);
			console.log('cancel all task result ===>', result);

			return true;
		} catch (e: any) {
			console.log(e.message);
			return false;
		}
	},

	async removeAllTask(removeFlag = false) {
		const instance = this.getTaskInstance();
		const userStore = useUserStore();
		try {
			const result = await instance.put(
				'/knowledge/download/larepass_remove_all_task',
				{
					larepass_id: userStore.id!,
					removeFlag: removeFlag
				}
			);
			console.log('remove all task result ===>', result);
			return true;
		} catch (e: any) {
			console.log(e.message);
			return false;
		}
	},

	async getDownloadHistory(
		update_time: number,
		larepass_id: string
	): Promise<DownloadRecord[]> {
		try {
			const instance = this.getTaskInstance();
			const history: any = await instance.get(
				'/knowledge/download/larepass_task_query',
				{
					params: {
						larepass_id,
						update_time
					}
				}
			);
			console.log('history ===>', history);
			// return history;
			return history &&
				history.data &&
				history.data.data &&
				history.data.data.items
				? history.data.data.items
				: [];
		} catch (e: any) {
			console.log(e.message);
			return [];
		}
	},

	taskIdentify(taskId: string) {
		return this.taskBaseIdentify() + '_' + `${taskId}`;
	},

	getTaskIdByIdentify(uniqueIdentifier?: string) {
		const userStore = useUserStore();
		console.log('getTaskIdByIdentify start====>', uniqueIdentifier);
		console.log('userStore.current_id ===>', userStore.current_id);

		if (
			!uniqueIdentifier ||
			(userStore.current_id &&
				!uniqueIdentifier.startsWith(userStore.current_id))
		) {
			return '';
		}

		const result = uniqueIdentifier.substring(
			((userStore.current_id || '') + '_').length
		);
		console.log('getTaskIdByIdentify end====>', result);

		return result;
	},

	async syncCloudData() {
		const transferStore = useTransfer2Store();
		if (transferStore.isIniting) {
			setTimeout(() => {
				this.syncCloudData();
			}, 500);
			return;
		}
		const userStore = useUserStore();
		// const transferStore = useTransfer2Store();
		let info = await this.getCurrentUserCloudTransferSaveInfo();
		if (!info) {
			info = {
				terminusID: userStore.current_user!.id,
				timer: 0
			};
		}
		let lastTimer = 0;
		if (info.terminusID != userStore.current_user?.olares_device_id) {
			//remove old data
			info = {
				terminusID: userStore.current_user!.olares_device_id,
				timer: 0
			};
			await this.saveCurrentUserCloudTransferSaveInfo(info.timer);
		} else {
			lastTimer = info.timer;
		}
		// const result = await this.getDownloadHistory(lastTimer, userStore.id!);

		let data: DownloadRecord[] = [];
		do {
			data = await this.getDownloadHistory(lastTimer, userStore.id!);
			data = data.sort((a, b) => {
				return b.update_time > a.update_time
					? 1
					: b.update_time < a.update_time
					? -1
					: 0;
			});
			for (const downloadItem of data) {
				const time = new Date(downloadItem.update_time).getTime();
				if (time > lastTimer) {
					lastTimer = time;
					await this.saveCurrentUserCloudTransferSaveInfo(time);
				}
				const tansferStore = useTransfer2Store();
				const taskIdentify = this.taskIdentify(`${downloadItem.id}`);
				const transfeId =
					transferStore.filesCloudTransferMap[taskIdentify] || -1;
				if (transfeId >= 0) {
					this.updateTransferItemStatus(
						tansferStore.transferMap[transfeId],
						downloadItem
					);
				} else {
					const socketStore = useWebsocketManager2Store();
					socketStore.apply('addTaskHistory', downloadItem);
				}
			}
		} while (data.length >= 100);
	},

	async getCurrentUserCloudTransferSaveInfo() {
		const userStore = useUserStore();
		if (!userStore.current_id) {
			return undefined;
		}
		let info = await getAppPlatform().userStorage.getItem(
			cloudHistoryIdentify + userStore.current_id
		);
		if (typeof info == 'string') {
			info = JSON.parse(info);
		}
		return info as { terminusID: string; timer: number };
	},
	async saveCurrentUserCloudTransferSaveInfo(timer: number) {
		const userStore = useUserStore();
		const saveInfo = {
			terminusID: userStore.current_user!.olares_device_id,
			timer
		};
		await getAppPlatform().userStorage.setItem(
			cloudHistoryIdentify + userStore.current_id,
			JSON.stringify(saveInfo)
		);
		return saveInfo;
	},
	async updateTransferItemStatus(
		item: TransferItem,
		downloadItem: DownloadRecord
	) {
		const transferStore = useTransfer2Store();

		const taskIdentify = this.taskIdentify(`${downloadItem.id}`);

		let downloadPath = downloadItem.path;
		if (downloadPath != undefined) {
			if (!downloadPath.startsWith('/')) {
				downloadPath = '/' + downloadPath;
			}

			if (!downloadPath.endsWith('/')) {
				downloadPath = downloadPath + '/';
			}

			downloadPath = '/Files/Home' + downloadPath;
		}
		if (!item.id) {
			return;
		}

		if (!transferStore.transferMap[item.id]) {
			transferStore.transferMap[item.id] =
				convertTransferItemToMemeryItem(item);
			transferStore.transfers.push(item.id);
			transferStore.filesCloudTransferMap[taskIdentify] = item.id;
		}

		if (Number(downloadItem.progress) > 100) {
			downloadItem.progress = '99';
		}

		if (
			downloadItem.size &&
			downloadItem.size > 0 &&
			transferStore.transferMap[item.id].size != downloadItem.size
		) {
			transferStore.transferMap[item.id].size = downloadItem.size;
			transferStore.update(item.id, {
				size: downloadItem.size
			});
		}

		if (
			Number(downloadItem.downloaded_bytes) > downloadItem.size &&
			Number(downloadItem.progress) > 0 &&
			Number(downloadItem.progress) < 100
		) {
			transferStore.transferMap[item.id].size =
				(Number(downloadItem.downloaded_bytes) * 1.0 * 100) /
				Number(downloadItem.progress);
			transferStore.update(item.id, {
				size: transferStore.transferMap[item.id].size
			});
		}

		if (downloadItem.status == ClouderTransferStatus.DOWNLOADING) {
			transferStore.transferMap[item.id].status = TransferStatus.Running;
			if (
				Number(downloadItem.progress) == 0 &&
				transferStore.transferMap[item.id].size == 0
			) {
				transferStore.transferMap[item.id].bytes = Number(
					downloadItem.downloaded_bytes
				);
			} else {
				transferStore.onFileProgress(
					item.id,
					Number(downloadItem.downloaded_bytes),
					TransferFront.cloud
				);
			}
		} else if (downloadItem.status == ClouderTransferStatus.COMPLETE) {
			if (
				transferStore.transferMap[item.id].path !=
					downloadPath + downloadItem.name ||
				transferStore.transferMap[item.id].name != downloadItem.name
			) {
				transferStore.transferMap[item.id].path =
					downloadPath + downloadItem.name;
				transferStore.transferMap[item.id].name = downloadItem.name;
				transferStore.update(item.id, {
					path: transferStore.transferMap[item.id].path,
					name: downloadItem.name
				});
			}
			transferStore.onFileComplete(item.id, TransferFront.cloud);
		} else if (downloadItem.status == ClouderTransferStatus.PAUSE) {
			transferStore.pausedOrResumeTaskStatus(item.id, true);
		} else if (downloadItem.status == ClouderTransferStatus.REMOVE) {
			transferStore.bulkRemove([item.id]);
		} else if (downloadItem.status == ClouderTransferStatus.CANCEL) {
			if (transferStore.transferMap[item.id]) {
				transferStore.transferMap[item.id].status = TransferStatus.Canceled;
				transferStore.update(item.id, {
					status: TransferStatus.Canceled
				});
			}
		} else if (downloadItem.status == ClouderTransferStatus.ERROR) {
			transferStore.onFileError(item.id, TransferFront.cloud);
		} else if (downloadItem.status == ClouderTransferStatus.WAITING) {
			transferStore.updateTaskStatus(item.id, TransferStatus.Pending);
		}
	},
	getQueryId(): string {
		const userStore = useUserStore();
		return userStore.id ? userStore.id : '';
	},
	taskBaseIdentify(): string {
		const userStore = useUserStore();
		return userStore.current_id || '';
	}
};

//  copy/move
export const commonCopier = {
	async start(_item: TransferItem) {
		return true;
	},
	cancel: async function (item: TransferItem): Promise<boolean> {
		return await Taskmanager.doAction(item, 'cancel');
	},
	pause: function (_item: TransferItem): Promise<boolean> {
		throw new Error('Function not implemented.');
	},
	resume: function (_item: TransferItem): Promise<boolean> {
		throw new Error('Function not implemented.');
	},
	complete: async function (_item: TransferItem): Promise<boolean> {
		return true;
	},
	getTransferInfo: async function (
		_item: TransferItem
	): Promise<{ id: number; bytes: number } | undefined> {
		return undefined;
	},
	restartEnable: async function (_item: TransferItem): Promise<boolean> {
		return false;
	}
};

export class CommonTransfer implements TransferClientService {
	downloader = commonDownloader;
	uploader = commonUploader;
	clouder = commonClouder;
	copier = commonCopier;
	restartAutoResume = false;
	errorRetryNumber = 1;
}
