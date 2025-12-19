import {
	PodItem,
	ResourcesResponse
} from '@apps/control-hub/src/types/network';
import { get, isArray, isEmpty, set } from 'lodash';
import { getLastMonitoringData } from '@apps/control-hub/src/utils/monitoring';
import { _capitalize } from '@apps/control-hub/src/utils';
import { firstToUpperWith_ } from '@apps/control-hub/src/constant';
import memory_icon from '@apps/dashboard/src/assets/memory.svg';
import memory_icon_dark from '@apps/dashboard/src/assets/memory-dark.svg';
import memory_active_icon from '@apps/dashboard/src/assets/memory_active.svg';
import memory_alt_icon_dark from '@apps/dashboard/src/assets/memory_alt-dark.svg';
import memory_alt_active_icon from '@apps/dashboard/src/assets/memory_alt_active.svg';
import memory_alt_icon from '@apps/dashboard/src/assets/memory_alt.svg';
import hard_drive_active_icon from '@apps/dashboard/src/assets/hard_drive_active.svg';
import hard_drive_icon from '@apps/dashboard/src/assets/hard_drive.svg';
import hard_drive_icon_dark from '@apps/dashboard/src/assets/hard_drive-dark.svg';
import package_2_active_icon from '@apps/dashboard/src/assets/package_2_active.svg';
import package_2_icon from '@apps/dashboard/src/assets/package_2.svg';
import package_2_icon_dark from '@apps/dashboard/src/assets/package_2-dark.svg';
import { t } from '@apps/control-hub/src/boot/i18n';

const resourceName_ = 'resource_name';

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
	const lastData: { [key: string]: any } = getLastMonitoringData(data);
	const result = [
		{
			name: t('CPU'),
			unitType: 'cpu',
			used: getValue(lastData[MetricTypes.cpu_usage]),
			total: getValue(lastData[MetricTypes.cpu_total]),
			img: isDark ? memory_icon_dark : memory_icon
		},
		{
			name: t('MEMORY'),
			unitType: 'memory',
			used: getValue(lastData[MetricTypes.memory_usage]),
			total: getValue(lastData[MetricTypes.memory_total]),
			img: isDark ? memory_alt_icon_dark : memory_alt_icon
		},
		{
			name: t('DISK'),
			unitType: 'disk',
			used: getValue(lastData[MetricTypes.disk_size_usage]),
			total: getValue(lastData[MetricTypes.disk_size_capacity]),
			img: isDark ? hard_drive_icon_dark : hard_drive_icon
		},
		{
			name: t('PODS'),
			unit: '',
			used: getValue(lastData[MetricTypes.pod_count]),
			total: getValue(lastData[MetricTypes.pod_capacity]),
			img: isDark ? package_2_icon_dark : package_2_icon
		}
	];

	return result;
};

export const getTabOptions3 = (
	data: { [key: string]: string },
	MetricTypes: any,
	index = 0,
	isDark = false
) => {
	const lastData: { [key: string]: any } = getLastMonitoringData(data);
	const result = [
		{
			name: t('CPU'),
			unitType: 'cpu',
			used: getValue(lastData[MetricTypes.cpu_usage]),
			total: getValue(lastData[MetricTypes.cpu_total]),
			img: isDark ? memory_icon_dark : memory_icon
		},
		{
			name: t('MEMORY'),
			unitType: 'memory',
			used: getValue(lastData[MetricTypes.memory_usage]),
			total: getValue(lastData[MetricTypes.memory_total]),
			img: isDark ? memory_alt_icon_dark : memory_alt_icon
		},
		{
			name: t('PODS'),
			unit: '',
			used: getValue(lastData[MetricTypes.pod_count]),
			total: getValue(lastData[MetricTypes.pod_capacity]),
			img: isDark ? package_2_icon_dark : package_2_icon
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
			legend: [t('USAGE')],
			data: [get(data, `${MetricTypes.cpu_utilisation}.data.result[${index}]`)]
		},
		{
			type: 'utilisation',
			title: t('MEMORY_USAGE'),
			unit: '%',
			legend: [t('USAGE')],
			data: [
				get(data, `${MetricTypes.memory_utilisation}.data.result[${index}]`)
			]
		},
		{
			type: 'utilisation',
			title: t('DISK_USAGE'),
			unit: '%',
			legend: [t('USAGE')],
			data: [get(data, `${MetricTypes.disk_utilisation}.data.result[${index}]`)]
		},
		{
			title: t('PODS'),
			unit: '',
			legend: [t('COUNT')],
			data: [get(data, `${MetricTypes.pod_count}.data.result[${index}]`)],
			img: package_2_icon,
			img_active: package_2_active_icon
		}
	];

	return result;
};

export const getContentOptions3 = (
	data: ResourcesResponse,
	MetricTypes: any,
	index = 0
) => {
	const result = [
		{
			type: 'utilisation',
			title: t('CPU_USAGE'),
			unit: '%',
			legend: [t('USAGE')],
			data: [get(data, `${MetricTypes.cpu_utilisation}.data.result[${index}]`)]
		},
		{
			type: 'utilisation',
			title: t('MEMORY_USAGE'),
			unit: '%',
			legend: [t('USAGE')],
			data: [
				get(data, `${MetricTypes.memory_utilisation}.data.result[${index}]`)
			]
		},
		{
			title: t('PODS'),
			unit: '',
			legend: [t('COUNT')],
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

const getRefreshResult = (newData = {}, origin = {}) => {
	const data = origin;

	if (isEmpty(data)) {
		return newData;
	}

	Object.values(newData).forEach((item) => {
		const key = get(item, 'metric_name');
		// eslint-disable-next-line @typescript-eslint/ban-ts-comment
		// @ts-ignore
		const metric = data[key];

		if (metric) {
			const currentResult = get(item, 'data.result') || [];
			const originResult =
				get(metric, 'data.result', get(metric, 'data.results')) || [];

			set(
				metric,
				'data.result',
				getNewRefreshedResult(currentResult, originResult)
			);
		}
	});

	return data;
};

const getNewValues = (origin = [], newValue = []) => {
	const values = isEmpty(origin) ? [] : [...origin];
	const value = newValue || [];

	if (!isEmpty(value)) {
		if (values.length > 10) {
			values.shift();
		}
		// eslint-disable-next-line @typescript-eslint/ban-ts-comment
		// @ts-ignore
		values.push(value);
	}

	return values;
};

const getNewRefreshedResult = (currentResult = [], originResult = []) => {
	const newResult = [...originResult];

	currentResult.forEach((record, index) => {
		const resourceName = get(record, `metric.${resourceName_}`);
		let recordData = undefined;

		if (resourceName) {
			const originRecord = newResult.find(
				(_record) => get(_record, `metric.${resourceName_}`) === resourceName
			);

			if (isEmpty(originRecord)) {
				newResult.push(record);
			} else {
				recordData = originRecord;
			}
		} else {
			recordData = newResult[index];
		}

		if (!isEmpty(recordData) && recordData) {
			// eslint-disable-next-line @typescript-eslint/ban-ts-comment
			// @ts-ignore
			const newValues = getNewValues(recordData.values, record.value);
			set(recordData, 'values', newValues);
		}
	});

	return newResult;
};
