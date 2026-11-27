import { defineStore } from 'pinia';
import axios from 'axios';
import { useTokenStore } from './token';
import { UpgradeState } from '@apps/desktop/type/types';
import { compareOlaresVersion } from '@bytetrade/core';
import { TerminusStatusEnum } from 'src/services/abstractions/mdns/service';

export type UpgradeStateStore = {
	state: UpgradeState;
	upgradeDetail: {
		upgradingState: string;
		terminusState: TerminusStatusEnum;
	};
	refreshTimer: any;
};

export const useUpgradeStore = defineStore('upgrade', {
	state: () => {
		return {
			state: UpgradeState.NotRunning,
			refreshTimer: undefined
		} as UpgradeStateStore;
	},
	getters: {
		isUpgrading(state): boolean {
			return (
				state.upgradeDetail &&
				state.upgradeDetail.terminusState == TerminusStatusEnum.Upgrading
			);
		},
		isUpgradingComplete(state): boolean {
			return (
				state.upgradeDetail &&
				state.upgradeDetail.terminusState == TerminusStatusEnum.TerminusRunning
			);
		},
		isUpgradingError(state): boolean {
			return (
				state.upgradeDetail &&
				state.upgradeDetail.terminusState == TerminusStatusEnum.SystemError
			);
		}
	},
	actions: {
		async update_upgrade_state_info() {
			const tokenStore = useTokenStore();
			if (
				!tokenStore.terminus.osVersion ||
				compareOlaresVersion(tokenStore.terminus.osVersion, '1.12.0-0')
					.compare < 0
			) {
				let modeStr = '';
				if (localStorage.getItem('dev_mode')) {
					modeStr = '?dev_mode=true';
				}
				try {
					const data: any = await axios.get(
						tokenStore.url + '/server/upgrade/state' + modeStr,
						{}
					);
					this.state = data.state;
				} catch (error) {
					console.log('update_upgrade_state_info error', error);
				}
			} else {
				this.requestUpgradeStatus();
			}
		},
		async requestUpgradeStatus() {
			try {
				const tokenStore = useTokenStore();
				const res: any = await axios.get(tokenStore.url + '/api/system/status');
				this.upgradeDetail = res;
				if (this.isUpgrading) {
					this.setStateByUpgradingState(this.upgradeDetail.upgradingState);
				} else {
					const isFailed =
						this.isUpgradingError ||
						this.upgradeDetail.upgradingState == 'failed';

					if (
						this.state == UpgradeState.StatusRunning &&
						!isFailed &&
						!this.isUpgradingComplete
					) {
						return;
					}

					if (this.state == UpgradeState.StatusRunning) {
						this.state = isFailed
							? UpgradeState.StatusFailed
							: UpgradeState.StatusComplete;
					} else {
						this.state = UpgradeState.NotRunning;
					}
					this.stopRefreshUpgradeState();
				}
			} catch (error) {
				/* empty */
			}
		},
		setStateByUpgradingState(upgradingState: string) {
			switch (upgradingState) {
				case 'failed':
					this.state = UpgradeState.StatusFailed;
					break;
				default:
					this.state = UpgradeState.StatusRunning;
					this.startRefreshActivedMachine();
					break;
			}
		},

		startRefreshActivedMachine(ms = 5000) {
			if (this.refreshTimer) {
				return;
			}
			this.refreshTimer = setInterval(async () => {
				this.update_upgrade_state_info();
			}, ms);
		},
		stopRefreshUpgradeState() {
			clearInterval(this.refreshTimer);
			this.refreshTimer = undefined;
		}
	}
});
