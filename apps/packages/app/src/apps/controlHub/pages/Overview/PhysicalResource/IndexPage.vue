<template>
	<bt-scroll-area
		style="height: calc(100vh - 44px)"
		class="my-scroll-container"
	>
		<MyLoading :showing="loading">
			<my-card flat title="Physical Resource Montoring">
				<template #extra>
					<div class="row">
						<DateRangeMonitoring
							:default-value="selectValue"
							@change="selecteChange"
						/>
						<q-btn
							outline
							padding="4px"
							dense
							color="grey-5"
							style="margin-left: 8px; border-radius: 8px"
							:disable="loading"
						>
							<q-icon
								name="refresh"
								class="refresh-icon"
								@click="fetchData"
							></q-icon>
						</q-btn>
					</div>
				</template>

				<div class="row q-col-gutter-lg">
					<div class="col-6" v-for="(item, index) in list" :key="`md-${index}`">
						<MylineChart :data="item" />
					</div>
				</div>
			</my-card>
		</MyLoading>
	</bt-scroll-area>
</template>

<script setup lang="ts">
import MyCard from '@apps/control-hub/src/components/MyCard.vue';
import MylineChart from '@apps/control-hub/src/components/Charts/MylineChart.vue';
import { getClusterMonitoring } from '@apps/control-hub/src/network';
import DateRangeMonitoring, {
	DateRangeItem,
	options
} from '@apps/control-hub/src/containers/DateRangeMonitoring.vue';
import { ref } from 'vue';
import {
	getAreaChartOps,
	getResult
} from '@apps/control-hub/src/utils/monitoring';
import { getMonitoringCfgs } from './config';
import MyLoading from '@apps/control-hub/src/components/MyLoading.vue';
import { date } from 'quasar';
const chartData = ref();
const selectValue = ref<DateRangeItem>(options[4]);
const loading = ref(false);
const list = ref();

const selecteChange = (value: DateRangeItem) => {
	selectValue.value = value;
	fetchData();
};

const fetchData = () => {
	const end = date.formatDate(Date.now(), 'X');
	const start = (Number(end) - selectValue.value.value).toString();

	const prams = {
		start,
		end,
		step: '120s',
		times: 50,
		metrics_filter:
			'cluster_cpu_utilisation|cluster_memory_utilisation|cluster_load1|cluster_load5|cluster_load15|cluster_disk_size_usage|cluster_disk_inode_utilisation|cluster_disk_inode_usage|cluster_disk_inode_total|cluster_disk_read_iops|cluster_disk_write_iops|cluster_disk_read_throughput|cluster_disk_write_throughput|cluster_net_bytes_transmitted|cluster_net_bytes_received|cluster_pod_running_count|cluster_pod_abnormal_count|cluster_pod_succeeded_count|pod_completed_count|$'
	};
	loading.value = true;
	getClusterMonitoring(prams)
		.then((res) => {
			const data = getResult(res.data.results);
			list.value = getMonitoringCfgs(data).map((item) => getAreaChartOps(item));
		})
		.finally(() => {
			loading.value = false;
		});
};

fetchData();
</script>

<style></style>
