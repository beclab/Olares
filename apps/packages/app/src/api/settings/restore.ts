import { RestorePlan, RestorePlanDetail } from '../../constant';
import axios from 'axios';
import { useBackupStore } from '../../stores/settings/backup';

export async function createRestoreBySnapshot(
	snapshotId: string,
	path: string
) {
	const backupStore = useBackupStore();
	return await axios.post(
		backupStore.baseUrl + '/apis/backup/v1/plans/restore',
		{
			snapshotId,
			path
		}
	);
}

export async function getSnapshotsByUrl(
	backupUrl: string,
	password: string,
	offset: number,
	limit: number
) {
	const backupStore = useBackupStore();
	return await axios.post(
		backupStore.baseUrl + '/apis/backup/v1/plans/restore/checkurl',
		{
			backupUrl,
			password,
			offset,
			limit
		}
	);
}

export async function createRestoreByUrl(
	backupUrl: string,
	password: string,
	path: string,
	dir: string
) {
	const backupStore = useBackupStore();
	return await axios.post(
		backupStore.baseUrl + '/apis/backup/v1/plans/restore',
		{
			backupUrl,
			password,
			path,
			dir
		}
	);
}

export async function getRestorePlan(id: string) {
	const backupStore = useBackupStore();
	return await axios.get(
		backupStore.baseUrl + '/apis/backup/v1/plans/restore/one/' + id
	);
}

export async function queryRestoreList(
	offset: number,
	limit: number
): Promise<RestorePlan[]> {
	try {
		const backupStore = useBackupStore();
		const { restores }: any = await axios.get(
			backupStore.baseUrl +
				'/apis/backup/v1/plans/restore?offset=' +
				offset +
				'&limit=' +
				limit
		);
		return restores || [];
	} catch (e) {
		console.log(e);
		return [];
	}
}

export async function getRestoreDetails(
	id: string
): Promise<RestorePlanDetail | null> {
	const backupStore = useBackupStore();
	const result: any = await axios.get(
		backupStore.baseUrl + '/apis/backup/v1/plans/restore/' + id
	);
	return result || null;
}

export async function cancelRestorePlan(id: string) {
	const backupStore = useBackupStore();
	await axios.delete(
		backupStore.baseUrl + '/apis/backup/v1/plans/restore/' + id,
		{
			data: {
				event: 'cancel'
			}
		}
	);
}
