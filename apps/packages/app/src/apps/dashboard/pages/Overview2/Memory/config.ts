import { PodItem } from '@apps/dashboard/src/types/network';
import {
	getAreaChartOps,
	getLastMonitoringData
} from '@apps/dashboard/src/utils/monitoring';
import { get, round } from 'lodash';
import { t } from '@apps/dashboard/boot/i18n';
import { getDiskSize } from '@apps/dashboard/src/utils/disk';
import { getThroughput } from '@apps/dashboard/src/utils/memory';
import { timeRangeFormate } from '@apps/control-panel-common/src/containers/Monitoring/utils';

export enum MemoryType {
	PHYSICAL_VIDEO_MEMORY = 'PHYSICAL_VIDEO_MEMORY',
	EXCHANGE = 'EXCHANGE'
}

export const memoryOptions = [
	{
		label: t('MEMORY_OP.PHYSICAL_VIDEO_MEMORY'),
		value: MemoryType.PHYSICAL_VIDEO_MEMORY
	},
	{
		label: t('MEMORY_OP.EXCHANGE'),
		value: MemoryType.EXCHANGE
	}
];

export const MetricTypes = {
	memory_utilisation: 'node_memory_utilisation',
	memory_total: 'node_memory_total',
	memory_buffer_bytes: 'node_memory_buffer_bytes',
	memory_cached_bytes: 'node_memory_cached_bytes',
	memory_system_reserved: 'node_memory_system_reserved',
	memory_available: 'node_memory_available',
	memory_usage_wo_cache: 'node_memory_usage_wo_cache',
	vmstat_pswpout_bytes: 'node_vmstat_pswpout_bytes',
	vmstat_pswpin_bytes: 'node_vmstat_pswpin_bytes'
};
export type MetricTypesType = typeof MetricTypes;

export const getValue = (data: PodItem) => get(data, 'value[1]', 0);

export const getMemoryList = (
	data: { [key: string]: string },
	MetricTypes: MetricTypesType
) => {
	const firstData: any = get(
		data,
		`${MetricTypes.memory_total}.data.result`,
		[]
	);
	return firstData.map((item, index) => {
		return {
			MemoryChartData: getAreaChartOps(getMonitoringCfgs(data, index)),
			MemoryChartDataExchange: getAreaChartOps(
				getMonitoringExchange(data, index)
			),
			memoryStructure: getMemoryOptions(data, MetricTypes, index),
			memoryStructure2: getMemoryOptions2(data, MetricTypes, index),
			memoryExchangeList: getMemoryExchange(data, MetricTypes, index),
			memoryExchangeList2: getMemoryExchange2(data, MetricTypes, index)
		};
	});
};

export const getMonitoringCfgs = (
	data: { [key: string]: string },
	index = 0
) => ({
	type: 'load',
	unit: '%',
	title: get(
		data,
		`${MetricTypes.memory_utilisation}.data.result[${index}].metric.node`,
		undefined
	),
	legend: [t('MEMORY_OP.UTILIZATION_RATE')],
	data: [
		get(data, `${MetricTypes.memory_utilisation}.data.result[${index}]`, {})
	]
});

export const getMemoryOptions = (
	data: { [key: string]: string },
	MetricTypes: MetricTypesType,
	index = 0
) => {
	const lastData: { [key: string]: any } = getLastMonitoringData(data, index);
	const memory_system_reserved = getValue(
		lastData[MetricTypes.memory_system_reserved]
	);
	const memory_usage_wo_cache = getValue(
		lastData[MetricTypes.memory_usage_wo_cache]
	);
	const memory_buffer_bytes = getValue(
		lastData[MetricTypes.memory_buffer_bytes]
	);
	const memory_cached_bytes = getValue(
		lastData[MetricTypes.memory_cached_bytes]
	);
	const memory_available = getValue(lastData[MetricTypes.memory_available]);
	const result = [
		{
			color: 'orange-default',
			label: t('MEMORY_OP.RESERVED'),
			value: getDiskSize(memory_system_reserved),
			size: memory_system_reserved,
			info: t('MEMORY_OP.RESERVED_INFO')
		},
		{
			color: 'warning',
			label: t('MEMORY_OP.USED'),
			value: getDiskSize(memory_usage_wo_cache),
			size: memory_usage_wo_cache,
			info: t('MEMORY_OP.USED_INFO')
		},
		{
			color: 'light-blue-default',
			label: t('MEMORY_OP.BUFFER'),
			value: getDiskSize(memory_buffer_bytes),
			size: memory_buffer_bytes,
			info: t('MEMORY_OP.BUFFER_INFO')
		},
		{
			color: 'positive',
			label: t('MEMORY_OP.CACHE'),
			value: getDiskSize(memory_cached_bytes),
			size: memory_cached_bytes,
			info: t('MEMORY_OP.CACHE_INFO')
		},
		{
			color: 'info',
			label: t('MEMORY_OP.AVAILABLE'),
			value: getDiskSize(memory_available),
			size: memory_available,
			info: t('MEMORY_OP.AVAILABLE_INFO')
		}
	];

	return result;
};

export const getMemoryOptions2 = (
	data: { [key: string]: string },
	MetricTypes: MetricTypesType,
	index = 0
) => {
	const lastData: { [key: string]: any } = getLastMonitoringData(data, index);

	const result = [
		{
			label: t('MEMORY_OP.TOTAL'),
			value: getDiskSize(getValue(lastData[MetricTypes.memory_total]))
		},
		{
			label: t('MEMORY_OP.UTILIZATION_RATE'),
			value:
				round(getValue(lastData[MetricTypes.memory_utilisation]) * 100, 2) + '%'
		}
	];

	return result;
};

export const getMonitoringExchange = (
	data: { [key: string]: string },
	index = 0
) => ({
	type: 'throughput',
	unitType: 'throughput',
	title: get(
		data,
		`${MetricTypes.vmstat_pswpout_bytes}.data.result[${index}].metric.node`,
		undefined
	),
	legend: [t('MEMORY_OP.SWAP_IN'), t('MEMORY_OP.EXCHANGE_OUT')],
	data: [
		get(data, `${MetricTypes.vmstat_pswpout_bytes}.data.result[${index}]`, {}),
		get(data, `${MetricTypes.vmstat_pswpin_bytes}.data.result[${index}]`, {})
	]
});

export const getMemoryExchange = (
	data: { [key: string]: string },
	MetricTypes: MetricTypesType,
	index = 0
) => {
	const lastData: { [key: string]: any } = getLastMonitoringData(data, index);

	const result = [
		{
			color: 'light-blue-default',
			label: t('MEMORY_OP.SWAP_IN'),
			value: getThroughput(getValue(lastData[MetricTypes.vmstat_pswpin_bytes]))
		},
		{
			color: 'positive',
			label: t('MEMORY_OP.EXCHANGE_OUT'),
			value: getThroughput(getValue(lastData[MetricTypes.vmstat_pswpout_bytes]))
		}
	];

	return result;
};

export const getMemoryExchange2 = (
	data: { [key: string]: string },
	MetricTypes: MetricTypesType,
	index = 0
) => {
	const lastData: { [key: string]: any } = getLastMonitoringData(data, index);
	const memory_usage_wo_cache = getValue(
		lastData[MetricTypes.memory_usage_wo_cache]
	);
	const memory_total = getValue(lastData[MetricTypes.memory_total]);
	const result = [
		{
			color: 'ink-3',
			label: t('MEMORY_OP.TOTAL'),
			value: getDiskSize(memory_total)
		},
		{
			color: 'ink-3',
			label: t('MEMORY_OP.USED'),
			value: getDiskSize(memory_usage_wo_cache)
		},
		{
			color: 'ink-3',
			label: t('MEMORY_OP.UTILIZATION_RATE'),
			value:
				round(getValue(lastData[MetricTypes.memory_utilisation]) * 100, 2) + '%'
		}
	];

	return result;
};
