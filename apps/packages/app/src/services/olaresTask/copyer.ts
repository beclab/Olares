import {
	TransferFront,
	TransferItem,
	TransferStatus
} from 'src/utils/interface/transfer';
import {
	OlaresTaskBaseItem,
	OlaresTaskStatus,
	OlaresCopyTask
} from '../abstractions/olaresTask/interface';
import { DriveType } from 'src/utils/interface/files';
import { useTransfer2Store } from 'src/stores/transfer2';
// import { dataAPIs } from 'src/api';
import { getFileIcon } from '@bytetrade/core';
import { appendPath } from 'src/api/files/path';
// import OriginV2 from 'src/api/files/v2/origin';
import { DriveAPI } from 'src/api/files/v2';
import { driveTypeByFileTypeAndFileExtend } from 'src/api/files/v2/common/common';

export interface CopyTaskItem extends OlaresTaskBaseItem {
	action: 'copy' | 'move';
	cancellable: boolean;
	dest: string;
	// dst_type: DriveType;
	log: string[];
	progress: number;
	relation_node: string;
	relation_task_id: string;
	source: string;
	// src_type: DriveType;
	total_file_size: number;
	transferred: number;
	file_type: string;
	filename: string;
	is_dir: boolean;
	failed_reason: string;
	dst_filename: string;
	current_phase: number;
	total_phases: number;
	pause_able: boolean;
}

export const getCopyTaskDriveType = (task: CopyTaskItem, isSrc = false) => {
	let filePath = isSrc ? task.source : task.dest;
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

class Copyer implements OlaresCopyTask {
	async setQueryResult(item: CopyTaskItem): Promise<CopyTaskItem> {
		if (!item.transfer_id) {
			return item;
		}
		const dstDriveType = getCopyTaskDriveType(item);
		const dataAPI = DriveAPI.getAPI(dstDriveType);
		await this.formatTransferData(item);

		const transferStore = useTransfer2Store();

		switch (item.status) {
			case OlaresTaskStatus.PENDING:
				transferStore.transferMap[item.transfer_id].status =
					TransferStatus.Pending;
				break;

			case OlaresTaskStatus.RUNNING:
				transferStore.onFileProgress(
					item.transfer_id,
					(item.total_file_size * item.progress) / 100,
					item.action == 'copy' ? TransferFront.copy : TransferFront.move
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
					item.action == 'copy' ? TransferFront.copy : TransferFront.move,
					item.current_phase
				);
				console.log('refresh 111', dataAPI);

				dataAPI.uploadSuccessRefreshData(item.transfer_id);

				break;
			case OlaresTaskStatus.PAUSED:
				await transferStore.pausedOrResumeTaskStatus(item.transfer_id, true);
				break;

			case OlaresTaskStatus.FAILED:
				transferStore.onFileError(
					item.transfer_id,
					item.action == 'copy' ? TransferFront.copy : TransferFront.move
				);
				break;

			default:
				break;
		}
		return item;
	}

	private async formatTransferData(updatedTask: CopyTaskItem) {
		const transferStore = useTransfer2Store();

		const isDir = updatedTask.is_dir;
		const appendSourceDir = appendPath(updatedTask.source, isDir ? '/' : '');
		const appendDestDir = appendPath(updatedTask.dest, isDir ? '/' : '');

		const dstDriveType = getCopyTaskDriveType(updatedTask);
		const srcDriveType = getCopyTaskDriveType(updatedTask, true);

		updatedTask.source = DriveAPI.getAPI(srcDriveType).formatCopyPath(
			appendSourceDir,
			updatedTask.filename,
			isDir,
			srcDriveType
		);

		updatedTask.dest = DriveAPI.getAPI(dstDriveType).formatCopyPath(
			appendDestDir,
			updatedTask.dst_filename,
			isDir,
			dstDriveType
		);
		console.log('updatedTask.dest ===>', updatedTask.dest);

		if (
			!updatedTask.transfer_id ||
			!transferStore.transferMap[updatedTask.transfer_id]
		) {
			return;
		}
		const transterItem = transferStore.transferMap[updatedTask.transfer_id];
		if (updatedTask.total_file_size !== transterItem.size) {
			transterItem.size = updatedTask.total_file_size;
			await transferStore.update(transterItem.id!, {
				size: updatedTask.total_file_size
			});
		}

		if (
			updatedTask.current_phase != transterItem.currentPhase ||
			updatedTask.total_phases != transterItem.totalPhase
		) {
			transterItem.currentPhase = updatedTask.current_phase;
			transterItem.totalPhase = updatedTask.total_phases;
			await transferStore.update(transterItem.id!, {
				currentPhase: updatedTask.current_phase,
				totalPhase: updatedTask.total_phases
			});
		}

		if (
			JSON.stringify(updatedTask.failed_reason) !==
			JSON.stringify(transterItem.message)
		) {
			transterItem.message = updatedTask.failed_reason;
			await transferStore.update(transterItem.id!, {
				message: updatedTask.failed_reason
			});
		}

		if (updatedTask.dest != transterItem.to) {
			transterItem.to = updatedTask.dest;
			await transferStore.update(transterItem.id!, {
				to: updatedTask.dest
			});
		}
	}

	async addCopyToTransfer(
		taskIdItem: CopyTaskItem,
		info: {
			dst_drive_type?: DriveType;
			type: TransferFront;
			node: string;
		}
	) {
		const dstDriveType = getCopyTaskDriveType(taskIdItem);

		await this.formatTransferData(taskIdItem);

		const transferItem = this.formatCopyToTransfer(
			taskIdItem,
			info.node,
			TransferStatus.Prepare,
			info.dst_drive_type ? info.dst_drive_type : dstDriveType
		);

		const transferStore = useTransfer2Store();
		const transferId = await transferStore.add(transferItem, info.type);
		taskIdItem.transfer_id = transferId;
		taskIdItem.node = info.node;
		return taskIdItem;
	}

	formatCopyToTransfer(
		taskIdItem: CopyTaskItem,
		node: string,
		status = TransferStatus.Prepare,
		driveType?: DriveType
	) {
		const type = taskIdItem.is_dir ? '' : getFileIcon(taskIdItem.filename);

		const transferItem: TransferItem = {
			name: taskIdItem.dst_filename,
			path: taskIdItem.dest,
			type,
			isFolder: taskIdItem.is_dir,
			front:
				taskIdItem.action == 'copy' ? TransferFront.copy : TransferFront.move,
			isPaused: false,
			size: taskIdItem.total_file_size,
			from: taskIdItem.source,
			to: taskIdItem.dest,
			cancellable: taskIdItem.cancellable,
			message: taskIdItem.failed_reason,
			params: node,
			node: node,
			currentPhase: taskIdItem.current_phase,
			totalPhase: taskIdItem.total_phases,
			phaseTaskId: taskIdItem.id,
			driveType: driveType,
			status: status,
			pauseDisable: !taskIdItem.pause_able || false
		};

		if (transferItem.driveType === DriveType.Sync) {
			let cur_path = transferItem.path.slice(0, transferItem.path.indexOf('?'));
			if (transferItem.isFolder) {
				cur_path = cur_path.slice(0, -1);
			}
			const lastIndex = cur_path.lastIndexOf('/');
			const parentPath = cur_path.split('/').slice(3, lastIndex).join('/');
			transferItem.parentPath = parentPath ? parentPath : '/';
		}

		return transferItem;
	}
}

export default new Copyer();
