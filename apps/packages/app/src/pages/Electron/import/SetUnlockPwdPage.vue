<template>
	<div class="set-unlock-pwd-page">
		<PadInputContentScrollView>
			<terminus-page-title
				:center="true"
				class="page-title"
				:label="t('Password settings')"
				:desc="
					t(
						'This password is only used for unlocking {AppName} on this device',
						{
							AppName: 'LarePass'
						}
					)
				"
			/>

			<terminus-password-validator
				ref="passwordValidator"
				v-model:button-status="btnStatusRef"
				v-model:button-text="btnTextRef"
				@keyup.enter="verifyPassword"
			/>
		</PadInputContentScrollView>
		<confirm-button
			class="set-unlock-pwd-page__button"
			:btn-title="btnTextRef"
			@onConfirm="verifyPassword"
			@onError="clearData"
			:btn-status="btnStatusRef"
		/>
	</div>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { useRouter } from 'vue-router';
import { useQuasar } from 'quasar';
import ConfirmButton from '../../../components/common/ConfirmButton.vue';
import { useI18n } from 'vue-i18n';
import { ConfirmButtonStatus } from '../../../utils/constants';
import { savePassword } from '../../../utils/UnlockBusiness';
import TerminusPageTitle from '../../../components/common/TerminusPageTitle.vue';
import TerminusPasswordValidator from '../../../components/common/TerminusPasswordValidator.vue';
import { notifyFailed } from '../../../utils/notifyRedefinedUtil';
import TerminusTipDialog from '../../../components/dialog/TerminusTipDialog.vue';
import { getAppPlatform, getNativeAppPlatform } from 'src/application/platform';
import PadInputContentScrollView from '../../../components/ios/PadInputContentScrollView.vue';

const $router = useRouter();
const $q = useQuasar();
const { t } = useI18n();

const btnTextRef = ref(t('next'));
const btnStatusRef = ref<ConfirmButtonStatus>(ConfirmButtonStatus.disable);
const passwordValidator = ref();

function clearData() {
	passwordValidator.value.clearPassword();
}

const verifyPassword = async () => {
	const password = passwordValidator.value.getValidPassword();
	if (!password) {
		notifyFailed(t('password_not_meet_rules'));
		return;
	}
	$q.loading.show();
	await savePassword(password, {
		async onSuccess() {
			const jumpToRegisterDid = () => {
				if (getAppPlatform().isPad) {
					$router.replace({
						name: 'setupSuccess'
					});
					return;
				}
				$router.push({
					name: 'InputMnemonic'
				});
			};

			if (await getNativeAppPlatform().biometricKeyStore.isSupported()) {
				$q.dialog({
					component: TerminusTipDialog,
					componentProps: {
						title: t('biometric_unlock'),
						message: t('biometric_unlock_desc'),
						navigation: t('skip')
					}
				})
					.onOk(async () => {
						const result = await getNativeAppPlatform().openBiometric();
						if (!result.status) {
							notifyFailed(t('errors.set_biometry_failure'));
							// return;
						}
						jumpToRegisterDid();
					})
					.onCancel(() => {
						jumpToRegisterDid();
					});
				return;
			}

			jumpToRegisterDid();
		},
		onFailure(message: string) {
			notifyFailed(message);
		}
	});
	$q.loading.hide();
};
</script>

<style lang="scss" scoped>
.set-unlock-pwd-page {
	width: 100%;
	height: 100%;
	background: $background-2;
	padding-top: 20px;
	padding-left: 32px;
	padding-right: 32px;
	position: relative;

	&__button {
		position: absolute;
		bottom: 52px;
		width: calc(100% - 64px);
		left: 32px;
		right: 32px;
	}
}
</style>
