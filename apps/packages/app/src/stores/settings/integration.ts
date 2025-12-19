import { IntegrationAccount } from 'src/services/abstractions/integration/integrationService';
import integrationService from 'src/services/integration/index';
import { AccountType, IntegrationAccountMiniData } from '@bytetrade/core';
import { useTokenStore } from './token';
import { defineStore } from 'pinia';
import axios from 'axios';
import { getRequireImage } from '../../utils/settings/helper';

import { olaresSpaceUrl } from 'src/constant';
import { i18n } from 'src/boot/i18n';

export enum PlanLevel {
	None = 0,
	Starter = 1,
	Basic = 2,
	Pro = 3
}

export enum PlanType {
	Starter = 'Starter',
	Basic = 'Basic',
	Pro = 'Pro'
}

export interface BackupPlan {
	totalSize: number;
	usageSize: number;
	canBackup: boolean;
	planLevel: PlanLevel;
}

export interface trafficPlan {
	totalSize: number;
	usedSize: number;
	overageSize: number;
}

export type IntegrationState = {
	accounts: IntegrationAccountMiniData[];
	accountLoading: boolean;
	backupPlan?: BackupPlan;
	trafficPlan?: trafficPlan;
};

export const useIntegrationStore = defineStore('settingsIntegration', {
	state: () => ({ accounts: [], accountLoading: false } as IntegrationState),

	getters: {
		backupAccounts: (state: IntegrationState) => {
			return state.accounts.filter(
				(item) =>
					item.type === AccountType.AWSS3 ||
					item.type === AccountType.Tencent ||
					item.type === AccountType.Space
			);
		},
		spaceAccount: (state: IntegrationState) => {
			return state.accounts.filter((item) => item.type === AccountType.Space);
		},
		backupUsageSize: (state: IntegrationState) => {
			if (!state.backupPlan) {
				return 0;
			}
			return state.backupPlan.usageSize;
		},
		backupTotalSize: (state: IntegrationState) => {
			if (!state.backupPlan) {
				return 0;
			}
			return state.backupPlan.totalSize;
		},
		trafficUsageSize: (state: IntegrationState) => {
			if (!state.trafficPlan) {
				return 0;
			}
			return state.trafficPlan.usedSize;
		},
		trafficPlanTotalSize: (state: IntegrationState) => {
			if (!state.trafficPlan) {
				return 0;
			}
			return state.trafficPlan.totalSize;
		},
		planLevel: (state: IntegrationState) => {
			let text = i18n.global.t('PlanLevel.none');
			let color = '';
			if (state.backupPlan) {
				switch (state.backupPlan.planLevel) {
					case PlanLevel.Starter:
						text = i18n.global.t('PlanLevel.starter');
						color = 'warning';
						break;
					case PlanLevel.Basic:
						text = i18n.global.t('PlanLevel.basic');
						color = 'warning';
						break;
					case PlanLevel.Pro:
						text = i18n.global.t('PlanLevel.pro');
						color = 'warning';
						break;
				}
			}
			return {
				color,
				text
			};
		}
	},

	actions: {
		async getAccount(
			type: AccountType | 'all'
		): Promise<IntegrationAccountMiniData[]> {
			try {
				const tokenStore = useTokenStore();
				if (this.accounts.length == 0) {
					this.accountLoading = true;
				}
				const result: any = await axios.get(
					`${tokenStore.url}/api/account/` + type
				);
				this.accounts = result;
			} catch (error) {
				/* empty */
			} finally {
				this.accountLoading = false;
			}
			return this.accounts;
		},
		async getAccountByTypeAndName(
			type: AccountType,
			name: string
		): Promise<IntegrationAccountMiniData[]> {
			const tokenStore = useTokenStore();
			const result: any = await axios.get(
				`${tokenStore.url}/api/account/` + type + '/' + name
			);
			return result;
		},
		async createAccount(data: IntegrationAccount) {
			const tokenStore = useTokenStore();
			return await axios.post(`${tokenStore.url}/api/account/create`, data);
		},
		async deleteAccount(data: IntegrationAccountMiniData) {
			const tokenStore = useTokenStore();
			const key = this.getStoreKey(data);
			return await axios.delete(`${tokenStore.url}/api/account/${key}`);
		},
		getStoreKey(data: IntegrationAccountMiniData | IntegrationAccount) {
			if (data.name) {
				return 'integration-account:' + data.type + ':' + data.name;
			} else {
				return 'integration-account:' + data.type;
			}
		},
		getAccountByType(data: IntegrationAccountMiniData | IntegrationAccount) {
			return integrationService.getAccountByType(data.type);
		},
		getAccountIcon(data: IntegrationAccountMiniData) {
			const account = this.getAccountByType(data);
			if (!account) {
				return '';
			}
			return getRequireImage(`integration/${account.detail.icon}`);
		},
		async getAccountFullData(
			data: IntegrationAccountMiniData | IntegrationAccount
		) {
			const tokenStore = useTokenStore();
			const key = this.getStoreKey(data);
			return await axios.post(`${tokenStore.url}/api/account/retrieve`, {
				name: key
			});
		},
		async getUsage() {
			const spaceList: any = this.spaceAccount;
			console.log(spaceList);
			if (spaceList.length > 0) {
				const fullData: any = await this.getAccountFullData(spaceList[0]);
				console.log(fullData);
				try {
					const response: any = await axios.post(
						olaresSpaceUrl + '/api/v1/resource/backup/usage',
						{
							userid: fullData.raw_data.userid,
							token: fullData.raw_data.access_token
						}
					);
					console.log(response.data);
					this.backupPlan = response.data;

					const response2: any = await axios.post(
						olaresSpaceUrl + '/v1/resource/traffic/usage',
						{
							userid: fullData.raw_data.userid,
							token: fullData.raw_data.access_token
						}
					);
					console.log(response2.data);
					this.trafficPlan = response2.data;
				} catch (error) {
					console.log(error);
				}
			}

			return { backup: this.backupPlan, traffic: this.trafficPlan };
		}
	}
});
