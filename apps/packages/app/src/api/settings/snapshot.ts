import axios from 'axios';
import { BackupSnapshotDetail } from '../../constant';
import { useBackupStore } from '../../stores/settings/backup';

export async function querySnapshotList(
	backupId: string,
	offset: number,
	limit: number
) {
	const backupStore = useBackupStore();
	return await axios.get(
		backupStore.baseUrl +
			'/apis/backup/v1/plans/backup/' +
			backupId +
			'/snapshots?limit=' +
			limit +
			'&offset=' +
			offset
	);
}

export async function createSnapshot(backupId: string) {
	const backupStore = useBackupStore();
	await axios.post(
		backupStore.baseUrl +
			'/apis/backup/v1/plans/backup/' +
			backupId +
			'/snapshots',
		{
			event: 'create'
		}
	);
}

export async function getBackupSnapshot(backupId: string, snapshotId: string) {
	const backupStore = useBackupStore();
	return await axios.get(
		backupStore.baseUrl +
			'/apis/backup/v1/plans/backup/' +
			backupId +
			'/snapshots/one/' +
			snapshotId
	);
}

export async function getSnapshotDetails(
	backupId: string,
	snapshotId: string
): Promise<BackupSnapshotDetail> {
	const backupStore = useBackupStore();
	return await axios.get(
		backupStore.baseUrl +
			'/apis/backup/v1/plans/backup/' +
			backupId +
			'/snapshots/' +
			snapshotId
	);
}

export async function cancelSnapshot(backupId: string, snapshotId: string) {
	const backupStore = useBackupStore();
	await axios.delete(
		backupStore.baseUrl +
			'/apis/backup/v1/plans/backup/' +
			backupId +
			'/snapshots/' +
			snapshotId,
		{
			data: {
				event: 'cancel'
			}
		}
	);
}
