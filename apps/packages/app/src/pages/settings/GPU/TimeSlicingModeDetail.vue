<template>
	<q-list
		:class="deviceStore.isMobile ? 'mobile-items-list' : 'q-list-class q-py-md'"
	>
		<div class="item-margin-left item-margin-right">
			<q-table
				tableHeaderStyle="height: 32px;"
				table-header-class="text-body3 text-ink-3"
				:tableRowStyleFn="
					() => {
						return 'height: 64px';
					}
				"
				flat
				:bordered="false"
				:rows="selectApps"
				:columns="appColumns"
				hide-pagination
				hide-selected-banner
				hide-bottom
				:rowsPerPageOptions="[0]"
			>
				<template v-slot:body-cell-appName="props">
					<q-td :props="props" no-hover>
						<ApplicationInfo
							:icon="props.row.icon"
							:state="props.row.state"
							:app="props.row.app"
						/>
					</q-td>
				</template>
				<template v-slot:body-cell-actions="props">
					<q-td
						:props="props"
						style="height: 64px"
						class="text-ink-2 row items-center justify-end"
						no-hover
					>
						<UnbindGPU
							v-if="unBindEnable(props.row.value)"
							:app="props.row.app"
							@un-bind-app="emit('unbind', props.row.value)"
						/>
						<SwitchGPU
							v-if="availableGpuList.length > 1"
							:currentGPU="currentGPU"
							:appName="props.row.value"
							:app="props.row.app"
						/>
					</q-td>
				</template>
			</q-table>
		</div>
		<EmptyApplication class="q-mt-md" v-if="selectApps.length == 0" />
	</q-list>
	<div
		class="full-width row justify-end q-mt-lg"
		v-if="availableApps.length > 0"
	>
		<q-btn
			dense
			class="bind-app q-px-md q-py-sm text-body3 text-ink-2 bg-background-1"
			:label="t('Bind App')"
			no-caps
			@click="bindApp"
		/>
	</div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import { computed } from 'vue';
import { useDeviceStore } from 'src/stores/settings/device';
import EmptyApplication from './EmptyApplication.vue';
import ApplicationInfo from './ApplicationInfo.vue';
import UnbindGPU from './Components/UnbindGPU.vue';
import SwitchGPU from './Components/SwitchGPU.vue';
import { GPUInfo } from 'src/stores/settings/gpu';
import { useGPUStore } from 'src/stores/settings/gpu';

interface Props {
	selectApps: {
		app: string;
		icon: string;
		size: number;
		value: string;
		state?: string;
	}[];
	availableApps: any[];
	availableGpuList: any[];
	currentGPU: GPUInfo;
}

const props = withDefaults(defineProps<Props>(), {
	selectApps: () => [],
	availableApps: () => [],
	availableGpuList: () => []
});

const { t } = useI18n();

const deviceStore = useDeviceStore();

const emit = defineEmits(['bindApp', 'switchApp', 'unbind', 'editVRAM']);

const gpuStore = useGPUStore();

const bindApp = () => {
	emit('bindApp');
};

const unBindApp = (app: string) => {
	emit('unbind', app);
};

const appColumns: any = computed(() => {
	return [
		{
			name: 'appName',
			align: 'left',
			label: t('application'),
			field: 'app',
			format: (val: any) => {
				return val;
			},
			sortable: false
		},

		{
			name: 'actions',
			label: t('Operation'),
			align: 'right',
			sortable: false
		}
	];
});

const unBindEnable = (appName: string) => {
	return (
		gpuStore.gpuList.filter(
			(e) => e.apps && e.apps.find((app) => app.appName == appName) != undefined
		).length > 1
	);
};
</script>

<style scoped lang="scss">
.bind-app {
	flex: 0 0 64;
	border: solid 1px $btn-stroke;
}

.detail-btn {
	cursor: pointer;
	height: 24px;
	width: 24px;
	color: $ink-2;
}
</style>
