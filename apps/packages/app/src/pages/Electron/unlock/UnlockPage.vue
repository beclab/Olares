<template>
	<div class="terminus-unlock-page column justify-center items-center">
		<q-img
			class="terminus-unlock-page__brand"
			:src="
				getRequireImage(
					$q.dark.isActive
						? 'login/larepass_brand_desktop_dark.svg'
						: 'login/larepass_brand_desktop_light.svg'
				)
			"
		/>
		<div class="terminus-unlock-box column justify-start items-center">
			<span class="terminus-unlock-box__desc login-sub-title">{{
				t('Enter the password to unlock LarePass')
			}}</span>
			<terminus-edit
				v-model="passwordRef"
				:label="t('password')"
				:show-password-img="true"
				class="terminus-unlock-box__edit"
				@update:model-value="onTextChange"
				@keyup.enter="loginByPassword(passwordRef)"
			/>
			<q-icon
				v-if="biometricIcon"
				style="margin-top: 20px"
				size="48px"
				:name="`sym_r_${biometricIcon}`"
				@click="unlockByBiometric"
			/>
			<confirm-button
				class="terminus-unlock-box__button"
				:btn-title="t('unlock.title')"
				:btn-status="btnStatusRef"
				@onConfirm="loginByPassword(passwordRef)"
			/>
		</div>
	</div>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { useRouter } from 'vue-router';
import { useQuasar } from 'quasar';
import { ConfirmButtonStatus } from '../../../utils/constants';
import TerminusEdit from '../../../components/common/TerminusEdit.vue';
import ConfirmButton from '../../../components/common/ConfirmButton.vue';
import { getRequireImage } from '../../../utils/imageUtils';
import { useI18n } from 'vue-i18n';
import { unlockByPwd } from '../../Mobile/login/unlock/UnlockBusiness';
import { notifyFailed } from '../../../utils/notifyRedefinedUtil';
import { useUserStore } from '../../../stores/user';
import { BiometryType } from '@capgo/capacitor-native-biometric';
import { onMounted } from 'vue';
import MonitorKeyboard from '../../../utils/monitorKeyboard';
import { onUnmounted } from 'vue';
import { getNativeAppPlatform } from '../../../application/platform';

const router = useRouter();
const passwordRef = ref('');
const { t } = useI18n();
const btnStatusRef = ref<ConfirmButtonStatus>(ConfirmButtonStatus.disable);
const userStore = useUserStore();
let monitorKeyboard: MonitorKeyboard | undefined = undefined;
const keyboardOpen = ref(false);
const $q = useQuasar();

onMounted(async () => {
	await setBiometric();

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

function onTextChange() {
	btnStatusRef.value =
		passwordRef.value.length < 8 || passwordRef.value.length > 32
			? ConfirmButtonStatus.disable
			: ConfirmButtonStatus.normal;
}

const loginByPassword = async (password: string) => {
	await unlockByPwd(password, {
		async onSuccess(data: any) {
			if (data) {
				if (userStore.current_user) {
					if (userStore.current_user.name) {
						router.replace('/connectLoading');
					}
				}
			} else {
				router.replace({ name: 'setupSuccess' });
			}
		},
		onFailure(message: string) {
			notifyFailed(message);
		}
	});
};
const biometricAvailable = ref(false);
const biometricIcon = ref();

const setBiometric = async () => {
	biometricAvailable.value = userStore.openBiometric;
	if (userStore.openBiometric) {
		try {
			const result =
				await getNativeAppPlatform().biometricKeyStore.isSupportedWithData();
			const type = result.biometryType;
			unlockByBiometric();
			if (type === BiometryType.NONE) {
				biometricIcon.value = '';
				return;
			}
			if (
				type == BiometryType.FACE_ID ||
				type == BiometryType.FACE_AUTHENTICATION
			) {
				biometricIcon.value = 'ar_on_you';
				return;
			}

			if (
				type == BiometryType.TOUCH_ID ||
				type == BiometryType.FINGERPRINT ||
				type == BiometryType.MULTIPLE
			) {
				biometricIcon.value = 'fingerprint';
				return;
			}

			if (type == BiometryType.IRIS_AUTHENTICATION) {
				biometricIcon.value = 'motion_sensor_active';
				return;
			}
		} catch (error) {
			console.error(error);
			notifyFailed(error.message);
		}
	}
};

const unlockByBiometric = async () => {
	const password = await getNativeAppPlatform().unlockByBiometric();
	if (!password || password.length === 0) {
		notifyFailed(
			t('errors.biometric_verify_error_please_unlock_with_password_try_again')
		);
		return;
	}
	await loginByPassword(password);
};
</script>

<style scoped lang="scss">
.terminus-unlock-page {
	width: 100%;
	height: 100%;
	// background: $background-3;
	// padding-top: 40px;
	// padding-bottom: 40px;

	&__brand {
		width: 225px;
		// height: 62px;
	}

	.terminus-unlock-box {
		width: 400px;
		margin-top: 32px;
		border-radius: 12px;
		padding: 20px;
		background: $background-1;
		border: 1px solid $separator;

		&__desc {
			margin-top: 12px;
		}

		&__edit {
			margin-top: 20px;
			width: 100%;
		}

		&__button {
			margin-top: 30px;
			width: calc(100%);
		}

		.item {
			border: 1px solid $separator;
			padding: 5rem;
			margin: 0.25rem;
			position: relative;
		}

		.header {
			position: absolute;
			top: 0;
			left: 0;
		}
	}
}
</style>
