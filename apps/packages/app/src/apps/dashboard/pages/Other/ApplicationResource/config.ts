import { ResourcesResponse } from '@apps/dashboard/src/types/network';
import { get, isEmpty } from 'lodash';
import {
	getAreaChartOps,
	getZeroValues
} from '@apps/dashboard/src/utils/monitoring';

export const ClusterMetricTypes = {
	cpu_usage: 'cluster_cpu_usage',
	memory_usage: 'cluster_memory_usage_wo_cache',
	disk_usage: 'cluster_disk_size_usage'
};

export const AppMetricTypes = {
	deployment_count: 'cluster_deployment_count',
	statefulset_count: 'cluster_statefulset_count',
	daemonset_count: 'cluster_daemonset_count',
	job_count: 'cluster_job_count',
	cronjob_count: 'cluster_cronjob_count',
	pvc_count: 'cluster_pvc_count',
	service_count: 'cluster_service_count',
	route_count: 'cluster_ingresses_extensions_count',
	pod_running_count: 'cluster_pod_running_count'
};

export const ProjectsMetricTypes = {
	namespace_count: 'cluster_namespace_count'
};
const MetricTypes = {
	// cluster
	...ClusterMetricTypes,
	// app
	...AppMetricTypes,
	// project
	...ProjectsMetricTypes
};

export function chartConfigCluster(
	data: ResourcesResponse,
	configs: { [key: string]: any }
) {
	const newData = configs.map((item: any) => {
		const ops = item;

		const itemData = get(data, `${item.metricType}.data.result`) || [];
		const config = getAreaChartOps({
			...ops,
			data: isEmpty(itemData) ? [{ values: getZeroValues() }] : itemData
		});

		return config;
	});

	return newData;
}

export const getMonitoringCfgs = () => [
	{
		type: 'cpu',
		title: 'CPU',
		unitType: 'cpu',
		legend: ['USAGE'],
		metricType: MetricTypes.cpu_usage
	},
	{
		type: 'memory',
		title: 'MEMORY',
		unitType: 'memory',
		legend: ['USAGE'],
		metricType: MetricTypes.memory_usage
	}
];

export function chartConfigApp() {
	return [
		{
			type: 'deployment',
			title: 'DEPLOYMENT',
			legend: ['DEPLOYMENT_PL'],
			metricType: MetricTypes.deployment_count
		},
		{
			type: 'statefulset',
			title: 'STATEFULSET',
			legend: ['STATEFULSET_PL'],
			metricType: MetricTypes.statefulset_count
		},
		{
			type: 'daemonset',
			title: 'DAEMONSET',
			legend: ['DAEMONSET_PL'],
			metricType: MetricTypes.daemonset_count
		},
		{
			type: 'job',
			title: 'JOB',
			legend: ['JOB_PL'],
			metricType: MetricTypes.job_count
		},
		{
			type: 'cronjob',
			title: 'CRONJOB',
			legend: ['CRONJOB_PL'],
			metricType: MetricTypes.cronjob_count
		},
		{
			type: 'pvc',
			title: 'PERSISTENT_VOLUME_CLAIM',
			legend: ['PERSISTENT_VOLUME_CLAIM_PL'],
			metricType: MetricTypes.pvc_count
		},
		{
			type: 'service',
			title: 'SERVICE',
			legend: ['SERVICE_PL'],
			metricType: MetricTypes.service_count
		},
		{
			type: 'routes',
			title: 'ROUTE',
			legend: ['ROUTE_PL'],
			metricType: MetricTypes.route_count
		},
		{
			type: 'pod',
			title: 'POD',
			legend: ['RUNNING_PODS'],
			metricType: MetricTypes.pod_running_count
		}
	];
}

export function chartConfigProject(data: ResourcesResponse) {
	return [
		{
			unit: '',
			data: get(data, `${MetricTypes.namespace_count}.data.result`)
		}
	];
}
