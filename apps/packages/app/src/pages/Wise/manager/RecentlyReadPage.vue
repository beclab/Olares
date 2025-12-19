<template>
	<div class="wise-page-root bg-color-white column justify-start">
		<title-bar>
			<template v-slot:before>
				<bt-breadcrumbs
					:title="t('main.recently_read')"
					icon="sym_r_history"
					margin="80px"
				/>
			</template>
			<template v-slot:after>
				<title-right-layout />
			</template>
		</title-bar>
		<div class="wise-page-content column justify-start">
			<q-tab-panels
				v-model="configStore.menuChoice.tab"
				animated
				class="wise-page-tab-panels"
				keep-alive
			>
				<q-tab-panel :name="TabType.Empty" class="wise-page-tab-panel">
					<bt-scroll-area
						class="full-width full-height"
						@scroll="onScroll"
						v-if="mergedRecords.length > 0 || firstLoading"
					>
						<q-list v-if="firstLoading">
							<div class="entry">
								<library-entry-view
									:skeleton="true"
									v-for="item in DefaultType.Limit"
									:key="item"
								/>
							</div>
						</q-list>
						<q-list v-else>
							<div
								v-for="(entry, index) in mergedRecords"
								:key="entry.id + entry.last_opened"
							>
								<library-entry-view
									:entry="entry"
									:selected="index === selectIndex"
									:show-read-status="false"
									:time="entry.last_opened"
									:time-prefix="t('base.last_opened')"
									@on-selected-change="onSelectedChange(index)"
									@on-entry-delete="onRecentlyDelete"
								/>
							</div>
						</q-list>
						<footer-loading-component :has-data="loadMoreEnable" />
					</bt-scroll-area>
					<empty-view v-else class="source-root" />
				</q-tab-panel>
			</q-tab-panels>
		</div>
	</div>
</template>

<script lang="ts" setup>
import FooterLoadingComponent from '../../../components/files/FooterLoadingComponent.vue';
import LibraryEntryView from '../../../components/rss/entry/LibraryEntryView.vue';
import TitleRightLayout from '../../../components/base/TitleRightLayout.vue';
import BtBreadcrumbs from '../../../components/base/BtBreadcrumbs.vue';
import EmptyView from '../../../components/rss/EmptyView.vue';
import TitleBar from '../../../components/rss/TitleBar.vue';
import { useConfigStore } from 'src/stores/rss-config';
import { DefaultType, Entry } from 'src/utils/rss-types';
import { useReaderStore } from 'src/stores/rss-reader';
import { getRecentlyEntryList } from 'src/api/wise';
import { CompareRecentlyEntry, extractHtml } from 'src/utils/rss-utils';
import { useRssStore } from 'src/stores/rss';
import { TabType } from 'src/utils/rss-menu';
import { onActivated } from 'vue-demi';
import { computed, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { binaryInsert } from 'src/utils/utils';
import cloneDeep from 'lodash/cloneDeep';

let loadingMore = false;
const { t } = useI18n();
const rssStore = useRssStore();
const selectIndex = ref(-1);
const showBarShadow = ref(false);
const loadMoreEnable = ref(true);
const firstLoading = ref(true);
const configStore = useConfigStore();
const readerStore = useReaderStore();
const historyList = ref([]);
const tempRecentlyList = ref([]);

onActivated(async () => {
	firstLoading.value = true;
	historyList.value = [];
	tempRecentlyList.value = cloneDeep(rssStore.recentlyList);
	await requestList();
	firstLoading.value = false;
});

const onSelectedChange = async (index: number) => {
	selectIndex.value = index;
};

const onRecentlyDelete = async (url: string, selected: boolean) => {
	rssStore.recentlyList = rssStore.recentlyList.filter(
		(item) => item.url !== url
	);
	tempRecentlyList.value = cloneDeep(rssStore.recentlyList);
	historyList.value = historyList.value.filter((item) => item.url !== url);
	await rssStore.removeEntry(url, selected);
};

const requestList = async () => {
	console.log(loadingMore);
	if (loadingMore) {
		return;
	}
	loadingMore = true;

	const listLength = await getRecentlyList();
	loadMoreEnable.value = listLength === DefaultType.Limit;
	loadingMore = false;
};

async function getRecentlyList(): Promise<number> {
	const list: Entry[] = await getRecentlyEntryList(
		historyList.value.length,
		DefaultType.Limit
	);
	for (let i = 0; i < list.length; i++) {
		list[i].summary = extractHtml(list[i]);
	}

	for (let i = 0; i < list.length; i++) {
		list[i].summary = extractHtml(list[i]);
	}
	for (let i = 0; i < list.length; i++) {
		const index = historyList.value.findIndex((l) => l.id == list[i].id);
		if (index >= 0) {
			historyList.value.splice(index, 1, list[i]);
		} else {
			binaryInsert<Entry>(historyList.value, list[i], CompareRecentlyEntry);
		}
	}
	return list.length;
}

const mergedRecords = computed(() => {
	const recordMap = new Map<string, Entry>();

	historyList.value.forEach((record) => {
		recordMap.set(record.id, record);
	});

	tempRecentlyList.value.forEach((item) => {
		recordMap.set(item.id, item);
	});

	const mergedList = Array.from(recordMap.values()).sort(CompareRecentlyEntry);

	readerStore.setNavigationList(mergedList);

	return mergedList;
});

const onScroll = async (info: any) => {
	showBarShadow.value = info.verticalPosition > 0;

	if (loadingMore || !loadMoreEnable.value || info.verticalSize <= 0) {
		return;
	}

	if (
		info.verticalPosition + info.verticalContainerSize >=
		info.verticalSize - 30
	) {
		console.log('trend scroll end');

		await requestList();
	}
};
</script>

<style scoped lang="scss"></style>
