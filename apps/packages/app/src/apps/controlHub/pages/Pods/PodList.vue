<template>
	<MyCard3 :title="$t('POD_LIST')">
		<template #extra>
			<div class="row items-center q-gutter-x-md">
				<Refresh @click="fetchData"></Refresh>
				<NodeSelect v-model="selectValue" @update:model-value="fetchData">
				</NodeSelect>
				<BtSelect
					style="min-width: 120px"
					v-model="statusValue"
					@update:model-value="fetchData"
					:options="statusOptions"
					dense
					outlined
				>
				</BtSelect>
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
		<QTableStyle2 sticky-first sticky-last>
			<q-table
				:rows="tableData"
				:columns="columns"
				color="primary"
				row-key="name"
				:pagination="pagination"
				flat
				binary-state-sort
				:loading="loading"
				:rows-per-page-label="$t('RECORDS_PER_PAGE')"
			>
				<template v-slot:body="props">
					<q-tr :props="props" class="text-body2">
						<q-td key="name" :props="props" class="text-subtitle2">
							<q-img
								:src="podIcon"
								:ratio="1"
								spinner-color="primary"
								width="32px"
							/>
							<router-link
								class="pod-item-link td-title q-ml-md text-subtitle2"
								:to="{
									path: `/${splitMenu.type}/pods/overview/${props.row.spec.nodeName}/${props.row.namespace}/${props.row.name}`
								}"
							>
								<span style="margin-left: 6px">{{ props.row.name }}</span>
							</router-link>
						</q-td>
						<q-td key="status" :props="props">
							<div class="row items-center no-wrap">
								<MyBadge :type="props.row.podStatus.type"></MyBadge>
								<span class="q-ml-sm">{{
									$t(props.row.podStatus.status.toUpperCase())
								}}</span>
							</div>
						</q-td>
						<q-td key="node" :props="props">
							{{ props.row.node }}({{ props.row.nodeIp }})
						</q-td>
						<q-td key="podIp" :props="props">
							{{ props.row.podIp || '-' }}
						</q-td>
						<q-td key="createTime" :props="props">
							{{ date.formatDate(props.row.createTime, 'YYYY-MM-DD HH:mm:ss') }}
						</q-td>
						<q-td key="operations" :props="props" style="width: 60px">
							<MoreSelection
								:options="options"
								@click="selectHandler(props.row)"
							></MoreSelection>
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
	<Yaml
		ref="yamlRef"
		:namespace="currentItem.namespace"
		:name="currentItem.name"
	></Yaml>
	<DeleteDialog
		:title="$t('DELETE_POD')"
		:desc="$t('POD')"
		:name="currentItem.name"
		:loading="deleteLoading"
		ref="deleteDialogRef"
		@submit="submitHandler"
		@hide="hideHandler"
	></DeleteDialog>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, reactive, ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { deletePod, getPodsList } from '@apps/control-hub/src/network';
import { get, isEmpty, omitBy, snakeCase, upperCase } from 'lodash';
import { UsePod } from '@apps/control-panel-common/src/stores/PodData';
import { useSplitMenu } from '@apps/control-panel-common/src/stores/menu';
import podIcon from '@apps/control-panel-common/src/assets/pod.svg';
import { ObjectMapper } from '@apps/control-panel-common/src/utils/object.mapper';
import MyBadge from '@apps/control-panel-common/src/components/MyBadge.vue';
import { date } from 'quasar';
import QTableStyle2 from '@apps/control-panel-common/src/components/QTableStyle2.vue';
import MyCard3 from '@apps/control-panel-common/src/components/MyCard3.vue';
import QInputStyle from '@apps/control-panel-common/src/components/QInputStyle.vue';
import Refresh from '@apps/control-panel-common/src/components/Refresh.vue';
import NodeSelect from '@apps/dashboard/src/containers/NodeSelect/IndexPage.vue';
import { selectValue } from '@apps/dashboard/src/containers/NodeSelect/config';
import Empty3 from '@apps/control-panel-common/src/components/Empty3.vue';
import MoreSelection from '@apps/control-panel-common/src/components/MoreSelection.vue';
import Yaml from '@apps/control-hub/src/pages/Pods/Yaml.vue';
import DeleteDialog from '@apps/control-panel-common/src/components/DeleteDialog.vue';
import { useQuasar } from 'quasar';
import MyLoading2 from '@apps/control-panel-common/src/components/MyLoading2.vue';
import { useI18n } from 'vue-i18n';
import QSectionStyle from '@apps/control-panel-common/src/components/QSectionStyle.vue';
import BtSelect from '@apps/control-panel-common/src/components/Select.vue';
import { useAppDetailStore } from '@apps/control-hub/stores/AppDetail';

const appDetailStore = useAppDetailStore();
const $q = useQuasar();
const { t } = useI18n();
const options = computed(() => {
	const option1 = [
		{
			label: t('VIEW_YAML'),
			value: 'yaml',
			icon: 'sym_r_preview',
			onClick: () => {
				yamlRef.value.show();
			}
		}
	];
	const option2 = [
		{
			label: t('DELETE'),
			value: 'delete',
			icon: 'sym_r_delete',
			onClick: () => {
				deleteDialogRef.value?.show();
			}
		}
	];

	return appDetailStore.isDemo ? option1 : option1.concat(option2);
});

const columns: any = [
	{
		name: 'name',
		label: t('NAME'),
		align: 'left',
		field: 'name'
	},
	{
		name: 'status',
		label: t('STATUS'),
		align: 'left',
		field: 'status'
	},
	{
		name: 'node',
		label: t('NODE'),
		align: 'left',
		field: 'node'
	},
	{
		name: 'podIp',
		label: t('POD_IP_ADDRESS_SCAP'),
		align: 'left',
		field: 'podIp'
	},
	{
		name: 'createTime',
		label: t('CREATION_TIME'),
		align: 'left',
		field: 'createTime',
		sortable: true,
		sortOrder: 'da'
	},
	{
		name: 'operations',
		label: t('OPERATIONS'),
		align: 'center',
		field: 'status'
	}
];

const initOptions = [
	{
		label: t('ALL_STATUS'),
		value: undefined
	},
	{
		label: t('WARNING'),
		value: 'Waiting'
	},
	{
		label: t('RUNNING'),
		value: 'Running'
	},
	{
		label: t('COMPLETED'),
		value: 'Completed'
	},
	{
		label: t('ERROR'),
		value: 'Error'
	}
];

interface Props {
	filter: Record<string, string>;
}

const props = withDefaults(defineProps<Props>(), {
	filter: {} as any
});

const idFormat = (namespace: any, name: any) =>
	namespace && name ? `${namespace}-${name}` : undefined;

const pagination = ref({
	sortBy: 'createTime',
	descending: true,
	page: 1,
	rowsPerPage: 10
});
const splitMenu = useSplitMenu();
const usePod = UsePod();
const route = useRoute();
const router = useRouter();
const name = ref();
const shouldExecuteResponseHandler = ref(true);
const yamlRef = ref();
const currentItem = ref<any>({});
const deleteDialogRef = ref();
const deleteLoading = ref(false);
const defaultOpeneds = ref([
	idFormat(route.params.namespace, route.params.name)
]);
const statusOptions = ref(initOptions);
const statusValue = ref(statusOptions.value[2]);

const containerId = (namespace: any, container: any) =>
	container ? `${namespace}-${container}` : undefined;
const defaultActive = ref(
	containerId(route.params.namespace, route.params.container)
);

const scrollToView = () => {
	const target = document.querySelector(`#${defaultOpeneds.value[0]}`);
	if (target) {
		target.scrollIntoView({ behavior: 'smooth', block: 'center' });
	}
};
const rows = ref();
let list = reactive<any>({
	items: [],
	totalItems: 0
});

const loading = ref(false);

const podData = ref();

const tableData = computed(() => {
	if (loading.value) {
		return [];
	}
	return rows.value;
});

function itemClick(id: string, target: any, routerLink = true) {
	const data = target.originData;
	podData.value = data;
	usePod.setDetail(data);
	if (routerLink) {
		router.push({
			path: `/${splitMenu.type}/pods/overview/${data.spec.nodeName}/${data.metadata.namespace}/${data.metadata.name}`
		});
		// nextTick(() => {
		//   myTreeRef.value && myTreeRef.value.setExpanded(id, true);
		// });
	}
}

const clearData = () => {
	list.items = [];
	rows.value = [];
};

const fetchData = () => {
	const params = {
		sortBy: pagination.value.sortBy,
		page: pagination.value.page,
		limit: -1,
		status: statusValue.value.value,
		ascending: !pagination.value.descending,
		name: name.value,
		nodeName: selectValue.value.value ? selectValue.value.value : undefined,
		...omitBy(props.filter, isEmpty)
	};
	clearData();
	loading.value = true;
	getPodsList(params)
		.then((res) => {
			if (!get(res, 'data.items')) {
				return;
			}
			const result = res.data.items.map((item) => ObjectMapper.pods(item));
			rows.value = result;
			list.items = result.map((item: any) => ({
				title: item.name,
				subTitle: item.namespace,
				id: idFormat(item.namespace, item.name),
				img: podIcon,
				originData: item,
				selectToExpend: true,
				status: item.podStatus.type,
				route: {
					path: `/${splitMenu.type}/pods/overview/${item.spec.nodeName}/${item.namespace}/${item.name}`
				}
				// children: tabList(item),
			}));

			if (shouldExecuteResponseHandler.value) {
				if (defaultOpeneds.value.length > 0) {
					const target = list.items.find(
						(item: any) =>
							item.id === idFormat(route.params.namespace, route.params.name)
					);
					target && itemClick('', target, false);
				} else {
					list.items[0] && itemClick('', list.items[0]);
					const value = idFormat(
						get(list.items[0], 'originData.metadata.namespace'),
						get(list.items[0], 'originData.metadata.name')
					);
					defaultActive.value = value;
				}
			}
		})
		.finally(() => {
			loading.value = false;
		});
};

const onRequest = (props: {
	pagination: {
		sortBy: string;
		descending: boolean;
		page: number;
		rowsPerPage: number;
	};

	filter?: string | any;
	getCellValue: (col: any, row: any) => any;
}) => {
	const { page, rowsPerPage, sortBy, descending } = props.pagination;
	const rowsPerPageNew =
		rowsPerPage === 0 ? pagination.value.rowsNumber : rowsPerPage;

	pagination.value.page = page;
	pagination.value.rowsPerPage = rowsPerPageNew;
	pagination.value.descending = descending;
	pagination.value.sortBy = sortBy;
	fetchData();
};

const selectHandler = (data: any) => {
	currentItem.value = data;
};

const submitHandler = async () => {
	try {
		const { data } = await deletePod(
			currentItem.value.namespace,
			currentItem.value.name
		);
		$q.notify({
			type: 'positive',
			message: 'Suscess'
		});
		fetchData();
	} catch (error: any) {
		$q.notify({
			type: 'negative',
			message: error
		});
	}
	deleteDialogRef.value?.hide();
	deleteLoading.value = false;
};
const hideHandler = () => {
	currentItem.value = {};
};
// watch(filterQuery, (newData, oldData) => {
//   if (!isEqual(newData, oldData)) {
//     defaultOpeneds.value = [];
//     defaultActive.value = undefined;
//     fetchData();
//   }
// });

onBeforeUnmount(() => {
	shouldExecuteResponseHandler.value = false;
});

fetchData();
</script>

<style lang="scss" scoped>
.my-nodes-container {
	padding: 0 16px;
	line-height: 44px;
	font-size: 12px;
	.my-tab-wrapper {
		font-family: 'Roboto';
		font-style: normal;
		font-weight: 500;
		font-size: 12px;
		line-height: 16px;
	}
}
.pod-item-link {
	color: $ink-1;
	font-weight: 500;
	text-decoration: none;
	&:hover {
		color: $blue-6;
	}
}
</style>
