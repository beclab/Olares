import { useTransfer2Store } from 'src/stores/transfer2';
import {
	getUploadType,
	TransferClientService
} from '../abstractions/transfer/interface';
import {
	IUploadStartFile,
	IDownloadStartFile
} from 'src/platform/interface/electron/interface';
import { Platform } from 'quasar';
import { useUserStore } from 'src/stores/user';
import { dataAPIs } from 'src/api';
// import { DriveType } from 'src/stores/files';
import { commonClouder } from './common';
import { compareUrlHost, replaceUrlHost } from 'src/utils/file';
import { TransferItem } from 'src/utils/interface/transfer';
import { filesIsV2 } from 'src/api/files';
import { dataAPIs as dataAPIsV2 } from 'src/api/files/v2';
import { busEmit } from 'src/utils/bus';
export class ElectronTransfer implements TransferClientService {
	downloader = {
		start: async function (item: TransferItem): Promise<boolean> {
			if (!item.id) {
				return false;
			}
			let url = item.url;
			const baseUrl = getFileBaseUrl(item);
			if (!compareUrlHost(item.url!, baseUrl)) {
				url = replaceUrlHost(item.url!, baseUrl);
			}
			const data: IDownloadStartFile = {
				id: item.id,
				url: url!,
				fileName: item.name,
				path: '',
				size: item.size
			};

			const transferStore = useTransfer2Store();
			const appendPath = Platform.is.win ? '\\' : '/';
			const transferInfo = await window.electron.api.download2.getDownloadInfo(
				item.id!
			);
			if (transferInfo) {
				return await this.resume(item);
			} else {
				if (!data.path) {
					if (item.task && item.task > 0) {
						const savePath =
							await window.electron.api.download2.getFolderSavePath(
								item.task,
								transferStore.transferMap[item.task].name
							);
						const defaultDownloadPath =
							await window.electron.api.download2.getDownloadPath();

						const dataAPI = dataAPIs(item.driveType);
						const { parentSavePath, itemSavePath } =
							dataAPI.formatFolderSubItemDownloadPath(
								item,
								transferStore.transferMap[item.task],
								savePath,
								defaultDownloadPath,
								appendPath
							);
						data.path = itemSavePath;

						transferStore.update(item.task, {
							to: parentSavePath
						});

						transferStore.transferMap[item.task].to = parentSavePath;
					} else {
						data.path = await window.electron.api.download2.getDownloadPath();
					}
				}
				await window.electron.api.download2.start(data);
				return true;
			}
		},
		cancel: async function (item: TransferItem): Promise<boolean> {
			const transferInfo = await this.getTransferInfo(item);
			if (!transferInfo) {
				return true;
			}
			return await window.electron.api.download2.cancel(item.id!);
		},
		pause: async function (item: TransferItem): Promise<boolean> {
			const transferInfo = await this.getTransferInfo(item);
			if (!transferInfo) {
				return true;
			}
			return await window.electron.api.download2.pause(item.id!);
		},
		resume: async function (item: TransferItem): Promise<boolean> {
			const transferInfo = await this.getTransferInfo(item);
			if (!transferInfo) {
				return true;
			}
			const options = {
				id: item.id!,
				url: item.url!
			};
			const baseUrl = getFileBaseUrl(item);
			if (!compareUrlHost(item.url!, baseUrl)) {
				options.url = replaceUrlHost(item.url!, baseUrl);
			}
			return await window.electron.api.download2.resume(
				options.id,
				options.url
			);
		},

		complete: async function (item: TransferItem): Promise<boolean> {
			if (!item.isFolder) {
				return true;
			}
			return await window.electron.api.download2.deleteFolderSavePath(item.id!);
		},

		getTransferInfo: async function (
			item: TransferItem
		): Promise<{ id: number; bytes: number } | undefined> {
			return await window.electron.api.download2.getDownloadInfo(item.id!);
		},

		restartEnable: async function (): Promise<boolean> {
			return true;
		},
		restartAutoResume: false
	};
	uploader = {
		async start(item: TransferItem) {
			try {
				const transferInfo = await window.electron.api.upload.getUploadInfo(
					item.id!
				);

				const userStore = useUserStore();
				const baseUrl = userStore.getModuleSever('files');

				if (transferInfo) {
					return await window.electron.api.upload.resume(item.id!, baseUrl);
				}

				const dataAPI = dataAPIs(item.driveType);

				if (item.size == 0 && filesIsV2()) {
					const apiV2 = dataAPIsV2(item.driveType);
					apiV2.uploadEmptyFile(item.id!).then(() => {
						window.electron.api.upload.clearData(item.id!);
						busEmit('fileUploadComleted', {
							code: 0,
							message: '',
							path: '',
							id: item.id,
							taskId: item.task
						});
					});
					return true;
				}

				let uploadPath = await dataAPI.formatUploadTransferPath(item);

				try {
					uploadPath = decodeURIComponent(uploadPath);
				} catch (error) {
					console.log('error ===>', error);
				}

				const moreInfo = dataAPI.getUploadTransferItemMoreInfo(item);

				const uploadOptions: IUploadStartFile = {
					id: item.id!,
					baseUrl,
					// uploadPath: uploadPath,
					uploadPath: uploadPath,
					filePath: item.from!,
					fileType: getUploadType(item),
					repoId: item.repo_id,
					uniqueIdentifier: item.uniqueIdentifier,
					relativePath: item.relatePath,
					size: item.size,
					node: item.node,
					moreInfo: moreInfo
				};

				return await window.electron.api.upload.start(uploadOptions);
			} catch (error) {
				return false;
			}
		},
		cancel: async function (item: TransferItem): Promise<boolean> {
			if ((await this.getTransferInfo(item)) == undefined) {
				return true;
			}
			return await window.electron.api.upload.cancel(item.id!);
		},
		pause: async function (item: TransferItem): Promise<boolean> {
			if ((await this.getTransferInfo(item)) == undefined) {
				return true;
			}
			return await window.electron.api.upload.pause(item.id!);
		},
		resume: async function (item: TransferItem): Promise<boolean> {
			if ((await this.getTransferInfo(item)) == undefined) {
				return true;
			}
			const userStore = useUserStore();
			const baseUrl = userStore.getModuleSever('files');
			return await window.electron.api.upload.resume(item.id!, baseUrl);
		},

		complete: async function (): Promise<boolean> {
			return true;
		},

		getTransferInfo: async function (
			item: TransferItem
		): Promise<{ id: number; bytes: number } | undefined> {
			return await window.electron.api.upload.getUploadInfo(item.id!);
		},
		restartEnable: async function (): Promise<boolean> {
			return true;
		},
		restartAutoResume: false
	};

	restartAutoResume = false;
	clouder = commonClouder;
	errorRetryNumber = 5;
}

function getFileBaseUrl(item: TransferItem) {
	const userStore = useUserStore();
	const baseUrl = userStore.getSelectUserModuleServer(
		'files',
		undefined,
		undefined,
		undefined,
		item.userId
	);
	return baseUrl;
}
