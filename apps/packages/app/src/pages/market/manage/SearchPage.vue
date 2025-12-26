<template>
	<page-container :title-height="deviceStore.isMobile ? 0 : 56">
		<template v-slot:page>
			<div
				class="update-scroll"
				:style="{ '--paddingX': deviceStore.isMobile ? '0px' : '44px' }"
			>
				<app-store-body
					:title="t('search')"
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
					<template v-slot:body>
						<div :style="{ padding: `0 ${deviceStore.isMobile ? 20 : 0}px` }">
							<div class="row search-input">
								<q-icon size="24px" name="sym_r_search" color="ink-1" />
								<q-input
									input-style="height:24px"
									borderless
									style="flex: 1"
									:placeholder="t('Enter a keyword to start searching')"
									class="q-mx-md"
									v-model="searchContent"
									@update:modelValue="search"
								/>
								<q-icon
									v-if="searchContent"
									size="24px"
									name="sym_r_close"
									class="cursor-pointer"
									color="ink-3"
									@click="clearSearch"
								/>
							</div>
							<empty-view
								v-if="applications.length === 0"
								:label="
									searchContent
										? t('no_find_app_in_search', { keyword: searchContent })
										: t('Quickly find the apps you need')
								"
								class="empty_view"
							/>
							<div
								v-else
								:class="
									deviceStore.isMobile
										? 'app-store-workflow-mobile'
										: 'app-store-workflow'
								"
								class="q-mt-lg"
							>
								<function-app-card
									v-for="item in applications"
									:key="item.application + item.sourceName"
									:app-name="item.application"
									:source-id="item.sourceName"
									:version-display-mode="VERSION_DISPLAY_MODE.PRIORITY_MY"
								/>
								<app-card-hide-border />
							</div>
						</div>
					</template>
				</app-store-body>
			</div>
		</template>
	</page-container>
</template>

<script setup lang="ts">
import AppCardHideBorder from '../../../components/appcard/AppCardHideBorder.vue';
import FunctionAppCard from '../../../components/appcard/FunctionAppCard.vue';
import PageContainer from '../../../components/base/PageContainer.vue';
import AppStoreBody from '../../../components/base/AppStoreBody.vue';
import EmptyView from '../../../components/base/EmptyView.vue';
import { VERSION_DISPLAY_MODE } from '../../../constant/constants';
import { useDeviceStore } from '../../../stores/settings/device';
import { useCenterStore } from '../../../stores/market/center';
import { useMenuStore } from '../../../stores/market/menu';
import SimpleWaiter from '../../../utils/simpleWaiter';
import { computed, onMounted, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { debounce } from 'lodash';
import { useRoute } from 'vue-router';

const applications = ref<{ application: string; sourceName: string }[]>([]);
const centerStore = useCenterStore();
const deviceStore = useDeviceStore();
const searchContent = ref('');
const menuStore = useMenuStore();
const { t } = useI18n();
const route = useRoute();
const waiter = new SimpleWaiter();

const calculateWeight = (appInfo: any, searchContent: string) => {
	if (!searchContent) return 0;
	const lowerSearch = searchContent.toLowerCase();
	let weight = 0;

	const WEIGHTS = {
		titleExact: 100,
		titleInclude: 20,
		descExact: 15,
		descInclude: 5,
		fullDescExact: 10,
		fullDescInclude: 3
	};

	let titleMatched = false;
	const titles = [
		appInfo?.app_title['en-US']?.toLowerCase(),
		appInfo?.app_title['zh-CN']?.toLowerCase()
	].filter(Boolean);

	if (titles.some((title) => title === lowerSearch)) {
		weight += WEIGHTS.titleExact;
		titleMatched = true;
	} else if (
		!titleMatched &&
		titles.some((title) => title.includes(lowerSearch))
	) {
		weight += WEIGHTS.titleInclude;
	}

	let descMatched = false;
	const descs = [
		appInfo?.app_description['en-US']?.toLowerCase(),
		appInfo?.app_description['zh-CN']?.toLowerCase()
	].filter(Boolean);

	if (descs.some((desc) => desc === lowerSearch)) {
		weight += WEIGHTS.descExact;
		descMatched = true;
	} else if (!descMatched && descs.some((desc) => desc.includes(lowerSearch))) {
		weight += WEIGHTS.descInclude;
	}

	let fullDescMatched = false;
	const fullDescs = [
		appInfo?.fullDescription?.['en-US']?.toLowerCase(),
		appInfo?.fullDescription?.['zh-CN']?.toLowerCase()
	].filter(Boolean);

	if (fullDescs.some((fullDesc) => fullDesc === lowerSearch)) {
		weight += WEIGHTS.fullDescExact;
		fullDescMatched = true;
	} else if (
		!fullDescMatched &&
		fullDescs.some((fullDesc) => fullDesc.includes(lowerSearch))
	) {
		weight += WEIGHTS.fullDescInclude;
	}
	return weight;
};

const search = debounce(() => {
	if (!searchContent.value) {
		applications.value = [];
		return;
	}
	applications.value = Array.from(centerStore.appSimpleInfoMap.entries())
		.map(([key, appInfo]) => {
			const weight = calculateWeight(
				appInfo.app_simple_info,
				searchContent.value
			);
			return { key, appInfo: appInfo.app_simple_info, weight };
		})
		.filter(({ weight }) => weight > 0)
		.sort((a, b) => b.weight - a.weight)
		.map(({ key }) => {
			const [sourceName, application] = key.split('_');
			return { sourceName, application };
		});
});

onMounted(() => {
	const initialKeyword = (route.query.keyword as string)?.trim();
	if (initialKeyword) {
		searchContent.value = initialKeyword;
		waiter.waitForCondition(
			() => {
				return Array.from(centerStore.appSimpleInfoMap.keys()).length > 0;
			},
			() => {
				search();
			}
		);
	}
});

watch(
	() => centerStore.appSimpleInfoMap,
	() => {
		search();
	},
	{
		deep: true
	}
);

const clearSearch = () => {
	searchContent.value = '';
	applications.value = [];
};
</script>

<style lang="scss" scoped>
.update-scroll {
	width: 100%;
	height: 100%;
	padding: 0 var(--paddingX);

	.empty_view {
		width: 100%;
		height: calc(100dvh - 112px - 64px);
	}

	.search-input {
		border-radius: 12px;
		border: 1px solid $input-stroke;
		height: 48px;
		padding: 12px 20px;
	}
}
</style>
