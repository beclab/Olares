import { getNameSpacePodsList } from '../../network';
import { joinSelector } from '../../utils';
import { get, has, isEmpty, set } from 'lodash';
import { ObjectMapper } from '../../utils/object.mapper';
const resourceName = 'pod';

export const MetricTypes = {
	cpu: 'pod_cpu_usage',
	memory: 'pod_memory_usage_wo_cache',
	net_transmitted: 'pod_net_bytes_transmitted',
	net_received: 'pod_net_bytes_received'
};
export const podMetricsFormat = (monitoringData: any, pod: any) => {
	const metrics: any = {};
	Object.entries(MetricTypes).forEach(([key, value]) => {
		const records = get(monitoringData, `${value}.data.result`) || [];
		metrics[key] = records.find(
			(item: any) => get(item, 'metric.pod') === pod.name
		);
	});
	return metrics;
};

export const getTimeRange = ({ step = '600s', times = 20 } = {}) => {
	const interval = parseFloat(step) * times;
	const end = Math.floor(Date.now() / 1000);
	const start = Math.floor(end - interval);

	return { start, end };
};

const getNewValues = (origin = [], newValue = []) => {
	const values: any[] = isEmpty(origin) ? [] : [...origin];
	const value = newValue || [];

	if (!isEmpty(value)) {
		if (values.length > 10) {
			values.shift();
		}
		values.push(value);
	}

	return values;
};

const getNewRefreshedResult = (
	currentResult = [],
	originResult: any[] = []
) => {
	const newResult = [...originResult];

	currentResult.forEach((record: any, index) => {
		const newResourceName = get(record, `metric.${resourceName}`);
		let recordData: any = null;

		if (newResourceName) {
			const originRecord = newResult.find(
				(_record) => get(_record, `metric.${resourceName}`) === newResourceName
			);

			if (isEmpty(originRecord)) {
				newResult.push(record);
			} else {
				recordData = originRecord;
			}
		} else {
			recordData = newResult[index];
		}

		if (!isEmpty(recordData)) {
			const newValues = getNewValues(recordData?.values, record?.value);
			set(recordData, 'values', newValues);
		}
	});

	return newResult;
};

export const getRefreshResult = (newData = {}, origin = {}) => {
	const data = origin;
	if (isEmpty(data)) {
		return newData;
	}
	Object.values(newData).forEach((item) => {
		const key = get(item, 'metric_name');
		const metric = data[key!];
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

function getFilterParam(params: any) {
	const result = { ...params };
	if (result.app) {
		result.labelSelector = result.labelSelector || '';
		result.labelSelector += `app.kubernetes.io/name=${result.app}`;
		delete result.app;
	}
	return result;
}

export const getParams = (detail: any) => {
	console.log('detail', detail);
	const { name, namespace, kind: _kind, selector, _originData } = detail || {};
	const kind = _kind || get(_originData, 'kind', '');

	let result: any = {};

	result.namespace = namespace;

	switch (kind) {
		case 'PVC':
			result.pvcName = name;
			break;
		case 'Node':
			result.nodeName = name;
			break;
		case 'Namespace':
			result.namespace = name;
			break;
		case 'Service':
		case 'IPPool':
			result.labelSelector = joinSelector(selector);
			break;
		default:
			result.ownerKind = kind === 'Deployment' ? 'ReplicaSet' : kind;
			result.labelSelector = joinSelector(selector);
	}

	if (has(result, 'labelSelector') && isEmpty(selector)) {
		result = {};
	}
	return result;
};

export function getPosdList(detail: any, rest = {}) {
	const params = getParams(detail);
	return getNameSpacePodsList(getFilterParam(params), rest);
}

export function getPosdListFormatter(res: any) {
	return res.data.items.map((item: any) => ObjectMapper.pods(item));
}
