<template>
	<PageBackground class="page-bg-wrapper"></PageBackground>
	<div class="row q-col-gutter-x-xl" style="margin-top: -12px">
		<q-resize-observer @resize="onResize" />

		<div style="flex: 1">
			<MyCard2
				:title="$t('CLUSTER_PHYSIC_RESOURCE')"
				:link="$t('MORE_DETAILS')"
				@link-handler="detailsHandler"
			>
				<template #avatar>
					<q-icon name="sym_r_candlestick_chart" size="32px" />
				</template>
				<ClusterResource
					:data="clusterData"
					:type="type.type"
					:loading="resourcesStore.loading"
				></ClusterResource>
			</MyCard2>
			<!-- <div class="row justify-between">
    <QSectionStyle>
      <q-select
        v-model="type"
        :options="options"
        outlined
        dense
        style="width: 240px"
        @update:model-value="typeChangeHandler"
      />
    </QSectionStyle>
  </div> -->
		</div>
	</div>
	<div style="overflow: auto" class="q-mt-xl">
		<div class="row q-col-gutter-xl">
			<MyCard2
				style="flex: 1"
				:title="$t('USER_RESOURCES', { name: appDetail.user.username })"
			>
				<template #avatar>
					<OlaresAvatar class="terminu-avatar-wrapper" />
				</template>
				<UserResource
					:data="userResourcesData"
					:loading="resourcesStore.loading"
					:cluster_cpu_total="cluster_cpu_total"
					:cluster_memory_total="cluster_memory_total"
				></UserResource>
			</MyCard2>
			<div style="height: 507px" class="col-12">
				<UsageRanking></UsageRanking>
			</div>
			<!-- <MyCard2
				:class="[!rightSideVisible ? 'col-6' : 'side-wrapper']"
				:title="$t('ANALYTICS')"
				:link="$t('MORE')"
				@link-handler="routeToAnalytics"
			>
				<Analytics></Analytics>
			</MyCard2> -->
		</div>
	</div>

	<RouterViewTransition></RouterViewTransition>
</template>

<script setup lang="ts">
import ClusterResource from './ClusterResource.vue';
import RouterViewTransition from '@apps/control-panel-common/src/components/RouterViewTransition.vue';

import { MetricTypesFormat, MetricTypesUser } from './config';
import { getResult } from '@apps/dashboard/src/utils/monitoring';
import {
	fillEmptyMetrics,
	getParams
} from '@apps/control-panel-common/src/containers/Monitoring/config';
import {
	getClusterMonitoring,
	getNodeMonitoring,
	getNodesList,
	getUserMetric
} from '@apps/control-panel-common/src/network';
import { computed, onBeforeUnmount, ref } from 'vue';
import { ResourcesResponse } from '@apps/control-panel-common/src/network/network';
import { getRefreshResult } from '@apps/control-panel-common/src/containers/PodsList/config';
import { useRouter } from 'vue-router';
import MyCard2 from '@apps/dashboard/components/MyCard2.vue';
import UserResource from './UserResource.vue';
import UsageRanking from './UsageRanking.vue';
import { useAppDetailStore } from '@apps/dashboard/src/stores/AppDetail';
import Analytics from './AnalyticsPage.vue';
import { get, last } from 'lodash';
import PageBackground from '@apps/dashboard/components/PageBackground.vue';
import OlaresAvatar from '@apps/dashboard/src/containers/OlaresAvatar.vue';
import { useResourcesStore } from '@apps/dashboard/src/stores/Resource';
import { Locker } from '@apps/dashboard/src/types/main';
import { timeParams } from '../../../controlPanelCommon/config/resource.common';

const resourcesStore = useResourcesStore();
const appDetail = useAppDetailStore();

const router = useRouter();
const defaultOptions = [
	{
		label: 'Cluster',
		value: 'cluster',
		type: 'cluster'
	}
];
const clusterData = ref([]);
const userResourcesData = ref([]);
const NodeData = ref([]);
const nodesList = ref<ResourcesResponse['items']>([]);
const nodes = ref<string[]>([]);
const options = ref(defaultOptions);
const type = ref(defaultOptions[0]);
const rightSideVisible = ref(false);

const nodeMetricShow = computed(() => nodes.value.length > 1);

let locker: Locker = undefined;
const MetricTypesNode = MetricTypesFormat('node');
const MetricTypesCluster = MetricTypesFormat('cluster');

const cluster_cpu_total = computed((): string | undefined =>
	appDetail.isAdmin
		? last(
				last(
					get(clusterData.value, 'cluster_cpu_total.data.result[0].values', [])
				)
		  )
		: undefined
);

const cluster_memory_total = computed((): string | undefined =>
	appDetail.isAdmin
		? last(
				last(
					get(
						clusterData.value,
						'cluster_memory_total.data.result[0].values',
						[]
					)
				)
		  )
		: undefined
);

const fetchData = async (autofresh = false) => {
	let filters: any = {};
	const filtersNode = {
		resources: [type.value.value],
		metrics: Object.values(MetricTypesNode),
		...timeParams
	};

	const filtersUser: any = {
		metrics: Object.values(MetricTypesUser),
		...timeParams
	};

	const filtersCluster = {
		metrics: Object.values(MetricTypesCluster),
		...timeParams
	};
	let fn = getNodeMonitoring;
	if (type.value.value === 'cluster') {
		filters = filtersCluster;
		fn = getClusterMonitoring;
	} else {
		filters = filtersNode;
		fn = getNodeMonitoring;
	}

	if (autofresh) {
		filters.last = true;
		filtersUser.last = true;
	} else {
		resourcesStore.loading = true;
	}

	const paramsCluster = getParams(filters);
	const paramsUser = getParams(filtersUser);

	Promise.all([
		fn(paramsCluster),
		getUserMetric(appDetail.user.username, paramsUser)
	])
		.then(([result, resultUser]) => {
			let clusterResultFormat = getResult(result.data.results);
			let userResultFormat = getResult(resultUser.data.results);

			if (autofresh) {
				clusterResultFormat = getRefreshResult(
					clusterResultFormat,
					clusterData.value
				);

				userResultFormat = getRefreshResult(
					userResultFormat,
					userResourcesData.value
				);
			}
			clusterData.value = fillEmptyMetrics(paramsCluster, clusterResultFormat);
			userResourcesData.value = fillEmptyMetrics(paramsUser, userResultFormat);

			resourcesStore.loading = false;
			refresh();
		})
		.catch(() => {
			refresh();
			resourcesStore.loading = false;
		});
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
	}, 60 * 1000);
};

const getNodes = () => {
	const params = {
		sortBy: 'createTime'
		// labelSelector: '!node-role.kubernetes.io/edge',
	};
	resourcesStore.loading = true;
	getNodesList(params).then((res) => {
		nodesList.value = res.data.items;
		if (nodesList.value.length > 1) {
			const nodesOptions = nodesList.value.map((item, index) => ({
				label: item.metadata.name,
				value: item.metadata.name,
				type: 'node'
			}));
			options.value = [...defaultOptions, ...nodesOptions];
		}
	});
};

const detailsHandler = () => {
	router.push({
		path: `/physical-resources/${type.value.value}`
	});
};

const routeToAnalytics = () => {
	router.push({
		path: '/analytics'
	});
};

const onResize = (size: { width: number; height: number }) => {
	if (size.width < 1599) {
		rightSideVisible.value = false;
	} else {
		rightSideVisible.value = true;
	}
};
// getNodes();
fetchData();

onBeforeUnmount(() => {
	clearLocker();
});
</script>

<style lang="scss" scoped>
.side-wrapper {
	width: 460px;
}
.terminu-avatar-wrapper {
	border-radius: 50%;
	overflow: hidden;
}
</style>
