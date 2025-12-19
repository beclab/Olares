<template>
	<div class="rest-account-root column justify-start items-start">
		<div
			class="rest-account-root__img column justify-center items-center"
			@click="onReturn"
		>
			<q-icon name="sym_r_chevron_left" size="20px" color="sub-title" />
		</div>
		<div class="rest-account-page column justify-start items-center">
			<span class="rest-account-page__title text-color-title text-h6">{{
				t('reset_account_warning')
			}}</span>
			<div class="rest-account-page__warning login-sub-title">
				{{ t('reset_account_warning_desc') }}
			</div>
			<confirm-button
				class="rest-account-page__button"
				:btn-title="t('confirm_reset')"
				:btn-status="ConfirmButtonStatus.normal"
				@onConfirm="resetAccount"
			/>
		</div>
	</div>
</template>

<script setup lang="ts">
import { ConfirmButtonStatus } from '../../../utils/constants';
import ConfirmButton from '../../../components/common/ConfirmButton.vue';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';

const { t } = useI18n();
const router = useRouter();

const resetAccount = () => {
	// router.push({ path : '/reset_account'})
};

const onReturn = () => {
	if (window.history.length <= 1) {
		return;
	}
	router.go(-1);
};
</script>

<style scoped lang="scss">
.rest-account-root {
	width: 100%;
	height: 100%;
	padding: 12px;
	border: 1px solid $separator;

	&__img {
		width: 32px;
		height: 32px;
	}

	.rest-account-page {
		width: 100%;
		height: calc(100% - 32px);
		padding-left: 20px;
		padding-right: 20px;
		position: relative;

		&__title {
			margin-top: 44px;
			color: $red;
			width: 100%;
		}

		&__warning {
			word-wrap: break-word;
			white-space: pre-wrap;
			margin-top: 20px;
			color: $red;
			width: 100%;
		}

		&__button {
			position: absolute;
			bottom: 40px;
			width: calc(100% - 40px);
			left: 20px;
			right: 20px;
		}
	}
}
</style>
