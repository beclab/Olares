<template>
	<page-title-component :show-back="true" :title="t('My hardware')" />
	<bt-scroll-area class="nav-height-scroll-area-conf">
		<div class="terminus-cloud-page column items-center">
			<QRCodeOlaresdCommand
				:command="command"
				:title="notificationTitle"
				:body="notificationBody"
				:data="terminusDStore.commandData"
				@success="commandSuccess"
			>
				<template v-slot:mode>
					<div class="row items-center justify-center">
						<q-img class="terminus-cloud-icon" src="settings/scan.svg" />
						<div class="q-ml-md text-h4 text-ink-1">
							{{ title }}
						</div>
					</div>
				</template>
			</QRCodeOlaresdCommand>
		</div>
	</bt-scroll-area>
</template>

<script setup lang="ts">
import PageTitleComponent from '../../../components/settings/PageTitleComponent.vue';
import QRCodeOlaresdCommand from '../../../components/settings/QRCodeOlaresdCommand.vue';
import { useRoute, useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { notifySuccess } from 'src/utils/settings/btNotify';
import { ref } from 'vue';
import { useTerminusDStore } from 'src/stores/settings/terminusd';

const { t } = useI18n();

const Route = useRoute();

const router = useRouter();

const command = ref(Route.params.command as string);

const terminusDStore = useTerminusDStore();

const notificationTitle = ref('');

const notificationBody = ref('');

const title = ref('');

const commandSuccess = () => {
	router.back();
	if (command.value == 'shutdown') {
		notifySuccess(
			t(
				'The system will be completely powered off and can only be restarted by physically powering it on.'
			)
		);
	}
	if (command.value == 'reboot') {
		notifySuccess(
			t(
				'After restarting, the system will be unavailable for a short period. Please wait patiently.'
			)
		);
	}

	if (command.value == 'ssh-password') {
		notifySuccess(
			t('SSH password reset to {password}', {
				password: terminusDStore.commandData?.password
			})
		);
	}
};

const configInfo = () => {
	if (command.value == 'shutdown') {
		title.value = t('login.scan_to_shutdown');
		notificationTitle.value = t('Confirm Shutdown');
		notificationBody.value = t(
			'Are you sure you want to shut down? After shutdown, the system will be completely powered off and can only be restarted by physically powering it on.'
		);
	} else if (command.value == 'reboot') {
		title.value = t('login.scan_to_reboot');
		notificationTitle.value = t('Confirm Restart');
		notificationBody.value = t(
			'Are you sure you want to restart? After restarting, the system will be unavailable for a short period. Please wait patiently.'
		);
	} else if (command.value == 'ssh-password') {
		title.value = t('login.scan_to_reset_ssh_password');
		notificationTitle.value = t('Confirm Reset');
		notificationBody.value = t(
			'Are you sure you want to reset your SSH login password?'
		);
	}
};
configInfo();
</script>

<style scoped lang="scss">
.terminus-cloud-page {
	width: 100%;
	height: calc(100% - 56px);
	.terminus-cloud-icon {
		width: 32px;
	}
}
</style>
