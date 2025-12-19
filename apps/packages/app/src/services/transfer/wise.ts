/* eslint-disable @typescript-eslint/no-unused-vars */
import { commonClouder, CommonTransfer, commonUploader } from './common';
import { axiosInstanceProxy } from '../../platform/httpProxy';
import { CloudFileInfo } from '../abstractions/transfer/interface';
import { downloadFile, queryDownloadFile } from '../../api/wise';
import { useTransferStore } from '../../stores/rss-transfer';
import { useTransfer2Store } from '../../stores/transfer2';
import { busOn } from '../../utils/bus';
import { useRssStore } from '../../stores/rss';
import { useWebsocketManager2Store } from 'src/stores/websocketManager2';
import { TransferItem } from 'src/utils/interface/transfer';
import { DownloadRecord } from 'src/utils/interface/rss';
import { useTerminusStore } from 'src/stores/terminus';

export const wiseUploader = {
	...commonUploader,
	complete: async function (_item: TransferItem): Promise<boolean> {
		console.log(_item);
		const rssTransfer = useTransferStore();
		const rssStore = useRssStore();
		const transferStore = useTransfer2Store();
		const data = await rssTransfer.uploadFile(_item.path, _item.name);
		console.log(data);
		if (data && data.upload_id && _item.id) {
			transferStore.transferMap[_item.id].wiseRecordId = data.upload_id;
			transferStore.update(_item.id, { wiseRecordId: data.upload_id });
		}
		rssStore.syncEntries();
		return true;
	}
};

const cloudHistoryIdentify = 'cloud-transfer-wise';

export const wiseClouder = {
	...commonClouder,
	init() {
		busOn(
			'CloudTransferUpdate',
			async (data: { item: TransferItem; data: DownloadRecord }) => {
				// this.updateTransferItemStatus({
				// 	...data,
				// 	id: data.task_id
				// });
				console.log(data);
				this.updateTransferItemStatus(data.item, data.data);
			}
		);
		busOn('account_update', () => {
			this.syncCloudData();
		});
	},
	getTaskInstance() {
		return axiosInstanceProxy({
			headers: {
				'Content-Type': 'application/json',
				Accept: 'application/json'
			}
		});
	},
	complete: async function (_item: TransferItem): Promise<boolean> {
		const rssStore = useRssStore();
		rssStore.syncEntries();
		return true;
	},
	async queryUrl(url: string): Promise<CloudFileInfo | undefined> {
		return await queryDownloadFile(url);
	},
	async downloadFile(
		name: string,
		download_url: string,
		larepass_id: string,
		path: string,
		file_type: string
	): Promise<DownloadRecord | undefined> {
		console.log(larepass_id);
		return downloadFile(name, file_type, download_url, path);
	},

	async getDownloadHistory(
		update_time: number,
		larepass_id: string
	): Promise<DownloadRecord[]> {
		try {
			console.log(larepass_id);
			const instance = this.getTaskInstance();
			const history: any = await instance.get(
				'/knowledge/download/larepass_task_query',
				{
					params: {
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

	getTaskIdByIdentify(uniqueIdentifier?: string) {
		const terminusStore = useTerminusStore();

		if (
			!uniqueIdentifier ||
			(terminusStore.olares_device_id &&
				!uniqueIdentifier.startsWith(terminusStore.olares_device_id))
		) {
			return '';
		}

		const result = uniqueIdentifier.substring(
			(terminusStore.olares_device_id + '_').length
		);
		console.log('getTaskIdByIdentify end====>', result);

		return result;
	},

	async syncCloudData() {
		const transferStore = useTransfer2Store();
		const terminusStore = useTerminusStore();
		if (transferStore.isIniting || !terminusStore.terminusInfo) {
			setTimeout(() => {
				this.syncCloudData();
			}, 500);
			return;
		}

		let info = await this.getCurrentUserCloudTransferSaveInfo();

		let lastTimer = 0;
		if (!info || info.terminusID != terminusStore.olares_device_id) {
			info = {
				terminusID: terminusStore.olares_device_id,
				timer: 0
			};
			await this.saveCurrentUserCloudTransferSaveInfo(info.timer);
		} else {
			lastTimer = info.timer;
		}
		let data: DownloadRecord[] = [];
		do {
			data = await this.getDownloadHistory(lastTimer, '');
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
		const info = localStorage.getItem(cloudHistoryIdentify);
		if (!info) {
			return undefined;
		}
		return JSON.parse(info) as { terminusID: string; timer: number };
	},
	async saveCurrentUserCloudTransferSaveInfo(timer: number) {
		const terminusStore = useTerminusStore();
		const saveInfo = {
			terminusID: terminusStore.terminusInfo
				? terminusStore.olares_device_id
				: '',
			timer
		};
		localStorage.setItem(cloudHistoryIdentify, JSON.stringify(saveInfo));
		return saveInfo;
	},
	getQueryId(): string {
		const terminusStore = useTerminusStore();
		return terminusStore.olares_device_id ? terminusStore.olares_device_id : '';
	},
	taskBaseIdentify(): string {
		const terminusStore = useTerminusStore();
		return terminusStore.olares_device_id || '';
	}
};

export class WiseTransfer extends CommonTransfer {
	clouder = wiseClouder;
	uploader = wiseUploader;
}
