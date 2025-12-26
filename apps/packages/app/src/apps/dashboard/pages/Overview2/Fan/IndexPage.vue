<template>
	<FullPageWithBack :title="$t('FAN_DETAILS')">
		<div class="row flex-gap-xl no-wrap">
			<div style="flex: 1" v-for="(item, index) in list" :key="index">
				<MyCard :title="item.title" class="full-height">
					<div class="row flex-gap-xxl no-wrap">
						<div class="fan-gua-wrapper">
							<MyGaugeChart
								detailLabelWrap
								style="width: 100%"
								:progressWidth="10"
								:gaugeMax="item.FanSpeedMax"
								:data="item.gaugeChartData"
								:colorStops="item.colorStops"
							></MyGaugeChart>
						</div>
						<div style="flex: 1" class="row">
							<div class="column flex-gap-md" style="flex: 1">
								<div
									v-for="(child, childIndex) in item.contentList"
									:key="child.title"
									:style="{ opacity: child.opacity }"
								>
									<div class="text-body3 text-ink-1">
										{{ child.title }}
									</div>
									<div class="relative-position q-mt-xs fan-progress-container">
										<div class="fan-progress-wrapper">
											<MyProgressBar
												:colorStops="child.colorStops"
												:data="child.ProgressChartData"
											></MyProgressBar>
										</div>
									</div>
								</div>
							</div>

							<div class="column flex-gap-md">
								<div
									class="q-ml-md text-body3 text-ink-1 text-right q-mt-lg row items-center justify-end"
									v-for="(child, childIndex) in item.contentList"
									:key="child.title"
									:style="{ opacity: child.opacity }"
								>
									<span>{{ child.value }}</span>
									<span class="q-ml-xs">{{ child.unit }}</span>
								</div>
							</div>
						</div>
					</div>
				</MyCard>
			</div>
		</div>

		<div class="q-mt-xl">
			<QTableStyle2 sticky-last>
				<q-table
					style="width: 100%"
					:rows="tableData"
					:columns="columns"
					:loading="loading"
					hide-pagination
					v-model:pagination="pagination"
					flat
					class="table-wrapper"
				>
				</q-table>
			</QTableStyle2>
		</div>
		<!-- <q-inner-loading :showing="loading"> </q-inner-loading>
		<Empty
			class="absolute-center"
			v-show="!loading && fanData.length == 0"
			@click="fetchData"
		></Empty> -->
	</FullPageWithBack>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import FullPageWithBack from '@apps/control-panel-common/src/components/FullPageWithBack2.vue';
import Empty from '@apps/control-panel-common/components/Empty3.vue';
import MyProgressBar from '@apps/control-panel-common/components/Charts/MyProgressBar.vue';
import MyCard from '@apps/dashboard/components/MyCard.vue';
import MyGaugeChart from '@apps/dashboard/components/Charts/MyGaugeChart.vue';
import QTableStyle2 from '@apps/control-panel-common/src/components/QTableStyle2.vue';
import { useI18n } from 'vue-i18n';
import {
	tableData,
	columns,
	FanSpeedMaxCPU,
	FanSpeedMaxGPU,
	gpuFanColorStops,
	cpuFanColorStops
} from './config';
import { useFanStore } from '@apps/dashboard/stores/Fan';
const { t } = useI18n();
const FanStore = useFanStore();

const loading = ref(false);
const chartData = computed(() => ({
	legend: ['rpm'],
	data: [[FanStore.data.cpu_fan_speed]],
	unit: 'RPM'
}));
const chartData2 = computed(() => ({
	legend: ['Used'],
	data: [[FanStore.data.gpu_fan_speed]],
	unit: 'RPM'
}));

const pagination = ref({
	rowsNumber: 0
});

const list = computed(() => {
	return [
		{
			title: t('CPU'),
			gaugeChartData: chartData.value,
			FanSpeedMax: FanSpeedMaxCPU,
			colorStops: cpuFanColorStops,
			contentList: [
				{
					title: t('DISK_OP.TEMPERATURE'),
					value: FanStore.data.cpu_temperature,
					unit: '°C',
					colorStops: [0, 0.6, 0.8],
					ProgressChartData: {
						legend: ['Used'],
						data: [[FanStore.data.cpu_temperature / 100]],
						unit: 'RPM'
					}
				},
				{
					title: t('FAN_OP.SPEED'),
					value: FanStore.data.cpu_fan_speed,
					unit: 'RPM',
					colorStops: cpuFanColorStops,
					ProgressChartData: {
						legend: ['Used'],
						data: [[FanStore.data.cpu_fan_speed / FanSpeedMaxCPU]],
						unit: 'RPM'
					}
				}
			]
		},
		{
			title: t('GPU_OP.GPU'),
			gaugeChartData: chartData2.value,
			FanSpeedMax: FanSpeedMaxGPU,
			colorStops: gpuFanColorStops,
			contentList: [
				{
					title: t('DISK_OP.TEMPERATURE'),
					value: FanStore.data.gpu_temperature,
					unit: '°C',
					colorStops: [0, 0.6, 0.8],
					ProgressChartData: {
						legend: ['Used'],
						data: [[FanStore.data.gpu_temperature / 100]],
						unit: 'RPM'
					}
				},
				{
					title: t('FAN_OP.POWER'),
					value: FanStore.data.gpu_power,
					unit: 'W',
					colorStops: [0, 0.6, 0.8],
					ProgressChartData: {
						legend: ['Used'],
						data: [[FanStore.data.gpu_power / FanStore.data.gpu_power_limit]],
						unit: 'W'
					}
				},
				{
					title: t('FAN_OP.SPEED'),
					value: FanStore.data.gpu_fan_speed,
					unit: 'RPM',
					colorStops: gpuFanColorStops,
					ProgressChartData: {
						legend: ['Used'],
						data: [[FanStore.data.gpu_fan_speed / FanSpeedMaxGPU]],
						unit: 'RPM'
					}
				}
			]
		}
	];
});
const fetchData = () => {
	//
};

onMounted(() => {
	fetchData();
});
</script>

<style lang="scss" scoped>
.fan-gua-wrapper {
	flex: 0 0 140px;
}
.fan-progress-container {
	height: 16px;
	.fan-progress-wrapper {
		position: absolute;
		top: 50%;
		right: 0px;
		left: 0px;
		transform: translateY(-50%);
	}
}
</style>
