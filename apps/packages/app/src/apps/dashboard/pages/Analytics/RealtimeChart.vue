<template>
	<MylineChart :data="chartData" />
</template>

<script setup lang="ts">
import MylineChart from '@apps/control-panel-common/src/components/Charts/MylineChart.vue';
import { reactive, watch } from 'vue';
import { subMinutes, startOfMinute, format } from 'date-fns';
import { mapData, REALTIME_RANGE } from './config';
import { getDateArray } from '@apps/dashboard/src/utils/analytics-date';

interface Props {
	data: any;
}

const props = withDefaults(defineProps<Props>(), {});

const chartData = reactive({
	unit: 'page views',
	legend: ['Page views', 'Unique visitors'],
	data: []
});

watch(
	() => props.data,
	(data) => {
		const endDate = startOfMinute(new Date());
		const startDate = subMinutes(endDate, REALTIME_RANGE);
		const unit = 'minute';
		const pageviewsData = getDateArray(
			mapData(data.pageviews),
			startDate,
			endDate,
			unit
		);

		const sessionsData = getDateArray(
			mapData(data.visitors),
			startDate,
			endDate,
			unit
		);

		const data1 = pageviewsData.map((item) => [format(item.x, 'H:m'), item.y]);
		const data2 = sessionsData.map((item) => [format(item.x, 'H:m'), item.y]);
		chartData.data = [data1, data2];
	},
	{
		deep: true
	}
);
</script>

<style></style>
