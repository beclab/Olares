<template>
	<MyLoading :showing="loading">
		<my-card bordered flat title="Monitoring" animated>
			<template #extra>
				<div class="row items-center q-gutter-x-md">
					<DateRangeMonitoring
						:step="selectValue.step"
						:times="selectValue.times"
						@change="selecteChange"
					/>
					<QButtonStyle>
						<q-btn outline dense :disable="loading">
							<q-icon name="refresh" @click="init"></q-icon>
						</q-btn>
					</QButtonStyle>
				</div>
			</template>
			<div class="row q-col-gutter-lg">
				<div class="col-12" v-for="item in list" :key="`md-${item.title}`">
					<MylineChart :data="item" style="height: 238px" :split-number-y="3" />
				</div>
			</div>
		</my-card>
	</MyLoading>
</template>

<script setup lang="ts">
import { useRoute } from 'vue-router';
import MylineChart from '../../components/Charts/MylineChart.vue';
import MyCard from '../../components/MyCard2.vue';
import { onMounted, ref } from 'vue';
import { getContainersMonitoring } from '@apps/control-panel-common/src/network';
import { chartConfig, getResult, MetricTypes } from './config';
import DateRangeMonitoring from './DateRangeMonitoring.vue';
import MyLoading from '../../components/MyLoading.vue';
import { getAreaChartOps } from '@apps/control-panel-common/src/utils/monitoring';
import { fillEmptyMetrics, getParams, ParamsFormat } from './config';
import QButtonStyle from '../../components/QButtonStyle.vue';

interface Props {
	createTime?: string;
	container?: string;
}

const props = withDefaults(defineProps<Props>(), {});

const selectValue = ref({
	step: '600s',
	times: 50
});
let loading = ref(false);
let list: any = ref([]);

const route = useRoute();

const init = () => {
	fetchList();
};

const fetchList = () => {
	const { namespace, name, container }: any = route.params;
	const containerName = container || props.container;
	const filters: any = {
		cluster: 'default',
		namespace,
		container: containerName,
		podName: name,
		metrics: Object.values(MetricTypes),
		...selectValue.value
	};
	const params = getParams(filters);

	const fillZero = true;
	loading.value = true;
	getContainersMonitoring(params)
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

<style lang="scss" scoped>
.my-scroll-container {
	margin: 8px;
}
</style>
