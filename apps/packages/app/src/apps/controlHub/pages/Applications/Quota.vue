<template>
	<div class="row q-gutter-x-lg">
		<InfoCardItem
			v-for="item in clusterOptions"
			:key="item.name"
			:used="item.used"
			:total="item.total"
			:name="item.name"
			:unit-type="item.unitType"
			:img="item.img"
			:icon="item.icon"
			class="quota-wrapper"
		></InfoCardItem>
	</div>
</template>

<script setup lang="ts">
import {
	fillEmptyMetrics,
	getParams,
	getResult
} from '@apps/control-panel-common/src/containers/Monitoring/config';
import InfoCardItem from '@apps/control-hub/components/InfoCard/InfoCardItem.vue';
import { getNamespaces } from '@apps/control-hub/src/network';
import { ref, watch } from 'vue';
import { useRoute } from 'vue-router';
import { getTabOptions, MetricTypesFormat } from './config';

const clusterOptions = ref();
const route = useRoute();

const metrics_filter = Object.keys(MetricTypesFormat('namespace')).join('|');

const fetchData = () => {
	const { namespace }: Record<string, any> = route.params;
	const params = {
		type: 'rank',
		metrics_filter,
		resources_filter: namespace
	};

	getNamespaces(params).then((res) => {
		let clusterResultFormat = getResult(res.data.results);

		const data = fillEmptyMetrics(params, clusterResultFormat);
		const MetricTypes = MetricTypesFormat('namespace');
		clusterOptions.value = getTabOptions(data, MetricTypes);
		console.log('data', clusterOptions.value);
	});
};
watch(
	() => route.params.namespace,
	() => {
		fetchData();
	},
	{
		immediate: true
	}
);
</script>

<style lang="scss" scoped>
.quota-wrapper {
	border: 1px solid $separator;
	border-radius: 8px;
}
</style>
