<template>
	<div class="terminus-login-root">
		<terminus-wizard-view
			:btn-status="btnStatusRef"
			:btn-title="btnTitle"
			:enable-overlay="false"
			:bottom-more-height="60"
			@on-confirm="onConfirm"
		>
			<template v-slot:content>
				<div class="terminus-login-page" style="position: relative">
					<div class="row items-center justify-center">
						<div class="terminus-login-page__image">
							<TerminusAvatar :info="userStore.terminusInfo()" :size="64" />
						</div>
					</div>
					<div
						class="terminus-text-ellipsis terminus-login-page__name text-h5 q-mt-md"
					>
						{{ userStore.current_user?.local_name }}
					</div>

					<div
						class="terminus-text-ellipsis terminus-login-page__desc text-body3 q-mt-xs"
					>
						{{ userStore.current_user?.name }}
					</div>

					<div class="row items-center justify-center" v-if="isLocalTest">
						<q-checkbox v-model="use_local"> local server </q-checkbox>
					</div>

					<terminus-edit
						v-model="osPwd"
						:label="t('password')"
						:show-password-img="true"
						class="terminus-login-page__edit"
						@update:model-value="pwdInputChange"
						@keyup.enter="onConfirm"
					/>
				</div>
			</template>
			<template v-slot:bottom-top>
				<TerminusExportMnemonicRoot :height="48" class="q-mb-sm" />
			</template>
		</terminus-wizard-view>
		<q-inner-loading :showing="loading" dark color="white" size="64px">
		</q-inner-loading>

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
import { ref, onMounted, computed } from 'vue';
import { useRouter } from 'vue-router';
import { useUserStore } from '../../../stores/user';
import { OlaresInfo, TerminusInfo } from '@bytetrade/core';
import { useQuasar } from 'quasar';
import { UserItem, MnemonicItem } from '@didvault/sdk/src/core';
import TerminusEdit from '../../../components/common/TerminusEdit.vue';
import { ConfirmButtonStatus } from '../../../utils/constants';
import MonitorKeyboard from '../../../utils/monitorKeyboard';
import TerminusChangeUserHeader from '../../../components/common/TerminusChangeUserHeader.vue';
import { useI18n } from 'vue-i18n';
import {
	connectTerminus,
	getTerminusInfo,
	loginTerminus
} from '../../../utils/BindTerminusBusiness';
import { busEmit } from '../../../utils/bus';
import { notifyFailed } from '../../../utils/notifyRedefinedUtil';
import { useTermipassStore } from '../../../stores/termipass';
import { getAppPlatform } from '../../../application/platform';
import { onUnmounted } from 'vue';
import TerminusWizardView from '../../../components/common/TerminusWizardView.vue';

const { t } = useI18n();
const $q = useQuasar();
const router = useRouter();
const terminusNameRef = ref<string>('');

const userStore = useUserStore();
const loading = ref(false);
const btnTitle = ref(t('complete'));

const termipassStore = useTermipassStore();

const osPwd = ref('');
const btnStatusRef = ref<ConfirmButtonStatus>(ConfirmButtonStatus.disable);
let monitorKeyboard: MonitorKeyboard | undefined = undefined;
const keyboardOpen = ref(false);

const user: UserItem = userStore.users!.items.get(userStore.current_id!)!;

terminusNameRef.value = user.name;

onMounted(() => {
	if ($q.platform.is.android) {
		monitorKeyboard = new MonitorKeyboard();
		monitorKeyboard.onStart();
		monitorKeyboard.onShow(() => (keyboardOpen.value = true));
		monitorKeyboard.onHidden(() => (keyboardOpen.value = false));
	}
	if (termipassStore.srpInvalid) {
		btnTitle.value = t('reconnect');
	} else if (termipassStore.ssoInvalid) {
		btnTitle.value = t('login.title');
	}
});

onUnmounted(() => {
	if ($q.platform.is.android) {
		if (monitorKeyboard) {
			monitorKeyboard.onEnd();
		}
	}
});

const calNextButtonEnable = () => {
	btnStatusRef.value =
		osPwd.value.length >= 6
			? ConfirmButtonStatus.normal
			: ConfirmButtonStatus.disable;
};

const pwdInputChange = () => {
	calNextButtonEnable();
};

const isLocalTest = computed(() => {
	return process.env.IS_PC_TEST;
});

const use_local = ref(false);

const onConfirm = async () => {
	if (!(await userStore.unlockFirst())) {
		return;
	}
	const mnemonic: MnemonicItem = userStore.users!.mnemonics.get(
		userStore.current_id!
	)!;
	loading.value = true;
	const pingResult: OlaresInfo | null = await getTerminusInfo(user); //terminus_name
	if (!pingResult) {
		notifyFailed(
			t(
				'errors.unable_to_connect_to_olares_please_check_if_the_machine_and_network_are_functioning_properly'
			)
		);
		loading.value = false;
		return;
	}
	try {
		await connectTerminus(
			user,
			mnemonic.mnemonic,
			osPwd.value,
			use_local.value
		);
		await loginTerminus(user, osPwd.value, true, use_local.value);

		busEmit('account_update', true);

		//connect success set terminusid
		user.olares_id = pingResult.id || pingResult.terminusId;
		user.os_version = pingResult.osVersion;
		await userStore.users!.items.update(user);
		await userStore.save();

		if (process.env.PLATFORM == 'DESKTOP' || getAppPlatform().isPad) {
			router.replace({ path: '/Files/Home/' });
		} else {
			router.replace({ path: '/home' });
		}
	} catch (e) {
		notifyFailed(e.message);
	} finally {
		loading.value = false;
	}
};
</script>

<style lang="scss" scoped>
.terminus-login-root {
	width: 100%;
	height: 100%;
	background: $background-2;
	position: relative;

	.terminus-login-page {
		width: 100%;
		height: 100%;

		&__image {
			width: 64px;
			height: 64px;
			margin-top: 62px;
			border-radius: 32px;
			overflow: hidden;
		}

		&__name {
			text-align: center;
			color: $ink-1;
			margin-top: 12px;
			width: calc(100%);
		}

		&__desc {
			text-align: center;
			color: $ink-2;
			width: calc(100%);
		}

		&__edit {
			margin-top: 20px;
			width: 100%;
		}

		&__box {
			height: calc(100vh - 572px - 138px);
		}

		&__scan {
			margin-top: 20px;
			margin-bottom: 10px;
			color: $blue-4;
		}
	}

	.terminus-login-root-button {
		width: calc(100% - 40px);
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
