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

				<div
					class="tab-parent row justify-start items-center"
					:style="{ height: deviceStore.isMobile ? '48px' : '80px' }"
				>
					<q-tabs
						:class="
							deviceStore.isMobile ? 'my-page-tab-mobile' : 'my-page-tabs'
						"
						v-model="selectedTab"
						dense
						outside-arrows
						active-color="primary"
						indicator-color="transparent"
						align="left"
						:narrow-indicator="false"
					>
						<q-tab
							v-for="item in showSource"
							:key="item.id"
							style="padding: 0"
							:name="item.id"
							@click="logStatus(item.id)"
						>
							<template v-slot:default>
								<market-tab-item
									:selected="item.id === selectedTab"
									:label="item.name"
								/>
							</template>
						</q-tab>
					</q-tabs>
				</div>

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
import BtTabItem from '../../../components/base/BtTabItem.vue';
import BtLabel from '../../../components/base/BtLabel.vue';
import MarketRemotePage from './MarketRemotePage.vue';
import MarketLocalPage from './MarketLocalPage.vue';
import { useDeviceStore } from '../../../stores/settings/device';
import { useSettingStore } from '../../../stores/market/setting';
import { useMenuStore } from '../../../stores/market/menu';
import { computed, onBeforeUnmount, onMounted, ref } from 'vue';
import { useCenterStore } from '../../../stores/market/center';
import SimpleWaiter from '../../../utils/simpleWaiter';
import { busOn, busOff } from '../../../utils/bus';
import { useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import {
	MARKET_SOURCE_OFFICIAL,
	MARKET_SOURCE_TYPE,
	TRANSACTION_PAGE
} from '../../../constant/constants';
import MarketTabItem from '../../../components/appintro/MarketTabItem.vue';

const { t } = useI18n();
const selectedTab = ref();
const router = useRouter();
const rssWaiter = new SimpleWaiter();
const centerStore = useCenterStore();
const settingStore = useSettingStore();
const deviceStore = useDeviceStore();
const menuStore = useMenuStore();

const showSource = computed(() => {
	if (
		settingStore.initialized &&
		centerStore.sources &&
		centerStore.sources.length > 0
	) {
		return centerStore.remoteSource.concat(centerStore.localSource);
	}
	return [];
});

rssWaiter.waitForCondition(
	() =>
		settingStore.initialized &&
		centerStore.sources &&
		centerStore.sources.length > 0,
	() => {
		selectedTab.value = centerStore.remoteSource.concat(
			centerStore.localSource
		)[0].id;
	}
);

const goLogPage = () => {
	router.push({
		name: TRANSACTION_PAGE.Log
	});
};

const goPreferencePage = () => {
	router.push({
		name: TRANSACTION_PAGE.Preference
	});
};

const routeToUpload = () => {
	const localSource = centerStore.localSource.find(
		(item) => item.id === MARKET_SOURCE_OFFICIAL.LOCAL.UPLOAD
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

const logStatus = (sourceId) => {
	console.log(centerStore.appStatusMap);
	console.log(centerStore.getSourceInstalledApp(sourceId));
};
</script>

<style scoped lang="scss">
.my-page-scroll {
	width: 100%;
	height: 100%;
	padding: var(--paddingTop) var(--paddingX) 0;

	.tab-parent {
		width: 100%;
		max-width: 100%;
		overflow: scroll;
	}

	.my-page-tabs {
		border-radius: 8px;
		border: 1px solid $separator;
	}

	.my-page-tab-mobile {
		width: 100%;
		padding: 0 8px;
		box-shadow: 0 2px 8px -2px rgba(0, 0, 0, 0.1);
	}

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

.q-tabs--dense .q-tab {
	min-height: 32px;
}
</style>
