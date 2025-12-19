import axios from 'axios';
import { defineStore } from 'pinia';
import { useTokenStore } from './token';
import { UpgradeStatus, VersionInfo } from '../../constant';
import { compareOlaresVersion } from '@bytetrade/core';

interface upgradeState {
	versionInfo: VersionInfo | undefined;
	upgradeState: string;
	upgradeDetail: {
		upgradingTarget: string;
		upgradingRetryNum: number;
		upgradingState: string;
		upgradingStep: string;
		upgradingProgress: string;
		upgradingError: string;
		terminusState: string;
	};
}

export const useUpgradeStore = defineStore('upgrade', {
	state: () => {
		return {
			versionInfo: undefined,
			upgradeState: ''
		} as upgradeState;
	},
	getters: {
		upgradeCompleted(state): boolean {
			return (
				state.upgradeDetail && state.upgradeDetail.terminusState != 'upgrading'
			);
		}
	},
	actions: {
		async upgrade(version?: string) {
			this.upgradeState = UpgradeStatus.Running;
			const tokenStore = useTokenStore();
			if (!version || compareOlaresVersion(version, '1.12.0-0').compare < 0) {
				return await axios.get(tokenStore.url + '/api/upgrade', {});
			}
			return await axios.post(tokenStore.url + '/api/command/upgrade', {
				version
			});
		},

		async cancelUpgrade(version?: string) {
			const tokenStore = useTokenStore();
			if (!version || compareOlaresVersion(version, '1.12.0-0').compare < 0) {
				return;
			}
			await axios.delete(tokenStore.url + '/api/command/upgrade');
			this.queryUpgradeState(version);
		},

		async checkLastOsVersion() {
			try {
				const tokenStore = useTokenStore();
				this.versionInfo = await axios.get(
					tokenStore.url + '/api/checkLastOsVersion'
				);
			} catch (error) {
				console.log(error);
			}
		},

		async queryUpgradeState(version?: string) {
			const tokenStore = useTokenStore();
			try {
				if (!version || compareOlaresVersion(version, '1.12.0-0').compare < 0) {
					const res: any = await axios.get(
						tokenStore.url + '/api/upgrade/state'
					);
					this.upgradeState = res.state;
				} else {
					const res: any = await axios.get(
						tokenStore.url + '/api/system/status'
					);
					this.upgradeDetail = res;
					if (
						this.upgradeDetail.upgradingState &&
						this.upgradeDetail.upgradingState.length > 0
					) {
						this.upgradeState = this.upgradeDetail.upgradingState;
					} else {
						if (
							this.upgradeState != UpgradeStatus.Completed &&
							this.upgradeCompleted
						) {
							this.upgradeState = UpgradeStatus.Completed;
						}
					}
				}
			} catch (error) {
				console.log('upgradeState error', error);
			}
		}
	}
});
