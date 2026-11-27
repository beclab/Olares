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
import SimpleWaiter from '../../../utils/simpleWaiter';
import { usePreCheckStore } from '../../../stores/market/preCheck';
import {
	TRANSACTION_PAGE,
	VERSION_DISPLAY_MODE
} from '../../../constant/constants';
import { useSettingStore } from '../../../stores/market/setting';
import { useDeviceStore } from '../../../stores/settings/device';
import { useCenterStore } from '../../../stores/market/center';
import { useAppStore } from '../../../stores/market/appStore';
import { useMenuStore } from '../../../stores/market/menu';
import { intersection } from '../../../utils/utils';
import { nsfwApp } from '../../../constant/config';
import { onMounted, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { debounce } from 'lodash';

const applications = ref<{ application: string; sourceName: string }[]>([]);
const centerStore = useCenterStore();
const appStore = useAppStore();
const settingsStore = useSettingStore();
const preCheckStore = usePreCheckStore();
const deviceStore = useDeviceStore();
const searchContent = ref('');
const menuStore = useMenuStore();
const { t } = useI18n();
const route = useRoute();
const router = useRouter();
const waiter = new SimpleWaiter();

const syncKeywordToUrl = () => {
	// Only mutate the URL while this kept-alive page is the active route,
	// otherwise we'd clobber another page's query while cached in background.
	if (route.name !== TRANSACTION_PAGE.Search) {
		return;
	}
	const next = (searchContent.value || '').trim();
	if (((route.query.keyword as string) ?? '').trim() === next) {
		return;
	}
	const query = { ...route.query };
	if (next) {
		query.keyword = next;
	} else {
		delete query.keyword;
	}
	router.replace({ query });
};

interface FieldWeights {
	exact: number;
	all: number;
	some: number;
	prefix: number;
}

interface SearchScore {
	weight: number;
	isPrefix: boolean;
	titleLen: number;
}

const tokenize = (value: string): string[] =>
	value.toLowerCase().split(/\s+/).filter(Boolean);

const scoreField = (
	values: string[],
	tokens: string[],
	fullSearch: string,
	weights: FieldWeights
): number => {
	let best = 0;
	for (const value of values) {
		if (value === fullSearch) {
			best = Math.max(best, weights.exact);
			continue;
		}
		let allHit = true;
		let hit = 0;
		for (const token of tokens) {
			if (value.includes(token)) {
				hit++;
			} else {
				allHit = false;
			}
		}
		if (allHit) {
			const prefixBonus = value.startsWith(tokens[0]) ? weights.prefix : 0;
			best = Math.max(best, weights.all + prefixBonus);
		} else if (hit > 0) {
			best = Math.max(best, (hit / tokens.length) * weights.some);
		}
	}
	return best;
};

const calculateWeight = (appInfo: any, searchContent: string): SearchScore => {
	const tokens = tokenize(searchContent);
	if (tokens.length === 0) {
		return { weight: 0, isPrefix: false, titleLen: Number.MAX_SAFE_INTEGER };
	}
	const fullSearch = tokens.join(' ');

	const names = [appInfo?.app_name?.toLowerCase()].filter(Boolean) as string[];
	const titles = [
		appInfo?.app_title?.['en-US']?.toLowerCase(),
		appInfo?.app_title?.['zh-CN']?.toLowerCase()
	].filter(Boolean) as string[];
	const descs = [
		appInfo?.app_description?.['en-US']?.toLowerCase(),
		appInfo?.app_description?.['zh-CN']?.toLowerCase()
	].filter(Boolean) as string[];
	const fullDescs = [
		appInfo?.fullDescription?.['en-US']?.toLowerCase(),
		appInfo?.fullDescription?.['zh-CN']?.toLowerCase()
	].filter(Boolean) as string[];

	let weight = 0;
	weight += scoreField(names, tokens, fullSearch, {
		exact: 120,
		all: 60,
		some: 30,
		prefix: 10
	});
	weight += scoreField(titles, tokens, fullSearch, {
		exact: 100,
		all: 40,
		some: 20,
		prefix: 10
	});
	weight += scoreField(descs, tokens, fullSearch, {
		exact: 15,
		all: 8,
		some: 5,
		prefix: 0
	});
	weight += scoreField(fullDescs, tokens, fullSearch, {
		exact: 10,
		all: 4,
		some: 3,
		prefix: 0
	});

	const isPrefix = [...names, ...titles].some((value) =>
		value.startsWith(tokens[0])
	);
	const titleLen = titles.length
		? Math.min(...titles.map((title) => title.length))
		: Number.MAX_SAFE_INTEGER;

	return { weight, isPrefix, titleLen };
};

const search = debounce(() => {
	syncKeywordToUrl();
	if (!searchContent.value) {
		applications.value = [];
		return;
	}
	if (
		!preCheckStore.initialized ||
		!preCheckStore.systemResource ||
		preCheckStore.systemResource.nodes.length <= 0
	) {
		return;
	}
	applications.value = Array.from(appStore.appSimpleInfoMap.entries())
		.filter(([, appInfo]) => {
			return !(settingsStore.nsfw && nsfwApp(appInfo));
		})
		.filter(([, appInfo]) => {
			return centerStore.allSourcesApps.has(appInfo.app_simple_info.app_name);
		})
		.filter(([, appInfo]) => {
			const intersectedArray = intersection(
				preCheckStore.systemResource.nodes,
				appInfo.app_simple_info.support_arch
					? appInfo.app_simple_info.support_arch
					: []
			);
			return intersectedArray.length > 0;
		})
		.map(([key, appInfo]) => {
			const score = calculateWeight(
				appInfo.app_simple_info,
				searchContent.value
			);
			return { key, ...score };
		})
		.filter(({ weight }) => weight > 0)
		.sort((a, b) => {
			if (b.weight !== a.weight) return b.weight - a.weight;
			// Same weight: prefix matches first, then shorter title.
			if (a.isPrefix !== b.isPrefix) return a.isPrefix ? -1 : 1;
			return a.titleLen - b.titleLen;
		})
		.map(({ key }) => {
			const [sourceName, application] = key.split('_');
			return { sourceName, application };
		});
}, 200);

onMounted(() => {
	const initialKeyword = (route.query.keyword as string)?.trim();
	if (initialKeyword) {
		searchContent.value = initialKeyword;
		waiter.waitForCondition(
			() => {
				return Array.from(appStore.appSimpleInfoMap.keys()).length > 0;
			},
			() => {
				search();
			}
		);
	}
});

// Keep the input in sync when the keyword query changes from outside (e.g.
// browser back, or an external jump to /search?keyword=). Guarded against the
// echo of our own router.replace in syncKeywordToUrl.
watch(
	() => route.query.keyword,
	(kw) => {
		const next = ((kw as string) ?? '').trim();
		if (next === (searchContent.value || '').trim()) {
			return;
		}
		searchContent.value = next;
		search();
	}
);

watch(
	() => appStore.appSimpleInfoMap,
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
	syncKeywordToUrl();
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
