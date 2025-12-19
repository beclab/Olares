import { ResourcesResponse } from '@apps/control-hub/src/types/network';
import { get, isArray, isEmpty } from 'lodash';

export const MetricTypes = {
	cpu_usage: 'pod_cpu_usage',
	memory_usage: 'pod_memory_usage_wo_cache',
	net_transmitted: 'pod_net_bytes_transmitted',
	net_received: 'pod_net_bytes_received'
};

export function chartConfig(data: ResourcesResponse) {
	return [
		{
			type: 'cpu',
			title: 'CPU_USAGE',
			unitType: 'cpu',
			legend: ['USAGE'],
			data: get(data, `${MetricTypes.cpu_usage}.data.result`)
		},
		{
			type: 'memory',
			title: 'MEMORY_USAGE',
			unitType: 'memory',
			legend: ['USAGE'],
			data: get(data, `${MetricTypes.memory_usage}.data.result`)
		},
		{
			type: 'bandwidth',
			title: 'OUTBOUND_TRAFFIC',
			unitType: 'bandwidth',
			legend: ['OUTBOUND'],
			data: get(data, `${MetricTypes.net_transmitted}.data.result`)
		},
		{
			type: 'bandwidth',
			title: 'INBOUND_TRAFFIC',
			unitType: 'bandwidth',
			legend: ['INBOUND'],
			data: get(data, `${MetricTypes.net_received}.data.result`)
		}
	];
}

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
