<template>
	<div class="q-gutter-y-md relative-position">
		<PodItem
			v-for="item in podList"
			:key="item.name"
			:data="item"
			:metrics="getPodMetrics(item)"
			:loading="loading3"
			:scroll="scroll"
			bordered
			@click="itemClick"
		></PodItem>
		<q-inner-loading :showing="loading2"> </q-inner-loading>
	</div>
	<Empty v-if="noData"></Empty>
</template>

<script setup lang="ts">
import { useRoute, useRouter, onBeforeRouteUpdate } from 'vue-router';
import { ref, watch, watchEffect, onBeforeUnmount, computed } from 'vue';
import { getMetrics } from '@apps/control-panel-common/src/network';
import { get, isArray, isEmpty, throttle } from 'lodash';
import { ObjectMapper } from '../../utils/object.mapper';
import PodItem from './PodItem.vue';
import { MonitoringResponse } from '@apps/control-panel-common/src/network/network';
import { usePodList } from '../../stores/podList';
import {
	getParams,
	getPosdList,
	getPosdListFormatter,
	getRefreshResult,
	getTimeRange,
	MetricTypes,
	podMetricsFormat
} from './config';
import Empty from '../../components/Empty.vue';
import axios from 'axios';
import isFunction from 'lodash';

import ReconnectingWebSocket from 'reconnecting-websocket';

interface Props {
	module?: string;
	detail?: any;
	scroll?: boolean;
	routePrefix?: string;
	path?: string;
	componentName?: string;
	routePushFunction?: (data: any) => void;
}
const PodListData = usePodList();
const CancelToken = axios.CancelToken;
let source = CancelToken.source();

const loading = ref(false);
const loading2 = ref(true);
const loading3 = ref(false);
const route = useRoute();
const router = useRouter();
const podList = ref();
const params = ref();
const monitoringData = ref();
const shouldExecuteResponseHandler = ref(true);

let locker: any = null;
let autofresh = false;
let createWS: any = undefined;

const props = withDefaults(defineProps<Props>(), {});
const emits = defineEmits(['podChange', 'update']);

const noData = computed(() => {
	return isEmpty(podList.value);
});

const fetchEnv = async (showLoading = true) => {
	params.value = getParams(props.detail);

	fetchPods(showLoading);
};

function updateDetail(data: any) {
	emits('podChange', data);
}

async function fetchPods(showLoading = true) {
	if (showLoading) {
		loading2.value = true;
		podList.value = [];
	}
	try {
		const res = await getPosdList(props.detail, { cancelToken: source.token });
		podList.value = getPosdListFormatter(res);
		PodListData.updateData(podList.value, Math.random());
		fetchMetrics();
		loading2.value = false;
		PodListData.updateLoading(false);
	} catch (error) {
		loading2.value = false;
	}
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
	if (createWS) {
		return;
	}
	const selectCluster = '';
	const { namespace, labelSelector }: any = params.value;
	const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';

	const urlProvider = async () => {
		return `${protocol}//${window.location.host}/api/v1/watch${
			selectCluster ? `/klusters/${selectCluster}` : ''
		}/namespaces/${namespace}/pods?labelSelector=${labelSelector}`;
	};

	createWS = new ReconnectingWebSocket(urlProvider);
	createWS.addEventListener('message', podWsHandler);
}

function closeWs() {
	if (createWS) {
		createWS.close(1000);
		createWS = null;
	}
}

function podWsHandler(MessageEvent: any) {
	const message = JSON.parse(MessageEvent.data);
	const selectCluster = '';
	const fetchPods2 = throttle(fetchPods, 350);
	const targetUrl = get(MessageEvent, 'currentTarget.url', '');
	targetUrl.split('app=')[1];

	if (targetUrl.split('app=')[1] !== route.params.pods_name) {
		return;
	}
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

let createWS2: any = undefined;
function monitoringWs() {
	if (createWS2) {
		return;
	}
	const {
		namespace,
		kind,
		pods_name: name
	}: Record<string, any> = route.params;

	const updateDetail2 = throttle(updateDetail, 350);

	const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
	const url = `${protocol}//${window.location.host}/apis/apps/v1/watch/namespaces/${namespace}/${kind}/${name}`;
	createWS2 = new ReconnectingWebSocket(url);

	createWS2.addEventListener('message', (MessageEvent: any) => {
		const message = JSON.parse(MessageEvent.data);
		updateDetail2(message.object);
	});
}

function closeWs2() {
	if (createWS2) {
		createWS2.close(1000);
		createWS2 = null;
	}
}

const getPodMetrics = (pod: any) => {
	const data = monitoringData.value;
	return podMetricsFormat(data, pod);
};

const getResult = (result: MonitoringResponse['results']) => {
	const data: any = {};
	const results = isArray(result) ? result : get(result, 'results', []) || [];

	if (isEmpty(results)) {
		const metricName = get(result, 'metric_name');

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
	const { namespace }: { [key: string]: any } = route.params;
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

	let timeParams = {
		step: '60s',
		times: 30
	};

	const timeRange = getTimeRange(timeParams);
	if (!autofresh) {
		param = { ...param, ...timeRange, ...timeParams };
		loading3.value = true;
	}

	getMetrics(namespace, param, { cancelToken: source.token })
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

const itemClick = (data: any) => {
	if (props.componentName) {
		router.push({
			name: props.componentName,
			params: {
				...route.params,
				name: data.name,
				uid: data.uid,
				node: data.node,
				createTime: data.createTime
			}
		});
		return;
	}
	if (props.routePushFunction && isFunction(props.routePushFunction)) {
		props.routePushFunction(data);
		return;
	}
	const podTemplateHash = data.labels['pod-template-hash'];
	const name = data.name.split(`-${podTemplateHash}`)[0];
	const path = props.path
		? props.path
		: `/${props.routePrefix || 'application-spaces'}/pods/overview/${
				data.node
		  }/${data.namespace}/${data.name}/${data.createTime}`;
	router.push({
		path,
		query: {
			type: 'pod',
			podName: name,
			uid: props.detail.uid
		}
	});
};

const CancelTokenHandler = () => {
	source.cancel();
};

watchEffect(() => {
	emits('update', podList.value);
});

watch(
	() => props.detail,
	async (newData, oldData) => {
		if (newData) {
			CancelTokenHandler();
			source = CancelToken.source();
			if (newData.uid === oldData?.uid) {
				fetchEnv(false);
			} else {
				fetchEnv();
				wsInit();
				monitoringWs();
			}
		}
	}
);

onBeforeRouteUpdate(() => {
	closeWs();
	closeWs2();
	CancelTokenHandler();
	shouldExecuteResponseHandler.value = false;
	PodListData.updateData([]);
	resetConfig();
});

onBeforeUnmount(() => {
	closeWs();
	closeWs2();
	CancelTokenHandler();
	shouldExecuteResponseHandler.value = false;
	PodListData.updateData([]);
	resetConfig();
});

defineExpose({ refresh: refreshPods });
</script>
