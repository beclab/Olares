import {
	getUploadType,
	TransferClientService
} from '../abstractions/transfer/interface';
import { useUserStore } from 'src/stores/user';
import { dataAPIs } from 'src/api';
import { filesIsV2 } from 'src/api/files';
import { dataAPIs as dataAPIsV2 } from 'src/api/files/v2';
import { compareUrlHost, replaceUrlHost } from 'src/utils/file';
import { TransferItem } from 'src/utils/interface/transfer';
import { busEmit } from 'src/utils/bus';
import { FileDownloadPlugin } from 'src/platform/interface/capacitor/plugins/download';
import { FileUploadPlugin } from 'src/platform/interface/capacitor/plugins/upload';

export class MobileTransfer implements TransferClientService {
	downloader = {
		async start(item: TransferItem) {
			if (!item.id) {
				return false;
			}
			try {
				const result = await FileDownloadPlugin.getTransferInfo({
					id: item.id
				});

				const options = {
					id: item.id!,
					url: item.url
				};
				const baseUrl = getFileBaseUrl(item);
				if (!compareUrlHost(item.url!, baseUrl)) {
					options.url = replaceUrlHost(item.url!, baseUrl);
				}

				if (!result) {
					return (
						await FileDownloadPlugin.start({
							id: item.id,
							url: options.url!,
							path: item.name || item.path,
							progress: true,
							fileSize: item.size
						})
					).status;
				} else {
					return (await FileDownloadPlugin.resume(options)).status;
				}
			} catch (error) {
				return false;
			}
		},
		cancel: async function (item: TransferItem): Promise<boolean> {
			if ((await this.getTransferInfo(item)) == undefined) {
				return true;
			}
			return (await FileDownloadPlugin.cancel({ id: item.id! })).status;
		},
		pause: async function (item: TransferItem): Promise<boolean> {
			if ((await this.getTransferInfo(item)) == undefined) {
				return true;
			}
			return (await FileDownloadPlugin.pause({ id: item.id! })).status;
		},
		resume: async function (item: TransferItem): Promise<boolean> {
			if ((await this.getTransferInfo(item)) == undefined) {
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
			return (await FileDownloadPlugin.resume(options)).status;
		},

		complete: async function (): Promise<boolean> {
			return true;
		},

		getTransferInfo: async function (
			item: TransferItem
		): Promise<{ id: number; bytes: number } | undefined> {
			return await FileDownloadPlugin.getTransferInfo({ id: item.id! });
		},
		restartEnable: async function (): Promise<boolean> {
			return true;
		},
		restartAutoResume: true
	};
	uploader = {
		async start(item: TransferItem) {
			try {
				const result = await FileUploadPlugin.getTransferInfo({
					id: item.id!
				});
				if (result) {
					return (
						await FileUploadPlugin.resume({
							id: item.id!,
							baseUrl: getFileBaseUrl(item)
						})
					).status;
				}

				const dataAPI = dataAPIs(item.driveType);

				if (item.size == 0 && filesIsV2()) {
					const apiV2 = dataAPIsV2(item.driveType);
					apiV2.uploadEmptyFile(item.id!).then(() => {
						FileUploadPlugin.clearData({
							savePath: item.from!
						});
						busEmit('fileUploadComleted', {
							code: 0,
							message: '',
							path: '',
							id: item.id
						});
					});
					return true;
				}

				const fullPathSplit = item.path.split('?');
				let parentPath = fullPathSplit[0];
				if (!parentPath.endsWith('/') && parentPath.endsWith(item.name)) {
					parentPath = parentPath.substring(
						0,
						parentPath.length - item.name.length
					);
				}

				const pathname = (await dataAPI.formatUploaderPath(parentPath)) || '/';

				const moreInfo = dataAPI.getUploadTransferItemMoreInfo(item);

				let uploadPath = pathname;

				try {
					uploadPath = decodeURIComponent(pathname);
				} catch (error) {
					/* empty */
				}

				let account = '';

				if (moreInfo) {
					account = moreInfo.account;
					uploadPath = moreInfo.cloudFilePath;
				}

				const uploadOptions = {
					id: item.id!,
					baseUrl: getFileBaseUrl(item),
					uploadPath: uploadPath,
					filePath: item.from!,
					fileType: getUploadType(item),
					repoId: item.repo_id,
					uniqueIdentifier: item.uniqueIdentifier,
					account,
					node: item.node
				};
				return (await FileUploadPlugin.start(uploadOptions)).status;
			} catch (error) {
				console.log('upload error ===>', error.message);
				return false;
			}
		},
		cancel: async function (item: TransferItem): Promise<boolean> {
			if ((await this.getTransferInfo(item)) == undefined) {
				FileUploadPlugin.clearData({
					savePath: item.from!
				});

				return true;
			}
			return (await FileUploadPlugin.cancel({ id: item.id! })).status;
		},
		pause: async function (item: TransferItem): Promise<boolean> {
			if ((await this.getTransferInfo(item)) == undefined) {
				return true;
			}
			return (await FileUploadPlugin.pause({ id: item.id! })).status;
		},
		resume: async function (item: TransferItem): Promise<boolean> {
			if ((await this.getTransferInfo(item)) == undefined) {
				return true;
			}
			const userStore = useUserStore();

			const options = {
				id: item.id!,
				baseUrl: userStore.getModuleSever('files')
			};
			return (await FileUploadPlugin.resume(options)).status;
		},

		complete: async function (): Promise<boolean> {
			return true;
		},

		getTransferInfo: async function (
			item: TransferItem
		): Promise<{ id: number; bytes: number } | undefined> {
			return await FileUploadPlugin.getTransferInfo({ id: item.id! });
		},
		restartEnable: async function (): Promise<boolean> {
			return true;
		},
		restartAutoResume: true
	};

	restartAutoResume = true;
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
