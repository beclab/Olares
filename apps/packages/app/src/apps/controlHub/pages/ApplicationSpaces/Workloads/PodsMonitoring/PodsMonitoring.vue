<template>
	<div class="q-gutter-y-lg">
		<div class="">
			<div class="text-h6 text-ink-1">{{ $t('BASIC_INFOMATION') }}</div>
			<q-card flat bordered class="q-pa-lg q-mt-lg">
				<DetailData ref="detaiDatalRef"></DetailData>
			</q-card>
		</div>
		<div class="row items-center justify-between">
			<div class="text-h6 text-ink-1">{{ $t('MONITORING') }}</div>
			<div class="row items-center q-col-gutter-x-xs">
				<QButtonStyle>
					<q-btn dense flat icon="refresh" :disable="loading" @click="refresh">
					</q-btn>
				</QButtonStyle>
				<QButtonStyle class="q-mr-md">
					<q-btn
						dense
						flat
						:icon="play ? 'sym_r_pause' : 'sym_r_slideshow'"
						@click="realtimeFetch"
					/>
				</QButtonStyle>
				<DateRangeMonitoring
					v-model="selectValue2"
					:times="selectValue.times"
					:step="selectValue.step"
					@change="selecteChange"
				/>
			</div>
		</div>
		<div v-for="(item, index) in list" :key="`md-${index}`">
			<q-card flat bordered class="q-pa-lg">
				<MylineChart
					:data="item"
					style="height: 238px"
					:loading="loading"
					:split-number-y="3"
				/>
			</q-card>
		</div>
	</div>
	<q-inner-loading :showing="mountLoading"> </q-inner-loading>
</template>

<script setup lang="ts">
import MylineChart from '@apps/control-panel-common/src/components/Charts/MylineChart.vue';
import {
	getWorkloadsMetrics2,
	getWorkloadsPods
} from '@apps/control-hub/src/network';
import DateRangeMonitoring from '@apps/control-panel-common/src/containers/Monitoring/DateRangeMonitoring.vue';
import { computed, ref, onMounted } from 'vue';
import {
	getAreaChartOps,
	getResult
} from '@apps/control-hub/src/utils/monitoring';
import {
	chartConfig as getMonitoringCfgs,
	MetricTypes,
	MONITOR_MODULES,
	initParams
} from './config';
import {
	fillEmptyMetrics,
	getParams
} from '@apps/control-panel-common/src/containers/Monitoring/config';
import QButtonStyle from '@apps/control-panel-common/src/components/QButtonStyle.vue';
import { useRoute } from 'vue-router';
import { getLastTimeStr } from '@apps/control-panel-common/src/containers/Monitoring/utils';
import { get, isEmpty } from 'lodash';
import { getRefreshResult } from '@apps/control-panel-common/src/containers/PodsList/config';
import DetailData from '@apps/control-hub/src/pages/ApplicationSpaces/Workloads/DetailData.vue';
import { timeParams } from '@apps/control-panel-common/src/config/resource.common';

const route = useRoute();
export type DateRangeItem = string;
let locker: undefined | NodeJS.Timeout = undefined;

let _pods = [];
const selectValue = ref(
	initParams({
		...timeParams,
		createTime: route.params.createTime
	})
);

const defaultdata = getMonitoringCfgs([], []).map((item) =>
	getAreaChartOps(item)
);

const selectValue2 = ref<DateRangeItem>(
	getLastTimeStr(selectValue.value.step, selectValue.value.times)
);

const mountLoading = ref(true);
const loading = ref(false);
const list = ref();
const clusterData = ref();
const play = ref(false);

const selecteChange = (value: any) => {
	selectValue.value = value;
	pause();
	fetchData();
};

const fetchData = async (autofresh = false) => {
	const {
		cluster = 'default',
		namespace,
		pods_name: name,
		kind
	}: Record<string, any> = route.params;
	if (isEmpty(_pods)) {
		const {
			data: { results }
		} = await getWorkloadsPods(namespace, MONITOR_MODULES[kind], name, {
			sort_metric: MetricTypes.cpu_usage,
			limit: 5,
			page: 1
		});
		const data = getResult(results);
		const result = get(data[MetricTypes.cpu_usage], 'data.result') || [];
		_pods = result.map((item) => get(item, 'metric.pod'));
	}

	const filters: any = {
		metrics: Object.values(MetricTypes),
		...selectValue.value,
		cluster,
		resources_filter: _pods.join('|'),
		sortBy: 'createTime'
	};

	if (autofresh) {
		filters.last = true;
	} else {
		loading.value = true;
	}

	const params = getParams(filters);

	getWorkloadsMetrics2(namespace, kind, name, params)
		.then((res) => {
			let result = getResult(res.data.results);

			if (autofresh) {
				result = getRefreshResult(result, clusterData.value);
			}

			clusterData.value = fillEmptyMetrics(params, result);
			list.value = getMonitoringCfgs(clusterData.value).map((item) =>
				getAreaChartOps(item)
			);
			autofresh && setLocker();
		})
		.finally(() => {
			loading.value = false;
			mountLoading.value = false;
		});
};

const realtimeFetch = () => {
	play.value = !play.value;
	if (!play.value) {
		clearLocker();
		return;
	}
	fetchData(true);
};
const setLocker = () => {
	clearLocker();
	locker = setTimeout(() => {
		fetchData(true);
	}, 3000);
};

const clearLocker = () => {
	locker && clearTimeout(locker);
};

const refresh = () => {
	pause();
	fetchData();
};

const pause = () => {
	clearLocker();
	play.value = false;
};

onMounted(() => {
	fetchData();
});
</script>

<style lang="scss">
.my-custom-drawer-wrapper {
	&.el-drawer {
		box-shadow: none;
	}
	.my-drawer-title {
		color: #1f1814;
		font-size: 14px;
		font-weight: 500;
		line-height: 20px;
		margin-left: 12px;
	}
	.el-drawer__header {
		margin-bottom: 0;
		padding: 0 18px;
		height: 56px;
	}
}
</style>
