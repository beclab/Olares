<template>
	<page-container>
		<template v-slot:page>
			<div
				class="home-page"
				:style="{
					'--paddingTop': deviceStore.isMobile
						? '0'
						: globalConfig.isOfficial
						? '56px'
						: '0',
					'--paddingBottom': deviceStore.isMobile ? '0' : '56px'
				}"
			>
				<bt-select
					v-if="globalConfig.isOfficial && !deviceStore.isMobile"
					:transparent="true"
					v-model="settingStore.currentLanguage"
					:options="supportLanguages"
					@update:modelValue="settingStore.languageUpdate"
					class="select-wrapper"
					text-class="text-subtitle1"
				/>
				<app-store-body
					v-if="globalConfig.isOfficial || deviceStore.isMobile"
					:title="t('main.discover')"
					:bottom-separator="true"
					:padding-exclude-body="deviceStore.isMobile ? 12 : 44"
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
				</app-store-body>

				<template
					v-for="(item, index) in pageData"
					:key="generatePageDataItemKey(item, index)"
				>
					<app-store-body
						v-if="item.type === CONTENT_TYPE.TOPIC"
						:label="
							deviceStore.isMobile ? '' : getI18nValue(item.title, locale)
						"
						:padding-exclude-body="paddingX"
						:bottom-separator="index !== pageData.length - 1"
						:show-body="item.content.length > 0"
						:body-margin-bottom="deviceStore.isMobile ? 20 : 32"
						:body-margin-top="
							globalConfig.isOfficial || deviceStore.isMobile ? 12 : 20
						"
					>
						<template v-slot:loading>
							<app-card-grid rule="app-store-topic" show-size="3,2,2">
								<template v-slot:card>
									<topic-view :skeleton="true" />
								</template>
							</app-card-grid>
						</template>
						<template v-slot:body>
							<app-store-swiper
								:data-array="item.content"
								show-size="3, 2, 2"
								:padding-x="paddingX"
								:navigation-offsite="40"
							>
								<template v-slot:swiper="{ item, index }">
									<topic-view
										class="cursor-pointer"
										:item="{
											...(getI18nValue < TopicInfo > (item, locale))
										}"
										@click="onTopicClick(getI18nValue<TopicInfo>(item, locale))"
									/>
								</template>
							</app-store-swiper>
						</template>
					</app-store-body>

					<app-store-body
						v-if="item.type === CONTENT_TYPE.APP"
						:label="
							deviceStore.isMobile ? '' : getI18nValue(item.title, locale)
						"
						:padding-exclude-body="paddingX"
						:bottom-separator="index !== pageData.length - 1"
						:show-body="item.content.length > 0"
						:body-margin-bottom="32"
					>
						<template v-slot:body>
							<app-store-swiper
								:data-array="item.content"
								:padding-x="paddingX"
								:navigation-offsite="40"
							>
								<template v-slot:swiper="{ item }">
									<topic-app-view
										:item="getI18nValue < TopicInfo > (item, locale)"
									/>
								</template>
							</app-store-swiper>
						</template>
					</app-store-body>

					<app-store-body
						v-if="item.type === CONTENT_TYPE.RECOMMENDS"
						:style="{
							paddingLeft: paddingX + 'px',
							paddingRight: paddingX + 'px'
						}"
						:show-body="item.content.length > 0"
						:label="getI18nValue(item.title, locale)"
						:right="t('base.see_all')"
						:no-label-padding-bottom="true"
						:bottom-separator="index !== pageData.length - 1"
						@on-right-click="clickList(item.type)"
					>
						<template v-slot:body>
							<app-card-grid
								rule="app-store-application"
								:app-list="item.content"
							>
								<template v-slot:card="{ app }">
									<base-app-card
										:app-name="app"
										:source-id="settingStore.marketSourceId"
									/>
								</template>
							</app-card-grid>
						</template>
					</app-store-body>

					<app-store-body
						v-if="item.type === CONTENT_TYPE.TOP"
						:style="{
							paddingLeft: paddingX + 'px',
							paddingRight: paddingX + 'px'
						}"
						:show-body="item.content.length > 0"
						:label="t('top_app_on_terminus')"
						:right="t('base.see_all')"
						:no-label-padding-bottom="true"
						:bottom-separator="index !== pageData.length - 1"
						@on-right-click="clickList(item.type)"
					>
						<template v-slot:body>
							<app-card-grid
								rule="app-store-application"
								:app-list="item.content"
							>
								<template v-slot:card="{ app }">
									<base-app-card
										:app-name="app"
										:source-id="settingStore.marketSourceId"
									/>
								</template>
							</app-card-grid>
						</template>
					</app-store-body>

					<app-store-body
						v-if="item.type === CONTENT_TYPE.LATEST"
						:style="{
							paddingLeft: paddingX + 'px',
							paddingRight: paddingX + 'px'
						}"
						:show-body="item.content.length > 0"
						:label="t('latest_app_on_terminus')"
						:right="t('base.see_all')"
						:bottom-separator="index !== pageData.length - 1"
						@on-right-click="clickList(item.type)"
					>
						<template v-slot:body>
							<app-card-grid
								rule="app-store-application"
								:app-list="item.content"
							>
								<template v-slot:card="{ app }">
									<base-app-card
										:app-name="app"
										:source-id="settingStore.marketSourceId"
									/>
								</template>
							</app-card-grid>
						</template>
					</app-store-body>
				</template>
			</div>
		</template>
	</page-container>
</template>

<script lang="ts" setup>
import PageContainer from '../../../components/base/PageContainer.vue';
import AppStoreSwiper from '../../../components/base/AppStoreSwiper.vue';
import BtSelect from '../../../components/settings/base/BtSelect.vue';
import TopicAppView from '../../../components/topic/TopicAppView.vue';
import AppCardGrid from '../../../components/appcard/AppCardGrid.vue';
import AppStoreBody from '../../../components/base/AppStoreBody.vue';
import BaseAppCard from '../../../components/appcard/BaseAppCard.vue';
import TopicView from '../../../components/topic/TopicView.vue';
import { useDeviceStore } from '../../../stores/settings/device';
import { useSettingStore } from '../../../stores/market/setting';
import { useCenterStore } from '../../../stores/market/center';
import { useMenuStore } from '../../../stores/market/menu';
import globalConfig from '../../../api/market/config';
import { SelectorProps } from '../../../constant';
import cloneDeep from 'lodash/cloneDeep';
import { useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { computed } from 'vue';
import {
	CONTENT_TYPE,
	TRANSACTION_PAGE,
	getI18nValue,
	TopicInfo,
	generatePageDataItemKey
} from '../../../constant/constants';

const supportLanguages: SelectorProps[] = [
	{ value: 'zh-CN', label: 'ZH' },
	{ value: 'en-US', label: 'EN' }
];

const settingStore = useSettingStore();
const deviceStore = useDeviceStore();
const centerStore = useCenterStore();
const menuStore = useMenuStore();
const { t, locale } = useI18n();
const router = useRouter();
const pageData = computed(() => {
	const data = centerStore.pagesMap.get('All') ?? [];
	if (deviceStore.isMobile) {
		const topicItem = data.find((item) => item.type === CONTENT_TYPE.TOPIC);
		const otherItems = data.filter((item) => item.type !== CONTENT_TYPE.TOPIC);

		const insertThreshold = 1;
		if (
			(topicItem &&
				topicItem.content &&
				topicItem.content.length <= insertThreshold) ||
			!topicItem ||
			otherItems.length === 0
		) {
			return data;
		}

		const result = [];
		let otherIndex = 0;

		topicItem.content.forEach((contentItem, index) => {
			const clonedContent = cloneDeep(topicItem);
			clonedContent.content = [clonedContent.content[index]];
			clonedContent.ids = [clonedContent.ids[index]];
			result.push(clonedContent);

			if (
				(index + 1) % insertThreshold === 0 &&
				otherIndex < otherItems.length
			) {
				result.push(otherItems[otherIndex]);
				otherIndex++;
			}
		});

		if (otherIndex < otherItems.length) {
			result.push(...otherItems.slice(otherIndex));
		}

		return result;
	} else {
		return data;
	}
});
const paddingX = computed(() => {
	return deviceStore.isMobile ? 20 : 44;
});

const clickList = (type: string) => {
	router.push({
		name: TRANSACTION_PAGE.List,
		params: {
			categories: 'All',
			type: type
		}
	});
};

const onTopicClick = (info: TopicInfo) => {
	router.push({
		name: TRANSACTION_PAGE.TOPIC,
		params: {
			category: 'All',
			topicId: info.topicId
		},
		query: {
			showMenu: 'false'
		}
	});
};
</script>
<style lang="scss" scoped>
.home-page {
	height: calc(100% - var(--paddingTop));
	width: 100%;
	margin-top: var(--paddingTop);
	padding: 0 0 var(--paddingBottom) 0;
	position: relative;

	.select-wrapper {
		position: absolute;
		top: -40px;
		right: 52px;

		::v-deep(.q-btn.q-btn-item) {
			padding: 0 12px !important;
		}

		::v-deep(.q-btn-dropdown__arrow) {
			margin-left: 4px;
		}
	}

	.empty_view {
		height: 120px;
		width: 100%;
	}
}
</style>
