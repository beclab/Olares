<template>
	<div
		class="q-py-sm q-pl-md q-pr-sm bg-background-3 cookie-message-container row no-wrap justify-between items-center flex-gap-x-sm"
		v-if="cookieRequire"
	>
		<div class="text-negative text-body3">
			{{ cookieIcon.tooltip }}
		</div>
		<q-btn
			:color="theme?.btnDefaultColor"
			padding="12px 8px"
			class="btn-wrapper"
			no-caps
			:loading="collectSiteStore.loading || pushLoading"
			:text-color="theme?.btnTextDefaultColor"
			@click="cookieHandler"
		>
			<div class="relative-position row items-center cursor-pointer no-wrap">
				<img :src="cookieIcon.icon" class="cookie-icon" />
				<span class="q-ml-sm text-body3">{{ $t('upload_cookies') }}</span>
			</div>
		</q-btn>
	</div>
</template>

<script setup lang="ts">
import { inject } from 'vue';
import { COLLECT_THEME_TYPE } from 'src/constant/theme';
import { COLLECT_THEME } from 'src/constant/provide';
import { COOKIE_LEVEL, CookieStatusCode } from 'src/utils/rss-types';
import { useCollectSiteStore } from 'src/stores/collect-site';
import { useCookieStatus } from 'src/composables/bex/useCookieStatus';

const theme = inject<COLLECT_THEME_TYPE>(COLLECT_THEME);
const collectSiteStore = useCollectSiteStore();

const { cookieIcon, pushLoading, cookieHandler, cookieRequire } =
	useCookieStatus();
</script>

<style lang="scss" scoped>
.cookie-message-container {
	border-radius: 12px;
	.cookie-icon {
		width: 16px;
		height: 16px;
	}
	.submit-icon {
		position: absolute;
		bottom: 0;
		right: 0;
	}
	.btn-wrapper {
		flex: 0 0 80px;
		::v-deep(.q-btn__content) {
			line-height: 16px;
		}
	}
}
</style>
