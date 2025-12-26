import axios from 'axios';
import { useTokenStore } from 'src/stores/settings/token';

export interface UserUsage {
	user_cpu_total: string;
	user_cpu_usage: string;
	user_memory_usage: string;
	user_memory_total: string;
	user_disk_usage: string;
	user_disk_total: string;
}

const metricKeyMap = {
	user_cpu_total: ['cpu_total'],
	user_cpu_usage: ['cpu_usage'],
	user_memory_usage: ['memory_usage_wo_cache'],
	user_memory_total: ['memory_total'],
	user_disk_usage: ['disk_size_usage'],
	user_disk_total: ['disk_size_capacity']
};

async function fetchUsageData(
	urlPath: string,
	metricsFilter: string
): Promise<UserUsage> {
	const tokenStore = useTokenStore();
	const data: any = await axios.get(
		`${tokenStore.url}/kapis/monitoring.kubesphere.io/v1alpha3/${urlPath}`,
		{ params: { metrics_filter: metricsFilter } }
	);

	const requestedMetrics = metricsFilter
		.split('|')
		.map((metric) => metric.replace('$', '').trim());

	const metricValueMap: Record<string, string> = {};
	data.results.forEach((re: any) => {
		if (re.metric_name && re.data?.result?.[0]?.value?.[1]) {
			metricValueMap[re.metric_name] = re.data.result[0].value[1];
		}
	});

	const getMatchedValue = (key: keyof typeof metricKeyMap) => {
		const keywords = metricKeyMap[key];
		const matchedMetric = requestedMetrics.find((metric) =>
			keywords.some((keyword) => metric.includes(keyword))
		);
		return matchedMetric && metricValueMap[matchedMetric]
			? metricValueMap[matchedMetric]
			: '0';
	};

	return {
		user_cpu_total: getMatchedValue('user_cpu_total'),
		user_cpu_usage: getMatchedValue('user_cpu_usage'),
		user_memory_usage: getMatchedValue('user_memory_usage'),
		user_memory_total: getMatchedValue('user_memory_total'),
		user_disk_usage: getMatchedValue('user_disk_usage'),
		user_disk_total: getMatchedValue('user_disk_total')
	};
}

export async function get_cluster_resource(): Promise<UserUsage> {
	return fetchUsageData(
		'cluster/',
		'cluster_cpu_usage|cluster_memory_usage_wo_cache|cluster_cpu_total|cluster_memory_total|cluster_disk_size_usage|cluster_disk_size_capacity$'
	);
}

export async function get_user_resource(username: string): Promise<UserUsage> {
	return fetchUsageData(
		`users/${username}`,
		'user_cpu_usage|user_memory_usage_wo_cache|user_cpu_total|user_memory_total|user_disk_size_usage|user_disk_size_capacity$'
	);
}
