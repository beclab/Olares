import { IP_METHOD_OPTION, SystemIFSItem } from './../types/network';
import { defineStore } from 'pinia';
import {
	getGraphicsList,
	getNodeMonitoring,
	getSystemFan,
	getSystemStatus
} from '@apps/dashboard/src/network';
import { Locker } from '../types/main';
import { GraphicsListParams } from '../types/gpu';
import { get, round } from 'lodash';
import { TerminusStatus } from 'src/services/abstractions/mdns/service';
import {
	getLastMonitoringData,
	getResult
} from '@apps/dashboard/src/utils/monitoring';
import { PodItem } from '@apps/control-panel-common/network/network';

const MetricTypes = {
	rapl_package_joules_power: 'node_rapl_package_joules_power',
	rapl_constraint_0_max_power_uw: 'node_rapl_constraint_0_max_power_uw',
	rapl_constraint_0_power_limit_uw: 'node_rapl_constraint_0_power_limit_uw'
};
export const cpuStopColorRatio = 0.44;
const getValue = (data: PodItem) => get(data, 'value[1]', 0);
const convertToWatt = (value: number) => round(value / 1000 ** 2, 2);
interface FanStore {
	data: {
		gpu_fan_speed: number;
		gpu_temperature: number;
		cpu_fan_speed: number;
		cpu_temperature: number;
		gpu_power: number;
		gpu_power_limit: number;
	};
	cpuData: {
		cpu_power: number;
		cpu_power_max: number;
		cpu_power_limit: number;
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
			gpu_power_limit: 1
		},
		cpuData: {
			cpu_power: 0,
			cpu_power_max: 0,
			cpu_power_limit: 1
		},
		systemStatus: undefined,
		loading: false,
		locker: undefined
	}),

	getters: {
		gpu_fan_speed_ratio: (state) => state.data.gpu_fan_speed,
		isOlaresOneDevice: (state) =>
			state.systemStatus?.device_name === 'Olares One',
		cpuwarningRatio: (state) =>
			state.cpuData.cpu_power_limit
				? round(state.cpuData.cpu_power_max / state.cpuData.cpu_power_limit, 2)
				: cpuStopColorRatio
	},
	actions: {
		init() {
			this.getFanData(false);
			this.getFanData(true);
			this.getCpuPower();
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

		async getCpuPower() {
			const params = {
				metrics_filter: Object.values(MetricTypes).join('|')
			};
			const res = await getNodeMonitoring(params);
			const data = getResult(res.data.results);

			const metrics = getLastMonitoringData(data);
			const cpu_power = getValue(
				metrics[MetricTypes.rapl_package_joules_power]
			);
			const rapl_constraint_0_max_power_uw = getValue(
				metrics[MetricTypes.rapl_constraint_0_max_power_uw]
			);
			const rapl_constraint_0_power_limit_uw = getValue(
				metrics[MetricTypes.rapl_constraint_0_power_limit_uw]
			);
			this.cpuData = {
				cpu_power: round(cpu_power, 2),
				cpu_power_max: convertToWatt(rapl_constraint_0_max_power_uw),
				cpu_power_limit: convertToWatt(rapl_constraint_0_power_limit_uw)
			};
		},
		refresh() {
			this.clearLocker();
			this.locker = setTimeout(() => {
				this.getFanData(true);
				this.getCpuPower();
			}, 2500);
		},
		clearLocker() {
			this.locker && clearTimeout(this.locker);
		}
	}
});
