<template>
	<MyCard3 title="Usage Ranking" flat>
		<template #extra>
			<div class="row items-center q-gutter-x-md">
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
			</div>
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
			>
				<template v-slot:body="props">
					<q-tr :props="props" class="my-tr-wrapper">
						<q-td key="namespace">
							<div class="row items-center no-wrap my-table-column-namespace">
								<q-img :src="appIcon" width="32px" ratio="1" no-spinner />
								<router-link
									class="pod-item-link td-title q-ml-md text-subtitle2"
									:to="{
										path: `/applications/${props.row.namespace}/pods`,
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
								<div class="text-body3 text-grey-5 q-mt-xs">
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
								<div class="text-body3 text-grey-5 q-mt-xs">
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
								<div class="text-body3 text-grey-5 q-mt-xs">
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
import appIcon from '@apps/dashboard/assets/user-space.svg';
import { computed, onMounted, ref } from 'vue';
import { getNamespaces } from '@apps/dashboard/src/network';
import { get, isEmpty } from 'lodash';
import {
	NamespacesResponse,
	MonitoringResponseData,
	NamespacesParam
} from '@apps/dashboard/src/types/network';
import { getSuitableValue } from '@apps/dashboard/src/utils/monitoring';
import { saveAs } from 'file-saver';
import { getPodsList } from '@apps/control-panel-common/src/network';
import RouterViewTransition from '@apps/control-panel-common/src/components/RouterViewTransition.vue';
import MyCard3 from '@apps/control-panel-common/src/components/MyCard3.vue';
import Refresh from '@apps/control-panel-common/src/components/Refresh.vue';
import QInputStyle from '@apps/control-panel-common/src/components/QInputStyle.vue';
import { t } from '@apps/dashboard/src/boot/i18n';
import { useAppDetailStore } from '@apps/dashboard/src/stores/AppDetail';
import UserSelect from '@apps/dashboard/src/containers/UserSelect/IndexPage.vue';
import { selectValue } from '@apps/dashboard/src/containers/UserSelect/config';
import QTableStyle2 from '@apps/control-panel-common/src/components/QTableStyle2.vue';
import MyLoading2 from '@apps/control-panel-common/src/components/MyLoading2.vue';
import Empty3 from '@apps/control-panel-common/src/components/Empty3.vue';

const appDetail = useAppDetailStore();

const columns: any = [
	{
		name: 'namespace',
		label: 'Namespaces',
		align: 'left',
		field: 'namespace',
		required: true
	},
	{
		name: 'namespace_cpu_usage',
		label: 'CPU Usage',
		align: 'left',
		field: 'namespace_cpu_usage',
		sortable: true,
		sortOrder: 'da'
	},
	{
		name: 'namespace_memory_usage_wo_cache',
		label: 'Memory Usage',
		align: 'left',
		field: 'namespace_memory_usage_wo_cache',
		sortable: true,
		sortOrder: 'da'
	},
	{
		name: 'namespace_pod_count',
		label: 'Pods',
		align: 'left',
		field: 'namespace_pod_count',
		sortable: true,
		sortOrder: 'da'
	},
	{
		name: 'namespace_net_bytes_transmitted',
		label: 'Outbound Traffic',
		align: 'left',
		field: 'namespace_net_bytes_transmitted',
		sortable: true,
		sortOrder: 'da'
	},
	{
		name: 'namespace_net_bytes_received',
		label: 'Inbound Traffic',
		align: 'left',
		field: 'namespace_net_bytes_received',
		sortable: true,
		sortOrder: 'da'
	}
];

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
	color: #1f1814;
	font-weight: 500;
	text-decoration: none;
	&:hover {
		color: $primary;
	}
}
::v-deep(.my-table-header-wrapper) {
	color: #b2b0af;
	font-size: 12px;
	font-style: normal;
	font-weight: 400;
	line-height: 16px;
	background: #fff;
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
