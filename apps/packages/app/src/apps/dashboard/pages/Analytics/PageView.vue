<template>
	<bt-scroll-area style="height: calc(100vh - 104px)">
		<MyCard flat>
			<q-list>
				<q-item>
					<q-item-section>
						<RealtimeHeader :data="currentData" />
					</q-item-section>
				</q-item>
				<q-item>
					<q-item-section>
						<RealtimeChart :data="realtimeData" style="height: 300px" />
					</q-item-section>
				</q-item>
				<q-separator spaced />
				<q-item>
					<q-item-section>
						<div class="row">
							<PagesTable class="col-5" :websiteId="id"></PagesTable>
							<div class="col-2">
								<q-separator spaced vertical />
							</div>
							<ReferrersTable class="col-5" :websiteId="id"></ReferrersTable>
						</div>
					</q-item-section>
				</q-item>
				<q-separator spaced />
				<q-item>
					<q-item-section>
						<div class="row q-col-gutter-x-md">
							<BrowsersTable class="col-4" :websiteId="id"></BrowsersTable>
							<OSTable class="col-4" :websiteId="id"></OSTable>
							<DevicesTable class="col-4" :websiteId="id"></DevicesTable>
						</div>
					</q-item-section>
				</q-item>
			</q-list>

			<div></div>
		</MyCard>
		<q-inner-loading :showing="loading"> </q-inner-loading>
	</bt-scroll-area>
</template>

<script setup lang="ts">
import { getRealtime } from '@apps/dashboard/src/network/analytics-service';
import { computed, ref, watch } from 'vue';
import MyCard from '@apps/control-panel-common/src/components/MyCard.vue';
import { subMinutes, startOfMinute } from 'date-fns';
import { percentFilter, REALTIME_RANGE } from './config';
import firstBy from 'thenby';
import { useRoute } from 'vue-router';
import axios from 'axios';
import RealtimeChart from './RealtimeChart.vue';
import RealtimeHeader from './RealtimeHeader.vue';
import BrowsersTable from './BrowsersTable.vue';
import PagesTable from './PagesTable.vue';
import OSTable from './OSTable.vue';
import DevicesTable from './DevicesTable.vue';
import ReferrersTable from './ReferrersTable.vue';

interface Props {
	id: string;
}

const props = withDefaults(defineProps<Props>(), {});
const CancelToken = axios.CancelToken;
const source = CancelToken.source();

const loading = ref(false);

const defaultCurrentData = {
	pageviews: [],
	sessions: [],
	events: [],
	countries: [],
	visitors: [],
	timestamp: 0
};
const currentData = ref({ ...defaultCurrentData });

function mergeData(state = [], data = [], time: number) {
	const ids = state.map(({ __id }) => __id);
	return state
		.concat(data.filter(({ __id }) => !ids.includes(__id)))
		.filter(({ timestamp }) => timestamp >= time);
}

const fetchData = (showLoading = true) => {
	if (showLoading) {
		loading.value = true;
	}
	getRealtime(props.id, currentData.value.timestamp, {
		cancelToken: source.token
	})
		.then((res) => {
			const data = res.data;
			const date = subMinutes(startOfMinute(new Date()), REALTIME_RANGE);
			const time = date.getTime();

			currentData.value.pageviews = mergeData(
				currentData.value?.pageviews,
				data.pageviews,
				time
			);
			currentData.value.sessions = mergeData(
				currentData.value?.sessions,
				data.sessions,
				time
			);
			currentData.value.events = mergeData(
				currentData.value?.events,
				data.events,
				time
			);
			currentData.value.timestamp = data.timestamp;
			autoFetch();
		})
		.finally(() => {
			loading.value = false;
		});
};

const realtimeData = computed(() => {
	const currentDataTemp = { ...currentData.value };
	if (!currentDataTemp) {
		return {
			pageviews: [],
			sessions: [],
			events: [],
			countries: [],
			visitors: []
		};
	}

	currentDataTemp.countries = percentFilter(
		currentDataTemp.sessions
			.reduce((arr, data) => {
				if (!arr.find(({ id }) => id === data.id)) {
					return arr.concat(data);
				}
				return arr;
			}, [])
			.reduce((arr, { country }) => {
				if (country) {
					const row = arr.find(({ x }) => x === country);

					if (!row) {
						arr.push({ x: country, y: 1 });
					} else {
						row.y += 1;
					}
				}
				return arr;
			}, [])
			.sort(firstBy('y', -1))
	);

	currentDataTemp.visitors = currentDataTemp.sessions.reduce((arr, val) => {
		if (!arr.find(({ id }) => id === val.id)) {
			return arr.concat(val);
		}
		return arr;
	}, []);

	return currentDataTemp;
});

let locker: any = undefined;
const clearLocker = () => {
	locker && clearTimeout(locker);
};
const autoFetch = () => {
	clearLocker();
	locker = setTimeout(() => {
		fetchData(false);
	}, 3000);
};

watch(
	() => props.id,
	(id) => {
		if (id) {
			clearLocker();
			source.cancel();
			currentData.value = { ...defaultCurrentData };
			fetchData();
		}
	}
);
</script>

<style></style>
