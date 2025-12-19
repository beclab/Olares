import {
	PodItem,
	ResourcesResponse
} from '@apps/control-hub/src/types/network';
import { get } from 'lodash';
import { getLastMonitoringData } from '@apps/control-hub/src/utils/monitoring';
import { _capitalize } from '@apps/control-hub/src/utils';
import { firstToUpperWith_ } from '@apps/control-hub/src/constant';

const MetricTypes = {
	cpu_usage: 'cluster_cpu_usage',
	cpu_total: 'cluster_cpu_total',
	cpu_utilisation: 'cluster_cpu_utilisation',
	memory_usage: 'cluster_memory_usage_wo_cache',
	memory_total: 'cluster_memory_total',
	memory_utilisation: 'cluster_memory_utilisation',
	disk_size_usage: 'cluster_disk_size_usage',
	disk_size_capacity: 'cluster_disk_size_capacity',
	disk_utilisation: 'cluster_disk_size_utilisation',
	pod_count: 'cluster_pod_running_count',
	pod_capacity: 'cluster_pod_quota'
};

const MetricTypes2 = {
	request_latencies_total: 'apiserver_request_latencies',
	request_latencies_apis: 'apiserver_request_by_verb_latencies',
	request_rate: 'apiserver_request_rate',
	schedule_attempts_count: 'scheduler_schedule_attempts',
	schedule_attempt_rate: 'scheduler_schedule_attempt_rate'
};

const getValue = (data: PodItem) => get(data, 'value[1]', 0);

export const getTabOptions = (data: { [key: string]: string }) => {
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
		},
		{
			name: 'DISK',
			unitType: 'disk',
			used: getValue(lastData[MetricTypes.disk_size_usage]),
			total: getValue(lastData[MetricTypes.disk_size_capacity])
		},
		{
			name: 'PODS',
			unit: '',
			used: getValue(lastData[MetricTypes.pod_count]),
			total: getValue(lastData[MetricTypes.pod_capacity])
		}
	];

	return result;
};

export const getContentOptions = (data: ResourcesResponse) => {
	const result = [
		{
			type: 'utilisation',
			title: 'CPU_USAGE',
			unit: '%',
			legend: ['USAGE'],
			data: get(data, `${MetricTypes.cpu_utilisation}.data.result`)
		},
		{
			type: 'utilisation',
			title: 'MEMORY_USAGE',
			unit: '%',
			legend: ['USAGE'],
			data: get(data, `${MetricTypes.memory_utilisation}.data.result`)
		},
		{
			type: 'utilisation',
			title: 'DISK_USAGE',
			unit: '%',
			legend: ['USAGE'],
			data: get(data, `${MetricTypes.disk_utilisation}.data.result`)
		},
		{
			title: 'POD_COUNT',
			unit: '',
			legend: ['COUNT'],
			data: get(data, `${MetricTypes.pod_count}.data.result`)
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

const getSpecificData = (
	metricName: Metric2Keys,
	type: string,
	value: string | number,
	metrics: { [key: string]: any }
) => {
	const data = get(metrics, `${MetricTypes2[metricName]}.data.result`) || [];
	return data.find((item: any) => get(item, `metric.${type}`) === value) || {};
};

const getVerbData = (value: string, data: { [key: string]: any }) =>
	getSpecificData('request_latencies_apis', 'verb', value, data);

type Metric2Keys = keyof typeof MetricTypes2;

export const getContentOptions2 = (data: { [key: string]: any }) => {
	const metrics = data;
	const result = [
		{
			type: 'area',
			title: 'REQUEST_LATENCY',
			unit: 'ms',
			legend: ['GET', 'POST', 'PATCH', 'DELETE', 'PUT', 'TOTAL'],
			data: [
				getVerbData('GET', data),
				getVerbData('POST', data),
				getVerbData('PATCH', data),
				getVerbData('DELETE', data),
				getVerbData('PUT', data),
				get(metrics, `${MetricTypes2.request_latencies_total}.data.result[0]`)
			]
		},
		{
			type: 'area',
			title: 'REQUEST_PER_SECOND',
			unit: _capitalize('TIMES_PER_SECOND'),
			legend: ['REQUEST'],
			data: get(metrics, `${MetricTypes2.request_rate}.data.result`)
		},
		{
			type: 'area',
			title: 'SCHEDULE_ATTEMPTS',
			unit: '',
			legend: ['SUCCESS', 'ERROR', 'FAILURE'],
			data: [
				getSpecificData('schedule_attempts_count', 'result', 'scheduled', data),
				getSpecificData('schedule_attempts_count', 'result', 'error', data),
				getSpecificData(
					'schedule_attempts_count',
					'result',
					'unschedulable',
					data
				)
			],
			areaColors: ['blue', 'red', 'yellow']
		},
		{
			type: 'area',
			title: 'SCHEDULING_RATE',
			unit: _capitalize('TIMES_PER_SECOND'),
			legend: ['SUCCESS', 'ERROR', 'FAILURE'],
			data: [
				getSpecificData('schedule_attempt_rate', 'result', 'scheduled', data),
				getSpecificData('schedule_attempt_rate', 'result', 'error', data),
				getSpecificData(
					'schedule_attempt_rate',
					'result',
					'unschedulable',
					data
				)
			],
			areaColors: ['blue', 'red', 'yellow']
		}
	];

	return result;
};
