<template>
	<PageCard :title="$t('collect')">
		<template #extra>
			<LinkToSetting :name="ROUTE_CONST.OPTIONS_COLLECT" />
		</template>
		<div v-if="appAbilitiesStore.wise.running">
			<EmptyData
				v-if="!validate.valid"
				:title="emptyDataDescTitle"
				class="absolute-center"
				btnHidden
			></EmptyData>
			<CollectionContent v-else></CollectionContent>
		</div>
		<EmptyUninstallWise v-else></EmptyUninstallWise>

		<SpinnerLoading
			class="absolute-center"
			:showing="collectSiteStore.loading && !collectSiteStore.hasCache"
			size="44px"
			:desc="$t('parsing_waitting')"
		></SpinnerLoading>
	</PageCard>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, provide, ref } from 'vue';
import { useCollectSiteStore } from 'src/stores/collect-site';
import CollectionContent from './CollectionContent.vue';
import { browser } from 'src/platform/interface/bex/browser/target';
import {
	URL_VALID_STATUS,
	UrlValidationResult,
	validateUrlWithReasonAsync
} from 'src/utils/url2';
import { useAppAbilitiesStore } from 'src/stores/appAbilities';
import PageCard from 'src/pages/Plugin/components/PageCard.vue';
import EmptyUninstallWise from 'src/pages/Plugin/components/EmptyUninstallWise.vue';
import {
	createTabChangeListenerInCurrentWindow,
	getCurrentTabInfo,
	TAB_CHANGE_TYPE
} from 'src/utils/bex/tabs';
import { COLLECT_THEME } from 'src/constant/provide';
import { BEX_COLLECT_THEME } from 'src/constant/theme';
import SpinnerLoading from 'src/components/common/SpinnerLoading.vue';
import EmptyData from 'src/pages/Plugin/components/EmptyData.vue';
import { useI18n } from 'vue-i18n';
import LinkToSetting from 'src/pages/Plugin/containers/LinkToSetting.vue';
import { ROUTE_CONST } from 'src/router/route-const';
import { useUserStore } from 'src/stores/user';
import { useBrowserCookieStore } from 'src/stores/settings/browserCookie';

const { t } = useI18n();
const browserCookieStore = useBrowserCookieStore();

provide(COLLECT_THEME, BEX_COLLECT_THEME);
let listener;

const appAbilitiesStore = useAppAbilitiesStore();

const collectSiteStore = useCollectSiteStore();
collectSiteStore.init();
const validate = ref<UrlValidationResult>({ valid: false });
const emptyDataDescTitle = computed(() =>
	validate.value.status === URL_VALID_STATUS.BLOCKED && validate.value.reason
		? validate.value.reason
		: t('no_data')
);

async function getInfo() {
	const tab = await getCurrentTabInfo();
	setData(tab);
}

async function setData(tab) {
	validate.value = await validateUrlWithReasonAsync(tab?.url);

	if (validate.value.valid) {
		collectSiteStore.search(tab.url);
		collectSiteStore.updateEntry({
			title: tab.title,
			url: tab.url,
			thumbnail: tab.favIconUrl
		});
	} else {
		collectSiteStore.reset();
	}
}

const handleTabInfo = (tab) => {
	if (tab) {
		setData(tab);
	}
};

const handleActivated = async (tab) => {
	handleTabInfo(tab);
};

onMounted(async () => {
	getInfo();

	const tab = await getCurrentTabInfo();
	const userStore = useUserStore();
	const url = userStore.getModuleSever('settings');
	if (userStore.current_user?.name) {
		browserCookieStore.init(
			tab,
			userStore.current_user?.name.split('@')[0],
			url
		);
	}

	listener = createTabChangeListenerInCurrentWindow(async (info) => {
		if (info.type === TAB_CHANGE_TYPE.STATUS_CHANGE) {
			return;
		}

		let tab = info.tab;
		if (!tab) {
			tab = await browser.tabs.get(info.tabId);
		}

		handleActivated(tab);

		if (userStore.current_user?.name) {
			browserCookieStore.init(
				tab,
				userStore.current_user?.name.split('@')[0],
				url
			);
		}
	});
});

onUnmounted(() => {
	listener && listener.remove();
});
</script>

<style></style>
