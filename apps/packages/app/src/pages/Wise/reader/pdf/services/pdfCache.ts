/**
 * PDF Cache Manager
 * Uses IndexedDB to store PDF file data for offline access
 *
 * Supports two storage modes:
 * - 'arraybuffer': Store as ArrayBuffer (default, compatible with older data)
 * - 'blob': Store as Blob (better for large files, lower memory usage)
 */

const DB_NAME = 'pdf-cache-db';
const DB_VERSION = 2; // Bump version to support storage type field
const STORE_NAME = 'pdf-files';
const DEFAULT_MAX_CACHE_SIZE = 1000 * 1024 * 1024; // Default max cache 1GB
const MAX_CACHE_AGE = 14 * 24 * 60 * 60 * 1000; // Cache validity 14 days
const QUOTA_USAGE_RATIO = 0.5; // Use at most 50% of browser quota

/**
 * Storage mode for PDF cache
 * - 'arraybuffer': Traditional mode, stores raw binary data
 * - 'blob': Blob mode, better for large files (browser can optimize storage)
 */
export type PDFStorageMode = 'arraybuffer' | 'blob';

interface CachedPDF {
	url: string;
	data: ArrayBuffer | Blob;
	timestamp: number;
	size: number;
	storageType?: PDFStorageMode; // Track how data was stored
	name?: string; // Human-readable file name for identification
}

class PDFCacheManager {
	private db: IDBDatabase | null = null;
	private dbPromise: Promise<IDBDatabase> | null = null;
	private storageMode: PDFStorageMode = 'blob'; // Default to blob for better performance

	/**
	 * Set storage mode
	 * @param mode 'arraybuffer' or 'blob'
	 */
	setStorageMode(mode: PDFStorageMode): void {
		this.storageMode = mode;
		console.log(`PDF Cache: Storage mode set to '${mode}'`);
	}

	/**
	 * Get current storage mode
	 */
	getStorageMode(): PDFStorageMode {
		return this.storageMode;
	}

	/**
	 * Initialize IndexedDB database
	 */
	private async initDB(): Promise<IDBDatabase> {
		if (this.db) {
			return this.db;
		}

		if (this.dbPromise) {
			return this.dbPromise;
		}

		this.dbPromise = new Promise((resolve, reject) => {
			const request = indexedDB.open(DB_NAME, DB_VERSION);

			request.onerror = () => {
				console.error('PDF Cache: Failed to open database', request.error);
				reject(request.error);
			};

			request.onsuccess = () => {
				this.db = request.result;
				resolve(this.db);
			};

			request.onupgradeneeded = (event) => {
				const db = (event.target as IDBOpenDBRequest).result;
				if (!db.objectStoreNames.contains(STORE_NAME)) {
					const store = db.createObjectStore(STORE_NAME, { keyPath: 'url' });
					store.createIndex('timestamp', 'timestamp', { unique: false });
				}
			};
		});

		return this.dbPromise;
	}

	/**
	 * Generate cache key (normalize URL)
	 */
	getCacheKey(url: string): string {
		try {
			const urlObj = new URL(url);
			// Remove query params that may change, keep core path
			return `${urlObj.origin}${urlObj.pathname}`;
		} catch {
			return url;
		}
	}

	/**
	 * Check if cache exists (without reading data, faster)
	 */
	async has(url: string): Promise<boolean> {
		try {
			const db = await this.initDB();
			const cacheKey = this.getCacheKey(url);

			return new Promise((resolve) => {
				const transaction = db.transaction(STORE_NAME, 'readonly');
				const store = transaction.objectStore(STORE_NAME);
				const request = store.get(cacheKey);

				request.onsuccess = () => {
					const cached = request.result as CachedPDF | undefined;
					if (cached && Date.now() - cached.timestamp < MAX_CACHE_AGE) {
						resolve(true);
					} else {
						resolve(false);
					}
				};

				request.onerror = () => {
					resolve(false);
				};
			});
		} catch (error) {
			return false;
		}
	}

	/**
	 * Get PDF from cache
	 * Automatically handles both ArrayBuffer and Blob storage formats
	 */
	async get(url: string): Promise<ArrayBuffer | null> {
		const startTime = performance.now();
		try {
			const db = await this.initDB();
			const cacheKey = this.getCacheKey(url);
			const dbReadyTime = performance.now();
			console.log(
				`PDF Cache: DB init took ${(dbReadyTime - startTime).toFixed(0)}ms`
			);

			return new Promise((resolve) => {
				const transaction = db.transaction(STORE_NAME, 'readonly');
				const store = transaction.objectStore(STORE_NAME);
				const request = store.get(cacheKey);

				request.onsuccess = async () => {
					const readTime = performance.now();
					const cached = request.result as CachedPDF | undefined;
					if (cached) {
						// Check if cache is expired
						if (Date.now() - cached.timestamp < MAX_CACHE_AGE) {
							const storageType = cached.storageType || 'arraybuffer';
							console.log(
								`PDF Cache: Cache hit (${storageType}), read took ${(
									readTime - dbReadyTime
								).toFixed(0)}ms, size: ${(cached.size / 1024 / 1024).toFixed(
									2
								)} MB`
							);

							// Convert Blob to ArrayBuffer if needed
							if (cached.data instanceof Blob) {
								try {
									const arrayBuffer = await cached.data.arrayBuffer();
									const convertTime = performance.now();
									console.log(
										`PDF Cache: Blob->ArrayBuffer conversion took ${(
											convertTime - readTime
										).toFixed(0)}ms`
									);
									resolve(arrayBuffer);
								} catch (e) {
									console.error('PDF Cache: Failed to convert Blob', e);
									resolve(null);
								}
							} else {
								resolve(cached.data as ArrayBuffer);
							}
						} else {
							console.log('PDF Cache: Cache expired', cacheKey);
							this.delete(cacheKey);
							resolve(null);
						}
					} else {
						console.log(
							`PDF Cache: Cache miss, query took ${(
								readTime - dbReadyTime
							).toFixed(0)}ms`
						);
						resolve(null);
					}
				};

				request.onerror = () => {
					console.error('PDF Cache: Read failed', request.error);
					resolve(null);
				};
			});
		} catch (error) {
			console.error('PDF Cache: Failed to get cache', error);
			return null;
		}
	}

	/**
	 * Check storage quota
	 */
	private async checkStorageQuota(): Promise<{
		usage: number;
		quota: number;
		available: number;
	}> {
		if ('storage' in navigator && 'estimate' in navigator.storage) {
			try {
				const estimate = await navigator.storage.estimate();
				const usage = estimate.usage || 0;
				const quota = estimate.quota || 0;
				return {
					usage,
					quota,
					available: quota - usage
				};
			} catch (e) {
				console.warn('PDF Cache: Unable to get storage quota', e);
			}
		}
		// If unable to get quota, return conservative estimate
		return { usage: 0, quota: 100 * 1024 * 1024, available: 100 * 1024 * 1024 };
	}

	/**
	 * Get effective max cache size
	 * Takes the minimum of config value and available space
	 */
	private async getEffectiveMaxCacheSize(): Promise<number> {
		const { quota, available } = await this.checkStorageQuota();

		// Calculate quota-based limit (use 30% of quota)
		const quotaBasedLimit = Math.floor(quota * QUOTA_USAGE_RATIO);

		// Calculate available-space-based limit (use 50% of available)
		const availableBasedLimit = Math.floor(available * 0.5);

		// Take minimum of: config value, quota limit, available space limit
		const effectiveLimit = Math.min(
			DEFAULT_MAX_CACHE_SIZE,
			quotaBasedLimit,
			availableBasedLimit
		);

		// Ensure at least 10MB cache space
		return Math.max(effectiveLimit, 10 * 1024 * 1024);
	}

	/**
	 * Extract file name from URL
	 */
	private extractFileName(url: string): string {
		try {
			const urlObj = new URL(url);
			const pathname = urlObj.pathname;
			// Get last segment of path and decode
			const fileName = decodeURIComponent(pathname.split('/').pop() || '');
			// Remove extension if it's .pdf
			return fileName || 'Unknown PDF';
		} catch {
			return 'Unknown PDF';
		}
	}

	/**
	 * Store PDF in cache
	 * @param url PDF URL
	 * @param data PDF binary data
	 * @param name Optional custom name (auto-extracted from URL if not provided)
	 */
	async set(url: string, data: ArrayBuffer, name?: string): Promise<void> {
		try {
			const fileSize = data.byteLength;
			const cacheKey = this.getCacheKey(url);

			console.log(
				`PDF Cache: Preparing to cache file ${(fileSize / 1024 / 1024).toFixed(
					2
				)} MB`
			);

			// Check available storage space
			const { available, quota } = await this.checkStorageQuota();

			// If file size exceeds available space, clean old cache first
			if (fileSize > available * 0.8) {
				console.warn(
					`PDF Cache: Insufficient space (need ${(
						fileSize /
						1024 /
						1024
					).toFixed(1)} MB, available ${(available / 1024 / 1024).toFixed(
						1
					)} MB), cleaning old cache`
				);
				await this.cleanup(fileSize);
			}

			// Check again after cleanup
			const afterCleanup = await this.checkStorageQuota();
			if (fileSize > afterCleanup.available * 0.9) {
				// If file size exceeds browser quota, skip
				if (fileSize > quota * 0.5) {
					console.warn(
						`PDF Cache: File too large (${(fileSize / 1024 / 1024).toFixed(
							1
						)} MB), exceeds 50% of browser storage quota, cannot cache`
					);
					return;
				}
				console.warn(
					'PDF Cache: Still insufficient space after cleanup, attempting forced cache'
				);
			}

			const db = await this.initDB();

			// Clean expired cache first
			await this.cleanup();

			// Prepare data based on storage mode
			let storageData: ArrayBuffer | Blob;
			if (this.storageMode === 'blob') {
				storageData = new Blob([data], { type: 'application/pdf' });
				console.log(`PDF Cache: Converting to Blob for storage`);
			} else {
				storageData = data;
			}

			const cachedPDF: CachedPDF = {
				url: cacheKey,
				data: storageData,
				timestamp: Date.now(),
				size: fileSize,
				storageType: this.storageMode,
				name: name || this.extractFileName(url)
			};

			return new Promise((resolve, reject) => {
				const transaction = db.transaction(STORE_NAME, 'readwrite');
				const store = transaction.objectStore(STORE_NAME);
				const request = store.put(cachedPDF);

				request.onsuccess = () => {
					const displayName = name || this.extractFileName(url);
					console.log(
						`PDF Cache: Cached "${displayName}"`,
						`(${(fileSize / 1024 / 1024).toFixed(2)} MB, ${this.storageMode})`
					);
					resolve();
				};

				request.onerror = () => {
					const error = request.error;
					// Handle quota exceeded error
					if (error?.name === 'QuotaExceededError') {
						console.warn(
							'PDF Cache: Storage quota exceeded, retrying after cleanup'
						);
						this.cleanup(fileSize)
							.then(() => {
								// Retry once
								const retryTx = db.transaction(STORE_NAME, 'readwrite');
								const retryStore = retryTx.objectStore(STORE_NAME);
								retryStore.put(cachedPDF);
								resolve();
							})
							.catch(() => resolve());
					} else {
						console.error('PDF Cache: Storage failed', error);
						resolve(); // Don't block main flow
					}
				};
			});
		} catch (error) {
			console.error('PDF Cache: Failed to set cache', error);
		}
	}

	/**
	 * List all cached PDFs (for debugging/management)
	 * Returns metadata only, not the actual data
	 */
	async list(): Promise<
		Array<{
			url: string;
			name: string;
			size: number;
			timestamp: number;
			storageType: string;
		}>
	> {
		try {
			const db = await this.initDB();

			return new Promise((resolve) => {
				const transaction = db.transaction(STORE_NAME, 'readonly');
				const store = transaction.objectStore(STORE_NAME);
				const request = store.getAll();

				request.onsuccess = () => {
					const items = (request.result as CachedPDF[]) || [];
					const list = items.map((item) => ({
						url: item.url,
						name: item.name || this.extractFileName(item.url),
						size: item.size,
						timestamp: item.timestamp,
						storageType: item.storageType || 'arraybuffer'
					}));
					// Sort by timestamp descending (newest first)
					list.sort((a, b) => b.timestamp - a.timestamp);
					resolve(list);
				};

				request.onerror = () => {
					console.error('PDF Cache: Failed to list cache');
					resolve([]);
				};
			});
		} catch (error) {
			console.error('PDF Cache: Failed to list cache', error);
			return [];
		}
	}

	/**
	 * Delete specific cache
	 */
	async delete(url: string): Promise<void> {
		try {
			const db = await this.initDB();
			const cacheKey = this.getCacheKey(url);

			return new Promise((resolve) => {
				const transaction = db.transaction(STORE_NAME, 'readwrite');
				const store = transaction.objectStore(STORE_NAME);
				store.delete(cacheKey);
				resolve();
			});
		} catch (error) {
			console.error('PDF Cache: Delete failed', error);
		}
	}

	/**
	 * Clean expired and over-limit cache
	 * @param requiredSpace Additional space needed (bytes)
	 */
	private async cleanup(requiredSpace = 0): Promise<void> {
		try {
			const db = await this.initDB();
			// Get dynamic max cache size
			const maxCacheSize = await this.getEffectiveMaxCacheSize();

			return new Promise((resolve) => {
				const transaction = db.transaction(STORE_NAME, 'readwrite');
				const store = transaction.objectStore(STORE_NAME);
				const index = store.index('timestamp');
				// Iterate in chronological order (oldest first)
				const request = index.openCursor();

				const allItems: CachedPDF[] = [];
				const now = Date.now();

				request.onsuccess = (event) => {
					const cursor = (event.target as IDBRequest<IDBCursorWithValue>)
						.result;
					if (cursor) {
						allItems.push(cursor.value as CachedPDF);
						cursor.continue();
					} else {
						// Iteration complete, start cleanup
						const toDelete: string[] = [];
						let totalSize = 0;
						let freedSpace = 0;

						// Sort by time (oldest first)
						allItems.sort((a, b) => a.timestamp - b.timestamp);

						for (const cached of allItems) {
							// Delete expired cache
							if (now - cached.timestamp > MAX_CACHE_AGE) {
								toDelete.push(cached.url);
								freedSpace += cached.size;
								continue;
							}

							totalSize += cached.size;

							// If total size exceeds limit or need more space
							const needMoreSpace =
								requiredSpace > 0 && freedSpace < requiredSpace;
							if (totalSize > maxCacheSize || needMoreSpace) {
								toDelete.push(cached.url);
								freedSpace += cached.size;
								totalSize -= cached.size;
							}
						}

						// Execute deletion
						toDelete.forEach((url) => store.delete(url));
						if (toDelete.length > 0) {
							console.log(
								`PDF Cache: Cleaned ${toDelete.length} cache entries, freed ${(
									freedSpace /
									1024 /
									1024
								).toFixed(2)} MB`
							);
							console.log(
								`PDF Cache: Current effective cache limit ${(
									maxCacheSize /
									1024 /
									1024
								).toFixed(2)} MB`
							);
						}
						resolve();
					}
				};

				request.onerror = () => {
					resolve();
				};
			});
		} catch (error) {
			console.error('PDF Cache: Cleanup failed', error);
		}
	}

	/**
	 * Clear all cache
	 */
	async clear(): Promise<void> {
		try {
			const db = await this.initDB();

			return new Promise((resolve, reject) => {
				const transaction = db.transaction(STORE_NAME, 'readwrite');
				const store = transaction.objectStore(STORE_NAME);
				const request = store.clear();

				request.onsuccess = () => {
					console.log('PDF Cache: All cache cleared');
					resolve();
				};

				request.onerror = () => {
					reject(request.error);
				};
			});
		} catch (error) {
			console.error('PDF Cache: Clear failed', error);
		}
	}
}

// Export singleton
export const pdfCache = new PDFCacheManager();

// Request lock to prevent concurrent requests for same URL
const pendingRequests = new Map<string, Promise<ArrayBuffer>>();

/**
 * PDF load result
 */
export interface PDFLoadResult {
	data: ArrayBuffer;
	fromCache: boolean;
}

// ==================== Progressive Loading ====================

/**
 * Download progress callback
 */
export interface DownloadProgress {
	loaded: number; // Bytes downloaded
	total: number; // Total bytes (0 if unknown)
	percent: number; // Percentage (0-100)
	speed: number; // Download speed (bytes/sec)
	supportsRange: boolean; // Whether server supports Range requests
}

/**
 * Progressive load options
 */
export interface ProgressiveLoadOptions {
	onProgress?: (progress: DownloadProgress) => void;
	chunkSize?: number; // Chunk size for Range requests (default: 256KB)
	signal?: AbortSignal; // AbortController signal for cancellation
}

/**
 * Check if server supports Range requests
 */
async function checkRangeSupport(
	url: string
): Promise<{ supportsRange: boolean; contentLength: number }> {
	try {
		const response = await fetch(url, {
			method: 'HEAD'
		});

		const acceptRanges = response.headers.get('Accept-Ranges');
		const contentLengthHeader = response.headers.get('Content-Length');
		// Validate content length is a reasonable number (> 1KB)
		const contentLength = contentLengthHeader
			? parseInt(contentLengthHeader, 10)
			: 0;
		const isValidLength = contentLength > 1024;

		return {
			supportsRange: acceptRanges === 'bytes' && isValidLength,
			contentLength: isValidLength ? contentLength : 0
		};
	} catch {
		return { supportsRange: false, contentLength: 0 };
	}
}

/**
 * Download with progress tracking using streaming
 */
async function downloadWithProgress(
	url: string,
	options: ProgressiveLoadOptions = {}
): Promise<ArrayBuffer> {
	const { onProgress, signal } = options;
	const startTime = performance.now();

	// First check Range support
	const { supportsRange, contentLength } = await checkRangeSupport(url);

	console.log(
		`PDF Download: Range support: ${supportsRange}, Content-Length: ${(
			contentLength /
			1024 /
			1024
		).toFixed(2)} MB`
	);

	const response = await fetch(url, { signal });

	if (!response.ok) {
		throw new Error(`Failed to download PDF: ${response.status}`);
	}

	// If no body or no reader, fall back to simple download
	if (!response.body) {
		const data = await response.arrayBuffer();
		onProgress?.({
			loaded: data.byteLength,
			total: data.byteLength,
			percent: 100,
			speed: 0,
			supportsRange: false
		});
		return data;
	}

	// Use streaming to track progress
	const reader = response.body.getReader();
	// Get content length from response header first, fall back to HEAD request result
	const responseContentLength = parseInt(
		response.headers.get('Content-Length') || '0',
		10
	);
	// Use response header if valid (> 1KB), otherwise use HEAD result, otherwise 0
	const total =
		responseContentLength > 1024
			? responseContentLength
			: contentLength > 1024
			? contentLength
			: 0;

	const chunks: Uint8Array[] = [];
	let loaded = 0;
	let lastLoaded = 0;
	let lastTime = startTime;
	let done = false;

	while (!done) {
		const result = await reader.read();
		done = result.done;

		if (result.value) {
			chunks.push(result.value);
			loaded += result.value.length;

			// Calculate speed
			const now = performance.now();
			const timeDelta = (now - lastTime) / 1000; // seconds
			const speed = timeDelta > 0 ? (loaded - lastLoaded) / timeDelta : 0;

			if (timeDelta > 0.1) {
				// Update every 100ms
				lastLoaded = loaded;
				lastTime = now;

				onProgress?.({
					loaded,
					total,
					percent: total > 0 ? Math.round((loaded / total) * 100) : 0,
					speed,
					supportsRange
				});
			}
		}
	}

	// Final progress update
	onProgress?.({
		loaded,
		total: loaded,
		percent: 100,
		speed: 0,
		supportsRange
	});

	// Combine chunks into single ArrayBuffer
	const result = new Uint8Array(loaded);
	let offset = 0;
	for (const chunk of chunks) {
		result.set(chunk, offset);
		offset += chunk.length;
	}

	const elapsed = (performance.now() - startTime) / 1000;
	console.log(
		`PDF Download: Complete in ${elapsed.toFixed(1)}s, avg speed: ${(
			loaded /
			elapsed /
			1024 /
			1024
		).toFixed(2)} MB/s`
	);

	return result.buffer;
}

/**
 * Load PDF with cache and progress tracking
 * @param url PDF file URL
 * @param options Progressive load options
 * @returns PDFLoadResult
 */
export async function fetchPDFWithProgress(
	url: string,
	options: ProgressiveLoadOptions = {}
): Promise<PDFLoadResult> {
	const { onProgress, signal } = options;
	const cacheKey = pdfCache.getCacheKey(url);

	// Try cache first
	const cached = await pdfCache.get(url);
	if (cached) {
		console.log('PDF Cache: Cache hit', cacheKey);
		onProgress?.({
			loaded: cached.byteLength,
			total: cached.byteLength,
			percent: 100,
			speed: 0,
			supportsRange: false
		});
		return { data: cached, fromCache: true };
	}

	// Check for pending request
	const pendingRequest = pendingRequests.get(cacheKey);
	if (pendingRequest) {
		console.log('PDF Cache: Waiting for pending request', cacheKey);
		const data = await pendingRequest;
		return { data, fromCache: false };
	}

	// Download with progress
	console.log('PDF Cache: Downloading with progress tracking', cacheKey);

	const requestPromise = (async () => {
		try {
			const arrayBuffer = await downloadWithProgress(url, {
				onProgress,
				signal
			});

			// Cache the result
			await pdfCache.set(url, arrayBuffer);

			return arrayBuffer;
		} finally {
			pendingRequests.delete(cacheKey);
		}
	})();

	pendingRequests.set(cacheKey, requestPromise);

	const data = await requestPromise;
	return { data, fromCache: false };
}

/**
 * Format bytes to human readable string
 */
export function formatBytes(bytes: number): string {
	if (bytes < 1024) return bytes + ' B';
	if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
	return (bytes / 1024 / 1024).toFixed(2) + ' MB';
}

/**
 * Format speed to human readable string
 */
export function formatSpeed(bytesPerSec: number): string {
	if (bytesPerSec < 1024) return bytesPerSec.toFixed(0) + ' B/s';
	if (bytesPerSec < 1024 * 1024)
		return (bytesPerSec / 1024).toFixed(1) + ' KB/s';
	return (bytesPerSec / 1024 / 1024).toFixed(2) + ' MB/s';
}
