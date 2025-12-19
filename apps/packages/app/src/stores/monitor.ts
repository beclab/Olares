import { defineStore } from 'pinia';
import { useUserStore } from './user';
import { Moniter, Usage, Net } from '@bytetrade/core';
import { axiosInstanceProxy } from '../platform/httpProxy';
import { i18n } from '../boot/i18n';
import { NetworkErrorMode, busEmit } from 'src/utils/bus';

export enum TerminusStatus {
	UNKNOWN = 0,
	NORMAL = 1,
	ERROR = 2
}

export type MonitorStoreState = {
	usages: Usage[] | undefined;
	background: string[];
	dark_background: string[];
	net: Net | undefined;
	status: TerminusStatus;
	icons: string[];
};

export const useMonitorStore = defineStore('monitor', {
	state: () => {
		return {
			usages: undefined,
			net: undefined,
			status: TerminusStatus.UNKNOWN,
			background: [
				'linear-gradient(127.05deg, #FFFFFF 4.41%, #FFFCE0 84.8%)',
				'linear-gradient(121.71deg, #FFFFFF 4.54%, #E5FFEE 89.25%)',
				'linear-gradient(127.31deg, #FFFFFF 4.44%, #EBF5FF 85.27%)'
			],
			dark_background: [
				'linear-gradient(127.05deg, #1F1F1F 4.41%, #41381A 84.8%)',
				'linear-gradient(121.71deg, #1F1F1F 4.54%, #20331F 89.25%)',
				'linear-gradient(127.31deg, #1F1F1F 4.44%, #262E37 85.27%)'
			],
			icons: ['cpu.svg', 'memory.svg', 'disk.svg'],
			icons_dark: ['cpu_dark.svg', 'memory_dark.svg', 'disk_dark.svg']
		} as MonitorStoreState;
	},
	getters: {
		formatNetReceived(state) {
			if (!state.net) {
				return '-- B/s';
			}
			const bytes = state.net.received;
			if (bytes < 1024) {
				return bytes.toFixed(2) + 'B/s';
			}

			const kb = bytes / 1024;

			if (kb < 1024) {
				return kb.toFixed(2) + ' KB/s';
			}

			const m = kb / 1024;
			return m.toFixed(2) + ' M/s';
		},
		formatNetTransmitted(state) {
			if (!state.net) {
				return '-- B/s';
			}
			const bytes = state.net.transmitted;
			if (bytes < 1024) {
				return bytes.toFixed(2) + 'B/s';
			}

			const kb = bytes / 1024;

			if (kb < 1024) {
				return kb.toFixed(2) + ' KB/s';
			}

			const m = kb / 1024;
			return m.toFixed(2) + ' M/s';
		}
	},
	actions: {
		clear() {
			this.usages = undefined;
			this.net = undefined;
			this.status = TerminusStatus.UNKNOWN;
		},
		error() {
			this.usages = undefined;
			this.net = undefined;
			this.status = TerminusStatus.ERROR;
		},
		async loadMonitor() {
			const userStore = useUserStore();
			if (!userStore.current_user) {
				this.clear();
				return;
			}

			if (!userStore.current_user.setup_finished) {
				this.clear();
				return;
			}

			if (process.env.IS_PC_TEST) {
				const data: Moniter = {
					cpu: {
						name: i18n.global.t('monitor.cpu.title'),
						color: 'yellow',
						uint: '40',
						ratio: 40,
						total: 100,
						usage: 40
					},
					memory: {
						name: i18n.global.t('monitor.memory.title'),
						color: 'light-green-12',
						uint: '40',
						ratio: 40,
						total: 100,
						usage: 40
					},
					disk: {
						name: i18n.global.t('monitor.disk.title'),
						color: 'cyan-4',
						uint: '40',
						ratio: 40,
						total: 100,
						usage: 30
					},
					net: {
						received: 100000,
						transmitted: 200000
					}
				};
				this.usages = [];
				this.usages.push(data.cpu);
				this.usages.push(data.memory);
				this.usages.push(data.disk);
				this.net = data.net;
				this.status = TerminusStatus.NORMAL;
				return;
			}

			try {
				const response = await this.getMonitorInfo();
				if (!response) {
					return;
				}
				if (
					!response ||
					(response.status != 200 && response.status != 201) ||
					!response.data
				) {
					this.error();
					throw Error('Network error, please try again later');
				}

				if (response.data.code != 0) {
					this.error();
					if (response.data.code == '100001' || response.data.code == 100001) {
						busEmit('network_error', {
							type: NetworkErrorMode.monitor
						});
					}
				} else {
					const data = response.data.data;
					console.log('cpu', data);
					this.usages = [];
					this.usages.push(data.cpu);
					this.usages.push(data.memory);
					this.usages.push(data.disk);

					this.usages[0].name = i18n.global.t('monitor.cpu.title');
					this.usages[0].color = 'yellow';
					this.usages[0].uint = '40';

					this.usages[1].name = i18n.global.t('monitor.memory.title');
					this.usages[1].color = 'light-green-12';
					this.usages[1].uint = '40';

					this.usages[2].name = i18n.global.t('monitor.disk.title');
					this.usages[2].color = 'cyan-4';
					this.usages[2].uint = '40';

					this.net = data.net;
					this.status = TerminusStatus.NORMAL;
				}
			} catch (error) {
				busEmit('network_error', {
					type: NetworkErrorMode.monitor
				});
			}
		},
		async getMonitorInfo() {
			const userStore = useUserStore();
			if (userStore.current_user?.isLargeVersion12_3) {
				const baseURL = userStore.getSelectUserModuleServer('settings');
				if (baseURL) {
					const instance = axiosInstanceProxy({
						baseURL: baseURL,
						timeout: 1000 * 10,
						headers: {
							'Content-Type': 'application/json'
						}
					});
					return await instance.get('/api/users/monitor/cluster');
				}
			}

			const baseURL = userStore.current_user?.auth_url;
			if (baseURL) {
				const instance = axiosInstanceProxy({
					baseURL: baseURL,
					timeout: 1000 * 10,
					headers: {
						'Content-Type': 'application/json',
						'X-Authorization': userStore.current_user.access_token
					}
				});
				return await instance.get('/bfl/monitor/v1alpha1/cluster', {});
			}
		}
	}
});
