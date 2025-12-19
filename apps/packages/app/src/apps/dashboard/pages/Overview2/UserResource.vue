<template>
	<div class="row q-col-gutter-x-xl">
		<div class="col-6">
			<MyCard>
				<div class="row items-center justify-between q-mb-xl">
					<div class="row items-center">
						<q-img
							:src="$q.dark.isActive ? cupIconDark : cupIcon"
							:ratio="1"
							width="24px"
							class="q-mr-md"
						/>
						<div class="text-subtitle1 text-ink-1">
							{{ $t('CPU') }} {{ cpuData.unit }}
						</div>
					</div>
					<div class="text-h5 text-ink-1">
						{{ cpuData.used }}/{{ cpuData.total || '-' }}
					</div>
				</div>
				<MylineChart
					:data="chartData[0]"
					:split-number-y="2"
					style="height: 189px"
					:loading="loading"
				/>
			</MyCard>
		</div>
		<div class="col-6">
			<MyCard>
				<div class="row items-center justify-between q-mb-xl">
					<div class="row items-center">
						<q-img
							:src="$q.dark.isActive ? memory2IconDark : memory2Icon"
							:ratio="1"
							width="24px"
							class="q-mr-md"
						/>
						<div class="text-subtitle1 text-ink-1">
							{{ $t('MEMORY') }} {{ memoryData.unit }}
						</div>
					</div>
					<div class="text-h5 text-ink-1">
						{{ memoryData.used }}/{{ memoryData.total || '-' }}
					</div>
				</div>
				<MylineChart
					:data="chartData[1]"
					style="height: 189px"
					:loading="loading"
				/>
			</MyCard>
		</div>
	</div>
</template>

<script setup lang="ts">
import { get, last } from 'lodash';
import {
	getAreaChartOps,
	getSuitableUnit,
	getValueByUnit
} from '@apps/dashboard/src/utils/monitoring';
import { computed } from 'vue';
import MylineChart from '@apps/control-panel-common/src/components/Charts/MylineChart3.vue';
import MyCard from '@apps/dashboard/components/MyCard.vue';
import cupIcon from '@apps/dashboard/assets/cup.svg';
import cupIconDark from '@apps/dashboard/assets/cup-dark.svg';
import memory2Icon from '@apps/dashboard/assets/memory2.svg';
import memory2IconDark from '@apps/dashboard/assets/memory2-dark.svg';
import {
	getContentOptionsUser,
	getTabOptionsUser,
	MetricTypesUser
} from './config';
import { useQuasar } from 'quasar';
const $q = useQuasar();

interface Props {
	data: any;
	loading: boolean;
	cluster_cpu_total?: string;
	cluster_memory_total?: string;
}
const props = withDefaults(defineProps<Props>(), {
	data: []
});

const chartData = computed(() => {
	return getContentOptionsUser(props.data, MetricTypesUser).map((item) =>
		getAreaChartOps(item)
	);
});
const userResourcesLast = computed(() =>
	getTabOptionsUser(props.data, MetricTypesUser)
);

const cpuData = computed(() => {
	const target: any = userResourcesLast.value[0];
	const totalValue = props.cluster_cpu_total || target.total;
	const _unit = getSuitableUnit(
		totalValue || target.used,
		target.unitType as any
	);
	const _used = getValueByUnit(target.used, _unit);
	const _total = getValueByUnit(totalValue, _unit);
	return {
		unit: _unit,
		used: _used,
		total: _total
	};
});

const memoryData = computed(() => {
	const target: any = userResourcesLast.value[1];
	const totalValue = props.cluster_memory_total || target.total;
	const _unit = getSuitableUnit(
		totalValue || target.used,
		target.unitType as any
	);
	const _used = getValueByUnit(target.used, _unit);
	const _total = getValueByUnit(totalValue, _unit);
	return {
		unit: _unit,
		used: _used,
		total: _total
	};
});
</script>

<style></style>
