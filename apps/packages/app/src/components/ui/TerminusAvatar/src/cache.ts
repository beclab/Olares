export const MAX_CACHE_BYTES = 20 * 1024 * 1024;
export const CACHE_DURATION = 7 * 24 * 3600 * 1000;
export const CHCHE_PREFIX = 'img_cache_list';

import localStorage from 'localforage/src/localforage';

export function estimateUtf8Bytes(str: string): number {
	return new TextEncoder().encode(str).length;
}

export function estimateListBytes(
	list: Array<{ key: string; value: string; expire: number }>
): number {
	return estimateUtf8Bytes(JSON.stringify(list));
}

export function trimCacheBySize(
	list: Array<{ key: string; value: string; expire: number }>,
	keepKey?: string
): Array<{ key: string; value: string; expire: number }> {
	let next = [...list];
	const totalBytes = estimateListBytes(next);
	if (totalBytes <= MAX_CACHE_BYTES) {
		return next;
	}
	next = next.filter((item) => item.expire > Date.now());

	while (estimateListBytes(next) > MAX_CACHE_BYTES && next.length > 0) {
		const sorted = next
			.map((item, i) => ({ item, i }))
			.sort((a, b) => a.item.expire - b.item.expire);
		const pick =
			sorted.find((x) => !keepKey || x.item.key !== keepKey) ?? sorted[0];
		next.splice(pick.i, 1);
	}
	return next;
}

export async function getCacheImg(src: string) {
	const imgCacheListString = await localStorage.getItem(CHCHE_PREFIX);
	const imgCacheList =
		imgCacheListString != undefined
			? typeof imgCacheListString == 'string'
				? JSON.parse(imgCacheListString)
				: imgCacheListString
			: [];

	const key = btoa(src);
	return imgCacheList.find(
		(item: { key: string; value: string }) => item.key === key
	);
}

export async function saveCacheImg(base64: string, src: string) {
	const key = btoa(src);
	const imgCacheListString = await localStorage.getItem(CHCHE_PREFIX);

	const imgCacheList =
		imgCacheListString != undefined
			? typeof imgCacheListString == 'string'
				? JSON.parse(imgCacheListString)
				: imgCacheListString
			: [];
	const isExit = imgCacheList.find((item: any) => item.key === key);
	if (!isExit) {
		imgCacheList.push({
			key,
			value: base64,
			expire: Date.now() + CACHE_DURATION
		});
	} else {
		isExit.value = base64;
		isExit.expire = Date.now() + CACHE_DURATION;
	}
	const trimmed = trimCacheBySize(imgCacheList, key);
	await localStorage.setItem(CHCHE_PREFIX, JSON.stringify(trimmed));
}

export const delay = (ms: number) => {
	return new Promise((resolve) => {
		const timer = setTimeout(() => {
			resolve(undefined);
			clearTimeout(timer);
		}, ms);
	});
};
