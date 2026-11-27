import { getFileIcon } from '@bytetrade/core';
import { appendPath } from 'src/api/files/path';
import { DriveAPI } from 'src/api/files/v2';
import { driveTypeByFileTypeAndFileExtend } from 'src/api/files/v2/common/common';
import { FileItem, useFilesStore } from 'src/stores/files';
import { useTransfer2Store } from 'src/stores/transfer2';
import {
	ArchiveConflict,
	ArchiveFormat,
	ArchiveSourceItem,
	getArchiveFileExtension
} from 'src/utils/interface/archive';
import { DriveType } from 'src/utils/interface/files';
import {
	TransferFront,
	TransferItem,
	TransferStatus
} from 'src/utils/interface/transfer';
import {
	OlaresArchiveTask,
	OlaresTaskBaseItem,
	OlaresTaskInterface,
	OlaresTaskStatus
} from '../abstractions/olaresTask/interface';
import { dataAPIs } from 'src/api';

export interface ArchiveTaskItem extends OlaresTaskBaseItem {
	action: 'compress' | 'extract';
	format?: ArchiveFormat;
	source: string;
	dest: string;
	filename: string;
	dst_filename: string;
	is_dir: boolean;
	total_file_size: number;
	transferred: number;
	progress: number;
	failed_reason: string;
	current_phase: number;
	total_phases: number;
	parentPath?: string;
	driveType?: DriveType;
	openFolderWhenDone?: boolean;
}

export interface ArchiveCreateParams {
	items: FileItem[];
	parentPath: string;
	driveType: DriveType;
	param?: string;
	format?: ArchiveFormat;
	node?: string;
	//Compress target folder name
	targetFileName?: string;
	//Extract target file name
	targetFolderName?: string;
	password?: string;
	splitVolume?: string | number;
	conflictStrategy?: ArchiveConflict;
	compressionLevel?: number;
	preserveSymbolicLinks?: boolean;
	extractToCurrentFolder?: boolean;
	openFolderWhenDone?: boolean;
}

const getArchiveTaskDriveType = (task: ArchiveTaskItem) => {
	let filePath = task.dest;
	if (filePath.startsWith('/')) {
		filePath = filePath.substring(1);
	}
	const filesPathSlitArray = filePath.split('/');
	if (filesPathSlitArray.length < 2) {
		return DriveType.Drive;
	}
	const fileType = filesPathSlitArray[0];
	const fileExtend = filesPathSlitArray[1];
	return driveTypeByFileTypeAndFileExtend(fileType, fileExtend);
};

const toArchiveSourceItem = (item: FileItem): ArchiveSourceItem => ({
	name: item.name,
	path: item.path,
	isDir: item.isDir,
	size: item.size
});

const getDefaultArchiveName = (items: FileItem[], format: ArchiveFormat) => {
	if (items.length === 1) {
		const item = items[0];
		const name = item.isDir ? item.name : item.name.replace(/\.[^/.]+$/, '');
		return `${name || item.name}.${format}`;
	}

	return `Archive.${format}`;
};

const getExtractFormat = (name?: string): ArchiveFormat | undefined => {
	const ext = getArchiveFileExtension(name);
	switch (ext) {
		case 'zip':
			return ArchiveFormat.Zip;
		case '7z':
			return ArchiveFormat.SevenZ;
		case 'tar':
			return ArchiveFormat.Tar;
		case 'tar.gz':
			return ArchiveFormat.TarGz;
		case 'tgz':
			return ArchiveFormat.TGZ;
		case 'tar.bz2':
			return ArchiveFormat.TarBz2;
		case 'tar.xz':
			return ArchiveFormat.TarXz;
		case 'gz':
		case 'gzip':
			return ArchiveFormat.GZIP;
		case 'bz2':
		case 'bzip2':
			return ArchiveFormat.BZIP2;
		case 'xz':
			return ArchiveFormat.XZ;
		default:
			return undefined;
	}
};

const normalizeSplitVolume = (splitVolume?: string | number) => {
	if (typeof splitVolume === 'number') {
		return splitVolume;
	}
	if (!splitVolume) {
		return undefined;
	}

	const [rawValue, rawUnit = 'MB'] = splitVolume.trim().split(/\s+/);
	const value = Number(rawValue);
	if (!Number.isFinite(value) || value <= 0) {
		return undefined;
	}

	const unit = rawUnit.toUpperCase();
	const mbValue =
		unit === 'GB' ? value * 1024 : unit === 'KB' ? value / 1024 : value;
	return Math.max(1, Math.ceil(mbValue));
};

const getPathName = (path?: string) => {
	const normalizedPath = (path || '').split('?')[0].replace(/\/+$/, '');
	const pathName = normalizedPath.split('/').filter(Boolean).pop();
	return pathName || '';
};

const getArchiveTaskDisplayName = (task: ArchiveTaskItem) => {
	return task.dst_filename || getPathName(task.dest) || task.filename;
};

const getArchiveFrontendPath = (path: string, driveType: DriveType) => {
	const dataAPI = dataAPIs(driveType);
	return dataAPI.displayPath({
		...dataAPI.pathToFrontendFile(path),
		fileType: ''
	});
};

class ArchiveTask implements OlaresArchiveTask {
	maxRetries = 1;
	async createCompressTask(params: ArchiveCreateParams) {
		const format = params.format || ArchiveFormat.Zip;
		const dstFilename =
			params.targetFileName || getDefaultArchiveName(params.items, format);
		const task = await this.createLocalTask({
			action: 'compress',
			format,
			items: params.items,
			parentPath: params.parentPath,
			driveType: params.driveType,
			param: params.param,
			node: params.node,
			targetFileName: dstFilename,
			isDir: false,
			password: params.password,
			splitVolume: params.splitVolume,
			conflictStrategy: params.conflictStrategy,
			compressionLevel: params.compressionLevel,
			preserveSymbolicLinks: params.preserveSymbolicLinks
		});

		return task;
	}

	async createExtractTask(params: ArchiveCreateParams) {
		const archive = params.items[0];
		if (!archive) {
			return;
		}

		const task = await this.createLocalTask({
			action: 'extract',
			format: params.format || getExtractFormat(archive.name),
			items: [archive],
			parentPath: params.parentPath,
			driveType: params.driveType,
			param: params.param,
			node: params.node,
			targetFolderName: params.targetFolderName,
			isDir: true,
			password: params.password,
			preserveSymbolicLinks: params.preserveSymbolicLinks,
			extractToCurrentFolder: params.extractToCurrentFolder,
			conflictStrategy: params.conflictStrategy,
			openFolderWhenDone: params.openFolderWhenDone
		});

		return task;
	}

	async setQueryResult(item: ArchiveTaskItem) {
		if (!item.transfer_id) {
			return item;
		}

		await this.formatTransferData(item);
		const transferStore = useTransfer2Store();
		const front =
			item.action === 'compress'
				? TransferFront.compress
				: TransferFront.extract;

		switch (item.status) {
			case OlaresTaskStatus.PENDING:
				transferStore.transferMap[item.transfer_id].status =
					TransferStatus.Pending;
				break;
			case OlaresTaskStatus.RUNNING:
				transferStore.onFileProgress(
					item.transfer_id,
					(item.total_file_size * item.progress) / 100,
					front
				);
				if (transferStore.transferMap[item.transfer_id].isPaused) {
					await transferStore.pausedOrResumeTaskStatus(item.transfer_id, false);
				}
				transferStore.transferMap[item.transfer_id].status =
					TransferStatus.Running;
				break;
			case OlaresTaskStatus.COMPLETED:
				await transferStore.onFileComplete(
					item.transfer_id,
					front,
					item.current_phase
				);
				DriveAPI.getAPI(
					item.driveType || getArchiveTaskDriveType(item)
				).uploadSuccessRefreshData(item.transfer_id);
				if (item.action === 'extract' && item.openFolderWhenDone) {
					await useFilesStore().setBrowserUrl(
						item.dest.endsWith('/') ? item.dest : `${item.dest}/`,
						item.driveType || getArchiveTaskDriveType(item)
					);
				}
				break;
			case OlaresTaskStatus.PAUSED:
				await transferStore.pausedOrResumeTaskStatus(item.transfer_id, true);
				break;
			case OlaresTaskStatus.FAILED:
				transferStore.onFileError(item.transfer_id, front, item.failed_reason);
				break;
			default:
				break;
		}

		return item;
	}

	formatArchiveToTransfer(
		task: ArchiveTaskItem,
		node: string,
		status = TransferStatus.Prepare,
		driveType?: DriveType
	): TransferItem {
		const front =
			task.action === 'compress'
				? TransferFront.compress
				: TransferFront.extract;
		const displayName = getArchiveTaskDisplayName(task);
		const type = task.is_dir ? 'Folder' : getFileIcon(displayName);

		return {
			name: displayName,
			path: task.dest,
			type,
			isFolder: task.is_dir,
			front,
			isPaused: false,
			size: task.total_file_size,
			from: task.source,
			to: task.dest,
			cancellable: task.pause_able,
			message: task.failed_reason,
			params: node,
			node,
			currentPhase: task.current_phase,
			totalPhase: task.total_phases,
			phaseTaskId: task.id,
			driveType,
			status,
			parentPath: task.parentPath,
			pauseDisable: !task.pause_able
		};
	}

	private async createLocalTask(params: {
		action: 'compress' | 'extract';
		format?: ArchiveFormat;
		items: FileItem[];
		parentPath: string;
		driveType: DriveType;
		param?: string;
		node?: string;
		targetFolderName?: string;
		targetFileName?: string;
		isDir: boolean;
		password?: string;
		splitVolume?: string | number;
		conflictStrategy?: ArchiveConflict;
		compressionLevel?: number;
		preserveSymbolicLinks?: boolean;
		extractToCurrentFolder?: boolean;
		openFolderWhenDone?: boolean;
	}) {
		const node = params.node || params.items[0]?.node || '';
		const { task_id, node: remoteNode } = await this.createRemoteTask({
			action: params.action,
			node,
			parentPath: params.parentPath,
			param: params.param,
			format:
				params.action === 'compress'
					? params.format || ArchiveFormat.Zip
					: params.format,
			items: params.items,
			targetFolderName: params.extractToCurrentFolder
				? ''
				: params.targetFolderName,
			targetFileName: params.targetFileName,
			password: params.password,
			splitVolume: params.splitVolume,
			conflictStrategy: params.conflictStrategy,
			compressionLevel: params.compressionLevel,
			preserveSymbolicLinks: params.preserveSymbolicLinks,
			driveType: params.driveType
		});
		const taskItem = this.buildTaskItem(params, remoteNode, task_id);
		const transferStore = useTransfer2Store();
		const transferItem = this.formatArchiveToTransfer(
			taskItem,
			remoteNode,
			TransferStatus.Prepare,
			params.driveType
		);

		const transferId = await transferStore.add(
			transferItem,
			params.action === 'compress'
				? TransferFront.compress
				: TransferFront.extract
		);
		taskItem.transfer_id = transferId;
		if (!transferStore.isUploadProgressDialogShow) {
			transferStore.isUploadProgressDialogShow = true;
		}
		return taskItem;
	}

	private buildTaskItem(
		params: {
			action: 'compress' | 'extract';
			format?: ArchiveFormat;
			items: FileItem[];
			parentPath: string;
			driveType: DriveType;
			param?: string;
			targetFolderName?: string;
			targetFileName?: string;
			isDir: boolean;
			extractToCurrentFolder?: boolean;
			openFolderWhenDone?: boolean;
		},
		node: string,
		taskId: string
	): ArchiveTaskItem {
		const totalSize = params.items.reduce(
			(total, item) => total + (item.size || 0),
			0
		);
		const parentPath = params.parentPath.endsWith('/')
			? params.parentPath
			: `${params.parentPath}/`;
		const dest = appendPath(
			parentPath,
			(params.extractToCurrentFolder
				? ''
				: params.action === 'compress'
				? params.targetFileName
				: params.targetFolderName) || '',
			params.action !== 'compress' && params.isDir ? '/' : ''
		);

		return {
			id: taskId,
			status: OlaresTaskStatus.PENDING,
			node,
			pending: false,
			task_type: params.action,
			retryCount: 0,
			maxRetries: this.maxRetries,
			pause_able: false,
			action: params.action,
			format: params.format,
			source: params.items.map((item) => item.path).join(','),
			dest,
			filename:
				params.items[0]?.name ||
				(params.action === 'compress'
					? params.targetFileName
					: params.targetFolderName) ||
				'',
			dst_filename:
				(params.action === 'compress'
					? params.targetFileName
					: params.targetFolderName) || '',
			is_dir: params.isDir,
			total_file_size: totalSize,
			transferred: 0,
			progress: 0,
			failed_reason: '',
			current_phase: 1,
			total_phases: 1,
			parentPath: params.parentPath,
			driveType: params.driveType,
			openFolderWhenDone: params.openFolderWhenDone
		};
	}

	private async createRemoteTask(params: {
		action: 'compress' | 'extract';
		node: string;
		parentPath: string;
		param?: string;
		format?: ArchiveFormat;
		items: FileItem[];
		targetFileName?: string;
		targetFolderName?: string;
		password?: string;
		splitVolume?: string | number;
		conflictStrategy?: ArchiveConflict;
		compressionLevel?: number;
		preserveSymbolicLinks?: boolean;
		driveType: DriveType;
	}) {
		return await DriveAPI.getAPI(params.driveType)[
			params.action === 'compress'
				? 'createArchiveCompressTask'
				: 'createArchiveExtractTask'
		]({
			node: params.node,
			parentPath: params.parentPath,
			param: params.param,
			format: params.format,
			items: params.items,
			targetFileName: params.targetFileName,
			targetFolderName: params.targetFolderName,
			password: params.password,
			compressionLevel: params.compressionLevel,
			splitVolume: normalizeSplitVolume(params.splitVolume),
			preserveSymbolicLinks: params.preserveSymbolicLinks ?? true,
			conflictStrategy: params.conflictStrategy
		} as any);
	}

	private async formatTransferData(updatedTask: ArchiveTaskItem) {
		updatedTask.dest = updatedTask.dest.replace(/\/\//g, '/');

		const transferStore = useTransfer2Store();

		const isDir = updatedTask.action == 'compress' ? false : updatedTask.is_dir;

		const dstDriveType = getArchiveTaskDriveType(updatedTask);

		if (updatedTask.action === 'extract') {
			updatedTask.dest = getArchiveFrontendPath(
				appendPath(updatedTask.dest, isDir ? '/' : ''),
				dstDriveType
			);
		} else {
			updatedTask.dest = dataAPIs(dstDriveType).formatCopyPath(
				updatedTask.dest,
				updatedTask.dst_filename,
				isDir,
				dstDriveType
			);
		}

		if (!updatedTask.transfer_id) {
			return;
		}

		const transferItem =
			transferStore.transferMap[updatedTask.transfer_id] ||
			transferStore.getSubTransferItem(
				updatedTask.action == 'compress'
					? TransferFront.compress
					: TransferFront.extract,
				updatedTask.transfer_id
			);
		if (!transferItem) {
			return;
		}
		if (updatedTask.total_file_size !== transferItem.size) {
			transferItem.size = updatedTask.total_file_size;
			await transferStore.update(transferItem.id!, {
				size: updatedTask.total_file_size
			});
		}

		const pauseDisable = !updatedTask.pause_able;
		if (pauseDisable !== transferItem.pauseDisable) {
			transferItem.pauseDisable = pauseDisable;
			await transferStore.update(transferItem.id!, {
				pauseDisable
			});
		}

		if (
			JSON.stringify(updatedTask.failed_reason) !==
			JSON.stringify(transferItem.message)
		) {
			transferItem.message = updatedTask.failed_reason;
			await transferStore.update(transferItem.id!, {
				message: updatedTask.failed_reason
			});
		}

		if (
			updatedTask.dest != transferItem.to ||
			transferItem.name != getArchiveTaskDisplayName(updatedTask)
		) {
			const displayName = getArchiveTaskDisplayName(updatedTask);
			transferItem.to = updatedTask.dest;
			transferItem.name = displayName;
			transferItem.path = updatedTask.dest;
			await transferStore.update(transferItem.id!, {
				to: updatedTask.dest,
				name: displayName,
				path: transferItem.path
			});
		}
	}
}

export default new ArchiveTask();
