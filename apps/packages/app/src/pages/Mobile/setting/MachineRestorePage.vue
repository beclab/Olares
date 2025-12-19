<template>
	<div class="restore-root">
		<terminus-title-bar :title="t('Restore factory status')" />
		<terminus-scroll-area class="restore-scroll-area">
			<template v-slot:content>
				<div v-if="!hasCheckPassword">
					<div class="text-body2 text-ink-2 q-pt-md">
						{{
							t(
								'This operation will completely delete thedata on the device, including:'
							)
						}}
					</div>
					<div
						class="row text-body2"
						v-for="(item, index) in reminderList"
						:key="item"
						:class="index == 0 ? 'q-mt-md' : 'q-mt-xs'"
					>
						<div
							class="q-mb-sm row justify-center"
							style="width: 20px; padding-top: 4px"
						>
							<div
								style="width: 8px; height: 8px; border-radius: 4px"
								class="bg-background-5"
							></div>
						</div>
						<div
							class="q-mb-sm text-ink-2"
							style="width: calc(100% - 20px)"
							v-html="item"
						/>
					</div>
					<div class="text-body2 text-ink-2 q-pt-md">
						{{ t('Before proceeding, please back up your important data.') }}
					</div>
				</div>
				<div v-else class="column items-center" style="margin-top: 120px">
					<q-circular-progress
						rounded
						:value="ageRef"
						size="160px"
						:thickness="0.15"
						color="light-blue-default"
						track-color="light-blue-alpha"
					/>
					<div class="text-ink-2 text-body2 q-mt-lg" style="text-align: center">
						{{
							t(
								'Operation will be completed in 5 minutes. Do not close Olares during this time.'
							)
						}}
					</div>
				</div>
			</template>
		</terminus-scroll-area>
		<div class="bottom-view column justify-end" v-if="!hasCheckPassword">
			<confirm-button
				:btn-title="t('Restore factory status')"
				:btn-status="
					mdnsStore.activedMachine == undefined
						? ConfirmButtonStatus.disable
						: ConfirmButtonStatus.normal
				"
				@click="openCheckLogin"
			/>
		</div>
	</div>
</template>
<script setup lang="ts">
import TerminusTitleBar from '../../../components/common/TerminusTitleBar.vue';
import TerminusScrollArea from '../../../components/common/TerminusScrollArea.vue';
import ConfirmButton from '../../../components/common/ConfirmButton.vue';
import { notifyFailed } from '../../../utils/notifyRedefinedUtil';
import InputPasswordDialog from '../../../components/setting/InputPasswordDialog.vue';
import { ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { useQuasar } from 'quasar';
import { useUserStore } from '../../../stores/user';
import { useMDNSStore } from '../../../stores/mdns';
import { TerminusStatusEnum } from '../../../services/abstractions/mdns/service';

import { onMounted, onUnmounted, watch } from 'vue';
import { useRouter } from 'vue-router';
import { useTermipassStore } from '../../../stores/termipass';
import { busEmit } from '../../../utils/bus';
import { setSenderUrl } from '../../../globals';
import { ConfirmButtonStatus } from 'src/utils/constants';

const { t } = useI18n();

const userStore = useUserStore();

const termipassStore = useTermipassStore();

const $q = useQuasar();

const mdnsStore = useMDNSStore();
mdnsStore.mdnsUsed = true;

const hasCheckPassword = ref(false);

const router = useRouter();

const ageRef = ref(0);

const reminderList = ref([
	t('All your accounts (including member accounts)'),
	t('System and application data and settings'),
	t('Installed third-party applications.'),
	t('All photos, videos, music, documents, and other files')
]);

const openCheckLogin = async () => {
	if (!(await userStore.unlockFirst(undefined, { hide: true }))) {
		return;
	}
	if (!userStore.passwordReseted) {
		busEmit('configPassword');
		return;
	}

	$q.dialog({
		component: InputPasswordDialog,
		componentProps: {
			title: t('Enter your unlock password'),
			passwordTitle: t('Enter password')
		}
	}).onOk((password: string) => {
		if (password != userStore.password) {
			notifyFailed(t('password_error'));
			return;
		}
		uninstallAction();
	});
};

const uninstallAction = async () => {
	if (!mdnsStore.activedMachine) {
		return;
	}
	hasCheckPassword.value = true;
	const result = await mdnsStore.uninstallMachineTerminus(
		mdnsStore.activedMachine
	);

	if (!result) {
		hasCheckPassword.value = false;
	}

	if (result && mdnsStore.activedMachine.isSettingsServer) {
		setTimeout(() => {
			uninstallSuccess();
		}, 2000);
	}
};

watch(
	() => mdnsStore.mdnsMachines,
	() => {
		if (mdnsStore.mdnsMachines.length > 0) {
			const activedMachine = mdnsStore.mdnsMachines.find(
				(e) => e.status && e.status.terminusName == userStore.current_user?.name
			);
			if (activedMachine) mdnsStore.setActivedMachine(activedMachine, 2000);
		}
	},
	{
		deep: true
	}
);

watch(
	() => mdnsStore.activedMachine,
	async () => {
		if (mdnsStore.activedMachine && mdnsStore.activedMachine.status) {
			ageRef.value = mdnsStore.activedMachine.status.uninstallingProgress
				? Number(
						mdnsStore.activedMachine.status.uninstallingProgress.split('%')[0]
				  )
				: 0;

			if (
				mdnsStore.activedMachine.status.terminusState ==
				TerminusStatusEnum.NotInstalled
			) {
				uninstallSuccess();
			}
		}
	},
	{
		deep: true
	}
);

onMounted(() => {
	mdnsStore.startSearchMdnsService();
});

onUnmounted(() => {
	mdnsStore.stopSearchMdnsService();
});

const uninstallSuccess = async () => {
	termipassStore.reactivation = true;
	const user = userStore.users!.items.get(userStore.current_id!)!;
	user.wizard = 'restore';
	user.terminus_activate_status = 'wait_activate_vault';
	user.setup_finished = false;
	user.isLocal = false;
	userStore.users!.items.update(user);
	await userStore.save();
	setSenderUrl({
		url: user.vault_url
	});
	router.push({
		path: `/Activate/${1}`
	});
};
</script>

<style scoped lang="scss">
.restore-root {
	width: 100%;
	height: 100%;

	.restore-scroll-area {
		width: 100%;
		height: calc(100% - 56px - 48px - 48px);
		padding-left: 20px;
		padding-right: 20px;
	}

	.bottom-view {
		width: 100%;
		padding-bottom: 48px;
		padding-left: 20px;
		padding-right: 20px;
	}
}
</style>
