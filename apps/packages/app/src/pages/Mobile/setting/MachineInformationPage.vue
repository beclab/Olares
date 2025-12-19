<template>
	<div class="machine-information-root">
		<terminus-title-bar
			right-icon="sym_r_mode_off_on"
			@on-right-click="shutdownAction"
			:title="t('Device Information')"
		/>
		<terminus-scroll-area class="machine-scroll-area">
			<template
				v-slot:content
				v-if="mdnsStore.activedMachine && mdnsStore.activedMachine.status"
			>
				<div class="machine-top">
					<div class="row items-center justify-between">
						<div
							class="row items-center justify-center platform-logo item-border item-width"
						>
							<q-img
								:src="deviceLogo(mdnsStore.activedMachine.status)"
								width="80px"
								height="80px"
							/>
						</div>
						<div class="column item-width justify-between platform-logo">
							<div
								class="item-height item-border item-padding column justify-center"
							>
								<div class="text-subtitle2 text-ink-1">
									{{ t('Device name') }}
								</div>
								<div class="text-body3 text-ink-3">
									{{ mdnsStore.activedMachine.status?.device_name || '--' }}
								</div>
							</div>
							<div
								class="item-height item-border item-padding column justify-center"
							>
								<div class="text-subtitle2 text-ink-1">
									{{ t('Host name') }}
								</div>
								<div class="text-body3 text-ink-3">
									{{ mdnsStore.activedMachine.status.host_name || '--' }}
								</div>
							</div>
						</div>
					</div>
					<div class="row items-center justify-between item-margin-top">
						<div
							class="item-width item-padding-y item-border item-padding column justify-center"
						>
							<div class="text-subtitle2 text-ink-1">
								{{ t('CPU') }}
							</div>
							<div class="text-body3 text-ink-3">
								{{ mdnsStore.activedMachine.status.cpu_info || '--' }}
							</div>
						</div>
						<div
							class="item-width item-padding-y item-border item-padding column justify-center"
						>
							<div class="text-subtitle2 text-ink-1">
								{{ t('GPU') }}
							</div>
							<div class="text-body3 text-ink-3">
								{{ mdnsStore.activedMachine.status.gpu_info || '--' }}
							</div>
						</div>
					</div>

					<div class="row items-center justify-between item-margin-top">
						<div
							class="item-width item-padding item-padding-y item-border column justify-center"
						>
							<div class="text-subtitle2 text-ink-1">
								{{ t('Memroy') }}
							</div>
							<div class="text-body3 text-ink-3">
								{{ mdnsStore.activedMachine.status.memory || '--' }}
							</div>
						</div>
						<div
							class="item-width item-padding item-padding-y item-border column justify-center"
						>
							<div class="text-subtitle2 text-ink-1">
								{{ t('DISK') }}
							</div>
							<div class="text-body3 text-ink-3">
								{{ mdnsStore.activedMachine.status.disk || '--' }}
							</div>
						</div>
					</div>
				</div>
				<div class="text-h6 text-ink-1 q-mt-lg">
					{{ t('System') }}
				</div>
				<div class="item-padding item-padding-y item-border q-mt-sm">
					<div class="text-subtitle2 text-ink-1">
						{{ t('Kernel version') }}
					</div>
					<div class="text-body3 text-ink-3">
						{{
							mdnsStore.activedMachine.status.os_type +
							' ' +
							mdnsStore.activedMachine.status.os_version
						}}
					</div>
				</div>

				<div class="item-padding item-padding-y item-border item-margin-top">
					<div class="text-subtitle2 text-ink-1">
						{{ t('Kernel architecture') }}
					</div>
					<div class="text-body3 text-ink-3">
						{{ mdnsStore.activedMachine.status.os_arch || '--' }}
					</div>
				</div>

				<div class="item-padding item-padding-y item-border item-margin-top">
					<div class="text-subtitle2 text-ink-1">
						{{ t('Olares OS') }}
					</div>
					<div class="text-body3 text-ink-3">
						{{ mdnsStore.activedMachine.status.terminusVersion || '--' }}
					</div>
				</div>

				<div class="text-h6 text-ink-1 q-mt-lg">
					{{ t('Network') }}
				</div>
				<div class="item-padding item-padding-y item-border q-mt-sm">
					<div class="text-subtitle2 text-ink-1">
						{{ t('External IP') }}
					</div>
					<div class="text-body3 text-ink-3">
						{{ mdnsStore.activedMachine.status.externalIp || '--' }}
					</div>
				</div>

				<div class="item-padding item-padding-y item-border item-margin-top">
					<div class="text-subtitle2 text-ink-1">
						{{ t('Intranet IP') }}
					</div>
					<div class="text-body3 text-ink-3">
						{{ mdnsStore.activedMachine.status.hostIp || '--' }}
					</div>
				</div>

				<div class="item-padding item-padding-y item-border item-margin-top">
					<div class="text-subtitle2 text-ink-1">
						{{ t('Wired connection status') }}
					</div>
					<div class="text-body3 text-ink-3">
						{{
							mdnsStore.activedMachine.status.wiredConnected
								? 'Connected'
								: '--'
						}}
					</div>
				</div>

				<div class="item-padding item-padding-y item-border item-margin-top">
					<div class="text-subtitle2 text-ink-1">
						{{ t('Wi-Fi connection status') }}
					</div>
					<div class="text-body3 text-ink-3">
						{{
							mdnsStore.activedMachine?.status?.wifiConnected &&
							mdnsStore.activedMachine?.status?.wifiSSID
								? t('Connected to Wi-Fi {wifi}', {
										wifi: `'${mdnsStore.activedMachine.status.wifiSSID}'`
								  })
								: mdnsStore.activedMachine?.status?.wifiConnected
								? t('Connected')
								: '--'
						}}
					</div>
				</div>
			</template>
		</terminus-scroll-area>
	</div>
</template>
<script setup lang="ts">
import { onMounted, onUnmounted, watch } from 'vue';
import { useMDNSStore } from '../../../stores/mdns';

import TerminusTitleBar from '../../../components/common/TerminusTitleBar.vue';
import TerminusScrollArea from '../../../components/common/TerminusScrollArea.vue';
import { deviceLogo } from '../../../services/abstractions/mdns/service';

import LocalMachineOpeationDialog from '../../../components/wizard/LocalMachineOpeationDialog.vue';

import { useUserStore } from '../../../stores/user';
import { useI18n } from 'vue-i18n';
import { useQuasar } from 'quasar';
import { useRouter } from 'vue-router';
const mdnsStore = useMDNSStore();

const userStore = useUserStore();
mdnsStore.mdnsUsed = false;

const { t } = useI18n();
const $q = useQuasar();

const router = useRouter();

onMounted(() => {
	mdnsStore.startSearchMdnsService();
});

onUnmounted(() => {
	mdnsStore.stopSearchMdnsService();
	mdnsStore.mdnsMachines = [];
});

const shutdownAction = async () => {
	if (!mdnsStore.activedMachine) {
		return;
	}
	if (mdnsStore.checkContainerMode()) {
		return;
	}
	$q.dialog({
		component: LocalMachineOpeationDialog
	}).onOk(async (action: 'shutdown' | 'restart') => {
		if (!mdnsStore.activedMachine) {
			return;
		}
		let result = false;
		if (action == 'shutdown') {
			result = await mdnsStore.shutdownMachineTerminus(
				mdnsStore.activedMachine
			);
		} else {
			result = await mdnsStore.rebootMachineTerminus(mdnsStore.activedMachine);
		}
		if (result) {
			router.back();
		}
	});
};

watch(
	() => mdnsStore.mdnsMachines,
	() => {
		if (mdnsStore.mdnsMachines.length > 0) {
			const activedMachine = mdnsStore.mdnsMachines.find(
				(e) => e.status && e.status.terminusName == userStore.current_user?.name
			);
			if (
				!mdnsStore.activedMachine ||
				mdnsStore.activedMachine.host != activedMachine?.host
			) {
				mdnsStore.setActivedMachine(activedMachine, undefined, false);
			}
		}
	},
	{
		deep: true
	}
);
</script>

<style scoped lang="scss">
.machine-information-root {
	width: 100%;
	height: 100%;

	.machine-scroll-area {
		width: 100%;
		height: calc(100% - 56px);
		padding: 20px;
		.item-border {
			border: 1px solid $separator;
			border-radius: 12px;
		}
		.item-margin-top {
			margin-top: 12px;
		}

		.item-padding {
			padding: 0 20px;
		}

		.item-padding-y {
			padding-top: 12px;
			padding-bottom: 12px;
		}

		.machine-top {
			width: 100%;
			// background-color: red;

			.item-width {
				width: calc(50% - 6px);
			}

			.item-height {
				height: calc(50% - 6px);
			}

			.platform-logo {
				height: 140px;
			}
		}

		.domain-information {
			border: 1px solid $separator;
			border-radius: 12px;
			width: 100%;
			height: 64px;
		}
	}
}
</style>
