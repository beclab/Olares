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
import { t } from 'src/boot/dashboard-i18n';
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
	net_received: 'namespace_net_bytes_received',
	pod_count: 'namespace_pod_count'
};
const appList = useAppList();

export type WorkloadsMetricsFormatOptions = {
	preserveOrder?: string[];
};

export const loadingApps = new Array(
	appList.appsWithNamespace.length || 9
).fill({});

export const loadingData: any = {
	cpu_usage: loadingApps,
	memory_usage: loadingApps,
	net_transmitted: loadingApps,
	net_received: loadingApps
};

export const buildSkeletonMonitoringData = (rowCount: number) => {
	const n = Math.max(1, rowCount);
	return {
		cpu_usage: Array.from({ length: n }, () => ({})),
		memory_usage: Array.from({ length: n }, () => ({})),
		net_transmitted: Array.from({ length: n }, () => ({})),
		net_received: Array.from({ length: n }, () => ({}))
	};
};

const podResultData = ref<any>([]);
const userResultData = ref<any>([]);

let metricsAbortController: AbortController | null = null;
let namespacesAbortController: AbortController | null = null;

let partialMetricsAbortController: AbortController | null = null;
let partialNamespacesAbortController: AbortController | null = null;

export const resetAppMetricsData = () => {
	podResultData.value = [];
	userResultData.value = [];
};

export const cancelMainMetricsFetch = () => {
	metricsAbortController?.abort();
	namespacesAbortController?.abort();
};

const buildEmptyDateRangeValues = (params: any): any[] => {
	if (!params?.times || !params?.start || !params?.end) return [];
	const step = Math.floor((params.end - params.start) / params.times);
	const correctCount = params.times + 1;
	const format = (num: number) => String(num).replace(/\..*$/, '');
	const result: any[] = [];
	for (let i = 0; i < correctCount; i++) {
		const time = format(params.start + i * step);
		result.push([time, undefined]);
	}
	return result;
};

export const fetchWorkloadsMetrics = async (
	apps,
	namespace,
	sort = 'desc',
	autofresh = false,
	formatOptions?: WorkloadsMetricsFormatOptions
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

	const metricsPromise =
		systemApps.length > 0 && resources_filter_system.length > 0
			? getMetrics(namespace, param, {
					signal: metricsAbortController?.signal
			  })
			: Promise.resolve({ data: { results: [] } });

	const namespacesPromise =
		resources_filter_custom.length > 0
			? getNamespaces(namespaceParams, {
					signal: namespacesAbortController?.signal
			  })
			: Promise.resolve({ data: { results: [] } });

	const [res, namespaceRes] = await Promise.all([
		metricsPromise,
		namespacesPromise
	]);

	let podFormateData = getResult(res.data.results);

	let namespaceFormatData = getResult(namespaceRes.data.results);

	if (autofresh) {
		podFormateData = getRefreshResult(podFormateData, podResultData.value);

		namespaceFormatData = getRefreshResult(
			namespaceFormatData,
			userResultData.value,
			'namespace'
		);
	}

	podResultData.value = podFormateData = fillEmptyMetrics(
		param,
		podFormateData
	);
	namespaceFormatData = fillEmptyMetrics(namespaceParams, namespaceFormatData);

	const tempObj = {};
	const emptyDateRangeValues = buildEmptyDateRangeValues(namespaceParams);

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
			values: emptyDateRangeValues.map((v) => [...v]),
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

	userResultData.value = isEmpty(namespaceData)
		? namespaceFormatData
		: namespaceData;

	const appsForFormat = [...systemApps, ...customApps];
	const formatted = formatResult(
		podFormateData,
		namespaceData,
		appsForFormat,
		sort,
		formatOptions
	);
	return Promise.resolve(formatted);
};

const mergeIntoCache = (
	currentCache: any,
	additionData: any,
	resourceKey: string
) => {
	const cache =
		currentCache && !Array.isArray(currentCache) ? { ...currentCache } : {};
	Object.entries(additionData || {}).forEach(([metricName, metricObj]: any) => {
		if (!cache[metricName]) {
			cache[metricName] = metricObj;
			return;
		}
		const existingResult = get(cache[metricName], 'data.result', []) as any[];
		const additionResult = get(metricObj, 'data.result', []) as any[];
		const additionKeys = new Set(
			additionResult
				.map((r: any) => get(r, `metric.${resourceKey}`))
				.filter(Boolean)
		);
		const filteredExisting = existingResult.filter(
			(r: any) => !additionKeys.has(get(r, `metric.${resourceKey}`))
		);
		cache[metricName] = {
			...cache[metricName],
			data: {
				...get(cache[metricName], 'data', {}),
				result: [...filteredExisting, ...additionResult]
			}
		};
	});
	return cache;
};

export const fetchWorkloadsMetricsForApps = async (
	apps: any[],
	namespace: string,
	sort = 'desc'
) => {
	if (!apps || apps.length === 0) {
		return null;
	}

	if (partialMetricsAbortController) {
		partialMetricsAbortController.abort();
	}
	if (partialNamespacesAbortController) {
		partialNamespacesAbortController.abort();
	}

	partialMetricsAbortController = new AbortController();
	partialNamespacesAbortController = new AbortController();

	const metricsPods = Object.values(MetricTypes);
	const metricsNamespace = Object.values(NamespaceMetricTypes);

	const systemApps = apps.filter((item) => item.isSystem);
	const customApps = apps.filter((item) => !item.isSystem);

	const timeRange = getTimeRange(timeParams);

	const resources_filter_system = systemApps
		.map((item, index) =>
			index !== systemApps.length - 1
				? `${item.deployment}.*|`
				: `${item.deployment}.*`
		)
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

	const param = getParams(param_filter);
	const namespaceParams = getParams(namespaceParams_filter);

	const metricsPromise =
		systemApps.length > 0 && resources_filter_system.length > 0
			? getMetrics(namespace, param, {
					signal: partialMetricsAbortController?.signal
			  })
			: Promise.resolve({ data: { results: [] } });

	const namespacesPromise =
		resources_filter_custom.length > 0
			? getNamespaces(namespaceParams, {
					signal: partialNamespacesAbortController?.signal
			  })
			: Promise.resolve({ data: { results: [] } });

	const [res, namespaceRes] = await Promise.all([
		metricsPromise,
		namespacesPromise
	]);

	let podFormateData = getResult(res.data.results);
	let namespaceFormatData = getResult(namespaceRes.data.results);

	podFormateData = fillEmptyMetrics(param, podFormateData);
	namespaceFormatData = fillEmptyMetrics(namespaceParams, namespaceFormatData);

	const tempObj: any = {};
	const emptyDateRangeValues = buildEmptyDateRangeValues(namespaceParams);
	for (const key in namespaceFormatData) {
		const target = namespaceFormatData[key];
		const namespaces = target?.data?.result?.map(
			(item: any) => item.metric.namespace
		);
		const rest = resources_filter_custom.filter(
			(item) => !namespaces?.includes(item)
		);
		const restObj = rest.map((item) => ({
			metric: {
				namespace: item,
				workspace: 'system-workspace'
			},
			values: emptyDateRangeValues.map((v) => [...v]),
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

	const namespaceData = isEmpty(tempObj)
		? {}
		: fillEmptyMetrics(namespaceParams, tempObj);

	podResultData.value = mergeIntoCache(
		podResultData.value,
		podFormateData,
		'pod'
	);
	userResultData.value = mergeIntoCache(
		userResultData.value,
		namespaceData,
		'namespace'
	);

	const appsForFormat = [...systemApps, ...customApps];
	const formatted = formatResult(
		podFormateData,
		namespaceData,
		appsForFormat,
		sort
	);
	return formatted;
};

export const removeAppsFromMetricsCache = (
	removedApps: any[],
	remainingApps: any[]
) => {
	if (!removedApps?.length) return;

	const remainingPodKeys = new Set<string>();
	remainingApps
		.filter((a) => a.isSystem)
		.forEach((a) => {
			if (a.deployment) remainingPodKeys.add(a.deployment);
			if (a.name) remainingPodKeys.add(a.name);
		});

	const podCache = podResultData.value;
	if (podCache && !Array.isArray(podCache)) {
		Object.values(podCache).forEach((metric: any) => {
			const result = get(metric, 'data.result');
			if (Array.isArray(result)) {
				metric.data.result = result.filter((row: any) => {
					const dep = get(row, 'metric.pod') ? podDeploymentName(row) : null;
					return dep ? remainingPodKeys.has(dep) : true;
				});
			}
		});
	}

	const remainingNsKeys = new Set<string>();
	remainingApps
		.filter((a) => !a.isSystem)
		.forEach((a) => {
			if (a.namespace) remainingNsKeys.add(a.namespace);
		});

	const nsCache = userResultData.value;
	if (nsCache && !Array.isArray(nsCache)) {
		Object.values(nsCache).forEach((metric: any) => {
			const result = get(metric, 'data.result');
			if (Array.isArray(result)) {
				metric.data.result = result.filter((row: any) => {
					const ns = get(row, 'metric.namespace');
					return ns ? remainingNsKeys.has(ns) : true;
				});
			}
		});
	}
};

const MONITORING_METRIC_KEYS = [
	'cpu_usage',
	'memory_usage',
	'net_transmitted',
	'net_received'
] as const;

export const mergeMonitoringData = (
	existing: any,
	addition: any,
	sort: string
) => {
	const sortDirs = [sort as 'desc' | 'asc', 'asc'] as const;
	const merged: any = {};
	MONITORING_METRIC_KEYS.forEach((key) => {
		const existingRows: any[] = Array.isArray(existing?.[key])
			? existing[key]
			: [];
		const additionRows: any[] = Array.isArray(addition?.[key])
			? addition[key]
			: [];
		const byName = new Map<string, any>();
		existingRows.forEach((r: any) => {
			if (r?.name) byName.set(r.name, r);
		});
		additionRows.forEach((r: any) => {
			if (r?.name) byName.set(r.name, r);
		});
		const total = Array.from(byName.values());
		merged[key] = unionBy(
			orderBy(total, ['value', 'title'], sortDirs),
			'title'
		);
	});
	return merged;
};

export const removeMonitoringDataApps = (
	existing: any,
	removedNames: string[]
) => {
	const removedSet = new Set(removedNames);
	const result: any = {};
	MONITORING_METRIC_KEYS.forEach((key) => {
		const rows: any[] = Array.isArray(existing?.[key]) ? existing[key] : [];
		result[key] = rows.filter((r: any) => !removedSet.has(r?.name));
	});
	return result;
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

const podDeploymentName = (item: any) =>
	item.metric.pod.split('-').slice(0, -2).join('-');

const podCountByDeploymentFromPodMetrics = (data: any) => {
	const rows = get(data, `${MetricTypes.cpu_usage}.data.result`, []) as any[];
	return rows.reduce<Record<string, number>>((acc, item) => {
		const key = podDeploymentName(item);
		acc[key] = (acc[key] || 0) + 1;
		return acc;
	}, {});
};

const getTabOptions = (data: any, apps: any[]) => {
	const podCountByDeployment = podCountByDeploymentFromPodMetrics(data);
	const result = {
		cpu_usage: get(data, `${MetricTypes.cpu_usage}.data.result`, []).map(
			(item: any) => {
				const name = podDeploymentName(item);
				return {
					isSystem: true,
					ownerKind: item.metric.owner_kind,
					name,
					value: getLastMonitoringData(item),
					pod_acount: podCountByDeployment[name] || 0,
					chartData: getAreaChartOps(chartConfigCpu(item))
				};
			}
		),
		memory_usage: get(data, `${MetricTypes.memory_usage}.data.result`, []).map(
			(item: any) => {
				const name = podDeploymentName(item);
				return {
					isSystem: true,
					ownerKind: item.metric.owner_kind,
					name,
					value: getLastMonitoringData(item),
					pod_acount: podCountByDeployment[name] || 0,
					chartData: getAreaChartOps(chartConfigMemory(item))
				};
			}
		),
		net_transmitted: get(
			data,
			`${MetricTypes.net_transmitted}.data.result`,
			[]
		).map((item: any) => {
			const name = podDeploymentName(item);
			return {
				isSystem: true,
				ownerKind: item.metric.owner_kind,
				name,
				value: getLastMonitoringData(item),
				pod_acount: podCountByDeployment[name] || 0,
				chartData: getAreaChartOps(chartConfigTraffic(item))
			};
		}),
		net_received: get(data, `${MetricTypes.net_received}.data.result`, []).map(
			(item: any) => {
				const name = podDeploymentName(item);
				return {
					isSystem: true,
					ownerKind: item.metric.owner_kind,
					name,
					value: getLastMonitoringData(item),
					pod_acount: podCountByDeployment[name] || 0,
					chartData: getAreaChartOps(chartConfigTraffic(item))
				};
			}
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

const namespacePodCountByNs = (data: any) => {
	const rows = get(
		data,
		`${NamespaceMetricTypes.pod_count}.data.result`,
		[]
	) as any[];
	return rows.reduce<Record<string, number>>((acc, item) => {
		const ns = item.metric?.namespace;
		if (ns) acc[ns] = getLastMonitoringData(item);
		return acc;
	}, {});
};

const getTabOptions2 = (data: any) => {
	const podCountByNs = namespacePodCountByNs(data);
	const result = {
		cpu_usage: get(
			data,
			`${NamespaceMetricTypes.cpu_usage}.data.result`,
			[]
		).map((item: any) => ({
			name: item.metric.namespace,
			value: getLastMonitoringData(item),
			pod_acount: podCountByNs[item.metric.namespace] ?? 0,
			chartData: getAreaChartOps(chartConfigCpu(item))
		})),
		memory_usage: get(
			data,
			`${NamespaceMetricTypes.memory_usage}.data.result`,
			[]
		).map((item: any) => ({
			name: item.metric.namespace,
			value: getLastMonitoringData(item),
			pod_acount: podCountByNs[item.metric.namespace] ?? 0,
			chartData: getAreaChartOps(chartConfigMemory(item))
		})),
		net_transmitted: get(
			data,
			`${NamespaceMetricTypes.net_transmitted}.data.result`,
			[]
		).map((item: any) => ({
			name: item.metric.namespace,
			value: getLastMonitoringData(item),
			pod_acount: podCountByNs[item.metric.namespace] ?? 0,
			chartData: getAreaChartOps(chartConfigMemory(item))
		})),
		net_received: get(
			data,
			`${NamespaceMetricTypes.net_received}.data.result`,
			[]
		).map((item: any) => ({
			name: item.metric.namespace,
			value: getLastMonitoringData(item),
			pod_acount: podCountByNs[item.metric.namespace] ?? 0,
			chartData: getAreaChartOps(chartConfigMemory(item))
		}))
	};

	return result;
};

const orderMetricRows = (
	total: any[],
	sort: string,
	preserveOrder?: string[]
) => {
	const sortDirs = [sort as 'desc' | 'asc', 'asc'] as const;
	const representatives = unionBy(
		orderBy(total, ['value', 'title'], ['desc', 'asc']),
		'title'
	);
	const sortedRepresentatives = orderBy(
		representatives,
		['value', 'title'],
		sortDirs
	);
	if (preserveOrder?.length) {
		return preserveOrder
			.map((n) => representatives.find((item: any) => item.name === n))
			.filter(Boolean);
	}
	return sortedRepresentatives;
};

const formatResult = (
	res: any,
	namespaceRes: any,
	apps,
	sort,
	options?: WorkloadsMetricsFormatOptions
) => {
	const preserveOrder = options?.preserveOrder;
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
	const cpu_usage = orderMetricRows(cpu_usage_total, sort, preserveOrder);

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
	const net_transmitted = orderMetricRows(
		net_transmitted_total,
		sort,
		preserveOrder
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
	const memory_usage = orderMetricRows(memory_usage_total, sort, preserveOrder);

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
	const net_received = orderMetricRows(net_received_total, sort, preserveOrder);
	return {
		cpu_usage: cpu_usage,
		memory_usage: memory_usage,
		net_transmitted: net_transmitted,
		net_received: net_received
	};
};
