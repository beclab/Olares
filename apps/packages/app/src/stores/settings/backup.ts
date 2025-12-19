import { defineStore } from 'pinia';
import {
	ApplicationSelectorState,
	BackupMessage,
	BackupPlan,
	BackupPolicy,
	BackupResourcesType,
	BackupSnapshotDetail,
	createLocationConfig,
	createPolicy,
	OlaresSpaceRegion,
	olaresSpaceUrl,
	RestoreMessage,
	RestorePlan,
	RestoreSnapshotInfo,
	SupportBackupAppList
} from 'src/constant';
import {
	createBackup,
	deleteBackupPlan,
	getBackupDetails,
	getBackupPlan,
	pauseBackupPlan,
	queryBackupList,
	resumeBackupPlan,
	setPlanPassword,
	updateBackupPolicy
} from 'src/api/settings/backup';
import {
	cancelRestorePlan,
	createRestoreBySnapshot,
	createRestoreByUrl,
	getRestoreDetails,
	getRestorePlan,
	getSnapshotsByUrl,
	queryRestoreList
} from 'src/api/settings/restore';
import {
	cancelSnapshot,
	createSnapshot,
	getSnapshotDetails,
	querySnapshotList
} from 'src/api/settings/snapshot';
import { binaryInsert, CompareBackup } from 'src/utils/rss-utils';
import logger from 'electron-log';
import { stringToBase64 } from '@didvault/sdk/src/core';
import axios from 'axios';
import { useIntegrationStore } from './integration';
import { useApplicationStore } from 'src/stores/settings/application';
import { APP_STATUS } from 'src/constant/constants';

export type BackupState = {
	inited: boolean;
	baseUrl: string;
	backupList: BackupPlan[];
	restoreList: RestorePlan[];
	loadedBackupIds: Set<string>;
	loadedRestoreIds: Set<string>;
	usage: number;
	total: number;
};

export const useBackupStore = defineStore('backup', {
	state: () => {
		return {
			inited: false,
			baseUrl: '',
			backupList: [],
			restoreList: [],
			loadedBackupIds: new Set(),
			loadedRestoreIds: new Set(),
			usage: 0,
			total: 0
		} as BackupState;
	},
	getters: {
		restoreOptions(state: BackupState) {
			return state.backupList.map((item) => {
				return {
					label: item.name,
					value: item.id,
					enable: true
				};
			});
		}
	},
	actions: {
		async init() {
			const limit = 50;
			this.baseUrl = window.location.origin;
			const [backupList, restoreList] = await Promise.all([
				queryBackupList(0, limit),
				queryRestoreList(0, limit)
			]);
			this.backupList = backupList;
			this.restoreList = restoreList;

			this.loadedBackupIds = new Set(backupList.map((item) => item.id));
			this.loadedRestoreIds = new Set(restoreList.map((item) => item.id));

			await this.loadData(
				this.backupList,
				this.loadedBackupIds,
				queryBackupList,
				limit
			);
			await this.loadData(
				this.restoreList,
				this.loadedRestoreIds,
				queryRestoreList,
				limit
			);

			this.inited = true;
		},
		async loadData(
			list: any[],
			loadedIds: Set<string>,
			queryFunction: any,
			limit: number,
			offset = 0
		) {
			try {
				const batch = await queryFunction(offset, limit);
				this.insertNewData(list, loadedIds, batch);

				if (batch.length === limit) {
					await this.loadData(
						list,
						loadedIds,
						queryFunction,
						limit,
						offset + limit
					);
				}
			} catch (error) {
				console.error('Error loading data:', error);
			}
		},
		insertNewData(list: any[], loadedIds: Set<string>, newItems: any[]) {
			const uniqueItems = newItems.filter((item) => !loadedIds.has(item.id));

			if (uniqueItems.length > 0) {
				uniqueItems.forEach((item) => {
					loadedIds.add(item.id);
					binaryInsert(list, item, CompareBackup);
				});
			}
		},
		getModuleSever(module: string, protocol = 'https:', suffix = '') {
			let replaceContent = process.env.DEV ? 'test' : 'settings';
			if (!module) {
				replaceContent = replaceContent + '.';
			}
			const url =
				protocol +
				'//' +
				window.location.hostname.replace(replaceContent, module) +
				suffix;
			console.log(url);
			return url;
		},
		async createBackupPlan(
			createType: BackupResourcesType,
			name: string,
			password: string,
			locationData: any,
			policy: BackupPolicy,
			backupApp?: ApplicationSelectorState,
			path?: string,
			region?: OlaresSpaceRegion
		) {
			await setPlanPassword(name, password);
			let createProps;
			if (createType === BackupResourcesType.app) {
				createProps = {
					name,
					backupType: {
						type: createType,
						name: backupApp?.value
					},
					location: locationData.type,
					locationConfig: createLocationConfig(locationData, region),
					backupPolicy: createPolicy(policy)
				};
			} else {
				createProps = {
					name,
					path,
					location: locationData.type,
					locationConfig: createLocationConfig(locationData, region),
					backupPolicy: createPolicy(policy)
				};
			}
			const response: any = await createBackup(createProps);
			console.log(response);
			if (response && response.id) {
				this.insertNewData(this.backupList, this.loadedBackupIds, [response]);
			}
		},
		async updateBackupPlan(id: string, policy: BackupPolicy) {
			const result: any = await updateBackupPolicy(id, createPolicy(policy));
			this.updateOneBackupPlan(id, result);
		},
		updateOneBackupPlan(id: string, value: any) {
			for (let i = 0; i < this.backupList.length; ++i) {
				if (this.backupList[i].id === id) {
					this.backupList[i] = {
						...this.backupList[i],
						...value
					};
				}
			}
		},
		updateBackupBySocket(data: BackupMessage) {
			console.log(data);
			if (!this.inited) {
				console.warn(
					'Backup message update attempted before initialization complete.'
				);
				return;
			}

			if (data) {
				if (this.loadedBackupIds.has(data.backupId)) {
					const backup = this.backupList.find(
						(item) => item.id === data.backupId
					);
					if (backup) {
						backup.status = data.status;
						backup.progress = data.progress;
						if (data.restoreSize) {
							backup.restoreSize = data.restoreSize;
						}
						if (data.size) {
							backup.size = data.size;
						}
					}
				} else {
					getBackupPlan(data.backupId)
						.then((response: any) => {
							if (response && response.id) {
								this.insertNewData(this.backupList, this.loadedBackupIds, [
									response
								]);
								this.loadedBackupIds.add(data.backupId);
							}
						})
						.catch((e) => {
							logger.error(e);
						});
				}
			}
		},

		updateRestoreBySocket(data: RestoreMessage) {
			console.log(data);
			if (!this.inited) {
				console.warn(
					'Restore message update attempted before initialization complete.'
				);
				return;
			}

			if (data) {
				if (this.loadedRestoreIds.has(data.id)) {
					const restore = this.restoreList.find((item) => item.id === data.id);
					if (restore) {
						restore.status = data.status;
						restore.progress = data.progress;
						if (data.endat && data.endat !== 0) {
							restore.endAt = data.endat;
						}
					}
				} else {
					getRestorePlan(data.id)
						.then((response: any) => {
							if (response && response.id) {
								this.insertNewData(this.restoreList, this.loadedRestoreIds, [
									response
								]);
								this.loadedRestoreIds.add(data.id);
							}
						})
						.catch((e) => {
							logger.error(e);
						});
				}
			}
		},
		async deleteBackupPlan(id: string) {
			await deleteBackupPlan(id);
			const index = this.backupList.findIndex((item) => item.id === id);
			if (index > -1) {
				this.backupList.splice(index, 1);
				this.loadedBackupIds.delete(id);
			}
		},
		async resumeBackup(backupId: string) {
			await resumeBackupPlan(backupId);
			return await getBackupDetails(backupId);
		},
		async pauseBackup(backupId: string) {
			await pauseBackupPlan(backupId);
			return await getBackupDetails(backupId);
		},
		async getBackupDetails(id: string) {
			return await getBackupDetails(id);
		},
		async getSnapshots(id: string, offset: number, limit: number) {
			return await querySnapshotList(id, offset, limit);
		},
		async createBackupSnapShot(id: string) {
			return await createSnapshot(id);
		},
		async getSnapShotDetail(
			backupId: string,
			snapShotId: string
		): Promise<BackupSnapshotDetail> {
			return await getSnapshotDetails(backupId, snapShotId);
		},
		async cancelBackupSnapShot(backupId: string, snapShotId: string) {
			await cancelSnapshot(backupId, snapShotId);
			await this.getSnapShotDetail(backupId, snapShotId);
		},
		async restoreBackup(snapshotId: string, path: string) {
			const response: any = await createRestoreBySnapshot(snapshotId, path);
			console.log(response);
			if (response && response.id) {
				this.insertNewData(this.restoreList, this.loadedRestoreIds, [response]);
			}
		},
		async getRestoreDetails(restoreId: string) {
			return await getRestoreDetails(restoreId);
		},
		async cancelRestore(restoreId: string) {
			await cancelRestorePlan(restoreId);
			await this.getRestoreDetails(restoreId);
		},
		async restoreCustomUrl(
			backupUrl: string,
			password: string,
			path: string,
			dir: string,
			encode = true,
			snapshot?: RestoreSnapshotInfo
		): Promise<void> {
			let resultUrl = backupUrl;
			if (snapshot) {
				resultUrl =
					backupUrl +
					'?snapshotId=' +
					snapshot?.id +
					'&snapshotTime=' +
					snapshot?.createAt +
					'&backupPath=' +
					stringToBase64(snapshot?.backupPath, false);
			}
			const response: any = await createRestoreByUrl(
				encode ? stringToBase64(resultUrl, false) : resultUrl,
				password,
				path,
				dir
			);
			console.log(response);
			if (response && response.id) {
				this.insertNewData(this.restoreList, this.loadedRestoreIds, [response]);
			}
		},
		async parseUrl(url: string, pwd: string, offset: number, limit: number) {
			return await getSnapshotsByUrl(url, pwd, offset, limit);
		},
		getSupportApplicationOptions() {
			const appStore = useApplicationStore();
			return appStore.applications
				.filter((item) => {
					return (
						item.state === APP_STATUS.RUNNING &&
						SupportBackupAppList.includes(item.name)
					);
				})
				.map((item) => {
					return {
						label: item.name,
						value: item.name,
						disable: false,
						hideLabel: false,
						app: item
					};
				});
		}
	}
});
