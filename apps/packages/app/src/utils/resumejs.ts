import Resumablejs from './resumable.js';
// import MD5 from 'MD5';
import md5 from 'js-md5';
import { useDataStore } from '../stores/data';
import { useTransfer2Store } from '../stores/transfer2';
import {
	FileItem,
	FilePath,
	FilesIdType,
	useFilesStore
} from '../stores/files';

import { dataAPIs } from '../api';
import { Platform } from 'quasar';
import { useUserStore } from '../stores/user';
import { common } from '../api';
import { getParams } from './utils';
import url from '../utils/url';
import TransferClient from '../services/transfer';
import Taskmanager from '../services/olaresTask';
import { DriveType } from 'src/utils/interface/files';
import {
	TransferFront,
	TransferItem,
	TransferStatus
} from 'src/utils/interface/transfer';
import mime from 'mime';

export interface UploaderFileItem extends FileItem {
	speed: number;
	progress: number;
	leftTime: number;
	isPaused: boolean;
	bytes?: number;
	status: TransferStatus;
	uniqueIdentifier: string;
}

type ChunkStatus = 'pending' | 'uploading' | 'success' | 'error';
type PreprocessState = 0 | 1 | 2; // 0 = unprocessed, 1 = processing, 2 = finished

const ignores = ['.DS_Store'];
const targetMap: Record<string, string> = {};

interface ResumableObj {
	getOpt: (key: string) => any;
	fire: (event: string, file: ResumableFile, message?: any) => void;
	uploadNextChunk: () => void;
	removeFile: (file: ResumableFile) => void;
	on: (event: string, callback: () => void) => void;
}

class ResumableChunk {
	status: () => ChunkStatus;
	send: () => void;
	progress: (relative?: boolean) => number;
	abort: () => void;
	preprocessState: PreprocessState;

	constructor(props?: Partial<ResumableChunk>) {
		Object.assign(this, props);
	}
}

class ResumableFile {
	opts: any;
	getOpt: (key: string) => any;
	resumableObj: ResumableObj;
	file: File;
	fileName: string;
	size: number;
	relativePath: string;
	uniqueIdentifier: string;
	container: string;
	preprocessState: PreprocessState;
	error: string | null;
	remainingTime: number;
	isSaved: boolean;
	newFileName: string;
	chunks: ResumableChunk[];
	formData: any;
	_pause: boolean;
	id?: number;
	path: string;
	isPaused: () => boolean;
	progress: (relative?: boolean) => number;
	pause: () => void;
	upload: () => void;
	cancel: () => void;
	retry: () => void;
	addFiles: (files: any, event: any) => void;

	constructor(props?: Partial<ResumableFile>) {
		Object.assign(this, props);
	}
}

type UploadObjType = {
	maxNumberOfFilesForFileupload: number | undefined;
	maxUploadFileSize: number;
	resumableUploadFileBlockSize: number;
	simultaneousUploads: number;
	bitrateInterval: number;
};

// const LastUploadInfoRecord: Record<string, any> = {};

class Resumable {
	private props: {
		uploadInput: HTMLInputElement | null;
		origin_id: number;
	};

	public origin_id = 0;

	filePath?: FilePath;
	uploadFileLength = 0;

	private resumable: any = null;

	// public uploadFileList = <ResumableFile[]>[];
	// public notifiedFolders = <any[]>[];
	public timestamp = <number | null>null;
	public isUploadLinkLoaded = false;
	// public loadedd = 0;

	public uploadBitrate = 0;
	public totalProgress = 0;
	public retryFileList = <ResumableFile[]>[];
	public forbidUploadFileList = <ResumableFile[]>[];
	public allFilesUploaded = false;
	public addTranferFiles = <TransferItem[]>[];
	private waitAddSubtasksMap: Record<number, ResumableFile[]> = [];

	public uploadObj: UploadObjType = <UploadObjType>{
		maxNumberOfFilesForFileupload: undefined,
		maxUploadFileSize: 1024 * 100,
		resumableUploadFileBlockSize: 8,
		simultaneousUploads: 1,
		bitrateInterval: 1000
	};

	constructor() {
		this.init();
	}

	init() {
		// this.setupResumable();
		this.resumable = new Resumablejs({
			target: this.getTarget,
			query: this.setQuery || {},
			fileType: [],
			maxFiles: this.uploadObj.maxNumberOfFilesForFileupload,
			maxFileSize: this.uploadObj.maxUploadFileSize * 1024 * 1024,
			method: 'post',
			testChunks: false,
			headers: this.setHeaders || {},
			withCredentials: false,
			chunkSize:
				this.uploadObj.resumableUploadFileBlockSize * 1024 * 1024 ||
				1 * 1024 * 1024,
			simultaneousUploads: this.uploadObj.simultaneousUploads || 1,
			generateUniqueIdentifier: this.generateUniqueIdentifier,
			isIgnoreFile: this.isIgnoreFile,
			forceChunkSize: true,
			maxChunkRetries: 3,
			minFileSize: 0,
			xhrTimeout: 600 * 1000,
			uploadSuccessCode: this.uploadSuccess,
			chunkMd5: this.chunkMd5,
			chunkRetryInterval: 5 * 1000,
			createEmptyFile: this.createEmptyFile
		});
		this.bindEventHandler();
	}

	public setupResumable(props: {
		uploadInput: HTMLInputElement | null;
		origin_id: number;
	}) {
		this.props = props;
		this.origin_id = this.props.origin_id;
		this.filePath = undefined;
		this.resumable.assignBrowse(this.props.uploadInput, true);
	}

	public getFiles() {
		return this.resumable.files as ResumableFile[];
	}

	public addFilesToUpload(
		target: any,
		filePath: FilePath,
		origin_id = FilesIdType.PAGEID
	) {
		this.origin_id = origin_id;
		this.filePath = filePath;
		this.uploadFileLength = target.target.files.length;
		this.resumable.addFiles(target.target.files, target);
	}

	public updateSubtasksSuccess(
		task: number,
		ids: number[],
		identifys?: string[]
	) {
		const files = this.waitAddSubtasksMap[task];

		if (files.length == ids.length) {
			files.forEach((e, index) => (e.id = ids[index]));
		} else {
			if (identifys && ids.length == identifys.length) {
				ids.forEach((id, index) => {
					const identify = identifys[index];
					const file = files.find((e) => e.uniqueIdentifier == identify);
					if (file) {
						file.id = id;
					}
				});
			}
		}
		delete this.waitAddSubtasksMap[task];
		this.startUpload();
	}

	private bindEventHandler() {
		this.resumable.on('chunkingComplete', this.onChunkingComplete.bind(this));
		this.resumable.on('fileAdded', this.onFileAdded.bind(this));
		this.resumable.on('filesAddedComplete', this.filesAddedComplete.bind(this));
		this.resumable.on('fileProgress', this.onFileProgress.bind(this));
		this.resumable.on('fileSuccess', this.onFileUploadSuccess.bind(this));
		this.resumable.on('progress', this.onProgress.bind(this));
		this.resumable.on('complete', this.onComplete.bind(this));
		this.resumable.on('fileError', this.onFileError.bind(this));
		this.resumable.on('error', this.onError.bind(this));
		// this.resumable.on('dragstart', this.onDragStart.bind(this));
		this.resumable.on('pause', this.onPause.bind(this));
		this.resumable.on('fileRetry', this.onFileRetry.bind(this));
		this.resumable.on('beforeCancel', this.onBeforeCancel.bind(this));
		this.resumable.on('cancel', this.onCancel.bind(this));
	}

	private async onChunkingComplete(resumableFile: any) {
		this.uploadFileLength -= 1;
		if (this.allFilesUploaded === true) {
			this.allFilesUploaded = false;
		}

		resumableFile.relativePath = resumableFile.file.fullPath
			? resumableFile.file.fullPath
			: resumableFile.relativePath;

		//get parent_dir relative_path
		const path = await this.getCurrentPath();

		if (!path.pathname.endsWith('/')) {
			path.pathname = path.pathname + '/';
		}
		const fileName = resumableFile.fileName;
		const relativePath = resumableFile.relativePath;
		const isFile = fileName === relativePath;

		//update formdata
		resumableFile.formData = {};
		if (isFile) {
			// upload file
			resumableFile.formData = {
				parent_dir: path.pathname,
				...path
			};
		} else {
			// upload folder
			const relative_path = relativePath.slice(
				0,
				relativePath.lastIndexOf('/') + 1
			);
			resumableFile.formData = {
				parent_dir: path.pathname,
				relative_path: relative_path,
				...path
				// md5: MD5
			};
		}
		if (this.filePath && this.uploadFileLength === 0) {
			this.filePath = undefined;
		}
	}

	private async onFileAdded() {}

	private uploadSuccess(obj: any) {
		const google_upload_interface = '/drive/direct_upload_file';
		const google_upload_interface_index = obj.responseURL.indexOf(
			google_upload_interface
		);
		if (google_upload_interface_index >= 0) {
			try {
				const responseText = JSON.parse(obj.responseText);
				const status_code = responseText.status_code;
				return status_code == 'SUCCESS';
			} catch (error) {
				return false;
			}
		}

		return [201, 200];
	}

	private async chunkMd5(bytes: Blob) {
		try {
			// only main support
			if (!process.env.VERSIONTAG || process.env.VERSIONTAG == 'MAIN') {
				const arrayBuffer = await bytes.arrayBuffer();
				const uint8Array = new Uint8Array(arrayBuffer);
				const mds = md5(uint8Array);
				return mds;
			}
			return '';
		} catch (error) {
			console.error('MD5 error:', error);
			return null;
		}
	}

	private getTarget(params: string[]) {
		const identifierEntry = params.find((entry) =>
			entry.startsWith('resumableIdentifier=')
		);
		const resumableIdentifier = identifierEntry
			? identifierEntry.split('=')[1]
			: null;
		if (!resumableIdentifier) {
			return '';
		}
		const target = targetMap[decodeURIComponent(resumableIdentifier)] || '';

		const google_upload_interface = '/drive/direct_upload_file';
		const google_upload_interface_index = target.indexOf(
			google_upload_interface
		);

		if (google_upload_interface_index >= 0) {
			const task_id = target.slice(
				google_upload_interface_index + google_upload_interface.length + 1,
				target.indexOf('?')
			);

			const snakeCaseArray: string[] = [];
			params.map((item) => {
				let key = item.split('=')[0];
				const value = item.split('=')[1];

				if (key === 'resumableFilename') {
					key = 'resumable_file_name';
				}
				const snakeKey = key.replace(/([A-Z])/g, '_$1').toLowerCase();
				snakeCaseArray.push(`${snakeKey}=${value}`);
			});

			const parmaString = snakeCaseArray.join('&');
			return (
				google_upload_interface + '?' + parmaString + '&task_id=' + task_id
			);
		}

		return target;
	}

	private groupedByFolder(data) {
		return data.reduce((acc, obj) => {
			const keys = Object.keys(acc);
			let key = obj.formData.relative_path || '/';
			if (key != '/') key = key.slice(0, key.indexOf('/'));

			const hasKey = keys.find((item) => key == item);
			if (!hasKey) {
				acc[key] = [];
				acc[key].push(obj);
			} else {
				acc[hasKey].push(obj);
			}
			return acc;
		}, {});
	}

	private async filesAddedComplete(resumable: any, files: ResumableFile[]) {
		if (files.length == 0) {
			return;
		}
		const path = files[0].formData;

		let isDir = false;
		if (files[0].newFileName !== files[0].relativePath) {
			isDir = true;
		}

		if (
			[DriveType.GoogleDrive, DriveType.Dropbox, DriveType.Awss3].includes(
				path.driveType
			) &&
			!isDir
		) {
			this.getGoogleUploadLink(resumable, files, path);
		} else {
			this.getUploadLink(resumable, files, path, isDir);
		}
	}

	private async getGoogleUploadLink(resumable, files, path) {
		const dataAPI = dataAPIs(path.driveType, this.origin_id);
		for (let index = 0; index < files.length; index++) {
			const file = files[index];
			const uploadLink = await dataAPI.getFileServerUploadLink(
				path.pathname,
				path.repoId
			);

			this.useUploadLink(resumable, [file], path, uploadLink);
		}
	}

	private async getUploadLink(resumable, files: ResumableFile[], path, isDir) {
		const dataAPI = dataAPIs(path.driveType, this.origin_id);

		let dirName = '';
		if (isDir) {
			if (files[0].newFileName !== files[0].relativePath) {
				dirName = files[0].relativePath.slice(
					0,
					files[0].relativePath.indexOf('/')
				);
			}
		}

		const uploadLink = await dataAPI.getFileServerUploadLink(
			path.pathname,
			path.repoId,
			dirName
		);

		this.useUploadLink(resumable, files, path, uploadLink);
	}

	private async useUploadLink(
		resumable,
		files: ResumableFile[],
		path,
		uploadLink
	) {
		const dataStore = useDataStore();

		const baseURL = dataStore.baseURL();
		const targetUrl = baseURL + uploadLink;

		const groupFolders = this.groupedByFolder(files);
		let taskId = '';

		if (uploadLink.indexOf('/drive/direct_upload_file') > -1) {
			taskId = uploadLink.substring(uploadLink.lastIndexOf('/') + 1);
			if (!resumable.formData) {
				resumable.formData = {};
			}
			resumable.formData.task_id = taskId;
			this.resumable.opts.method = 'octet';
		} else {
			this.resumable.opts.method = 'post';
		}

		for (const relative_path in groupFolders) {
			await this.addSumableToTransfer(
				groupFolders[relative_path],
				path,
				targetUrl,
				taskId
			);
		}
	}

	private async addSumableToTransfer(
		files: ResumableFile[],
		path,
		targetUrl,
		task_id
	) {
		const dataAPI = dataAPIs(path.driveType, this.origin_id);
		const transferStore = useTransfer2Store();
		const tasks: TransferItem[] = [];

		for (let index = 0; index < files.length; index++) {
			const file = files[index];

			const relative_path = dataAPI.getResumePath(
				path.fullPath,
				files[index].relativePath
			);
			file.path = relative_path;
			file.pause();
			const task = await this.resumableToTransferItem(file);
			tasks.push(task);

			targetMap[file.uniqueIdentifier] = targetUrl;
		}

		const isFile = files[0].fileName === files[0].relativePath;

		if (isFile) {
			const transferStore = useTransfer2Store();
			for (let index = 0; index < tasks.length; index++) {
				const task = tasks[index];
				const id = await transferStore.add(task, TransferFront.upload);
				if (id < 0) {
					continue;
				}
				files[index].id = id;
			}
			if (tasks.length == 1) {
				await this.resumableUpload(files[0], task_id);
			}
			this.startUpload();
			return;
		}

		const dirItem = JSON.parse(JSON.stringify(tasks[0]));

		dirItem.isFolder = true;
		dirItem.name = files[0].formData.relative_path.slice(
			0,
			files[0].formData.relative_path.indexOf('/')
		);
		dirItem.folderTotalCount = this.addTranferFiles.length;
		dirItem.size = this.addTranferFiles.reduce((accumulator, item) => {
			return accumulator + item.size;
		}, 0);
		dirItem.uniqueIdentifier = '';

		dirItem.path = path.fullPath + dirItem.name + '/';
		dirItem.totalPhase = 1;

		const taskId = await transferStore.add(dirItem, TransferFront.upload);
		if (taskId < 0) {
			return;
		}

		TransferClient.waitAddSubtasks[taskId] = {
			finished: true,
			subtasks: tasks,
			offset: transferStore.transferMap[taskId].folderTotalCount
		};
		this.waitAddSubtasksMap[taskId] = files;
	}

	private async startUpload() {
		const transferStore = useTransfer2Store();
		transferStore.isUploadProgressDialogShow = true;
	}

	private async resumableUpload(resumableFile: any, taskId: string) {
		const path = resumableFile.formData;

		const dataAPI = dataAPIs(path.driveType, this.origin_id);
		await dataAPI
			.getFileUploadedBytes(
				path.pathname,
				resumableFile.fileName,
				path.repoId,
				taskId
			)
			.then((res) => {
				const uploadedBytes = res.uploadedBytes;
				const blockSize =
					this.uploadObj.resumableUploadFileBlockSize * 1024 * 1024 ||
					1024 * 1024;

				const offset = Math.floor(uploadedBytes / blockSize);

				resumableFile.markChunksCompleted(offset);
				this.resumable.upload();
			})
			.catch((error) => {
				console.error('errMessage000', error);
			});
	}

	private async onFileProgress(resumableFile: ResumableFile) {
		const transferStore = useTransfer2Store();
		const item = await this.resumableToTransferItemMemory(resumableFile);
		transferStore.onFileProgress(item.id, item.bytes, item.front);
	}

	private onProgress() {
		const progress = Math.round(this.resumable.progress() * 100);
		this.totalProgress = progress;
	}

	private async onFileUploadSuccess(resumableFile, message) {
		const transferStore = useTransfer2Store();
		delete targetMap[resumableFile.uniqueIdentifier];

		if (resumableFile.id) {
			const dataAPI = dataAPIs(
				resumableFile.formData.driveType,
				this.origin_id
			);
			const messageRes = JSON.parse(message);
			let nextTaskId = undefined;
			if (messageRes && Array.isArray(messageRes)) {
				const item = messageRes[0];
				if (item && item.taskId && item.taskId.length > 0) {
					await Taskmanager.addTask(
						item.taskId,
						resumableFile.formData.node,
						'upload',
						{
							node: resumableFile.formData.node,
							transfer_id: resumableFile.id
						}
					);
					nextTaskId = item.taskId;
				}
			}

			const res = await dataAPI.transferItemUploadSuccessResponse(
				resumableFile.id,
				message
			);

			if (res) {
				if (!nextTaskId) {
					dataAPI.uploadSuccessRefreshData(resumableFile.id);
				}
				await transferStore.onFileComplete(
					resumableFile.id,
					TransferFront.upload,
					1,
					nextTaskId
				);
			}
		}
	}

	private async onFileError(resumableFile, message) {
		const transferStore = useTransfer2Store();
		const item = await this.resumableToTransferItemMemory(resumableFile);
		let error = '';
		if (!message) {
			error = 'Network error';
		} else {
			try {
				error = message;
				const errorMessage = JSON.parse(message);
				error =
					errorMessage.message ||
					errorMessage.fail_reason ||
					errorMessage.error;
			} catch (error) {
				/* empty */
			}
		}

		if (item.id) {
			await transferStore.onFileError(item.id, item.front, error);
		}
	}

	private onComplete() {
		// this.notifiedFolders = [];
		// reset upload link loaded
		this.isUploadLinkLoaded = false;
		this.allFilesUploaded = true;
	}

	private onError() {
		// reset upload link loaded
		this.isUploadLinkLoaded = false;
	}

	private onPause() {
		this.resumable.upload();
	}

	private onFileRetry() {
		// todo, cancel upload file, uploded again;
		console.log('onFileRetry');
		console.log('onFileRetry-upload');
	}

	private onBeforeCancel() {
		// todo, giving a pop message ?
		console.log('onBeforeCancel');
	}

	private onCancel() {
		console.log('onCancel');
	}

	private setQuery(resumableFile: ResumableFile) {
		const formData = resumableFile.formData;

		return {
			...formData,
			resumableType:
				mime.getType(resumableFile.fileName) || 'application/octet-stream'
		};
	}

	private setHeaders(resumableFile, resumable) {
		const offset = resumable.offset;
		const chunkSize = resumable.getOpt('chunkSize');
		const fileSize = resumableFile.size === 0 ? 1 : resumableFile.size;
		const startByte = offset !== 0 ? offset * chunkSize : 0;
		let endByte = Math.min(fileSize, (offset + 1) * chunkSize) - 1;

		if (
			fileSize - resumable.endByte < chunkSize &&
			!resumable.getOpt('forceChunkSize')
		) {
			endByte = fileSize;
		}

		const headers = {
			Accept: 'application/json; text/javascript, */*; q=0.01',
			'Content-Disposition':
				'attachment; filename="' + encodeURI(resumableFile.fileName) + '"',
			'Content-Range': 'bytes ' + startByte + '-' + endByte + '/' + fileSize
		};

		if (Platform.is.nativeMobile) {
			const userStore = useUserStore();
			headers['X-Authorization'] = userStore.current_user?.access_token;
		}
		return headers;
	}

	private generateUniqueIdentifier(file) {
		const relativePath =
			file.fullPath ||
			file.webkitRelativePath ||
			file.relativePath ||
			file.fileName ||
			file.name;
		return md5(relativePath + new Date()) + relativePath;
	}

	private isIgnoreFile(file) {
		const item = ignores.findIndex((e) => e == file.name);
		return item >= 0;
	}

	private async createEmptyFile(resumableFile) {
		if (resumableFile.id) {
			const dataAPI = dataAPIs(resumableFile.formData.driveType);
			if ((dataAPI as any).uploadEmptyFile) {
				try {
					const res = await (dataAPI as any).uploadEmptyFile(
						resumableFile.id,
						''
					);
					return res;
				} catch (error) {
					return false;
				}
			}
			return true;
		}
		return false;
	}

	public onCloseUploadDialog() {
		// this.loadedd = 0;
		// this.resumable.files = [];
		// reset upload link loaded
		this.isUploadLinkLoaded = false;
		// const filesStore = useFilesStore();
		const transferStore = useTransfer2Store();

		transferStore.isUploadProgressDialogShow = false;
		// filesStore.uploadFileList[this.origin_id] = [];
	}

	public async resumableToTransferItem(
		resumableFile: any
	): Promise<TransferItem> {
		const path = resumableFile.formData;

		[DriveType.GoogleDrive, DriveType.Dropbox, DriveType.Awss3].includes(
			path.driveType
		);

		const driveType = common().formatUrltoDriveType(resumableFile.path);
		let totalPhase = 1;
		if (
			driveType &&
			[DriveType.GoogleDrive, DriveType.Dropbox, DriveType.Awss3].includes(
				driveType
			)
		) {
			totalPhase = 2;
		}

		const item: TransferItem = {
			name: resumableFile.newFileName,
			path: resumableFile.path,
			type: resumableFile.file.type,
			isFolder: false,
			driveType: common().formatUrltoDriveType(resumableFile.path),
			front: TransferFront.upload,
			status: this.uploadState(resumableFile),
			url: '',
			startTime: new Date().getTime(),
			endTime: 0,
			updateTime: new Date().getTime(),
			from: '',
			to: resumableFile.path,
			size: resumableFile.size,
			message: '',
			uniqueIdentifier: resumableFile.uniqueIdentifier,
			repo_id: path.repoId || '',
			isPaused: false,
			id: resumableFile.id,
			parentPath: resumableFile.formData.parent_dir,
			node: path.node,
			currentPhase: 1,
			totalPhase: totalPhase,
			phaseTaskId: ''
		};

		return item;
	}

	public async resumableToTransferItemMemory(
		resumableFile: any
	): Promise<{ id: number | undefined; front: TransferFront; bytes: number }> {
		const progress = resumableFile.progress();
		const item: {
			id: number | undefined;
			front: TransferFront;
			bytes: number;
		} = {
			front: TransferFront.upload,
			bytes: progress * resumableFile.size,
			id: resumableFile.id || undefined
		};

		return item;
	}

	public uploadState(resumableFile: any) {
		let uploadState = TransferStatus.Running;

		if (resumableFile.error) {
			uploadState = TransferStatus.Error;
		} else if (resumableFile.isPaused()) {
			uploadState = TransferStatus.Running;
		} else {
			if (resumableFile.remainingTime === 0 && !resumableFile.isSaved) {
				uploadState = TransferStatus.Checking;
			}

			if (resumableFile.isSaved) {
				uploadState = TransferStatus.Completed;
			}
		}

		if (uploadState === TransferStatus.Running) {
			if (!resumableFile.isUploading()) {
				uploadState = TransferStatus.Pending;
				// if (resumableFile.remainingTime === -1) {
				// 	uploadState = TransferStatus.Prepare;
				// } else {
				// 	uploadState = TransferStatus.Pending;
				// }
			} else {
				uploadState = TransferStatus.Running;
			}
		}

		return uploadState;
	}

	private async getCurrentPath() {
		let fullPath = '';
		let fullPathSplit: string[] = [];
		let driveType = DriveType.Drive;

		const filesStore = useFilesStore();
		if (!this.filePath) {
			if (this.origin_id) {
				driveType = filesStore.currentPath[this.origin_id].driveType;
				fullPathSplit = [filesStore.currentPath[this.origin_id].path];
				if (filesStore.currentPath[this.origin_id].param) {
					fullPathSplit.push(filesStore.currentPath[this.origin_id].param);
				}
				fullPath = fullPathSplit.join('');
			} else {
				driveType =
					common().formatUrltoDriveType(url.getWindowPathname()) ||
					DriveType.Drive;
				fullPath = url.getWindowFullpath();
				fullPathSplit = url.getWindowFullpath().split('?');
			}
		} else {
			driveType = this.filePath.driveType;
			fullPathSplit = [this.filePath.path];
			if (this.filePath.param) {
				fullPathSplit.push(this.filePath.param);
			}
			fullPath = fullPathSplit.join('');
		}

		const dataAPI = dataAPIs(driveType);

		const pathname =
			(await dataAPI.formatUploaderPath(fullPathSplit[0])) || '/';
		const repoId = fullPathSplit[1] ? getParams(fullPathSplit[1], 'id') : '';
		const node = dataAPI.getUploadNode();
		const res = {
			fullPath,
			pathname: decodeURIComponent(pathname),
			repoId,
			driveType: driveType,
			node
		};

		return res;
	}
}

export default new Resumable();
