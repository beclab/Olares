<template>
	<FullPageWithBack :title="$t('MEMORY_DETAILS')">
		<div class="column no-wrap flex-gap-y-xl">
			<div v-for="(node, index) in memoryList" :key="index">
				<MyCard>
					<q-responsive
						:ratio="3.5"
						style="min-height: 264px; max-height: 320px"
					>
						<MylineChart
							v-if="selectValue.value === MemoryType.PHYSICAL_VIDEO_MEMORY"
							:data="node.MemoryChartData"
							class="full-height"
							:loading="loading"
							:title-format="['title']"
						>
							<template #extra>
								<BtSelect
									v-model="selectValue"
									:options="memoryOptions"
									@update:model-value="changeHandler"
									dense
									outlined
									style="width: 200px"
								/>
							</template>
						</MylineChart>

						<MylineChart
							v-else
							:data="node.MemoryChartDataExchange"
							class="full-height"
							:loading="loading"
							:legendHide="true"
							:title-format="['title']"
						>
							<template #extra>
								<BtSelect
									v-model="selectValue"
									:options="memoryOptions"
									@update:model-value="fetchData"
									dense
									outlined
									style="width: 200px"
								/>
							</template>
						</MylineChart>
					</q-responsive>
					<div
						class="column flex-gap-md q-mt-xl"
						v-if="selectValue.value === MemoryType.PHYSICAL_VIDEO_MEMORY"
					>
						<div class="row justify-between items-center q-py-md">
							<span class="text-h6 text-ink-1">{{
								$t('MEMORY_OP.MEMORY_STRUCTURE')
							}}</span>
							<div class="row items-center flex-gap-xl">
								<div
									v-for="(item, index) in node.memoryStructure2"
									:key="index"
									class="text-ink-1"
								>
									<span class="text-subtitle3">{{ item.label }}</span>
									<span class="text-h6 q-ml-xs">{{ item.value }}</span>
								</div>
							</div>
						</div>
						<div class="row memory-linear-progress-container bg-background-3">
							<div
								v-for="(item, index) in node.memoryStructure"
								:key="index"
								:style="{ flex: item.size }"
							>
								<q-linear-progress :color="item.color" :value="1" size="6px" />
								<q-tooltip anchor="top middle" self="bottom middle">{{
									item.label
								}}</q-tooltip>
							</div>
						</div>
						<div class="row flex-gap-xl">
							<ContainerBox
								v-for="(item, index) in node.memoryStructure"
								:key="index"
								:color="item.color"
							>
								<div class="column justify-around full-height flex-gap-xs">
									<div
										class="row items-center flex-gap-xs text-subtitle3 text-ink-2"
									>
										{{ item.label }}
										<template v-if="item.info">
											<q-icon name="sym_r_info" color="ink-3" size="16px">
											</q-icon>
											<q-tooltip anchor="top middle" self="bottom middle">{{
												item.info
											}}</q-tooltip>
										</template>
									</div>
									<div class="text-h6 text-ink-1">{{ item.value }}</div>
								</div>
							</ContainerBox>
						</div>
					</div>
					<div v-else class="row justify-between q-mt-xl">
						<div class="row flex-gap-xl">
							<ContainerBox
								v-for="(item, index) in node.memoryExchangeList"
								:key="index"
								:color="item.color"
							>
								<div class="column justify-around full-height flex-gap-xs">
									<div class="text-subtitle3 text-ink-2">{{ item.label }}</div>
									<div class="text-h6 text-ink-1">{{ item.value }}</div>
								</div>
							</ContainerBox>
						</div>
						<div class="row flex-gap-xl">
							<ContainerBox
								v-for="(item, index) in node.memoryExchangeList2"
								:key="index"
								:color="item.color"
							>
								<div class="column justify-around full-height flex-gap-xs">
									<div class="text-subtitle3 text-ink-2">
										{{ item.label }}
									</div>
									<div class="text-h6 text-ink-1">{{ item.value }}</div>
								</div>
							</ContainerBox>
						</div>
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
import {
	MetricTypes,
	memoryOptions,
	MemoryType,
	getMemoryList
} from './config';
import {
	fillEmptyMetrics,
	getParams
} from '@apps/control-panel-common/src/containers/Monitoring/config';
import { getResult } from '@apps/dashboard/src/utils/monitoring';
import MylineChart from '@apps/control-panel-common/src/components/Charts/MylineChart.vue';
import ContainerBox from '../components/ContainerBox.vue';
import BtSelect from '@apps/control-panel-common/src/components/Select.vue';
import { Locker } from '@apps/dashboard/src/types/main';
import { getRefreshResult } from '@apps/control-panel-common/src/containers/PodsList/config';
import { timeRangeDefault } from '../../../../controlPanelCommon/config/resource.common';

const MemoryData = ref({});
const loading = ref(false);
const selectValue = ref(memoryOptions[0]);

const memoryList = computed(() => {
	return getMemoryList(MemoryData.value, MetricTypes);
});

const fetchData = async (autofresh = false) => {
	let filters = {
		metrics: Object.values(MetricTypes),
		last: false
	};
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
			result = getRefreshResult(result, MemoryData.value);
		}
		MemoryData.value = fillEmptyMetrics(params, result);
		refresh();
	} catch (error) {
		loading.value = false;
	}
	loading.value = false;
};

const changeHandler = () => {
	clearLocker();
	fetchData();
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
.memory-linear-progress-container {
	border-radius: 3px;
	overflow: hidden;
}
</style>
