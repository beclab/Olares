<template>
	<FullPageWithBack :title="$t('GPU_OP.TASK_MANAGEMENT_DETAILS')">
		<template #extra>
			<DatePicker v-model="times"></DatePicker>
		</template>
		<MyCard>
			<div class="text-h6 text-ink-1 q-pb-xl">
				{{ $t('GPU_OP.DETAILS_INFORMATION') }}
			</div>
			<div class="row items-center wrap flex-gap-y-xxxxl flex-gap-x-xl">
				<div>
					<GPUDetailList :data="detail" :columns="columns" :maxColumn="2">
						<template #field-status="{ value }">
							<TaskStatus :status="value"></TaskStatus>
						</template>
						<template #field-deviceIds="{ value }">
							<span v-if="!value">-</span>
							<span v-else>
								{{ value.join(',') }}
								<q-tooltip
									v-if="value.length > 0"
									anchor="top middle"
									self="bottom middle"
								>
									<div v-for="item in value" :key="item">{{ item }}</div>
								</q-tooltip>
							</span>
						</template>
						<template #field-nodeName="{ value }">
							<TextPlus :text="value" copy />
						</template>

						<template #field-allocatedMem="{ value }">
							<span v-if="value">{{ roundToDecimal(value / 1024, 2) }} Gi</span>
							<span v-else>--</span>
						</template>

						<template #field-createTime="{ value }">
							{{ timeParse(value) }}
						</template>
					</GPUDetailList>
				</div>
				<div class="row no-wrap justify-end flex-gap-y-xxxxl flex-gap-x-xl">
					<div v-for="(item, index) in gaugeConfig" :key="index">
						<div class="column items-center justify-start">
							<MyGaugeChart
								style="width: 140px"
								:data="{
									title: item.title,
									unit: item.gaugeUnit,
									data: [[roundToDecimal(item.percent)]]
								}"
							></MyGaugeChart>
							<div class="text-body3 text-ink-1 q-mt-sm">
								<span class="">{{
									item.title.includes('使用')
										? $t('GPU_OP.USE')
										: $t('GPU_OP.ALLOCATION')
								}}</span>
								<span v-if="!item.title.includes('算力')"
									>({{ item.unit }})</span
								>
								:
								<span class="text-subtitle3"
									>{{ roundToDecimal(item.used) }}/{{
										roundToDecimal(item.total)
									}}</span
								>
							</div>
						</div>
					</div>
				</div>
			</div>
		</MyCard>

		<div class="q-mt-xl">
			<MyGridLayout col-width="540px" gap="xl">
				<div v-for="(item, index) in lineConfig" :key="index">
					<MyCard>
						<MylineChart
							:data="{
								title: item.title,
								unit: item.unit,
								legend: item.legend,
								data: item.data
							}"
							:loading="item.loading"
							style="height: 234px"
						></MylineChart>
					</MyCard>
				</div>
			</MyGridLayout>
		</div>
	</FullPageWithBack>
</template>

<script setup lang="ts">
import FullPageWithBack from '@apps/control-panel-common/src/components/FullPageWithBack2.vue';
import { getRangeVector, getTaskDetail } from '@apps/dashboard/src/network/gpu';
import { ref, onMounted, watch, computed } from 'vue';
import { useRoute } from 'vue-router';
import { useInstantVector, getStepWithTimeRange } from './config';
import MyGaugeChart from '@apps/dashboard/src/components/Charts/MyGaugeChart.vue';
import GPUDetailList from './GPUDetailList.vue';
import MyCard from '@apps/dashboard/components/MyCard.vue';
import { roundToDecimal, timeParse } from '@apps/dashboard/src/utils/gpu';
import MylineChart from '@apps/control-panel-common/src/components/Charts/MylineChart6.vue';
import TextPlus from '@apps/dashboard/src/components/TextPlus.vue';
import TaskStatus from './TaskStatus.vue';
import MyGridLayout from '@apps/control-panel-common/src/components/MyGridLayout.vue';
import { useI18n } from 'vue-i18n';
import DatePicker from './DatePicker.vue';
import { get, round } from 'lodash';
import { date } from 'quasar';
import { VRAMMode } from 'src/constant';
import { ShareMode } from '@apps/dashboard/src/types/gpu';

const route = useRoute();
const { t } = useI18n();
const { name, pod_uid } = route.params as { [key: string]: string };
const { sharemode } = route.query as { sharemode: ShareMode };
const displayAllocation = computed(() => sharemode !== VRAMMode.TimeSlicing);

const end = new Date();
const start = new Date();
start.setTime(start.getTime() - 3600 * 1000);

const times = ref([
	date.formatDate(start, 'YYYY-MM-DD HH:mm:ss'),
	date.formatDate(end, 'YYYY-MM-DD HH:mm:ss')
]);
const detail = ref();

const columns = computed(() => {
	const list = displayAllocation.value
		? [
				{
					label: t('GPU_OP.ALLOCATABLE_COMPUTING_POWER'),
					field: 'allocatedCores'
				},
				{
					label: t('GPU_OP.ALLOCATABLE_MEMORY'),
					field: 'allocatedMem'
				}
		  ]
		: [];
	return [
		{
			label: t('GPU_OP.TASK_STATUS'),
			field: 'status'
		},
		{
			label: t('GPU_OP.GRAPHICS_CARD_BELONGS'),
			field: 'deviceIds'
		},
		{
			label: t('GPU_OP.AFFILIATED_NODE'),
			field: 'nodeName'
		},
		{
			label: t('GPU_OP.GRAPHICS_TYPE'),
			field: 'type'
		},
		...list,
		{
			label: t('GPU_OP.APP_NAME'),
			field: 'appName'
		},
		{
			label: t('GPU_OP.TASK_CREATION_TIME'),
			field: 'createTime'
		}
	];
});

const gaugeConfig = useInstantVector(
	[
		{
			title: t('GPU_OP.CPU_R'),
			percent: 0,
			query:
				'avg(sum(hami_container_core_used{container_name="$container",pod_name=~"$pod",namespace_name="$namespace"}) by (instance))',
			totalQuery:
				'avg(sum(hami_container_vcore_allocated{container_name="$container",pod_name=~"$pod",namespace_name="$namespace"}) by (instance))',
			percentQuery:
				'avg(sum(hami_container_core_used{container_name="$container",pod_name=~"$pod",namespace_name="$namespace"}) by (instance)) / avg(sum(hami_container_vcore_allocated{container_name="$container",pod_name=~"$pod",namespace_name="$namespace"}) by (instance)) *100',
			total: 0,
			used: 0,
			unit: '%',
			data: []
		},
		{
			title: t('GPU_OP.VRAM_USAGE_RATE'),
			percent: 0,
			query:
				'avg(sum(hami_container_memory_used{container_name="$container",pod_name=~"$pod",namespace_name="$namespace"}) by (instance))/ 1024',
			totalQuery:
				'avg(sum(hami_container_vmemory_allocated{container_name="$container",pod_name=~"$pod",namespace_name="$namespace"}) by (instance))/1024',
			percentQuery:
				'(avg(sum(hami_container_memory_used{container_name="$container",pod_name=~"$pod",namespace_name="$namespace"})/ 1024)/(avg(sum(hami_container_vmemory_allocated{container_name="$container",pod_name=~"$pod",namespace_name="$namespace"}) by (instance))/1024) *100)',
			total: 0,
			used: 0,
			unit: 'GiB',
			data: []
		}
	],
	(query) =>
		query
			.replaceAll('$container', detail.value.name)
			.replaceAll('$namespace', detail.value.namespace)
			.replaceAll('$pod', detail.value.appName)
);

const lineConfig = ref([
	{
		title: t('GPU_OP.COMPUTE_POWER_USAGE_TREND'),
		query:
			'avg(sum(hami_container_core_util{container_name=~"$container",pod_name=~"$pod",namespace_name="$namespace"}) by (instance))',
		unit: '%',
		legend: [t('USAGE')],
		data: [[]],
		loading: false
	},
	{
		title: t('GPU_OP.VRAM_USAGE_TREND'),
		query:
			'avg(sum(hami_container_memory_util{container_name=~"$container",pod_name=~"$pod",namespace_name="$namespace"}) by (instance))',
		unit: '%',
		legend: [t('USAGE')],
		data: [[]],
		loading: false
	}
]);

const fetchLineData = async () => {
	lineConfig.value.map((item, index) => {
		lineConfig.value[index].loading = true;
		const range = {
			start: timeParse(new Date(times.value[0])),
			end: timeParse(new Date(times.value[1])),
			step: getStepWithTimeRange(times.value)
		};
		getRangeVector({
			range,
			query: item.query
				.replace('$container', detail.value.name)
				.replace('$namespace', detail.value.namespace)
				.replace('$pod', detail.value.appName)
		})
			.then((res) => {
				const data = get(res, `data.data[0].values`, []);
				const list = data.map((item) => [
					item.timestamp,
					isNaN(item.value) || item.value === 'Infinity'
						? '-'
						: round(item.value, 2)
				]);
				lineConfig.value[index].data = [list];
			})
			.finally(() => {
				lineConfig.value[index].loading = false;
			})
			.catch(() => {
				lineConfig.value[index].loading = false;
			});
	});
};

const fetchDetail = async () => {
	const res = await getTaskDetail({ name, podUid: pod_uid });
	detail.value = res.data;
	fetchLineData();
};

watch(times, () => {
	fetchLineData();
});

onMounted(() => {
	fetchDetail();
});
</script>

<style lang="scss" scoped></style>
