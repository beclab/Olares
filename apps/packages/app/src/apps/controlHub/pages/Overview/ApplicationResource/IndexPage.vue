<template>
	<div class="q-gutter-sm">
		<my-card flat title="Cluster Resource Usage">
			<template #extra>
				<div class="row items-center">
					<DateRangeMonitoring
						:default-value="selectValueCluster"
						@change="selecteChangeCluster"
					/>
					<q-btn
						outline
						padding="4px"
						dense
						:disable="loading[0]"
						color="grey-5"
						style="margin-left: 8px; border-radius: 8px"
						@click="fetchCluster"
					>
						<q-icon name="refresh"></q-icon>
					</q-btn>
				</div>
			</template>
			<MyLoading style="min-height: 127px" :showing="loading[0]">
				<div class="row q-col-gutter-lg">
					<div
						class="col-6"
						v-for="(item, index) in clusterList"
						:key="`md-${index}`"
					>
						<MylineChartMini style="margin-top: 24px" :data="item" />
					</div>
				</div>
			</MyLoading>
		</my-card>

		<q-space />

		<my-card flat title="Application Resource Usage">
			<template #extra>
				<div class="row items-center">
					<DateRangeMonitoring
						:default-value="selectValueApp"
						@change="selecteChangeApp"
					/>
					<q-btn
						outline
						padding="4px"
						dense
						:disable="loading[1]"
						color="grey-5"
						style="margin-left: 8px; border-radius: 8px"
						@click="fetchApplication"
					>
						<q-icon name="refresh"></q-icon>
					</q-btn>
				</div>
			</template>
			<MyLoading style="min-height: 381px" :showing="loading[1]">
				<div class="row q-col-gutter-lg">
					<div
						class="col-4"
						v-for="(item, index) in appList"
						:key="`md-${index}`"
					>
						<MylineChartMini style="margin-top: 24px" :data="item" />
					</div>
				</div>
			</MyLoading>
		</my-card>

		<my-card flat title="Projects">
			<MyLoading style="min-height: 134px" :showing="loading[2]">
				<div class="row q-col-gutter-lg">
					<div
						class="col-12"
						v-for="(item, index) in projectList"
						:key="`md-${index}`"
					>
						<MylineChart style="height: 110px" :data="item" />
					</div>
				</div>
			</MyLoading>
		</my-card>
	</div>
</template>

<script setup lang="ts">
import { getClusterMonitoring } from '@apps/control-hub/src/network';
import { onMounted, ref } from 'vue';
import {
	chartConfigCluster,
	chartConfigProject,
	getMonitoringCfgs,
	chartConfigApp
} from './config';
import MylineChart from '@apps/control-hub/components/Charts/MylineChart.vue';
import MylineChartMini from '@apps/control-hub/components/Charts/MylineChartMini.vue';
import MyCard from '@apps/control-hub/src/components/MyCard.vue';
import DateRangeMonitoring, {
	DateRangeItem,
	options
} from '@apps/control-hub/src/containers/DateRangeMonitoring.vue';
import { date } from 'quasar';
import {
	getAreaChartOps,
	getResult
} from '@apps/control-hub/src/utils/monitoring';
import MyLoading from '@apps/control-hub/src/components/MyLoading.vue';
import { timeParams } from '@apps/control-panel-common/src/config/resource.common';

const clusterList = ref();
const appList = ref();
const projectList = ref();
const loading = ref([false, false, false]);
const selectValueCluster = ref<DateRangeItem>(options[6]);
const selectValueApp = ref<DateRangeItem>(options[6]);

const selecteChangeCluster = (value: DateRangeItem) => {
	selectValueCluster.value = value;
	fetchCluster();
};

const selecteChangeApp = (value: DateRangeItem) => {
	selectValueApp.value = value;
	fetchApplication();
};

const fetchCluster = () => {
	const end = date.formatDate(Date.now(), 'X');
	const start = (Number(end) - selectValueCluster.value.value).toString();
	const params = {
		...timeParams,
		start,
		end,
		metrics_filter:
			'cluster_cpu_usage|cluster_memory_usage_wo_cache|cluster_disk_size_usage$'
	};
	loading.value[0] = true;
	getClusterMonitoring(params)
		.then((res) => {
			const data = getResult(res.data.results);
			clusterList.value = chartConfigCluster(data, getMonitoringCfgs());
		})
		.finally(() => {
			loading.value[0] = false;
		});
};

const fetchApplication = () => {
	const end = date.formatDate(Date.now(), 'X');
	const start = (Number(end) - selectValueApp.value.value).toString();
	const params = {
		...timeParams,
		start,
		end,
		metrics_filter:
			'cluster_deployment_count|cluster_statefulset_count|cluster_daemonset_count|cluster_job_count|cluster_cronjob_count|cluster_pvc_count|cluster_service_count|cluster_ingresses_extensions_count|cluster_pod_running_count$'
	};
	loading.value[1] = true;
	getClusterMonitoring(params)
		.then((res) => {
			const data = getResult(res.data.results);
			appList.value = chartConfigCluster(data, chartConfigApp());
		})
		.finally(() => {
			loading.value[1] = false;
		});
};

const fetchProject = () => {
	const end = date.formatDate(Date.now(), 'X');
	const start = (Number(end) - selectValueApp.value.value).toString();
	const params = {
		...timeParams,
		start,
		end,
		metrics_filter: 'cluster_namespace_count$'
	};
	loading.value[2] = true;
	getClusterMonitoring(params)
		.then((res) => {
			const data = getResult(res.data.results);
			projectList.value = chartConfigProject(data).map((item) =>
				getAreaChartOps(item)
			);
		})
		.finally(() => {
			loading.value[2] = false;
		});
};

onMounted(() => {
	fetchCluster();
	fetchApplication();
	fetchProject();
});
</script>

<style></style>
