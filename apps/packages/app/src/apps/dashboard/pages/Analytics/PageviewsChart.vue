<template>
	<div class="pageview-chart-container">
		<div class="pageview-chart-wrapper">
			<QBtnToggleStyle class="chart-type-toggle" v-if="toggle">
				<q-btn-toggle
					v-model="type"
					dense
					no-caps
					unelevated
					text-color="ink-2"
					toggle-color="light-blue-soft"
					:options="[
						{ value: 'line', slot: 'line' },
						{ value: 'bar', slot: 'bar' }
					]"
				>
					<template #line>
						<q-icon
							name="sym_r_planner_review"
							:color="type === 'line' ? 'light-blue' : ''"
							style="padding: 3px"
							size="18px"
						/>
					</template>
					<template #bar>
						<q-icon
							name="sym_r_insert_chart"
							:color="type === 'bar' ? 'light-blue' : ''"
							style="padding: 3px"
							size="18px"
						/>
					</template>
				</q-btn-toggle>
			</QBtnToggleStyle>

			<MyBarChart
				v-if="type === 'bar' && toggle"
				style="height: 100%"
				:data="chartData"
				:loading="loading"
				:splitNumberY="!toggle ? 2 : undefined"
			/>
			<MylineChart2
				v-else
				style="height: 100%"
				:data="chartData"
				:loading="loading"
				:splitNumberY="!toggle ? 2 : undefined"
			/>
		</div>
	</div>
</template>

<script setup lang="ts">
import MyBarChart from '@apps/control-panel-common/src/components/Charts/MyBar2Chart.vue';
import { reactive, ref, watch } from 'vue';
import { getDateArray } from '@apps/dashboard/src/utils/analytics-date';
import { date } from 'quasar';
import { useRoute } from 'vue-router';
import { dateFormatStr, dateRange } from './DatefilterState';
import MylineChart2 from '@apps/control-panel-common/src/components/Charts/MylineChart2.vue';
import QBtnToggleStyle from '@apps/control-panel-common/src/components/QBtnToggleStyle.vue';
import { t } from '@apps/dashboard/boot/i18n';
interface Props {
	data: any;
	websiteId: string;
	loading?: boolean;
	/**
	 * show chart type bar chart toggle
	 */
	toggle?: boolean;
}

const props = withDefaults(defineProps<Props>(), {});
const type = ref('line');
const chartData = reactive({
	unit: [t('VISITORS'), t('VIEWS')],
	legend: [t('UNIQUE_VISITORS'), t('PAGE_VIEWS')],
	data: [],
	type: 'line',
	toggle: false
});
const route = useRoute();

watch(
	() => props.data,
	(data) => {
		const { startDate, endDate, unit, value, modified } = dateRange(
			props.websiteId
		);
		let dataFormat = dateFormatStr[unit];

		const pageviews = getDateArray(data.pageviews, startDate, endDate, unit);
		const sessions = getDateArray(data.sessions, startDate, endDate, unit);

		const pageviewsData = pageviews.map((item: any) => [
			date.formatDate(item.x, dataFormat),
			item.y
		]);
		const sessionsData = sessions.map((item: any) => [
			date.formatDate(item.x, dataFormat),
			item.y
		]);
		chartData.data = [sessionsData, pageviewsData];
	},
	{
		deep: true
	}
);
</script>

<style lang="scss" scoped>
.pageview-chart-container {
	position: relative;
	.pageview-chart-wrapper {
		position: absolute;
		top: 0;
		left: 0;
		right: 0;
		bottom: 0;
	}
	.chart-type-toggle {
		position: absolute;
		right: 0;
		top: 0;
		z-index: 1;
	}
}
</style>
