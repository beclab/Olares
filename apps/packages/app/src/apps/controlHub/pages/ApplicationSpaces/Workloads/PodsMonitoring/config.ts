import { ResourcesResponse } from '@apps/control-hub/src/types/network';
import { get, isArray, isEmpty } from 'lodash';
import { t } from '@apps/control-hub/src/boot/i18n';

interface metric {
	metric: {
		namespace: string;
		node: string;
		owner_kind: string;
		owner_name: string;
		pod: string;
	};
	values: Array<[string, string]>;
	min_value: string;
	max_value: string;
	avg_value: string;
	sum_value: string;
	fee: string;
	resource_unit: string;
	currency_unit: string;
}

export const MetricTypes = {
	cpu_usage: 'pod_cpu_usage',
	memory_usage: 'pod_memory_usage_wo_cache',
	net_transmitted: 'pod_net_bytes_transmitted',
	net_received: 'pod_net_bytes_received'
};

export const MONITOR_MODULES = {
	deployments: 'deployment',
	statefulsets: 'statefulset',
	daemonsets: 'daemonset'
};

export function chartConfig(data: ResourcesResponse) {
	const result_one: metric[] = get(
		data,
		`${MetricTypes.cpu_usage}.data.result`,
		[]
	);
	const legend = result_one.map((item) => item.metric.pod);

	return [
		{
			type: 'cpu',
			title: t('CPU_USAGE'),
			unitType: 'cpu',
			legend,
			data: get(data, `${MetricTypes.cpu_usage}.data.result`)
		},
		{
			type: 'memory',
			title: t('MEMORY_USAGE'),
			unitType: 'memory',
			legend,
			data: get(data, `${MetricTypes.memory_usage}.data.result`)
		},
		{
			type: 'bandwidth',
			title: t('OUTBOUND_TRAFFIC'),
			unitType: 'bandwidth',
			legend,
			data: get(data, `${MetricTypes.net_transmitted}.data.result`)
		},
		{
			type: 'bandwidth',
			title: t('INBOUND_TRAFFIC'),
			unitType: 'bandwidth',
			legend,
			data: get(data, `${MetricTypes.net_received}.data.result`)
		}
	];
}

export const initParams = (props) => {
	const { times, step, createTime } = props || {};
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
