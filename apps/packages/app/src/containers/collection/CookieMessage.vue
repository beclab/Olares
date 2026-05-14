<template>
	<div
		:class="[!isBex ? 'cookie-message-container' : '']"
		v-if="collectSiteStore.errorCode"
	>
		<EmptyData
			v-if="errorPageConfig"
			:title="errorPageConfig.title"
			:subtitle="errorPageConfig.subtitle"
			:listItems="errorPageConfig.listItems"
			:btnLabel="errorPageConfig.btnLabel"
			:empty-icon="errorPageConfig.emptyIcon"
			class="absolute-center"
			btnHidden
		>
			<template
				#action
				v-if="collectSiteStore.errorCode === HTTP_STATUS_CODE.COOKIE_REQUIRED"
			>
				<CustomButton
					:label="errorPageConfig.btnLabel"
					:color="theme?.btnDefaultColor"
					:text-color="theme?.btnTextDefaultColor"
					class="q-mt-lg"
					:loading="collectSiteStore.loading || pushLoading"
					:icon="cookieUploadIcon"
					@click="cookieHandler(collectSiteStore.retryUrl || '')"
				></CustomButton>
			</template>

			<template
				#action
				v-else-if="collectSiteStore.errorCode === HTTP_STATUS_CODE.UNKNOWN"
			>
				<CustomButton
					:label="errorPageConfig.btnLabel"
					:color="theme?.btnDefaultColor"
					:text-color="theme?.btnTextDefaultColor"
					class="q-mt-lg"
					icon-font="sym_r_autorenew"
					@click="refreshHandler"
				></CustomButton>
			</template>
		</EmptyData>
	</div>
</template>

<script setup lang="ts">
import { computed, inject, ref } from 'vue';
import { COLLECT_THEME_TYPE } from 'src/constant/theme';
import { COLLECT_THEME } from 'src/constant/provide';
import { useCollectSiteStore } from 'src/stores/collect-site';
import { useCookieStatus } from 'src/composables/bex/useCookieStatus';
import { useI18n } from 'vue-i18n';
import EmptyData from 'src/pages/Plugin/components/EmptyData.vue';
import CustomButton from 'src/pages/Plugin/components/CustomButton.vue';
import cookieUploadIconDarkDark from 'src/assets/plugin/cookie-upload-dark.svg';
import cookieUploadIconLight from 'src/assets/plugin/cookie-upload.svg';
import { useQuasar } from 'quasar';
import emptyIcon1 from 'src/assets/plugin/empty2.svg';
import emptyIcon2 from 'src/assets/plugin/empty.svg';
import emptyIcon3 from 'src/assets/plugin/empty3.svg';
import emptyIcon4 from 'src/assets/plugin/empty4.svg';

const HTTP_STATUS_CODE = {
	FORBIDDEN: 403,
	INTERNAL_SERVER_ERROR: 500,
	COOKIE_REQUIRED: 501,
	SERVICE_UNAVAILABLE: 503,
	GATEWAY_TIMEOUT: 504,
	NETWORK: 505,
	UNAVAILABLE: 506,
	PRIVATE: 507,
	DELETED: 508,
	COPYRIGHT: 509,
	URL_INVALID: 510,
	AUTHORIZATION_FAILED: 511,
	BOT_DETECTED: 512,
	UNKNOWN: 513,
	SERVER_ERROR_MIN: 504,
	SERVER_ERROR_MAX: 600
} as const;

const isBex = ref(process.env.IS_BEX);

const { t } = useI18n();
const $q = useQuasar();

const theme = inject<COLLECT_THEME_TYPE>(COLLECT_THEME);
const collectSiteStore = useCollectSiteStore();

const { cookieIcon, pushLoading, cookieHandler, cookieRequire } =
	useCookieStatus();

const cookieUploadIcon = computed(() =>
	$q.dark.isActive ? cookieUploadIconDarkDark : cookieUploadIconLight
);
const errorPageConfig = computed(() => {
	if (!collectSiteStore.errorCode) {
		return null;
	}

	const errorCode = collectSiteStore.errorCode;
	let title = '';
	let subtitle = '';
	let listItems: string[] = [];
	let btnHidden = false;
	let btnLabel = '';
	let emptyIcon = emptyIcon2;

	switch (errorCode) {
		case HTTP_STATUS_CODE.FORBIDDEN:
			title = t('bex.enable_collect');
			subtitle = t('bex.collect_error_403_desc');
			emptyIcon = emptyIcon1;
			break;
		case HTTP_STATUS_CODE.INTERNAL_SERVER_ERROR:
			title = t('bex.enable_collect');
			subtitle = t('bex.collect_error_500_desc');
			listItems = [
				t('bex.collect_error_500_list_1'),
				t('bex.collect_error_500_list_2'),
				t('bex.collect_error_500_list_3')
			];
			emptyIcon = emptyIcon2;
			break;
		case HTTP_STATUS_CODE.COOKIE_REQUIRED:
			title = t('bex.collect_error_501');
			subtitle = t('bex.collect_error_501_desc');
			btnLabel = t('bex.cookie_upload_tooltip');
			emptyIcon = emptyIcon3;
			break;
		case HTTP_STATUS_CODE.SERVICE_UNAVAILABLE:
			title = t('bex.enable_collect');
			subtitle = t('bex.collect_error_503_desc');
			emptyIcon = emptyIcon2;
			break;
		case HTTP_STATUS_CODE.GATEWAY_TIMEOUT:
			title = t('bex.collect_error_504');
			subtitle = t('bex.collect_error_504_desc');
			emptyIcon = emptyIcon4;
			break;
		case HTTP_STATUS_CODE.NETWORK:
			title = t('bex.enable_collect');
			subtitle = t('bex.collect_error_505_desc');
			emptyIcon = emptyIcon4;
			btnLabel = t('bex.try_again');
			break;
		case HTTP_STATUS_CODE.UNAVAILABLE:
			title = t('bex.enable_collect');
			subtitle = t('bex.collect_error_506_desc');
			emptyIcon = emptyIcon2;
			break;
		case HTTP_STATUS_CODE.PRIVATE:
			title = t('bex.enable_collect');
			subtitle = t('bex.collect_error_507_desc');
			emptyIcon = emptyIcon2;
			break;
		case HTTP_STATUS_CODE.DELETED:
			title = t('bex.enable_collect');
			subtitle = t('bex.collect_error_508_desc');
			emptyIcon = emptyIcon2;
			break;
		case HTTP_STATUS_CODE.COPYRIGHT:
			title = t('bex.enable_collect');
			subtitle = t('bex.collect_error_509_desc');
			emptyIcon = emptyIcon2;
			break;
		case HTTP_STATUS_CODE.URL_INVALID:
			title = t('bex.enable_collect');
			subtitle = t('bex.collect_error_510_desc');
			emptyIcon = emptyIcon2;
			break;
		case HTTP_STATUS_CODE.AUTHORIZATION_FAILED:
			title = t('bex.enable_collect');
			subtitle = t('bex.collect_error_511_desc');
			emptyIcon = emptyIcon2;
			break;
		case HTTP_STATUS_CODE.BOT_DETECTED:
			title = t('bex.enable_collect');
			subtitle = t('bex.collect_error_512_desc');
			emptyIcon = emptyIcon2;
			break;
		case HTTP_STATUS_CODE.UNKNOWN:
			title = t('bex.enable_collect');
			subtitle = t('bex.collect_error_513_desc');
			emptyIcon = emptyIcon4;
			btnLabel = t('bex.try_again');
			break;
		default:
			if (
				errorCode > HTTP_STATUS_CODE.SERVER_ERROR_MIN &&
				errorCode < HTTP_STATUS_CODE.SERVER_ERROR_MAX
			) {
				title = t('bex.collect_error_5xx');
				subtitle = t('bex.collect_error_5xx_desc');
				emptyIcon = emptyIcon4;
				btnLabel = t('bex.try_again');
			} else {
				return null;
			}
	}

	return {
		title,
		subtitle,
		listItems,
		btnHidden,
		btnLabel,
		emptyIcon
	};
});
const refreshHandler = () => {
	if (collectSiteStore.retryUrl) {
		collectSiteStore.deleteCache(collectSiteStore.retryUrl);
		collectSiteStore.search(collectSiteStore.retryUrl);
	}
};
</script>

<style lang="scss" scoped>
.cookie-message-container {
	position: relative;
	height: 270px;
}
</style>
