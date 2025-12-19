<template>
	<FullPageWithBack :title="$t('PHYSICAL_RESOURCE_MONTORING')">
		<template #extra>
			<DateRangeMonitoring
				v-model="selectValue2"
				:times="selectValue.times"
				:step="selectValue.step"
				@change="selecteChange"
			/>
			<QButtonStyle>
				<q-btn
					class="q-pa-xs"
					dense
					icon="refresh"
					color="ink-2"
					outline
					style="margin-left: 16px; border-radius: 8px"
					:disable="loading"
					@click="fetchData"
				>
				</q-btn>
			</QButtonStyle>
		</template>
		<MyGridLayout col-width="542px" gap="xl">
			<div
				class="col-12 col-md-6"
				v-for="(item, index) in list"
				:key="`md-${index}`"
			>
				<MyCard>
					<MylineChart :data="item" style="height: 236px" :loading="loading" />
				</MyCard>
			</div>
		</MyGridLayout>
	</FullPageWithBack>
</template>

<script setup lang="ts">
import MylineChart from '@apps/control-panel-common/src/components/Charts/MylineChart.vue';
import { getClusterMonitoring } from '@apps/dashboard/src/network';
import DateRangeMonitoring from '@apps/control-panel-common/src/containers/Monitoring/DateRangeMonitoring.vue';
import { ref } from 'vue';
import {
	getAreaChartOps,
	getResult
} from '@apps/dashboard/src/utils/monitoring';
import { getMonitoringCfgs, MetricTypes } from './config';
import MyCard from '@apps/dashboard/components/MyCard.vue';
import {
	fillEmptyMetrics,
	getParams
} from '@apps/control-panel-common/src/containers/Monitoring/config';
import FullPageWithBack from '@apps/control-panel-common/src/components/FullPageWithBack2.vue';
import MyGridLayout from '@apps/control-panel-common/src/components/MyGridLayout.vue';
import QButtonStyle from '@apps/control-panel-common/src/components/QButtonStyle.vue';
import { getLastTimeStr } from '@apps/control-panel-common/src/containers/Monitoring/utils';
import { timeParams } from '@apps/control-panel-common/src/config/resource.common';
export type DateRangeItem = string;

const defaultdata = getMonitoringCfgs([]).map((item) => getAreaChartOps(item));

const selectValue = ref(timeParams);

const selectValue2 = ref<DateRangeItem>(
	getLastTimeStr(selectValue.value.step, selectValue.value.times)
);

const loading = ref(false);
const list = ref(defaultdata);

const selecteChange = (value: any) => {
	selectValue.value = value;
	fetchData();
};

const fetchData = () => {
	const filters: any = {
		metrics: Object.values(MetricTypes),
		...selectValue.value
	};

	const params = getParams(filters);
	const fillZero = true;
	loading.value = true;
	getClusterMonitoring(params)
		.then((res) => {
			const result = getResult(res.data.results);
			const data = fillZero ? fillEmptyMetrics(params, result) : result;
			list.value = getMonitoringCfgs(data).map((item) => getAreaChartOps(item));
		})
		.finally(() => {
			loading.value = false;
		});
};

fetchData();
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
