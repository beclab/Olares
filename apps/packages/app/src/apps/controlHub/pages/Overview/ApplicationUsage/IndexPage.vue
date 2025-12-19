<template>
	<my-card flat>
		<q-table
			:rows="rows"
			:columns="columns"
			color="primary"
			row-key="name"
			v-model:pagination="pagination"
			@request="onRequest"
			flat
			:loading="loading"
		>
			<template v-slot:top>
				<QSectionStyle>
					<q-select
						v-model="selectValue"
						@update:model-value="fetchData"
						:options="options"
						dense
						outlined
					/>
				</QSectionStyle>

				<q-btn
					outline
					dense
					color="grey-5"
					:style="{
						marginLeft: '8px',
						borderRadius: '8px',
						lineHeight: '32px',
						height: '32px',
						width: '32px'
					}"
					@click="sortBy"
				>
					<img
						:class="[
							pagination.sortBy === 'desc' ? 'app-usage-sort-transform' : ''
						]"
						:src="sortUpIcon"
						alt=""
					/>
				</q-btn>
				<q-space />
				<q-btn
					outline
					dense
					color="grey-5"
					padding="4px 12px"
					style="border-radius: 8px"
				>
					<span class="export-icon" @click="download('project.usage.rank.json')"
						>Export</span
					>
				</q-btn>
			</template>
			<template v-slot:body="props">
				<q-tr :props="props">
					<q-td key="namespace">
						<div class="row items-center">
							<img :src="appIcon" alt="" />
							<span style="margin-left: 6px">{{ props.row.namespace }}</span>
						</div>
					</q-td>
					<q-td key="namespace_cpu_usage">
						<div>
							<div>
								{{
									getSuitableValue(props.row.namespace_cpu_usage, 'cpu', '-')
								}}
							</div>
							<div class="usage-rangking-cell-sub">
								<span>{{ $t('QUOTA') }}:&nbsp;</span>
								<span>
									{{
										getSuitableValue(
											props.row.namespace_cpu_limit_hard,
											'cpu',
											'-'
										)
									}}</span
								>
							</div>
						</div>
					</q-td>
					<q-td key="namespace_memory_usage_wo_cache">
						{{
							getSuitableValue(
								props.row.namespace_memory_usage_wo_cache,
								'memory',
								'-'
							)
						}}
						<div>
							<div class="usage-rangking-cell-sub">
								<span>{{ $t('QUOTA') }}:&nbsp;</span>
								<span>{{
									getSuitableValue(
										props.row.namespace_memory_limit_hard,
										'memory',
										'-'
									)
								}}</span>
							</div>
						</div>
					</q-td>
					<q-td key="namespace_pod_count">
						<div>{{ props.row.namespace_pod_count || '-' }}</div>
						<div>
							<div class="usage-rangking-cell-sub">
								<span>{{ $t('QUOTA') }}:&nbsp;</span>
								<span>
									{{ props.row.namespace_pod_count_hard || '-' }}
								</span>
							</div>
						</div>
					</q-td>
					<q-td key="namespace_net_bytes_transmitted">
						{{
							getSuitableValue(
								props.row.namespace_net_bytes_transmitted,
								'bandwidth',
								'-'
							)
						}}
					</q-td>
					<q-td key="namespace_net_bytes_received">
						{{
							getSuitableValue(
								props.row.namespace_net_bytes_received,
								'bandwidth',
								'-'
							)
						}}
					</q-td>
				</q-tr>
			</template>
			<template v-slot:no-data>
				<Empty></Empty>
			</template>
		</q-table>
	</my-card>
</template>

<script setup lang="ts">
import MyCard from '@apps/control-hub/src/components/MyCard.vue';
import sortUpIcon from '@apps/control-hub/assets/sort-up.svg';
import appIcon from '@apps/control-hub/assets/app.svg';
import { ref } from 'vue';
import { getNamespaces } from '@apps/control-hub/src/network';
import { get, isEmpty } from 'lodash';
import {
	NamespacesResponse,
	MonitoringResponseData,
	NamespacesParam
} from '@apps/control-hub/src/types/network';
import { getSuitableValue } from '@apps/control-hub/src/utils/monitoring';
import { _capitalize } from '@apps/control-hub/src/utils/index';
import saveAs from 'file-saver';
import MyLoading from '@apps/control-hub/src/components/MyLoading.vue';
import QSectionStyle from '@apps/control-hub/src/components/QSectionStyle.vue';
import Empty from '@apps/control-panel-common/src/components/Empty2.vue';
import { useI18n } from 'vue-i18n';
const { t } = useI18n();
const options = [
	{
		label: t('SORT_BY_NODE_CPU_UTILISATION'),
		value: 'namespace_cpu_usage'
	},
	{
		label: t('SORT_BY_NODE_MEMORY_UTILISATION'),
		value: 'namespace_memory_usage_wo_cache'
	},
	{
		label: t('SORT_BY_NAMESPACE_POD_COUNT'),
		value: 'namespace_pod_count'
	},
	{
		label: t('SORT_BY_INBOUND_TRAFFIC'),
		value: 'namespace_net_bytes_received'
	},
	{
		label: t('SORT_BY_OUTBOUND_TRAFFIC'),
		value: 'namespace_net_bytes_transmitted'
	}
];
const selectValue = ref(options[0]);

const columns: any = [
	{
		name: 'namespace',
		label: 'Application',
		align: 'left',
		field: 'namespace'
	},
	{
		name: 'cpu',
		label: 'CPU Usage',
		align: 'left',
		field: 'namespace_cpu_usage'
	},
	{
		name: 'namespace_memory_usage_wo_cache',
		label: 'Memory Usage',
		align: 'left',
		field: 'namespace_memory_usage_wo_cache'
	},
	{
		name: 'namespace_pod_count',
		label: 'Pods',
		align: 'left',
		field: 'namespace_pod_count'
	},
	{
		name: 'namespace_net_bytes_transmitted',
		label: 'Outbound Traffic',
		align: 'left',
		field: 'namespace_net_bytes_transmitted'
	},
	{
		name: 'namespace',
		label: 'Inbound Traffic',
		align: 'left',
		field: 'namespace'
	}
];
const rows = ref([]);
const loading = ref(false);
const pagination = ref({
	sortBy: 'desc',
	// descending: false,
	page: 1,
	rowsPerPage: 20,
	rowsNumber: 0
});

const handleResult = (response: NamespacesResponse) => {
	const results: NamespacesResponse['results'] = get(response, 'results', []);

	if (isEmpty(results)) {
		return [];
	}

	return results.reduce((rankList: any, metric: any = {}) => {
		const sortName = metric.metric_name;
		const result: MonitoringResponseData['result'] =
			get(metric, 'data.result') || [];
		result.forEach((data, index) => {
			const item = rankList[index] || {};
			const value = get(data, 'value', []);
			const describeMsg = get(data, 'metric', {});
			item[sortName] = value[1];
			Object.assign(item, describeMsg);
			rankList[index] = item;
		});
		return rankList;
	}, []);
};

const sortBy = () => {
	if (pagination.value.sortBy === 'desc') {
		pagination.value.sortBy = 'asc';
	} else {
		pagination.value.sortBy = 'desc';
	}
	fetchData();
};
const metrics_filter =
	'namespace_memory_usage_wo_cache|namespace_memory_limit_hard|namespace_cpu_usage|namespace_cpu_limit_hard|namespace_pod_count|namespace_pod_count_hard|namespace_net_bytes_received|namespace_net_bytes_transmitted';

const onRequest = (props: {
	pagination: {
		sortBy: string;
		page: number;
		rowsPerPage: number;
	};
	filter?: any;
	getCellValue: (col: any, row: any) => any;
}) => {
	const { page, rowsPerPage, sortBy } = props.pagination;
	const params = {
		type: 'rank',
		metrics_filter,
		page: page,
		limit: rowsPerPage,
		sort_type: sortBy,
		sort_metric: selectValue.value.value
	};
	fetchNamespaces(params);
};

const fetchData = () => {
	const params = {
		type: 'rank',
		metrics_filter,
		page: pagination.value.page,
		limit: pagination.value.rowsPerPage,
		sort_type: pagination.value.sortBy,
		sort_metric: selectValue.value.value
	};
	fetchNamespaces(params);
};

const fetchNamespaces = (params: NamespacesParam) => {
	loading.value = true;
	getNamespaces(params)
		.then((res) => {
			pagination.value.rowsNumber = res.data.total_item;
			rows.value = handleResult(res.data);
			pagination.value.page = res.data.page;
		})
		.finally(() => {
			loading.value = false;
		});
};

const download = (fileName: string) => {
	const json = JSON.stringify(rows.value, null, 2);
	const blob = new Blob([json], {
		type: 'text/plain;charset=utf-8'
	});
	saveAs(blob, fileName);
};

fetchData();
</script>

<style lang="scss" scoped>
.app-usage-sort-transform {
	transform: rotateZ(180deg) rotateY(180deg);
}
.usage-rangking-cell-sub {
	font-weight: 400;
	font-size: 12px;
	line-height: 12px;
	color: #7e7e7e;
	margin-top: 4px;
}
.export-icon {
	color: #414141;
}
</style>
