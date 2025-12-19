<template>
	<div class="row justify-center fit">
		<div class="unlock-content column justify-between items-center no-wrap">
			<q-scroll-area
				class="terminus-unlock-scroll"
				ref="scrollAreaRef"
				:class="
					$q.platform.is.mobile
						? 'scroll-area-conf-open1-mobile'
						: 'scroll-area-conf-open1'
				"
			>
				<div class="terminus-unlock-page column justify-start items-center">
					<q-img
						class="terminus-unlock-page__brand"
						:src="getRequireImage(logo)"
					/>
					<span class="terminus-unlock-page__desc login-sub-title">{{
						detailText
					}}</span>
					<terminus-edit
						v-model="passwordRef"
						:label="t('password')"
						:show-password-img="true"
						class="terminus-unlock-page__edit"
						@update:model-value="onTextChange"
						@keyup.enter="loginByPassword(passwordRef)"
					/>
					<div class="terminus-unlock-page__reminder text-body3 q-mt-md">
						{{ reminderText }}
					</div>
				</div>
			</q-scroll-area>
			<div class="bottom-content row items-center justify-center">
				<q-icon
					v-if="biometricIcon"
					size="48px"
					class="q-mb-md"
					:name="`sym_r_${biometricIcon}`"
					@click="unlockByBiometric"
				/>
				<confirm-button
					class="common-width"
					:btn-title="t('unlock.title')"
					:btn-status="btnStatusRef"
					@onConfirm="loginByPassword(passwordRef)"
				/>
				<q-btn
					v-if="cancel"
					class="q-mt-md common-width cancel"
					flat
					no-caps
					@click="onCancelClick"
				>
					<div>{{ t('cancel') }}</div>
				</q-btn>
			</div>
		</div>
	</div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import ConfirmButton from '../../common/ConfirmButton.vue';
import TerminusEdit from '../../common/TerminusEdit.vue';
import { getRequireImage } from '../../../utils/imageUtils';
import { ref } from 'vue';
import { ConfirmButtonStatus } from '../../../utils/constants';
import { notifyFailed } from '../../../utils/notifyRedefinedUtil';
import { unlockByPwd, unlockByDefaultPwd } from '../../../utils/UnlockBusiness';
import { useUserStore } from '../../../stores/user';
import { BiometryType } from '@capgo/capacitor-native-biometric';
import { onMounted } from 'vue';
import { useQuasar } from 'quasar';
import MonitorKeyboard from '../../../utils/monitorKeyboard';
import { watch } from 'vue';
import { app } from '../../../globals';
import { formatMinutesTime } from '../../../utils/utils';
import { sendUnlock } from '../../../utils/bexFront';
import { onUnmounted } from 'vue';
import { _t } from '../../../utils/i18n';
import { getNativeAppPlatform } from 'src/application/platform';

const props = defineProps({
	detailText: {
		type: String,
		required: false,
		default: ''
	},
	logo: {
		type: String,
		required: false,
		default: ''
	},
	cancel: {
		type: Boolean,
		default: true,
		required: false
	},
	biometryAutoUnlock: {
		type: Boolean,
		required: false,
		default: true
	}
});

const { t } = useI18n();
const biometricIcon = ref('');
const btnStatusRef = ref<ConfirmButtonStatus>(ConfirmButtonStatus.disable);
const passwordRef = ref('');
const biometricAvailable = ref(false);
const userStore = useUserStore();
const $q = useQuasar();
let monitorKeyboard: MonitorKeyboard | undefined = undefined;
const keyboardOpen = ref(false);

const lockTime = ref(app.settings.autoLockDelay);
const autoUnlock = ref(app.settings.autoLock);

const reminderText = ref(
	autoUnlock.value
		? _t('unlock.auth_lock_reminder', {
				time: formatMinutesTime(lockTime.value)
		  })
		: t('unlock.unauth_lock_reminder')
);

onMounted(async () => {
	await setBiometric();
	if ($q.platform.is.android) {
		monitorKeyboard = new MonitorKeyboard();
		monitorKeyboard.onStart();
		monitorKeyboard.onShow(() => (keyboardOpen.value = true));
		monitorKeyboard.onHidden(() => (keyboardOpen.value = false));
	}
	unlockByDefaultPwd(unlockResult);
});

onUnmounted(() => {
	if ($q.platform.is.android) {
		if (monitorKeyboard) {
			monitorKeyboard.onEnd();
		}
	}
});

const setBiometric = async () => {
	biometricAvailable.value = userStore.openBiometric;
	if (userStore.openBiometric) {
		try {
			const result =
				await getNativeAppPlatform().biometricKeyStore.isSupportedWithData();
			const type = result.biometryType;
			if (props.biometryAutoUnlock) {
				unlockByBiometric();
			}
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

function onTextChange() {
	btnStatusRef.value =
		passwordRef.value.length < 8 || passwordRef.value.length > 32
			? ConfirmButtonStatus.disable
			: ConfirmButtonStatus.normal;
}

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

const loginByPassword = async (password: string) => {
	await unlockByPwd(password, unlockResult);
};

const unlockResult = {
	onSuccess: async (data: any) => {
		sendUnlock();
		emit('unlockSuccess', data);
	},
	onFailure: (message: string) => {
		notifyFailed(message);
	}
};

const scrollAreaRef = ref();
watch(
	() => keyboardOpen.value,
	() => {
		if (keyboardOpen.value == true) {
			setTimeout(() => {
				scrollAreaRef.value.setScrollPosition('vertical', 1000, 300);
			}, 100);
		} else {
			scrollAreaRef.value.setScrollPosition('vertical', 0);
		}
	}
);

const onCancelClick = () => {
	emit('cancel');
};

const emit = defineEmits(['unlockSuccess', 'cancel']);
</script>

<style scoped lang="scss">
.unlock-content {
	width: 100%;
	height: 100%;
	max-width: 500px;

	.scroll-area-conf-open1 {
		height: calc(100% - 140px);
	}

	.scroll-area-conf-open1-mobile {
		height: calc(100% - 200px);
	}
	.terminus-unlock-scroll {
		width: 100%;

		.terminus-unlock-page {
			width: 100%;
			padding-left: 20px;
			padding-right: 20px;

			&__brand {
				margin-top: 80px;
				width: 120px;
			}

			&__desc {
				margin-top: 20px;
				text-align: center;
			}

			&__edit {
				margin-top: 64px;
				width: 100%;
			}
			&__box {
				height: calc(100% - 452px);
			}

			&__reminder {
				color: $light-blue-default;
				margin-bottom: 40px;
			}
		}
	}

	.bottom-content {
		width: 100%;
		.common-width {
			width: calc(100% - 40px);
		}
		.cancel {
			height: 48px;
			border-radius: 10px;
			// box-shadow: none;
			color: $light-blue-default;

			// &:before {
			// 	box-shadow: none;
			// }
		}
		padding-bottom: 32px;
	}
}
</style>
