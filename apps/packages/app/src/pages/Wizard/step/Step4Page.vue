<template>
	<div>
		<CheckVault
			v-if="
				tokenStore.user.wizardStatus == 'wait_activate_vault' ||
				tokenStore.user.wizardStatus == 'vault_activating' ||
				tokenStore.user.wizardStatus == 'vault_activate_failed'
			"
		/>
		<CheckSystem
			v-if="
				tokenStore.user.wizardStatus == 'wait_activate_system' ||
				tokenStore.user.wizardStatus == 'system_activating' ||
				tokenStore.user.wizardStatus == 'system_activate_failed'
			"
		/>
		<CheckNetwork
			v-if="
				tokenStore.user.wizardStatus == 'wait_activate_network' ||
				tokenStore.user.wizardStatus == 'network_activating' ||
				tokenStore.user.wizardStatus == 'network_activate_failed'
			"
		/>

		<div
			v-if="
				tokenStore.user.wizardStatus == 'wait_reset_password' ||
				tokenStore.user.wizardStatus == 'completed'
			"
		>
			<CheckPassword
				v-if="tokenStore.pingResult === true"
				:wizard_succeed="wizard_succeed"
			/>
			<CheckDNS
				v-else
				:step8Img="step8Img"
				:step8CertificateImg="step8CertificateImg"
			/>
		</div>

		<img :src="step8Img" style="display: none" />
		<img :src="step8CertificateImg" style="display: none" />
		<img :src="wizard_succeed" style="display: none" />
	</div>
</template>

<script lang="ts" setup>
import { onMounted, onUnmounted } from 'vue';
import { useTokenStore } from '../../../stores/wizard-step';

import CheckVault from './status/CheckVault.vue';
import CheckSystem from './status/CheckSystem.vue';
import CheckNetwork from './status/CheckNetwork.vue';
import CheckDNS from './status/CheckDNS.vue';
import CheckPassword from './status/CheckPassword.vue';

const step8Img = require('../../../assets/wizard-web/wizard_resolutiona.png');
const step8CertificateImg = require('../../../assets/wizard-web/wizard_resolution.png');
const wizard_succeed = require('../../../assets/wizard-web/wizard_succeed.png');

const tokenStore = useTokenStore();

let user_info_interval: any = null;

let last_set_wait_reset_password_time = 0;

onMounted(async () => {
	user_info_interval = setInterval(async () => {
		if (
			tokenStore.user.wizardStatus == 'wait_reset_password' ||
			tokenStore.user.wizardStatus == 'completed'
		) {
			const pingResult = await tokenStore.ping2();
			const now = new Date().getTime();
			if (
				tokenStore.user.wizardStatus == 'wait_reset_password' &&
				last_set_wait_reset_password_time > 0 &&
				now - last_set_wait_reset_password_time > 75 * 1000 &&
				pingResult
			) {
				tokenStore.pingResult = true;
			}

			if (tokenStore.user.wizardStatus == 'completed') {
				clearInterval(user_info_interval);
				window.location.replace(tokenStore.get_terminus_url!);
			}
		} else {
			await tokenStore.loadData(true);
		}
		if (
			last_set_wait_reset_password_time == 0 &&
			tokenStore.user.wizardStatus == 'wait_reset_password'
		) {
			last_set_wait_reset_password_time = new Date().getTime();
		}
	}, 2 * 1000);
});

onUnmounted(() => {
	if (user_info_interval) {
		clearInterval(user_info_interval);
	}
});
</script>
