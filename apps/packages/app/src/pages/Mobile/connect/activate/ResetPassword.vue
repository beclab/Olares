<template>
	<div class="terminus-login-root">
		<terminus-wizard-view
			:btn-status="btnStatusRef"
			:btn-title="btnTextRef"
			:enable-overlay="false"
			:bottom-more-height="60"
			@on-confirm="onConfirm"
			@onError="clearData"
		>
			<template v-slot:content>
				<div class="terminus-login-root__content">
					<div class="row items-center justify-center">
						<div class="terminus-login-root__content__image">
							<TerminusAvatar :info="userStore.terminusInfo()" :size="100" />
						</div>
					</div>
					<div class="terminus-login-root__content__name text-h5 q-mt-md">
						{{ userStore.current_user?.local_name }}
					</div>
					<div class="terminus-login-root__content__info text-body2 q-mt-xs">
						{{ t('reset_your_olares_device_password') }}
					</div>

					<terminus-password-validator
						class="q-mt-md"
						:edt-transaction="true"
						ref="passwordValidator"
						v-model:button-status="btnStatusRef"
						v-model:button-text="btnTextRef"
					/>
				</div>
			</template>
			<template v-slot:bottom-top>
				<TerminusExportMnemonicRoot :height="48" class="q-mb-sm" />
			</template>
		</terminus-wizard-view>

		<div
			class="terminus-login-root__img row items-center justify-center"
			style="top: 20px"
		>
			<TerminusChangeUserHeader :scan="false">
				<template v-slot:avatar>
					<q-icon name="account_circle" size="24px" color="grey-8" />
				</template>
			</TerminusChangeUserHeader>
		</div>
	</div>
</template>

<script lang="ts" setup>
import { ref, watch } from 'vue';
import { useRouter } from 'vue-router';
import { useUserStore } from '../../../../stores/user';

import { useQuasar } from 'quasar';
import { SSHPassworResetType, UserItem } from '@didvault/sdk/src/core';
import TerminusWizardView from '../../../../components/common/TerminusWizardView.vue';
import { ConfirmButtonStatus } from '../../../../utils/constants';
import { useI18n } from 'vue-i18n';
import { loginTerminus } from '../../../../utils/BindTerminusBusiness';
// import axios from 'axios';
import TerminusPasswordValidator from '../../../../components/common/TerminusPasswordValidator.vue';
import { busEmit } from '../../../../utils/bus';
import { notifyFailed } from '../../../../utils/notifyRedefinedUtil';
import TerminusChangeUserHeader from '../../../../components/common/TerminusChangeUserHeader.vue';
import TerminusExportMnemonicRoot from '../../../../components/common/TerminusExportMnemonicRoot.vue';
import { reset_password } from '../../../../utils/account';
import { useMDNSStore } from 'src/stores/mdns';
import {
	getSettingsServerMdnsRequestApi,
	isOlaresGlobalDevice,
	MdnsApiEmum,
	TerminusStatus
} from 'src/services/abstractions/mdns/service';
import { WizardInfo } from 'src/utils/interface/wizard';

const { t } = useI18n();
const $q = useQuasar();
const router = useRouter();
const terminusNameRef = ref<string>('');
const userStore = useUserStore();

const btnTextRef = ref(t('complete'));

const btnStatusRef = ref<ConfirmButtonStatus>(ConfirmButtonStatus.disable);
const user: UserItem = userStore.users!.items.get(userStore.current_id!)!;
const wizard: WizardInfo = JSON.parse(user.wizard);
terminusNameRef.value = user.name;

let baseURL = wizard.url;
if (process.env.IS_PC_TEST) {
	baseURL = window.location.origin;
} else {
	baseURL = user.auth_url;
}

if (baseURL.endsWith('/')) {
	baseURL = baseURL.slice(0, -1);
}

const onConfirm = async () => {
	if (!(await userStore.unlockFirst())) {
		return;
	}
	if (btnStatusRef.value != ConfirmButtonStatus.normal) {
		return;
	}
	const newPassword = passwordValidator.value.getValidPassword();
	if (!newPassword) {
		notifyFailed(t('password_not_meet_rules'));
		return;
	}

	$q.loading.show();

	try {
		await loginTerminus(user, wizard.password!, false);
		await reset_password(
			baseURL,
			user.local_name,
			wizard.password,
			newPassword,
			user.access_token,
			userStore.getUserTerminusInfo(user.id).osVersion
		);
		await loginTerminus(user, newPassword, true);
	} catch (e) {
		$q.loading.hide();
		notifyFailed(e.message);
		return;
	}

	try {
		user.setup_finished = true;
		user.wizard = '';
		const olaresInfo = userStore.getUserTerminusInfo(user.id);
		const needResetSSH = await needResetSSHPassword();
		if (needResetSSH) {
			user.owner_ssh_is_default = SSHPassworResetType.NEED_RESET;
		} else {
			user.owner_ssh_is_default = SSHPassworResetType.SKIP;
		}
		if (olaresInfo && olaresInfo.id) {
			user.olares_id = olaresInfo.id;
		} else if (olaresInfo && olaresInfo.terminusId) {
			user.olares_id = olaresInfo.terminusId;
		}
		user.os_version = olaresInfo.osVersion;
		await userStore.users!.items.update(user);
		await userStore.save();

		busEmit('terminus_actived');

		router.push({ path: '/home' });
	} catch (e) {
		notifyFailed(e.message);
	} finally {
		$q.loading.hide();
	}
};

const passwordValidator = ref();
const clearData = () => {
	passwordValidator.value.clearPassword();
};

watch(
	() => btnStatusRef.value,
	() => {
		if (btnStatusRef.value === ConfirmButtonStatus.error) {
			btnTextRef.value = t('reset');
		} else {
			btnTextRef.value = t('complete');
		}
	}
);

const needResetSSHPassword = async () => {
	if (!user.isLargeVersion12_2) {
		return false;
	}
	try {
		const mdnsStore = useMDNSStore();
		const instance = await mdnsStore.getOlaresStatusInfoInstance();
		const res = await instance.get('/api/system/status');

		const status: TerminusStatus = res.data.data;
		if (status.terminusName != user.name || !isOlaresGlobalDevice(status)) {
			return false;
		}
	} catch (error) {
		return false;
	}
	return true;
};
</script>

<style lang="scss" scoped>
.terminus-login-root {
	width: 100%;
	height: 100%;
	// background: $background;
	position: relative;

	&__content {
		width: 100%;
		height: 100%;
		padding-top: 80px;

		&__image {
			height: 100px;
			width: 100px;
			border-radius: 50px;
			overflow: hidden;
		}

		&__name {
			text-align: center;
			color: $ink-1;
		}

		&__info {
			text-align: center;
			color: $ink-2;
		}
	}
	&__img {
		width: 100%;
		height: 40px;
		position: absolute;
		right: 0px;
		border-radius: 16px;
		overflow: hidden;

		.header-content {
			width: 100%;
			.scan-icon {
				width: 40px;
				height: 40px;
			}
		}
		.avatar {
			height: 40px;
			width: 40px;
			overflow: hidden;
			border-radius: 20px;
		}
	}
}
</style>
