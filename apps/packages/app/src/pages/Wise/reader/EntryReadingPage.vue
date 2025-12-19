<template>
	<div class="container-root column items-center justify-center">
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
						<pdf-title-bar v-if="isPDF" />
						<div v-else />
						<common-title-bar />
					</div>
				</bt-scroll-area>
			</template>
		</title-bar>

		<div class="content-bg" id="content-bg">
			<q-scroll-area
				:thumb-style="{
					right: '2px',
					borderRadius: '3px',
					backgroundColor: '#BCBDBE',
					width: '6px',
					height: '6px',
					opacity: '1'
				}"
				ref="scrollAreaRef"
				class="content-scroll-area"
				@scroll="onScroll"
			>
				<empty-view
					v-if="!isEntryReadable"
					:is-table="true"
					:content="t('main.wise_is_loading_article')"
				/>

				<div
					v-else
					class="article-content-parent column justify-center items-center"
				>
					<div
						id="article-parent"
						class="article-content-root column justify-start items-start"
						:style="{
							padding: $q.screen.sm || $q.screen.xs ? 15 : 30,
							height: '100%',
							maxWidth: isPDF
								? '100%'
								: $q.screen.sm || $q.screen.xs
								? maxWidthStyles.small
								: maxWidthStyles.default,
							'--text-align': justifyTextValue(
								justifyTextOverride ?? justifyText
							),
							'--text-font-size': `${styles.fontSize}px`,
							'--line-height': `${styles.lineHeight}%`,
							'--blockquote-padding':
								$q.screen.sm || $q.screen.xs ? '1em 2em' : '0.5em 1em',
							'--blockquote-icon-font-size':
								$q.screen.sm || $q.screen.xs ? '1.7rem' : '1.3rem',
							'--figure-margin':
								$q.screen.sm || $q.screen.xs ? '2.6875rem auto' : '1.6rem auto',
							'--hr-margin': $q.screen.sm || $q.screen.xs ? '2em' : '1em',
							'--font-color': styles.readerFontColor,
							'--table-header-color': styles.readerTableHeaderColor
						}"
					>
						<article-header v-if="!isPDF" />
						<!--						<div class="reading-indicator" :style="indicatorStyle" />-->
						<component
							:is="currentReader"
							ref="pdfController"
							:style="{
								height:
									readerStore.readingEntry &&
									readerStore.readingEntry.file_type === FILE_TYPE.EBOOK
										? 'calc(100vh - 200px)'
										: ''
							}"
						/>
					</div>
				</div>
			</q-scroll-area>
			<entry-topic v-if="readerStore.readingEntry" :pdf="isPDF" />
		</div>
	</div>
</template>
<script lang="ts" setup>
import { useReadingProgressStore } from '../../../stores/rss-reading-progress';
import { MenuType, SupportDetails } from '../../../utils/rss-menu';
import HotkeyManager from '../../../directives/hotkeyManager';
import EmptyView from '../../../components/rss/EmptyView.vue';
import { useConfigStore } from '../../../stores/rss-config';
import TitleBar from '../../../components/rss/TitleBar.vue';
import { useReaderStore } from '../../../stores/rss-reader';
import { extractHtml } from '../../../utils/rss-utils';
import ArticleHeader from './preview/ArticleHeader.vue';
import CommonTitleBar from './title/CommonTitleBar.vue';
import { ENTRY_STATUS, FILE_TYPE } from '../../../utils/rss-types';
import EntryTopic from './preview/EntryTopic.vue';
import PdfTitleBar from './title/PdfTitleBar.vue';
import { useRssStore } from '../../../stores/rss';
import { useRoute, useRouter } from 'vue-router';
import cloneDeep from 'lodash/cloneDeep';
import { useColor } from '@bytetrade/ui';
import '../../../css/omnivore.scss';
import { useQuasar } from 'quasar';
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
import { WISE_HOTKEY } from '../../../directives/wiseHotkey';

const route = useRoute();
const { t } = useI18n();
const router = useRouter();
const rssStore = useRssStore();
const readerStore = useReaderStore();
const configStore = useConfigStore();
const readingProgressStore = useReadingProgressStore();

const scrollAreaRef = ref();
const showBarShadow = ref(false);
const isScrollAble = ref(false);
const pdfController = ref();
const indicatorStyle = ref({
	transform: 'translateY(0px)',
	height: '0px'
});
let intersectionTrackingObserver = null;
let intersectionTitleObserver = null;
// const clickedElements: HTMLElement[] = [];

watch(
	() => route.params.id,
	(id) => {
		if (id) {
			readingProgressStore.onreset(id);
			readerStore.entryUpdate(id as string).then((entry) => {
				readerStore.updateImpression(entry, { clicked: true });
				const tempEntry = cloneDeep(entry);
				tempEntry.last_opened = new Date().getTime() / 1000;
				console.log(tempEntry.last_opened);
				rssStore.addRecentlyEntry(tempEntry);
				if (readerStore.readingEntry.file_type === FILE_TYPE.ARTICLE) {
					nextTick(() => {
						isScrollAble.value = false;
						setTimeout(() => {
							if (!isScrollAble.value) {
								readingProgressStore.updateProgress(1);
							}
						}, 2000);
					});
				}
			});
		} else {
			console.log('entry empty');
			readerStore.clearNavigationList();
		}
	},
	{
		immediate: true
	}
);

function checkAndScrollToPosition() {
	if (
		readerStore.readingEntry &&
		readerStore.readingEntry.file_type === FILE_TYPE.ARTICLE &&
		readerStore.readingEntry.progress
	) {
		scrollAreaRef.value.setScrollPercentage(
			'vertical',
			Number(readerStore.readingEntry.progress) / 100
		);
	}
}

const articleContentRef = ref(null);

onMounted(() => {
	HotkeyManager.setScope('reading');
	HotkeyManager.registerHotkeys(
		{
			[WISE_HOTKEY.ENTRY.BACK]: () => {
				onBackClick();
			}
		},
		['reading']
	);
	HotkeyManager.logAllKeyCodes();
	articleContentRef.value = document.getElementById('article-parent');

	readingProgressStore.startReading();

	nextTick(() => {
		const dummy = document.body.offsetHeight;
		setTimeout(() => {
			checkAndScrollToPosition();
			observerVisibleTitle();
			markVisibleElement();
		}, 1000);
	});
});

onBeforeUnmount(() => {
	readingProgressStore.stopReading();

	if (intersectionTrackingObserver) {
		intersectionTrackingObserver.disconnect();
	}

	if (intersectionTitleObserver) {
		intersectionTitleObserver.disconnect();
	}

	// clickedElements.forEach((el) => {
	// 	el.removeEventListener('click', handleElementClick);
	// });
	// clickedElements.length = 0;
});

const onBackClick = () => {
	if (route.params.path) {
		if (route.params.path === MenuType.Entry) {
			configStore.setMenuType(MenuType.History);
		} else if (SupportDetails.includes(route.params.path)) {
			configStore.setMenuType(route.params.path);
		} else {
			configStore.setMenuType(route.params.path, {
				filterId: route.params.path
			});
		}
	} else {
		router.back();
	}
};

const onScroll = (info: any) => {
	showBarShadow.value = info.verticalPosition > 0;

	if (
		readerStore.readingEntry &&
		readerStore.readingEntry.file_type === FILE_TYPE.ARTICLE
	) {
		if (info.verticalSize > info.verticalContainerSize) {
			readingProgressStore.updateProgress(info.verticalPercentage);
			isScrollAble.value = true;
		}
		readingProgressStore.setTotalProgress(
			1,
			extractHtml(readerStore.readingEntry.full_content!.trim()).length
		);
	}

	if (info.verticalPosition > 0 && isPDF.value && pdfController.value) {
		pdfController.value.changePageByScroll(info.verticalPosition);
	}
};

let selectedElement = null;
let clickedElement = null;

function markVisibleElement() {
	if (intersectionTrackingObserver) {
		intersectionTrackingObserver.disconnect();
	}
	intersectionTrackingObserver = new IntersectionObserver(
		(entries) => {
			const visibleEntries = entries
				.filter((entry) => entry.isIntersecting)
				.sort((a, b) => a.boundingClientRect.top - b.boundingClientRect.top);

			if (visibleEntries.length > 0) {
				selectedElement = visibleEntries[0].target;
				if (clickedElement) {
					clickedElement = null;
				}
				updateIndicator(selectedElement);
			}
		},
		{
			threshold: 0.1
		}
	);

	// const rootElement = pdfController.value.$el;
	//
	// const targetTags = ['p', 'article', 'section', 'ul', 'ol'];
	//
	// const targetLayer = findTargetLayer(rootElement, targetTags);

	if (pdfController.value && pdfController.value.$el) {
		const targetLayer = pdfController.value.$el.querySelectorAll(
			'section, p, h1, h2, h3, h4, h5, h6, img, ul, ol, blockquote, pre,table'
		);

		console.log(targetLayer);

		targetLayer.forEach((element) => {
			intersectionTrackingObserver.observe(element);
			// element.addEventListener('click', handleElementClick);
			// clickedElements.push(element);
		});
	}
}

const handleElementClick = (e: MouseEvent) => {
	clickedElement = e.target as HTMLElement;
	updateIndicator(clickedElement);
};

function findTargetLayer(element, targetTags) {
	const children = Array.from(element.children);

	const filteredChildren = children.filter((child) => {
		const tagName = child.tagName.toLowerCase();
		return targetTags.includes(tagName);
	});

	const invalidTags = ['hr', 'br'];

	const articles = filteredChildren.filter(
		(child) => child.tagName.toLowerCase() === 'article'
	);
	if (articles.length > 0) {
		const childrenArray = articles
			.map((article) => Array.from(cleanArticle(article).children))
			.flat();

		console.log(childrenArray);

		if (childrenArray.length > 0) {
			const validChildren = childrenArray.filter(
				(child) => !invalidTags.includes(child.tagName.toLowerCase())
			);
			if (validChildren.length > 0) {
				return validChildren;
			}
		}
	}

	const validChildren = filteredChildren.filter(
		(child) => !invalidTags.includes(child.tagName.toLowerCase())
	);
	if (validChildren.length > 0) {
		return validChildren;
	}

	for (const child of children) {
		const result = findTargetLayer(child, targetTags);
		if (result) {
			return result;
		}
	}

	return null;
}

function cleanArticle(article) {
	const children = Array.from(article.children);

	const validChildren = children.filter((child) => {
		const tagName = child.tagName.toLowerCase();
		return tagName !== 'header' && tagName !== 'details';
	});

	validChildren.forEach((child) => cleanArticle(child));

	return article;
}

let debounceTimer = null;
const updateIndicator = (element: HTMLElement) => {
	if (debounceTimer) {
		clearTimeout(debounceTimer);
	}

	debounceTimer = setTimeout(() => {
		if (element.offsetTop === 0 && element.offsetHeight === 0) {
			return;
		}
		const additionalOffset = element.offsetTop === 0 ? 0 : 101 + 20;

		indicatorStyle.value = {
			transform: `translateY(${element.offsetTop + additionalOffset}px)`,
			height: `${element.offsetHeight}px`
		};
	}, 100);
};

function observerVisibleTitle() {
	if (intersectionTitleObserver) {
		intersectionTitleObserver.disconnect();
	}

	intersectionTitleObserver = new IntersectionObserver(
		(entries) => {
			entries.forEach((entry) => {
				if (entry.isIntersecting) {
					console.log('topic target ', entry);
					console.log('topic array', readerStore.topicArray);
					readerStore.updateReadingTopic(entry.target.id);
					console.log('topic target result ', readerStore.readingTopic);
				}
			});
		},
		{
			rootMargin: '0px 0px -90% 0px',
			threshold: 0
		}
	);

	const elements = document.querySelectorAll('h1, h2, h3, h4, h5, h6');
	elements.forEach((title) => intersectionTitleObserver.observe(title));
}

watch(
	() => readerStore.readingTopic,
	(newTopic) => {
		if (newTopic && newTopic.jump && articleContentRef.value) {
			const targetElement = articleContentRef.value.querySelector(
				`#${newTopic.id}`
			);
			if (targetElement) {
				targetElement.scrollIntoView({ block: 'start' });
			}
		}
	}
);

const isPDF = computed(() => {
	return (
		readerStore.readingEntry &&
		readerStore.readingEntry.file_type === FILE_TYPE.PDF
	);
});

const isEntryReadable = computed(() => {
	if (!readerStore.readingEntry) {
		return false;
	}
	return (
		readerStore.readingEntry.status === ENTRY_STATUS.Extracted ||
		readerStore.readingEntry.status === ENTRY_STATUS.Staging ||
		readerStore.readingEntry.status === ENTRY_STATUS.Completed
	);
});

/***
 * second priority
 */
const props = defineProps({
	margin: {
		type: Number,
		default: 290
	},
	lineHeight: {
		type: Number,
		default: 150
	},
	maxWidthPercentage: {
		type: Number,
		default: 0
	},
	fontFamily: {
		type: String,
		default: 'Robot'
	},
	highContrastText: {
		type: Boolean,
		default: false
	},
	justifyText: Boolean,
	fontSize: {
		type: Number,
		default: 20
	}
});

/***
 * first priority
 */
const maxWidthPercentageOverride = null;
const lineHeightOverride = null;
const fontFamilyOverride = null;
const highContrastTextOverride = undefined;
const justifyTextOverride = undefined;
const justifyTextValue = (isJustified: boolean) => {
	return isJustified ? 'justify' : 'start';
};
const { color: background1 } = useColor('background-1');
const { color: ink2 } = useColor('ink-2');
const $q = useQuasar();

const textColorValue = (isHighContrast: boolean) => {
	return isHighContrast ? theme.readerFontHighContrast : theme.readerFont;
};

const theme = {
	readerTableHeader: '#FFFFFF',
	readerFontHighContrast: '#0A0806',
	readerFont: ink2.value,
	readerBg: background1.value
};

const styles = {
	fontSize: props.fontSize,
	margin: props.margin,
	maxWidthPercentage: maxWidthPercentageOverride ?? props.maxWidthPercentage,
	lineHeight: lineHeightOverride ?? props.lineHeight,
	fontFamily: fontFamilyOverride ?? props.fontFamily,
	readerFontColor: textColorValue(
		highContrastTextOverride ?? props.highContrastText
	),
	readerTableHeaderColor: theme.readerTableHeader,
	readerHeadersColor: theme.readerFont
};

const maxWidthStyles = {
	default: styles.maxWidthPercentage
		? `${styles.maxWidthPercentage}%`
		: `${1024 - styles.margin}px`,
	small: styles.maxWidthPercentage
		? `${styles.maxWidthPercentage}%`
		: `calc(${120 - Math.round((styles.margin * 10) / 100) / 100}% * 100vw)`
};

const currentReader = computed(() => {
	switch (readerStore.readingEntry.file_type) {
		// case FILE_TYPE.VIDEO:
		// 	return defineAsyncComponent(() => import('./preview/RssVideoPlayer.vue'));
		// case FILE_TYPE.AUDIO:
		// 	return defineAsyncComponent(() => import('./preview/RssAudioPlayer.vue'));
		case FILE_TYPE.EBOOK:
			return defineAsyncComponent(() => import('./preview/RssEbookReader.vue'));
		case FILE_TYPE.PDF:
			return defineAsyncComponent(() => import('./preview/RssPdfReader.vue'));
		default:
			return defineAsyncComponent(
				() => import('./preview/RssArticleReader.vue')
			);
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

					//.reading-indicator {
					//	position: absolute;
					//	top: 0;
					//	left: -20px;
					//	width: 2px;
					//	background: $orange-default;
					//	fill: #fff;
					//	transition: transform 0.2s, height 0.2s;
					//	z-index: 10;
					//	pointer-events: none;
					//}
				}
			}
		}
	}
}
</style>
