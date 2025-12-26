<template>
	<div class="change-password-root">
		<TerminusMobileTitleConfirmView
			:title="t(isFirstSetPwd ? 'Set up a password' : 'change_local_password')"
			:btn-title="btnTextRef"
			:btn-status="totalBtnStatusRef"
			:bottomLineHide="true"
			@on-confirm="verifyPassword"
			@on-error="clearData"
		>
			<template v-slot:content>
				<div
					class="change-password-root__scroll__page column justify-start items-center"
				>
					<terminus-edit
						v-if="!isFirstSetPwd"
						v-model="oldPasswordRef"
						:label="t('please_enter_the_old_password')"
						:show-password-img="true"
						class="change-password-root__scroll__page__edit"
						@update:model-value="oldPwdInputChange"
					/>

					<terminus-password-validator
						ref="passwordValidator"
						v-model:button-status="newPasswordStatusRef"
						v-model:button-text="btnTextRef"
						:repeat-enable="isFirstSetPwd"
					/>
				</div>
			</template>
		</TerminusMobileTitleConfirmView>
	</div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import { ref, watch } from 'vue';
import TerminusEdit from '../../../components/common/TerminusEdit.vue';
import { ConfirmButtonStatus } from '../../../utils/constants';
import TerminusPasswordValidator from '../../../components/common/TerminusPasswordValidator.vue';

import TerminusMobileTitleConfirmView from '../../../components/common/TerminusMobileTitleConfirmView.vue';
import { useUserStore, defaultPassword } from '../../../stores/user';
import { useRouter } from 'vue-router';
import { notifyFailed } from '../../../utils/notifyRedefinedUtil';
import { useQuasar } from 'quasar';
import { getNativeAppPlatform } from '../../../application/platform';

const { t } = useI18n();

const oldPasswordRef = ref();

const btnTextRef = ref(t('next'));

const passwordValidator = ref();

const router = useRouter();

const userStore = useUserStore();

const $q = useQuasar();

const newPasswordStatusRef = ref<ConfirmButtonStatus>(
	ConfirmButtonStatus.disable
);

const totalBtnStatusRef = ref(ConfirmButtonStatus.disable);

const isFirstSetPwd = ref(!userStore.passwordReseted);

if (isFirstSetPwd.value) {
	oldPasswordRef.value = defaultPassword;
}

watch(
	() => newPasswordStatusRef.value,
	() => {
		setButtonStatus();
	}
);

function oldPwdInputChange() {
	setButtonStatus();
}

function setButtonStatus() {
	if (!oldPasswordRef.value) {
		totalBtnStatusRef.value = ConfirmButtonStatus.disable;
		btnTextRef.value = t('next');
		return;
	}
	totalBtnStatusRef.value = newPasswordStatusRef.value;
}

function clearData() {
	passwordValidator.value.clearPassword();
	oldPasswordRef.value = '';
	oldPwdInputChange();
}

const verifyPassword = async () => {
	if (!oldPasswordRef.value) {
		return;
	}

	const newPassword = passwordValidator.value.getValidPassword();
	if (!newPassword) {
		notifyFailed(t('password_not_meet_rules'));
		return;
	}

	try {
		if (!userStore.users || userStore.users.locked) {
			notifyFailed(t('please_unlock_first'));
			return;
		}
		await userStore.users.unlock(oldPasswordRef.value).then(() => {
			resetPasswordConfirm();
		});
	} catch (error) {
		notifyFailed(t('wrong_password_please_try_again'));
	}
};

const resetPasswordConfirm = async () => {
	const newPassword = passwordValidator.value.getValidPassword();
	try {
		const resetPasswordStatus = await userStore.updateUserPassword(
			oldPasswordRef.value,
			newPassword
		);

		if (
			$q.platform.is.nativeMobile &&
			(userStore.openBiometric || isFirstSetPwd.value)
		) {
			const result = await getNativeAppPlatform().openBiometric();
			if (!result.status) {
				await userStore.updateOpenBiometricStatus(false);
			}
		}

		if (resetPasswordStatus.status) {
			// app.lock();
			router.back();
		} else {
			notifyFailed(resetPasswordStatus.message);
		}
	} catch (error) {
		if (error.message) {
			notifyFailed(error.message);
		}
	}
};
</script>

<style scoped lang="scss">
.change-password-root {
	widows: 100%;
	height: 100%;

	&__scroll {
		&__page {
			width: 100%;
			height: 100%;
			padding-left: 20px;
			padding-right: 20px;

			&__edit {
				margin-top: 20px;
				width: 100%;
			}
		}
	}
}
</style>
