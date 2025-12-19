import { defineStore } from 'pinia';
import { Usage } from '@bytetrade/core';
import { get_cluster_resource } from 'src/types/resource';
import { getSuitableUnit, getValueByUnit } from 'src/utils/settings/monitoring';

export type MonitorStoreState = {
	usages: Usage[];
};

export const useMonitorStore = defineStore('monitor', {
	state: () => {
		return {
			usages: []
		} as MonitorStoreState;
	},
	getters: {},
	actions: {
		async loadMonitor() {
			const userUsage = await get_cluster_resource();
			this.usages = [];

			const resourceConfigs: Array<{
				name: 'cpu' | 'disk' | 'memory';
				totalField: 'user_cpu_total' | 'user_disk_total' | 'user_memory_total';
				usageField: 'user_cpu_usage' | 'user_disk_usage' | 'user_memory_usage';
				color: string;
			}> = [
				{
					name: 'cpu',
					totalField: 'user_cpu_total',
					usageField: 'user_cpu_usage',
					color: 'yellow-12'
				},
				{
					name: 'disk',
					totalField: 'user_disk_total',
					usageField: 'user_disk_usage',
					color: 'light-blue-13'
				},
				{
					name: 'memory',
					totalField: 'user_memory_total',
					usageField: 'user_memory_usage',
					color: 'light-green-13'
				}
			];

			const processResource = (config: (typeof resourceConfigs)[0]) => {
				const totalValue = userUsage[config.totalField];
				const usageValue = userUsage[config.usageField];

				const unit: string = getSuitableUnit(
					totalValue || usageValue,
					config.name
				);

				const total = getValueByUnit(`${totalValue}`, unit);
				const usage = getValueByUnit(`${usageValue}`, unit);

				let percent = 0;
				if (totalValue && usageValue) {
					percent = Number(
						((Number(usageValue) / Number(totalValue)) * 100).toFixed(0)
					);
				}

				this.usages.push({
					total,
					usage,
					ratio: percent,
					uint: unit,
					name: config.name,
					color: config.color
				});
			};

			resourceConfigs.forEach(processResource);
		}
	}
});
