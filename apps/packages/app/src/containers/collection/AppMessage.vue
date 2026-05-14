<template>
	<div class="q-pa-md column flex-gap-y-md sdk-messge-container">
		<div>
			<div class="row items-center flex-gap-x-xs">
				<q-icon
					name="sym_r_error"
					color="red-default"
					style="margin-top: 0px"
					size="12px"
				/>
				<div class="text-negative text-subtitle3">
					{{ title }}
				</div>

				<q-btn
					v-if="appName"
					color="orange-default"
					padding="8px 24px"
					class="btn-wrapper"
					no-caps
					text-color="white"
					@click="onClickHandler"
				>
				</q-btn>
			</div>
			<div class="app-skd-required-tooltip text-overline text-ink-3">
				{{ t('bex.required_plugin_not_available') }}
			</div>
		</div>
		<SDKSearch
			v-for="value in sdk"
			:title="value"
			:key="value"
			@search="onClickHandler(value)"
		></SDKSearch>
	</div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import { useUserStore } from 'src/stores/user';
import { useConfigStore } from 'src/stores/rss-config';
import SDKSearch from './SDKSearch.vue';
import { computed } from 'vue';
import {
	useAppAbilitiesStore,
	defaultData,
	APP_KEYS
} from 'src/stores/appAbilities';
import { useCollectSiteStore } from 'src/stores/collect-site';

interface Props {
	message: string;
	appName?: string;
	type: 'collect' | 'download' | 'feed';
}

const props = defineProps<Props>();
const appAbilitiesStore = useAppAbilitiesStore();
const collectSiteStore = useCollectSiteStore();

const title = computed(() => {
	switch (props.type) {
		case 'collect':
			return t('bex.unable_to_collect');

		case 'download':
			return t('bex.unable_to_get_downloadable_files');

		case 'feed':
			return t('bex.unable_to_get_rss_feed');
	}

	return '';
});
const sdk = computed(() => {
	switch (props.type) {
		case 'collect':
			return props.message ? collectSiteStore.data.entry_plugin_dependency : [];

		case 'download':
			return props.message
				? collectSiteStore.data.download_plugin_dependency
				: [];

		case 'feed':
			return props.message ? collectSiteStore.data.feed_plugin_dependency : [];
	}

	return [];
});
const { t } = useI18n();

const onClickHandler = (appName) => {
	if (process.env.PLATFORM_BEX_ALL) {
		const userStore = useUserStore();
		const url =
			userStore.getModuleSever('market') + '/search?keyword=' + appName;
		window.open(url, '_blank');
	} else {
		const configStore = useConfigStore();
		const url =
			configStore.getModuleSever('market') + '/search?keyword=' + appName;
		window.open(url, '_blank');
	}
};
</script>

<style lang="scss" scoped>
.sdk-messge-container {
	border-radius: 12px;
	border: 1px solid $separator;
}
.app-skd-required-tooltip {
	margin-left: 17px;
}
.cookie-message-container {
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
		::v-deep(.q-btn__content) {
			line-height: 16px;
		}
	}
}
</style>
