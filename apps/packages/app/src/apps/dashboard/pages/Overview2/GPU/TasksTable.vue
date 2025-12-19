<template>
	<QTableStyle2 sticky-last>
		<q-table
			style="width: 100%"
			:rows="GpuStore.taskList"
			:columns="columns"
			row-key="uuid"
			:loading="loading"
			hide-pagination
			v-model:pagination="pagination"
			flat
			class="table-wrapper"
		>
			<template #body-cell-uuid="props">
				<q-td :props="props">
					{{ props.value }}
				</q-td>
			</template>

			<template #body-cell-status="props">
				<q-td :props="props">
					<TaskStatus :status="props.row.status"></TaskStatus>
				</q-td>
			</template>

			<template #body-cell-deviceIds="props">
				<q-td :props="props">
					<div
						class="row inline items-center q-py-xs q-px-md text-light-blue-default bg-light-blue-alpha"
						style="border-radius: 4px"
					>
						{{
							$t('GPU_OP.V_GPU_COUNT', { count: props.row.deviceIds.length })
						}}
						<q-tooltip>
							<div v-for="id in props.row.deviceIds" :key="id">{{ id }}</div>
						</q-tooltip>
					</div>
				</q-td>
			</template>

			<template #body-cell-allocatedMem="props">
				<q-td :props="props">
					{{ roundToDecimal(props.row.allocatedMem / 1024, 2) }} Gi
				</q-td>
			</template>

			<template #body-cell-createTime="props">
				<q-td :props="props">
					{{ timeParse(props.row.createTime) }}
				</q-td>
			</template>
			<template #body-cell-operations="props">
				<q-td :props="props">
					<span
						style="width: 68px"
						class="text-right text-body2 text-light-blue-default cursor-pointer"
						@click="routeTo(props.row)"
						>{{ $t('VIEW_DETAIL') }}</span
					>
				</q-td>
			</template>
			<template v-slot:no-data>
				<div class="row justify-center full-width q-mt-lg">
					<Empty v-show="!loading"></Empty>
				</div>
			</template>
			<template v-slot:loading>
				<q-inner-loading showing />
			</template>
		</q-table>
	</QTableStyle2>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import {
	TaskItemWithShareMode,
	TaskListParams
} from '@apps/dashboard/src/types/gpu';
import { getTaskList } from '@apps/dashboard/src/network/gpu';
import QTableStyle2 from '@apps/control-panel-common/src/components/QTableStyle2.vue';
import Empty from '@apps/control-panel-common/src/components/Empty.vue';
import { roundToDecimal, timeParse } from '@apps/dashboard/src/utils/gpu';
import { useRouter } from 'vue-router';
import { ROUTE_NAME } from '@apps/dashboard/src/router/const';
import { useI18n } from 'vue-i18n';
import TaskStatus from './TaskStatus.vue';
import { useGpuStore } from '@apps/dashboard/src/stores/GpuStore';
import { VRAMModeLabel, VRAMModeOptions } from 'src/constant';
import { round } from 'lodash';
import { getDiskSize } from '@apps/dashboard/src/utils/disk';
const GpuStore = useGpuStore();
const router = useRouter();
const { t } = useI18n();
const loading = ref(false);
const uid_search = ref();
const pagination = ref({
	sortBy: 'uuid',
	descending: true,
	rowsNumber: 0
});

const columns: any = [
	{
		name: 'name',
		label: t('GPU_OP.TASK_NAME'),
		field: 'name',
		align: 'left'
	},
	{
		name: 'status',
		label: t('GPU_OP.TASK_STATUS'),
		field: 'status',
		align: 'left'
	},
	{
		name: 'mode',
		align: 'left',
		label: t('GPU Mode'),
		field: 'deviceShareModes',
		format: (val: any) => {
			return t(VRAMModeLabel[val[0]]);
		},
		sortable: false
	},
	{
		name: 'nodeName',
		label: t('GPU_OP.AFFILIATED_NODE'),
		field: 'nodeName',
		align: 'left'
	},
	{
		name: 'devicesCoreUtilizedPercent',
		label: t('GPU_OP.CPU_R'),
		field: 'devicesCoreUtilizedPercent',
		format: (val: any) => {
			return `${round(val[0], 2)}%`;
		},
		align: 'left'
	},
	{
		name: 'devicesMemUtilized',
		label: t('GPU_OP.VIDEO_MEMORY_USAGE_OP'),
		field: 'devicesMemUtilized',
		format: (val: any) => {
			return val[0] ? `${round(val[0] / 1024, 2)}Gi` : '0Gi';
		},
		align: 'left'
	},
	{
		name: 'operations',
		label: t('OPERATIONS'),
		field: 'name',
		align: 'right'
	}
];

const fetchData = async (
	filters: TaskListParams['filters'] = {},
	init = false
) => {
	loading.value = true;
	try {
		const params: TaskListParams = {
			filters,
			pageRequest: {
				sort: pagination.value.descending ? 'DESC' : 'ASC',
				sortField: 'id'
			}
		};

		const res = await getTaskList(params);
		const items = res.data.items;
		GpuStore.updateTaskList(items, init);
		pagination.value.rowsNumber = items.length;
	} finally {
		loading.value = false;
	}
};

const routeTo = (data: TaskItemWithShareMode) => {
	router.push({
		name: ROUTE_NAME.TASKS_DETAILS,
		params: {
			name: data.name,
			pod_uid: data.podUid
		},
		query: {
			sharemode: data.deviceShareModes[0]
		}
	});
};

onMounted(() => {
	fetchData({}, true);
	GpuStore.getDeviceIds();
});

defineExpose({ search: fetchData });
</script>

<style lang="scss" scoped>
.q-table {
	background: white;
}
.table-wrapper {
	width: calc(100vw - 328px);
}
</style>
