<template>
	<MyCard3 :title="$t('USAGE_RANKING')" flat>
		<template #extra>
			<Refresh @click="fetchData"></Refresh>
			<UserSelect v-model="selectValue" @update:model-value="selectHandler">
			</UserSelect>

			<QInputStyle>
				<q-input
					v-model="name"
					type="search"
					outlined
					debounce="500"
					:placeholder="t('SEARCH')"
					clearable
					style="width: 240px"
					@update:model-value="fetchData"
				>
					<template v-slot:prepend>
						<q-icon name="search" size="xs" />
					</template>
				</q-input>
			</QInputStyle>
			<!-- <QSectionStyle class="q-ml-sm">
      <q-select
        v-model="nodeSelected"
        :options="options2"
        dense
        outlined
        v-if="nodes.length > 1"
      />
    </QSectionStyle> -->
		</template>
		<QTableStyle2>
			<q-table
				:rows="tableData"
				:columns="columns"
				color="primary"
				row-key="name"
				v-model:pagination="pagination"
				@request="onRequest"
				flat
				binary-state-sort
				:loading="loading"
				:rows-per-page-label="$t('RECORDS_PER_PAGE')"
			>
				<template v-slot:body="props">
					<q-tr :props="props" class="my-tr-wrapper">
						<q-td key="namespace">
							<div class="row items-center no-wrap my-table-column-namespace">
								<q-img :src="appIcon" width="32px" ratio="1" no-spinner />
								<router-link
									class="pod-item-link td-title q-ml-md text-subtitle2"
									:to="{
										path: `/namespace/${props.row.namespace}/pods`,
										query: {
											node:
												nodeSelected === DEFAULT_OPTION
													? undefined
													: nodeSelected
										}
									}"
								>
									<span style="margin-left: 6px">{{
										props.row.namespace
									}}</span>
								</router-link>
							</div>
						</q-td>
						<q-td key="namespace_cpu_usage">
							<div>
								<div class="td-title">
									{{
										getSuitableValue(props.row.namespace_cpu_usage, 'cpu', '-')
									}}
								</div>
								<div class="text-body3 text-ink-3 q-mt-xs">
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
								<div class="text-body3 text-ink-3 q-mt-xs">
									<span>{{ $t('QUOTA') }}&nbsp;</span>
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
								<div class="text-body3 text-ink-3 q-mt-xs">
									<span>{{ $t('QUOTA') }}&nbsp;</span>
									<span>
										{{ props.row.namespace_pod_count_hard || '-' }}
									</span>
								</div>
							</div>
						</q-td>
						<q-td key="namespace_net_bytes_transmitted" class="text-right">
							{{
								getSuitableValue(
									props.row.namespace_net_bytes_transmitted,
									'bandwidth',
									'-'
								)
							}}
						</q-td>
						<q-td key="namespace_net_bytes_received" class="text-right">
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
				<template #no-data>
					<div style="height: 560px">
						<Empty3 @click="fetchData" v-show="!loading"></Empty3>
					</div>
				</template>
				<template #loading>
					<MyLoading2> </MyLoading2>
				</template>
			</q-table>
		</QTableStyle2>
	</MyCard3>

	<RouterViewTransition></RouterViewTransition>
</template>

<script setup lang="ts">
import sortUpIcon from '../../assets/sort-up.svg';
import appIcon from '@apps/control-hub/assets/user-space.svg';
import { computed, nextTick, onMounted, ref } from 'vue';
import { getNamespaces } from '@apps/control-hub/src/network';
import { get, isEmpty, set, uniq } from 'lodash';
import {
	NamespacesResponse,
	MonitoringResponseData,
	NamespacesParam
} from '@apps/control-hub/src/types/network';
import { getSuitableValue } from '@apps/control-hub/src/utils/monitoring';
import { _capitalize } from '@apps/control-hub/src/utils/index';
import { saveAs } from 'file-saver';
import { getPodsList } from '@apps/control-panel-common/src/network';
import { useRoute } from 'vue-router';
import RouterViewTransition from '@apps/control-panel-common/src/components/RouterViewTransition.vue';
import SortButtom from '@apps/control-panel-common/src/components/SortButton.vue';
import MyCard3 from '@apps/control-panel-common/src/components/MyCard3.vue';
import Refresh from '@apps/control-panel-common/src/components/Refresh.vue';
import QInputStyle from '@apps/control-panel-common/src/components/QInputStyle.vue';
import QSectionStyle from '@apps/control-panel-common/src/components/QSectionStyle.vue';
import { useAppDetailStore } from '@apps/control-hub/src/stores/AppDetail';
import { GLOBAL_ROLE } from '@apps/control-panel-common/src/constant/user';
import UserSelect from '@apps/dashboard/src/containers/UserSelect/IndexPage.vue';
import { selectValue } from '@apps/dashboard/src/containers/UserSelect/config';
import QTableStyle2 from '@apps/control-panel-common/src/components/QTableStyle2.vue';
import MyLoading2 from '@apps/control-panel-common/src/components/MyLoading2.vue';
import Empty3 from '@apps/control-panel-common/src/components/Empty3.vue';
import { useI18n } from 'vue-i18n';
const { t } = useI18n();
const appDetail = useAppDetailStore();

const columns: any = computed(() => [
	{
		name: 'namespace',
		label: t('NAMESPACE'),
		align: 'left',
		field: 'namespace',
		required: true
	},
	{
		name: 'namespace_cpu_usage',
		label: t('CPU_USAGE'),
		align: 'left',
		field: 'namespace_cpu_usage',
		sortable: true,
		sortOrder: 'da'
	},
	{
		name: 'namespace_memory_usage_wo_cache',
		label: t('MEMORY_USAGE'),
		align: 'left',
		field: 'namespace_memory_usage_wo_cache',
		sortable: true,
		sortOrder: 'da'
	},
	{
		name: 'namespace_pod_count',
		label: t('PODS'),
		align: 'left',
		field: 'namespace_pod_count',
		sortable: true,
		sortOrder: 'da'
	},
	{
		name: 'namespace_net_bytes_transmitted',
		label: t('OUTBOUND_TRAFFIC'),
		align: 'right',
		field: 'namespace_net_bytes_transmitted',
		sortable: true,
		sortOrder: 'da'
	},
	{
		name: 'namespace_net_bytes_received',
		label: t('INBOUND_TRAFFIC'),
		align: 'right',
		field: 'namespace_net_bytes_received',
		sortable: true,
		sortOrder: 'da'
	}
]);

const DEFAULT_OPTION = t('ALL');
const options2 = ref([DEFAULT_OPTION]);
const nodeSelected = ref(options2.value[0]);
const namespaceInNode: any = ref([]);
const rows = ref([]);
const loading = ref(false);
const name = ref();

const pagination = ref({
	sortBy: 'namespace_cpu_usage',
	descending: true,
	page: 1,
	rowsPerPage: 10,
	rowsNumber: 20
});
const podList = ref();
const nodes = ref([]);
const tableData = computed(() => {
	if (loading.value) {
		return [];
	}
	const target = namespaceInNode.value.find(
		(item: any) => item.label === nodeSelected.value
	);
	if (nodeSelected.value === DEFAULT_OPTION) {
		return rows.value;
	} else {
		return rows.value.filter((item: any) =>
			target.value.includes(item.namespace)
		);
	}
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

const sortBy = (status: string) => {
	pagination.value.sortBy = status;
	fetchData();
};
const metrics_filter =
	'namespace_memory_usage_wo_cache|namespace_memory_limit_hard|namespace_cpu_usage|namespace_cpu_limit_hard|namespace_pod_count|namespace_pod_count_hard|namespace_net_bytes_received|namespace_net_bytes_transmitted';

const onRequest = (props: {
	pagination: {
		sortBy: string;
		page: number;
		rowsPerPage: number;
		descending: boolean;
	};
	filter?: any;
	getCellValue: (col: any, row: any) => any;
}) => {
	const { page, rowsPerPage, sortBy, descending } = props.pagination;
	const rowsPerPageNew =
		rowsPerPage === 0 ? pagination.value.rowsNumber : rowsPerPage;
	const params = {
		type: 'rank',
		metrics_filter,
		page: page,
		limit: rowsPerPageNew,
		sort_type: descending ? 'desc' : 'asc',
		sort_metric: sortBy,
		resources_filter: resources_filter(),
		name: name.value
	};
	pagination.value.page = page;
	pagination.value.rowsPerPage = rowsPerPageNew;
	pagination.value.sortBy = sortBy;
	pagination.value.descending = descending;
	fetchNamespaces(params);
};

const resources_filter = () => {
	const isAdmin = appDetail.isAdmin;
	const selected = !isAdmin ? appDetail.user.username : selectValue.value.value;
	return name.value
		? `.*(?:${selected}).*?(?:${name.value}).*|.*(?:${name.value}).*?(?:${selected}).*`
		: `.*${selected}.*`;
};

const selectHandler = () => {
	name.value = undefined;
	fetchData();
};
const fetchData = () => {
	console.log('testaaaa', name.value, selectValue.value, resources_filter());
	const params = {
		type: 'rank',
		metrics_filter,
		page: pagination.value.page,
		limit: pagination.value.rowsPerPage,
		sort_type: pagination.value.descending ? 'desc' : 'asc',
		sort_metric: pagination.value.sortBy,
		resources_filter: resources_filter(),
		name: name.value
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

const fetchPodsList = () => {
	const params = {
		sortBy: 'startTime'
	};
	getPodsList(params).then((res) => {
		podList.value = res.data.items;
		const nodesObj: any = [];
		let nodeName = '';
		let namespace = '';
		podList.value.forEach((item: any) => {
			nodeName = item.spec.nodeName;
			namespace = item.metadata.namespace;
			const nodeTarget = nodesObj.find(
				(child: any) => child.label === nodeName
			);
			if (nodeTarget) {
				const data: string[] = nodeTarget.value;
				const target = data.find((child) => child === namespace);
				!target && nodeTarget.value.push(namespace);
			} else {
				nodesObj.push({
					label: nodeName,
					value: [namespace]
				});
			}
		});
		console.log('test', nodesObj);
		namespaceInNode.value = nodesObj;
		nodes.value = nodesObj.map((item: any) => item.label);
		options2.value = [DEFAULT_OPTION, ...nodes.value];
	});
};

onMounted(() => {
	fetchData();
});

// fetchPodsList();
</script>

<style lang="scss" scoped>
.app-usage-sort-transform {
	transform: rotateZ(180deg) rotateY(180deg);
}
.my-table-column-namespace {
}
.td-title {
	color: $grey-10;
	font-weight: 400;
	line-height: 16px;
}

.export-icon {
	color: #414141;
}
.pod-item-link {
	color: $ink-1;
	font-weight: 500;
	text-decoration: none;
	&:hover {
		color: $blue-6;
	}
}
::v-deep(.my-table-header-wrapper) {
	font-size: 12px;
	font-style: normal;
	font-weight: 400;
	line-height: 16px;
}
::v-deep(.my-tr-wrapper td) {
	padding: 12px 8px 12px 0;
}
::v-deep(.my-table-header-wrapper th) {
	padding: 12px 8px 12px 0;
}
::v-deep(.q-table th.sorted .q-table__sort-icon) {
	color: $ink-2;
}
::v-deep(.q-table__card .q-table__bottom) {
	font-size: 12px;
	color: $ink-2;
}
</style>
