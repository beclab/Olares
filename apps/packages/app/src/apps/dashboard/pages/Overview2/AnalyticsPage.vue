<template>
	<div>
		<MyCard>
			<AnalyticCardItem
				type="view"
				:value="pageviewsSum"
				:change="pageviewsChange"
			></AnalyticCardItem>
			<q-separator color="separator" class="q-my-xl" />
			<AnalyticCardItem
				type="visit"
				:value="uniquesSum"
				:change="uniquesChange"
			></AnalyticCardItem>
		</MyCard>
	</div>
</template>

<script setup lang="ts">
import AnalyticCardItem from './components/AnalyticCardItem.vue';
import MyCard from '@apps/dashboard/components/MyCard.vue';
import { getAnalyticsStats, getWebsites } from '@apps/dashboard/src/network';
import { computed, onMounted, ref } from 'vue';
import { parseDateRange } from '../Analytics/date';
import { DEFAULT_DATE_OPTIONS_VALUE } from '../Analytics/DatefilterState';
import { sumBy } from 'lodash';

const list = ref();
const fetchData = async () => {
	const { data } = await getWebsites();

	const { startDate, endDate, unit } = parseDateRange(
		DEFAULT_DATE_OPTIONS_VALUE
	);
	const params = {
		startAt: +startDate,
		endAt: +endDate
	};

	Promise.all(data.map((item: any) => getAnalyticsStats(item.id, params))).then(
		(res) => {
			list.value = res.map((item) => item.data);
		}
	);
};

const pageviewsSum = computed(() =>
	sumBy(list.value, (item: any) => item.pageviews.value)
);
const pageviewsChange = computed(() =>
	sumBy(list.value, (item: any) => item.pageviews.change)
);
const uniquesSum = computed(() =>
	sumBy(list.value, (item: any) => item.uniques.value)
);
const uniquesChange = computed(() =>
	sumBy(list.value, (item: any) => item.uniques.change)
);

onMounted(() => {
	fetchData();
});
</script>

<style lang="scss" scoped>
.parent-container {
	width: 48px;
	height: 48px;
	display: flex;
	align-items: center;
	justify-content: center;
	border-radius: 12px !important;
	overflow: hidden;
}
</style>
