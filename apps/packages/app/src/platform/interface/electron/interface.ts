import { SupportLanguageType } from 'src/i18n';
import { ClientUploadFileType } from 'src/services/abstractions/transfer/upload';
import { TransferFront } from 'src/utils/interface/transfer';

export type IPCVpnEventName =
	| 'openVpn'
	| 'closeVpn'
	| 'statusVpn'
	| 'currentNodeId'
	| 'peersState'
	| 'electronVpnStatusUpdate'
	| 'resendCache';

export type IPCFilesEventName =
	| 'loginAccount'
	| 'setAutoSyncEnable'
	| 'isAutoSyncEnable'
	| 'hasLocalRepo'
	| 'openLocalRepo'
	| 'repoAddSync'
	| 'repoRemoveSync'
	| 'syncRepoImmediately'
	| 'repoSyncInfo'
	| 'defaultSyncSavePath'
	| 'selectSyncSavePath'
	| 'removeCurrentAccount'
	| 'updateSyncStatus'
	| 'updateSyncServerUrl'
	| 'repoSyncPathIsExist'
	| 'syncNotificationMessage';

export type IPCDownloadEventName =
	| 'newDownloadFile'
	| 'pauseOrResume'
	| 'allPauseOrResume'
	| 'removeDownloadItem'
	| 'removeAllDownloadItems'
	| 'newDownloadItem'
	| 'downloadItemUpdate'
	| 'downloadItemDone'
	| 'getDownloadPath'
	| 'setDownloadPath'
	| 'selectDownloadPath'
	| 'getDownloadedInfo'
	| 'getFolderSavePath'
	| 'deleteFolderSavePath';

export type IPCDownload2EventName =
	| 'start'
	| 'resume'
	| 'pause'
	| 'cancel'
	| 'getDownloadPath'
	| 'setDownloadPath'
	| 'newDownloadItem'
	| 'downloadItemUpdate'
	| 'downloadItemDone'
	| 'getFolderSavePath'
	| 'selectDownloadPath'
	| 'deleteFolderSavePath'
	| 'getDownloadInfo';

export type IPCUploadEventName =
	| 'getFilePath'
	| 'getPathInfo'
	| 'getSubItems'
	| 'start'
	| 'resume'
	| 'pause'
	| 'cancel'
	| 'getUploadInfo'
	| 'uploadItemUpdate'
	| 'uploadItemDone'
	| 'uploadCloudDriveTaskIdUpdate'
	| 'selectUploadFiles'
	| 'selectUploadFolder'
	| 'clearData';

export type IPCStoreEventName =
	| 'electronStoreGet'
	| 'electronStoreSet'
	| 'electronStoreRemove'
	| 'electronStoreClear';

export type IPCTransferEventName =
	| 'getDownloadData'
	| 'getUploadData'
	| 'getCompleteData'
	| 'clearCompleteData'
	| 'openFile'
	| 'openFileInFolder';

export type IPCWindowsHeaderMenusEventName =
	| 'minimize'
	| 'maximize'
	| 'close'
	| 'isMaximized'
	| 'winMove';

export type IPCSettingsEventName =
	| 'getAutomaticallyStartBoot'
	| 'setAutomaticallyStartBoot'
	| 'getAppInfo'
	| 'listenerNetworkUpdate'
	| 'getTaskPreventSleepBoot'
	| 'setTaskPreventSleepBoot'
	| 'openUrl'
	| 'updateLanguage'
	| 'checkNewVersion'
	| 'getUpdateStatus'
	| 'updateStatusEvent'
	| 'newVersionEvent'
	| 'skipNewVersion'
	| 'updateNewVersion';

export interface IOpenVpnInterface {
	server: string;
	authKey: string;
	acceptDns: boolean;
}

export interface IFilesLoginAccountInterface {
	url: string;
	username: string;
	authToken: string;
}

export interface IFilesRepoAddSyncInterface {
	worktree: string;
	repo_id: string;
	name: string;
	password: string;
	readonly: boolean;
}

export type DownloadItemState =
	| 'progressing'
	| 'completed'
	| 'cancelled'
	| 'interrupted';

export interface ITransferFile {
	id: string;
	fileName: string;
	icon: string;
	totalBytes: number;
	front: TransferFront;
	from: string;
	to: string;
	startTime: number;
	endTime: number;
	openPath: string;
	speed: number;
	progress: number;
	leftTimes: number; //unit s
}

export interface ITransferDownloadFile extends ITransferFile {
	receivedBytes: number;
	paused: boolean;
	state: DownloadItemState;
	downloadId: number;
	retryCount?: number;
}

export interface ITransferUploadFile extends ITransferFile {
	paused: boolean;
	extensionInfo: any;
}

export interface IPagination {
	pageIndex: number;
	pageCount: number;
}

export interface INewDownloadFile {
	url: string;
	fileName?: string;
	path: string;
	totalBytes: number;
	downloadId: number;
}

export interface IAppInfo {
	version: string;
	name: string;
	displayName: string;
	description: string;
	hostname: string;
}

export interface IFilesSyncStatus {
	syncing: boolean;
	error: boolean;
	pause: boolean;
	done: boolean;
}

export interface IUploadFileInfo {
	path: string;
	name: string;
	size: number;
	mimeType: string;
	isFolder: boolean;
	relatePath?: string;
}

export interface IUploadStartFile {
	id: number;
	baseUrl: string;
	uploadPath: string;
	filePath: string;
	fileType: ClientUploadFileType;
	repoId?: string;
	uniqueIdentifier?: string;
	relativePath?: string;
	size: number;

	moreInfo?: IUploadCloudParams;

	node?: string;
}

export interface IUploadCloudParams {
	taskId?: string;
	isFolder: boolean;
	account: string;
	// parentPath: string;
	cloudFilePath: string;
	folderName: string;
	relativePath: string;
}

export interface IDownloadStartFile {
	id: number;
	url: string;
	fileName?: string;
	path: string;
	size: number;
	savePath?: string;
}

export type LarePassElectronUpdateStatus =
	| 'normal'
	| 'checking'
	| 'latest'
	| 'error'
	| 'new'
	| 'cancel'
	| 'skip'
	| 'downloading'
	| 'downloaded'
	| 'prepareInstall'
	| 'restart';

declare global {
	interface Window {
		electron: {
			api: {
				vpn: {
					openVpn: (data: IOpenVpnInterface) => void;
					closeVpn: () => void;
					statusVpn: () => Promise<number>;
					currentNodeId: () => Promise<string>;
					peersState: () => Promise<any>;
					listenerVpnStatusUpdate: (
						callback: (_event: any, list: any) => void
					) => void;
					resendCache: (data: { server: string }) => Promise<void>;
				};
				files: {
					/** seafile **/
					loginAccount: (data: IFilesLoginAccountInterface) => Promise<boolean>;
					setAutoSyncEnable: (authSync: boolean) => Promise<boolean>;
					isAutoSyncEnable: () => Promise<boolean>;
					hasLocalRepo: (repo_id: string) => Promise<boolean>;
					openLocalRepo: (repo_id: string, subPath?: string) => void;
					repoAddSync: (data: IFilesRepoAddSyncInterface) => Promise<boolean>;
					repoRemoveSync: (repo_id: string) => Promise<boolean>;
					syncRepoImmediately: (repo_id: string) => Promise<boolean>;
					repoSyncInfo: (repo_id: string) => Promise<string>;
					defaultSyncSavePath: () => Promise<string>;
					selectSyncSavePath: () => Promise<string>;
					removeCurrentAccount: (
						data: IFilesLoginAccountInterface
					) => Promise<boolean>;
					updateSyncStatus: (status: IFilesSyncStatus) => void;
					updateSyncServerUrl: (url: string) => void;
					repoSyncPathIsExist: (path: string) => Promise<boolean>;
					syncNotificationMessage: () => void;
				};
				download: {
					newDownloadFile: (formData: INewDownloadFile) => Promise<void>;
					pauseOrResume: (
						pause: boolean,
						downloadId: number
					) => Promise<boolean>;
					allPauseOrResume: (
						pause: boolean,
						downloadIds: number[]
					) => Promise<boolean>;
					removeDownloadItem: (downloadId: number) => Promise<boolean>;
					removeAllDownloadItems: (downloadIds: number[]) => Promise<boolean>;
					clearDownloadDone: () => Promise<ITransferDownloadFile[]>;
					listenerNewDownloadItem: (
						callback: (_event: any, item: ITransferDownloadFile) => void
					) => void;
					listenerDownloadItemUpdate: (
						callback: (_event: any, item: ITransferDownloadFile) => void
					) => void;
					listenerDownloadItemDone: (
						callback: (_event: any, item: ITransferDownloadFile) => void
					) => void;
					getDownloadPath: () => Promise<string>;
					setDownloadPath: (path: string) => Promise<boolean>;
					selectDownloadPath: () => Promise<string>;
					getDownloadedInfo: (
						downloadId: number
					) => Promise<{ id: number; bytes: number } | undefined>;
					getFolderSavePath: (
						downloadId: number,
						folderName: string
					) => Promise<string>;

					deleteFolderSavePath: (downloadId: number) => Promise<boolean>;
				};
				download2: {
					start: (item: IDownloadStartFile) => Promise<boolean>;
					cancel: (downloadId: number) => Promise<boolean>;
					pause: (downloadId: number) => Promise<boolean>;
					resume: (downloadId: number, url?: string) => Promise<boolean>;
					complete: (downloadId: number) => Promise<boolean>;
					getDownloadInfo: (
						downloadId: number
					) => Promise<{ id: number; bytes: number } | undefined>;
					getDownloadPath: () => Promise<string>;
					setDownloadPath: (path: string) => Promise<boolean>;
					getFolderSavePath: (
						downloadId: number,
						folderName: string
					) => Promise<string>;
					selectDownloadPath: () => Promise<string>;
					deleteFolderSavePath: (downloadId: number) => Promise<boolean>;

					listenerNewDownloadItem: (
						callback: (_event: any, item: IDownloadStartFile) => void
					) => void;
					listenerDownloadItemUpdate: (
						callback: (_event: any, item: { id: number; bytes: number }) => void
					) => void;
					listenerDownloadItemDone: (
						callback: (
							_event: any,
							item: { id: number; code: number; message: string }
						) => void
					) => void;
				};

				upload: {
					getFilePath: (file: File) => Promise<string>;
					getPathInfo: (path: string) => Promise<IUploadFileInfo>;
					getSubItems: (path: string) => Promise<IUploadFileInfo[]>;
					start: (file: IUploadStartFile) => Promise<boolean>;
					getUploadInfo: (
						uploadId: number
					) => Promise<{ id: number; bytes: number } | undefined>;

					resume: (uploadId: number, baseUrl?: string) => Promise<boolean>;
					pause: (uploadId: number) => Promise<boolean>;
					cancel: (uploadId: number) => Promise<boolean>;
					clearData: (uploadId: number) => Promise<boolean>;

					listenerUploadItemUpdate: (
						callback: (
							_event: any,
							item: { uploadId: number; bytes: number }
						) => void
					) => void;
					listenerUploadItemDone: (
						callback: (
							_event: any,
							item: {
								id: number;
								code: number;
								message: string;
								taskId?: string;
							}
						) => void
					) => void;

					listerUploadCloudDriveTaskIdUpdate: (
						callback: (
							_event: any,
							item: { id: number; taskId: string }
						) => void
					) => void;

					selectUploadFiles: () => Promise<string[]>;
					selectUploadFolder: () => Promise<string>;
				};

				transfer: {
					getDownloadData: () => Promise<ITransferDownloadFile[]>;
					getUploadData: () => Promise<ITransferUploadFile[]>;
					getCompleteData: () => Promise<ITransferFile[]>;
					clearCompleteData: (list: string[]) => Promise<boolean>;
					openFileInFolder: (path: string) => Promise<boolean>;
					openFile: (path: string) => Promise<string>;
				};

				windows: {
					minimize: () => Promise<void>;
					maximize: () => Promise<void>;
					close: () => Promise<void>;
					isMaximized: () => Promise<boolean>;
					winMove: (isMove: boolean) => Promise<void>;
				};

				settings: {
					getAutomaticallyStartBoot: () => Promise<boolean>;
					setAutomaticallyStartBoot: (enable: boolean) => Promise<void>;
					getAppInfo: () => Promise<IAppInfo>;
					listenerNetworkUpdate: (
						callback: (_event: any, value: any) => void
					) => void;
					getTaskPreventSleepBoot: () => Promise<boolean>;
					setTaskPreventSleepBoot: (status: boolean) => Promise<void>;
					openUrl: (url: string) => Promise<void>;
					updateLanguage: (language: SupportLanguageType) => Promise<boolean>;
					checkNewVersion: () => Promise<boolean>;
					getUpdateStatus: () => Promise<{
						status: LarePassElectronUpdateStatus;
						process: number;
						message: string;
						version: string;
					}>;
					listenerUpdateStatusEvent: (
						callback: (
							_event: any,
							data: {
								status: LarePassElectronUpdateStatus;
								process: number;
								message: string;
								version: string;
							}[]
						) => void
					) => void;

					listenerNewVersionEvent: (
						callback: (
							_event: any,
							data: {
								currentVersion: string;
								lastVersion: string;
							}[]
						) => void
					) => void;

					skipNewVersion: (version: string) => Promise<void>;
					updateNewVersion: (autoUpdate: boolean) => Promise<void>;
				};
			};

			store: {
				get: (key: string) => any;
				set: (key: string, val: any) => void;
				remove: (key: string) => void;
				clear: () => void;
			};
		};
	}
}
