/**
 * PDF Lazy Load Composable
 * Manages PDF page lazy loading logic to prevent memory overflow from rendering all pages at once
 *
 * Uses viewport strategy - lazy loads pages based on current page position
 */

import { ref, watch, computed, nextTick } from 'vue';
import { usePDfStore } from 'src/stores/pdf';

// ==================== Type Definitions ====================

export interface PdfLazyLoadOptions {
	/** Preload buffer size (pages before and after current page) */
	buffer?: number;
	/** Whether to auto scroll to current page */
	autoScrollToCurrentPage?: boolean;
}

// ==================== Viewport Strategy Composable ====================

export function usePdfLazyLoad(options: PdfLazyLoadOptions = {}) {
	const pdfConfigStore = usePDfStore();

	const { buffer = 3, autoScrollToCurrentPage = true } = options;

	// Visible pages list
	const visiblePages = ref<number[]>([]);

	// Current page number
	const currentPage = computed(() => pdfConfigStore.pageNum ?? 1);
	// Total pages
	const totalPages = computed(() => pdfConfigStore.numPages ?? 0);

	/**
	 * Generate page number array (for v-for)
	 */
	const pageNumbers = computed(() => {
		const total = totalPages.value;
		if (total <= 0) return [];
		return Array.from({ length: total }, (_, i) => i + 1);
	});

	/**
	 * Update visible page range
	 */
	function updateVisiblePages() {
		const current = currentPage.value || 1;
		const total = totalPages.value;

		if (total <= 0) {
			visiblePages.value = [];
			return;
		}

		const start = Math.max(1, current - buffer);
		const end = Math.min(total, current + buffer);

		const pages: number[] = [];
		for (let i = start; i <= end; i++) {
			pages.push(i);
		}
		visiblePages.value = pages;
	}

	/**
	 * Check if page should be rendered
	 */
	function shouldRenderPage(pageNum: number): boolean {
		return visiblePages.value.includes(pageNum);
	}

	/**
	 * Scroll to element corresponding to current page
	 */
	function scrollToCurrentPage(prefix = 'thumb_') {
		const page = currentPage.value;
		nextTick(() => {
			const el = document.getElementById(prefix + page);
			if (el) {
				el.scrollIntoView({ behavior: 'smooth', block: 'center' });
			}
		});
	}

	// ========== Watchers ==========

	// Watch page number changes to update visible range
	watch(currentPage, updateVisiblePages);
	watch(totalPages, updateVisiblePages, { immediate: true });

	// Auto scroll to current page when it changes
	if (autoScrollToCurrentPage) {
		watch(currentPage, () => {
			scrollToCurrentPage();
		});
	}

	return {
		/** Visible pages list */
		visiblePages,
		/** All page numbers array */
		pageNumbers,
		/** Current page number */
		currentPage,
		/** Total pages */
		totalPages,
		/** Update visible pages */
		updateVisiblePages,
		/** Check if page should be rendered */
		shouldRenderPage,
		/** Scroll to current page */
		scrollToCurrentPage,
		/** pdfConfigStore reference */
		pdfConfigStore
	};
}

// ==================== Preset Configurations ====================

export const PDF_LAZY_LOAD_PRESETS = {
	/** Main reader: 2 pages before and after */
	reader: {
		buffer: 2
	},
	/** Thumbnail panel: 3 thumbnails before and after (used for scroll sync only, visibility handled by IntersectionObserver) */
	thumbnail: {
		buffer: 3,
		autoScrollToCurrentPage: true
	}
} as const;
