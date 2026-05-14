<template>
	<div class="container-root column items-center justify-center">
		<!-- Title Bar -->
		<title-bar
			:use-back="true"
			:progress="readingProgressStore.progressPercentage"
			@on-back-click="onBackClick"
		>
			<template v-slot:after>
				<bt-scroll-area class="reading-title">
					<div
						class="full-width row justify-between items-center no-wrap"
						style="height: 56px"
					>
						<pdf-title-bar v-if="isPdf" />
						<div v-else />
						<common-title-bar />
					</div>
				</bt-scroll-area>
			</template>
		</title-bar>

		<!-- Content Area -->
		<div class="content-bg" id="content-bg">
			<q-scroll-area
				:thumb-style="scrollThumbStyle"
				ref="scrollAreaRef"
				class="content-scroll-area"
				@scroll="onScroll"
			>
				<!-- Empty State -->
				<empty-view
					v-if="!isEntryReadable"
					:is-table="true"
					:content="t('main.wise_is_loading_article')"
				/>

				<!-- Article Content -->
				<div
					v-else
					class="article-content-parent column justify-center items-center"
				>
					<div
						id="article-parent"
						class="article-content-root column justify-start items-start"
						:style="articleContentStyle"
					>
						<!-- Article Header (non-PDF) -->
						<article-header v-if="!isPdf" />

						<!-- Dynamic Reader Component -->
						<component
							:is="currentReader"
							ref="readerController"
							:style="readerStyle"
						/>
					</div>
				</div>
			</q-scroll-area>

			<!-- Sidebar -->
			<entry-topic v-if="readerStore.readingEntry" :pdf="isPdf" />
		</div>
	</div>
</template>

<script lang="ts" setup>
import EmptyView from '../../../components/rss/EmptyView.vue';
import TitleBar from '../../../components/rss/TitleBar.vue';
import PdfTitleBar from './pdf/components/PdfTitleBar.vue';
import ArticleHeader from './preview/ArticleHeader.vue';
import CommonTitleBar from './title/CommonTitleBar.vue';
import EntryTopic from './preview/EntryTopic.vue';
import { useReadingProgressStore } from '../../../stores/rss-reading-progress';
import { ENTRY_STATUS, FILE_TYPE } from '../../../utils/rss-types';
import { MenuType, SupportDetails } from '../../../utils/rss-menu';
import HotkeyManager from '../../../directives/hotkeyManager';
import { WISE_HOTKEY } from '../../../directives/wiseHotkey';
import { useConfigStore } from '../../../stores/rss-config';
import { useReaderStore } from '../../../stores/rss-reader';
import { extractHtml } from '../../../utils/rss-utils';
import { useRssStore } from '../../../stores/rss';
import { useRoute, useRouter } from 'vue-router';
import cloneDeep from 'lodash/cloneDeep';
import { useI18n } from 'vue-i18n';
import {
	computed,
	defineAsyncComponent,
	nextTick,
	onBeforeUnmount,
	onMounted,
	ref,
	watch
} from 'vue';

import {
	useReadingProgress,
	useTitleObserver,
	useReaderStyles,
	useScrollHandler
} from './composables';

// ==================== Props ====================
const props = defineProps({
	margin: { type: Number, default: 290 },
	lineHeight: { type: Number, default: 150 },
	maxWidthPercentage: { type: Number, default: 0 },
	fontFamily: { type: String, default: 'Robot' },
	highContrastText: { type: Boolean, default: false },
	justifyText: Boolean,
	fontSize: { type: Number, default: 20 }
});

// ==================== Stores ====================
const route = useRoute();
const router = useRouter();
const { t } = useI18n();
const rssStore = useRssStore();
const readerStore = useReaderStore();
const configStore = useConfigStore();
const readingProgressStore = useReadingProgressStore();

// ==================== Refs ====================
const scrollAreaRef = ref();
const readerController = ref();
const articleContentRef = ref<HTMLElement | null>(null);

// ==================== Composables Setup ====================

// Reader styles
const { cssVariables, contentMaxWidth, contentPadding, isSmallScreen } =
	useReaderStyles(props);

// Reading progress tracking
const { indicatorStyle, startObserving: startProgressObserving } =
	useReadingProgress(articleContentRef);

// Title observer for TOC highlighting
const { startObserving: startTitleObserving } = useTitleObserver(
	articleContentRef,
	(id) => readerStore.updateReadingTopic(id)
);

// Scroll handler
const {
	handleScroll: baseHandleScroll,
	isPdf,
	restoreScrollPosition,
	isScrollable
} = useScrollHandler({
	fileType: computed(() => readerStore.readingEntry?.file_type),
	onArticleProgress: (percentage) => {
		readingProgressStore.updateProgress(percentage);
	},
	onPdfPageChange: (scrollY) => {
		if (readerController.value?.changePageByScroll) {
			readerController.value.changePageByScroll(scrollY);
		}
	}
});

// ==================== Computed ====================

const isEntryReadable = computed(() => {
	if (!readerStore.readingEntry) return false;
	return [
		ENTRY_STATUS.Extracted,
		ENTRY_STATUS.Staging,
		ENTRY_STATUS.Completed
	].includes(readerStore.readingEntry.status);
});

const scrollThumbStyle = {
	right: '2px',
	borderRadius: '3px',
	backgroundColor: '#BCBDBE',
	width: '6px',
	height: '6px',
	opacity: '1'
};

const articleContentStyle = computed(() => ({
	padding: contentPadding.value,
	height: '100%',
	maxWidth: isPdf.value ? '100%' : contentMaxWidth.value,
	...cssVariables.value
}));

const readerStyle = computed(() => ({
	height:
		readerStore.readingEntry?.file_type === FILE_TYPE.EBOOK
			? 'calc(100vh - 200px)'
			: ''
}));

const currentReader = computed(() => {
	switch (readerStore.readingEntry?.file_type) {
		case FILE_TYPE.EBOOK:
			return defineAsyncComponent(() => import('./preview/RssEbookReader.vue'));
		case FILE_TYPE.PDF:
			return defineAsyncComponent(
				() => import('./pdf/components/PdfReader.vue')
			);
		default:
			return defineAsyncComponent(
				() => import('./preview/RssArticleReader.vue')
			);
	}
});

// ==================== Methods ====================

function onBackClick() {
	if (route.params.path) {
		if (route.params.path === MenuType.Entry) {
			configStore.setMenuType(MenuType.History);
		} else if (SupportDetails.includes(route.params.path as string)) {
			configStore.setMenuType(route.params.path as string);
		} else {
			configStore.setMenuType(route.params.path as string, {
				filterId: route.params.path as string
			});
		}
	} else {
		router.back();
	}
}

function onScroll(info: any) {
	baseHandleScroll(info);

	// Update total progress for articles
	if (
		readerStore.readingEntry?.file_type === FILE_TYPE.ARTICLE &&
		readerStore.readingEntry.full_content
	) {
		readingProgressStore.setTotalProgress(
			1,
			extractHtml(readerStore.readingEntry.full_content.trim()).length
		);
	}
}

function checkAndScrollToPosition() {
	if (
		readerStore.readingEntry?.file_type === FILE_TYPE.ARTICLE &&
		readerStore.readingEntry.progress
	) {
		restoreScrollPosition(scrollAreaRef, readerStore.readingEntry.progress);
	}
}

// ==================== Watchers ====================

watch(
	() => route.params.id,
	(id) => {
		if (id) {
			readingProgressStore.onreset(id as string);
			readerStore.entryUpdate(id as string).then((entry) => {
				readerStore.updateImpression(entry, { clicked: true });

				const tempEntry = cloneDeep(entry);
				tempEntry.last_opened = new Date().getTime() / 1000;
				rssStore.addRecentlyEntry(tempEntry);

				if (readerStore.readingEntry?.file_type === FILE_TYPE.ARTICLE) {
					// Reset scrollable state for new entry
					isScrollable.value = false;
					nextTick(() => {
						setTimeout(() => {
							// Only set to 100% if user hasn't scrolled (content is too short to scroll)
							if (!isScrollable.value) {
								readingProgressStore.updateProgress(1);
							}
						}, 2000);
					});
				}
			});
		} else {
			readerStore.clearNavigationList();
		}
	},
	{ immediate: true }
);

// Watch entry status: when entry is not readable (e.g. after retry),
// periodically re-fetch from store to pick up external polling updates
let statusPollingTimer: ReturnType<typeof setInterval> | null = null;

function stopStatusPolling() {
	if (statusPollingTimer) {
		clearInterval(statusPollingTimer);
		statusPollingTimer = null;
	}
}

watch(
	() => readerStore.readingEntry?.status,
	(status) => {
		if (
			status &&
			![ENTRY_STATUS.Completed, ENTRY_STATUS.Failed].includes(status)
		) {
			// Entry exists but is not readable, start polling to sync updates
			if (!statusPollingTimer) {
				statusPollingTimer = setInterval(async () => {
					if (route.params.id) {
						await readerStore.entryUpdate(route.params.id as string);
					}
				}, 3000);
			}
		} else {
			stopStatusPolling();
		}
	}
);

// Watch topic selection for jump
watch(
	() => readerStore.readingTopic,
	(newTopic) => {
		if (newTopic?.jump && articleContentRef.value) {
			const targetElement = document.getElementById(newTopic.id);
			targetElement?.scrollIntoView({ block: 'start' });
		}
	}
);

// ==================== Hotkeys ====================

let hotkeyMap: Record<string, () => void>;

function registerHotkeys() {
	HotkeyManager.setScope('reading');
	hotkeyMap = {
		[WISE_HOTKEY.ENTRY.BACK]: onBackClick
	};
	HotkeyManager.registerHotkeys(hotkeyMap, ['reading']);
}

// ==================== Lifecycle ====================

onMounted(() => {
	articleContentRef.value = document.getElementById('article-parent');
	readingProgressStore.startReading();
	registerHotkeys();

	nextTick(() => {
		setTimeout(() => {
			checkAndScrollToPosition();
			startTitleObserving();
			startProgressObserving();
		}, 1000);
	});
});

onBeforeUnmount(() => {
	stopStatusPolling();
	readingProgressStore.stopReading();
	if (hotkeyMap) {
		HotkeyManager.unregisterHotkeys(hotkeyMap);
	}
});
</script>

<style lang="scss" scoped>
.container-root {
	height: 100vh;
	width: 100%;
	overflow: hidden;

	.reading-title {
		width: calc(50% + 141px - 56px);
		height: 56px;
	}

	.content-bg {
		width: 100%;
		height: calc(100% - 56px);
		position: relative;

		.content-scroll-area {
			height: calc(100% - 3px);
			margin-top: 3px;
			width: 100%;

			.article-content-parent {
				width: 100%;
				height: 100%;
				padding-bottom: 10px;

				.article-content-root {
					width: 100%;
					height: 100%;
					position: relative;
				}
			}
		}
	}
}
</style>
