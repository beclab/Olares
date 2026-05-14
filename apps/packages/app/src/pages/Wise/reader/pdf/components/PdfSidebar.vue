<template>
	<div class="pdf-sidebar">
		<!-- Page Indicator -->
		<div class="page-indicator">
			<span class="current-page">{{ pdfConfigStore.pageNum }}</span>
			<span class="separator">/</span>
			<span class="total-pages">{{ pdfConfigStore.numPages }}</span>
		</div>

		<!-- Page Slider -->
		<div class="page-slider">
			<input
				type="range"
				:min="1"
				:max="pdfConfigStore.numPages"
				:value="pdfConfigStore.pageNum"
				@input="handleSliderChange"
				class="slider"
			/>
		</div>

		<!-- View Tabs -->
		<div class="view-tabs">
			<div
				class="pdf-tab"
				:class="{ active: activeView === 'thumbnails' }"
				@click="activeView = 'thumbnails'"
			>
				{{ t('Thumbnail') }}
			</div>
			<div
				class="pdf-tab"
				:class="{
					active: activeView === 'outline',
					disabled: !hasPdfOutline
				}"
				@click="hasPdfOutline && (activeView = 'outline')"
			>
				{{ t('TOC') }}
			</div>
		</div>

		<!-- Thumbnails View -->
		<bt-scroll-area
			class="full-width thumbnail-scroll"
			v-if="activeView === 'thumbnails'"
		>
			<div class="thumbnail-grid">
				<PdfThumbnailItem
					v-for="i in pageNumbers"
					:key="i"
					:id="'thumb_' + i"
					:pdf-document="pdfDocument"
					:page="i"
					:source-key="sourceKey"
					:active="i === pdfConfigStore.pageNum"
					:width="120"
					@click="handleThumbnailClick(i)"
				/>
			</div>
		</bt-scroll-area>

		<!-- Outline View -->
		<bt-scroll-area
			class="full-width outline-scroll"
			v-if="activeView === 'outline' && hasPdfOutline"
		>
			<div class="pdf-outline">
				<template v-for="(item, index) in flattenedOutline" :key="index">
					<div
						class="outline-item cursor-pointer"
						:class="{
							active: isOutlineItemActive(item),
							['level-' + item.level]: true
						}"
						@click="handleOutlineClick(item)"
					>
						<span class="outline-title">{{ item.title }}</span>
						<span class="outline-page">{{ item.page }}</span>
					</div>
				</template>
			</div>
		</bt-scroll-area>
	</div>
</template>

<script setup lang="ts">
import PdfThumbnailItem from './PdfThumbnailItem.vue';
import { computed, nextTick, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import {
	usePdfLazyLoad,
	PDF_LAZY_LOAD_PRESETS
} from 'src/pages/Wise/reader/pdf';
import { useReadingProgressStore } from 'src/stores/rss-reading-progress';
const rssReadingStore = useReadingProgressStore();

const { t } = useI18n();

// Use lazy load logic for page numbers
const { pageNumbers, pdfConfigStore } = usePdfLazyLoad(
	PDF_LAZY_LOAD_PRESETS.thumbnail
);

// PDF document reference
const pdfDocument = computed(() => pdfConfigStore.pdfDocument);

// Source key for thumbnail caching
const sourceKey = computed(() => {
	const source = pdfConfigStore.source;
	if (!source) return '';
	if (typeof source === 'string') return source;
	if (source.url) return source.url;
	return `pdf-${Date.now()}`;
});

// Current active view
const activeView = ref<'thumbnails' | 'outline'>('thumbnails');

// Check if PDF has outline
const hasPdfOutline = computed(() => {
	return pdfConfigStore.pdfOutline && pdfConfigStore.pdfOutline.length > 0;
});

// Flatten outline for display
const flattenedOutline = computed(() => {
	if (!pdfConfigStore.pdfOutline) return [];

	const result: any[] = [];

	function flatten(items: any[], level = 0) {
		for (const item of items) {
			result.push({ ...item, level });
			if (item.children && item.children.length > 0) {
				flatten(item.children, level + 1);
			}
		}
	}

	flatten(pdfConfigStore.pdfOutline);
	return result;
});

// Check if outline item is active
function isOutlineItemActive(item: any): boolean {
	const currentPage = pdfConfigStore.pageNum;
	const itemIndex = flattenedOutline.value.indexOf(item);
	const nextItem = flattenedOutline.value[itemIndex + 1];

	if (!nextItem) {
		return currentPage >= item.page;
	}

	return currentPage >= item.page && currentPage < nextItem.page;
}

// Navigation handlers
function goToPage(page: number) {
	console.log('PdfSidebar: navigate to page', page);
	pdfConfigStore.pageNum = page;
	nextTick(() => {
		pdfConfigStore.handleItemClick(page);
		rssReadingStore.updateProgress(page);
	});
}

function handleOutlineClick(item: any) {
	goToPage(item.page);
}

function handleThumbnailClick(page: number) {
	goToPage(page);
	scrollToPageCell(page);
}

function handleSliderChange(event: Event) {
	const target = event.target as HTMLInputElement;
	const page = parseInt(target.value, 10);
	goToPage(page);
	scrollToPageCell(page);
}

// Scroll to page element
function scrollToPageCell(page: number) {
	nextTick(() => {
		let el: HTMLElement | null = null;

		if (activeView.value === 'thumbnails') {
			el = document.getElementById('thumb_' + page);
		}

		if (el) {
			el.scrollIntoView({ behavior: 'smooth', block: 'center' });
		}
	});
}

// Watch page number changes
watch(
	() => pdfConfigStore.pageNum,
	(newPage) => {
		if (newPage) {
			scrollToPageCell(newPage);
		}
	}
);

// Scroll to current page when switching views
watch(activeView, () => {
	if (pdfConfigStore.pageNum) {
		scrollToPageCell(pdfConfigStore.pageNum);
	}
});
</script>

<style scoped lang="scss">
.pdf-sidebar {
	padding: 12px;
	display: flex;
	flex-direction: column;
	gap: 12px;
	height: 100%;

	// Page indicator
	.page-indicator {
		display: flex;
		align-items: baseline;
		justify-content: center;
		gap: 4px;

		.current-page {
			font-size: 28px;
			font-weight: 600;
			color: $orange-default;
		}

		.separator {
			font-size: 18px;
			color: #999;
		}

		.total-pages {
			font-size: 16px;
			color: #666;
		}
	}

	// Page slider
	.page-slider {
		padding: 0 8px;

		.slider {
			width: 100%;
			height: 6px;
			-webkit-appearance: none;
			appearance: none;
			background: #e0e0e0;
			border-radius: 3px;
			outline: none;
			cursor: pointer;

			&::-webkit-slider-thumb {
				-webkit-appearance: none;
				appearance: none;
				width: 18px;
				height: 18px;
				background: $orange-default;
				border-radius: 50%;
				cursor: pointer;
				box-shadow: 0 2px 6px rgba(0, 0, 0, 0.2);
				transition: transform 0.15s ease;

				&:hover {
					transform: scale(1.15);
				}
			}

			&::-moz-range-thumb {
				width: 18px;
				height: 18px;
				background: $orange-default;
				border-radius: 50%;
				cursor: pointer;
				border: none;
			}
		}
	}

	// View tabs
	.view-tabs {
		display: flex;
		background: #f0f0f0;
		border-radius: 8px;
		padding: 3px;

		.pdf-tab {
			flex: 1;
			text-align: center;
			padding: 6px 8px;
			font-size: 12px;
			color: #666;
			border-radius: 6px;
			cursor: pointer;
			transition: all 0.2s ease;

			&:hover:not(.disabled) {
				color: #333;
			}

			&.active {
				background: white;
				color: $orange-default;
				font-weight: 500;
				box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
			}

			&.disabled {
				color: #bbb;
				cursor: not-allowed;
			}
		}
	}

	// Thumbnail scroll
	.thumbnail-scroll {
		flex: 1;
		min-height: 0;
	}

	// Thumbnail grid
	.thumbnail-grid {
		display: flex;
		flex-direction: column;
		align-items: center;
		padding: 8px;
		gap: 0;
	}

	// Outline scroll
	.outline-scroll {
		flex: 1;
		min-height: 0;
	}

	// Outline styles
	.pdf-outline {
		display: flex;
		flex-direction: column;
		gap: 2px;
		padding: 4px 0;

		.outline-item {
			display: flex;
			align-items: center;
			justify-content: space-between;
			padding: 8px 10px;
			border-radius: 6px;
			transition: all 0.15s ease;
			gap: 8px;

			&:hover {
				background: #f5f5f5;
			}

			&.active {
				background: rgba($orange-default, 0.1);

				.outline-title {
					color: $orange-default;
					font-weight: 500;
				}

				.outline-page {
					color: $orange-default;
				}
			}

			.outline-title {
				flex: 1;
				font-size: 13px;
				color: #333;
				overflow: hidden;
				text-overflow: ellipsis;
				white-space: nowrap;
			}

			.outline-page {
				font-size: 11px;
				color: #999;
				flex-shrink: 0;
			}

			// Indentation levels
			&.level-0 {
				padding-left: 10px;
			}

			&.level-1 {
				padding-left: 22px;
				font-size: 12px;
			}

			&.level-2 {
				padding-left: 34px;
				font-size: 12px;
			}

			&.level-3 {
				padding-left: 46px;
				font-size: 11px;
			}

			&.level-4 {
				padding-left: 58px;
				font-size: 11px;
			}

			&.level-5 {
				padding-left: 70px;
				font-size: 11px;
			}
		}
	}
}
</style>
