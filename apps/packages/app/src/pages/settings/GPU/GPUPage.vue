<template>
	<page-title-component
		:show-back="currentGpu && gpuStore.gpuList.length > 1"
		:title="t(`home_menus.${MENU_TYPE.GPU.toLowerCase()}`)"
		:customBack="true"
		@onBackClick="backAction"
	/>
	<bt-scroll-area class="nav-height-scroll-area-conf">
		<AdaptiveLayout>
			<template v-slot:pc>
				<div v-if="currentGpu">
					<bt-list first>
						<bt-form-item
							:title="t('GPU Type')"
							:margin-top="false"
							:chevron-right="false"
							:widthSeparator="true"
							:data="currentGpu.type"
						/>
						<bt-form-item
							:title="t('Node')"
							:margin-top="false"
							:chevron-right="false"
							:widthSeparator="true"
							:data="currentGpu.nodeName"
						/>
						<bt-form-item
							:title="t('Video memory size')"
							:margin-top="false"
							:chevron-right="false"
							:widthSeparator="true"
							:data="gpuSize(currentGpu.devmem)"
						/>
						<bt-form-item
							:margin-top="false"
							:chevron-right="false"
							:widthSeparator="false"
						>
							<template v-slot:title>
								<div class="text-body1 q-mt-md">
									{{ t('GPU Mode') }}
								</div>
								<div
									class="text-ink-3 q-mt-sm q-mb-md text-body3"
									style="max-width: 80%; word-break: break-word"
								>
									{{ memoryOptions?.description }}
								</div>
							</template>
							<bt-select
								v-model="memoryMode"
								:options="VRAMModeOptions()"
								@update:modelValue="vramUpdate"
							/>
						</bt-form-item>
					</bt-list>
					<div class="row items-center justify-between q-mt-lg q-mb-sm">
						<div class="row justify-start items-center">
							<div class="text-h6 text-ink-1">
								{{ memoryOptions?.subTitle }}
							</div>
							<settings-tooltip
								v-if="memoryOptions?.subDesc"
								:description="memoryOptions?.subDesc"
							/>
						</div>
					</div>
					<div>
						<template v-if="lastMemoryMode == VRAMMode.Single">
							<SingleModeDetail
								:selectApps="selectApps"
								:availableApps="selectApplicationsOptions"
								:availableGpuList="gpuStore.gpuList"
								:currentGPU="currentGpu"
								@bind-app="addApp"
								@switch-app="switchAppAction"
								@unbind="cancelBind"
							/>
						</template>
						<template v-else-if="lastMemoryMode == VRAMMode.MemorySlicing">
							<MemorySlicingModeDetail
								:selectApps="selectApps"
								:availableApps="selectApplicationsOptions"
								:availableGpuList="gpuStore.gpuList"
								:currentGPU="currentGpu"
								@bind-app="addApp"
								@switch-app="switchAppAction"
								@editVRAM="editVRAM"
								@unbind="cancelBind"
							/>
						</template>
						<template v-else-if="lastMemoryMode == VRAMMode.TimeSlicing">
							<TimeSlicingModeDetail
								:selectApps="selectApps"
								:availableApps="selectApplicationsOptions"
								:availableGpuList="gpuStore.gpuList"
								@bind-app="addApp"
								@switch-app="switchAppAction"
								@editVRAM="editVRAM"
								@unbind="cancelBind"
								:currentGPU="currentGpu"
							/>
						</template>
					</div>
				</div>
				<div v-else-if="gpuStore.gpuList && gpuStore.gpuList.length > 1">
					<app-menu-feature :menu-type="MENU_TYPE.GPU" />
					<q-list class="q-pt-md q-list-class q-mt-md">
						<div class="column item-margin-left item-margin-right">
							<q-table
								tableHeaderStyle="height: 32px;"
								table-header-class="text-body3 text-ink-3"
								flat
								:bordered="false"
								:rows="gpuStore.gpuList"
								:columns="gupListColumns"
								row-key="id"
								hide-pagination
								hide-selected-banner
								hide-bottom
								:rowsPerPageOptions="[0]"
								:tableRowStyleFn="
									() => {
										return 'height: 64px';
									}
								"
							>
								<template v-slot:body-cell-type="props">
									<q-td :props="props" class="text-ink-1" no-hover>
										<div
											style="
												max-width: 120px;
												white-space: nowrap;
												overflow: hidden;
												text-overflow: ellipsis;
											"
										>
											{{ props.row.type }}
										</div>
									</q-td>
								</template>
								<template v-slot:body-cell-node="props">
									<q-td :props="props" class="text-ink-1" no-hover>
										{{ props.row.nodeName }}
									</q-td>
								</template>
								<template v-slot:body-cell-mode="props">
									<q-td :props="props" class="text-ink-1" no-hover>
										{{
											VRAMModeOptions().filter(
												(e) => e.value == props.row.sharemode
											)[0].label
										}}
									</q-td>
								</template>
								<template v-slot:body-cell-size="props">
									<q-td :props="props" class="text-ink-1" no-hover>
										{{
											format.formatFileSize(
												props.row.devmem * 1024 * 1024,
												0,
												' '
											)
										}}
									</q-td>
								</template>
								<template v-slot:body-cell-actions="props">
									<q-td
										:props="props"
										class="text-ink-2 row justify-end items-center"
										style="height: 64px"
										no-hover
									>
										<div
											class="detail-btn row justify-center items-center"
											@click="enterDetail(props.row)"
										>
											<q-icon
												size="20px"
												name="sym_r_keyboard_arrow_right"
												color="ink-1"
											/>
										</div>
									</q-td>
								</template>
							</q-table>
						</div>
					</q-list>
				</div>

				<app-menu-empty
					v-else
					:menu-type="MENU_TYPE.GPU"
					:title="t('No GPU found')"
				/>
			</template>
			<template v-slot:mobile>
				<div v-if="currentGpu">
					<bt-list>
						<bt-form-item
							:title="t('GPU Type')"
							:margin-top="false"
							:chevron-right="false"
							:widthSeparator="true"
							:data="currentGpu.type"
						/>
						<bt-form-item
							:title="t('Node')"
							:margin-top="false"
							:chevron-right="false"
							:widthSeparator="true"
							:data="currentGpu.nodeName"
						/>
						<bt-form-item
							:title="t('Video memory size')"
							:margin-top="false"
							:chevron-right="false"
							:widthSeparator="true"
							:data="gpuSize(currentGpu.devmem)"
						/>
						<bt-form-item
							:margin-top="false"
							:chevron-right="false"
							:widthSeparator="false"
						>
							<template v-slot:title>
								<div class="text-subtitle3-m q-mt-md">
									{{ t('GPU Mode') }}
								</div>
								<div
									class="text-ink-3 q-mt-sm q-mb-md text-body3"
									style="max-width: 80%; word-break: break-word"
								>
									{{ memoryOptions?.description }}
								</div>
							</template>
							<bt-select
								v-model="memoryMode"
								:options="VRAMModeOptions()"
								@update:modelValue="vramUpdate"
							/>
						</bt-form-item>
					</bt-list>
					<div class="row items-center justify-between q-mt-lg q-mb-sm">
						<div class="row justify-start items-center">
							<div class="text-h6 text-ink-1">
								{{ memoryOptions?.subTitle }}
							</div>
							<settings-tooltip
								v-if="memoryOptions?.subDesc"
								:description="memoryOptions?.subDesc"
							/>
						</div>
						<!-- <q-btn
							flat
							dense
							@click="addApp"
							v-if="lastMemoryMode != VRAMMode.Single"
						>
							<q-icon size="20px" name="sym_r_add" color="info" />
							<div class="text-info text-body2 q-ml-xs text-capitalize">
								{{ t('Add an application') }}
							</div>
						</q-btn>
						<q-btn
							v-else-if="
								selectGpu && selectGpu.apps && selectGpu.apps?.length > 0
							"
							flat
							dense
							@click="singleRemove"
						>
							<q-icon size="20px" name="sym_r_cancel" color="info" />
							<div class="text-info text-body2 q-ml-xs text-capitalize">
								{{ t('btn_unbind') }}
							</div>
						</q-btn> -->
					</div>
					<div>
						<template v-if="lastMemoryMode == VRAMMode.Single">
							<SingleModeDetail
								:selectApps="selectApps"
								:availableApps="selectApplicationsOptions"
								:availableGpuList="gpuStore.gpuList"
								:currentGPU="currentGpu"
								@bind-app="addApp"
								@switch-app="switchAppAction"
								@unbind="cancelBind"
							/>
						</template>
						<template v-else-if="lastMemoryMode == VRAMMode.MemorySlicing">
							<MemorySlicingModeDetail
								:selectApps="selectApps"
								:availableApps="selectApplicationsOptions"
								:availableGpuList="gpuStore.gpuList"
								:currentGPU="currentGpu"
								@bind-app="addApp"
								@switch-app="switchAppAction"
								@editVRAM="editVRAM"
								@unbind="cancelBind"
							/>
						</template>
						<template v-else-if="lastMemoryMode == VRAMMode.TimeSlicing">
							<TimeSlicingModeDetail
								:selectApps="selectApps"
								:availableApps="selectApplicationsOptions"
								:availableGpuList="gpuStore.gpuList"
								:currentGPU="currentGpu"
								@bind-app="addApp"
								@switch-app="switchAppAction"
								@editVRAM="editVRAM"
								@unbind="cancelBind"
							/>
						</template>
					</div>
				</div>
				<div v-else-if="gpuStore.gpuList && gpuStore.gpuList.length > 1">
					<app-menu-feature :menu-type="MENU_TYPE.GPU" />
					<bt-grid
						class="mobile-items-list"
						:repeat-count="2"
						v-for="(gpu, index) in gpuStore.gpuList"
						:key="index"
						:paddingY="12"
					>
						<template v-slot:title>
							<div
								class="text-subtitle3-m row justify-between items-center clickable-view q-mb-md"
							>
								<div>
									{{ gpu.type }}
								</div>
								<div
									class="detail-btn row justify-center items-center"
									@click="enterDetail(gpu)"
								>
									<q-icon
										size="20px"
										name="sym_r_keyboard_arrow_right"
										color="ink-1"
									/>
								</div>
							</div>
						</template>

						<template v-slot:grid>
							<bt-grid-item
								:label="t('Node')"
								mobileTitleClasses="text-body3-m"
								:value="gpu.nodeName"
							/>
							<bt-grid-item
								:label="t('Video memory size')"
								mobileTitleClasses="text-body3-m"
								:value="format.formatFileSize(gpu.devmem * 1024 * 1024, 0, ' ')"
							/>
							<bt-grid-item
								:label="t('GPU Mode')"
								mobileTitleClasses="text-body3-m"
								:value="
									VRAMModeOptions().filter((e) => e.value == gpu.sharemode)[0]
										?.label
								"
							/>
						</template>
					</bt-grid>
				</div>
				<app-menu-empty
					v-else
					:menu-type="MENU_TYPE.GPU"
					:title="t('No GPU found')"
				/>
			</template>
		</AdaptiveLayout>
	</bt-scroll-area>
</template>

<script setup lang="ts">
import ReminderDialogComponent from 'src/components/settings/ReminderDialogComponent.vue';
import PageTitleComponent from 'src/components/settings/PageTitleComponent.vue';
import AppSelect from 'src/pages/settings/Developer/pages/dialog/AppSelect.vue';
import AppMenuFeature from 'src/components/settings/AppMenuFeature.vue';
import EmptyComponent from 'src/components/settings/EmptyComponent.vue';
import AdaptiveLayout from 'src/components/settings/AdaptiveLayout.vue';
import BtFormItem from 'src/components/settings/base/BtFormItem.vue';
import BtGridItem from 'src/components/settings/base/BtGridItem.vue';
import AppMenuEmpty from 'src/components/settings/AppMenuEmpty.vue';
import BtSelect from 'src/components/settings/base/BtSelect.vue';
import BtGrid from 'src/components/settings/base/BtGrid.vue';
import EditAppGpuDialog from './EditAppGpuDialog.vue';
import { useApplicationStore } from 'src/stores/settings/application';
import { MENU_TYPE, VRAMMode, VRAMModeOptions } from 'src/constant';
import { GPUInfo, useGPUStore } from 'src/stores/settings/gpu';
import { useDeviceStore } from 'src/stores/settings/device';
import { computed, onMounted, ref, watch } from 'vue';
import { format } from 'src/utils/format';
import { useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { useQuasar } from 'quasar';
import BtList from 'src/components/settings/base/BtList.vue';
import SettingsTooltip from 'src/components/settings/base/SettingsTooltip.vue';
import { notifyWarning } from 'src/utils/settings/btNotify';
import SingleModeDetail from './SingleModeDetail.vue';
import MemorySlicingModeDetail from './MemorySlicingModeDetail.vue';
import TimeSlicingModeDetail from './TimeSlicingModeDetail.vue';

import { useGPU } from './gpu';

const { t } = useI18n();

const memoryMode = ref(VRAMMode.Single);
let lastMemoryMode = ref(VRAMMode.Single);
const $q = useQuasar();

const appSingleName = ref('');

const {
	selectApps,
	selectGpu,
	selectApplicationsOptions,
	gpuStore,
	currentGpu
} = useGPU();

onMounted(() => {
	gpuStore.getGpuList();
});

const memoryOptions = computed(() => {
	return VRAMModeOptions().find((e) => e.value == memoryMode.value);
});

const vramUpdate = (mode: VRAMMode) => {
	if (!selectGpu.value) {
		return;
	}
	const vramItem = VRAMModeOptions().find((e) => e.value == mode);
	if (!vramItem) {
		return;
	}
	$q.dialog({
		component: ReminderDialogComponent,
		componentProps: {
			title: t('Switch VRAM mode'),
			message: t('Are you sure you need to switch the VRAM mode to {mode}?', {
				mode: vramItem.label
			}),
			useCancel: true,
			confirmText: t('confirm'),
			cancelText: t('cancel')
		}
	})
		.onOk(async () => {
			$q.loading.show();
			await gpuStore.setGpuMode(mode, selectGpu.value!.id);
			if (mode != VRAMMode.Single) {
				appSingleName.value = '';
			}
			$q.loading.hide();
		})
		.onCancel(() => {
			memoryMode.value = lastMemoryMode.value;
		});
};

const gpuSize = (size: number) => {
	return format.formatFileSize(size * 1024 * 1024, 0, ' ');
};

watch(
	() => gpuStore.gpuList,
	() => {
		if (gpuStore.gpuList.length == 1) {
			selectGpu.value = gpuStore.gpuList[0];
		} else if (gpuStore.gpuList.length > 1 && selectGpu.value) {
			selectGpu.value = gpuStore.gpuList.find(
				(e) => e.id == selectGpu.value?.id
			);
		}
	},
	{
		immediate: true
	}
);

watch(
	() => selectGpu.value,
	() => {
		if (selectGpu.value) {
			memoryMode.value = selectGpu.value.sharemode;
			lastMemoryMode.value = selectGpu.value.sharemode;
			if (
				selectGpu.value.apps &&
				selectGpu.value.apps.length > 0 &&
				selectGpu.value.sharemode == VRAMMode.Single
			) {
				appSingleName.value = selectGpu.value.apps[0].appName;
			}
		}
	},
	{
		immediate: true
	}
);

const addApp = () => {
	if (selectApplicationsOptions.value.length == 0) {
		notifyWarning(t('No apps available yet'));
		return;
	}
	$q.dialog({
		component: EditAppGpuDialog,
		componentProps: {
			maxValue: currentGpu.value?.memoryAvailable,
			memoryInput: lastMemoryMode.value == VRAMMode.MemorySlicing,
			selectApplicationsOptions: selectApplicationsOptions.value
		}
	}).onOk(async (data: { app: string; memoryLimit: number }) => {
		await gpuStore.updateApplication(
			lastMemoryMode.value,
			selectGpu.value!.id,
			data.app,
			data.memoryLimit
		);
		if (lastMemoryMode.value != VRAMMode.Single) {
			appSingleName.value = '';
		}
	});
};

const cancelBind = async (app: string) => {
	$q.loading.show();
	await gpuStore.unbindApplication(
		lastMemoryMode.value,
		selectGpu.value!.id,
		app
	);
	$q.loading.hide();
};

const singleRemove = () => {
	cancelBind(selectGpu.value!.apps![0].appName);
	appSingleName.value = '';
};

const selectAppAction = (app: any) => {
	gpuStore.updateApplication(lastMemoryMode.value, selectGpu.value!.id, app);
};

const switchAppAction = () => {
	if (selectApplicationsOptions.value.length == 0) {
		notifyWarning(t('No apps available yet'));
		return;
	}
	$q.dialog({
		component: EditAppGpuDialog,
		componentProps: {
			maxValue: currentGpu.value?.memoryAvailable,
			memoryInput: lastMemoryMode.value == VRAMMode.MemorySlicing,
			selectApplicationsOptions: selectApplicationsOptions.value.map((e) => {
				return {
					...e,
					isDefault: e.value == appSingleName.value
				};
			})
		}
	}).onOk(async (data: { app: string; memoryLimit: number }) => {
		await gpuStore.updateApplication(
			lastMemoryMode.value,
			selectGpu.value!.id,
			data.app,
			data.memoryLimit
		);
		if (lastMemoryMode.value != VRAMMode.Single) {
			appSingleName.value = '';
		}
	});
};

const editVRAM = (app: string) => {
	const apps = selectApps.value.filter((e) => e.value == app);

	if (apps.length == 0) {
		notifyWarning(t('No apps available yet'));
		return;
	}
	$q.dialog({
		component: EditAppGpuDialog,
		componentProps: {
			maxValue: currentGpu.value?.memoryAvailable,
			memoryInput: lastMemoryMode.value == VRAMMode.MemorySlicing,
			memeryInit: Number(
				format.formatFileSize(apps[0].size, undefined, undefined, false)
			),
			selectApplicationsOptions: apps.map((e) => {
				return {
					label: e.app,
					value: e.value,
					icon: e.icon,
					state: e.state
				};
			}),
			title: t('Edit VRAM Allocatin')
		}
	}).onOk(async (data: { app: string; memoryLimit: number }) => {
		await gpuStore.updateApplication(
			lastMemoryMode.value,
			selectGpu.value!.id,
			data.app,
			data.memoryLimit
		);
		if (lastMemoryMode.value != VRAMMode.Single) {
			appSingleName.value = '';
		}
	});
};

const enterDetail = (info: GPUInfo) => {
	appSingleName.value = '';
	selectGpu.value = info;
};
const router = useRouter();
const backAction = () => {
	if (gpuStore.gpuList.length > 1 && selectGpu.value) {
		selectGpu.value = undefined;
	} else {
		router.back();
	}
};

const appColumns: any = computed(() => {
	if (lastMemoryMode.value != VRAMMode.MemorySlicing) {
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
				label: t('operations'),
				align: 'right',
				sortable: false
			}
		];
	}
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
			name: 'size',
			align: 'right',
			label: t('Video Memory'),
			field: 'size',
			format: (val: any) => {
				return format.humanStorageSize(val);
			},
			sortable: false
		},
		{
			name: 'actions',
			label: t('btn_unbind'),
			align: 'right',
			sortable: false
		}
	];
});

const gupListColumns: any = [
	{
		name: 'type',
		align: 'left',
		label: t('GPU Type'),
		field: 'type',
		format: (val: any) => {
			return val;
		},
		sortable: false
	},
	{
		name: 'node',
		align: 'left',
		label: t('Node'),
		field: 'nodeName',
		format: (val: any) => {
			return val;
		},
		sortable: false
	},
	{
		name: 'size',
		align: 'left',
		label: t('Video memory size'),
		field: 'devmem',
		format: (val: any) => {
			return format.formatFileSize(val * 1024 * 1024, 0, ' ');
		},
		sortable: false
	},
	{
		name: 'mode',
		align: 'left',
		label: t('GPU Mode'),
		field: 'sharemode',
		format: (val: any) => {
			return VRAMModeOptions().filter((e) => e.value == val)[0].label;
		},
		sortable: false
	},
	{
		name: 'actions',
		align: 'right',
		label: t('action'),
		sortable: false
	}
];
</script>

<style scoped lang="scss">
.list_section {
	min-height: 56px;
}

.upgradeNow {
	border: 1px solid $btn-stroke;
	padding: 8px 12px;
	border-radius: 8px;
	cursor: pointer;
	// background-color: $background-3;

	&:hover {
		background: $background-5;
	}
}

.loader {
	width: 1rem;
	height: 1rem;
	color: inherit;
	vertical-align: middle;
	border: 0.2em solid transparent;
	border-top-color: currentcolor;
	border-bottom-color: currentcolor;
	border-radius: 50%;
	position: relative;
	animation: 1s loader-30 linear infinite;
	display: inline-block;

	&:before,
	&:after {
		content: '';
		display: block;
		width: 0;
		height: 0;
		position: absolute;
		border: 0.2em solid transparent;
		border-bottom-color: currentcolor;
	}

	&:before {
		transform: rotate(135deg);
		right: -0.2em;
		top: -0.05em;
	}

	&:after {
		transform: rotate(-45deg);
		left: -0.2em;
		bottom: -0.05em;
	}
}

.detail-btn {
	border-radius: 8px;
	border: 1px solid $separator;
	cursor: pointer;
	// padding: 5px 8px;
	// text-decoration: none;
	height: 32px;
	width: 32px;

	.add-title {
		color: $ink-2;
	}
}

.detail-btn:hover {
	background-color: $background-3;
}

@keyframes loader-30 {
	0% {
		transform: rotate(0deg);
	}

	100% {
		transform: rotate(360deg);
	}
}
</style>
