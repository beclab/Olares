<template>
	<div class="active-wizard-root">
		<div class="wizard-content">
			<check-system
				:wizardStatus="wizardStatus"
				v-if="
					wizardStatus == 'wait_activate_system' ||
					wizardStatus == 'system_activating' ||
					wizardStatus == 'system_activate_failed'
				"
			/>
			<CheckNetwork
				:wizard="wizard"
				:baseURL="baseURL"
				:wizardStatus="wizardStatus"
				:retrying="hadActivateNetwork"
				:taskState="!user.isLargeVersion12_2"
				v-else-if="
					wizardStatus == 'wait_activate_network' ||
					wizardStatus == 'network_activating' ||
					wizardStatus == 'network_activate_failed'
				"
				@onNetworkSetup="resetNetwork"
			/>

			<CheckDNS
				:wizard="wizard"
				:failed="failed"
				:wizardStatus="wizardStatus"
				:reminderInfo="t(reminderInfoRef)"
				@dns-retry="dnsConfigRetry"
				v-else-if="wizardStatus == 'wait_reset_password'"
			/>
			<div v-else>{{ t('loading') }}</div>
		</div>
	</div>
</template>

<script lang="ts" setup>
import { ref, onMounted, onUnmounted, onBeforeUnmount } from 'vue';
import { useUserStore } from '../../../../stores/user';
import { OlaresInfo } from '@bytetrade/core';
import { useQuasar } from 'quasar';
import { useRouter } from 'vue-router';
import axios from 'axios';
import CheckSystem from './CheckSystem.vue';
import CheckNetwork from './CheckNetwork.vue';
import CheckDNS from './CheckDNS.vue';
import { UserItem } from '@didvault/sdk/src/core';
import { setSenderUrl } from '../../../../globals';
import { getAppPlatform } from '../../../../application/platform';
import { useI18n } from 'vue-i18n';
import { commonInterceptValue } from '../../../../utils/response';

import './wizard.scss';
import { getOlaresInfo } from '../../../../utils/account';
import { getBflTerminusInfo } from '../../../../utils/BindTerminusBusiness';
import AndroidPlugins from 'src/platform/interface/capacitor/android/androidPlugins';
import { WizardInfo } from 'src/utils/interface/wizard';

const failed = ref(false);

const userStore = useUserStore();
const router = useRouter();
const $q = useQuasar();
const { t } = useI18n();

const user: UserItem = userStore.users!.items.get(userStore.current_id!)!;
const wizard: WizardInfo = JSON.parse(user.wizard);

const wizardStatus = ref<string>(user.terminus_activate_status);

const hadActivateNetwork = ref(false);

let updateTerminusInfoIng = false;

let user_info_interval: any = null;

let baseURL = wizard.url;
if (process.env.IS_PC_TEST) {
	baseURL = window.location.origin;
} else {
	if (baseURL.endsWith('/server')) {
		baseURL = baseURL.substring(0, baseURL.length - 7);
	}
}

let times = 0;

async function updateUserStatus(status: string) {
	user.terminus_activate_status = status;
	await userStore.users!.items.update(user);
	await userStore.save();
}

async function gotoResetPassword() {
	wizardStatus.value = 'wait_reset_password';
	await updateUserStatus('wait_reset_password');

	setSenderUrl({
		url: user.vault_url
	});
	router.push({ path: '/ResetPassword' });
}

let last_set_wait_reset_password_time = 0;

let getHostTerminusCount = 0;

async function updateTerminusInfo(): Promise<string | undefined> {
	if (wizardStatus.value == 'wait_reset_password') {
		return undefined;
	}

	if (updateTerminusInfoIng) {
		return;
	}
	updateTerminusInfoIng = true;

	try {
		const data: OlaresInfo = await getOlaresInfo(baseURL, 5000, {
			t: new Date().getTime()
		});

		if (commonInterceptValue.includes(data as any)) {
			return undefined;
		}
		await userStore.setUserTerminusInfo(user.id, data);

		wizardStatus.value = data.wizardStatus;
		if (user.terminus_activate_status != wizardStatus.value) {
			await updateUserStatus(wizardStatus.value);
		}
		return wizardStatus.value;
	} catch (e) {
		// failed.value = true;
		// return undefined;
		try {
			return await authRequestTerminusInfo();
		} catch (error) {
			if (getHostTerminusCount < 10) {
				getHostTerminusCount += 1;
			} else {
				failed.value = true;
			}

			return undefined;
		}
	} finally {
		updateTerminusInfoIng = false;
		if (
			last_set_wait_reset_password_time == 0 &&
			wizardStatus.value == 'wait_reset_password'
		) {
			last_set_wait_reset_password_time = new Date().getTime();
		}
	}
}

async function authRequestTerminusInfo() {
	const data = await getBflTerminusInfo(user, undefined, {
		t: new Date().getTime()
	});
	if (!data || commonInterceptValue.includes(data as any)) {
		return;
	}

	wizardStatus.value = data.wizardStatus;

	if (user.terminus_activate_status != wizardStatus.value) {
		await updateUserStatus(wizardStatus.value);
	}
	return wizardStatus.value;
}

async function startSctivate() {
	if (user.isLargeVersion12_2) {
		await activate();
	} else {
		await configSystem();
	}
}

async function configSystem() {
	wizardStatus.value = 'system_activating';
	try {
		await axios.post(
			baseURL + '/bfl/settings/v1alpha1/config-system',
			wizard.system
		);
	} catch (e) {
		wizardStatus.value = 'system_activate_failed';
	}
	await updateTerminusInfo();
}

async function activate() {
	wizardStatus.value = 'system_activating';
	try {
		const frpInfo = wizard.system.frp;
		if (frpInfo) {
			const jws = await userStore.activeOlaresJws({
				...frpInfo,
				jws: undefined
			});
			frpInfo.jws = jws;
		}
		wizard.system.frp = frpInfo;
		await axios.post(
			baseURL + '/bfl/settings/v1alpha1/activate',
			wizard.system
		);
	} catch (e) {
		wizardStatus.value = 'network_activate_failed';
	}
	await updateTerminusInfo();
}

async function resetNetwork() {
	hadActivateNetwork.value = false;
	if (user.isLargeVersion12_2) {
		await activate();
	} else {
		await configNetwork();
	}
}

async function configNetworkNew() {
	if (user.isLargeVersion12_2) {
		return;
	}
	configNetwork();
}

async function configNetwork() {
	wizardStatus.value = 'network_activating';
	if (hadActivateNetwork.value) {
		return;
	}
	hadActivateNetwork.value = true;
	try {
		let ob: any = {};
		if (wizard.network.enable_tunnel) {
			ob.enable_reverse_proxy = true;
			ob.enable_tunnel = true;
			ob.ip = '';
		} else {
			ob.enable_reverse_proxy = false;
			ob.enable_tunnel = false;
			ob.ip = wizard.network.external_ip;
		}

		await axios.post(baseURL + '/bfl/settings/v1alpha1/ssl/enable', ob);
	} catch (e) {
		hadActivateNetwork.value = false;
		wizardStatus.value = 'network_activate_failed';
	}
	await updateTerminusInfo();
}

async function updateInfo() {
	if (updateTerminusInfoIng) {
		return;
	}
	await updateTerminusInfo();

	if (
		wizardStatus.value == 'vault_activate_failed' ||
		wizardStatus.value == 'system_activate_failed' ||
		wizardStatus.value == 'network_activate_failed'
	) {
		return;
	}

	if (
		wizardStatus.value == 'vault_activating' ||
		wizardStatus.value == 'system_activating' ||
		wizardStatus.value == 'network_activating'
	) {
		return;
	}

	if (wizardStatus.value == 'completed') {
		router.push({ path: '/home' });
	} else if (wizardStatus.value == 'wait_activate_system') {
		await startSctivate();
	} else if (wizardStatus.value == 'wait_activate_network') {
		await configNetworkNew();
	} else if (wizardStatus.value == 'wait_reset_password') {
		if (process.env.IS_PC_TEST) {
			times++;
			if (times > 5) {
				await gotoResetPassword();
				return;
			}
		}
		try {
			const now = new Date().getTime();
			const checkDNSTimer = now - last_set_wait_reset_password_time;

			if (last_set_wait_reset_password_time > 0) {
				if (checkDNSTimer >= 180 * 1000 && checkDNSTimer < 300 * 1000) {
					reminderInfoRef.value =
						'The DNS settings may take longer to take effect. Please wait patiently.';
				} else if (checkDNSTimer >= 300 * 1000) {
					reminderInfoRef.value =
						'There was an issue applying the DNS settings. Please switch your phone’s network and try again. If the problem persists, consider seeking help in the Olares forum.';
				}
			}

			if (
				last_set_wait_reset_password_time > 0 &&
				now - last_set_wait_reset_password_time > 75 * 1000
			) {
				await authRequestTerminusInfo();
				await gotoResetPassword();
			}
		} catch (e) {
			console.error(e);
		}
	}
}

onMounted(async () => {
	if ($q.platform.is.android) {
		// StatusBar.setOverlaysWebView({ overlay: true });
		await AndroidPlugins.EdgeToEdge.disable();
	}
	user_info_interval = setInterval(async () => {
		await updateInfo();
	}, 2 * 1000);

	await updateInfo();
	getAppPlatform().hookServerHttp = false;
});

onUnmounted(() => {
	if (user_info_interval) {
		clearInterval(user_info_interval);
	}
	getAppPlatform().hookServerHttp = true;
});

if (user.terminus_activate_status == 'wait_reset_password') {
	setSenderUrl({
		url: user.vault_url
	});
	router.push({ path: '/ResetPassword' });
} else if (user.terminus_activate_status == 'completed') {
	router.push({ path: '/home' });
}

const dnsConfigRetry = () => {
	failed.value = false;
	getHostTerminusCount = 0;
	updateInfo();
};

onBeforeUnmount(async () => {
	if ($q.platform.is.android) {
		// StatusBar.setOverlaysWebView({ overlay: false });
		await AndroidPlugins.EdgeToEdge.enable();
	}
});

const reminderInfoRef = ref('dns_resolution_introduce');
</script>

<style lang="scss" scoped>
.active-wizard-root {
	width: 100%;
	height: 100%;

	&__content {
		width: 100%;
		height: 100%;
	}
}
</style>
