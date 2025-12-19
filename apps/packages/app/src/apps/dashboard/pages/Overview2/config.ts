import { PodItem, ResourcesResponse } from '@apps/dashboard/src/types/network';
import { get, isArray, isEmpty } from 'lodash';
import { getLastMonitoringData } from '@apps/dashboard/src/utils/monitoring';
import { _capitalize } from '@apps/dashboard/src/utils';
import { firstToUpperWith_ } from '@apps/dashboard/src/constant';
import memory_icon from '../../assets/memory.svg';
import memory_icon_dark from '../../assets/memory-dark.svg';
import memory_alt_icon from '../../assets/memory_alt.svg';
import memory_alt_icon_dark from '../../assets/memory_alt-dark.svg';
import hard_drive_icon from '../../assets/hard_drive.svg';
import hard_drive_icon_dark from '../../assets/hard_drive-dark.svg';
import package_2_active_icon from '../../assets/package_2_active.svg';
import package_2_icon from '../../assets/package_2.svg';
import package_2_icon_dark from '../../assets/package_2-dark.svg';
import { t } from '@apps/dashboard/boot/i18n';
import { ROUTE_NAME } from '@apps/dashboard/src/router/const';

export const MetricTypesFormat = (type = 'cluster') => ({
	cpu_usage: `${type}_cpu_usage`,
	cpu_total: `${type}_cpu_total`,
	cpu_utilisation: `${type}_cpu_utilisation`,
	memory_usage: `${type}_memory_usage_wo_cache`,
	memory_total: `${type}_memory_total`,
	memory_utilisation: `${type}_memory_utilisation`,
	disk_size_usage: `${type}_disk_size_usage`,
	disk_size_capacity: `${type}_disk_size_capacity`,
	disk_utilisation: `${type}_disk_size_utilisation`,
	pod_count: `${type}_pod_running_count`,
	pod_capacity: `${type}_pod_quota`
});

export const getValue = (data: PodItem) => get(data, 'value[1]', 0);

export const getTabOptions = (
	data: { [key: string]: string },
	MetricTypes: any,
	index = 0,
	isDark = false
) => {
	const lastData: { [key: string]: any } = getLastMonitoringData(data, index);
	const result = [
		{
			name: t('CPU'),
			unitType: 'cpu',
			used: getValue(lastData[MetricTypes.cpu_usage]),
			total: getValue(lastData[MetricTypes.cpu_total]),
			img: isDark ? memory_icon_dark : memory_icon,
			route: {
				name: ROUTE_NAME.CPU_DETAIL
			}
		},
		{
			name: t('MEMORY'),
			unitType: 'memory',
			used: getValue(lastData[MetricTypes.memory_usage]),
			total: getValue(lastData[MetricTypes.memory_total]),
			img: isDark ? memory_alt_icon_dark : memory_alt_icon,
			route: {
				name: ROUTE_NAME.MEMORY_DETAIL
			}
		},
		{
			name: t('DISK'),
			unitType: 'disk',
			used: getValue(lastData[MetricTypes.disk_size_usage]),
			total: getValue(lastData[MetricTypes.disk_size_capacity]),
			img: isDark ? hard_drive_icon_dark : hard_drive_icon,
			route: {
				name: ROUTE_NAME.DISK_DETAIL
			}
		},
		{
			name: t('PODS'),
			unit: '',
			used: getValue(lastData[MetricTypes.pod_count]),
			total: getValue(lastData[MetricTypes.pod_capacity]),
			img: isDark ? package_2_icon_dark : package_2_icon,
			route: {
				name: ROUTE_NAME.PODS_DETAIL
			}
		}
	];

	return result;
};

export const getContentOptions = (
	data: ResourcesResponse,
	MetricTypes: any,
	index = 0
) => {
	const result = [
		{
			type: 'utilisation',
			title: t('CPU_USAGE'),
			unit: '%',
			legend: ['USAGE'],
			data: [get(data, `${MetricTypes.cpu_utilisation}.data.result[${index}]`)]
		},
		{
			type: 'utilisation',
			title: t('MEMORY_USAGE'),
			unit: '%',
			legend: ['USAGE'],
			data: [
				get(data, `${MetricTypes.memory_utilisation}.data.result[${index}]`)
			]
		},
		{
			type: 'utilisation',
			title: t('DISK_USAGE'),
			unit: '%',
			legend: ['USAGE'],
			data: [get(data, `${MetricTypes.disk_utilisation}.data.result[${index}]`)]
		},
		{
			title: t('PODS'),
			unit: '',
			legend: ['COUNT'],
			data: [get(data, `${MetricTypes.pod_count}.data.result[${index}]`)],
			img: package_2_icon,
			img_active: package_2_active_icon
		}
	];

	return result;
};

export const getTabOptions2 = [
	{
		name: 'API_SERVER',
		title: firstToUpperWith_('REQUEST_LATENCY'),
		icon: 'server'
	},
	{
		name: 'API_SERVER',
		title: firstToUpperWith_('REQUEST_RATE'),
		icon: 'server'
	},
	{
		name: 'SCHEDULER',
		title: firstToUpperWith_('SCHEDULE_ATTEMPTS'),
		icon: 'ring'
	},
	{
		name: 'SCHEDULER',
		title: firstToUpperWith_('SCHEDULING_RATE'),
		icon: 'ring'
	}
];

export function getResult(result: any) {
	const data: any = {};
	const results = isArray(result) ? result : get(result, 'results', []) || [];

	if (isEmpty(results)) {
		const metricName = get(result, 'metric_name');

		if (metricName) {
			data[metricName] = result;
		}
	} else {
		results.forEach((item: any) => {
			data[item.metric_name] = item;
		});
	}

	return data;
}

export const MetricTypesUser = {
	cpu_usage: 'user_cpu_usage',
	memory_usage: 'user_memory_usage_wo_cache',
	cpu_total: 'user_cpu_total',
	memory_total: 'user_memory_total'
};

export const getTabOptionsUser = (
	data: { [key: string]: string },
	MetricTypes: any
) => {
	const lastData: { [key: string]: any } = getLastMonitoringData(data);
	const result = [
		{
			name: 'CPU',
			unitType: 'cpu',
			used: getValue(lastData[MetricTypes.cpu_usage]),
			total: getValue(lastData[MetricTypes.cpu_total])
		},
		{
			name: 'MEMORY',
			unitType: 'memory',
			used: getValue(lastData[MetricTypes.memory_usage]),
			total: getValue(lastData[MetricTypes.memory_total])
		}
	];

	return result;
};
export const getContentOptionsUser = (
	data: any,
	MetricTypes: any,
	index = 0
) => {
	const result = [
		{
			type: 'utilisation',
			unitType: 'cpu',
			legend: ['USAGE'],
			data: [get(data, `${MetricTypes.cpu_usage}.data.result[${index}]`)]
		},
		{
			type: 'utilisation',
			unitType: 'memory',
			legend: ['USAGE'],
			data: [get(data, `${MetricTypes.memory_usage}.data.result[${index}]`)]
		}
	];

	return result;
};
