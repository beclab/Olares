import { DriveType } from 'src/utils/interface/files';
import { DownloadRecord } from 'src/utils/interface/rss';
import {
	TransferFront,
	TransferItem,
	TransferStatus
} from 'src/utils/interface/transfer';
import { TransferDatabase } from 'src/utils/interface/transferDB';

export enum WiseWSType {
	DOWNLOAD_PROCESS = 'download_process'
}

const getCloudTaskById = async (task_id: string, db: TransferDatabase) => {
	const result = await db.transferData
		.where('uniqueIdentifier')
		.equals(task_id)
		.toArray();
	return result;
};

const getTaskId = (recordId: string | number, terminusId?: string) => {
	if (!terminusId) {
		return null;
	}
	return terminusId + '_' + recordId;
};
export const wiseInsertTransferItem = (
	downloadItem: DownloadRecord,
	db: TransferDatabase,
	terminusId?: string,
	callBack?: (item: TransferItem) => void
) => {
	const taskId = getTaskId(downloadItem.id, terminusId);
	let downloadPath = downloadItem.path;
	if (downloadPath != undefined) {
		if (!downloadPath.startsWith('/')) {
			downloadPath = '/' + downloadPath;
		}

		if (!downloadPath.endsWith('/')) {
			downloadPath = downloadPath + '/';
		}

		downloadPath = '/Files/Home' + downloadPath;
	}
	if (taskId) {
		getCloudTaskById(taskId, db).then((task) => {
			if (task.length > 0) {
				if (callBack) {
					callBack(task[0]);
				}
			} else {
				if (downloadItem.status == 'remove') {
					return;
				}
				const transferItem: TransferItem = {
					task: -1,
					name: downloadItem.name,
					path: downloadPath,
					type: downloadItem.mimeType || downloadItem.file_type || '',
					isFolder: false,
					driveType: DriveType.Drive,
					front: TransferFront.cloud,
					status: TransferStatus.Prepare,
					url: downloadItem.url,
					startTime: downloadItem.startTime
						? downloadItem.startTime * 1000
						: downloadItem.created_time
						? new Date(downloadItem.created_time).getTime()
						: new Date().getTime(),
					endTime: 0,
					updateTime: downloadItem.startTime
						? downloadItem.startTime * 1000
						: downloadItem.created_time
						? new Date(downloadItem.created_time).getTime()
						: new Date().getTime(),
					from: downloadItem.url,
					to: downloadPath,
					isPaused: false,
					size: downloadItem.size || 0,
					uniqueIdentifier: taskId,
					userId: terminusId,
					currentPhase: 1,
					totalPhase: 1
				};
				db.transferData.add(transferItem).then((id: number) => {
					transferItem.id = id;
					if (callBack) {
						callBack(transferItem);
					}
				});
			}
		});
	} else {
		console.error('download task id error ' + taskId);
	}
};
