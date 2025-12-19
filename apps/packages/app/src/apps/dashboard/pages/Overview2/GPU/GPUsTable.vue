<template>
	<QTableStyle2 sticky-last>
		<q-table
			:rows="GpuStore.gpuList"
			:columns="columns"
			row-key="uuid"
			:loading="loading"
			hide-pagination
			v-model:pagination="pagination"
			flat
			class="table-wrapper"
		>
			<template #body-cell-nodeUid="props">
				<q-td :props="props">
					<div style="width: 160px" class="ellipsis relative-position">
						{{ props.row.nodeUid }}
						<q-tooltip>
							{{ props.row.nodeUid }}
						</q-tooltip>
					</div>
				</q-td>
			</template>
			<template #body-cell-uuid="props">
				<q-td :props="props">
					{{ props.value }}
				</q-td>
			</template>

			<template #body-cell-health="props">
				<q-td :props="props">
					<GPUStatus
						:health="props.row.health"
						:is-external="props.row.isExternal"
					></GPUStatus>
				</q-td>
			</template>
			<template #body-cell-type="props">
				<q-td :props="props">
					<div style="width: 160px" class="ellipsis relative-position">
						{{ props.row.type }}
						<q-tooltip>
							{{ props.row.type }}
						</q-tooltip>
					</div>
				</q-td>
			</template>
			<template #body-cell-vgpu="props">
				<q-td :props="props">
					{{ props.row.isExternal ? '--' : props.row.vgpuUsed }}/{{
						props.row.isExternal ? '--' : props.row.vgpuTotal
					}}
				</q-td>
			</template>

			<template #body-cell-compute="props">
				<q-td :props="props">
					{{ props.row.isExternal ? '--' : props.row.coreUsed }}/{{
						props.row.coreTotal
					}}
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
import { Graphics, GraphicsListParams } from '@apps/dashboard/src/types/gpu';
import { getGraphicsList } from '@apps/dashboard/src/network/gpu';
import QTableStyle2 from '@apps/control-panel-common/src/components/QTableStyle2.vue';
import Empty from '@apps/control-panel-common/src/components/Empty.vue';
import { useRouter } from 'vue-router';
import { ROUTE_NAME } from '@apps/dashboard/src/router/const';
import { useI18n } from 'vue-i18n';
import GPUStatus from '@apps/dashboard/src/pages/Overview2/GPU/GPUStatus.vue';
import { useGpuStore } from '@apps/dashboard/src/stores/GpuStore';
import { VRAMModeLabel } from 'src/constant';
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
		name: 'nodeUid',
		label: t('GPU_OP.GRAPHICS_CARD_NODE'),
		field: 'nodeUid',
		align: 'left'
	},
	{
		name: 'type',
		label: t('GPU_OP.GRAPHICS_MODEL'),
		field: 'type',
		align: 'left'
	},
	{
		name: 'mode',
		align: 'left',
		label: t('GPU Mode'),
		field: 'shareMode',
		format: (val: any) => {
			return t(VRAMModeLabel[val]);
		}
	},
	{
		name: 'nodeName',
		label: t('GPU_OP.AFFILIATED_NODE'),
		field: 'nodeName',
		align: 'left'
	},
	{
		name: 'health',
		label: t('GPU_OP.GRAPHICS_STATUS'),
		field: 'health',
		align: 'left'
	},
	{
		name: 'coreUtilizedPercent',
		label: t('GPU_OP.CPU_R'),
		field: 'coreUtilizedPercent',
		format: (value) => `${round(value, 2)}%`,
		align: 'left'
	},
	{
		name: 'memoryTotal',
		label: t('GPU_OP.VIDEO_MEMORY_SIZE'),
		field: 'memoryTotal',
		format: (value) => getDiskSize(value * 1024 ** 2),
		align: 'left'
	},
	{
		name: 'memoryUtilizedPercent',
		label: t('GPU_OP.VRAM_USAGE_RATE'),
		field: 'memoryUtilizedPercent',
		format: (value) => `${round(value, 2)}%`,
		align: 'left'
	},
	{
		name: 'power',
		label: t('GPU_OP.GRAPHICS_CARD_POWER'),
		field: 'power',
		format: (value) => `${round(value, 2)}W`,
		align: 'left'
	},
	{
		name: 'temperature',
		label: t('GPU_OP.GRAPHICS_CARD_TEMP'),
		field: 'temperature',
		format: (value) => `${round(value, 2)}℃`,
		align: 'left'
	},
	{
		name: 'operations',
		label: t('OPERATIONS'),
		field: 'uuid',
		align: 'right'
	}
];

const fetchData = async (
	filters: GraphicsListParams['filters'] = {},
	init = false
) => {
	loading.value = true;
	try {
		const uid = uid_search.value;
		const params: GraphicsListParams = {
			filters,
			pageRequest: {
				sort: pagination.value.descending ? 'DESC' : 'ASC',
				sortField: 'id'
			}
		};

		const res = await getGraphicsList(params);

		const list = res.data.list;
		GpuStore.updateGpuList(list, init);
		pagination.value.rowsNumber = list.length;
	} finally {
		loading.value = false;
	}
};

const routeTo = (data: Graphics) => {
	router.push({
		name: ROUTE_NAME.GPUS_DETAILS,
		params: {
			uuid: data.uuid
		}
	});
};

onMounted(() => {
	fetchData({}, true);
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
