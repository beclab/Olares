<template>
	<div class="metrics-table-container">
		<bt-scroll-area style="height: 265px">
			<table class="my-data-detail-table text-ink-1">
				<thead class="table-header">
					<tr>
						<th colspan="1">
							<div class="text-h6 text-uppercase-first text-left">
								{{ label ? label : type }}
							</div>
						</th>
						<th colspan="2">
							<div class="text-h6 text-uppercase-first text-right">
								{{ valueLabel }}
							</div>
						</th>
					</tr>
				</thead>
				<tr v-for="item in filteredData" :key="item.name" class="table-wrapper">
					<td class="first-td">
						<span class="text-subtitle1 text-ink-1">{{
							item.x || '(Unknown)'
						}}</span>
					</td>
					<td>
						<div style="width: 100px" class="my-progress-container">
							<q-linear-progress
								class="progress-wrapper"
								size="20px"
								color="light-blue-soft"
								track-color="transparent"
								:value="progressLabel(item.z) / 100"
							>
								<div class="absolute-full flex items-center">
									<span
										style="font-size: 12px"
										class="text-light-blue-default text-subtitle3 q-pr-sm my-progress-label"
									>
										{{ progressLabel(item.z) }}%
									</span>
								</div>
							</q-linear-progress>
						</div>
					</td>
					<td>
						<div class="text-right text-h6 q-pl-lg">
							{{ item.y }}
						</div>
					</td>
				</tr>
				<WebsiteCard :loading="loading" />
			</table>
		</bt-scroll-area>
		<Empty center v-if="data.length === 0 && !loading"></Empty>
	</div>
</template>

<script setup lang="ts">
import { percentFilter } from '@apps/dashboard/src/utils/filters';
import { computed, watch, ref, toRefs } from 'vue';
import firstBy from 'thenby';
import { round } from 'lodash';
import { getAnalyticsMetrics } from '@apps/dashboard/src/network';
import { dateRange, dateValue } from './DatefilterState';
import Empty from '@apps/control-panel-common/src/components/Empty2.vue';
import WebsiteCard from '@apps/dashboard/components/Skeleton/WebsiteCard.vue';
import { t } from '@apps/dashboard/src/boot/i18n';

interface Props {
	websiteId: string;
	type: string;
	dataFilter?: (data: any, filterOptions: any) => any;
	filterOptions?: any;
	limit?: number;
	label?: string;
	valueLabel?: string;
}

const props = withDefaults(defineProps<Props>(), {
	limit: undefined,
	valueLabel: t('VISITORS')
});

const data = ref([]);
const loading = ref(false);

const pages = computed(() => {
	return [];
});

const progressLabel = (value: number) => {
	return round(value, 1);
};

const getData = () => {
	const { startDate, endDate, unit } = dateRange(props.websiteId);
	data.value = [];

	const params = {
		type: props.type,
		startAt: +startDate,
		endAt: +endDate
	};
	loading.value = true;
	getAnalyticsMetrics(props.websiteId, params)
		.then((res) => {
			data.value = res.data;
		})
		.finally(() => {
			loading.value = false;
		});
};

const filteredData = computed(() => {
	const { limit, dataFilter, filterOptions } = toRefs(props);
	if (data.value) {
		let items = percentFilter(
			dataFilter.value
				? dataFilter.value(data.value, filterOptions)
				: data.value
		);
		if (limit.value) {
			items = items.filter((e, i) => i < limit.value);
		}
		if (filterOptions.value?.sort === false) {
			return items;
		}
		return items.sort(firstBy('y', -1).thenBy('x'));
	}
	return [];
});

watch(() => dateValue[props.websiteId], getData, {
	immediate: true
});
</script>

<style lang="scss" scoped>
.metrics-table-container {
	.progress-wrapper {
		&::after {
			content: '';
			position: absolute;
			left: 0;
			top: 0;
			bottom: 0;
			width: 2px;
			background: $light-blue-default;
		}
	}

	.table-header {
		position: sticky;
		top: 0;
		z-index: 1;
	}

	.text-uppercase-first::first-letter {
		text-transform: uppercase;
	}
	.my-progress-container {
		transform: rotateY(180deg);
	}
	.my-progress-label {
		transform: rotateY(180deg);
	}
}
</style>
<style lang="scss" scoped>
.my-data-detail-table {
	border-collapse: separate;
	border-spacing: map-get($space-xs, x) map-get($space-md, x);
	width: 100%;
	margin: -12px -4px;
	.text-wrap {
		white-space: wrap;
	}
}

td:first-child {
	width: auto;
}

td:nth-child(2) {
	width: 100px;
}

td:last-child {
	width: 1%;
	white-space: nowrap;
}
</style>
