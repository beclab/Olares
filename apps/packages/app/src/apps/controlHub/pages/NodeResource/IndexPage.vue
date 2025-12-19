<template>
	<MyCard :title="$t('RESOURCE')" flat>
		<template #extra>
			<div class="row q-gutter-x-md">
				<Refresh dense flat icon="sym_r_refresh" @click="init" />
				<NodeSelect @update:model-value="typeChangeHandler"> </NodeSelect>
			</div>
		</template>
		<ClusterResource
			:data="clusterData"
			:type="selectValue.type"
			:loading="loading"
		></ClusterResource>
	</MyCard>

	<RouterViewTransition></RouterViewTransition>
</template>

<script setup lang="ts">
import ClusterResource from './ClusterResource.vue';
import RouterViewTransition from '@apps/control-panel-common/src/components/RouterViewTransition.vue';

import { getContentOptions, getTabOptions, MetricTypesFormat } from './config';
import {
	getAreaChartOps,
	getResult
} from '@apps/control-hub/src/utils/monitoring';
import {
	fillEmptyMetrics,
	getParams
} from '@apps/control-panel-common/src/containers/Monitoring/config';
import {
	getClusterMonitoring,
	getNodeMonitoring,
	getNodesList
} from '@apps/control-panel-common/src/network';
import { computed, onBeforeUnmount, onMounted, ref } from 'vue';
import { ResourcesResponse } from '@apps/control-panel-common/src/network/network';
import { getRefreshResult } from '@apps/control-panel-common/src/containers/PodsList/config';
import MyLoading from '@apps/control-panel-common/src/components/MyLoading.vue';
import QSectionStyle from '@apps/control-panel-common/src/components/QSectionStyle.vue';
import { useRouter } from 'vue-router';
import QButtonStyle from '@apps/control-panel-common/src/components/QButtonStyle.vue';
import MyCard from '@apps/control-panel-common/src/components/MyCard3.vue';
import Refresh from '@apps/control-panel-common/src/components/Refresh.vue';
import NodeSelect from '@apps/dashboard/src/containers/NodeSelect/IndexPage.vue';
import { selectValue } from '@apps/dashboard/src/containers/NodeSelect/config';
import { timeParams } from '@apps/control-panel-common/src/config/resource.common';

const clusterData = ref([]);
const loading = ref(false);
const nodes = ref<string[]>([]);

type Locker = string | number | NodeJS.Timeout | undefined;
let locker: Locker = undefined;
const MetricTypesNode = MetricTypesFormat('node');
const MetricTypesCluster = MetricTypesFormat('cluster');

const fetchData = async (autofresh = false) => {
	let filters: any = {};
	const filtersNode = {
		resources: [selectValue.value.value],
		metrics: Object.values(MetricTypesNode),
		...timeParams
	};

	const filtersCluster = {
		metrics: Object.values(MetricTypesCluster),
		...timeParams
	};
	let fn = getNodeMonitoring;
	if (selectValue.value.type === 'cluster') {
		filters = filtersCluster;
		fn = getClusterMonitoring;
	} else {
		filters = filtersNode;
		fn = getNodeMonitoring;
	}

	if (autofresh) {
		filters.last = true;
	} else {
		loading.value = true;
	}
	const paramsCluster = getParams(filters);

	const result = await fn(paramsCluster);

	let clusterResultFormat = getResult(result.data.results);
	if (autofresh) {
		clusterResultFormat = getRefreshResult(
			clusterResultFormat,
			clusterData.value
		);
	}
	clusterData.value = fillEmptyMetrics(paramsCluster, clusterResultFormat);
	loading.value = false;

	refresh();
};

const typeChangeHandler = () => {
	clearLocker();
	fetchData();
};

const clearLocker = () => {
	locker && clearTimeout(locker);
};

const refresh = () => {
	clearLocker();
	locker = setTimeout(() => {
		fetchData(true);
	}, 10000);
};

const init = () => {
	fetchData();
};

onMounted(() => {
	init();
});

onBeforeUnmount(() => {
	clearLocker();
});
</script>

<style></style>
