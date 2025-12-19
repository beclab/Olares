<template>
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
						<template v-slot:body-cell-name="props">
							<q-td :props="props">
								<div class="cronjob-list-name" @click="routeTo(props.row)">
									{{ props.value }}
								</div>
							</q-td>
						</template>
						<template v-slot:body-cell-status="props">
							<q-td :props="props">
								<div class="row items-center">
									<Status :type="lowerCase(props.value)"></Status>
									{{ $t(props.value) }}
								</div>
							</q-td>
						</template>
						<template v-slot:body-cell-uid="props">
							<q-td :props="props">
								<QButtonStyle>
									<q-btn
										flat
										dense
										icon="sym_r_reopen_window"
										@click="returnHandler(props.row)"
									/>
								</QButtonStyle>
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

			<MetadataVue :data="metadata"></MetadataVue>

			<MyCard4 square flat :title="$t('EVENT_PL')">
				<template #extra>
					<Refresh @click="fetchEvent"></Refresh>
				</template>
				<Event :data="evnetList"></Event>
			</MyCard4>
		</MyPage2>
		<q-inner-loading :showing="loading"> </q-inner-loading>
	</MyContentPage>
	<RouterViewTransition></RouterViewTransition>

	<Yaml2
		ref="yamlRef"
		:apiVersion="apiVersion.replace('apis/', '')"
		:name="detail.name"
		:namespace="detail.namespace"
		:originData="detail._originData"
		:module="module"
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
import { ref, watch, computed } from 'vue';
import MyContentPage from '@apps/control-hub/src/components/MyContentPage.vue';
import {
	getCornJobsDetail,
	getJobEvent,
	fetchListByK8s,
	jobRerun,
	deleteJob,
	toggleJob
} from '@apps/control-hub/src/network';
import {
	API_VERSIONS,
	MODULE_KIND_MAP
} from '@apps/control-hub/src/utils/constants';
import { useRoute, useRouter } from 'vue-router';
import { ObjectMapper } from '@apps/control-panel-common/src/utils/object.mapper';
import DetailPage from '@apps/control-panel-common/src/containers/DetailPage.vue';
import { get, isEmpty, orderBy, lowerCase } from 'lodash';
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
import { getJobStatus } from '@apps/control-hub/src/utils/status';
import Empty from '@apps/control-panel-common/src/components/Empty.vue';
import QButtonStyle from '@apps/control-panel-common/src/components/QButtonStyle.vue';
import { componentName } from '@apps/control-hub/src/router/const';
import RouterViewTransition from '@apps/control-panel-common/src/components/RouterViewTransition.vue';

const $q = useQuasar();
const { t } = useI18n();

const module = 'cronjobs';
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
const options = computed(() => [
	{
		label: detail.value.suspend ? t('START') : t('PAUSE'),
		value: 'pause',
		icon: detail.value.suspend ? 'sym_r_play_circle' : 'sym_r_pause_circle',
		onClick: () => {
			rerunHanlder();
		}
	},
	{
		label: t('EDIT_YAML'),
		value: 'edit',
		icon: 'sym_r_edit_square',
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
]);

const metadata = computed(() => get(detail.value, '_originData.metadata', {}));

const attrData = computed(() => getAttrs());
const pagination = ref({
	rowsNumber: 0
});
const getCronJobStatus = (spec) => {
	if (spec.suspend) {
		return 'Paused';
	}
	return 'running';
};

const columns = [
	{
		label: t('JOB'),
		name: 'name',
		field: 'name',
		align: 'left'
	},
	{
		label: t('STATUS'),
		name: 'status',
		field: (row) => getJobStatus(row),
		align: 'left'
	},
	{
		label: t('START_TIME'),
		name: 'startTime',
		field: (row) =>
			getLocalTime(row.status.startTime).format('YYYY-MM-DD HH:mm:ss'),
		align: 'left'
	},
	{
		label: t('END_TIME'),
		name: 'completionTime',
		field: (row) =>
			getLocalTime(row.status.completionTime).format('YYYY-MM-DD HH:mm:ss'),
		align: 'left'
	},
	{
		label: t('OPERATIONS'),
		name: 'uid',
		align: 'center'
	}
];
const getAttrs = () => {
	if (isEmpty(detail.value)) {
		return;
	}
	const { spec = {} } = detail.value || { spec: {} };
	const status = getCronJobStatus(spec);

	const { namespace }: Record<string, any> = route.params;

	return [
		{
			name: t('CLUSTER'),
			value: 'default'
		},
		{
			name: t('PROJECT'),
			value: namespace
		},
		{
			name: t('STATUS'),
			value: t(status)
		},
		{
			name: t('SCHEDULE'),
			value: spec.schedule
		},
		{
			name: t('MAXIMUM_DELAY'),
			value: spec.startingDeadlineSeconds
		},
		{
			name: t('SUCCESSFUL_JOBS_RETAINED'),
			value: spec.successfulJobsHistoryLimit
		},
		{
			name: t('FAILED_JOBS_RETAINED'),
			value: spec.failedJobsHistoryLimit
		},
		{
			name: t('CONCURRENCY_POLICY'),
			value:
				spec.concurrencyPolicy === 'Allow'
					? t('RUN_JOBS_CONCURRENTLY')
					: spec.concurrencyPolicy === 'Forbid'
					? t('SKIP_NEW_JOB')
					: t('SKIP_OLD_JOB')
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
	const { namespace, name }: Record<string, any> = route.params;
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
		const {
			data: { items }
		} = await fetchListByK8s({
			...params,
			cluster,
			namespace,
			selector,
			module: 'jobs'
		});

		tableList.value = items.map(ObjectMapper[module]);
	} catch (error) {
		//
	}

	tableLoading.value = false;
};

const hideHandler = () => {
	// currentYamlData.value = {};
};

const confirmHandler = async () => {
	const { namespace, name }: Record<string, any> = route.params;
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
	const { namespace, name }: Record<string, any> = route.params;
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
		const params2 = { spec: { suspend: !detail.value.suspend } };
		await toggleJob({ apiVersion, name, namespace }, params2);
		fetchData();
	} catch (error) {
		//
	}
	loading.value = false;
};

const returnHandler = async (row) => {
	console.log('row', row);
	// jobRerun()

	const { namespace, name }: Record<string, any> = row;
	const params = {};

	const module = 'jobs';
	const apiVersion = API_VERSIONS[module];
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
		loading.value = false;
	}
};

const routeTo = (row) => {
	console.log('row', row);

	router.push({
		name: componentName.JOB_DETAILS,
		params: {
			namespace: row.namespace,
			jobName: row.name,
			jobUid: row.uid
		}
	});
};

watch(
	() => route.params.jobUid,
	() => {
		fetchData();
	},
	{
		immediate: true
	}
);
</script>
<style lang="scss" scoped>
.cronjob-list-name {
	font-weight: 500;
	&:hover {
		color: $blue-6;
		cursor: pointer;
	}
}
</style>
