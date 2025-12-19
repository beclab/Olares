import { defineStore } from 'pinia';
import { OPERATE_ACTION } from '../utils/contact';
import { RouteLocationNormalizedLoaded } from 'vue-router';
import { useFilesStore, FilesIdType } from './files';
import { useFilesCopyStore } from './files-copy';
import { BtNotify, NotifyDefinedType, BtDialog } from '@bytetrade/ui';

import { useDataStore } from './data';
import { MenuItem } from '../utils/contact';

import { CommonFetch, common, filesUtil, filesIsV2 } from '../api';
import { dataAPIs } from '../api/files';
import { useTransfer2Store } from './transfer2';
import { i18n } from '../boot/i18n';
import TransferClient from '../services/transfer';
import { getParams } from '../utils/utils';

import { DriveType } from 'src/utils/interface/files';
import {
	TransferFront,
	TransferItem,
	TransferStatus
} from 'src/utils/interface/transfer';
import { useUserStore } from './user';
import { notifySuccess } from 'src/utils/notifyRedefinedUtil';
import { Platform } from 'quasar';
import { ShareType } from 'src/utils/interface/share';
import { getApplication } from 'src/application/base';

export interface EventType {
	type?: DriveType;
	isSelected: boolean; //	true: Right click on a certain item; false: Right click on a blank area on the files
	hasCopied: boolean;
	showRename: boolean;
	isHomePage: boolean;
	selectCount: number;
	rw: boolean;
	isExternal: boolean;
	canBackup: boolean;
	isDir: boolean;
	isShareItem: boolean;
	isShareRoot: boolean;
	isOwner: boolean;
	// isShareEditEnable: boolean;
	isEditPermissionsEnable: boolean;
	isPublicShare: boolean;
}
export interface ContextType {
	name: string;
	icon: string;
	type?: string;
	action: OPERATE_ACTION;
	condition: (event: EventType) => boolean;
}

export interface CopyStoragesType {
	from: string;
	to: string;
	name: string;
	src_drive_type?: DriveType;
	dst_drive_type?: DriveType;
	key?: string;
	isDir?: boolean;
	src_node?: string;
	dst_node?: string;
}

export type DataState = {
	contextmenu: ContextType[];
	disableMenuItem: string[];
	copyFiles: CopyStoragesType[];
	defaultPath: string;
	executing: boolean;
	isCut: boolean;
};

export const useOperateinStore = defineStore('operation', {
	state: () => {
		return {
			contextmenu: [
				{
					name: 'buttons.download',
					icon: 'sym_r_browser_updated',
					action: OPERATE_ACTION.DOWNLOAD,
					condition: (event: EventType) =>
						event.isSelected && !event.isExternal && !event.isShareItem
				},
				{
					name: 'files_popup_menu.Share to Internal',
					icon: 'sym_r_folder_supervised',
					action: OPERATE_ACTION.SHARE_IN_INTERNAL,
					condition: (event: EventType) =>
						filesIsV2() &&
						(event.type === DriveType.Drive ||
							event.type === DriveType.Data ||
							event.type === DriveType.Sync ||
							event.type === DriveType.External ||
							event.type === DriveType.Cache) &&
						event.isSelected &&
						event.isDir
				},
				{
					name: 'files_popup_menu.Share to SMB',
					icon: 'sym_r_smb_share',
					action: OPERATE_ACTION.SHARE_IN_SMB,
					condition: (event: EventType) =>
						filesIsV2() &&
						(event.type === DriveType.Drive ||
							event.type === DriveType.Data ||
							event.type === DriveType.External ||
							event.type === DriveType.Cache) &&
						event.isSelected &&
						event.isDir
				},
				{
					name: 'files_popup_menu.Share to Public',
					icon: 'sym_r_link',
					action: OPERATE_ACTION.SHARE_IN_PUBLIC,
					condition: (event: EventType) =>
						filesIsV2() &&
						(event.type === DriveType.Drive || event.type === DriveType.Data) &&
						event.isSelected &&
						event.isDir
				},
				{
					name: 'files_popup_menu.Reset Password',
					icon: 'sym_r_visibility_lock',
					action: OPERATE_ACTION.RESET_PASSWORD,
					condition: (event: EventType) =>
						filesIsV2() &&
						event.type === DriveType.Share &&
						event.isSelected &&
						event.isPublicShare
				},

				{
					name: 'copy',
					icon: 'sym_r_content_copy',
					action: OPERATE_ACTION.COPY,
					condition: (event: EventType) =>
						event.isSelected &&
						!event.isExternal &&
						event.isSelected &&
						!event.isShareItem
				},
				{
					name: 'files.cut',
					icon: 'sym_r_move_up',
					action: OPERATE_ACTION.CUT,
					condition: (event: EventType) =>
						event.isSelected &&
						event.rw &&
						!event.isExternal &&
						!event.isHomePage &&
						!event.isShareItem
				},
				{
					name: 'files_popup_menu.rename',
					icon: 'sym_r_edit_square',
					action: OPERATE_ACTION.RENAME,
					condition: (event: EventType) =>
						event.isSelected &&
						event.showRename &&
						!event.isHomePage &&
						event.rw &&
						!event.isExternal &&
						!event.isShareItem
				},
				{
					name: 'files_popup_menu.backup',
					icon: 'sym_r_backup',
					action: OPERATE_ACTION.BACKUP,
					condition: (event: EventType) =>
						event.canBackup && event.selectCount == 1 && filesIsV2()
				},
				{
					name: 'files_popup_menu.Edit permissions',
					icon: 'sym_r_admin_panel_settings',
					action: OPERATE_ACTION.EDIT_PERMISSIONS,
					condition: (event: EventType) =>
						filesIsV2() && event.isEditPermissionsEnable && event.isSelected
				},
				{
					name: 'files_popup_menu.delete',
					icon: 'sym_r_delete',
					action: OPERATE_ACTION.DELETE,
					condition: (event: EventType) =>
						event.isSelected &&
						!event.isHomePage &&
						event.rw &&
						!event.isExternal &&
						!event.isShareItem
				},
				{
					name: 'files_popup_menu.Revoke sharing',
					icon: 'sym_r_do_not_disturb_on',
					action: OPERATE_ACTION.REVOKE_SHARING,
					condition: (event: EventType) =>
						event.isSelected && event.isShareItem && event.isOwner
				},
				{
					name: 'files_popup_menu.attributes',
					icon: 'sym_r_ballot',
					action: OPERATE_ACTION.ATTRIBUTES,
					condition: (event: EventType) =>
						event.isSelected && event.selectCount == 1 && !event.isExternal
				},
				{
					name: 'files_popup_menu.new_folder',
					icon: 'sym_r_create_new_folder',
					action: OPERATE_ACTION.CREATE_FOLDER,
					condition: (event: EventType) =>
						!event.isSelected &&
						event.rw &&
						!event.isExternal &&
						!event.isShareRoot
				},
				{
					name: 'files_popup_menu.upload_file',
					icon: 'sym_r_upload_file',
					action: OPERATE_ACTION.UPLOAD_FILES,
					condition: (event: EventType) =>
						!event.isSelected &&
						event.rw &&
						!event.isExternal &&
						!event.isShareRoot
				},
				{
					name: 'files_popup_menu.upload_folder',
					icon: 'sym_r_drive_folder_upload',
					action: OPERATE_ACTION.UPLOAD_FOLDER,
					condition: (event: EventType) =>
						!event.isSelected &&
						event.rw &&
						!event.isExternal &&
						!event.isShareRoot
				},
				{
					name: 'paste',
					icon: 'sym_r_content_paste',
					action: OPERATE_ACTION.PASTE,
					condition: (event: EventType) =>
						!event.isSelected &&
						event.hasCopied &&
						event.rw &&
						!event.isExternal &&
						!event.isShareRoot
				},
				{
					name: 'files.refresh',
					icon: 'sym_r_replay',
					action: OPERATE_ACTION.REFRESH,
					condition: (event: EventType) => !event.isSelected
				},
				{
					name: 'cancel',
					icon: 'sym_r_cancel',
					action: OPERATE_ACTION.CANCEL,
					condition: (event: EventType) =>
						event.hasCopied && !event.isExternal && !event.isShareItem
				},
				{
					name: 'files_popup_menu.unmount',
					icon: 'sym_r_usb_off',
					action: OPERATE_ACTION.UNMOUNT,
					condition: (event: EventType) =>
						event.isSelected && event.isExternal && event.selectCount == 1
				}
			],
			disableMenuItem: [
				MenuItem.HOME,
				MenuItem.DOCUMENTS,
				MenuItem.PICTURES,
				MenuItem.MOVIES,
				MenuItem.DOWNLOADS,
				MenuItem.DATA,
				MenuItem.CACHE,
				MenuItem.CODE,
				MenuItem.MUSIC
			],
			executing: false,
			copyFiles: [],
			defaultPath: '/Files/Home/',
			isCut: false
		} as DataState;
	},

	getters: {},

	actions: {
		async handleFileOperate(
			origin_id: number = FilesIdType.PAGEID,
			e: any,
			route: RouteLocationNormalizedLoaded,
			action: OPERATE_ACTION,
			driveType: DriveType,
			callback: (action: OPERATE_ACTION, data: any) => Promise<void>
		): Promise<void> {
			e && e.preventDefault();
			e && e.stopPropagation();

			const dataStore = useDataStore();
			const filesStore = useFilesStore();

			switch (action) {
				case OPERATE_ACTION.CREATE_FOLDER:
					dataStore.showHover('newDir');
					break;

				case OPERATE_ACTION.CREATE_REPO:
					dataStore.showHover('NewLib');
					break;

				case OPERATE_ACTION.DOWNLOAD:
					await this.download(route.path, origin_id);
					callback(action, undefined);
					break;

				case OPERATE_ACTION.UPLOAD_FILES:
					this.uploadFiles();
					break;

				case OPERATE_ACTION.UPLOAD_FOLDER:
					this.uploadFolder();
					break;

				case OPERATE_ACTION.ATTRIBUTES:
					dataStore.showHover('info');
					break;

				case OPERATE_ACTION.COPY:
					this.copyCatalogue(driveType, origin_id);
					callback && callback(action, undefined);
					break;

				case OPERATE_ACTION.CUT:
					this.cutCatalogue(driveType, origin_id);
					callback && callback(action, undefined);
					break;

				case OPERATE_ACTION.PASTE:
					this.pasteCatalogue(
						filesStore.currentPath[origin_id].path +
							filesStore.currentPath[origin_id].param,
						driveType,
						origin_id,
						callback
					);
					break;

				case OPERATE_ACTION.MOVE:
					this.moveCatalogue(route, driveType, origin_id, callback);
					break;

				case OPERATE_ACTION.RENAME:
					dataStore.showHover('rename');
					break;

				case OPERATE_ACTION.DELETE:
					dataStore.showHover('delete');
					break;

				case OPERATE_ACTION.REFRESH:
					{
						const filesStore = useFilesStore();
						const currentPath = filesStore.currentPath[origin_id];
						await filesStore.refushCurrentRouter(
							currentPath.path + currentPath.param,
							filesStore.activeMenu(origin_id).driveType,
							origin_id
						);
						notifySuccess(i18n.global.t('files.Refresh successful'));
					}
					break;

				case OPERATE_ACTION.SHARE:
					// dataStore.showHover('share-dialog2');
					break;
				case OPERATE_ACTION.SHARE_IN_INTERNAL:
					{
						const isMobile =
							process.env.PLATFORM == 'MOBILE' || Platform.is.mobile;
						dataStore.showHover(
							isMobile
								? 'share-internal-mobile-dialog'
								: 'share-internal-dialog'
						);
					}
					break;
				case OPERATE_ACTION.SHARE_IN_SMB:
					{
						const isMobile =
							process.env.PLATFORM == 'MOBILE' || Platform.is.mobile;
						dataStore.showHover(
							isMobile ? 'share-smb-mobile-dialog' : 'share-smb-dialog'
						);
					}
					break;
				case OPERATE_ACTION.SHARE_IN_PUBLIC:
					{
						const isMobile =
							process.env.PLATFORM == 'MOBILE' || Platform.is.mobile;
						dataStore.showHover(
							isMobile ? 'share-public-mobile-dialog' : 'share-public-dialog'
						);
					}
					break;
				case OPERATE_ACTION.RESET_PASSWORD:
					{
						dataStore.showHover('share-reset-password');
					}
					break;
				case OPERATE_ACTION.EDIT_PERMISSIONS:
					{
						const isMobile =
							process.env.PLATFORM == 'MOBILE' || Platform.is.mobile;

						const selectFiles =
							filesStore.currentFileList[origin_id]?.items[
								filesStore.selected[origin_id][0]
							];
						if (!selectFiles) {
							return;
						}
						if (
							selectFiles.share_type &&
							selectFiles.share_type == ShareType.SMB
						) {
							dataStore.showHover(
								isMobile ? 'share-smb-mobile-dialog' : 'share-smb-dialog'
							);
						} else {
							dataStore.showHover(
								isMobile
									? 'share-internal-mobile-dialog'
									: 'share-internal-dialog'
							);
						}
					}

					break;
				case OPERATE_ACTION.REVOKE_SHARING:
					dataStore.showHover('revoke_share');
					break;

				case OPERATE_ACTION.SYNCHRONIZE_TO_LOCAL:
					break;

				case OPERATE_ACTION.OPEN_LOCAL_SYNC_FOLDER:
					this.openLocalFolder(route, origin_id);
					break;
				case OPERATE_ACTION.CANCEL:
					this.resetCopyFiles();
					break;

				case OPERATE_ACTION.UNMOUNT:
					this.unmount(route, origin_id);
					break;

				case OPERATE_ACTION.BACKUP:
					this.openSetting(origin_id);
					break;

				default:
					break;
			}
		},

		async openSetting(origin_id) {
			const filesStore = useFilesStore();
			const selectFiles =
				filesStore.currentFileList[origin_id]?.items[
					filesStore.selected[origin_id][0]
				];
			if (!selectFiles) {
				return;
			}

			let origin = window.location.origin;
			if (process.env.APPLICATION == 'LAREPASS') {
				const user = useUserStore();
				origin = user.getModuleSever('files');
			}
			const settingOrigin = origin.replace('files', 'settings');
			const url = `${settingOrigin}/backup/create_backup/files/${encodeURIComponent(
				selectFiles.path
			)}`;

			getApplication().openUrl(url);
		},

		async download(path: string, origin_id: number) {
			const dataStore = useDataStore();
			const transferStore = useTransfer2Store();
			const dataAPI = dataAPIs(undefined, origin_id);

			let downloadInfoRes;

			if (dataStore.preview.isShow) {
				downloadInfoRes = await filesUtil().getPreviewDownloadInfo(origin_id);
			} else {
				downloadInfoRes = await dataAPI.getDownloadInfo(path);
			}

			const hasCanDownload = downloadInfoRes.find(
				(item) =>
					item.params && getParams(item.params, 'canDownload') === 'false'
			);

			if (hasCanDownload) {
				BtDialog.show({
					title: i18n.global.t('tips'),
					message: i18n.global.t('files.message_download_folder'),
					okStyle: {
						background: 'yellow-default',
						color: '#1F1F1F'
					},
					cancel: false,
					okText: i18n.global.t('base.confirm')
				});

				return false;
			}

			if (process.env.APPLICATION === 'FILES') {
				const hasFolder = downloadInfoRes.find((item) => item.isFolder);
				if (hasFolder) {
					BtDialog.show({
						title: i18n.global.t('tips'),
						message: i18n.global.t('files.message_download_folder'),
						okStyle: {
							background: 'yellow-default',
							color: '#1F1F1F'
						},
						cancel: false,
						okText: i18n.global.t('base.confirm')
					});
				}
			}

			// const transferStore = useTransfer2Store();
			for (let i = 0; i < downloadInfoRes.length; i++) {
				const info = downloadInfoRes[i];
				// info.isf = info.isFolder;
				if (process.env.APPLICATION === 'FILES') {
					if (info.isFolder) {
						continue;
					}
				}

				if (info.params && getParams(info.params, 'canDownload') === 'false') {
					continue;
				}

				const taskId = await transferStore.add(
					JSON.parse(JSON.stringify(info)),
					TransferFront.download
				);
				if (taskId < 0) {
					continue;
				}

				if (info.isFolder) {
					TransferClient.waitAddSubtasks[taskId] = {
						finished: false,
						subtasks: []
					};
					const response: any = await fetch(info.url);
					const reader = response.body.getReader();
					const decoder = new TextDecoder('utf-8');
					let lastLeftString = '';
					/* eslint-disable no-constant-condition */
					while (true) {
						try {
							const { done, value } = await reader.read();
							const result =
								lastLeftString + decoder.decode(value, { stream: true });
							const strameData = this.formatStremDataToTransferItems(
								result,
								path,
								info.path,
								origin_id
							);
							lastLeftString = strameData.leftString;
							TransferClient.waitAddSubtasks[taskId].finished = done;
							TransferClient.waitAddSubtasks[taskId].subtasks =
								TransferClient.waitAddSubtasks[taskId].subtasks.concat(
									strameData.datas
								);
							if (done) {
								break;
							}
						} catch (error) {
							break;
						}
					}
				}
			}
		},
		formatStremDataToTransferItems(
			rawData,
			parentPath: string,
			infoPath: string,
			origin_id: number
		) {
			const driveType = common().formatUrltoDriveType(parentPath);
			const dataAPI = dataAPIs(driveType, origin_id);
			const dataArray: TransferItem[] = [];
			const lines = rawData.trim().split('\n');
			let leftString = '';

			lines.forEach((line) => {
				if (!line) return false;
				const jsonString = line.replace('data: ', '');
				try {
					const item = JSON.parse(jsonString);
					dataAPI.formatSteamDownloadItem(item, infoPath, parentPath);
					const obj: TransferItem = {
						name: item.name,
						path: item.path,
						parentPath: item.parentPath,
						type: item.type,
						isFolder: false,
						driveType: item.driveType,
						front: TransferFront.download,
						status: TransferStatus.Pending,
						url: item.url,
						startTime: new Date().getTime(),
						endTime: 0,
						updateTime: new Date().getTime(),
						from: item.path,
						to: '',
						size: item.size,
						message: '',
						uniqueIdentifier: item.uniqueIdentifier,
						isPaused: false,
						params: item.params,
						node: '',
						currentPhase: 1,
						totalPhase: 1
					};
					dataArray.push(obj);
				} catch (error) {
					console.log('err', error);
					leftString = line;
				}
			});
			return {
				datas: dataArray,
				leftString
			};
		},

		async copyCatalogue(driveType: DriveType, origin_id: number) {
			const dataAPI = dataAPIs(driveType, origin_id);
			const filesStore = useFilesStore();
			const copyStorages: CopyStoragesType[] = [];

			for (const item of filesStore.selected[origin_id]) {
				const el = filesStore.getTargetFileItem(item, origin_id);
				if (!el) {
					continue;
				}

				const copyItem = await dataAPI.copy(el, 'copy');
				copyStorages.push(copyItem);
			}

			this.updateCopyFiles(copyStorages, false);
		},

		async pasteCatalogue(
			path: string,
			driveType: DriveType,
			origin_id: number,
			callback: (action: OPERATE_ACTION, data: any) => Promise<void>
		) {
			const filesCopyStore = useFilesCopyStore();
			const transferStore = useTransfer2Store();
			const dataAPI = dataAPIs(driveType, origin_id);
			if (this.copyFiles == undefined || this.copyFiles.length == 0) {
				return;
			}

			try {
				const tasks = await dataAPI.paste(path, callback);
				if (tasks) {
					if (tasks && Array.isArray(tasks) && tasks.length > 0) {
						for (let i = 0; i < tasks.length; i++) {
							const task = tasks[i];
							await filesCopyStore.createTask({
								...task,
								action: 'copy'
							});
						}
						this.resetCopyFiles();
						if (!transferStore.isUploadProgressDialogShow) {
							transferStore.isUploadProgressDialogShow = true;
						}
					}
				} else {
					this.resetCopyFiles();
					const filesStore = useFilesStore();
					await filesStore.refushCurrentRouter(path, driveType);
				}
			} catch (error) {
				console.log('error', error);
			}
		},

		async cutCatalogue(driveType: DriveType, origin_id: number) {
			const dataAPI = dataAPIs(driveType, origin_id);
			const filesStore = useFilesStore();
			const copyStorages: CopyStoragesType[] = [];

			for (const item of filesStore.selected[origin_id]) {
				const el = filesStore.getTargetFileItem(item, origin_id);
				if (!el) {
					continue;
				}

				const copyItem = await dataAPI.copy(el, 'cut');
				copyStorages.push(copyItem);
			}

			this.updateCopyFiles(copyStorages, true);
		},

		async moveCatalogue(
			route: RouteLocationNormalizedLoaded,
			driveType: DriveType,
			origin_id: number,
			callback: (action: OPERATE_ACTION, data: any) => Promise<void>
		) {
			const filesCopyStore = useFilesCopyStore();
			const transferStore = useTransfer2Store();
			const dataAPI = dataAPIs(driveType, origin_id);

			try {
				const tasks = await dataAPI.move(route.path, callback);
				if (tasks && Array.isArray(tasks)) {
					for (let i = 0; i < tasks.length; i++) {
						const task = tasks[i];
						await filesCopyStore.createTask({
							...task,
							action: 'move'
						});
					}
				}

				if (!transferStore.isUploadProgressDialogShow) {
					transferStore.isUploadProgressDialogShow = true;
				}
			} catch (error) {
				console.log('error', error);
			}
		},

		uploadFiles() {
			const fileStore = useFilesStore();
			fileStore.selectUploadFiles();
		},

		uploadFolder() {
			const fileStore = useFilesStore();
			fileStore.selectUploadFolder();
		},

		openLocalFolder(route, origin_id) {
			const filesStore = useFilesStore();
			const repo_id = route.query.id as string;

			const item =
				filesStore.currentFileList[origin_id]?.items[
					filesStore.selected[origin_id][0]
				];
			if (!item || !item.isDir) {
				return undefined;
			}
			const itemUrl = decodeURIComponent(item.path);
			const pathFromStart =
				itemUrl.indexOf(filesStore.activeMenu(origin_id).label) +
				filesStore.activeMenu(origin_id).label.length;
			const path = itemUrl.slice(pathFromStart, itemUrl.length - 1);
			if (process.env.PLATFORM == 'DESKTOP') {
				window.electron.api.files.openLocalRepo(repo_id, path);
			}
		},

		updateCopyFiles(copyStorages: CopyStoragesType[], isCut: boolean) {
			this.isCut = isCut;
			this.copyFiles = copyStorages;
		},

		resetCopyFiles() {
			if (this.isCut) {
				this.copyFiles = [];
				this.isCut = false;
			}
		},

		async unmount(route, origin_id) {
			const dataStore = useDataStore();
			const filesStore = useFilesStore();

			const selectFile =
				filesStore.currentFileList[origin_id]?.items[
					filesStore.selected[origin_id][0]
				];
			let node = '';
			if (filesStore.nodes.length > 0) {
				node = filesStore.currentNode[FilesIdType.PAGEID].name;
			}
			if (!selectFile) {
				return;
			}

			let path =
				'/api/unmount/external/' +
				selectFile.name +
				'?external_type=' +
				selectFile.externalType;
			if (filesIsV2()) {
				path =
					'/api/unmount/' +
					selectFile.fileType +
					'/' +
					selectFile.fileExtend +
					'/' +
					selectFile.name +
					'/' +
					'?external_type=' +
					selectFile.externalType;
			}

			const data = await CommonFetch.post(dataStore.baseURL() + path, {});

			if (data.data.code === 200) {
				BtNotify.show({
					type: NotifyDefinedType.SUCCESS,
					message: i18n.global.t('unmount_success')
				});
				filesStore.setBrowserUrl(
					route.fullPath,
					filesStore.activeMenu(origin_id).driveType
				);
			}
		},

		async getMd5(selectFiles) {
			try {
				const purePath = dataAPIs().getDiskPath(selectFiles, 'md5');
				const res = await CommonFetch.get(purePath, {});
				return res.md5;
			} catch (error) {
				return 'error';
			}
		},

		async getPermission(selectFiles) {
			try {
				const purePath = await dataAPIs().getDiskPath(
					selectFiles,
					'permission'
				);
				const res = await CommonFetch.get(purePath);
				return res.uid;
			} catch (error) {
				console.log('error', error);
			}
		},

		async setPermission(selectFiles, uid: string, recursive: boolean) {
			try {
				const purePath = await dataAPIs().getDiskPath(
					selectFiles,
					'permission'
				);

				let queryParams = `?uid=${uid}`;
				if (recursive) {
					queryParams += `&recursive=1`;
				}

				const res = await CommonFetch.put(`${purePath}${queryParams}`, {});

				return res;
			} catch (error) {
				console.log('error', error);
			}
		},

		isDisableMenuItem(name: string, path: string) {
			if (path === '/Files/Home/' && this.disableMenuItem.includes(name)) {
				return true;
			}

			return false;
		}
	}
});
