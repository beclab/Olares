<template>
	<page-title-component
		:title="t('Manage Node')"
		:show-back="true"
		:custom-back="true"
		@onBackClick="backAction"
	/>
	<bt-scroll-area class="nav-height-scroll-area-conf">
		<div
			v-for="device in devices"
			:key="device.id"
			class="manage-device-card q-pa-lg q-mt-md"
		>
			<!-- device header + supportType selector -->
			<div
				class="manage-device-header"
				:class="
					deviceStore.isMobile
						? 'column items-stretch'
						: 'row justify-between items-center'
				"
			>
				<div class="row items-center no-wrap" style="min-width: 0">
					<div class="manage-device-icon row items-center justify-center">
						<q-icon name="sym_r_memory" size="24px" class="text-ink-2" />
					</div>
					<div class="column q-ml-md" style="min-width: 0">
						<span class="text-body1 text-ink-1 ellipsis">{{
							deviceName(device)
						}}</span>
						<div class="text-body3 text-ink-3 q-mt-xs">
							{{ device.cardModel }}
						</div>
					</div>
				</div>

				<div :class="{ 'q-mt-md': deviceStore.isMobile }">
					<compute-support-type-select
						:model-value="device.supportType"
						:available-support-types="device.availableSupportTypes"
						:select-key="device.id"
						:reset-key="selectResetKey"
						@update:model-value="(value) => onSelectSupportType(device, value)"
					/>
				</div>
			</div>

			<!-- bound apps -->
			<template v-if="store.deviceApps(device).length > 0">
				<q-separator class="full-width q-mt-lg" />
				<compute-bound-apps
					class="q-mt-lg"
					bg-color="bg-background-1"
					:apps="boundAppsView(device)"
					:collapsible="false"
					@unbind="onUnbind"
				/>
			</template>
		</div>
	</bt-scroll-area>
</template>

<script setup lang="ts">
import PageTitleComponent from 'src/components/settings/PageTitleComponent.vue';
import ReminderDialogComponent from 'src/components/settings/ReminderDialogComponent.vue';
import ComputeBoundApps from 'src/components/compute/ComputeBoundApps.vue';
import ComputeSupportTypeSelect from 'src/components/compute/ComputeSupportTypeSelect.vue';
import SwitchSupportTypeDialog from './SwitchSupportTypeDialog.vue';
import UnbindAppDialog from 'src/components/compute/UnbindAppDialog.vue';
import {
	ComputeResourceDevice,
	ComputeSupportTypeAppRef,
	acceleratorSupportTypeLabelKey,
	applyExclusiveBoundAppMemory,
	computeModeTitle
} from 'src/constant/compute';
import {
	AcceleratorAppView,
	useAcceleratorStore
} from 'src/stores/settings/accelerator';
import { useDeviceStore } from 'src/stores/settings/device';
import { computed, onMounted, ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { useQuasar } from 'quasar';

const { locale, t } = useI18n();
const route = useRoute();
const router = useRouter();
const $q = useQuasar();
const store = useAcceleratorStore();
const deviceStore = useDeviceStore();

const nodeName = computed(() => decodeURIComponent(String(route.params.node)));

const devices = computed<ComputeResourceDevice[]>(
	() => store.getNode(nodeName.value)?.devices || []
);

// Bumped to force the support-type dropdown to re-sync with the committed
// device.supportType when a mode switch is cancelled or fails (the dropdown
// otherwise keeps its optimistic local selection).
const selectResetKey = ref(0);

onMounted(() => {
	if (store.nodes.length === 0) {
		store.fetch();
	}
});

function backAction() {
	const parentRouteName = route.meta.parentRouteName as string | undefined;
	if (parentRouteName) {
		router.replace({ name: parentRouteName });
	} else {
		router.back();
	}
}

function deviceName(device: ComputeResourceDevice): string {
	return (
		(device.name as string) ||
		device.cardModel ||
		(device.mode ? computeModeTitle(String(device.mode)) : '') ||
		device.id
	);
}

// Exclusive mode = the whole card is dedicated to its app, so the bound app is
// shown as occupying the full card capacity (overriding the backend memory).
function boundAppsView(device: ComputeResourceDevice): AcceleratorAppView[] {
	return applyExclusiveBoundAppMemory(
		store.deviceApps(device),
		device.supportType,
		device.memory
	);
}

function onSelectSupportType(device: ComputeResourceDevice, value: string) {
	if (!value || value === device.supportType) {
		return;
	}
	const appNames = store.deviceApps(device).map((app) => app.title);

	// No bound apps -> nothing to unbind, switch directly without confirmation.
	if (appNames.length === 0) {
		performSwitchSupportType(device, value);
		return;
	}

	$q.dialog({
		component: SwitchSupportTypeDialog,
		componentProps: {
			modeLabel: t(acceleratorSupportTypeLabelKey(value)),
			appNames
		}
	})
		.onOk(() => {
			performSwitchSupportType(device, value);
		})
		.onCancel(() => {
			// Dismissed via cancel / close / backdrop: keep the original mode.
			selectResetKey.value++;
		});
}

async function performSwitchSupportType(
	device: ComputeResourceDevice,
	value: string
) {
	$q.loading.show();
	try {
		const result = await store.switchSupportType(
			device.nodeName || nodeName.value,
			device.id,
			value
		);
		if (result?.status === 'bound-apps-stop-blocked') {
			showBlocked(result.blockedApps ?? []);
			// Mode was not actually switched; revert the dropdown selection.
			selectResetKey.value++;
		}
	} catch (e) {
		console.error('switchSupportType failed:', e);
		// Switch failed; revert the dropdown to the committed mode.
		selectResetKey.value++;
	} finally {
		$q.loading.hide();
	}
}

function showBlocked(blockedApps: ComputeSupportTypeAppRef[]) {
	const lines = blockedApps
		.map((a) => {
			const title = store.resolveApp(a.appName).title;
			return `${title}: ${a.reason || t('Unable to switch mode')}`;
		})
		.join('<br/>');
	$q.dialog({
		component: ReminderDialogComponent,
		componentProps: {
			title: t('Unable to switch mode'),
			message: lines,
			useCancel: false,
			confirmText: t('confirm')
		}
	});
}

function onUnbind(app: AcceleratorAppView) {
	$q.dialog({
		component: UnbindAppDialog,
		componentProps: {
			title: t('Confirm unbinding', { app: app.title }),
			// Multi-card apps release every linked card at once; single-card apps
			// only release this one binding.
			messageKeypath: app.multiCard
				? 'All cards linked to this application will be unlinked, and the application will stop simultaneously. Confirm this operation for {app}?'
				: 'This application will be unlinked and will stop simultaneously. Confirm this operation for {app}?',
			appTitle: app.title,
			okText: t('Unbind')
		}
	}).onOk(async () => {
		$q.loading.show();
		try {
			await store.releaseApp(app.appName);
		} finally {
			$q.loading.hide();
		}
	});
}
</script>

<style scoped lang="scss">
.manage-device-card {
	border: 1px solid $separator;
	border-radius: 12px;
	background-color: $background-1;
}

.manage-device-icon {
	width: 40px;
	height: 40px;
	border-radius: 50%;
	background-color: $background-3;
	flex-shrink: 0;
}
</style>
