/**
 * PDF Thumbnail Generator
 * Generates lightweight thumbnail images from PDF pages
 * Uses canvas rendering and converts to data URL to minimize memory usage
 */

import * as pdfjsLib from 'pdfjs-dist';

// Thumbnail cache (page number -> data URL)
const thumbnailCache = new Map<string, Map<number, string>>();

// Pending render tasks (to prevent duplicate renders)
const pendingRenders = new Map<string, Map<number, Promise<string>>>();

/**
 * Generate cache key for a PDF source
 */
function getSourceKey(source: any): string {
	if (typeof source === 'string') {
		return source;
	}
	if (source?.url) {
		return source.url;
	}
	if (source?.data) {
		// For ArrayBuffer/Uint8Array, use a hash or just timestamp
		return `data-${Date.now()}`;
	}
	return 'unknown';
}

/**
 * Get or create cache for a PDF source
 */
function getCache(sourceKey: string): Map<number, string> {
	if (!thumbnailCache.has(sourceKey)) {
		thumbnailCache.set(sourceKey, new Map());
	}
	return thumbnailCache.get(sourceKey)!;
}

/**
 * Clear thumbnail cache for a specific source or all
 */
export function clearThumbnailCache(source?: any): void {
	if (source) {
		const key = getSourceKey(source);
		thumbnailCache.delete(key);
		pendingRenders.delete(key);
	} else {
		thumbnailCache.clear();
		pendingRenders.clear();
	}
}

/**
 * Options for thumbnail generation
 */
export interface ThumbnailOptions {
	/** Thumbnail width in pixels (default: 150) */
	width?: number;
	/** Image quality for JPEG (0-1, default: 0.7) */
	quality?: number;
	/** Image format (default: 'image/jpeg') */
	format?: 'image/jpeg' | 'image/png' | 'image/webp';
}

/**
 * Generate thumbnail for a specific page
 * @param pdfDocument PDF.js document object
 * @param pageNumber Page number (1-based)
 * @param options Thumbnail options
 * @returns Data URL of the thumbnail image
 */
export async function generateThumbnail(
	pdfDocument: pdfjsLib.PDFDocumentProxy,
	pageNumber: number,
	options: ThumbnailOptions = {}
): Promise<string> {
	const { width = 150, quality = 0.7, format = 'image/jpeg' } = options;

	// Get page
	const page = await pdfDocument.getPage(pageNumber);

	// Calculate scale based on desired width
	const viewport = page.getViewport({ scale: 1 });
	const scale = width / viewport.width;
	const scaledViewport = page.getViewport({ scale });

	// Create canvas
	const canvas = document.createElement('canvas');
	canvas.width = scaledViewport.width;
	canvas.height = scaledViewport.height;

	const context = canvas.getContext('2d');
	if (!context) {
		throw new Error('Failed to get canvas context');
	}

	// Render page to canvas
	const renderContext = {
		canvasContext: context,
		viewport: scaledViewport
	};

	await page.render(renderContext).promise;

	// Convert to data URL
	const dataUrl = canvas.toDataURL(format, quality);

	// Cleanup - release resources
	canvas.width = 0;
	canvas.height = 0;
	page.cleanup();

	return dataUrl;
}

/**
 * Generate thumbnail with caching
 * @param pdfDocument PDF.js document object
 * @param pageNumber Page number (1-based)
 * @param sourceKey Cache key for the PDF source
 * @param options Thumbnail options
 * @returns Data URL of the thumbnail image
 */
export async function getThumbnail(
	pdfDocument: pdfjsLib.PDFDocumentProxy,
	pageNumber: number,
	sourceKey: string,
	options: ThumbnailOptions = {}
): Promise<string> {
	// Check cache first
	const cache = getCache(sourceKey);
	if (cache.has(pageNumber)) {
		return cache.get(pageNumber)!;
	}

	// Check if already rendering
	let pendingMap = pendingRenders.get(sourceKey);
	if (!pendingMap) {
		pendingMap = new Map();
		pendingRenders.set(sourceKey, pendingMap);
	}

	if (pendingMap.has(pageNumber)) {
		return pendingMap.get(pageNumber)!;
	}

	// Generate thumbnail
	const renderPromise = generateThumbnail(pdfDocument, pageNumber, options)
		.then((dataUrl) => {
			// Cache the result
			cache.set(pageNumber, dataUrl);
			// Remove from pending
			pendingMap?.delete(pageNumber);
			return dataUrl;
		})
		.catch((error) => {
			pendingMap?.delete(pageNumber);
			throw error;
		});

	pendingMap.set(pageNumber, renderPromise);
	return renderPromise;
}

/**
 * Batch generate thumbnails for multiple pages
 * @param pdfDocument PDF.js document object
 * @param pageNumbers Array of page numbers
 * @param sourceKey Cache key for the PDF source
 * @param options Thumbnail options
 * @param onProgress Progress callback
 */
export async function generateThumbnailBatch(
	pdfDocument: pdfjsLib.PDFDocumentProxy,
	pageNumbers: number[],
	sourceKey: string,
	options: ThumbnailOptions = {},
	onProgress?: (completed: number, total: number) => void
): Promise<Map<number, string>> {
	const results = new Map<number, string>();
	let completed = 0;

	// Generate sequentially to avoid memory spikes
	for (const pageNum of pageNumbers) {
		try {
			const dataUrl = await getThumbnail(
				pdfDocument,
				pageNum,
				sourceKey,
				options
			);
			results.set(pageNum, dataUrl);
		} catch (error) {
			console.warn(`Failed to generate thumbnail for page ${pageNum}:`, error);
		}

		completed++;
		onProgress?.(completed, pageNumbers.length);
	}

	return results;
}

/**
 * Vue composable for PDF thumbnails
 */
export function usePdfThumbnails() {
	const thumbnails = new Map<number, string>();
	let currentSourceKey = '';

	/**
	 * Load thumbnail for a page
	 */
	async function loadThumbnail(
		pdfDocument: pdfjsLib.PDFDocumentProxy,
		pageNumber: number,
		sourceKey: string,
		options?: ThumbnailOptions
	): Promise<string | null> {
		try {
			currentSourceKey = sourceKey;
			const dataUrl = await getThumbnail(
				pdfDocument,
				pageNumber,
				sourceKey,
				options
			);
			thumbnails.set(pageNumber, dataUrl);
			return dataUrl;
		} catch (error) {
			console.warn(`Failed to load thumbnail for page ${pageNumber}:`, error);
			return null;
		}
	}

	/**
	 * Get cached thumbnail
	 */
	function getCachedThumbnail(pageNumber: number): string | undefined {
		return thumbnails.get(pageNumber);
	}

	/**
	 * Clear thumbnails
	 */
	function clear(): void {
		thumbnails.clear();
		if (currentSourceKey) {
			clearThumbnailCache({ url: currentSourceKey });
		}
	}

	return {
		thumbnails,
		loadThumbnail,
		getCachedThumbnail,
		clear
	};
}
