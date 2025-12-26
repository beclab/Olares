<template>
	<bt-scroll-area
		class="full-width full-height"
		@scroll="onScroll"
		v-if="trend_recommends.length > 0 || firstLoading"
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
		<div v-else>
			<q-intersection
				once
				v-for="(entry, index) in trend_recommends"
				:key="entry.id"
				style="height: 132px"
				:root="rootElement"
				@visibility="handleVisibility(entry, $event)"
			>
				<library-entry-view
					:entry="entry"
					:selected="index === selectIndex"
					:show-read-status="true"
					:time-type="SORT_TYPE.PUBLISHED"
					@on-selected-change="onSelectedChange(index)"
					@on-entry-delete="onEntryDelete"
				/>
			</q-intersection>
		</div>
		<footer-loading-component :has-data="loadMoreEnable" />
	</bt-scroll-area>
	<empty-view v-else class="source-root" />
</template>

<script lang="ts" setup>
import FooterLoadingComponent from '../../../components/files/FooterLoadingComponent.vue';
import LibraryEntryView from '../../../components/rss/entry/LibraryEntryView.vue';
import { DefaultType, SimpleEntry, SORT_TYPE } from '../../../utils/rss-types';
import EmptyView from '../../../components/rss/EmptyView.vue';
import { useConfigStore } from '../../../stores/rss-config';
import { useReaderStore } from '../../../stores/rss-reader';
import { computed, onMounted, ref, watch } from 'vue';
import { useRssStore } from '../../../stores/rss';
import { fireImpression } from '../../../api/wise';
import { MenuType } from '../../../utils/rss-menu';
import { useQuasar } from 'quasar';

let loadingMore = false;
const selectIndex = ref(-1);
const showBarShadow = ref(false);
const loadMoreEnable = ref(true);
const firstLoading = ref(true);
const $q = useQuasar();
const rssStore = useRssStore();
const configStore = useConfigStore();
const readerStore = useReaderStore();
const rootElement = ref();

const props = defineProps({
	algorithm: {
		type: String,
		required: true
	}
});

onMounted(() => {
	console.log('trend source page');
	rootElement.value =
		window.self !== window.top ? document.documentElement : null;
	requestRecommend();
});

const onSelectedChange = async (index: number) => {
	selectIndex.value = index;
};

const onEntryDelete = async (url: string) => {
	rssStore.show_recommends = rssStore.show_recommends.filter((item) => {
		return item.url !== url;
	});
};

const trend_recommends = computed(() => {
	return rssStore.show_recommends.filter(
		(item) => item.source === props.algorithm
	);
});

watch(
	() => [
		trend_recommends.value,
		configStore.menuChoice.type,
		configStore.menuChoice.tab
	],
	() => {
		if (configStore.menuChoice.type === MenuType.Trend) {
			readerStore.setNavigationList(trend_recommends.value);
		}
	},
	{
		immediate: true
	}
);

const requestRecommend = async () => {
	console.log(firstLoading.value);
	console.log(loadingMore);
	if (trend_recommends.value.length == 0) {
		firstLoading.value = true;
		const { data, message } = await rssStore.getRecommendList(props.algorithm);
		if (!message) {
			loadMoreEnable.value = data.length == DefaultType.Limit;
		} else {
			$q.notify(message);
		}
		firstLoading.value = false;
	} else {
		if (loadingMore) {
			return;
		}
		loadingMore = true;
		const { data, message } = await rssStore.getRecommendList(props.algorithm);
		if (!message) {
			loadMoreEnable.value = data.length == DefaultType.Limit;
		} else {
			$q.notify(message);
		}
		loadingMore = false;
	}
};

const handleVisibility = async (entry: SimpleEntry, isVisible: boolean) => {
	console.log('handleVisibility');
	if (isVisible) {
		console.log(`Item ${entry.id} is visible`);
		fireImpression(entry.source, entry.id);
	} else {
		console.log(`Item ${entry.id} is not visible`);
	}
};

const onScroll = async (info: any) => {
	showBarShadow.value = info.verticalPosition > 0;

	if (
		loadingMore ||
		firstLoading.value ||
		!loadMoreEnable.value ||
		info.verticalSize <= 0
	) {
		return;
	}

	if (
		info.verticalPosition + info.verticalContainerSize >=
		info.verticalSize - 30
	) {
		console.log('trend scroll end');

		await requestRecommend();
	}
};
</script>

<style scoped lang="scss"></style>
