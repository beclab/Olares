<template>
	<RouteBackBar :title="name"></RouteBackBar>
	<div>
		<bt-scroll-area
			style="height: calc(100vh - 64px)"
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
import { getPodMonitoring } from '@apps/control-hub/src/network';
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
const { namespace, name }: any = route.params;

const init = () => {
	fetchList();
};

const fetchList = () => {
	const end = date.formatDate(Date.now(), 'X');
	const start = (Number(end) - selectValue.value.value).toString();
	const params = {
		namespace,
		podName: name,
		cluster: 'default',
		start,
		end,
		step: '600s',
		times: 50,
		metrics_filter:
			'pod_cpu_usage|pod_memory_usage_wo_cache|pod_net_bytes_transmitted|pod_net_bytes_received$'
	};

	loading.value = true;
	getPodMonitoring(params)
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
