<template>
	<FullPageWithBack :title="name" :avatar="icon" @scroll="scrollHandler">
		<template #extra>
			<div class="row q-gutter-x-md" v-if="!visible">
				<QButtonStyle>
					<q-btn
						class="q-pa-xs"
						icon="refresh"
						dense
						outline
						color="ink-2"
						@click="refresh"
					/>
				</QButtonStyle>
				<DateFilter :website-id="websiteId"></DateFilter>
			</div>
		</template>
		<div class="q-gutter-y-xl">
			<div class="row justify-between items-center q-gutter-col-lg">
				<MetricsBar
					:websiteId="websiteId"
					style="max-width: 780px"
					:loading="loading"
				></MetricsBar>
				<div class="row q-gutter-x-md" v-show="visible">
					<QButtonStyle>
						<q-btn
							class="q-pa-xs"
							icon="refresh"
							dense
							outline
							color="ink-2"
							@click="refresh"
						/>
					</QButtonStyle>
					<DateFilter
						:website-id="websiteId"
						@change="changeHandler"
					></DateFilter>
				</div>
			</div>
			<MyCard>
				<PageviewsChart
					:data="data"
					:websiteId="websiteId"
					:loading="loading"
					style="height: 380px"
					toggle
				/>
			</MyCard>
			<MyGridLayout col-width="500px" gap="xl" :key="updateKey">
				<MyCard>
					<PagesTable :websiteId="websiteId" :loading="loading"></PagesTable>
				</MyCard>
				<MyCard>
					<ReferrersTable
						:websiteId="websiteId"
						:loading="loading"
					></ReferrersTable>
				</MyCard>
				<MyCard>
					<BrowsersTable
						:websiteId="websiteId"
						:loading="loading"
					></BrowsersTable>
				</MyCard>
				<MyCard>
					<OSTable :websiteId="websiteId" :loading="loading"></OSTable>
				</MyCard>
				<MyCard>
					<DevicesTable
						:websiteId="websiteId"
						:loading="loading"
					></DevicesTable>
				</MyCard>
				<MyCard>
					<CitiesTable :websiteId="websiteId" :loading="loading"></CitiesTable>
				</MyCard>
			</MyGridLayout>
		</div>
	</FullPageWithBack>
</template>

<script setup lang="ts">
import { getPageviews } from '@apps/dashboard/src/network/analytics-service';
import { ref, toRefs, watch } from 'vue';
import { useRoute } from 'vue-router';
import BrowsersTable from './BrowsersTable.vue';
import PagesTable from './PagesTable.vue';
import OSTable from './OSTable.vue';
import DevicesTable from './DevicesTable.vue';
import ReferrersTable from './ReferrersTable.vue';
import MetricsBar from './MetricsBar.vue';
import CitiesTable from './CitiesTable.vue';
import { getDateArray } from '@apps/dashboard/src/utils/analytics-date';
import PageviewsChart from './PageviewsChart.vue';
import { dateRange, dateValue } from './DatefilterState';
import MyCard from '@apps/dashboard/src/components/MyCard.vue';
import MyGridLayout from '@apps/control-panel-common/src/components/MyGridLayout.vue';
import { getTimezone } from './date';
import DateFilter from './DateFilter.vue';
import QButtonStyle from '@apps/control-panel-common/src/components/QButtonStyle.vue';
import FullPageWithBack from '@apps/control-panel-common/src/components/FullPageWithBack2.vue';

interface Props {
	websiteId: string;
}

const props = withDefaults(defineProps<Props>(), {});
const route = useRoute();
const data = ref();
const chartData = ref();
const loading = ref(false);
const { websiteId } = toRefs(props);
const updateKey = ref(Math.random().toString());
const visible = ref(false);
const { name, icon }: Record<string, any> = route.query;

const scrollHandler = (data: any) => {
	if (data.verticalPosition < 120) {
		visible.value = true;
	} else {
		visible.value = false;
	}
};
const fetchData = () => {
	const { startDate, endDate, unit } = dateRange(websiteId.value);
	const timezone = getTimezone();

	const params = {
		startAt: +startDate,
		endAt: +endDate,
		unit,
		timezone
	};
	loading.value = true;
	getPageviews(websiteId.value, params)
		.then((res) => {
			data.value = res.data;
		})
		.finally(() => {
			loading.value = false;
		});
};

const refresh = () => {
	fetchData();
	updateKey.value = Math.random().toString();
};

const changeHandler = () => {
	fetchData();
};

watch(
	() => websiteId.value,
	() => {
		if (websiteId.value) {
			fetchData();
		}
	},
	{
		immediate: true
	}
);
</script>

<style lang="scss" scoped>
.data-page-wrapper {
	flex: 1;
}
</style>
