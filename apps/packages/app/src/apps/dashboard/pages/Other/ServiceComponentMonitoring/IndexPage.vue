<template>
	<div class="">
		<my-card flat title="Service Component Monitoring">
			<MyLoading style="min-height: 268px" :showing="loading">
				<div class="row q-gutter-x-lg">
					<InfoCardRadio2
						:list="getTabOptions2"
						@change="listchange2"
					></InfoCardRadio2>
					<MylineChart
						style="flex: 1; height: 268px"
						v-if="clusterResourceData"
						:key="activeIndex2"
						:data="clusterResourceData[activeIndex2]"
					/>
				</div>
			</MyLoading>
		</my-card>
	</div>
</template>

<script setup lang="ts">
import InfoCardRadio2 from '@apps/dashboard/components/InfoCard/InfoCardRadio2.vue';
import MylineChart from '@apps/control-panel-common/src/components/Charts/MylineChart.vue';
import MyCard from '@apps/control-panel-common/src/components/MyCard.vue';
import { ref, onBeforeUnmount } from 'vue';
import MyLoading from '@apps/control-panel-common/src/components/MyLoading.vue';
import {
	getComponentsApiserver,
	getComponentsScheduler
} from '@apps/dashboard/src/network';
import {
	ApiMetricTypes,
	getContentOptions2,
	SchedulerMetricTypes
} from './config';
import {
	getAreaChartOps,
	getResult
} from '@apps/dashboard/src/utils/monitoring';
import { getTabOptions2 } from './config';
import {
	fillEmptyMetrics,
	getParams
} from '@apps/control-panel-common/src/containers/Monitoring/config';
import { getRefreshResult } from '@apps/control-panel-common/src/containers/PodsList/config';

const loading = ref(false);
const clusterResourceData = ref();
const activeIndex2 = ref(0);
const clusterData = ref([]);

const listchange2 = (data: any, index: number) => {
	activeIndex2.value = index;
};

const fetchClusterResource = (showLoading = true, autofresh = false) => {
	const timeRange = {
		step: '5m',
		times: 100
	};

	const filters1: any = {
		metrics: Object.values(ApiMetricTypes),
		...timeRange
	};

	const filters2: any = {
		metrics: Object.values(SchedulerMetricTypes),
		...timeRange
	};

	if (autofresh) {
		filters1.last = true;
		filters2.last = true;
	}

	const paramsApi = getParams(filters1);
	const paramsScheduler = getParams(filters2);

	if (showLoading) {
		loading.value = true;
	}

	Promise.all([
		getComponentsApiserver(paramsApi),
		getComponentsScheduler(paramsScheduler)
	])
		.then(([resApi, resScheduler]) => {
			let data = {
				...getResult(resApi.data.results),
				...getResult(resScheduler.data.results)
			};
			const params = { ...paramsApi, ...paramsScheduler };
			let clusterResult = data;
			if (autofresh) {
				clusterResult = getRefreshResult(clusterResult, clusterData.value);
			}
			clusterData.value = fillEmptyMetrics(params, clusterResult);

			clusterResourceData.value = getContentOptions2(clusterData.value).map(
				(item) => getAreaChartOps(item)
			);
		})
		.finally(() => {
			refresh2();
			loading.value = false;
		});
};
type Locker = string | number | NodeJS.Timeout | undefined;
let locker2: Locker = undefined;

const refresh2 = () => {
	locker2 && clearTimeout(locker2);
	locker2 = setTimeout(() => {
		fetchClusterResource(false, true);
	}, 10000);
};

onBeforeUnmount(() => {
	clearTimeout(locker2);
});

fetchClusterResource();
</script>

<style></style>
