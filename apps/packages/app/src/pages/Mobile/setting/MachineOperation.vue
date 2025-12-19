<template>
	<div class="local-machine-root">
		<terminus-title-bar
			right-icon="sym_r_mode_off_on"
			:show-right-icon="!isScaning && mdnsStore.activedMachine != undefined"
			@on-right-click="shutdownAction"
			:title="t('Olares Settings')"
		/>
		<terminus-scroll-area
			class="local-machine-scroll"
			:style="{ '--bluetooth_height': '58px' }"
			v-if="isScaning || !mdnsStore.activedMachine"
		>
			<template v-slot:content>
				<div class="wizard-content">
					<scan-local-machine
						:reminder-content="
							mdnsStore.isShutdown
								? t('Olares is powered off.')
								: mdnsStore.isReboot
								? t(
										'Olares is restarting. Check again when the restart is finished.'
								  )
								: undefined
						"
						:title="
							!isScaning && !mdnsStore.activedMachine
								? t('Olares not found')
								: undefined
						"
						:icon="
							!isScaning && !mdnsStore.activedMachine
								? 'wizard/machine-not-found.svg'
								: undefined
						"
					/>
				</div>
			</template>
		</terminus-scroll-area>
		<terminus-scroll-area class="machines-scroll-area" v-else>
			<template v-slot:content>
				<div class="machine-info row justify-between" @click="enterInformation">
					<q-img
						:src="deviceLogo(mdnsStore.activedMachine.status)"
						width="80px"
						height="80px"
					/>
					<div class="machine-detail">
						<div
							class="text-h6 text-ink-1 q-pb-xs row items-center justify-between"
						>
							<div>
								{{ mdnsStore.activedMachine?.status?.device_name || '--' }}
							</div>
							<q-icon
								name="sym_r_keyboard_arrow_right"
								size="24px"
								color="ink-2"
							/>
						</div>
						<div class="text-body3 text-ink-2 q-pb-xs">
							{{
								t('System version') +
									': ' +
									mdnsStore.activedMachine?.status?.terminusVersion || '--'
							}}
						</div>
						<div class="text-body3 text-ink-2 q-pb-xs">
							{{
								t('Activation Time') +
								': ' +
								formattedDateToStr(
									mdnsStore.activedMachine?.status?.initializedTime || 0
								)
							}}
						</div>
						<div class="text-body3 text-ink-2 q-pb-xs">
							{{
								t('Run Time') +
								': ' +
								formattedDateCompareCurrentDate(
									mdnsStore.activedMachine?.status?.installedTime || 0
								)
							}}
						</div>
					</div>
				</div>
				<div
					class="machine_status q-mt-md"
					v-if="mdnsStore.activedMachine?.status"
				>
					<div class="text-subtitle2 text-ink-1">
						{{ t('Device status') }}
					</div>
					<div class="row items-center" style="margin-top: 4px">
						<olares-status
							:status="mdnsStore.activedMachine?.status"
							:node-size="8"
						/>
					</div>
				</div>

				<terminus-item class="q-mt-md" @click="copyFunc" :item-height="64">
					<template v-slot:title>
						<div class="text-subtitle2">
							{{ t('Domain information') }}
						</div>
					</template>
					<template v-slot:detail>
						<div class="text-body3 text-ink-3">
							{{ domain }}
						</div>
					</template>
				</terminus-item>

				<terminus-item
					img-bg-classes="bg-background-3"
					class="q-mt-md"
					icon-name="sym_r_wifi"
					:whole-picture-size="32"
					:item-height="64"
					@click="enterWifiConfig"
				>
					<template v-slot:title>
						<div class="text-subtitle1">
							{{ t('Wi-Fi Configuration') }}
						</div>
					</template>
					<template v-slot:side>
						<div class="row items-center justify-end">
							<div class="app-version text-body3 text-ink-3 q-pr-xs">
								{{
									mdnsStore.activedMachine?.status?.wifiConnected
										? t('ON')
										: t('OFF')
								}}
							</div>
							<q-icon
								v-if="
									mdnsStore.activedMachine.status?.terminusState ==
									TerminusStatusEnum.TerminusRunning
								"
								name="sym_r_keyboard_arrow_right"
								size="20px"
								color="ink-3"
							/>
						</div>
					</template>
				</terminus-item>

				<terminus-item
					img-bg-classes="bg-background-3"
					class="q-mt-md"
					icon-name="sym_r_keyboard_previous_language"
					:whole-picture-size="32"
					@click="enterRestore"
					:item-height="64"
				>
					<template v-slot:title>
						<div class="text-subtitle1">
							{{ t('Restore factory status') }}
						</div>
					</template>
					<template v-slot:side>
						<q-icon
							name="sym_r_keyboard_arrow_right"
							size="20px"
							color="ink-3"
						/>
					</template>
				</terminus-item>

				<!--				<terminus-item-->
				<!--					img-bg-classes="bg-background-3"-->
				<!--					class="q-mt-md"-->
				<!--					icon-name="sym_r_keyboard_previous_language"-->
				<!--					:whole-picture-size="32"-->
				<!--					@click="enterBlueToothConfigNetwork"-->
				<!--					:item-height="64"-->
				<!--				>-->
				<!--					<template v-slot:title>-->
				<!--						<div class="text-subtitle1">-->
				<!--							{{ t('Bluetooth network setup') }}-->
				<!--						</div>-->
				<!--					</template>-->
				<!--					<template v-slot:side>-->
				<!--						<q-icon-->
				<!--							name="sym_r_keyboard_arrow_right"-->
				<!--							size="20px"-->
				<!--							color="ink-3"-->
				<!--						/>-->
				<!--					</template>-->
				<!--				</terminus-item>-->

				<terminus-item
					v-if="userStore.current_user?.isLargeVersion12"
					img-bg-classes="bg-background-3"
					class="q-mt-md"
					icon-name="sym_r_computer_arrow_up"
					:whole-picture-size="32"
					@click="enterUpgrade"
					:item-height="64"
				>
					<template v-slot:title>
						<div class="text-subtitle1">
							{{ t('System update') }}
						</div>
					</template>
					<template v-slot:side>
						<div class="row items-center justify-end">
							<div
								v-if="mdnsStore.upgradeEnable"
								class="row itesm-center justify-center bg-negative text-white"
								style="width: 20px; height: 20px; border-radius: 10px"
							>
								1
							</div>
							<q-icon
								name="sym_r_keyboard_arrow_right"
								size="20px"
								color="ink-3"
							/>
						</div>
					</template>
				</terminus-item>
			</template>
		</terminus-scroll-area>
		<div
			class="bottom-view column justify-end"
			v-if="!isScaning && !mdnsStore.activedMachine"
		>
			<confirm-button
				:btn-title="t('Rescan Olares in LAN')"
				@onConfirm="rescanTerminus"
			/>
			<div
				class="bluetooth row items-center justify-center t text-body2"
				@click="enterBlueToothConfigNetwork"
			>
				{{ t('Bluetooth network setup') }}
			</div>
		</div>
	</div>
</template>
<script setup lang="ts">
import { onMounted, onUnmounted, ref, watch } from 'vue';
import { useMDNSStore } from '../../../stores/mdns';

import TerminusTitleBar from '../../../components/common/TerminusTitleBar.vue';
import TerminusScrollArea from '../../../components/common/TerminusScrollArea.vue';
import TerminusItem from '../../../components/common/TerminusItem.vue';
import LocalMachineOpeationDialog from '../../../components/wizard/LocalMachineOpeationDialog.vue';
import { getPlatform } from '@didvault/sdk/src/core';
import ConfirmButton from '../../../components/common/ConfirmButton.vue';
import OlaresStatus from '../../../components/base/OlaresStatus.vue';

import {
	displayStatus,
	deviceLogo,
	TerminusStatusEnum,
	UpgradeLevel
} from '../../../services/abstractions/mdns/service';
import { useUserStore } from '../../../stores/user';
import { useI18n } from 'vue-i18n';
import { useQuasar, date } from 'quasar';
import { useRouter } from 'vue-router';
import { notifySuccess } from '../../../utils/notifyRedefinedUtil';
import { formatDateToDMHM } from '../../../utils/utils';
import ScanLocalMachine from '../connect/activate/ScanLocalMachine.vue';

const isScaning = ref(false);

const mdnsStore = useMDNSStore();
mdnsStore.mdnsUsed = false;

const userStore = useUserStore();

const { t } = useI18n();
const $q = useQuasar();

const domain = ref(
	userStore.getModuleSever('desktop', undefined, undefined, false)
);

onMounted(() => {
	mdnsStore.startSearchMdnsService();
	if (!mdnsStore.activedMachine || !mdnsStore.activedMachine.isSettingsServer) {
		startFoundMachines();
	} else {
		updateUpgradeInfo();
	}
});

onUnmounted(() => {
	// if ($q.platform.is.nativeMobile) {
	mdnsStore.stopSearchMdnsService();
	// }
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
		if (action == 'shutdown') {
			await mdnsStore.shutdownMachineTerminus(mdnsStore.activedMachine);
		} else {
			await mdnsStore.rebootMachineTerminus(mdnsStore.activedMachine);
		}
	});
};

watch(
	() => mdnsStore.mdnsMachines,
	() => {
		if (mdnsStore.mdnsMachines.length > 0) {
			isScaning.value = false;
			const activedMachine = mdnsStore.mdnsMachines.find(
				(e) => e.status && e.status.terminusName == userStore.current_user?.name
			);
			if (
				!mdnsStore.activedMachine ||
				mdnsStore.activedMachine.host != activedMachine?.host
			) {
				mdnsStore.setActivedMachine(activedMachine, undefined);
			}
		}
	},
	{
		deep: true
	}
);

const router = useRouter();

const enterRestore = async () => {
	if (!(await userStore.unlockFirst())) {
		return;
	}
	router.push('/machine/restore');
};

const enterUpgrade = async () => {
	if (!(await userStore.unlockFirst())) {
		return;
	}
	router.push('/machine/upgrade');
};

const copyFunc = async () => {
	try {
		await getPlatform().setClipboard(domain.value);
		notifySuccess(t('copy_success'));
	} catch (error) {
		console.error(error.message);
	}
};

const enterWifiConfig = () => {
	if (
		mdnsStore.activedMachine?.status?.terminusState !=
		TerminusStatusEnum.TerminusRunning
	) {
		return;
	}
	if (mdnsStore.checkContainerMode()) {
		return;
	}
	router.push('/machine/wifi');
};

const enterInformation = () => {
	router.push('/machine/information');
};

const formattedDateToStr = (datetime: number) => {
	if (datetime <= 0) {
		return '--';
	}
	return date.formatDate(datetime * 1000, 'YYYY-MM-DD HH:mm:ss');
};

const formattedDateCompareCurrentDate = (datetime: number) => {
	if (datetime <= 0) {
		return '--';
	}
	const date = new Date(datetime * 1000);
	return formatDateToDMHM(new Date(), date);
};

let leftTime = 10;
let foundMachinesTimer: any = undefined;

const rescanTerminus = () => {
	mdnsStore.stopSearchMdnsService();
	mdnsStore.startSearchMdnsService();
	startFoundMachines();
};

const startFoundMachines = () => {
	if (foundMachinesTimer) {
		clearfoundMachinesTimer();
	}
	leftTime = 10;
	isScaning.value = true;

	foundMachinesTimer = setInterval(async () => {
		leftTime = leftTime - 1;
		if (leftTime == 0) {
			clearfoundMachinesTimer();
			isScaning.value = false;
		}
	}, 1 * 1000);
};

const clearfoundMachinesTimer = () => {
	if (foundMachinesTimer) {
		clearInterval(foundMachinesTimer);
		foundMachinesTimer = undefined;
	}
};

const enterBlueToothConfigNetwork = () => {
	router.push({ path: '/bluetooth/discover/machine' });
};

const updateUpgradeInfo = () => {
	if (mdnsStore.activedMachine && !mdnsStore.activedMachine?.version) {
		if (userStore.current_user?.isLargeVersion12) {
			mdnsStore
				.checkLastOsVersion(
					mdnsStore.activedMachine,
					mdnsStore.includeRC ? UpgradeLevel.RC : UpgradeLevel.NONE
				)
				.then((version: any) => {
					if (version) {
						mdnsStore!.activedMachine!.version = version;
					}
				});
		}
	}
};
</script>

<style scoped lang="scss">
.local-machine-root {
	width: 100%;
	height: 100%;

	.machines-scroll-area {
		width: 100%;
		height: calc(100% - 56px);
		padding: 20px;

		.machine-info {
			width: 100%;
			padding: 20px;
			// height: 152px;
			border-radius: 8px;
			border: 1px solid $separator;

			.machine-detail {
				width: calc(100% - 90px);
			}
		}
		.machine_status {
			border: 1px solid $separator;
			border-radius: 12px;
			padding: 12px 20px;

			.status-node {
				width: 12px;
				height: 12px;
				border-radius: 50%;
			}
		}

		.domain-information {
			border: 1px solid $separator;
			border-radius: 12px;
			width: 100%;
			height: 64px;
		}

		.upgradeNow {
			border: 1px solid $btn-stroke;
			padding: 8px 12px;
			border-radius: 8px;
			cursor: pointer;
			margin-top: 10px;
			display: inline-block;
		}
	}

	.local-machine-scroll {
		width: 100%;
		height: calc(100% - 56px - 48px - 48px - var(--bluetooth_height));
		.machines-list {
			padding-left: 20px;
			padding-right: 20px;
			padding-bottom: 20px;
			width: 100%;
		}
	}

	.bottom-view {
		width: 100%;
		padding-bottom: 48px;
		padding-left: 20px;
		padding-right: 20px;

		.bluetooth {
			height: 48px;
			margin-top: 10px;
			width: 100%;
			color: $light-blue-default;
		}

		.bluetooth:hover {
			// cursor: pointer;
			// background-color: red;
			color: $blue-default;
		}
	}
}
</style>
