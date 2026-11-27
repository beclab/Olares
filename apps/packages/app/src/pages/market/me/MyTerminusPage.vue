<template>
	<page-container>
		<template v-slot:page>
			<div
				class="my-page-scroll"
				:style="{
					'--paddingTop': deviceStore.isMobile ? '0' : '56px',
					'--paddingX': deviceStore.isMobile ? '0' : '44px'
				}"
			>
				<app-store-body
					:title="t('main.my_terminus')"
					:title-separator="true"
					:padding-exclude-body="deviceStore.isMobile ? 12 : 0"
				>
					<template v-slot:left>
						<q-icon
							v-if="deviceStore.isMobile"
							size="20px"
							style="margin: 6px"
							name="sym_r_menu_open"
							class="cursor-pointer"
							@click="menuStore.setDrawerOpen(true)"
						/>
					</template>
					<template v-slot:right>
						<div class="row justify-end items-center">
							<bt-label
								v-if="!deviceStore.isMobile"
								name="sym_r_settings"
								:label="t('Settings')"
								@click="goPreferencePage"
							/>
							<bt-upload-chart v-if="!deviceStore.isMobile" class="q-ml-lg">
								<bt-label
									name="sym_r_upload_file"
									:label="t('my.upload_custom_chart')"
								/>
							</bt-upload-chart>
							<bt-label
								class="q-ml-lg"
								name="sym_r_assignment"
								:label="deviceStore.isMobile ? '' : t('my.logs')"
								@click="goLogPage"
							/>
						</div>
					</template>
				</app-store-body>

				<source-tabs-with-more
					v-model="selectedTab"
					:sources="showSource"
					:is-mobile="deviceStore.isMobile"
					:more-label="t('More')"
					:tab-height="deviceStore.isMobile ? '48px' : '80px'"
				/>

				<q-tab-panels
					v-model="selectedTab"
					animated
					class="my-page-panels"
					keep-alive
				>
					<q-tab-panel
						v-for="item in showSource"
						:key="item.id"
						:name="item.id"
						class="my-page-panel"
						:style="{ padding: deviceStore.isMobile ? '0 20px' : '0' }"
					>
						<market-remote-page
							:source-id="item.id"
							v-if="item.type === MARKET_SOURCE_TYPE.REMOTE"
						/>
						<market-local-page
							:source-id="item.id"
							v-if="item.type === MARKET_SOURCE_TYPE.LOCAL"
						/>
					</q-tab-panel>
				</q-tab-panels>
			</div>
		</template>
	</page-container>
</template>

<script lang="ts" setup>
import PageContainer from '../../../components/base/PageContainer.vue';
import BtUploadChart from '../../../components/base/BtUploadChart.vue';
import AppStoreBody from '../../../components/base/AppStoreBody.vue';
import BtLabel from '../../../components/base/BtLabel.vue';
import SourceTabsWithMore from './SourceTabsWithMore.vue';
import MarketRemotePage from './MarketRemotePage.vue';
import MarketLocalPage from './MarketLocalPage.vue';
import { useDeviceStore } from '../../../stores/settings/device';
import { useSettingStore } from '../../../stores/market/setting';
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue';
import { useAppStore } from '../../../stores/market/appStore';
import { useMenuStore } from '../../../stores/market/menu';
import SimpleWaiter from '../../../utils/simpleWaiter';
import { busOn, busOff } from '../../../utils/bus';
import { useRoute, useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { pushWithParent } from '../../../utils/marketBack';
import {
	MARKET_SOURCE_OFFICIAL,
	MARKET_SOURCE_TYPE,
	MarketSource,
	TRANSACTION_PAGE
} from '../../../constant/constants';

const { t } = useI18n();
const selectedTab = ref();
const router = useRouter();
const route = useRoute();
const rssWaiter = new SimpleWaiter();
const settingStore = useSettingStore();
const deviceStore = useDeviceStore();
const menuStore = useMenuStore();
const appStore = useAppStore();

const showSource = computed(() => {
	if (
		settingStore.initialized &&
		appStore.sources &&
		appStore.sources.length > 0
	) {
		return appStore.remoteSource.concat(appStore.localSource);
	}
	return [];
});

// Restore the selected source from `?source=` when it points to an existing
// source; returns false when no valid source query is present.
const applySourceFromQuery = (): boolean => {
	const sources = appStore.remoteSource.concat(appStore.localSource);
	if (sources.length === 0) {
		return false;
	}
	const querySource = route.query.source as string;
	if (
		querySource &&
		sources.some((item: MarketSource) => item.id === querySource)
	) {
		selectedTab.value = querySource;
		return true;
	}
	return false;
};

rssWaiter.waitForCondition(
	() =>
		settingStore.initialized && appStore.sources && appStore.sources.length > 0,
	() => {
		if (!applySourceFromQuery()) {
			selectedTab.value = appStore.remoteSource.concat(
				appStore.localSource
			)[0].id;
		}
	}
);

// Mirror the selected source into the URL so refresh / desktop window
// maximize / open-in-new-tab keep the same tab. Guarded by route name to avoid
// clobbering a sub-page's query while this page is kept alive.
watch(selectedTab, (val) => {
	if (!val || route.name !== TRANSACTION_PAGE.MyTerminus) {
		return;
	}
	if ((route.query.source as string) === val) {
		return;
	}
	router.replace({ query: { ...route.query, source: val } });
});

// Keep the tab in sync when the source query changes from outside (browser
// back / forward). Guarded against the echo of our own router.replace.
watch(
	() => route.query.source,
	(src) => {
		const next = (src as string) ?? '';
		if (!next || next === selectedTab.value) {
			return;
		}
		const sources = appStore.remoteSource.concat(appStore.localSource);
		if (sources.some((item: MarketSource) => item.id === next)) {
			selectedTab.value = next;
		}
	}
);

const goLogPage = () => {
	pushWithParent(router, {
		name: TRANSACTION_PAGE.Log
	});
};

const goPreferencePage = () => {
	pushWithParent(router, {
		name: TRANSACTION_PAGE.Preference
	});
};

const routeToUpload = () => {
	const localSource = appStore.localSource.find(
		(item: any) => item.id === MARKET_SOURCE_OFFICIAL.LOCAL.UPLOAD
	);
	if (localSource) {
		selectedTab.value = localSource.id;
	}
};

onMounted(() => {
	busOn('uploadOK', routeToUpload);
});

onBeforeUnmount(() => {
	busOff('uploadOK', routeToUpload);
});
</script>

<style scoped lang="scss">
.my-page-scroll {
	width: 100%;
	height: 100%;
	padding: var(--paddingTop) var(--paddingX) 0;

	.my-page-panels {
		width: 100%;
		height: calc(100vh - var(--paddingX) - 84px - 52px);

		.my-page-panel {
			width: 100%;
			height: 100%;
			padding: 0;
		}
	}
}
</style>
