import {
	TransferFront,
	TransferItemInMemory,
	TransferStatus
} from 'src/utils/interface/transfer';
import {
	OlaresTaskBaseItem,
	OlaresTaskStatus,
	OlaresCloudUploadTask
} from '../abstractions/olaresTask/interface';
import { useTransfer2Store } from 'src/stores/transfer2';
import { dataAPIs } from 'src/api';
import { DriveType } from 'src/utils/interface/files';
import { appendPath } from 'src/api/files/path';
import OriginV2 from 'src/api/files/v2/origin';
import { driveTypeByFileTypeAndFileExtend } from 'src/api/files/v2/common/common';

export interface UploadTaskItem extends OlaresTaskBaseItem {
	action: TransferFront;
	// cancellable: boolean;
	dest: string;
	is_dir: boolean;
	progress: number;
	current_phase: number;
	total_phase: number;
	total_file_size: number;
	// dst_type: DriveType;
	dst_filename: string;
	failed_reason?: string;
	// dest: string;
}

export const getUploadTaskDriveType = (task: UploadTaskItem) => {
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

class CloudUpload implements OlaresCloudUploadTask {
	async setQueryResult(item: UploadTaskItem) {
		if (!item.transfer_id) {
			return item;
		}

		await this.formatTransferData(item);
		const transferStore = useTransfer2Store();
		const dstDriveType = getUploadTaskDriveType(item);
		const dataAPI = dataAPIs(dstDriveType);

		switch (item.status) {
			case OlaresTaskStatus.RUNNING:
			case OlaresTaskStatus.COMPLETED:
			case OlaresTaskStatus.PENDING:
				transferStore.onFileProgress(
					item.transfer_id,
					(item.total_file_size * item.progress) / 100,
					TransferFront.upload
				);
				if (item.status === OlaresTaskStatus.COMPLETED) {
					dataAPI.uploadSuccessRefreshData(item.transfer_id);
					await transferStore.onFileComplete(
						item.transfer_id,
						TransferFront.upload,
						item.current_phase
					);
				}

				break;
			case OlaresTaskStatus.PAUSED:
				await transferStore.pausedOrResumeTaskStatus(item.transfer_id, true);
				break;
			case OlaresTaskStatus.FAILED:
				transferStore.onFileError(item.transfer_id, TransferFront.upload);
				break;

			default:
				break;
		}
		return item;
	}

	async configUploadTask(
		taskIdItem: UploadTaskItem,
		info: {
			transfer_id: number;
			node: string;
		}
	) {
		taskIdItem.transfer_id = info.transfer_id;
		taskIdItem.node = info.node;
		taskIdItem.task_type = 'upload';
		return taskIdItem;
	}

	addTasks(items: TransferItemInMemory[]): UploadTaskItem[] {
		items = items.filter(
			(e) =>
				e.status == TransferStatus.Running || e.status == TransferStatus.Pending
		);

		if (items.length == 0) {
			return [];
		}
		const transferStore = useTransfer2Store();
		const oldItems =
			transferStore.taskCurrentSingleFiles[TransferFront.upload][
				items[0].task!
			];

		const combined = [...oldItems, ...items];

		const result = combined.filter((item, index, self) => {
			return self.findIndex((t) => t.id === item.id) === index;
		});

		transferStore.taskCurrentSingleFiles[TransferFront.upload][items[0].task!] =
			result;

		return items.map((e) => {
			return {
				id: e.phaseTaskId || '',
				status: OlaresTaskStatus.PENDING,
				transfer_id: e.id,
				node: e.node || '',
				pending: false,
				task_type: 'upload',
				retryCount: 0,
				action: TransferFront.upload,
				pause_able:
					e.pauseDisable == undefined ? true : e.pauseDisable == false,
				progress: 0,
				current_phase: e.currentPhase,
				total_phase: e.totalPhase,
				total_file_size: e.size,
				dst_type: e.driveType as DriveType,
				is_dir: e.isFolder,
				dest: e.to || '',
				dst_filename: e.name,
				failed_reason: e.message
			};
		});
	}

	private async formatTransferData(updatedTask: UploadTaskItem) {
		const transferStore = useTransfer2Store();

		const isDir = updatedTask.is_dir;

		const appendDestDir = appendPath(updatedTask.dest, isDir ? '/' : '');

		const dstDriveType = getUploadTaskDriveType(updatedTask);
		console.log('dstDriveType ===>', dstDriveType);

		updatedTask.dest = (dataAPIs(dstDriveType) as OriginV2).formatCopyPath(
			appendDestDir,
			updatedTask.dst_filename,
			isDir,
			dstDriveType
		);

		if (!updatedTask.transfer_id) {
			return;
		}

		const transferItem =
			transferStore.transferMap[updatedTask.transfer_id] ||
			transferStore.getSubTransferItem(
				TransferFront.upload,
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

		if (
			JSON.stringify(updatedTask.failed_reason) !==
			JSON.stringify(transferItem.message)
		) {
			transferItem.message = updatedTask.failed_reason;
			await transferStore.update(transferItem.id!, {
				message: updatedTask.failed_reason
			});
		}

		if (updatedTask.dest != transferItem.to) {
			transferItem.to = updatedTask.dest;
			transferItem.name = updatedTask.dst_filename;
			transferItem.path = updatedTask.dest;
			await transferStore.update(transferItem.id!, {
				to: updatedTask.dest,
				name: updatedTask.dst_filename,
				path: transferItem.path
			});
		}
	}
}

export default new CloudUpload();
