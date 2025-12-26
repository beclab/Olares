import { ResourcesResponse } from '../../network/network';
import { get, isArray, isEmpty } from 'lodash';

export const MetricTypes = {
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

const getTimeRange = ({ step = '600s', times = 20 } = {}) => {
	console.log('getTimeRange', step, times);
	const interval = parseFloat(step) * times;
	const end = Math.floor(Date.now() / 1000);
	const start = Math.floor(end - interval);

	return { start, end };
};

const getMinuteValue = (timeStr = '60s', hasUnit = true) => {
	const unit = timeStr.slice(-1);
	let value = parseFloat(timeStr);

	switch (unit) {
		default:
		case 's':
			break;
		case 'm':
			value *= 60;
			break;
		case 'h':
			value *= 60 * 60;
			break;
		case 'd':
			value = value * 24 * 60 * 60;
			break;
	}
	return hasUnit ? `${value}s` : value;
};

const handleParams = (params: any) => params;

interface MonitoringParams {
	start?: number;
	end?: number;
	step?: string;
	times?: number;
	resources?: string[];
	metrics?: string[];
	last?: boolean;
	[key: string]: any; // 用于支持 rest 参数
}

interface ProcessedParams {
	start?: number;
	end?: number;
	step?: string;
	times?: number;
	resources_filter?: string;
	metrics_filter?: string;
	[key: string]: any;
}

export const getParams = ({
	start = 0,
	end = 0,
	step = '600s',
	times = 20,
	resources = [],
	metrics = [],
	last = false, // last time
	...rest
}: MonitoringParams = {}): ProcessedParams => {
	const params: any = {
		...rest
	};
	if (!last) {
		Object.assign(params, {
			start,
			end,
			step: getMinuteValue(step),
			times
		});

		if (!start || !end) {
			console.log('start and end', start, end);
			const timeRange = getTimeRange(params);
			params.start = timeRange.start;
			params.end = timeRange.end;
		}
	}

	if (params.start) {
		params.start = Math.floor(params.start);
	}

	if (params.end) {
		params.end = Math.floor(params.end);
	}

	if (!isEmpty(resources)) {
		params.resources_filter = `${resources.join('|')}$`;
	}

	if (!isEmpty(metrics)) {
		params.metrics_filter = `${metrics.join('|')}$`;
	}
	console.log('start and end2', params);

	return handleParams(params);
};

export const fillEmptyMetrics = (params: any, result: any) => {
	console.log('fillEmptyMetrics', params, result);
	if (!params.times || !params.start || !params.end) {
		return result;
	}

	const format = (num: number) => String(num).replace(/\..*$/, '');
	const step = Math.floor((params.end - params.start) / params.times);
	const correctCount = params.times + 1;

	Object.values(result).forEach((item) => {
		const _result: any = get(item, 'data.result');
		if (!isEmpty(_result)) {
			_result.forEach((resultItem: any) => {
				const curValues = resultItem.values || [];
				const curValuesMap: any = curValues.reduce(
					(prev: any, cur: any) => ({
						...prev,
						[format(cur[0])]: cur[1]
					}),
					{}
				);

				if (curValues.length < correctCount) {
					const newValues: any = [];
					for (let index = 0; index < correctCount; index++) {
						const time = format(params.start + index * step);
						const data: any = [time, curValuesMap[time] || '0'];
						newValues.push(data);
					}
					resultItem.values = newValues;
				}
			});
		}
	});

	return result;
};

export const ParamsFormat = (
	createTime: string,
	times: number,
	step: string
) => {
	const params = {
		times,
		step
	};

	if (createTime) {
		const create = new Date(createTime).valueOf() / 1000;
		const now = Date.now() / 1000;
		const interval = now - create;

		switch (true) {
			// half an hour
			case interval <= 1800:
				params.times = 30;
				params.step = '1m';
				break;
			// an hour
			case interval <= 3600:
				params.times = 60;
				params.step = '1m';
				break;
			// five hours
			case interval <= 3600 * 5:
				params.times = 60;
				params.step = '5m';
				break;
			default:
				break;
		}
	}

	return params;
};
