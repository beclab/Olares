import { ResourcesResponse } from '@apps/control-hub/src/types/network';
import { get, isArray, isEmpty } from 'lodash';

const MetricTypes = {
	cpu_usage: 'container_cpu_usage',
	memory_usage: 'container_memory_usage_wo_cache'
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
