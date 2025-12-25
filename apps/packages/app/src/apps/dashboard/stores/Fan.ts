import { IP_METHOD_OPTION, SystemIFSItem } from './../types/network';
import { defineStore } from 'pinia';
import {
	getGraphicsList,
	getSystemFan,
	getSystemStatus
} from '@apps/dashboard/src/network';
import { Locker } from '../types/main';
import { GraphicsListParams } from '../types/gpu';
import { round } from 'lodash';
import { TerminusStatus } from 'src/services/abstractions/mdns/service';

interface FanStore {
	data: {
		gpu_fan_speed: number;
		gpu_temperature: number;
		cpu_fan_speed: number;
		cpu_temperature: number;
		gpu_power: number;
		gpu_power_limit: number;
	};
	systemStatus: TerminusStatus | undefined;
	loading: boolean;
	locker: Locker;
}
export const useFanStore = defineStore('FanStore', {
	state: (): FanStore => ({
		data: {
			gpu_fan_speed: 0,
			gpu_temperature: 0,
			cpu_fan_speed: 0,
			cpu_temperature: 0,
			gpu_power: 0,
			gpu_power_limit: 0
		},
		systemStatus: undefined,
		loading: false,
		locker: undefined
	}),

	getters: {
		gpu_fan_speed_ratio: (state) => state.data.gpu_fan_speed,
		isOlaresOneDevice: (state) =>
			state.systemStatus?.device_name === 'Olares One'
	},
	actions: {
		init() {
			this.getFanData(false);
			this.getFanData(true);
			this.fetchSystemStatus();
		},
		async fetchSystemStatus() {
			const res = await getSystemStatus();
			this.systemStatus = res.data.data;
		},
		async getFanData(autofresh = false) {
			if (!autofresh) {
				this.loading = true;
			}
			try {
				const params: GraphicsListParams = {
					filters: {},
					pageRequest: {
						sort: 'ASC',
						sortField: 'id'
					}
				};

				const gpusRes = await getGraphicsList(params);
				const gpuData = gpusRes.data.list[0];

				const res = await getSystemFan();
				this.data = {
					...res.data.data,
					gpu_power: round(gpuData.power, 2),
					gpu_power_limit: gpuData.powerLimit
				};
				this.refresh();
			} catch (error) {
				this.loading = false;
			}
			this.loading = false;
		},
		refresh() {
			this.clearLocker();
			this.locker = setTimeout(() => {
				this.getFanData(true);
			}, 2500);
		},
		clearLocker() {
			this.locker && clearTimeout(this.locker);
		}
	}
});
