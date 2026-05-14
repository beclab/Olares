/* eslint-disable no-undef */
import { kissLog } from './log';

class PerformanceMonitor {
	constructor() {
		this.metrics = {
			cacheHits: 0,
			cacheMisses: 0,
			totalRequests: 0,
			avgResponseTime: 0,
			requestTimes: [],
			streamFirstByteTime: []
		};
		this.maxSamples = 100;
	}

	recordCacheHit() {
		this.metrics.cacheHits++;
		this.metrics.totalRequests++;
	}

	recordCacheMiss() {
		this.metrics.cacheMisses++;
		this.metrics.totalRequests++;
	}

	recordRequestTime(time) {
		this.metrics.requestTimes.push(time);
		if (this.metrics.requestTimes.length > this.maxSamples) {
			this.metrics.requestTimes.shift();
		}

		const sum = this.metrics.requestTimes.reduce((a, b) => a + b, 0);
		this.metrics.avgResponseTime = sum / this.metrics.requestTimes.length || 0;
	}

	recordStreamTTFB(time) {
		this.metrics.streamFirstByteTime.push(time);
		if (this.metrics.streamFirstByteTime.length > this.maxSamples) {
			this.metrics.streamFirstByteTime.shift();
		}
	}

	getCacheHitRate() {
		if (this.metrics.totalRequests === 0) return 0;
		return (
			(this.metrics.cacheHits / this.metrics.totalRequests) *
			100
		).toFixed(2);
	}

	getAvgStreamTTFB() {
		if (this.metrics.streamFirstByteTime.length === 0) return 0;
		const sum = this.metrics.streamFirstByteTime.reduce((a, b) => a + b, 0);
		return (sum / this.metrics.streamFirstByteTime.length).toFixed(2);
	}

	getMetrics() {
		return {
			cacheHitRate: this.getCacheHitRate() + '%',
			cacheHits: this.metrics.cacheHits,
			cacheMisses: this.metrics.cacheMisses,
			totalRequests: this.metrics.totalRequests,
			avgResponseTime: this.metrics.avgResponseTime.toFixed(2) + 'ms',
			avgStreamTTFB: this.getAvgStreamTTFB() + 'ms',
			samples: {
				requestTimes: this.metrics.requestTimes.length,
				streamTTFB: this.metrics.streamFirstByteTime.length
			}
		};
	}

	printReport() {
		const metrics = this.getMetrics();
		console.group('📊 Translation Performance Report');
		console.log('Cache Hit Rate:', metrics.cacheHitRate);
		console.log('Total Requests:', metrics.totalRequests);
		console.log('Cache Hits:', metrics.cacheHits);
		console.log('Cache Misses:', metrics.cacheMisses);
		console.log('Average Response Time:', metrics.avgResponseTime);
		console.log('Average Stream TTFB:', metrics.avgStreamTTFB);
		console.log('Sample Sizes:', metrics.samples);
		console.groupEnd();
	}

	reset() {
		this.metrics = {
			cacheHits: 0,
			cacheMisses: 0,
			totalRequests: 0,
			avgResponseTime: 0,
			requestTimes: [],
			streamFirstByteTime: []
		};
	}

	async measureTime(fn, label = 'Operation') {
		const startTime = performance.now();
		try {
			const result = await fn();
			const endTime = performance.now();
			const duration = endTime - startTime;

			kissLog(`${label} took ${duration.toFixed(2)}ms`, 'performance');

			return result;
		} catch (err) {
			const endTime = performance.now();
			const duration = endTime - startTime;
			kissLog(`${label} failed after ${duration.toFixed(2)}ms`, 'performance');
			throw err;
		}
	}
}

const performanceMonitor = new PerformanceMonitor();

export { performanceMonitor, PerformanceMonitor };

export const withPerformanceTracking = (fetchFn) => {
	return async (input, options = {}) => {
		const startTime = performance.now();
		const { useCache } = options;

		try {
			const result = await fetchFn(input, options);
			const endTime = performance.now();
			const duration = endTime - startTime;

			performanceMonitor.recordRequestTime(duration);

			if (useCache && duration < 10) {
				performanceMonitor.recordCacheHit();
			} else if (useCache) {
				performanceMonitor.recordCacheMiss();
			}

			return result;
		} catch (err) {
			const endTime = performance.now();
			const duration = endTime - startTime;
			performanceMonitor.recordRequestTime(duration);

			if (useCache) {
				performanceMonitor.recordCacheMiss();
			}

			throw err;
		}
	};
};

export async function* withStreamPerformanceTracking(streamGenerator) {
	const startTime = performance.now();
	let firstByteReceived = false;

	try {
		for await (const chunk of streamGenerator) {
			if (!firstByteReceived) {
				const ttfb = performance.now() - startTime;
				performanceMonitor.recordStreamTTFB(ttfb);
				firstByteReceived = true;
			}
			yield chunk;
		}
	} catch (err) {
		kissLog('Stream error:', err);
		throw err;
	}
}

if (typeof window !== 'undefined') {
	window.__TERMIPASS_PERF__ = performanceMonitor;
}
