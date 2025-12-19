<template>
	<FullPageWithBack :title="$t('PODS_DETAILS')">
		<div class="column no-wrap flex-gap-y-xl">
			<div v-for="(item, index) in podsList" :key="index">
				<MyCard>
					<q-responsive
						:ratio="3.5"
						style="min-height: 264px; max-height: 320px"
					>
						<MylineChart :data="item" class="full-height" :loading="loading">
						</MylineChart>
					</q-responsive>
				</MyCard>
			</div>
		</div>
	</FullPageWithBack>
</template>

<script setup lang="ts">
import FullPageWithBack from '@apps/control-panel-common/src/components/FullPageWithBack2.vue';
import MyCard from '@apps/dashboard/components/MyCard.vue';
import { getNodeMonitoring } from '@apps/dashboard/src/network';
import { computed, onBeforeUnmount, onMounted, ref } from 'vue';
import { getMonitoringCfgs, MetricTypes, getPodsList } from './config';
import {
	fillEmptyMetrics,
	getParams
} from '@apps/control-panel-common/src/containers/Monitoring/config';
import {
	getAreaChartOps,
	getResult
} from '@apps/dashboard/src/utils/monitoring';
import MylineChart from '@apps/control-panel-common/src/components/Charts/MylineChart.vue';
import { Locker } from '@apps/dashboard/src/types/main';
import { getRefreshResult } from '@apps/control-panel-common/src/containers/PodsList/config';
import { timeRangeDefault } from '../../../../controlPanelCommon/config/resource.common';

const PodsData = ref({});
const loading = ref(false);

const podsList = computed(() => getPodsList(PodsData.value));

const PodsChartData = computed(() =>
	getAreaChartOps(getMonitoringCfgs(PodsData.value))
);

const fetchData = async (autofresh = false) => {
	let filters = {
		metrics: Object.values(MetricTypes),
		last: false
	};
	if (autofresh) {
		filters.last = true;
	} else {
		filters = { ...filters, ...timeRangeDefault };
		loading.value = true;
	}
	try {
		const params = getParams(filters);
		const res = await getNodeMonitoring(params);
		let result = getResult(res.data.results);

		if (autofresh) {
			result = getRefreshResult(result, PodsData.value);
		}
		PodsData.value = fillEmptyMetrics(params, result);
		refresh();
	} catch (error) {
		loading.value = false;
	}
	loading.value = false;
};

let locker: Locker = undefined;

const clearLocker = () => {
	locker && clearTimeout(locker);
};

const refresh = () => {
	clearLocker();
	locker = setTimeout(() => {
		fetchData(true);
	}, 60 * 1000);
};

onMounted(() => {
	fetchData();
});

onBeforeUnmount(() => {
	clearLocker();
});
</script>
