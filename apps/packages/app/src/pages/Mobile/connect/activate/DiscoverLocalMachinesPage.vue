<template>
	<div class="local-machine-root">
		<terminus-title-bar
			:title="isScaning || isNotFound ? '' : t('Activate Olares')"
		/>

		<terminus-scroll-area
			class="local-machine-scroll"
			:style="{ '--bluetooth_height': $q.platform.is.ios ? '0px' : '58px' }"
		>
			<template v-slot:content>
				<div v-if="isScaning || isNotFound" class="wizard-content">
					<scan-local-machine
						:title="isNotFound ? t('Olares not found') : undefined"
						:icon="isNotFound ? 'wizard/machine-not-found.svg' : undefined"
						:reminder-content="
							isNotFound
								? t(
										'No Olares device pending activation found. Please ensure your phone and the device to be activated are on the same network.'
								  )
								: undefined
						"
					/>
				</div>
				<dir
					v-else-if="mdnsStore.discoverPageDisplayMachines.length == 0"
					class="wizard-content"
				>
					<scan-local-machine
						:title="t('Olares not found')"
						:icon="'wizard/machine-not-found.svg'"
						:reminder-content="
							t(
								'No Olares device pending activation found. Please ensure your phone and the device to be activated are on the same network.'
							)
						"
					/>
				</dir>
				<q-list class="machines-list" v-else>
					<local-machine-item
						v-for="machine in mdnsStore.discoverPageDisplayMachines"
						:key="machine.host"
						:machine="machine"
						@install-action="installAction"
						@active-action="activeLocal"
						@uninstall-action="uninstallAction"
					/>
				</q-list>
			</template>
		</terminus-scroll-area>

		<div class="bottom-view column justify-end" v-if="!isScaning">
			<confirm-button
				:btn-title="t('Rescan Olares in LAN')"
				:btnStatus="
					mdnsStore.discoverPageDisplayMachines.length > 0 &&
					mdnsStore.isInstalled
						? ConfirmButtonStatus.disable
						: ConfirmButtonStatus.normal
				"
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
import { useMDNSStore } from '../../../../stores/mdns';
import { TerminusServiceInfo } from '../../../../services/abstractions/mdns/service';
import { useUserStore } from '../../../../stores/user';
import { WizardInfo, SystemOption } from 'src/utils/interface/wizard';
import axios from 'axios';
import { notifyFailed } from '../../../../utils/notifyRedefinedUtil';
import { useQuasar } from 'quasar';
import { userBindTerminus } from '../../../../utils/BindTerminusBusiness';
import { useRouter } from 'vue-router';
import TerminusTitleBar from '../../../../components/common/TerminusTitleBar.vue';
import TerminusScrollArea from '../../../../components/common/TerminusScrollArea.vue';
import ConfirmButton from '../../../../components/common/ConfirmButton.vue';
import ScanLocalMachine from './ScanLocalMachine.vue';
import { useI18n } from 'vue-i18n';
import { i18n } from '../../../../boot/i18n';
import LocalMachineItem from '../../../../components/wizard/LocalMachineItem.vue';
import InputPasswordDialog from '../../../../components/setting/InputPasswordDialog.vue';
import { onFirstFactor } from '../../../../utils/account';
import { ConfirmButtonStatus } from '../../../../utils/constants';
import TerminusTipDialog from '../../../../components/dialog/TerminusTipDialog.vue';
import { useDeviceStore } from '../../../../stores/device';
import ActiveOlaresReminderDialog from '../../../../components/dialog/ActiveOlaresReminderDialog.vue';

import './wizard.scss';
import { ThemeDefinedMode } from '@bytetrade/ui';
import { UserItem } from '@didvault/sdk/src/core';
import { getNativeAppPlatform } from 'src/application/platform';

const mdnsStore = useMDNSStore();
mdnsStore.mdnsUsed = true;

const userStore = useUserStore();

const { t } = useI18n();

const isScaning = ref(true);

const isNotFound = ref(false);

onMounted(() => {
	mdnsStore.startSearchMdnsService();
	startFoundMachines();
});

onUnmounted(() => {
	mdnsStore.stopSearchMdnsService();
	clearfoundMachinesTimer();
});

const installAction = async (machine: TerminusServiceInfo) => {
	$q.loading.show();
	const result = await mdnsStore.installMachineTerminus(machine);
	$q.loading.hide();
	if (result) {
		mdnsStore.setActivedMachine(machine, 2000);
	}
};

const uninstallAction = async (machine: TerminusServiceInfo) => {
	$q.dialog({
		component: TerminusTipDialog,
		componentProps: {
			title: i18n.global.t('Uninstall'),
			message: i18n.global.t(
				'Are you sure you want to uninstall the device with IP address {ip}?',
				{
					ip: machine.host
				}
			)
		}
	}).onOk(async () => {
		$q.loading.show();
		const result = await mdnsStore.uninstallMachineTerminus(machine);
		$q.loading.hide();
		if (result) {
			mdnsStore.setActivedMachine(machine, 2000);
		}
	});
};

let foundMachinesTimer: any = undefined;

watch(
	() => mdnsStore.mdnsMachines,
	() => {
		const hostInfo = userStore.current_user?.localMachine
			? JSON.parse(userStore.current_user?.localMachine)
			: undefined;

		if (hostInfo) {
			const localMachine = mdnsStore.mdnsMachines.find(
				(e) => e.host == hostInfo.host
			);
			if (localMachine) {
				mdnsStore.setActivedMachine(localMachine, 2000);
			}
		}
		if (mdnsStore.mdnsMachines.length > 0) {
			isScaning.value = false;
			isNotFound.value = false;
		}
	},
	{
		deep: true
	}
);

const clearfoundMachinesTimer = () => {
	if (foundMachinesTimer) {
		clearInterval(foundMachinesTimer);
		foundMachinesTimer = undefined;
	}
};

let leftTime = 10;
const startFoundMachines = () => {
	if (foundMachinesTimer) {
		clearfoundMachinesTimer();
	}
	leftTime = 3;
	isNotFound.value = false;
	isScaning.value = true;

	foundMachinesTimer = setInterval(async () => {
		leftTime = leftTime - 1;
		if (leftTime == 0) {
			clearfoundMachinesTimer();
			if (mdnsStore.mdnsMachines.length == 0) {
				isNotFound.value = true;
			}
			isScaning.value = false;
		}
	}, 1 * 1000);
};

const rescanTerminus = () => {
	mdnsStore.stopSearchMdnsService();
	mdnsStore.startSearchMdnsService();
	startFoundMachines();
};

const $q = useQuasar();
const router = useRouter();
const deviceStore = useDeviceStore();

const activeLocal = async (machine: TerminusServiceInfo) => {
	if (!userStore.current_user?.localMachine) {
		$q.dialog({
			component: InputPasswordDialog,
			componentProps: {
				title: t('Enter your Olares password'),
				passwordTitle: t('password')
			}
		}).onOk(async (password: string) => {
			const user = userStore.users!.items.get(userStore.current_id!)!;
			let baseURL = `http://${machine.host}:30180`;

			try {
				await onFirstFactor(
					baseURL,
					user.name,
					user.local_name,
					password,
					false,
					undefined,
					machine.status?.terminusVersion
				);
				activeTerminus(machine, password);
			} catch (error) {
				notifyFailed(error);
			}
		});
		return;
	}
	const hostInfo = JSON.parse(userStore.current_user?.localMachine);
	if (hostInfo.host != machine.host || !hostInfo.password) {
		return;
	}
	activeTerminus(machine, hostInfo.password);
};

const activeTerminus = async (
	machine: TerminusServiceInfo,
	password: string
) => {
	const user = userStore.users!.items.get(userStore.current_id!)!;
	if (
		machine.status?.terminusVersion &&
		mdnsStore.isNewActivedMineVersion(machine.status.terminusVersion)
	) {
		$q.loading.show();
		const frpList = await mdnsStore.getFrpList(user.name);
		$q.loading.hide();
		if (!frpList) {
			notifyFailed('get frp list error');
			return;
		}

		$q.dialog({
			component: ActiveOlaresReminderDialog,
			componentProps: {
				frpList
			}
		}).onOk(async (host: string) => {
			active(machine, user, password, host);
		});
	} else {
		active(machine, user, password, '');
	}
};

const active = async (
	machine: TerminusServiceInfo,
	user: UserItem,
	password: string,
	host: string
) => {
	if (userStore.current_user?.access_token) {
		userStore.current_user.access_token = '';
	}
	getNativeAppPlatform().hookServerHttp = false;
	$q.loading.show();

	let system: SystemOption = {
		language: i18n.global.locale.value,
		location: 'Singapore'
	};

	const obj: WizardInfo = {
		step: 4,
		username: user.name,
		password: password,
		url: `http://${machine.host}:30180/server`,
		system: system,
		network: {
			enable_tunnel: true,
			external_ip: ''
		}
	};

	const isNewActivedMachine =
		machine.status?.terminusVersion != undefined &&
		mdnsStore.isNewActivedMineVersion(machine.status.terminusVersion);
	if (isNewActivedMachine) {
		obj.system.frp = {
			host: host,
			jws: ''
		};
		obj.system.theme =
			deviceStore.theme == ThemeDefinedMode.DARK ? 'dark' : 'light';
	} else {
		try {
			const data: any = await axios.get(
				`http://${machine.host}:30180/bfl/backend/v1/ip`
			);
			const external = data.masterExternalIP;
			obj.network.external_ip = external;
		} catch (e) {
			$q.loading.hide();
			notifyFailed(e.message);
			return;
		}
	}

	const mnemonic = userStore.users!.mnemonics.get(userStore.current_id!)!;

	await userBindTerminus(
		user,
		mnemonic,
		obj.url,
		obj.password!,
		machine.status?.terminusVersion || '',
		{
			async onSuccess() {
				$q.loading.hide();
				const new_user = userStore.users!.items.get(userStore.current_id!)!;
				new_user.wizard = JSON.stringify(obj);
				new_user.terminus_activate_status = 'wait_activate_system';
				new_user.setup_finished = false;
				await userStore.users!.items.update(new_user);
				await userStore.save();
				router.replace({ path: '/ActivateWizard' });
			},
			onFailure(message: string) {
				$q.loading.hide();
				getNativeAppPlatform().hookServerHttp = true;
				if (message && message.length > 0) {
					notifyFailed(message);
				}
			}
		}
	);
};

const enterBlueToothConfigNetwork = () => {
	router.push({ path: '/bluetooth/discover/machine' });
};
</script>

<style scoped lang="scss">
.local-machine-root {
	width: 100%;
	height: 100%;

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
