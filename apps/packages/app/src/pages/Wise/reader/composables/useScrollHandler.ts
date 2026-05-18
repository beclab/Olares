/**
 * Scroll Handler Composable
 *
 * Manages scroll events for article and PDF reading
 * Provides unified scroll handling with progress tracking
 */

import { ref, computed, Ref } from 'vue';
import { FILE_TYPE } from 'src/utils/rss-types';

export interface ScrollInfo {
	/** Vertical scroll position */
	verticalPosition: number;
	/** Vertical container size */
	verticalContainerSize: number;
	/** Vertical content size */
	verticalSize: number;
	/** Vertical scroll percentage (0-1) */
	verticalPercentage: number;
}

export interface UseScrollHandlerOptions {
	/** File type of current content */
	fileType: Ref<string | undefined>;
	/** Callback for article scroll progress */
	onArticleProgress?: (percentage: number) => void;
	/** Callback for PDF page change */
	onPdfPageChange?: (scrollY: number) => void;
	/** Callback for shadow visibility */
	onShadowChange?: (visible: boolean) => void;
}

export function useScrollHandler(options: UseScrollHandlerOptions) {
	const { fileType, onArticleProgress, onPdfPageChange, onShadowChange } =
		options;

	// Shadow visibility state
	const showBarShadow = ref(false);

	// Track if content is scrollable
	const isScrollable = ref(false);

	// Check if current content is PDF
	const isPdf = computed(() => fileType.value === FILE_TYPE.PDF);

	// Check if current content is Article
	const isArticle = computed(() => fileType.value === FILE_TYPE.ARTICLE);

	/**
	 * Handle scroll event from QScrollArea
	 */
	function handleScroll(info: ScrollInfo) {
		// Update shadow visibility
		const shadowVisible = info.verticalPosition > 0;
		showBarShadow.value = shadowVisible;
		onShadowChange?.(shadowVisible);

		// Handle Article scroll
		if (isArticle.value) {
			handleArticleScroll(info);
		}

		// Handle PDF scroll
		if (isPdf.value && info.verticalPosition > 0) {
			onPdfPageChange?.(info.verticalPosition);
		}
	}

	/**
	 * Handle article-specific scroll logic
	 */
	function handleArticleScroll(info: ScrollInfo) {
		// Check if content is scrollable
		if (info.verticalSize > info.verticalContainerSize) {
			isScrollable.value = true;
			onArticleProgress?.(info.verticalPercentage);
		}
	}

	/**
	 * Scroll to specific percentage
	 */
	function scrollToPercentage(scrollAreaRef: Ref<any>, percentage: number) {
		if (scrollAreaRef.value) {
			scrollAreaRef.value.setScrollPercentage('vertical', percentage / 100);
		}
	}

	/**
	 * Scroll to element by ID
	 */
	function scrollToElement(
		elementId: string,
		behavior: ScrollBehavior = 'auto'
	) {
		const element = document.getElementById(elementId);
		if (element) {
			element.scrollIntoView({ behavior, block: 'start' });
		}
	}

	/**
	 * Check and restore scroll position from saved progress
	 */
	function restoreScrollPosition(
		scrollAreaRef: Ref<any>,
		progress: number | string | undefined
	) {
		if (progress && scrollAreaRef.value) {
			const percentage = Number(progress);
			if (!isNaN(percentage) && percentage > 0) {
				scrollAreaRef.value.setScrollPercentage('vertical', percentage / 100);
			}
		}
	}

	return {
		/** Shadow visibility state */
		showBarShadow,
		/** Whether content is scrollable */
		isScrollable,
		/** Whether current content is PDF */
		isPdf,
		/** Whether current content is Article */
		isArticle,
		/** Handle scroll event */
		handleScroll,
		/** Scroll to percentage */
		scrollToPercentage,
		/** Scroll to element by ID */
		scrollToElement,
		/** Restore scroll position */
		restoreScrollPosition
	};
}
