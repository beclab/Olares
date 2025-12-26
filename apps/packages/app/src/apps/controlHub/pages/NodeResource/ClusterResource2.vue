<template>
	<div class="row no-wrap">
		<InfoCardRadio
			:list="clusterOptions"
			@change="listchange"
			style="min-width: 280px"
			:loading="loading"
		></InfoCardRadio>
		<MyCard
			style="flex: 1; align-self: stretch; border-radius: 12px"
			class="q-ml-lg"
		>
			<MylineChart
				v-if="clusterResource"
				style="height: 100%"
				:data="clusterResource[activeIndex]"
				:loading="loading"
				theme="theme2"
			/>
		</MyCard>
	</div>
</template>

<script setup lang="ts">
import InfoCardRadio from '@apps/dashboard/src/components/InfoCard/InfoCardRadio2.vue';
import { InfoCardItemProps } from '@apps/dashboard/src/components/InfoCard/InfoCardItem.vue';
import MylineChart from '@apps/control-panel-common/src/components/Charts/MylineChart.vue';
import { ref, watch } from 'vue';
import {
	getContentOptions,
	getContentOptions3,
	getTabOptions3,
	MetricTypesFormat
} from './config';
import { getAreaChartOps } from '@apps/control-hub/src/utils/monitoring';
import MyCard from '@apps/dashboard/src/components/MyCard.vue';
import { useQuasar } from 'quasar';
const $q = useQuasar();

interface Props {
	data: any;
	type: string;
	loading: boolean;
}
const props = withDefaults(defineProps<Props>(), {});

const clusterOptions = ref();
const clusterResource = ref();
const activeIndex = ref(0);

const listchange = (data: InfoCardItemProps, index: number) => {
	activeIndex.value = index;
};

watch(
	[() => props.data, () => $q.dark.isActive],
	(data) => {
		const MetricTypes = MetricTypesFormat(props.type);
		clusterOptions.value = getTabOptions3(
			props.data,
			MetricTypes,
			0,
			$q.dark.isActive
		);
		clusterResource.value = getContentOptions3(props.data, MetricTypes).map(
			(item) => getAreaChartOps(item)
		);
	},
	{
		deep: true,
		immediate: true
	}
);
</script>

<style></style>
