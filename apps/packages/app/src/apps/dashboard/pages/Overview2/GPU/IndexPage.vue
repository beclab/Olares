<template>
	<div class="fixed-full z-top">
		<FullPageWithBack :title="$t('GPU_DETAILS')">
			<template #extra>
				<QButtonStyle>
					<q-btn
						class="q-pa-xs"
						dense
						icon="sym_r_refresh"
						color="ink-2"
						outline
						:disable="loading"
						@click="refreshHandler"
						narrow-indicator
					>
					</q-btn>
				</QButtonStyle>
			</template>

			<div class="gpu-tabs-wrapper q-pt-md q-mt-md">
				<q-tabs
					v-model="tab"
					align="left"
					content-class="tabs-content-wrapper"
					active-color="primary"
					:breakpoint="0"
					no-caps
					narrow-indicator
					@update:model-value="tabChangeHandler"
				>
					<q-tab :ripple="false" content-class="tabs-content-wrapper" :name="1">
						{{ $t('GPU_OP.GRAPHICS_MANAGEMENT') }}
					</q-tab>
					<q-tab :ripple="false" :name="2">{{
						$t('GPU_OP.TASK_MANAGEMENT')
					}}</q-tab>
				</q-tabs>
				<q-separator />
			</div>

			<div class="q-mt-xl my-tabs-panel-wrapper">
				<q-tab-panels v-model="tab" swipeable vertical>
					<q-tab-panel :name="1" class="q-pa-none">
						<GPUsTable ref="GPUsTableRef"></GPUsTable>
					</q-tab-panel>
					<q-tab-panel :name="2" class="q-pa-none">
						<TasksTable ref="TasksTableRef"></TasksTable>
					</q-tab-panel>
				</q-tab-panels>
			</div>
		</FullPageWithBack>
		<RouterViewTransition></RouterViewTransition>
	</div>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import FullPageWithBack from '@apps/control-panel-common/src/components/FullPageWithBack2.vue';
import QInputStyle from '@apps/control-panel-common/src/components/QInputStyle.vue';
import QButtonStyle from '@apps/control-panel-common/src/components/QButtonStyle.vue';
import GPUsTable from './GPUsTable.vue';
import TasksTable from './TasksTable.vue';
import RouterViewTransition from '@apps/control-panel-common/src/components/RouterViewTransition.vue';
import { useGpuStore } from '@apps/dashboard/src/stores/GpuStore';
import BtSelect from '@apps/control-panel-common/src/components/Select.vue';
import { useI18n } from 'vue-i18n';
import { TaskStatusOptions } from './config';
import QTableStyle2 from '@apps/control-panel-common/components/QTableStyle2.vue';

const GpuStore = useGpuStore();
const { t } = useI18n();

const loading = ref(false);
const gpuUid = ref();
const gpuType = ref();
const gpuNodeName = ref();

const taskName = ref();
const taskNodeName = ref();
const taskStatus = ref();
const taskDeviceId = ref();
const tab = ref(1);
const GPUsTableRef = ref();
const TasksTableRef = ref();

const searchGpu = async () => {
	const uid = gpuUid.value || undefined;
	const nodeName = gpuNodeName.value;
	const type = gpuType.value;
	const filter = {
		uid,
		nodeName,
		type
	};
	GPUsTableRef.value.search(filter);
};

const searchTask = () => {
	const name = taskName.value || undefined;
	const nodeName = taskNodeName.value;
	const status = taskStatus.value?.value;
	const deviceId = taskDeviceId.value?.value;
	const filter = {
		name,
		nodeName,
		status,
		deviceId
	};
	TasksTableRef.value.search(filter);
};

const refreshHandler = () => {
	if (tab.value === 1) {
		searchGpu();
	} else {
		searchTask();
	}
};

const clearForm = () => {
	clearGPUForm();
	clearTaskForm();
};

const clearGPUForm = () => {
	gpuUid.value = undefined;
	gpuNodeName.value = undefined;
	gpuType.value = undefined;
};

const clearTaskForm = () => {
	taskName.value = undefined;
	taskNodeName.value = undefined;
	taskStatus.value = undefined;
	taskDeviceId.value = undefined;
};

const tabChangeHandler = () => {
	clearForm();
};
</script>

<style lang="scss" scoped>
.q-table {
	background: white;
}
.gpu-tabs-wrapper {
	position: relative;
	font-size: 16px;
	::v-deep(.tabs-content-wrapper .q-tab__content) {
		padding: 16px 0;
	}

	::v-deep(.tabs-content-wrapper .q-tab) {
		padding: 0 12px;
	}
}
.my-tabs-panel-wrapper {
	::v-deep(.q-table th) {
		font-size: 14px;
	}
}
</style>
