<template>
	<div class="q-gutter-sm">
		<my-card flat title="Cluster Resource Usage">
			<MyLoading style="min-height: 289px" :showing="loading[0]">
				<InfoCardRadio
					:list="clusterOptions"
					@change="listchange"
				></InfoCardRadio>
				<MylineChart
					v-if="clusterResource"
					style="margin-top: 24px"
					:key="activeIndex"
					:data="clusterResource[activeIndex]"
				/>
			</MyLoading>
		</my-card>
		<q-space></q-space>
		<my-card flat title="Service Component Monitoring">
			<MyLoading style="min-height: 291px" :showing="loading[1]">
				<InfoCardRadio2
					:list="getTabOptions2"
					@change="listchange2"
				></InfoCardRadio2>
				<MylineChart
					style="margin-top: 24px"
					v-if="clusterResourceData"
					:key="activeIndex2"
					:data="clusterResourceData[activeIndex2]"
				/>
			</MyLoading>
		</my-card>
	</div>
</template>

<script setup lang="ts">
import InfoCardRadio from '@apps/control-hub/components/InfoCard/InfoCardRadio.vue';
import InfoCardRadio2 from '@apps/control-hub/components/InfoCard/InfoCardRadio2.vue';
import { InfoCardItemProps } from '@apps/control-hub/src/components/InfoCard/InfoCardItem.vue';
import MylineChart from '@apps/control-hub/components/Charts/MylineChart.vue';
import MyCard from '@apps/control-hub/src/components/MyCard.vue';
import { date } from 'quasar';
import { ref, reactive, computed, onBeforeUnmount } from 'vue';
import MyLoading from '@apps/control-hub/src/components/MyLoading.vue';
import {
	getClusterMonitoring,
	getComponenthealth,
	getComponentsApiserver,
	getComponentsScheduler
} from '@apps/control-hub/src/network';
import { getContentOptions, getContentOptions2, getTabOptions } from './config';
import {
	getAreaChartOps,
	getResult
} from '@apps/control-hub/src/utils/monitoring';
import { getTabOptions2 } from './config';

const clusterOptions = ref();
const loading = ref([false]);
const clusterResource = ref();
const SchedulerData = ref();
const ApiserverData = ref();
const clusterResourceData = ref();
const activeIndex = ref(0);
const activeIndex2 = ref(0);
const listchange = (data: InfoCardItemProps, index: number) => {
	activeIndex.value = index;
};

const listchange2 = (data: any, index: number) => {
	activeIndex2.value = index;
};

const fetchCluster = (showLoading = true) => {
	const end = date.formatDate(Date.now(), 'X');
	const start = Number(end) - 30000;
	const params = {
		start,
		end,
		step: '300s',
		times: 100,
		metrics_filter:
			'cluster_cpu_usage|cluster_cpu_total|cluster_cpu_utilisation|cluster_memory_usage_wo_cache|cluster_memory_total|cluster_memory_utilisation|cluster_disk_size_usage|cluster_disk_size_capacity|cluster_disk_size_utilisation|cluster_pod_running_count|cluster_pod_quota$'
	};
	if (showLoading) {
		loading.value[0] = true;
	}
	getClusterMonitoring(params)
		.then((res) => {
			const data = getResult(res.data.results);
			clusterOptions.value = getTabOptions(data);
			clusterResource.value = getContentOptions(data).map((item) =>
				getAreaChartOps(item)
			);
		})
		.finally(() => {
			refresh();
			loading.value[0] = false;
		});
};

const fetchClusterResource = (showLoading = true) => {
	const end = date.formatDate(Date.now(), 'X');
	const start = Number(end) - 30000;
	const paramsApi = {
		start,
		end,
		step: '300s',
		times: 100,
		metrics_filter:
			'apiserver_request_latencies|apiserver_request_by_verb_latencies|apiserver_request_rate$'
	};
	const paramsScheduler = {
		start,
		end,
		step: '300s',
		times: 100,
		metrics_filter:
			'scheduler_schedule_attempts|scheduler_schedule_attempt_rate$'
	};

	if (showLoading) {
		loading.value[1] = true;
	}

	Promise.all([
		getComponentsApiserver(paramsApi),
		getComponentsScheduler(paramsScheduler)
	])
		.then(([resApi, resScheduler]) => {
			const data = {
				...getResult(resApi.data.results),
				...getResult(resScheduler.data.results)
			};
			clusterResourceData.value = getContentOptions2(data).map((item) =>
				getAreaChartOps(item)
			);
		})
		.finally(() => {
			refresh2();
			loading.value[1] = false;
		});
};
type Locker = string | number | NodeJS.Timeout | undefined;
let locker: Locker = undefined;
let locker2: Locker = undefined;

const refresh = () => {
	locker && clearTimeout(locker);
	locker = setTimeout(() => {
		fetchCluster(false);
	}, 10000);
};

const refresh2 = () => {
	locker2 && clearTimeout(locker2);
	locker2 = setTimeout(() => {
		fetchClusterResource(false);
	}, 10000);
};

onBeforeUnmount(() => {
	clearTimeout(locker);
	clearTimeout(locker2);
});

fetchCluster();
fetchClusterResource();
</script>

<style></style>
