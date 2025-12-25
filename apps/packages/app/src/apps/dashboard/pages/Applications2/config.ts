import {
	getRefreshResult,
	getTimeRange
} from '@apps/control-panel-common/src/containers/PodsList/config';
import {
	getMetrics,
	getNamespaces
} from '@apps/control-panel-common/src/network';
import { get, unionBy, orderBy, last, isEmpty, compact } from 'lodash';
import {
	fillEmptyMetrics,
	getParams,
	getResult
} from '@apps/control-panel-common/src/containers/Monitoring/config';

import {
	getAreaChartOps,
	getSuitableUnit,
	getValueByUnit
} from '@apps/dashboard/src/utils/monitoring';
import { t } from '@apps/dashboard/src/boot/i18n';
import { useAppList } from '@apps/dashboard/src/stores/AppList';
import { timeParams } from '../../../controlPanelCommon/config/resource.common';
import { ref } from 'vue';
const SYSTEM_FRONTEND_DEPLOYMENT = 'system-frontend-deployment';

const MetricTypes = {
	cpu_usage: 'pod_cpu_usage',
	memory_usage: 'pod_memory_usage_wo_cache',
	net_transmitted: 'pod_net_bytes_transmitted',
	net_received: 'pod_net_bytes_received'
};
const NamespaceMetricTypes = {
	cpu_usage: 'namespace_cpu_usage',
	memory_usage: 'namespace_memory_usage_wo_cache',
	net_transmitted: 'namespace_net_bytes_transmitted',
	net_received: 'namespace_net_bytes_received'
};
const appList = useAppList();

export const loadingApps = new Array(
	appList.appsWithNamespace.length || 9
).fill({});

export const loadingData: any = {
	cpu_usage: loadingApps,
	memory_usage: loadingApps,
	net_transmitted: loadingApps,
	net_received: loadingApps
};

const podResultData = ref([]);
const userResultData = ref([]);

let metricsAbortController: AbortController | null = null;
let namespacesAbortController: AbortController | null = null;

export const resetAppMetricsData = () => {
	podResultData.value = [];
	userResultData.value = [];
};

export const fetchWorkloadsMetrics = async (
	apps,
	namespace,
	sort = 'desc',
	autofresh = false
) => {
	if (metricsAbortController) {
		metricsAbortController.abort();
	}
	if (namespacesAbortController) {
		namespacesAbortController.abort();
	}

	metricsAbortController = new AbortController();
	namespacesAbortController = new AbortController();

	const metricsPods = Object.values(MetricTypes);
	const metricsNamespace = Object.values(NamespaceMetricTypes);

	const systemApps = apps.filter((item) => item.isSystem);
	const customApps = apps.filter((item) => !item.isSystem);

	const timeRange = getTimeRange(timeParams);

	const resources_filter_system = systemApps
		.map((item, index) => {
			if (index !== systemApps.length - 1) {
				return `${item.deployment}.*|`;
			} else {
				return `${item.deployment}.*`;
			}
		})
		.join('');

	const resources_filter_custom = compact(
		customApps.map((item) => item.namespace)
	);
	const param_filter = {
		cluster: 'default',
		metrics_filter: `${metricsPods.join('|')}$`,
		resources_filter: `${resources_filter_system}$`,
		sort_type: 'desc',
		...timeRange,
		...timeParams,
		last: false
	};

	const namespaceParams_filter = {
		metrics_filter: `${metricsNamespace.join('|')}$`,
		resources_filter: `${resources_filter_custom.join('|')}`,
		sort_type: 'desc',
		...timeRange,
		...timeParams,
		last: false
	};

	if (autofresh) {
		param_filter.last = true;
		namespaceParams_filter.last = true;
	}

	const param = getParams(param_filter);
	const namespaceParams = getParams(namespaceParams_filter);

	const [res, namespaceRes] = await Promise.all([
		getMetrics(namespace, param, {
			signal: metricsAbortController?.signal
		}),
		resources_filter_custom.length > 0
			? getNamespaces(namespaceParams, {
					signal: namespacesAbortController?.signal
			  })
			: Promise.resolve({ data: { results: [] } })
	]);

	let podFormateData = getResult(res.data.results);

	let namespaceFormatData = getResult(namespaceRes.data.results);

	if (autofresh) {
		podFormateData = getRefreshResult(podFormateData, podResultData.value);

		namespaceFormatData = getRefreshResult(
			namespaceFormatData,
			userResultData.value
		);
	}

	podResultData.value = podFormateData = fillEmptyMetrics(
		param,
		podFormateData
	);
	userResultData.value = namespaceFormatData = fillEmptyMetrics(
		namespaceParams,
		namespaceFormatData
	);

	const tempObj = {};

	for (const key in namespaceFormatData) {
		const target = namespaceFormatData[key];
		const namespaces = target?.data?.result?.map(
			(item) => item.metric.namespace
		);
		const rest = resources_filter_custom.filter(
			(item) => !namespaces?.includes(item)
		);
		const restObj = rest.map((item) => ({
			metric: {
				namespace: item,
				workspace: 'system-workspace'
			},
			values: [[]],
			min_value: '',
			max_value: '',
			avg_value: '',
			sum_value: '',
			fee: '',
			resource_unit: '',
			currency_unit: ''
		}));

		if (
			resources_filter_custom.length <= 1 &&
			isEmpty(get(target, 'data.result'))
		) {
			tempObj[key] = {
				...target,
				data: { ...target.data, result: restObj }
			};
		} else {
			tempObj[key] = {
				...target,
				data: { ...target.data, result: target?.data?.result?.concat(restObj) }
			};
		}
	}

	const namespaceResult = tempObj;
	const namespaceData = isEmpty(namespaceResult)
		? {}
		: fillEmptyMetrics(namespaceParams, namespaceResult);

	return Promise.resolve(
		formatResult(podFormateData, namespaceData, apps, sort)
	);
};

const getLastMonitoringData = (data: any) => {
	return parseFloat(get(last(get(data, 'values')), '[1]') || '0');
};

function chartConfigCpu(data: any) {
	return {
		type: 'cpu',
		title: t('CPU_USAGE'),
		unitType: 'cpu',
		legend: [t('CPU')],
		data: [data]
	};
}

function chartConfigMemory(data: any) {
	return {
		type: 'memory',
		title: t('MEMORY_USAGE'),
		unitType: 'memory',
		legend: [t('MEMORY')],
		data: [data]
	};
}

function chartConfigTraffic(data: any) {
	return {
		type: 'bandwidth',
		title: t('OUTBOUND_TRAFFIC'),
		unitType: 'bandwidth',
		legend: [t('OUTBOUND')],
		data: [data]
	};
}

const getTabOptions = (data: any, apps: any[]) => {
	const result = {
		cpu_usage: get(data, `${MetricTypes.cpu_usage}.data.result`, []).map(
			(item: any) => ({
				isSystem: true,
				ownerKind: item.metric.owner_kind,
				name: item.metric.pod.split('-').slice(0, -2).join('-'),
				value: getLastMonitoringData(item),
				chartData: getAreaChartOps(chartConfigCpu(item))
			})
		),
		memory_usage: get(data, `${MetricTypes.memory_usage}.data.result`, []).map(
			(item: any) => ({
				isSystem: true,
				ownerKind: item.metric.owner_kind,
				name: item.metric.pod.split('-').slice(0, -2).join('-'),
				value: getLastMonitoringData(item),
				chartData: getAreaChartOps(chartConfigMemory(item))
			})
		),
		net_transmitted: get(
			data,
			`${MetricTypes.net_transmitted}.data.result`,
			[]
		).map((item: any) => ({
			isSystem: true,
			ownerKind: item.metric.owner_kind,
			name: item.metric.pod.split('-').slice(0, -2).join('-'),
			value: getLastMonitoringData(item),
			chartData: getAreaChartOps(chartConfigTraffic(item))
		})),
		net_received: get(data, `${MetricTypes.net_received}.data.result`, []).map(
			(item: any) => ({
				isSystem: true,
				ownerKind: item.metric.owner_kind,
				name: item.metric.pod.split('-').slice(0, -2).join('-'),
				value: getLastMonitoringData(item),

				chartData: getAreaChartOps(chartConfigTraffic(item))
			})
		)
	};
	for (const key in result) {
		const systemFrontend = result[key].find(
			(item) => item.name === SYSTEM_FRONTEND_DEPLOYMENT
		);
		const systemFrontendApps = apps
			.filter((app) => app.deployment === SYSTEM_FRONTEND_DEPLOYMENT)
			.map((app) => ({ ...systemFrontend, name: app.name }));

		result[key] = result[key].concat(systemFrontendApps);
	}

	return result;
};

const getTabOptions2 = (data: any) => {
	const result = {
		cpu_usage: get(
			data,
			`${NamespaceMetricTypes.cpu_usage}.data.result`,
			[]
		).map((item: any) => ({
			name: item.metric.namespace,
			value: getLastMonitoringData(item),
			chartData: getAreaChartOps(chartConfigCpu(item))
		})),
		memory_usage: get(
			data,
			`${NamespaceMetricTypes.memory_usage}.data.result`,
			[]
		).map((item: any) => ({
			name: item.metric.namespace,
			value: getLastMonitoringData(item),
			chartData: getAreaChartOps(chartConfigMemory(item))
		})),
		net_transmitted: get(
			data,
			`${NamespaceMetricTypes.net_transmitted}.data.result`,
			[]
		).map((item: any) => ({
			name: item.metric.namespace,
			value: getLastMonitoringData(item),
			chartData: getAreaChartOps(chartConfigMemory(item))
		})),
		net_received: get(
			data,
			`${NamespaceMetricTypes.net_received}.data.result`,
			[]
		).map((item: any) => ({
			name: item.metric.namespace,
			value: getLastMonitoringData(item),
			chartData: getAreaChartOps(chartConfigMemory(item))
		}))
	};

	return result;
};

const formatResult = (res: any, namespaceRes: any, apps, sort) => {
	const data1 = getTabOptions(res, apps);
	const data2 = getTabOptions2(namespaceRes);
	const cpu_usage_total = data1.cpu_usage
		.concat(data2.cpu_usage)
		.map((item: any) => {
			const app = apps.find((app) =>
				app.isSystem
					? item.isSystem &&
					  (app.deployment === item.name || app.name === item.name)
					: app.namespace === item.name
			);
			if (app) {
				const unit = getSuitableUnit(item.value, 'cpu');
				const used = getValueByUnit(item.value, unit);
				return {
					unit,
					used,
					...item,
					...app
				};
			}
			return false;
		})
		.filter((item: any) => item);
	const cpu_usage = unionBy(
		orderBy(cpu_usage_total, ['value', 'title'], [sort, 'asc']),
		'title'
	);

	const net_transmitted_total = data1.net_transmitted
		.concat(data2.net_transmitted)
		.map((item: any) => {
			const app = apps.find((app) =>
				app.isSystem
					? item.isSystem &&
					  (app.deployment === item.name || app.name === item.name)
					: app.namespace === item.name
			);

			if (app) {
				const unit = getSuitableUnit(item.value, 'bandwidth');
				const used = getValueByUnit(item.value, unit);
				return {
					unit,
					used,
					...item,
					...app
				};
			}
			return false;
		})
		.filter((item: any) => item);
	const net_transmitted = unionBy(
		orderBy(net_transmitted_total, ['value', 'title'], [sort, 'asc']),
		'title'
	);

	const memory_usage_total = data1.memory_usage
		.concat(data2.memory_usage)
		.map((item: any) => {
			const app = apps.find((app) =>
				app.isSystem
					? item.isSystem &&
					  (app.deployment === item.name || app.name === item.name)
					: app.namespace === item.name
			);
			if (app) {
				const unit = getSuitableUnit(item.value, 'memory');
				const used = getValueByUnit(item.value, unit);
				return {
					unit,
					used,
					...item,
					...app
				};
			}
			return false;
		})
		.filter((item: any) => item);
	const memory_usage = unionBy(
		orderBy(memory_usage_total, ['value', 'title'], [sort, 'asc']),
		'title'
	);

	const net_received_total = data1.net_received
		.concat(data2.net_received)
		.map((item: any) => {
			const app = apps.find((app) =>
				app.isSystem
					? item.isSystem &&
					  (app.deployment === item.name || app.name === item.name)
					: app.namespace === item.name
			);
			if (app) {
				const unit = getSuitableUnit(item.value, 'bandwidth');
				const used = getValueByUnit(item.value, unit);
				return {
					unit,
					used,
					...item,
					...app
				};
			}
			return false;
		})
		.filter((item: any) => item);
	const net_received = unionBy(
		orderBy(net_received_total, ['value', 'title'], [sort, 'asc']),
		'title'
	);
	return {
		cpu_usage: cpu_usage,
		memory_usage: memory_usage,
		net_transmitted: net_transmitted,
		net_received: net_received
	};
};
