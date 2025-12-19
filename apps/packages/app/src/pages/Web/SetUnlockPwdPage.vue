<template>
	<div class="set-unlock-pwd-page column justify-start items-center">
		<terminus-page-title
			:center="true"
			class="page-title"
			:label="t('Password settings')"
			:desc="
				t('This password is only used for unlocking {AppName} on this device', {
					AppName: 'Vault'
				})
			"
			imgSrc="/vault/vault-logo.svg"
		/>

		<terminus-password-validator
			ref="passwordValidator"
			v-model:button-status="btnStatusRef"
			v-model:button-text="btnTextRef"
			@keyup.enter="verifyPassword"
		/>

		<confirm-button
			class="set-unlock-pwd-page__button"
			:btn-title="btnTextRef"
			@onConfirm="verifyPassword"
			@onError="clearData"
			:btn-status="btnStatusRef"
		/>
	</div>
</template>

<script lang="ts" setup>
import { ref } from 'vue';
import { useRouter } from 'vue-router';
import { useUserStore } from '../../stores/user';
import { useQuasar } from 'quasar';
import TerminusPageTitle from '../../components/common/TerminusPageTitle.vue';
import TerminusPasswordValidator from '../../components/common/TerminusPasswordValidator.vue';
import ConfirmButton from '../../components/common/ConfirmButton.vue';
import { useI18n } from 'vue-i18n';
import { ConfirmButtonStatus } from '../../utils/constants';
import { notifyFailed } from '../../utils/notifyRedefinedUtil';

const $q = useQuasar();
const { t } = useI18n();
const btnTextRef = ref(t('next'));
const btnStatusRef = ref<ConfirmButtonStatus>(ConfirmButtonStatus.disable);
const passwordValidator = ref();
const router = useRouter();

function clearData() {
	passwordValidator.value.clearPassword();
}

const verifyPassword = async () => {
	try {
		const password = passwordValidator.value.getValidPassword();
		if (!password) {
			notifyFailed(t('password_not_meet_rules'));
			return;
		}

		$q.loading.show();

		const userStore = useUserStore();
		await userStore.create(password);
		userStore.password = password;
		router.push('/import_mnemonic');

		$q.loading.hide();
	} catch (e) {
		notifyFailed(e.message);
	} finally {
		$q.loading.hide();
	}
};
</script>

<style lang="scss" scoped>
.set-unlock-pwd-page {
	width: 100%;
	height: 100%;
	background: $background-2;
	padding-left: 20px;
	padding-right: 20px;
	position: relative;
	border: 1px solid $separator;

	&__button {
		position: absolute;
		bottom: 20px;
		width: calc(100% - 40px);
		left: 20px;
		right: 20px;
	}
}
</style>
