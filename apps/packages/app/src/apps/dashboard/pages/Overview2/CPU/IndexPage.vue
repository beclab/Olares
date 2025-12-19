<template>
	<FullPageWithBack :title="$t('CPU_DETAILS')">
		<div class="column no-wrap flex-gap-y-xl">
			<div v-for="node in cpuList" :key="node.name">
				<MyCard>
					<q-responsive
						:ratio="3.5"
						style="min-height: 264px; max-height: 320px"
					>
						<MylineChart
							:data="node.cpuChartData"
							class="full-height"
							:loading="loading"
							:title-format="['title']"
						>
							<template #extra>
								<div class="row items-center">
									<div
										v-for="(item, index) in node.cpuBase.list"
										:key="index"
										class="row"
									>
										<q-separator
											class="q-mx-md"
											color="ink-3"
											vertical
											v-if="!!index"
										/>
										<span>{{ item }}</span>
									</div>
								</div>
							</template>
						</MylineChart>
					</q-responsive>

					<div class="row flex-gap-x-xxxxl flex-gap-y-xl q-mt-xl">
						<ContainerBox color="light-blue-default" class="usage-box-wrapper">
							<div class="row items-center usage-rate">
								<div
									class="column justify-center flex-gap-sm q-pr-xxl text-subtitle3 text-ink-2"
								>
									<div>
										{{ $t('CPU_OP.UTILIZATION_RATE') }}
									</div>
									<div class="text-h6 text-ink-1">
										{{ node.usageTateList[0].value }}
										{{ node.usageTateList[0].unit }}
									</div>
								</div>
								<div class="column flex-gap-xs no-wrap">
									<div
										v-for="(item, index) in node.usageTateList.slice(1)"
										:key="index"
										class="row items-center text-right text-body3 text-ink-3"
									>
										<div class="usage-label-wrapper">
											{{ item.title }}
											<q-icon name="sym_r_info" color="ink-3" size="16px" />
											:&ensp;
										</div>
										<span class="text-subtitle3 text-ink-1"
											>{{ item.value }}{{ item.unit }}</span
										>
										<q-tooltip anchor="top middle" self="bottom middle">{{
											item.info
										}}</q-tooltip>
									</div>
								</div>
							</div>
						</ContainerBox>
						<ContainerBox color="green-default">
							<div
								class="column justify-center flex-gap-sm full-height text-subtitle3 text-ink-2"
							>
								<div>{{ node.temperature.name }}</div>
								<div
									class="text-h6"
									:class="[temperatureColor(node.temperature.value)]"
								>
									{{ node.temperature.value }}{{ node.temperature.unit }}
								</div>
							</div>
						</ContainerBox>

						<ContainerBox color="ink-3">
							<div
								class="column justify-center flex-gap-sm full-height text-subtitle3 text-ink-2"
							>
								<div>
									{{ $t('CPU_OP.AVERAGE_LOAD') }}
								</div>
								<div class="row flex-gap-xxl">
									<div v-for="(item, index) in node.AverageLoad" :key="index">
										<span class="text-h6 text-ink-1">
											<span>{{ item.value }}</span>
											<span class="text-body3 text-ink-3"
												>&nbsp;/{{ item.unit }}</span
											>
										</span>
									</div>
								</div>
							</div>
						</ContainerBox>
					</div>
				</MyCard>
			</div>
		</div>
	</FullPageWithBack>
</template>

<script setup lang="ts">
import FullPageWithBack from '@apps/control-panel-common/src/components/FullPageWithBack2.vue';
import MyCard from '@apps/dashboard/components/MyCard.vue';
import { getNodeMonitoring } from '@apps/dashboard/src/network';
import { computed, onBeforeUnmount, onMounted, ref } from 'vue';
import { MetricTypes, getCpuList } from './config';
import {
	fillEmptyMetrics,
	getParams
} from '@apps/control-panel-common/src/containers/Monitoring/config';
import { getResult } from '@apps/dashboard/src/utils/monitoring';
import MylineChart from '@apps/control-panel-common/src/components/Charts/MylineChart.vue';
import ContainerBox from '../components/ContainerBox.vue';
import { isNumber } from 'lodash';
import { Locker } from '@apps/dashboard/src/types/main';
import { getRefreshResult } from '@apps/control-panel-common/src/containers/PodsList/config';
import { resourceStatusColor } from '@apps/dashboard/src/utils/status';
import { timeRangeDefault } from '../../../../controlPanelCommon/config/resource.common';

const cpuData = ref({});
const loading = ref(false);

const cpuList = computed(() => {
	return getCpuList(cpuData.value, MetricTypes);
});

const temperatureColor = (value) => {
	if (isNumber(value)) {
		return `text-${resourceStatusColor(value)}`;
	} else {
		return '';
	}
};

const fetchData = async (autofresh = false) => {
	let filters = {
		metrics: Object.values(MetricTypes),
		last: false
	};
	4;
	if (autofresh) {
		filters.last = true;
	} else {
		filters = { ...filters, ...timeRangeDefault };
		loading.value = true;
	}
	try {
		const params = getParams(filters);
		const res = await getNodeMonitoring(params);
		let result = getResult(res.data.results);

		if (autofresh) {
			result = getRefreshResult(result, cpuData.value);
		}

		cpuData.value = fillEmptyMetrics(params, result);
		refresh();
	} catch (error) {
		loading.value = false;
	}
	loading.value = false;
};

let locker: Locker = undefined;

const clearLocker = () => {
	locker && clearTimeout(locker);
};

const refresh = () => {
	clearLocker();
	locker = setTimeout(() => {
		fetchData(true);
	}, 60 * 1000);
};

onMounted(() => {
	fetchData();
});

onBeforeUnmount(() => {
	clearLocker();
});
</script>

<style lang="scss" scoped>
.usage-rate {
	position: relative;
}
.usage-label-wrapper {
	width: 120px;
	white-space: nowrap;
}
.usage-box-wrapper {
	min-width: 267px;
}
</style>
