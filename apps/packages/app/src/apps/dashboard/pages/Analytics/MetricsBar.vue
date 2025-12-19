<template>
	<MyGridLayout col-width="135px">
		<MetricCard
			:value="data.pageviews.value"
			:label="$t('VIEWS')"
			:change="data.pageviews.change"
			:loading="loading"
		/>
		<MetricCard
			:value="data.uniques.value"
			:label="$t('VISITORS')"
			:change="data.uniques.change"
			:loading="loading"
		/>
		<!-- <MetricCard
      :value="data.uniques.value ? (num / data.uniques.value) * 100 : 0"
      label="Bounce rate"
      unit="%"
      :change="
        data.uniques.value && data.uniques.change
          ? (num / data.uniques.value) * 100 -
              (Math.min(diffs.uniques, diffs.bounces) / diffs.uniques) * 100 ||
            0
          : 0
      "
    /> -->
		<MetricCard
			:loading="loading"
			:value="
				data.totaltime.value && data.pageviews.value
					? data.totaltime.value / (data.pageviews.value - data.bounces.value)
					: 0
			"
			:label="$t('AVERAGE_VISIT_TIME')"
			unit="s"
			:change="
				data.totaltime.value && data.pageviews.value
					? (diffs.totaltime / (diffs.pageviews - diffs.bounces) -
							data.totaltime.value /
								(data.pageviews.value - data.bounces.value)) *
							-1 || 0
					: 0
			"
			:format="
        (n: number) =>
          `${n < 0 ? '-' : ''}${formatShortTime(
            Math.abs(~~n),
            ['m', 's'],
            ' '
          )}`
      "
		/>
	</MyGridLayout>
</template>

<script setup lang="ts">
import { getAnalyticsStats } from '@apps/dashboard/src/network';
import { computed, ref, watch } from 'vue';
import MetricCard from './MetricCard.vue';
import { formatShortTime } from './format';
import { dateRange, dateValue } from './DatefilterState';
import MyGridLayout from '@apps/control-panel-common/src/components/MyGridLayout.vue';

interface Props {
	websiteId: string;
	loading?: boolean;
}

const props = withDefaults(defineProps<Props>(), {});
const dataInit = {
	pageviews: {
		value: 0,
		change: 0
	},
	uniques: {
		value: 0,
		change: 0
	},
	bounces: {
		value: 0,
		change: 0
	},
	totaltime: {
		value: 0,
		change: 0
	}
};
const data = ref(dataInit);
// const loading = ref(true);

const num = computed(() =>
	Math.min(
		data.value && data.value.uniques.value,
		data.value && data.value.bounces.value
	)
);
const diffs = computed(
	() =>
		data.value && {
			pageviews: data.value.pageviews.value - data.value.pageviews.change,
			uniques: data.value.uniques.value - data.value.uniques.change,
			bounces: data.value.bounces.value - data.value.bounces.change,
			totaltime: data.value.totaltime.value - data.value.totaltime.change
		}
);

const fetchData = () => {
	const { startDate, endDate, unit } = dateRange(props.websiteId);
	const params = {
		startAt: +startDate,
		endAt: +endDate
	};
	// loading.value = true;

	getAnalyticsStats(props.websiteId, params)
		.then((res) => {
			data.value = res.data;
		})
		.finally(() => {
			// loading.value = false;
		});
};
watch(
	() => dateValue[props.websiteId],
	() => {
		if (dateValue[props.websiteId]) {
			fetchData();
		}
	},
	{
		immediate: true
	}
);

defineExpose({ refresh: fetchData });
</script>

<style></style>
