import { useRoute } from 'vue-router';
import { ref } from 'vue';
import {
	getMetrics,
	getNameSpacePodsList,
	patchWorkloadsControler
} from '@apps/control-panel-common/src/network';
import { get, isArray, isEmpty, set, throttle } from 'lodash';
import { ObjectMapper } from '@apps/control-panel-common/src/utils/object.mapper';
import SocketClient from '@apps/control-panel-common/src/utils/socket.client';
import { MonitoringResponse } from '@apps/control-panel-common/src/network/network';
export function useSub() {
	const loading = ref(false);
	const loading2 = ref(false);
	const loading3 = ref(false);
	const route = useRoute();
	const podList = ref();
	const params = ref();
	const childComponentRef = ref();
	const monitoringData = ref();
	const resourceName = ref('pod');

	let locker: any = null;
	let autofresh = false;
	let createWS: any = undefined;

	const emits = defineEmits(['podChange']);

	function updateDetail(data: any) {
		emits('podChange', data);
	}

	function getFilterParam(params: any) {
		const result = { ...params };
		if (result.app) {
			result.labelSelector = result.labelSelector || '';
			result.labelSelector += `app.kubernetes.io/name=${result.app}`;
			delete result.app;
		}
		return result;
	}

	function fetchPods(showLoading = true) {
		if (showLoading) {
			loading2.value = true;
		}
		getNameSpacePodsList(getFilterParam(params.value))
			.then((res) => {
				podList.value = res.data.items.map((item) => ObjectMapper.pods(item));
				fetchMetrics();
			})
			.finally(() => {
				loading2.value = false;
			});
	}

	const refreshPods = () => {
		resetConfig();
		fetchPods();
	};

	const resetConfig = () => {
		clearLocker();

		autofresh = false;
	};

	function wsInit() {
		const selectCluster = '';
		const { namespace, labelSelector }: any = params.value;
		const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
		const url = `${protocol}//${window.location.host}/api/v1/watch${
			selectCluster ? `/klusters/${selectCluster}` : ''
		}/namespaces/${namespace}/pods?labelSelector=${labelSelector}`;

		closeWs();
		createWS = new SocketClient(url, {
			onmessage: podWsHandler
		});
	}

	function closeWs() {
		if (createWS) {
			createWS.close(true);
			createWS = null;
		}
	}

	function podWsHandler(message: any) {
		const selectCluster = '';
		const fetchPods2 = throttle(fetchPods, 2000);
		const updateDetail2 = throttle(updateDetail, 2000);

		if (message.object.kind === 'Pod') {
			if (message.type === 'MODIFIED') {
				const data = {
					cluster: selectCluster,
					...ObjectMapper.pods(message.object)
				};
				updateItem(data);
			} else if (message.type === 'DELETED' || message.type === 'ADDED') {
				if (autofresh) {
					fetchPods2(false);
				}
			}
		} else {
			updateDetail2(message.object);
		}
	}

	function updateItem(item: any) {
		if (item.uid) {
			const index = podList.value.findIndex(
				(_item: any) => _item.uid === item.uid
			);
			if (index >= 0) {
				podList.value[index] = item;
			}
		}
	}

	function replicaChange(value: number) {
		const params = { spec: { replicas: value } };
		const { namespace, kind, name }: { [key: string]: any } = route.params;
		patchWorkloadsControler(namespace, kind, name, params).finally(() => {
			loading.value = false;
		});
	}

	function refreshEvent() {
		childComponentRef.value && childComponentRef.value.fetchData();
	}

	const getTimeRange = ({ step = '600s', times = 20 } = {}) => {
		const interval = parseFloat(step) * times;
		const end = Math.floor(Date.now() / 1000);
		const start = Math.floor(end - interval);

		return { start, end };
	};
	const MetricTypes = {
		cpu: 'pod_cpu_usage',
		memory: 'pod_memory_usage_wo_cache'
	};

	const getPodMetrics = (pod: any) => {
		const data = monitoringData.value;
		const metrics: any = {};

		Object.entries(MetricTypes).forEach(([key, value]) => {
			const records = get(data, `${value}.data.result`) || [];
			metrics[key] = records.find(
				(item: any) => get(item, 'metric.pod') === pod.name
			);
		});
		return metrics;
	};

	const getResult = (result: MonitoringResponse['results']) => {
		const data: any = {};
		const results = isArray(result) ? result : get(result, 'results', []) || [];

		if (isEmpty(results)) {
			const metricName = get(result, 'metric_name') as unknown as string;

			if (metricName) {
				data[metricName] = result;
			}
		} else {
			results.forEach((item) => {
				data[item.metric_name] = item;
			});
		}

		return data;
	};

	function fetchMetrics() {
		const {
			namespace,
			kind,
			name: podName
		}: { [key: string]: any } = route.params;
		const podListData = podList.value ?? [];

		const resources = podListData.map((item: any) => item.name);
		const metrics = Object.values(MetricTypes);
		let param = {
			cluster: 'default',
			ownerKind: params.value.ownerKind,
			labelSelector: params.value.labelSelector,
			resources_filter: `${resources.join('|')}$`,
			metrics_filter: `${metrics.join('|')}$`
		};

		const timeParams = {
			step: '60s',
			times: 30
		};

		const timeRange = getTimeRange(timeParams);
		if (!autofresh) {
			param = { ...param, ...timeRange, ...timeParams };
			loading3.value = true;
		}

		getMetrics(namespace, param)
			.then((res) => {
				if (autofresh) {
					monitoringData.value = getRefreshResult(
						res.data.results,
						monitoringData.value
					);
				} else {
					autofresh = true;
					monitoringData.value = getResult(res.data.results);
				}
				fetchMetricsAuto();
			})
			.finally(() => {
				loading3.value = false;
			});
	}

	const fetchMetricsAuto = () => {
		clearLocker();
		locker = setTimeout(() => {
			fetchMetrics();
		}, 3 * 1000);
	};

	const clearLocker = () => {
		if (locker) {
			clearTimeout(locker);
		}
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
			const newResourceName = get(record, `metric.${resourceName.value}`);
			let recordData: any = null;

			if (newResourceName) {
				const originRecord = newResult.find(
					(_record) =>
						get(_record, `metric.${resourceName.value}`) === newResourceName
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

	const getRefreshResult = (newData = {}, origin = {}) => {
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

	// watch(
	//   () => props.detail,
	//   async (newData, oldData) => {
	//     if (newData) {
	//       if (newData.uid === oldData?.uid) {
	//         fetchEnv(false);
	//       } else {
	//         resetConfig();
	//         fetchEnv();
	//       }
	//     }
	//   }
	// );

	return {
		podList,
		refreshPods
	};
}
