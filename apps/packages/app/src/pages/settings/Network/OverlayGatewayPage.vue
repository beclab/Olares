<template>
	<page-title-component :show-back="true" :title="t('Overlay Gateway')" />

	<bt-scroll-area class="nav-height-scroll-area-conf">
		<!-- Header card -->

		<app-menu-feature
			image="settings/imgs/root/overlay-gateway.svg"
			:label="t('Overlay Gateway')"
			:description="
				t(
					'Assigns a dedicated LAN IP to apps using virtual network interface. Enables direct local network access for seamless screen mirroring, DLNA, and device discovery.'
				)
			"
		>
			<template
				v-if="
					(!overlayStore.isAvailable &&
						overlayStore.unavailableReasons.length > 0) ||
					overlayStore.errorMessage
				"
				#tips
			>
				<div
					style="border-radius: 8px"
					class="row justify-start text-orange-default q-mb-lg q-mx-lg bg-background-3 q-pa-md"
				>
					<q-icon name="sym_r_error" size="16px" />
					<div class="column justify-start q-ml-md text-subtitle3">
						<div class="text-bold">
							{{ t('Overlay Gateway Unavailable') }}
						</div>
						<div
							v-if="overlayStore.errorMessage"
							class="text-orange-default text-body3 q-mt-xs"
						>
							{{ overlayStore.errorMessage }}
						</div>
						<template v-else>
							<div
								v-for="reason in overlayStore.unavailableReasons"
								:key="reason.key"
								class="text-body3 q-mt-xs"
							>
								<span
									style="
										border-radius: 50%;
										width: 4px;
										height: 4px;
										display: inline-block;
									"
									class="q-mx-sm q-my-xs bg-orange-default"
								/>
								<span>{{ reason.label }}</span>
							</div>
						</template>
					</div>
				</div>
			</template>
		</app-menu-feature>

		<!-- Feature toggle -->
		<bt-list>
			<bt-form-item
				:title="t('Enable Overlay Gateway')"
				:width-separator="false"
			>
				<bt-loading
					v-if="overlayStore.isGatewayBusy"
					:loading="true"
					size="24px"
					class="q-mr-sm"
				/>
				<bt-switch
					v-else
					truthy-track-color="blue-default"
					:model-value="overlayStore.isEnabled"
					:disable="
						!meStore.isOwner ||
						!overlayStore.isAvailable ||
						overlayStore.isGatewayBusy
					"
					@update:model-value="onToggleGateway"
				/>
			</bt-form-item>
		</bt-list>

		<div v-if="overlayStore.isAvailable && overlayStore.isEnabled">
			<!-- Empty state -->
			<bt-list v-if="overlayStore.apps.length === 0" :label="t('Applications')">
				<div
					class="overlay-gateway-page__empty column items-center justify-center q-mx-lg"
				>
					<div
						class="overlay-gateway-page__empty-icon column justify-center items-center"
					>
						<q-icon name="sym_r_action_key" size="20px" color="ink-1" />
					</div>
					<div class="text-body1 text-ink-1 q-mt-md">
						{{ t('No apps found') }}
					</div>
					<div class="text-body2 text-ink-3 q-mt-md text-center">
						{{ t('None of your installed apps have overlay permission') }}
					</div>
				</div>
			</bt-list>
			<template v-else>
				<overlay-app-item
					v-for="(app, index) in overlayStore.apps"
					:key="app.name"
					:app="app"
					:index="index"
					@toggle="onToggleApp"
				/>
			</template>
		</div>
	</bt-scroll-area>
</template>

<script lang="ts" setup>
import BtLoading from 'src/components/base/BtLoading.vue';
import BtList from 'src/components/settings/base/BtList.vue';
import BtFormItem from 'src/components/settings/base/BtFormItem.vue';
import AppMenuFeature from 'src/components/settings/AppMenuFeature.vue';
import OverlayAppItem from 'src/components/settings/network/OverlayAppItem.vue';
import PageTitleComponent from 'src/components/settings/PageTitleComponent.vue';
import OverlayConfirmDialog from 'src/components/settings/network/OverlayConfirmDialog.vue';
import { onBeforeUnmount, onMounted, ref, watch } from 'vue';
import { APP_STATUS } from 'src/constant/constants';
import { notifyWarning } from 'src/utils/settings/btNotify';
import { useMeStore } from 'src/stores/settings/me';
import { useI18n } from 'vue-i18n';
import { useQuasar } from 'quasar';
import {
	OVERLAY_GATEWAY_POLL_MS,
	useOverlayGatewayStore
} from 'src/stores/settings/overlayGateway';

const { t } = useI18n();
const $q = useQuasar();
const meStore = useMeStore();
const overlayStore = useOverlayGatewayStore();

let pollTimer: ReturnType<typeof setInterval> | null = null;
const hasShownTransientTip = ref(false);

const RESTARTING_TIP_KEY =
	'After restarting, the system will be unavailable for a short period. Please wait patiently.';

async function pollOverlayStatus() {
	const result = await overlayStore.fetchStatus();

	if (result.ok) {
		hasShownTransientTip.value = false;
	} else if (result.transient) {
		if (!hasShownTransientTip.value) {
			notifyWarning(t(RESTARTING_TIP_KEY));
			hasShownTransientTip.value = true;
		}
	} else {
		console.error('[overlayGateway] poll failed:', result.message);
	}
}

function startOverlayPolling() {
	if (pollTimer) {
		return;
	}
	pollTimer = setInterval(() => {
		void pollOverlayStatus();
	}, OVERLAY_GATEWAY_POLL_MS);
}

function stopOverlayPolling() {
	if (pollTimer) {
		clearInterval(pollTimer);
		pollTimer = null;
	}
}

// Start/stop polling purely based on store state, so it reacts the moment a
// gateway/app toggle registers a pending sync target (no timing gaps).
watch(
	() => overlayStore.needsStatusPolling,
	(need) => {
		if (need) {
			startOverlayPolling();
		} else {
			hasShownTransientTip.value = false;
			stopOverlayPolling();
		}
	},
	{ immediate: true }
);

onMounted(() => {
	void overlayStore.fetchStatus();
});

onBeforeUnmount(() => {
	stopOverlayPolling();
});

function onToggleGateway(value: boolean) {
	void overlayStore.toggleGateway(value);
}

function onToggleApp(name: string, enable: boolean) {
	const app = overlayStore.apps.find((a) => a.name === name);
	if (!app) return;

	const isStopped = app.status === APP_STATUS.STOP.COMPLETED;

	// Stopped app + disable overlay: no confirmation dialog and no restart.
	if (isStopped && !enable) {
		void overlayStore.toggleAppOverlay(name, enable, $q);
		return;
	}

	$q.dialog({
		component: OverlayConfirmDialog,
		componentProps: {
			action: enable ? 'enable' : 'disable',
			appTitle: app.title,
			isStopped
		}
	}).onOk(() => {
		void overlayStore.toggleAppOverlay(name, enable, $q);
	});
}
</script>

<style scoped lang="scss">
.overlay-gateway-page {
	&__header-icon {
		width: 40px;
		height: 40px;
		flex-shrink: 0;
	}

	&__header-content {
		flex: 1;
		min-width: 0;
	}

	&__empty {
		padding-bottom: 50px;
		padding-top: 50px;
	}

	&__empty-icon {
		width: 40px;
		height: 40px;
		background-color: $background-3;
		border-radius: 50%;
	}
}
</style>
