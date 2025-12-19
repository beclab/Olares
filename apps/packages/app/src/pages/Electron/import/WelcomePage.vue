<template>
	<div class="welcome-page column justify-start items-center">
		<q-img
			class="welcome-page__logo"
			:src="getRequireImage('login/termipass_logo.svg')"
		/>
		<terminus-page-title
			:center="true"
			class="page-title"
			:label="t('Welcome to LarePass')"
			:desc="t('Your journey with Olares starts here!')"
		/>
		<confirm-button
			class="welcome-page__button"
			:btn-title="t('start')"
			@onConfirm="onConfirm"
			:btn-status="ConfirmButtonStatus.normal"
		/>
	</div>
</template>

<script lang="ts" setup>
import { useRouter } from 'vue-router';
import ConfirmButton from '../../../components/common/ConfirmButton.vue';
import { ConfirmButtonStatus } from '../../../utils/constants';
import { useI18n } from 'vue-i18n';
import TerminusPageTitle from '../../../components/common/TerminusPageTitle.vue';
import { getRequireImage } from '../../../utils/imageUtils';
import { saveDefaultPassword } from '../../../utils/UnlockBusiness';
import { getAppPlatform } from '../../../application/platform';
import { notifyFailed } from '../../../utils/notifyRedefinedUtil';

const router = useRouter();
const { t } = useI18n();

async function onConfirm() {
	// router.push({ path: '/setUnlockPassword' });
	await saveDefaultPassword({
		async onSuccess() {
			const jumpToRegisterDid = () => {
				if (getAppPlatform().isPad) {
					router.replace({
						name: 'setupSuccess'
					});
					return;
				}
				router.push({
					name: 'InputMnemonic'
				});
			};
			jumpToRegisterDid();
		},
		onFailure(message: string) {
			notifyFailed(message);
		}
	});
}
</script>

<style lang="scss" scoped>
.welcome-page {
	width: 100%;
	height: 100%;
	background: $background-1;
	padding-top: 20px;
	padding-left: 32px;
	padding-right: 32px;
	position: relative;

	&__logo {
		width: 64px;
		height: 64px;
		margin-top: 96px;
		margin-bottom: 16px;
	}

	&__button {
		position: absolute;
		bottom: 52px;
		width: calc(100% - 64px);
		left: 32px;
		right: 32px;
	}
}
</style>
