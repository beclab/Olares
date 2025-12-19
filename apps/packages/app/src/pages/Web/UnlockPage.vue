<template>
	<div class="terminus-unlock-page column justify-center items-center">
		<q-img
			class="terminus-unlock-page__brand"
			:src="
				$q.dark.isActive
					? getRequireImage('login/vault_brand_web_dark.png')
					: getRequireImage('login/vault_brand_web_light.png')
			"
		/>
		<div class="terminus-unlock-box column justify-start items-center">
			<span class="terminus-unlock-box__desc login-sub-title">{{
				t('Enter the password to unlock Vault')
			}}</span>
			<terminus-edit
				v-model="passwordRef"
				:label="t('password')"
				:show-password-img="true"
				class="terminus-unlock-box__edit"
				@update:model-value="onTextChange"
				@keyup.enter="onSubmit"
			/>
			<confirm-button
				class="terminus-unlock-box__button"
				:btn-status="btnStatusRef"
				:btn-title="t('unlock.title')"
				@onConfirm="onSubmit()"
			/>
			<!-- :btn-status="btnStatusRef" -->
		</div>
	</div>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { useRouter } from 'vue-router';
import { useQuasar } from 'quasar';
import { ConfirmButtonStatus } from '../../utils/constants';
import TerminusEdit from '../../components/common/TerminusEdit.vue';
import ConfirmButton from '../../components/common/ConfirmButton.vue';
import { useI18n } from 'vue-i18n';
import { getRequireImage } from '../../utils/imageUtils';
import { MnemonicItem } from '@didvault/sdk/src/core';
import { useUserStore } from '../../stores/user';
import { app } from '../../globals';
import { notifyFailed } from '../../utils/notifyRedefinedUtil';

const $q = useQuasar();
const router = useRouter();
const passwordRef = ref('');
const { t } = useI18n();
const btnStatusRef = ref<ConfirmButtonStatus>(ConfirmButtonStatus.disable);
let failedCount = 0;
const userStore = useUserStore();

function onTextChange() {
	btnStatusRef.value =
		passwordRef.value.length < 8 || passwordRef.value.length > 32
			? ConfirmButtonStatus.disable
			: ConfirmButtonStatus.normal;
}

const loginByPassword = async (password: string) => {
	await userStore.unlock(password);

	if (!userStore.current_user) {
		router.push({ path: '/import_mnemonic' });
	} else {
		const mnemonicItem: MnemonicItem = userStore.current_mnemonic!;
		await app.load(undefined);
		await app.unlock(mnemonicItem.mnemonic);
		router.push('/items');
	}
};

async function onSubmit() {
	if (btnStatusRef.value === ConfirmButtonStatus.disable) return;

	await app.loaded;
	try {
		if (!passwordRef.value) {
			notifyFailed(t('password_not_empty'));
			return;
		}

		await loginByPassword(passwordRef.value);
	} catch (e) {
		notifyFailed(t('password_error'));
		failedCount++;
		if (failedCount > 3) {
			await app.logout();
			await userStore.clear();
			router.push({ path: '/setUnlockPassword' });
			notifyFailed(t('failed_to_unlock_too_many_times'));
			return;
		}
		failedCount++;
	}
}
</script>

<style scoped lang="scss">
.terminus-unlock-page {
	width: 100%;
	height: 100%;
	background: $background-1;

	&__brand {
		width: 124px;
	}

	.terminus-unlock-box {
		width: 400px;
		margin-top: 32px;
		border-radius: 12px;
		padding: 20px;
		background: $background-2;
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
	}
}
</style>
