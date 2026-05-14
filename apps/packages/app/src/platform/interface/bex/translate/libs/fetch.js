/* eslint-disable no-undef */
import { isExt, isGm } from './client';
import { sendBgMsg } from './msg';
import { taskPool } from './pool';
import {
	MSG_FETCH,
	MSG_GET_HTTPCACHE,
	MSG_PUT_HTTPCACHE,
	MSG_CLEAR_CACHES,
	CACHE_NAME,
	DEFAULT_FETCH_INTERVAL,
	DEFAULT_FETCH_LIMIT,
	DEFAULT_CACHE_TIMEOUT,
	DEFAULT_HTTP_TIMEOUT
} from '../config';
import { isBg } from './browser';
import { genTransReq } from '../apis/trans';
import { kissLog } from './log';
import { blobToBase64 } from './utils';
import { performanceMonitor } from './performance';

const TIMEOUT = DEFAULT_HTTP_TIMEOUT;

const newCacheReq = async (input, init) => {
	let request = new Request(input, init);
	if (request.method !== 'GET') {
		const body = await request.text();
		const cacheUrl = new URL(request.url);
		cacheUrl.pathname += body;
		request = new Request(cacheUrl.toString(), { method: 'GET' });
	}

	return request;
};

export const fetchGM = async (input, { method = 'GET', headers, body } = {}) =>
	new Promise((resolve, reject) => {
		GM.xmlHttpRequest({
			method,
			url: input,
			headers,
			data: body,
			timeout: TIMEOUT,
			onload: ({ response, responseHeaders, status, statusText }) => {
				const headers = {};
				responseHeaders.split('\n').forEach((line) => {
					const [name, value] = line.split(':').map((item) => item.trim());
					if (name && value) {
						headers[name] = value;
					}
				});
				resolve({
					body: response,
					headers,
					status,
					statusText
				});
			},
			onerror: reject
		});
	});

export const fetchPatcher = async (input, init, transOpts, apiSetting) => {
	if (transOpts?.translator) {
		[input, init] = await genTransReq(transOpts, apiSetting);
	}

	if (!input) {
		throw new Error('url is empty');
	}

	if (isGm) {
		let info;
		if (window.KISS_GM) {
			info = await window.KISS_GM.getInfo();
		} else {
			info = GM.info;
		}

		const connects = info?.script?.connects || info?.script?.connect || [];
		const url = new URL(input);
		const isSafe = connects.find((item) => url.hostname.endsWith(item));

		if (isSafe) {
			const { body, headers, status, statusText } = window.KISS_GM
				? await window.KISS_GM.fetch(input, init)
				: await fetchGM(input, init);

			return new Response(body, {
				headers: new Headers(headers),
				status,
				statusText
			});
		}
	}

	if (AbortSignal?.timeout) {
		Object.assign(init, { signal: AbortSignal.timeout(TIMEOUT) });
	}

	return fetch(input, init);
};

const parseResponse = async (res) => {
	if (!res) {
		return null;
	}

	const contentType = res.headers.get('Content-Type');
	if (contentType?.includes('json')) {
		return await res.json();
	} else if (contentType?.includes('audio')) {
		const blob = await res.blob();
		return await blobToBase64(blob);
	}
	return await res.text();
};

export const getHttpCache = async (input, { method, headers, body }) => {
	try {
		const req = await newCacheReq(input, { method, headers, body });
		const cache = await caches.open(CACHE_NAME);
		const res = await cache.match(req);

		if (!res) {
			return null;
		}

		const cacheControl = res.headers.get('Cache-Control');
		if (cacheControl) {
			const maxAgeMatch = cacheControl.match(/max-age=(\d+)/);
			if (maxAgeMatch) {
				const maxAge = parseInt(maxAgeMatch[1], 10);
				const cachedTime = parseInt(
					res.headers.get('X-Cached-Time') || '0',
					10
				);
				const now = Math.floor(Date.now() / 1000);

				if (now - cachedTime > maxAge) {
					await cache.delete(req);
					return null;
				}
			}
		}

		return parseResponse(res);
	} catch (err) {
		kissLog(err, 'get cache');
	}
	return null;
};

export const putHttpCache = async (
	input,
	{ method, headers, body },
	res,
	maxAge = DEFAULT_CACHE_TIMEOUT
) => {
	try {
		const req = await newCacheReq(input, { method, headers, body });
		const cache = await caches.open(CACHE_NAME);

		const clonedRes = res.clone();
		const resBody = await clonedRes.text();
		const newHeaders = new Headers(clonedRes.headers);

		newHeaders.set('Cache-Control', `max-age=${maxAge}`);
		newHeaders.set('X-Cached-Time', Math.floor(Date.now() / 1000).toString());

		const newRes = new Response(resBody, {
			status: clonedRes.status,
			statusText: clonedRes.statusText,
			headers: newHeaders
		});

		await cache.put(req, newRes);
	} catch (err) {
		kissLog(err, 'put cache');
	}
};

export const fetchHandle = async ({
	input,
	useCache,
	transOpts,
	apiSetting,
	...init
}) => {
	const res = await fetchPatcher(input, init, transOpts, apiSetting);
	if (!res) {
		throw new Error('Unknow error');
	} else if (!res.ok) {
		const msg = {
			url: res.url,
			status: res.status
		};
		if (res.headers.get('Content-Type')?.includes('json')) {
			msg.response = await res.json();
		}
		throw new Error(JSON.stringify(msg));
	}

	if (useCache) {
		await putHttpCachePolyfill(input, init, res.clone());
	}

	return parseResponse(res);
};

export const fetchPolyfill = (args) => {
	console.log('fetchPolyfill', isExt, !isBg);
	if (isExt && !isBg()) {
		return sendBgMsg(MSG_FETCH, args);
	}

	return fetchHandle(args);
};

export const getHttpCachePolyfill = (input, init) => {
	if (isExt && !isBg()) {
		return sendBgMsg(MSG_GET_HTTPCACHE, { input, init });
	}

	return getHttpCache(input, init);
};

export const putHttpCachePolyfill = (input, init, res, maxAge) => {
	if (isExt && !isBg()) {
		return sendBgMsg(MSG_PUT_HTTPCACHE, { input, init, data: res, maxAge });
	}

	return putHttpCache(input, init, res, maxAge);
};

export const clearAllCaches = async () => {
	try {
		if (isExt && !isBg()) {
			await sendBgMsg(MSG_CLEAR_CACHES);
		} else {
			await caches.delete(CACHE_NAME);
		}
	} catch (err) {
		kissLog(err, 'clear caches');
	}
};

export const fetchPool = taskPool(
	fetchPolyfill,
	null,
	DEFAULT_FETCH_INTERVAL,
	DEFAULT_FETCH_LIMIT
);

export const fetchData = async (input, { useCache, usePool, ...args } = {}) => {
	if (!input?.trim()) {
		throw new Error('URL is empty');
	}

	const startTime = performance.now();

	if (useCache) {
		const cache = await getHttpCachePolyfill(input, args);
		if (cache) {
			const duration = performance.now() - startTime;
			performanceMonitor.recordCacheHit();
			performanceMonitor.recordRequestTime(duration);
			return cache;
		}
	}

	try {
		let result;

		if (usePool) {
			result = await fetchPool.push({ input, useCache, ...args });
		} else {
			result = await fetchPolyfill({ input, useCache, ...args });
		}

		const duration = performance.now() - startTime;
		if (useCache) {
			performanceMonitor.recordCacheMiss();
		}
		performanceMonitor.recordRequestTime(duration);

		return result;
	} catch (err) {
		const duration = performance.now() - startTime;
		if (useCache) {
			performanceMonitor.recordCacheMiss();
		}
		performanceMonitor.recordRequestTime(duration);
		throw err;
	}
};

export const updateFetchPool = (interval, limit) => {
	fetchPool.update(interval, limit);
};

export const clearFetchPool = () => {
	fetchPool.clear();
};
