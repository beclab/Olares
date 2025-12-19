<template>
	<div class="terminus-bind-url-root">
		<back-title-bar />
		<q-scroll-area
			:class="keyboardOpen ? 'scroll-area-conf-open' : 'scroll-area-conf-close'"
		>
			<div class="terminus-bind-url">
				<page-title
					label="Terminus bind url"
					desc="Please input your terminus bind url."
					style="margin-top: 20px"
				/>
				<terminus-edit
					v-model="url"
					style="margin-top: 36px"
					label="Bind url"
					@update:model-value="terminusUrlUpdate"
				/>
			</div>
		</q-scroll-area>
		<confirm-button
			class="button-confirm"
			btn-title="Signup"
			@onConfirm="onSubmit"
			:btn-status="btnStatusRef"
		/>
	</div>
</template>

<script lang="ts" setup>
import { onMounted, ref, watch } from 'vue';
import { useRouter } from 'vue-router';
import { base64ToString, UserItem, MnemonicItem } from '@didvault/sdk/src/core';
import MonitorKeyboard from '../../../../utils/monitorKeyboard';
import { useQuasar } from 'quasar';
import TerminusEdit from '../../../../components/common/TerminusEdit.vue';
import BackTitleBar from '../../../../components/common/BackTitleBar.vue';
import PageTitle from '../../../../components/common/TerminusPageTitle.vue';
import ConfirmButton from '../../../../components/common/ConfirmButton.vue';
import { ConfirmButtonStatus } from '../../../../utils/constants';
import { useUserStore } from '../../../../stores/user';
import { userBindTerminus } from '../../../../utils/BindTerminusBusiness';
import { notifyFailed } from '../../../../utils/notifyRedefinedUtil';
import { onUnmounted } from 'vue';
import { getOlaresInfo } from '../../../../utils/account';
import { WizardInfo } from 'src/utils/interface/wizard';

const router = useRouter();

const $q = useQuasar();
const userStore = useUserStore();

let monitorKeyboard: MonitorKeyboard | undefined = undefined;
const keyboardOpen = ref(false);
const btnStatusRef = ref<ConfirmButtonStatus>(ConfirmButtonStatus.disable);

onMounted(async () => {
	if ($q.platform.is.android) {
		monitorKeyboard = new MonitorKeyboard();
		monitorKeyboard.onStart();
		monitorKeyboard.onShow(() => (keyboardOpen.value = true));
		monitorKeyboard.onHidden(() => (keyboardOpen.value = false));
	}
});

onUnmounted(() => {
	if ($q.platform.is.android) {
		if (monitorKeyboard) {
			monitorKeyboard.onEnd();
		}
	}
});

const url = ref<string>('');

async function onSubmit() {
	if (url.value.startsWith('http')) {
		notifyFailed('the url is error,please retry scan');
		btnStatusRef.value = ConfirmButtonStatus.error;
		return;
	}

	if (
		url.value.startsWith('https://file.bttcdn.com') ||
		url.value.startsWith('https://app.cdn.olares.com')
	) {
		notifyFailed('the url is error,please retry scan');
		return;
	}

	if (!userStore.current_id) {
		notifyFailed('Please choose user');
		return;
	}

	if (
		!url.value.startsWith('active-vault' + '://') &&
		!url.value.startsWith('active_vault' + '://')
	) {
		notifyFailed('the url is error,please retry scan');
		return;
	}

	const content = url.value.split('://')[1];

	$q.loading.show();

	const obj: WizardInfo = JSON.parse(base64ToString(content));
	if (obj.username?.split('@').length != 2) {
		notifyFailed('Username is error,please retry scan ' + obj.username);
		return;
	}

	if (obj.username !== userStore.current_user?.name) {
		notifyFailed('Username is error,please retry scan ' + obj.username);
		return;
	}

	const user: UserItem = userStore.users!.items.get(userStore.current_id!)!;
	const mnemonic: MnemonicItem = userStore.users!.mnemonics.get(
		userStore.current_id!
	)!;

	const terminusInfo = await getOlaresInfo(obj.url);

	await userBindTerminus(
		user,
		mnemonic,
		obj.url,
		obj.password!,
		terminusInfo.osVersion,
		{
			async onSuccess() {
				const new_user: UserItem = userStore.users!.items.get(
					userStore.current_id!
				)!;
				new_user.wizard = JSON.stringify(obj);
				new_user.terminus_activate_status = 'wait_activate_system';
				new_user.setup_finished = false;
				await userStore.users!.items.update(new_user);
				await userStore.save();

				router.push({ path: '/ActivateWizard' });
			},
			onFailure(message: string) {
				$q.loading.hide();
				if (message && message.length > 0) {
					notifyFailed(message);
				}
			}
		}
	);

	$q.loading.hide();
}

const terminusUrlUpdate = (terminusUrl: string) => {
	btnStatusRef.value = terminusUrl
		? ConfirmButtonStatus.normal
		: ConfirmButtonStatus.error;
};

watch(url, () => {
	btnStatusRef.value =
		url.value && url.value.length > 0
			? ConfirmButtonStatus.normal
			: ConfirmButtonStatus.disable;
});
</script>

<style scoped lang="scss">
.terminus-bind-url-root {
	width: 100%;
	height: 100%;

	.scroll-area-conf-open {
		width: 100%;
		height: calc(100% - 48px - 20px - 44px);
		padding-bottom: 10px;
	}

	.scroll-area-conf-close {
		width: 100%;
		height: calc(100% - 48px - 160px - 44px);
	}

	.terminus-bind-url {
		height: calc(100% - 44px);
		width: 100%;
		padding-top: 20px;
		padding-left: 24px;
		padding-right: 24px;
	}

	.button-confirm {
		width: 100%;
		padding-left: 24px;
		padding-right: 24px;
	}
}
</style>
