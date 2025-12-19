<template>
	<FullPageWithBack :title="node">
		<template #extra>
			<div class="row q-gutter-md">
				<DateRangeMonitoring
					:step="selectValue.step"
					:times="selectValue.times"
					@change="selecteChange"
				/>
				<q-btn
					outline
					padding="4px"
					dense
					flat
					color="grey-5"
					:disable="loading"
				>
					<q-icon name="refresh" class="refresh-icon" @click="init"></q-icon>
				</q-btn>
			</div>
		</template>
		<div class="row q-col-gutter-x-lg q-col-gutter-y-xl">
			<div
				class="col-12 col-md-6"
				v-for="item in list"
				:key="`md-${item.title}`"
			>
				<MylineChart :data="item" />
			</div>
		</div>
		<q-inner-loading :showing="loading"> </q-inner-loading>
	</FullPageWithBack>
</template>

<script setup lang="ts">
import { useRoute } from 'vue-router';
import MylineChart from '@apps/control-panel-common/src/components/Charts/MylineChart.vue';
import { onMounted, ref } from 'vue';
import { chartConfig, getResult, MetricTypes } from './config';
import {
	fillEmptyMetrics,
	getParams
} from '@apps/control-panel-common/src/containers/Monitoring/config';
import DateRangeMonitoring from '@apps/control-panel-common/src/containers/Monitoring/DateRangeMonitoring.vue';
import { getAreaChartOps } from '@apps/dashboard/src/utils/monitoring';
import { getNodeMonitoring } from '@apps/control-panel-common/src/network';
import FullPageWithBack from '@apps/control-panel-common/src/components/FullPageWithBack.vue';

const selectValue = ref({
	times: 50,
	step: '2m'
});
let loading = ref(false);
let list: any = ref([]);

const route = useRoute();
const { node }: any = route.params;

const init = () => {
	fetchList();
};

const fetchList = () => {
	const filters: any = {
		resources: [node],
		metrics: Object.values(MetricTypes),
		...selectValue.value
	};

	const params = getParams(filters);
	// fillZero: !role.includes('edge'),
	const fillZero = true;
	loading.value = true;
	getNodeMonitoring(params)
		.then((res) => {
			const result = getResult(res.data);
			const data = fillZero ? fillEmptyMetrics(params, result) : result;
			list.value = chartConfig(data).map((item) => getAreaChartOps(item));
		})
		.finally(() => {
			loading.value = false;
		});
};
const selecteChange = (value: any) => {
	selectValue.value = value;
	fetchList();
};

onMounted(() => {
	init();
});
</script>
