<template>
	<RouteBackBar
		:title="container"
		:sub-title="`${namespace} ${name}`"
	></RouteBackBar>
	<div>
		<bt-scroll-area
			style="height: calc(100vh - 30px)"
			class="my-scroll-container"
		>
			<MyLoading :showing="loading">
				<my-card flat title="Monitoring" animated>
					<template #extra>
						<div class="row items-center">
							<DateRangeMonitoring
								:default-value="selectValue"
								@change="selecteChange"
							/>
							<q-btn
								outline
								padding="4px"
								dense
								color="grey-5"
								style="margin-left: 8px; border-radius: 8px"
								:disable="loading"
							>
								<q-icon
									name="refresh"
									class="refresh-icon"
									@click="init"
								></q-icon>
							</q-btn>
						</div>
					</template>
					<div class="row q-col-gutter-lg">
						<div class="col-12" v-for="item in list" :key="`md-${item.title}`">
							<MylineChart :data="item" />
						</div>
					</div>
				</my-card>
			</MyLoading>
		</bt-scroll-area>
	</div>
</template>

<script setup lang="ts">
import RouteBackBar from '@apps/control-panel-common/src/components/RouteBackBar.vue';
import { useRoute } from 'vue-router';
import MylineChart from '@apps/control-hub/components/Charts/MylineChart.vue';
import MyCard from '@apps/control-panel-common/src/components/MyCard.vue';
import { onMounted, ref } from 'vue';
import { getContainersMonitoring } from '@apps/control-hub/src/network';
import { date } from 'quasar';
import { chartConfig, getResult } from './config';
import DateRangeMonitoring, {
	DateRangeItem,
	options
} from '@apps/control-panel-common/src/containers/DateRangeMonitoring.vue';
import MyLoading from '@apps/control-panel-common/src/components/MyLoading.vue';
import { getAreaChartOps } from '@apps/control-hub/src/utils/monitoring';

const selectValue = ref<DateRangeItem>(options[4]);
let loading = ref(false);
let list: any = ref([]);

const route = useRoute();
const { namespace, name, container }: any = route.params;

const init = () => {
	fetchList();
};

const fetchList = () => {
	const end = date.formatDate(Date.now(), 'X');
	const start = (Number(end) - selectValue.value.value).toString();
	const params = {
		start,
		end,
		cluster: 'default',
		namespace,
		container,
		podName: name,
		step: '600s',
		times: 50,
		metrics_filter: 'container_cpu_usage|container_memory_usage_wo_cache$'
	};

	loading.value = true;
	getContainersMonitoring(params)
		.then((res) => {
			const data = getResult(res.data.results);
			list.value = chartConfig(data).map((item) => getAreaChartOps(item));
		})
		.finally(() => {
			loading.value = false;
		});
};

const selecteChange = (value: DateRangeItem) => {
	selectValue.value = value;
	fetchList();
};

onMounted(() => {
	init();
});
</script>

<style lang="scss" scoped>
.my-scroll-container {
	margin: 8px;
}
</style>
