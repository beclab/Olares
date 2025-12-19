import { defineStore } from 'pinia';
import { getFileIcon } from '@bytetrade/core';
// import Origin from '../api/origin';
import { CommonFetch, dataAPIs, common, filesIsV1 } from './../api';

import { FilesSortType } from '../utils/contact';
import { deduplicateByField } from '../utils/file';
import { useMenuStore } from './files-menu';
import { useUserStore } from './user';
import { Platform } from 'quasar';
import { busEmit } from 'src/utils/bus';
import { useOperateinStore } from './operation';
import Resumable from '../utils/resumejs';

import { i18n } from 'src/boot/i18n';
import { MenuItem } from '../utils/contact';
import { DriveType } from '../utils/interface/files';
import { translateFolderName } from './../utils/file';
import { notifySuccess } from './../utils/notifyRedefinedUtil';

import { IUploadFileInfo } from 'src/platform/interface/electron/interface';
import { useDataStore } from './data';
import { delay } from 'src/utils/utils';
import { filesIsV2 } from 'src/api';
import { SharePermission, ShareType } from 'src/utils/interface/share';
import { SyncRepoMineType } from 'src/api/files/v2/sync/type';
import { appendPath } from 'src/api/files/path';
import { getNativeAppPlatform } from 'src/application/platform';

export enum FilesIdType {
	PAGEID = 0,
	SHARE = 1
}

export enum PickType {
	FOLDER = 'FOLDER',
	FILE = 'FILE'
}
export interface MenuItemType {
	label: string;
	key: string | number;
	icon: string;
	expationFlag?: boolean;
	driveType?: DriveType;
	children?: MenuItemType[];
}
export interface RecentFolderItem {
	name: string;
	path: string;
	url: string;
	driveType: DriveType;
}

export interface DriveSortingType {
	asc: boolean;
	by: string;
}

export enum ExoirationTime {
	all = 0,
	within1days = 1,
	within7days = 2,
	within30days = 3,
	within1year = 4,
	over1year = 5
}

export interface ShareFilter {
	shared: {
		byMe: boolean;
		withMe: boolean;
	};

	owner: string[];

	scope: {
		public: boolean;
		smb: boolean;
		internal: boolean;
	};

	permission: {
		manage: boolean;
		edit: boolean;
		view: boolean;
	};

	expire: ExoirationTime;

	ownerInit: boolean;
}

export interface ActiveSortType {
	by: FilesSortType;
	asc: boolean;
}

export interface SmbMountType {
	password?: string;
	timestamp?: number;
	url: string;
	username?: string;
}

export enum ExternalType {
	SMB = 'smb',
	USB = 'usb',
	HDD = 'hdd',
	OTHERS = 'others'
}

export interface SMBUser {
	id: string;
	name: string;
	password?: string;
}

export interface SMBPermissionUser extends SMBUser {
	permission: SharePermission;
	editingPermission?: SharePermission;
	pwdDisplay: boolean;
}

export interface ShareUser {
	name: string;
	role: string;
	status: string;
	olaresId: string;
}

export interface ShareItemUser extends ShareUser {
	permission: SharePermission;
	isOwner: boolean;
	editingPermission?: SharePermission;
}

export interface ShareUserList {
	owner: string;
	olaresId: string;
	users: ShareUser[];
}

export interface ShareMember {
	// id: number;
	path_id: string;
	share_member: string;
	permission: SharePermission;
	// create_time: string;
	// update_time: string;
}

export const sharePermissionStr = (permission?: number) => {
	if (permission == undefined) {
		return '';
	}
	if (permission == SharePermission.View) {
		return i18n.global.t('files.permissions.view');
	}
	if (permission == SharePermission.Edit) {
		return i18n.global.t('files.permissions.edit');
	}
	if (permission == SharePermission.UploadOnly) {
		return i18n.global.t('files.permissions.upload_only');
	}
	if (permission == SharePermission.ADMIN) {
		return i18n.global.t('files.permissions.admin');
	}
	return '';
};

export const shareTypeStr = (shareType: ShareType) => {
	if (shareType == ShareType.INTERNAL) {
		return i18n.global.t('files.Internal');
	}

	if (shareType == ShareType.PUBLIC) {
		return i18n.global.t('files.Public');
	}

	if (shareType == ShareType.SMB) {
		return 'SMB';
	}

	return '';
};

export const permissionOpt = () => {
	return [
		SharePermission.ADMIN,
		SharePermission.Edit,
		SharePermission.View
	].map((e) => {
		return {
			label: sharePermissionStr(e),
			value: e
		};
	});
};

export const smbPermissiontOpt = () => {
	return [SharePermission.Edit, SharePermission.View].map((e) => {
		return {
			label: sharePermissionStr(e),
			value: e
		};
	});
};

export class FilePath {
	isDir: boolean;
	path: string;
	driveType: DriveType;
	param: string;

	constructor(props?: Partial<FilePath>) {
		props && Object.assign(this, props);
	}
}

export class FileResType {
	fileExtend: string;
	filePath: string;
	fileType: string;
	extension: string;
	fileSize?: number;
	isDir: boolean;
	isSymlink: boolean;
	mode: number;
	modified: number;
	name: string;
	numDirs: number;
	numFiles: number;
	numTotalFiles?: number;
	path: string;
	size: number;
	type?: string;
	items: FileItem[];
	sorting: DriveSortingType;
	url?: string;
	driveType: DriveType;
	shareRoot?: boolean;
	isShareItem?: boolean;
	shared_by_me?: boolean;
	permission?: SharePermission;
	node?: string;

	constructor(props?: Partial<FileResType>) {
		props && Object.assign(this, props);
	}
}

export interface FileNode {
	name: string;
	master: boolean;
}

export class FileItem {
	extension: string;
	isDir: boolean;
	isSymlink: boolean;
	mode: number;
	modified: number;
	name: string;
	path: string;
	size: number;
	type: string;
	parentPath?: string;
	sorting?: DriveSortingType;
	numDirs?: number;
	numFiles?: number;
	numTotalFiles?: number;
	encoded_thumbnail_src?: string;
	index: number;
	url: string;
	content?: string;
	id?: string;
	driveType: DriveType;
	param: string;
	externalType?: ExternalType;
	uniqueIdentifier?: string;
	fileExtend: string;
	fileType?: string;
	filePath?: string;
	isNode?: boolean;
	oPath?: string;
	oParentPath?: string;
	isShareItem?: boolean;
	expire_time?: string;
	share_type?: ShareType;
	permission?: number;
	owner?: string;
	shared_by_me?: boolean;
	users?: any;
	public_smb?: boolean;

	sync_repo_name?: string;
	extend?: string;
	file_type?: string;
	node?: string;

	constructor(props?: Partial<FileItem>) {
		props && Object.assign(this, props);
	}
}

export type FileState = {
	isInPreview: Record<number, string>;
	sort: Record<number, DriveSortingType>;
	currentPath: Record<number, FilePath>;
	backStack: Record<number, FilePath[]>;
	previousStack: Record<number, FilePath[]>;
	currentFileList: Record<number, FileResType | undefined>;
	cached: Record<number, FileResType | undefined>;
	selected: Record<number, number[]>;
	previewItem: Record<number, any>;
	activeSort: Record<number, ActiveSortType>;
	loading: Record<number, boolean>;
	// uploadFileList: Record<number, any[]>;
	isShard: boolean;
	mobileRepo: any;
	googleDirMap: Record<string, any>;
	menu: Record<number, MenuItemType[]>;
	filterHiddenDir: Record<number, boolean>;
	nodes: FileNode[];
	users: ShareUserList | undefined;
	currentNode: Record<number, FileNode>;
	onlyMasterNodes: Record<number, boolean>;

	shareFilter: ShareFilter;
	shareRepoInfo: SyncRepoMineType | undefined;
};

export const useFilesStore = defineStore('files', {
	state: () => {
		return {
			sort: {},
			isInPreview: {},
			currentPath: {},
			backStack: {},
			previousStack: {},
			currentFileList: {},
			cached: {},
			selected: {},
			previewItem: {},
			activeSort: {},
			loading: {},
			uploadFileList: {},
			isShard: false,
			mobileRepo: '',
			menu: {},
			googleDirMap: {},
			filterHiddenDir: {},
			nodes: [],
			currentNode: {},
			onlyMasterNodes: {},
			users: undefined,
			shareFilter: {
				shared: {
					byMe: true,
					withMe: true
				},
				owner: [],
				scope: {
					public: true,
					smb: true,
					internal: true
				},
				permission: {
					manage: true,
					edit: true,
					view: true
				},
				expire: ExoirationTime.all,
				ownerInit: false
			},
			shareRepoInfo: undefined
		} as FileState;
	},
	getters: {
		currentFileItems:
			(state) =>
			(id = FilesIdType.PAGEID) => {
				return state.currentFileList[id]?.items?.filter((item) => !item.isDir);
			},

		currentDirItems:
			(state) =>
			(id = FilesIdType.PAGEID) => {
				return state.currentFileList[id]?.items?.filter((item) => item.isDir);
			},

		currentFileIds:
			(state) =>
			(id = FilesIdType.PAGEID) => {
				const operateinStore = useOperateinStore();
				const filesStore = useFilesStore();
				const items = (
					state.currentFileList[id]?.items?.filter((item) => item.isDir) || []
				).concat(
					state.currentFileList[id]?.items?.filter((item) => !item.isDir) || []
				);

				return items?.map((item, index) => {
					return {
						id: index + '_' + item.name,
						selectedEnable: (value: string) => {
							return !operateinStore.isDisableMenuItem(
								filesStore.currentFileListMap(id)[value].name,
								filesStore.currentFileListMap(id)[value].path
							);
						}
					};
				});
			},

		currentFileListMap:
			(state) =>
			(id = FilesIdType.PAGEID) => {
				const fileMap = {};
				const items = (
					state.currentFileList[id]?.items?.filter((item) => item.isDir) || []
				).concat(
					state.currentFileList[id]?.items?.filter((item) => !item.isDir) || []
				);
				items.map((item, index) => {
					const id = index + '_' + item.name;
					fileMap[id] = item;
				});
				return fileMap;
			},

		selectedCount:
			(state) =>
			(id = FilesIdType.PAGEID) => {
				return state.selected[id] ? state.selected[id].length : 0;
			},

		hasPrevPath:
			(state) =>
			(id = FilesIdType.PAGEID) => {
				return state.previousStack[id] && state.previousStack[id].length > 0
					? true
					: false;
			},

		hasBackPath:
			(state) =>
			(id = FilesIdType.PAGEID) => {
				return state.backStack[id] && state.backStack[id].length > 0
					? true
					: false;
			},

		getCurrentRepo:
			(state) =>
			(path: string, id = FilesIdType.PAGEID) => {
				if (state.backStack[id].length == 0) {
					return '0';
				}
				const isEndWith =
					state.backStack[id][state.backStack[id].length - 1].path.endsWith(
						'/'
					);
				const currentPath = decodeURIComponent(
					state.backStack[id][state.backStack[id].length - 1].path
				).split('/');

				let name =
					currentPath.length < 1
						? ''
						: currentPath[
								isEndWith ? currentPath.length - 2 : currentPath.length - 1
						  ];
				if (
					state.backStack[id][state.backStack[id].length - 1].driveType ===
					DriveType.GoogleDrive
				) {
					if (state.googleDirMap[id] && state.googleDirMap[id][name]) {
						name = state.googleDirMap[id][name];
					}
				}
				return translateFolderName(path, name, true);
			},

		activeMenu:
			(state) =>
			(id = FilesIdType.PAGEID) => {
				if (!state.currentPath[id]) {
					return {
						label: 'Home',
						id: 'Home',
						driveType: DriveType.Drive
					};
				}
				const item = common().formatUrltoActiveMenu(
					state.currentPath[id].path + state.currentPath[id].param
				);
				if (state.currentPath[id].driveType != item.driveType) {
					item.driveType = state.currentPath[id].driveType;
				}

				return item;
			},
		masterNode(state) {
			const masterNode = state.nodes.find((e) => e.master == true);
			return masterNode
				? masterNode.name
				: state.nodes.length > 0
				? state.nodes[0].name
				: '';
		}
	},
	actions: {
		initIdState(id: number = FilesIdType.PAGEID) {
			this.sort[id] = {
				by: '',
				asc: false
			};
			this.currentPath[id] = {
				isDir: true,
				path: '/',
				driveType: DriveType.Drive,
				param: ''
			};
			this.backStack[id] = [];
			this.previousStack[id] = [];
			this.currentFileList[id] = undefined;
			this.selected[id] = [];
			this.previewItem[id] = {};
			this.activeSort[id] = {
				by: FilesSortType.Modified,
				asc: true
			};
			this.loading[id] = false;

			// this.uploadFileList[id] = [];
			this.menu[id] = [];
			this.googleDirMap[id] = {};
			this.filterHiddenDir[id] = false;
			this.onlyMasterNodes[id] = false;
		},

		removeIdState(id: number = FilesIdType.PAGEID) {
			delete this.sort[id];
			delete this.isInPreview[id];
			delete this.currentPath[id];
			delete this.backStack[id];
			delete this.previousStack[id];
			delete this.currentFileList[id];
			delete this.selected[id];
			delete this.previewItem[id];
			delete this.activeSort[id];
			delete this.loading[id];
			// delete this.uploadFileList[id];
			delete this.menu[id];
			delete this.googleDirMap[id];
			delete this.filterHiddenDir[id];
		},

		async fetchData(
			path: FilePath,
			key: string,
			requestUrl: string,
			id: number = FilesIdType.PAGEID
		): Promise<FileItem[]> {
			const res = await dataAPIs(path.driveType, id).fetch(requestUrl);
			const fileList: FileItem[] = res.items;
			this.cached[key] = res;

			if (
				path.driveType == this.currentPath[id].driveType &&
				path.path == this.currentPath[id].path &&
				path.param == this.currentPath[id].param
			) {
				this.currentFileList[id] = res;
			}
			return fileList;
		},

		async setFilePath(
			path: FilePath,
			isBack = false,
			isPrev = true,
			id: number = FilesIdType.PAGEID
		) {
			this.loading[id] = true;
			if (!path.isDir) {
				this.openPreviewDialog();
				return true;
			}

			const key = this.registerUniqueKey(path.path, path.driveType, path.param);

			if (key in this.cached) {
				this.currentFileList[id] = this.cached[key];
			} else {
				this.currentFileList[id] = undefined;
			}

			this.currentPath[id] = path;

			if (!isBack && isPrev) {
				this.backStack[id].push(path);
				this.previousStack[id] = [];
			}

			const params = new URLSearchParams(path.param);
			const query = Object.fromEntries(params);

			if (id === FilesIdType.PAGEID) {
				this.router.push({
					path: path.path,
					query
				});
			}

			const requestUrl = this.formatPathtoUrl(path);

			try {
				await this.fetchData(path, key, requestUrl, id);
				this.loading[id] = false;
				return true;
			} catch (error) {
				this.loading[id] = false;
				return false;
			}
		},

		async setBrowserUrl(
			url: string,
			driveType: DriveType = DriveType.Drive,
			isPrev = true,
			id: number = FilesIdType.PAGEID
		) {
			const splitUrl = url.split('?');
			const lastItem = splitUrl[0].split('/')[url.split('/').length - 1];
			if (!lastItem) {
				const path = new FilePath({
					path: splitUrl[0],
					param: splitUrl[1] ? `?${splitUrl[1]}` : '',
					isDir: true,
					driveType: driveType
				});

				return await this.setFilePath(path, false, isPrev, id);
			} else {
				const lastIndex = splitUrl[0].lastIndexOf('/');
				const newPath =
					lastIndex !== -1 ? splitUrl[0].slice(0, lastIndex + 1) : splitUrl[0];

				const path = new FilePath({
					path: newPath,
					param: splitUrl[1] ? `?${splitUrl[1]}` : '',
					isDir: true,
					driveType: driveType
				});
				const result = await this.setFilePath(path, false, true, id);
				this.addSelected(
					this.currentFileList[id]?.items?.findIndex(
						(item) => item.name === decodeURIComponent(lastItem)
					),
					id
				);
				await this.openPreviewDialog();
				this.resetSelected(id);

				return result;
			}
		},

		async refushCurrentRouter(
			fullPath: string,
			driveType: DriveType,
			id: number = FilesIdType.PAGEID
		) {
			const splitFullPath = fullPath.split('?');

			await this.setFilePath(
				{
					path: splitFullPath[0],
					isDir: true,
					driveType,
					param: splitFullPath[1] ? `?${splitFullPath[1]}` : ''
				},
				false,
				false,
				id
			);
		},

		registerUniqueKey(path: string, driveType: DriveType, param: string) {
			const userStore = useUserStore();

			let key = userStore.current_user?.name + '=' + path + '=' + driveType;
			if (param) {
				key = key + '=' + param;
			}
			return key;
		},

		updateActiveSort(
			type: FilesSortType,
			asc: boolean,
			id: number = FilesIdType.PAGEID
		) {
			this.activeSort[id] = {
				by: type,
				asc
			};
			if (this.currentFileList[id])
				this.currentFileList[id]!.items = this.sortList(
					this.currentFileItems(id),
					this.currentDirItems(id),
					id
				);
		},

		sortList(fileItems, dirItems, id: number = FilesIdType.PAGEID) {
			return [
				...this._sortList(dirItems, id),
				...this._sortList(fileItems, id)
			];
		},

		_sortList(list: any, id: number = FilesIdType.PAGEID) {
			const cur_list = JSON.parse(JSON.stringify(list));
			if (cur_list) {
				const list1 = cur_list.sort((a, b) => {
					if (this.activeSort[id].by == FilesSortType.TYPE) {
						return this.activeSort[id].asc
							? a.type.localeCompare(b.type)
							: -a.type.localeCompare(b.type);
					} else if (this.activeSort[id].by == FilesSortType.NAME) {
						return this.activeSort[id].asc
							? a.name.localeCompare(b.name)
							: -a.name.localeCompare(b.name);
					} else if (this.activeSort[id].by == FilesSortType.SIZE) {
						return this.activeSort[id].asc ? a.size - b.size : b.size - a.size;
					} else {
						if (typeof a.modified == 'string') {
							return this.activeSort[id].asc
								? a.modified.localeCompare(b.modified)
								: -a.modified.localeCompare(b.modified);
						} else {
							return this.activeSort[id].asc
								? a.modified - b.modified
								: b.modified - a.modified;
						}
					}
				});

				return list1;
			}
		},

		async back(id: number = FilesIdType.PAGEID) {
			if (!this.backStack[id] || this.backStack[id].length == 0) {
				return;
			}
			if (id == FilesIdType.SHARE && this.backStack[id].length == 1) {
				return;
			}
			const path = this.backStack[id].pop();

			if (this.backStack[id].length == 0 && process.env.PLATFORM == 'MOBILE') {
				return;
			}

			const initPath = new FilePath({
				path: '/Files/Home',
				param: '',
				isDir: true,
				driveType: DriveType.Drive
			});

			const currentPath =
				this.backStack[id][this.backStack[id].length - 1] || initPath;

			if (path) {
				if (!this.previousStack[id]) {
					this.previousStack[id] = [path];
				} else {
					this.previousStack[id].push(path);
				}

				this.setFilePath(currentPath, true, true, id);
			}
		},

		previous(id: number = FilesIdType.PAGEID) {
			if (this.previousStack[id].length == 0) {
				return;
			}

			const path = this.previousStack[id].pop();

			if (path) {
				this.backStack[id].push(path);
				this.setFilePath(path, true, true, id);
			}
		},

		getTargetFileItem(index: number, id: number = FilesIdType.PAGEID) {
			console.log(' this.currentFileList[id] ===>', this.currentFileList[id]);

			return this.currentFileList[id]?.items.find(
				(item) => item.index == index
			);
		},

		addSelected(value: any, id: number = FilesIdType.PAGEID) {
			if (!this.selected[id]) {
				this.selected[id] = [];
			}
			this.selected[id] = [...new Set((this.selected[id] || []).concat(value))];
		},

		removeSelected(value: any, id: number = FilesIdType.PAGEID) {
			const i = this.selected[id].indexOf(value);
			if (i === -1) return;
			this.selected[id].splice(i, 1);
		},

		resetSelected(id: number = FilesIdType.PAGEID) {
			this.selected[id] = [];
		},

		async formatRepotoPath(value, id: number = FilesIdType.PAGEID) {
			return await dataAPIs(value.driveType, id).formatRepotoPath(value);
		},

		formatPathtoUrl(value: FilePath) {
			return dataAPIs(value.driveType).formatPathtoUrl(value.path, value.param);
		},

		async openPreviewDialog(
			itemFile?: FileItem,
			id: number = FilesIdType.PAGEID
		) {
			this.previewItem[id] = {};

			let cur_item;
			if (itemFile) {
				cur_item = JSON.parse(JSON.stringify(itemFile));
			} else {
				cur_item = this.currentFileList[id]?.items.find(
					(item) => item.index === this.selected[id][0]
				);
			}

			const api = dataAPIs(cur_item.driveType, id);
			let isVideo = getFileIcon(cur_item.name) === 'video';
			if (isVideo && !api.videoPlayEnable) {
				isVideo = false;
			}
			const store = useDataStore();
			if (!store.preview.isShow) {
				busEmit('filesPreviewDisplay', isVideo, id);
			}
			let awaitNumber = 0;
			while (
				!store.preview.isShow ||
				(this.masterNode == '' && filesIsV2() && awaitNumber < 5)
			) {
				awaitNumber++;
				await delay(300);
			}
			const res = await api.openPreview(cur_item);
			res.type = getFileIcon(res.name);
			this.previewItem[id] = res;
			if (itemFile) this.selected[id] = [itemFile.index];
		},

		getDownloadURL(file: FileItem, inline: boolean, download = false): string {
			return dataAPIs(file.driveType).getDownloadURL(file, inline, download);
		},

		getPreviewURL(file: FileItem, size: 'big' | 'thumb'): string {
			return dataAPIs(file.driveType).getPreviewURL(file, size);
		},

		addRecentFolder(data: RecentFolderItem) {
			const recentFolder: RecentFolderItem[] = localStorage.getItem(
				'recentFolder'
			)
				? JSON.parse(localStorage.getItem('recentFolder') as string)
				: [];

			recentFolder.unshift(data);

			if (recentFolder.length > 3) {
				recentFolder.pop();
			}

			localStorage.setItem('recentFolder', JSON.stringify(recentFolder));
		},

		sharedToFile(item: any) {
			console.log('sharedToFile', item);
		},

		async requestPathItems(
			url: string,
			driveType: DriveType = DriveType.Drive,
			id: number = FilesIdType.PAGEID
		) {
			try {
				const splitUrl = url.split('?');

				const path = new FilePath({
					path: splitUrl[0],
					param: splitUrl[1] ? `?${splitUrl[1]}` : '',
					isDir: true,
					driveType: driveType
				});

				const requestUrl = this.formatPathtoUrl(path);

				const key = this.registerUniqueKey(
					path.path,
					path.driveType,
					path.param
				);
				await dataAPIs(path.driveType, id)
					.fetch(requestUrl)
					.then((res: FileResType) => {
						this.loading[id] = false;
						this.cached[key] = res;
					})
					.catch(() => {
						// console.error('Error fetching items', error);
					});
			} catch (error) {
				console.log(error);
			}
		},

		async refreshPathItems(id: number = FilesIdType.PAGEID) {
			if (!this.currentPath[id]) {
				return;
			}
			const key = this.registerUniqueKey(
				this.currentPath[id].path,
				this.currentPath[id].driveType,
				this.currentPath[id].param
			);

			this.currentFileList[id] = this.cached[key];
		},
		async selectUploadFiles(isImage = false, id: number = FilesIdType.PAGEID) {
			if (Platform.is.nativeMobile) {
				getNativeAppPlatform().selectUploadFiles(
					this.currentPath[id].driveType,
					this.currentPath[id].path,
					this.currentPath[id].param,
					isImage
				);
			} else if (Platform.is.electron) {
				const paths = await window.electron.api.upload.selectUploadFiles();
				busEmit('electronUploadPaths', paths);
			} else {
				const dataAPI = dataAPIs();
				dataAPI.uploadFiles();
			}
		},
		async selectUploadFolder() {
			if (Platform.is.electron) {
				const path = await window.electron.api.upload.selectUploadFolder();
				if (path && path.length > 0) {
					busEmit('electronUploadPaths', [path]);
				}
			} else {
				const dataAPI = dataAPIs();
				dataAPI.uploadFolder();
			}
		},

		async selectSystemFile() {
			if (Platform.is.electron) {
				const paths = await window.electron.api.upload.selectUploadFiles();
				if (paths.length == 0) {
					return;
				}
				const files: IUploadFileInfo[] = [];
				for (let index = 0; index < paths.length; index++) {
					const filePath = paths[index];
					const info = await window.electron.api.upload.getPathInfo(filePath);
					files.push(info);
				}
				return {
					target: paths,
					files
				};
			} else {
				const dataAPI = dataAPIs();
				dataAPI.uploadFiles();
			}
		},

		async uploadSelectFile(
			target: any,
			filePath: FilePath,
			origin_id = FilesIdType.PAGEID
		) {
			if (Platform.is.electron) {
				busEmit('electronUploadPaths', target, filePath);
			} else {
				Resumable.addFilesToUpload(target, filePath, origin_id);
			}
		},

		driveTypeToApi(driveType: DriveType) {
			switch (driveType) {
				case DriveType.Drive:
					return {
						// api: new DriveDataAPI(),
						menuBar: {
							label: i18n.global.t(`files_menu.${MenuItem.DRIVE}`),
							key: MenuItem.DRIVE,
							icon: '',
							children: []
						}
					};

				case DriveType.Sync:
					return {
						// api: new SyncDataAPI(),
						menuBar: {
							label: i18n.global.t(`files_menu.${MenuItem.SYNC}`),
							key: MenuItem.SYNC,
							icon: '',
							children: []
						}
					};

				case DriveType.External:
					return {
						// api: new ExternalDataAPI(),
						menuBar: {
							label: i18n.global.t(`files_menu.${MenuItem.EXTERNAL}`),
							key: MenuItem.DRIVE,
							icon: '',
							children: []
						}
					};

				case DriveType.Data:
					return {
						// api: new DataDataAPI(),
						menuBar: {
							label: i18n.global.t(`files_menu.${MenuItem.APPLICATION}`),
							key: MenuItem.APPLICATION,
							icon: '',
							children: []
						}
					};

				case DriveType.Cache:
					return {
						// api: new CacheDataAPI(),
						menuBar: {
							label: i18n.global.t(`files_menu.${MenuItem.APPLICATION}`),
							key: MenuItem.APPLICATION,
							icon: '',
							children: []
						}
					};

				case DriveType.GoogleDrive:
					return {
						// api: new GoogleDataAPI(),
						menuBar: {
							label: i18n.global.t(`files_menu.${MenuItem.CLOUDDRIVE}`),
							key: MenuItem.CLOUDDRIVE,
							icon: '',
							children: []
						}
					};
				case DriveType.Share:
					return {
						menuBar: {
							label: i18n.global.t(`files_menu.${MenuItem.SHARE}`),
							key: MenuItem.SHARE,
							icon: '',
							children: []
						}
					};

				default:
					break;
			}
		},

		async getMenu(
			origins: DriveType[] = [
				DriveType.Drive,
				DriveType.External,
				DriveType.Sync,
				DriveType.Data,
				DriveType.Cache,
				DriveType.GoogleDrive,
				DriveType.Share
			],
			id: number = FilesIdType.PAGEID
		) {
			for (let i = 0; i < origins.length; i++) {
				const origin = origins[i];
				const cur_api = this.driveTypeToApi(origin);
				const driveAPI = dataAPIs(origin);
				if (!cur_api || !driveAPI) {
					continue;
				}
				if (process.env.PLATFORM === 'MOBILE') {
					if (origin === DriveType.Cache || origin === DriveType.Data) {
						continue;
					}
				}

				if (!this.menu[id]) this.menu[id] = [];

				const curMenuIndex = this.menu[id].findIndex(
					(menu) => menu.key === cur_api.menuBar?.key
				);

				const curDrive: any[] = await driveAPI.fetchMenuRepo();

				if (curDrive.length <= 0) {
					continue;
				}

				if (process.env.PLATFORM == 'DESKTOP' && origin === DriveType.Sync) {
					const menuStore = useMenuStore();
					const syncIds = curDrive.filter((item) => item.id).map((e) => e.id);
					menuStore.addSyncUpdateRepos(syncIds);
				}

				if (curMenuIndex > -1) {
					if (origin === DriveType.Sync) {
						this.menu[id][curMenuIndex].children = [];
					}

					const curDriveMenu = [
						...this.menu[id][curMenuIndex].children!,
						...curDrive
					];

					if (origin !== DriveType.Sync) {
						this.menu[id][curMenuIndex].children = deduplicateByField(
							curDriveMenu,
							'key'
						);
					} else {
						this.menu[id][curMenuIndex].children = curDriveMenu;
					}
				} else {
					const cur_menu: any = cur_api.menuBar;
					cur_menu.children = curDrive;
					this.menu[id].push(cur_api.menuBar);
				}
			}
		},

		async getMobileMenu(
			origins: DriveType[] = [DriveType.Drive, DriveType.Sync]
		): Promise<MenuItemType[]> {
			const menuRes: MenuItemType[] = [];
			for (let i = 0; i < origins.length; i++) {
				const origin = origins[i];

				switch (origin) {
					case DriveType.Drive:
						menuRes.push({
							label: i18n.global.t(`files_menu.${MenuItem.DRIVE}`),
							key: MenuItem.HOME,
							icon: 'file-drive.svg',
							driveType: DriveType.Drive
						});
						break;

					case DriveType.Sync:
						menuRes.push({
							label: i18n.global.t(`files_menu.${MenuItem.SYNC}`),
							key: MenuItem.SYNC,
							icon: 'file-sync.svg',
							driveType: DriveType.Sync
						});
						break;

					case DriveType.External:
						menuRes.push({
							label: i18n.global.t(`files_menu.${MenuItem.EXTERNAL}`),
							key: MenuItem.EXTERNAL,
							icon: 'file-external.svg',
							driveType: DriveType.External
						});
						break;

					case DriveType.Data:
						menuRes.push({
							label: i18n.global.t(`files_menu.${MenuItem.DATA}`),
							key: MenuItem.DATA,
							icon: 'file-data.svg',
							driveType: DriveType.Data
						});
						break;

					case DriveType.Cache:
						menuRes.push({
							label: i18n.global.t(`files_menu.${MenuItem.CACHE}`),
							key: MenuItem.CACHE,
							icon: 'file-cache.svg',
							driveType: DriveType.Cache
						});
						break;
				}
			}

			return menuRes;
		},
		async mountSmbInExternal(
			connectData: SmbMountType
		): Promise<{ code: number; data: any }> {
			let node = '';
			if (this.nodes.length > 0) {
				node = this.currentNode[FilesIdType.PAGEID].name;
			}

			const res = await CommonFetch.post(
				`/api/mount${node.length ? '/' + node + '/' : ''}?external_type=smb`,
				{
					smbPath: connectData.url,
					user: connectData.username,
					password: connectData.password
				}
			);

			if (res.data.code === 200) {
				notifySuccess(i18n.global.t('files.server_connect_success'));
				return {
					code: res.data.code,
					data: res.data
				};
			} else if (res.data.code === 300) {
				const data = res.data.data;
				data.map((item) => {
					const slashIndex = item.path.slice(2).indexOf('/') + 2;
					item.sambaPath = item.path.slice(0, slashIndex);
					item.dir = item.path.slice(slashIndex);
				});

				return {
					code: res.data.code,
					data
				};
			} else {
				// notifyFailed(res.data.message);
				throw new Error(res.data.message);
			}
		},
		setCurrentNode(node: string, id: number = FilesIdType.PAGEID) {
			if (this.nodes.length == 0) {
				return;
			}
			const fNode = this.nodes.find((e) => e.name == node);
			if (fNode) {
				this.currentNode[id] = {
					...fNode
				};
				return;
			}
			delete this.currentNode[id];
		},
		shareBaseUrl() {
			if (process.env.APPLICATION === 'FILES') {
				return this.getModuleSever('share', 'https:');
			} else if (process.env.APPLICATION == 'LAREPASS') {
				const user = useUserStore();
				const baseURL = user.getModuleSever('share');
				return baseURL;
			}
			return '';
		},
		getShareLinkAddress(id: string) {
			const shareBaseUrl = this.shareBaseUrl();
			return appendPath(shareBaseUrl, '/sharable-link/' + id, '/');
		},
		getModuleSever(module: string, protocol = 'https:', suffix = '') {
			let url = protocol + '//';
			const parts = window.location.hostname.split('.');
			if (parts.length > 1 && module && module.length > 0) {
				parts[0] = module;
				const processedHostname = parts.join('.');
				url = url + processedHostname + suffix;
			} else {
				url = url + module + window.location.hostname + suffix;
			}
			return url;
		},
		resetShareFilter() {
			this.shareFilter = {
				shared: {
					byMe: true,
					withMe: true
				},
				owner: this.users?.users.map((e) => e.name) || [],
				scope: {
					public: true,
					smb: true,
					internal: true
				},
				permission: {
					manage: true,
					edit: true,
					view: true
				},
				expire: ExoirationTime.all,
				ownerInit: this.shareFilter.ownerInit
			};
		}
	}
});
