<template>
	<div class="absolute-position">
		<MyContentPage>
			<template #extra>
				<div class="col-auto" v-permission>
					<QButtonStyle v-for="item in options" :key="item.value">
						<q-btn dense flat :icon="item.icon" @click="item.onClick">
							<q-tooltip>
								<div style="white-space: nowrap">
									{{ item.label }}
								</div>
							</q-tooltip>
						</q-btn>
					</QButtonStyle>
				</div>
			</template>
			<MyPage2>
				<MyCard4 square flat :title="$t('DETAILS')" :loading="loading">
					<DetailPage :data="attrData"></DetailPage>
				</MyCard4>
				<MyCard4 square flat :title="$t('RUN_RECORDS')">
					<template #extra>
						<Refresh @click="fetchRecords"></Refresh>
					</template>
					<QTableStyle>
						<q-table
							flat
							hide-pagination
							v-model:pagination="pagination"
							:rows="tableList"
							:columns="columns"
							:loading="tableLoading"
							row-key="name"
						>
							<template v-slot:body-cell-status="props">
								<q-td :props="props" class="row items-center">
									<Status :type="props.value"></Status>
									{{ $t(props.value.toUpperCase()) }}
									<span v-show="props.row.succeed">
										<span
											>({{ props.row.succeed }}/{{ props.row.desire }})</span
										>
									</span>
								</q-td>
							</template>
							<template #no-data>
								<div class="row justify-center full-width q-mt-lg">
									<Empty v-show="!tableLoading" @click="fetchRecords"></Empty>
								</div>
							</template>
						</q-table>
					</QTableStyle>
				</MyCard4>

				<MyCard4 no-content-gap square flat :title="t('PODS')">
					<PodContainer
						:detail="detail"
						ref="PodContainerRef"
						:route-prefix="`job-inner/${detail.name}/${detail.uid}`"
					></PodContainer>
				</MyCard4>

				<MetadataVue :data="metadata"></MetadataVue>

				<MyCard4
					no-content-gap
					square
					flat
					:title="t('ENVIRONMENT_VARIABLE_PL')"
				>
					<EnvironmentsLayout :detail="detail"></EnvironmentsLayout>
				</MyCard4>

				<MyCard4 square flat :title="$t('EVENT_PL')">
					<template #extra>
						<Refresh @click="fetchEvent"></Refresh>
					</template>
					<Event :data="evnetList"></Event>
					<q-inner-loading :showing="loading2"> </q-inner-loading>
				</MyCard4>
			</MyPage2>
			<q-inner-loading :showing="loading"> </q-inner-loading>
		</MyContentPage>
		<RouterViewTransition></RouterViewTransition>
	</div>

	<Yaml2
		ref="yamlRef"
		:apiVersion="apiVersion"
		:name="detail.name"
		:namespace="detail.namespace"
		:originData="detail._originData"
		readonly
		@hide="hideHandler"
		@change="fetchData"
	>
	</Yaml2>

	<DeleteDialog
		:title="`${$t('DELETE')} ${$t('JOB')}`"
		:desc="$t('JOB')"
		:name="detail.name"
		:loading="deleteLoading"
		ref="deleteDialogRef"
		@submit="confirmHandler"
		@hide="hideHandler"
	></DeleteDialog>
</template>

<script setup lang="ts">
import { ref, onMounted, watch, computed } from 'vue';
import EnvironmentVariables from '@apps/control-hub/src/containers/EnvironmentVariables.vue';
import Yaml from '@apps/control-hub/src/pages/NamespacePods/Yaml.vue';
import MyContentPage from '@apps/control-hub/src/components/MyContentPage.vue';
import MoreSelection from '@apps/control-panel-common/src/components/MoreSelection.vue';
import {
	getCornJobsDetail,
	getJobEvent,
	fetchJobRecords,
	jobRerun,
	deleteJob
} from '@apps/control-hub/src/network';
import {
	API_VERSIONS,
	MODULE_KIND_MAP
} from '@apps/control-hub/src/utils/constants';
import { useRoute, useRouter } from 'vue-router';
import { ObjectMapper } from '@apps/control-panel-common/src/utils/object.mapper';
import DetailPage from '@apps/control-panel-common/src/containers/DetailPage.vue';
import { get, isEmpty, orderBy } from 'lodash';
import MyPage2 from '@apps/control-panel-common/src/containers/MyPage2.vue';
import MyCard4 from '@apps/control-panel-common/src/components/MyCard2.vue';
import MetadataVue from '@apps/control-panel-common/src/containers/Metadata.vue';
import { getLocalTime, joinSelector } from '@apps/control-hub/src/utils';
import Event from '@apps/control-panel-common/src/containers/Event.vue';
import Refresh from '@apps/control-panel-common/src/components/Refresh.vue';
import QTableStyle from '@apps/control-panel-common/src/components/QTableStyle.vue';
import Status from '@apps/control-panel-common/src/components/Status.vue';
import Yaml2 from '../NamespacePods/Yaml2.vue';
import DeleteDialog from '@apps/control-panel-common/src/components/DeleteDialog.vue';
import { useQuasar } from 'quasar';
import { useI18n } from 'vue-i18n';
import PodContainer from '@apps/control-panel-common/src/containers/PodsList/PodContainer.vue';
import RouterViewTransition from '@apps/control-panel-common/src/components/RouterViewTransition.vue';
import { getJobStatus } from '@apps/control-hub/src/utils/status';
import EnvironmentsLayout from '@apps/control-panel-common/src/containers/EnvironmentsLayout.vue';
import QButtonStyle from '@apps/control-panel-common/src/components/QButtonStyle.vue';
import Empty from '@apps/control-panel-common/src/components/Empty.vue';

const $q = useQuasar();
const { t } = useI18n();

const module = 'jobs';
const apiVersion = API_VERSIONS[module];
const detail = ref<any>({});
const evnetList = ref();
const route = useRoute();
const router = useRouter();
const yamlRef = ref();
const loading = ref(false);
const loading2 = ref(false);
const tableList = ref([]);
const deleteDialogRef = ref();
const deleteLoading = ref(false);
const tableLoading = ref(false);
const options = [
	{
		label: t('RERUN'),
		value: 'refresh',
		icon: 'sym_r_reopen_window',
		onClick: () => {
			rerunHanlder();
		}
	},
	{
		label: t('VIEW_YAML'),
		value: 'edit',
		icon: 'sym_r_preview',
		onClick: () => {
			yamlRef.value.show();
		}
	},
	{
		label: t('DELETE'),
		value: 'delete',
		icon: 'sym_r_delete',
		onClick: () => {
			deleteDialogRef.value && deleteDialogRef.value.show();
		}
	}
];

const metadata = computed(() => get(detail.value, '_originData.metadata', {}));

const attrData = computed(() => getAttrs());
const pagination = ref({
	rowsNumber: 0
});

const columns = [
	{
		label: t('SN_NO'),
		name: 'id',
		field: 'id',
		align: 'left'
	},
	{
		label: t('STATUS'),
		name: 'status',
		field: 'status',
		align: 'left'
	},
	{
		label: t('MESSAGE'),
		name: 'messages',
		field: (row) => get(row, 'messages[0]') || '-',
		align: 'left'
	},
	{
		label: t('START_TIME'),
		name: 'start-time',
		field: (row) =>
			getLocalTime(row['start-time']).format('YYYY-MM-DD HH:mm:ss'),
		align: 'left'
	},
	{
		label: t('END_TIME'),
		name: 'completion-time',
		field: (row) =>
			getLocalTime(row['completion-time']).format('YYYY-MM-DD HH:mm:ss'),
		align: 'left'
	}
];
const getAttrs = () => {
	if (isEmpty(detail.value)) {
		return;
	}
	const { cluster = 'default', namespace }: Record<string, any> = route.params;

	const { spec = {} } = detail.value;

	const status = getJobStatus(detail.value);

	return [
		{
			name: t('CLUSTER'),
			value: cluster
		},
		{
			name: t('PROJECT'),
			value: namespace
		},
		{
			name: t('STATUS'),
			value: status
		},
		{
			name: t('MAXIMUM_RETRIES'),
			value: spec.backoffLimit
		},
		{
			name: t('COMPLETE_PODS'),
			value: spec.completions
		},
		{
			name: t('PARALLEL_PODS'),
			value: spec.parallelism
		},
		{
			name: t('MAXIMUM_DURATION'),
			value: spec.activeDeadlineSeconds
		},
		{
			name: t('CREATION_TIME_TCAP'),
			value: getLocalTime(detail.value.createTime).format('YYYY-MM-DD HH:mm:ss')
		},
		{
			name: t('CREATOR'),
			value: detail.value.creator
		}
	];
};

const fetchData = async () => {
	const { namespace, jobName: name }: Record<string, any> = route.params;
	const params = {};
	loading.value = true;
	try {
		const { data } = await getCornJobsDetail(
			apiVersion,
			namespace,
			module,
			name,
			params
		);
		detail.value = ObjectMapper[module](data);
		fetchEvent();
		fetchRecords();
	} catch (error) {
		//
	}
	loading.value = false;
};

const kind = MODULE_KIND_MAP[module];

const fetchEvent = async () => {
	const { uid, name, namespace, _originData = {} } = detail.value;

	const fields = {
		'involvedObject.name': name,
		'involvedObject.namespace': namespace,
		'involvedObject.kind': _originData.kind || kind,
		'involvedObject.uid': uid
	};
	loading2.value = true;
	try {
		const {
			data: { items }
		} = await getJobEvent({
			namespace,
			cluster: '',
			fieldSelector: joinSelector(fields)
		});
		evnetList.value = orderBy(items.map(ObjectMapper.events), 'lastTimestamp');
	} catch (error) {
		//
	}

	loading2.value = false;
};

const fetchRecords = async (params = {}) => {
	const { cluster, namespace, name } = detail.value;
	const selector = get(detail.value, 'spec.jobTemplate.metadata.labels', {});
	tableLoading.value = true;
	try {
		const { data } = await fetchJobRecords({
			...params,
			cluster,
			namespace,
			selector,
			module,
			name
		});

		const temp = ObjectMapper[module](data);

		const records = JSON.parse(get(temp, 'annotations.revisions', {}));
		tableList.value = Object.entries(records).map(([key, value]) => {
			const typedValue = value as Record<string, any>;

			return {
				...typedValue,
				id: key
			};
		});
	} catch (error) {
		//
	}
	tableLoading.value = false;
};

const hideHandler = () => {
	// currentYamlData.value = {};
};

const confirmHandler = async () => {
	const { namespace, jobName: name }: Record<string, any> = route.params;
	const params = {
		kind: 'DeleteOptions',
		apiVersion: 'v1',
		propagationPolicy: 'Background'
	};
	deleteLoading.value = true;
	try {
		const res = await deleteJob(apiVersion, namespace, module, name, params);
		$q.notify({
			type: 'positive',
			message: t('DELETED_SUCCESSFULLY')
		});

		deleteDialogRef.value && deleteDialogRef.value.hide();
		router.replace({
			path: '/jobs',
			query: {
				refresh: 1
			}
		});
	} catch (error) {
		$q.notify({
			type: 'negative',
			message: JSON.stringify(error || '')
		});
	}
	deleteLoading.value = false;
};

const rerunHanlder = async () => {
	const { namespace, jobName: name }: Record<string, any> = route.params;
	const params = {};
	loading.value = true;

	try {
		const { data } = await getCornJobsDetail(
			apiVersion,
			namespace,
			module,
			name,
			params
		);
		const resourceVersion = get(data, 'metadata.resourceVersion');

		await jobRerun(resourceVersion, { name, namespace });
		fetchData();
	} catch (error) {
		//
	}
	loading.value = false;
};

watch(
	() => route.params.jobName,
	() => {
		fetchData();
	},
	{
		immediate: true
	}
);
</script>
