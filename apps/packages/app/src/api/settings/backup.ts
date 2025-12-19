import axios from 'axios';
import {
	BackupCreat,
	BackupPlan,
	BackupPlanDetail,
	BackupPolicy,
	OlaresSpaceRegion
} from '../../constant';
import { useBackupStore } from '../../stores/settings/backup';

export async function getOlaresSpaceRegion(): Promise<OlaresSpaceRegion[]> {
	const backupStore = useBackupStore();
	const result: any = await axios.get(
		backupStore.baseUrl + '/apis/backup/v1/regions'
	);
	return result || [];
}

export async function setPlanPassword(name: string, password: string) {
	const backupStore = useBackupStore();

	await axios.put(backupStore.baseUrl + '/api/backup/password/' + name, {
		password
	});
}

export async function getBackupPlan(id: string) {
	const backupStore = useBackupStore();
	return await axios.get(
		backupStore.baseUrl + '/apis/backup/v1/plans/backup/one/' + id
	);
}

export async function createBackup(create: BackupCreat) {
	const backupStore = useBackupStore();
	return await axios.post(
		backupStore.baseUrl + '/apis/backup/v1/plans/backup',
		{
			...create
		}
	);
}

export async function queryBackupList(
	offset: number,
	limit: number
): Promise<BackupPlan[]> {
	const backupStore = useBackupStore();
	try {
		const { backups }: any = await axios.get(
			backupStore.baseUrl +
				'/apis/backup/v1/plans/backup?' +
				'offset=' +
				offset +
				'&limit=' +
				limit
		);
		return backups || [];
	} catch (e) {
		console.log(e);
		return [];
	}
}

export async function getBackupDetails(
	id: string
): Promise<BackupPlanDetail | null> {
	const backupStore = useBackupStore();
	const result: any = await axios.get(
		backupStore.baseUrl + '/apis/backup/v1/plans/backup/' + id
	);
	return result || null;
}

export async function updateBackupPolicy(
	id: string,
	backupPolicy: BackupPolicy
) {
	const backupStore = useBackupStore();
	return await axios.put(
		backupStore.baseUrl + '/apis/backup/v1/plans/backup/' + id,
		{
			backupPolicy: { ...backupPolicy }
		}
	);
}

export async function deleteBackupPlan(id: string) {
	const backupStore = useBackupStore();
	await axios.delete(
		backupStore.baseUrl + '/apis/backup/v1/plans/backup/' + id
	);
}

export async function pauseBackupPlan(id: string) {
	const backupStore = useBackupStore();
	return await axios.post(
		backupStore.baseUrl + '/apis/backup/v1/plans/backup/' + id,
		{
			event: 'pause'
		}
	);
}

export async function resumeBackupPlan(id: string) {
	const backupStore = useBackupStore();
	return await axios.post(
		backupStore.baseUrl + '/apis/backup/v1/plans/backup/' + id,
		{
			event: 'resume'
		}
	);
}
