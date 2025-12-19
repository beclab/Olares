<template>
	<q-page class="absolute-full">
		<FullPageWithBack :title="namespace">
			<div class="q-gutter-y-xl">
				<q-card flat>
					<MyTitle size="lg">{{ $t('QUOTA') }}</MyTitle>
					<Quota class="q-mt-sm"></Quota>
				</q-card>
				<q-card class="my-card" flat>
					<MyTitle size="lg">Pods</MyTitle>
					<div class="row items-center q-mt-xs">
						<div class="row q-col-gutter-x-md">
							<QInputStyle>
								<q-input
									v-model="name"
									type="search"
									outlined
									debounce="500"
									:placeholder="t('SEARCH_BY_NAME')"
									clearable
									style="width: 240px"
									@update:model-value="change"
								>
									<template v-slot:prepend>
										<q-icon name="search" size="xs" />
									</template>
								</q-input>
							</QInputStyle>
							<QSectionStyle>
								<q-select
									v-model="selectValue"
									@update:model-value="fetchData"
									:options="options"
									dense
									outlined
									style="width: 220px"
								/>
							</QSectionStyle>
						</div>
						<SortButtom
							@change="sortChange"
							:default-value="sort_type"
						></SortButtom>
					</div>
					<div class="q-gutter-y-lg q-mt-lg">
						<div v-for="(item, index) in podListBySort" :key="item.name">
							<PodItem
								bordered
								:data="item"
								:metrics="getPodMetrics(item)"
								:loading="loading3"
								:class="index === 0 ? '' : 'q-mt-md'"
								@click="itemClick"
							></PodItem>
						</div>
					</div>
					<Empty
						size="large"
						class="q-mt-xl"
						v-if="(!podList || podList.length === 0) && !loading"
					></Empty>
				</q-card>
			</div>
			<q-inner-loading :showing="loading"> </q-inner-loading>
		</FullPageWithBack>
		<RouterViewTransition></RouterViewTransition>
	</q-page>
</template>

<script setup lang="ts">
import {
	getPosdListFormatter,
	getRefreshResult,
	getTimeRange,
	MetricTypes,
	podMetricsFormat
} from '@apps/control-panel-common/src/containers/PodsList/config';
import PodItem from '@apps/control-panel-common/src/containers/PodsList/PodItem.vue';
import {
	getMetrics,
	getNameSpacePodsList
} from '@apps/control-panel-common/src/network';
import { usePodList } from '@apps/control-panel-common/src/stores/podList';
import { onMounted, ref, watch, onBeforeUnmount, computed } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import Empty from '@apps/control-panel-common/src/components/Empty.vue';
import { getResult } from '@apps/dashboard/src/utils/monitoring';
import { t } from '@apps/dashboard/src/boot/i18n';
import axios from 'axios';
import { UsePod } from '@apps/control-panel-common/src/stores/PodData';
import FullPageWithBack from '@apps/control-panel-common/src/components/FullPageWithBack.vue';
import RouterViewTransition from '@apps/control-panel-common/src/components/RouterViewTransition.vue';
import QInputStyle from '@apps/control-panel-common/src/components/QInputStyle.vue';
import SortButtom from '@apps/control-panel-common/src/components/SortButton.vue';
import MyTitle from '@apps/control-panel-common/src/components/MyListItem/MyTitle.vue';
import Quota from './QuotaPage.vue';
import { get } from 'lodash';
import QSectionStyle from '@apps/control-panel-common/src/components/QSectionStyle.vue';

const options = [
	{
		label: t('SORT_BY_NODE_MEMORY_UTILISATION'),
		value: 'pod_memory_usage_wo_cache',
		sortBy: 'memory_usage_wo_cache'
	},
	{
		label: t('SORT_BY_NODE_CPU_UTILISATION'),
		value: 'pod_cpu_usage',
		sortBy: 'cpu_usage'
	},
	{
		label: t('SORT_BY_INBOUND_TRAFFIC'),
		value: 'pod_net_bytes_received',
		sortBy: 'net_bytes_received'
	},
	{
		label: t('SORT_BY_OUTBOUND_TRAFFIC'),
		value: 'pod_net_bytes_transmitted',
		sortBy: 'net_bytes_transmitted'
	}
];
const selectValue = ref(options[0]);
const usePod = UsePod();

const CancelToken = axios.CancelToken;
const source = CancelToken.source();

const name = ref();
const loading = ref(false);
const podList = ref([]);
const loading3 = ref(false);
const monitoringData = ref();
const sort_type = ref('desc');
let autofresh = false;
let locker: any = null;

const getPodMetrics = (pod: any) => {
	const data = monitoringData.value;
	return podMetricsFormat(data, pod);
};
const route = useRoute();
const router = useRouter();

const { namespace }: Record<string, any> = route.params;

const podListBySort: any = computed(() => podList.value);

function fetchMetrics() {
	const {
		namespace,
		kind,
		name: podName
	}: { [key: string]: any } = route.params;
	const podListData = podList.value ?? [];

	const resources = podListData.map((item: any) => item.name);
	const metrics = Object.values(MetricTypes);
	const test = getTimeRange();
	let param = {
		cluster: 'default',
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

	getMetrics(namespace, param, {
		cancelToken: source.token
	})
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
const clearLocker = () => {
	if (locker) {
		clearTimeout(locker);
	}
};

const fetchMetricsAuto = () => {
	clearLocker();
	locker = setTimeout(() => {
		fetchMetrics();
	}, 3 * 1000);
};

const sortChange = (type: string) => {
	sort_type.value = type;
	fetchData();
};

const fetchData = () => {
	const { namespace }: Record<string, any> = route.params;
	const { node: nodeName, sort: sortBy }: Record<string, any> = route.query;

	const params = {
		namespace,
		name: name.value,
		nodeName
	};

	const metricsParams = {
		sort_metric: selectValue.value.value,
		metrics_filter: selectValue.value.value,
		sort_type: sort_type.value,
		limit: 9999
	};
	loading.value = true;

	Promise.all([
		getNameSpacePodsList(params),
		getMetrics(namespace, metricsParams)
	])
		.then(([res, res2]) => {
			const results = res2.data.results;
			const target = results.find(
				(item) => item.metric_name == selectValue.value.value
			);
			let podsRes: any = res;
			if (target) {
				const podsSort = get(target.data, 'result', []).map((item) =>
					get(item, 'metric.pod')
				);
				const podsFilter = podsSort.filter((item) =>
					podsRes.data.items.find((pod: any) => pod.metadata.name === item)
				);
				podsRes.data.items = podsFilter.map((item) =>
					podsRes.data.items.find((pod: any) => pod.metadata.name === item)
				);
			}
			podList.value = getPosdListFormatter(podsRes);

			fetchMetrics();
		})
		.finally(() => {
			loading.value = false;
		});
};

const change = () => {
	source.cancel();
	autofresh = false;
	fetchData();
};

const itemClick = (data: any) => {
	const { sort: sortBy }: Record<string, any> = route.query;
	usePod.setDetail(data);
	const podTemplateHash = data.labels['pod-template-hash'];
	const name = data.name.split(`-${podTemplateHash}`)[0];
	const path = `/applications/pods/overview/${data.node}/${data.namespace}/${data.name}/${data.createTime}`;
	router.push({
		path,
		query: {
			sort: sortBy
		}
	});
};

watch(
	() => route.params.namespace,
	() => {
		autofresh = false;
		fetchData();
	}
);

onMounted(() => {
	if (route.query.sort) {
		const target = options.find((item) => item.sortBy === route.query.sort);
		if (target) {
			selectValue.value = target;
		}
	}
	fetchData();
});

onBeforeUnmount(() => {
	clearLocker();
});
</script>

<style lang="scss" scoped>
.border-grey-2 {
	border: 1px solid $grey-2;
}
</style>
