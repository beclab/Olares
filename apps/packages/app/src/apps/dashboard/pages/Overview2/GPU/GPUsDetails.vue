<template>
	<FullPageWithBack :title="$t('GPU_OP.GRAPHICS_MANAGEMENT_DETAILS')">
		<template #extra>
			<DatePicker
				v-model="times"
				:disabled="loading"
				@update:modelValue="fetchLineData"
			></DatePicker>
		</template>
		<div class="column flex-gap-xl">
			<MyCard>
				<div class="text-h6 text-ink-1 q-pb-xl">
					{{ $t('GPU_OP.DETAILS_INFORMATION') }}
				</div>
				<div class="row q-col-gutter-xl">
					<GPUDetailList :data="details" :columns="columns">
						<template #field-health="{ data }">
							<GPUStatus
								:isExternal="data.isExternal"
								:health="data.health"
							></GPUStatus>
						</template>
						<template #field-mode="{ value, data }">
							{{ data.type?.split('-')[0] === 'NVIDIA' ? value : 'default' }}
						</template>
					</GPUDetailList>
				</div>
			</MyCard>
			<MyCard>
				<div class="row justify-between flex-gap-y-lg">
					<div v-for="(item, index) in gaugeConfigShow" :key="index">
						<div class="column items-center justify-start">
							<MyGaugeChart
								:data="{
									title: item.title,
									unit: item.gaugeUnit,
									data: [[round(item.percent, 2)]]
								}"
							></MyGaugeChart>
						</div>
					</div>
					<div v-for="(item, index) in lineTools" :key="index">
						<MyGaugeChart
							:data="{
								title: item.title,
								unit: item.gaugeUnit,
								data: [[round(item.percent, 2)]]
							}"
						></MyGaugeChart>
					</div>
				</div>
			</MyCard>
		</div>
		<div class="q-mt-xl">
			<MyGridLayout col-width="540px" gap="xl">
				<MyCard v-for="item in trendList" :key="item.title">
					<MylineChart
						:data="item"
						:splitNumberY="4"
						:loading="item.loading"
						style="height: 234px"
					>
					</MylineChart>
				</MyCard>
			</MyGridLayout>
		</div>
	</FullPageWithBack>
</template>

<script setup lang="ts">
import FullPageWithBack from '@apps/control-panel-common/src/components/FullPageWithBack2.vue';
import MyCard from '@apps/dashboard/components/MyCard.vue';
import {
	getGraphicsDetails,
	getInstantVector,
	getRangeVector
} from '@apps/dashboard/src/network/gpu';
import { GraphicsDetailsResponse } from '@apps/dashboard/src/types/gpu';
import { computed, onMounted, ref, watch } from 'vue';
import { useRoute } from 'vue-router';
import { getStepWithTimeRange, useInstantVector } from './config';
import { timeParse } from '@apps/dashboard/src/utils/gpu';
import { useI18n } from 'vue-i18n';
import MyGridLayout from '@apps/control-panel-common/src/components/MyGridLayout.vue';
import MyGaugeChart from '@apps/dashboard/src/components/Charts/MyGaugeChart.vue';
import MylineChart from '@apps/control-panel-common/src/components/Charts/MylineChart6.vue';
import GPUDetailList from './GPUDetailList.vue';
import GPUStatus from '@apps/dashboard/src/pages/Overview2/GPU/GPUStatus.vue';
import DatePicker from './DatePicker.vue';
import { get, round } from 'lodash';
import { date } from 'quasar';

const end = new Date();
const start = new Date();
start.setTime(start.getTime() - 8 * 3600 * 1000);

const times = ref([
	date.formatDate(start, 'YYYY-MM-DD HH:mm:ss'),
	date.formatDate(end, 'YYYY-MM-DD HH:mm:ss')
]);
const router = useRoute();
const route = useRoute();
const { t } = useI18n();
const detail1 = ref<GraphicsDetailsResponse>();
const detail2 = ref({});
const loading = computed(() => {
	const loading1 = gaugeConfig.value.some((item) => item.loading);
	const loading2 = trendList.value.some((item) => item.loading);
	return loading1 || loading2;
});

const columns = [
	{
		label: t('GPU_OP.GRAPHICS_STATUS'),
		field: 'health'
	},
	{
		label: t('GPU_OP.GRAPHICS_ID'),
		field: 'uuid'
	},
	{
		label: t('GPU_OP.AFFILIATED_NODE'),
		field: 'nodeName'
	},
	{
		label: t('GPU_OP.GRAPHICS_MODEL'),
		field: 'type'
	},
	{
		label: t('GPU_OP.DEVICE_NUMBER'),
		field: 'device_no'
	},
	{
		label: t('GPU_OP.DRIVER_VERSION'),
		field: 'driver_version'
	}
];

const details = computed(() => {
	const data = { ...detail2.value, ...detail1.value };
	return data;
});

const fetchDetails = async () => {
	const params = {
		uid: router.params.uuid as string
	};
	const res = await getGraphicsDetails(params);
	detail1.value = res.data;
};

const gaugeConfigShow = computed(() => gaugeConfig.value.slice(2));

const gaugeConfig = useInstantVector(
	[
		{
			title: t('GPU_OP.CALCULATION_POWER_ALLOCATION_RATIO'),
			percent: 0,
			query:
				'avg(sum(hami_container_vcore_allocated{deviceuuid=~"$deviceuuid"}) by (instance))',
			totalQuery:
				'avg(sum(hami_core_size{deviceuuid=~"$deviceuuid"}) by (instance))',
			percentQuery:
				'avg(sum(hami_container_vcore_allocated{deviceuuid=~"$deviceuuid"}) by (instance))/avg(sum(hami_core_size{deviceuuid=~"$deviceuuid"}) by (instance)) *100',
			total: 0,
			used: 0,
			unit: ' ',
			data: [],
			loading: false
		},
		{
			title: t('GPU_OP.VIDEO_MEMORY_ALLOCATION_RATIO'),
			percent: 0,
			query:
				'avg(sum(hami_container_vmemory_allocated{deviceuuid=~"$deviceuuid"}) by (instance)) / 1024',
			totalQuery:
				'avg(sum(hami_memory_size{deviceuuid=~"$deviceuuid"}) by (instance)) / 1024',
			percentQuery:
				'(avg(sum(hami_container_vmemory_allocated{deviceuuid=~"$deviceuuid"}) by (instance)) / 1024 )/(avg(sum(hami_memory_size{deviceuuid=~"$deviceuuid"}) by (instance)) / 1024) *100 ',
			total: 0,
			used: 0,
			unit: 'Gi',
			data: [],
			loading: false
		},
		{
			title: t('GPU_OP.CPU_R'),
			percent: 0,
			query:
				'avg(sum(hami_core_util{deviceuuid=~"$deviceuuid"}) by (instance))',
			percentQuery:
				'avg(sum(hami_core_util_avg{deviceuuid=~"$deviceuuid"}) by (instance))',
			total: 100,
			used: 0,
			unit: ' ',
			data: [],
			loading: false
		},
		{
			title: t('GPU_OP.VRAM_USAGE_RATE'),
			percent: 0,
			query:
				'avg(sum(hami_memory_used{deviceuuid=~"$deviceuuid"}) by (instance)) / 1024',
			totalQuery:
				'avg(sum(hami_memory_size{deviceuuid=~"$deviceuuid"}) by (instance))/1024',
			percentQuery:
				'(avg(sum(hami_memory_used{deviceuuid=~"$deviceuuid"}) by (instance)) / 1024)/(avg(sum(hami_memory_size{deviceuuid=~"$deviceuuid"}) by (instance))/1024)*100',
			total: 0,
			used: 0,
			unit: 'Gi',
			data: [],
			loading: false
		}
	],
	(query) => query.replaceAll('$deviceuuid', route.params.uuid),
	times
);

const trendList = computed(() => [
	{
		title: t('GPU_OP.RESOURCE_ALLOCATION_TREND'),
		unit: '%',
		legend: [t('core'), t('MEMORY')],
		data: [gaugeConfig.value[0].data, gaugeConfig.value[1].data],
		loading: gaugeConfig.value[0].loading || gaugeConfig.value[1].loading
	},
	{
		title: t('GPU_OP.RESOURCE_USAGE_TREND'),
		unit: '%',
		legend: [t('core'), t('MEMORY')],
		data: [gaugeConfig.value[2].data, gaugeConfig.value[3].data],
		loading: gaugeConfig.value[2].loading || gaugeConfig.value[3].loading
	},
	...lineTools.value.map((item) => ({
		title: item.title,
		unit: item.unit,
		legend: [item.title],
		data: item.data,
		loading: item.loading
	}))
]);

const lineTools = ref([
	{
		title: t('GPU_OP.GPU_POWER'),
		query:
			'avg by (device_no,driver_version) (hami_device_power{deviceuuid=~"$deviceuuid"})',
		data: [[]],
		unit: 'W',
		gaugeUnit: 'W',
		percent: 0,
		total: 0,
		hideInfo: true,
		loading: false
	},
	{
		title: t('GPU_OP.GPU_TEMP'),
		query:
			'avg(sum(hami_device_temperature{deviceuuid=~"$deviceuuid"}) by (instance))',
		data: [[]],
		unit: '℃',
		gaugeUnit: '℃',
		percent: 0,
		total: 0,
		hideInfo: true,
		loading: false
	}
]);

const fetchLineData = async () => {
	lineTools.value.forEach((item, index) => {
		lineTools.value[index].loading = true;
		const range = {
			start: timeParse(new Date(times.value[0])),
			end: timeParse(new Date(times.value[1])),
			step: getStepWithTimeRange(times.value)
		};
		getRangeVector({
			range,
			query: item.query.replace('$deviceuuid', route.params.uuid as string)
		})
			.then((res) => {
				const { device_no, driver_version } = res.data.data[0].metric;
				if (device_no && driver_version) {
					detail2.value = { ...detail2.value, device_no, driver_version };
				}
				const data = get(res, 'data.data[0].values', []);
				const list = data.map((item) => [item.timestamp, round(item.value, 2)]);
				lineTools.value[index].data = [list];
			})
			.finally(() => {
				lineTools.value[index].loading = false;
			})
			.catch(() => {
				lineTools.value[index].loading = false;
			});

		getInstantVector({
			query: item.query.replace('$deviceuuid', route.params.uuid as string)
		}).then((res) => {
			lineTools.value[index].percent = res.data.data[0]?.value || 0;
		});
	});
};

onMounted(() => {
	fetchDetails();
	fetchLineData();
});
</script>

<style lang="scss" scoped></style>
