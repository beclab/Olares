import { get } from 'lodash';
import { _capitalize } from '@apps/dashboard/src/utils';
import { firstToUpperWith_ } from '@apps/dashboard/src/constant';

export const ApiMetricTypes = {
	request_latencies_total: 'apiserver_request_latencies',
	request_latencies_apis: 'apiserver_request_by_verb_latencies',
	request_rate: 'apiserver_request_rate'
};

export const SchedulerMetricTypes = {
	schedule_attempts_count: 'scheduler_schedule_attempts',
	schedule_attempt_rate: 'scheduler_schedule_attempt_rate'
};

const MetricTypes2 = {
	...ApiMetricTypes,
	...SchedulerMetricTypes
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
