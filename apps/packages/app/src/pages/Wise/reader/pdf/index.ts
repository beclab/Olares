/**
 * PDF Module - Unified Export
 *
 * This module provides all PDF-related components, composables, and services
 */

// ==================== Components ====================
export { default as PdfReader } from './components/PdfReader.vue';
export { default as PdfTitleBar } from './components/PdfTitleBar.vue';
export { default as PdfSidebar } from './components/PdfSidebar.vue';
export { default as PdfThumbnailItem } from './components/PdfThumbnailItem.vue';

// ==================== Composables ====================
export {
	usePdfLazyLoad,
	PDF_LAZY_LOAD_PRESETS,
	PdfLazyLoadOptions
} from './composables/usePdfLazyLoad';

// ==================== Services ====================
export {
	pdfCache,
	fetchPDFWithProgress,
	formatBytes,
	formatSpeed,
	PDFStorageMode,
	PDFLoadResult,
	DownloadProgress,
	ProgressiveLoadOptions
} from './services/pdfCache';

export {
	getThumbnail,
	generateThumbnail,
	clearThumbnailCache,
	usePdfThumbnails,
	ThumbnailOptions
} from './services/pdfThumbnail';

export { configurePdfWorker } from './services/pdfWorkerConfig';

// ==================== Store ====================
// Note: Store is typically imported directly from stores path
// export { usePDfStore } from './stores/pdf';
