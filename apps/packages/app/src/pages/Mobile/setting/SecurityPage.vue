<template>
	<terminus-title-bar :title="t('settings.safety')" />
	<TerminusScrollArea class="security-root">
		<template v-slot:content>
			<div class="q-mt-lg q-px-lg">
				<terminus-item
					v-if="isBex || userStore.currentUserBackup"
					:show-board="true"
					icon-name="sym_r_fact_check"
					class="q-mt-lg"
					:wholePictureSize="20"
					@click="startBackUp"
				>
					<template v-slot:title> {{ t('mnemonic_phrase') }}</template>

					<template v-slot:side>
						<q-icon
							name="sym_r_keyboard_arrow_right"
							size="20px"
							color="ink-3"
						/>
					</template>
				</terminus-item>
				<div
					v-else
					class="security-root__backup-mneminic q-mt-lg q-pl-lg q-py-lg row items-center justify-between"
					:class="!$q.dark.isActive ? 'backup-mnemonic-bg' : 'bg-background-1'"
				>
					<div class="security-root__backup-mneminic__introduce">
						<div class="title text-h5">
							{{ t('please_backup_your_mnemonic_phrase') }}
						</div>
						<div class="detail text-body2">
							{{ t('back_up_your_mnemonic_phrase_immediately_to_safe') }}
						</div>
						<div
							class="backup text-subtitle3 q-mt-md row items-center justify-center"
							@click="startBackUp"
						>
							{{ t('start_backup') }}
						</div>
					</div>
					<div class="security-root__backup-mneminic__image">
						<img src="../../../assets/account/backup_mnemonic_image.svg" />
					</div>
				</div>

				<terminus-item
					v-if="!isBex"
					:show-board="true"
					icon-name="sym_r_badge"
					:wholePictureSize="20"
					class="q-mt-lg"
					@click="enterVCManager"
				>
					<template v-slot:title> {{ t('vc_management') }}</template>

					<template v-slot:side>
						<q-icon
							name="sym_r_keyboard_arrow_right"
							size="20px"
							color="ink-3"
						/>
					</template>
				</terminus-item>

				<terminus-item
					v-if="$q.platform.is.nativeMobile"
					:show-board="true"
					icon-name="sym_r_ink_highlighter"
					:wholePictureSize="20"
					class="q-mt-lg"
					@click="setPath('/setting/autofill')"
				>
					<template v-slot:title> {{ t('autofill') }}</template>

					<template v-slot:side>
						<q-icon
							name="sym_r_keyboard_arrow_right"
							size="20px"
							color="ink-3"
						/>
					</template>
				</terminus-item>

				<terminus-item
					@click="changePwd"
					icon-name="sym_r_lock"
					class="q-mt-md"
					:wholePictureSize="20"
				>
					<template v-slot:title>
						<div class="text-subtitle2 security-root__title">
							{{
								t(isFirstSetPwd ? 'Set up a password' : 'change_local_password')
							}}
						</div>
					</template>
					<template v-slot:side>
						<q-icon name="keyboard_arrow_right" size="20px" color="grey-3" />
					</template>
				</terminus-item>

				<terminus-item
					class="q-mt-md"
					v-if="!isBex"
					:clickable="false"
					:icon-name="biometricIcon"
					:wholePictureSize="20"
				>
					<template v-slot:title>
						<div class="text-subtitle2 security-root__title">
							{{ t('use_biometrics') }}
						</div>
					</template>
					<template v-slot:side>
						<bt-switch
							size="sm"
							truthy-track-color="light-blue-default"
							v-model="unlockByBiometricStatus"
							@update:model-value="changeBiometric"
						/>
					</template>
				</terminus-item>

				<div class="lock-content q-mt-md">
					<div
						class="lock-content__header row items-center justify-between q-px-md"
					>
						<div class="row items-center">
							<q-icon name="sym_r_lock_clock" size="20px" color="ink-2" />
							<div class="lock-content__header__title text-subtitle2">
								{{ t('autolock.title') }}
							</div>
						</div>

						<bt-switch
							size="sm"
							truthy-track-color="light-blue-default"
							v-model="lockStatus"
							@update:model-value="changeAutoLock"
						/>
					</div>
					<div
						v-show="lockStatus"
						class="bg-separator"
						style="
							height: 1px;
							width: calc(100% - 60px);
							margin-left: 40px;
							margin-right: 20px;
						"
					></div>
					<div class="settings-module-root__content text-body2">
						<q-slide-transition>
							<div v-show="lockStatus">
								<div
									class="row items-center q-ml-md q-mr-md"
									style="height: 60px"
								>
									<!-- <div class="text-body3 ink-1 q-mr-sm">
										{{ t('after') }}
									</div> -->
									<q-slider
										v-model="lockTime"
										:step="5"
										:min="0"
										:max="3 * 24 * 60"
										style="width: auto; margin-left: 5px; flex: 1"
										color="yellow"
										@change="changeAutoLockDelay"
									/>
									<div class="text-body3 ink-1 q-ml-sm">
										{{ formatMinutesTime(lockTime) }}
									</div>
								</div>
							</div>
						</q-slide-transition>
					</div>
				</div>
				<q-item class="q-pa-none q-mt-md" v-if="lockStatus">
					<q-item-section class="userinfo">
						<q-item-label class="q-pt-md text-ink-2 q-mb-sm">
							{{ t('autolock.reminderTitle') }}
						</q-item-label>
						<q-item-label
							class="row text-body3"
							v-for="item in reminderList"
							:key="item"
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
							<div class="q-mb-sm text-ink-3" style="width: calc(100% - 20px)">
								{{ item }}
							</div>
						</q-item-label>
					</q-item-section>
				</q-item>
			</div>
		</template>
	</TerminusScrollArea>
</template>

<script lang="ts" setup>
import { onMounted, ref } from 'vue';
import { app } from '../../../globals';
import TerminusTitleBar from '../../../components/common/TerminusTitleBar.vue';
import { useUserStore } from '../../../stores/user';

import { useI18n } from 'vue-i18n';
import TerminusItem from '../../../components/common/TerminusItem.vue';
import { useRouter } from 'vue-router';

import { formatMinutesTime } from '../../../utils/utils';
import { notifyFailed } from '../../../utils/notifyRedefinedUtil';
import TerminusScrollArea from '../../../components/common/TerminusScrollArea.vue';
import { busEmit } from '../../../utils/bus';
import { BiometryType } from '@capgo/capacitor-native-biometric';
import { getNativeAppPlatform } from '../../../application/platform';

const userStore = useUserStore();
const isFirstSetPwd = ref(!userStore.passwordReseted);
const { t } = useI18n();
const $router = useRouter();
const isBex = ref(process.env.IS_BEX);

const lockTime = ref(app.settings.autoLockDelay);
const lockStatus = ref(app.settings.autoLock);
const reminderList = ref([
	t('autolock.reminder1'),
	t('autolock.reminder2'),
	t('autolock.reminder3')
]);

const changeAutoLockDelay = (value: number) => {
	app.setSettings({ autoLockDelay: value });
};

const changeAutoLock = (value: any) => {
	app.setSettings({ autoLock: value });
};

const unlockByBiometricStatus = ref<boolean>(userStore.openBiometric);

const changeBiometric = async () => {
	if (!(await userStore.unlockFirst())) {
		return;
	}
	let result = {
		status: false,
		message: ''
	};

	if (!userStore.openBiometric) {
		result = await getNativeAppPlatform().openBiometric();
	} else {
		result = await getNativeAppPlatform().closeBiometric();
	}
	unlockByBiometricStatus.value = userStore.openBiometric;
	if (!result.status && result.message.length > 0) {
		notifyFailed(result.message);
	}
};

const changePwd = async () => {
	if (!(await userStore.unlockFirst(undefined, { hide: true }))) {
		return;
	}

	$router.push({
		path: '/change_pwd'
	});
};

const selectionReport = ref([] as string[]);

onMounted(() => {
	setBiometric();
	let securityReport = app.account?.settings.securityReport;
	for (const key in securityReport) {
		if (Object.prototype.hasOwnProperty.call(securityReport, key)) {
			const element = securityReport[key];
			if (element) {
				selectionReport.value.push(key);
			}
		}
	}
});

const biometricIcon = ref('sym_r_fingerprint');

const setBiometric = async () => {
	if (userStore.openBiometric) {
		try {
			const result =
				await getNativeAppPlatform().biometricKeyStore.isSupportedWithData();
			const type = result.biometryType;
			if (type === BiometryType.NONE) {
				biometricIcon.value = '';
				return;
			}
			if (
				type == BiometryType.FACE_ID ||
				type == BiometryType.FACE_AUTHENTICATION
			) {
				biometricIcon.value = 'sym_r_ar_on_you';
				return;
			}

			if (
				type == BiometryType.TOUCH_ID ||
				type == BiometryType.FINGERPRINT ||
				type == BiometryType.MULTIPLE
			) {
				biometricIcon.value = 'sym_r_fingerprint';
				return;
			}

			if (type == BiometryType.IRIS_AUTHENTICATION) {
				biometricIcon.value = 'sym_r_motion_sensor_active';
				return;
			}
		} catch (error) {
			console.error(error);
			notifyFailed(error.message);
		}
	}
};

const startBackUp = async () => {
	if (!(await userStore.unlockFirst(undefined, { hide: true }))) {
		return;
	}
	if (!userStore.passwordReseted) {
		busEmit('configPassword');
		return;
	}
	$router.push({
		path: '/backup_mnemonics'
	});
};

const enterVCManager = async () => {
	if (!(await userStore.unlockFirst())) {
		return;
	}
	$router.push({
		path: '/vc_manage'
	});
};

const setPath = (path: string) => {
	$router.push({
		path
	});
};
</script>

<style lang="scss" scoped>
.lock-slider {
	height: 60px;
	transition: height 0.5s;
	overflow: hidden;
	min-height: 0 !important;
	padding-top: 0px !important;
	padding-bottom: 0px !important;
}

.security-root {
	width: 100%;
	height: calc(100% - 56px);

	.security-root__title {
		color: $ink-1;
	}

	.backup-mnemonic-bg {
		background: linear-gradient(
			127.05deg,
			#fffef7 4.41%,
			rgba(249, 254, 199, 0.5) 49.41%,
			rgba(243, 254, 194, 0.5) 84.8%
		);
	}

	&__backup-mneminic {
		border: 1px solid $separator;
		width: 100%;
		border-radius: 16px;
		position: relative;
		min-height: 190px;

		&__introduce {
			width: calc(100% - 135px);
			height: 100%;

			.title {
				color: $ink-1;
			}

			.detail {
				color: $ink-2;
			}

			.backup {
				background: $yellow;
				border-radius: 8px;
				height: 32px;
				text-align: center;
				color: $grey-10;
				width: 93px;
			}
		}

		&__image {
			width: 127px;
			position: absolute;
			right: 15px;
			bottom: 20px;
		}
	}

	.lock-content {
		width: 100%;
		border: 1px solid $separator;
		background-color: $background-1;
		border-radius: 12px;

		&__header {
			height: 64px;

			&__title {
				margin-left: 16px;
			}
		}
	}
}
</style>
