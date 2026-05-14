import { defineStore } from 'pinia';
import axios from 'axios';
import { useTokenStore } from './token';
import { UpgradeState } from '@apps/desktop/type/types';
import { compareOlaresVersion } from '@bytetrade/core';

export type UpgradeStateStore = {
	state: UpgradeState;
	upgradeDetail: {
		upgradingState: string;
		terminusState: string;
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
				state.upgradeDetail && state.upgradeDetail.terminusState == 'upgrading'
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
					this.stopRefreshUpgradeState();
					if (this.state == UpgradeState.StatusRunning) {
						this.state = UpgradeState.StatusComplete;
					} else {
						this.state = UpgradeState.NotRunning;
					}
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
