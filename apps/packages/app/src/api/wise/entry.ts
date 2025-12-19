import axios, { CancelToken } from 'axios';
import {
	CreateEntry,
	EntriesStatusUpdateRequest,
	UpdateImpression,
	Entry,
	Enclosure,
	UrlCheck,
	SOURCE_TYPE
} from 'src/utils/rss-types';
import { useConfigStore } from 'src/stores/rss-config';

/**
 * When deleting an entry, if the user hovers over the current item while the deletion is in progress and the
 * onEntryUpdate method triggers an update, it will result in failure to retrieve the entry data from the remote
 * server, throwing an error: "not found 'xx'". Therefore, a noToast request config is added to this interface to
 * avoid this issue.
 * @param id
 */
export async function getRemoteEntryById(id: string): Promise<Entry> {
	console.log('test111 RemoteEntryById', id);
	const configStore = useConfigStore();
	return await axios.get(configStore.url + '/knowledge/entry/' + id, {
		noToast: true
	});
}

export async function syncEntries(time: number): Promise<Entry[]> {
	const configStore = useConfigStore();
	return await axios.get(configStore.url + '/knowledge/entry/sync/' + time);
}

export async function removeEntrySource(
	source: string,
	entry_urls: string[],
	removeFile: boolean
) {
	const configStore = useConfigStore();
	return await axios.delete(configStore.url + '/knowledge/entry/' + source, {
		data: {
			entry_urls: entry_urls,
			download_remove_flag: removeFile
		}
	});
}

export async function removeEntry(urls: string[], removeFile: boolean) {
	const configStore = useConfigStore();
	return await axios.delete(configStore.url + '/knowledge/entry/wise/remove', {
		data: {
			entry_urls: urls,
			remove_flag: removeFile
		}
	});
}

export async function updateReadProgress(
	id: string,
	progress: string,
	played_time: string,
	remaining_time: string
): Promise<Entry> {
	const configStore = useConfigStore();
	return await axios.put(configStore.url + '/knowledge/entry/process', {
		entry_id: id,
		process: progress,
		played_time,
		remaining_time
	});
}

export async function saveEntry(
	req: CreateEntry[],
	baseUrl = ''
): Promise<Entry[]> {
	let requestBase = '';
	if (baseUrl) {
		requestBase = baseUrl;
	} else {
		const configStore = useConfigStore();
		requestBase = configStore.url;
	}
	return await axios.post(requestBase + '/knowledge/entry', req);
}

export async function urlCheck(
	url: string,
	cancelToken?: CancelToken
): Promise<UrlCheck | null> {
	const configStore = useConfigStore();
	try {
		return await axios.get(
			configStore.url + '/knowledge/entry/entryCookieCheck?url=' + url,
			{
				cancelToken
			}
		);
	} catch (e) {
		return null;
	}
}

export async function updateEntryUnread(
	q: EntriesStatusUpdateRequest
): Promise<string[]> {
	const configStore = useConfigStore();
	return await axios.put(configStore.url + '/knowledge/entry/unread', q);
}

export async function updateEntryReadLater(
	q: EntriesStatusUpdateRequest
): Promise<string[]> {
	const configStore = useConfigStore();
	return await axios.put(configStore.url + '/knowledge/entry/read-later', q);
}

export async function updateImpression(impression: UpdateImpression) {
	const configStore = useConfigStore();

	await axios.put(configStore.url + '/knowledge/impression', impression);
}

export async function findEnclosure(id: string): Promise<Enclosure[]> {
	const configStore = useConfigStore();
	return await axios.get(configStore.url + '/knowledge/entry/enclosures/' + id);
}

export async function getRecentlyEntryList(
	offset: number,
	limit: number
): Promise<Entry[]> {
	const configStore = useConfigStore();
	try {
		const result: any = await axios.get(
			configStore.url + '/knowledge/entry/recentlyRead/',
			{
				params: {
					offset,
					limit
				}
			}
		);
		console.log(result);
		if (result && result.items) {
			return result.items;
		}
		return [];
	} catch (e) {
		return [];
	}
}

export async function queryEntry(
	url: string,
	cancelToken?: CancelToken,
	baseUrl = ''
) {
	let requestBase = '';
	if (baseUrl) {
		requestBase = baseUrl;
	} else {
		const configStore = useConfigStore();
		requestBase = configStore.url;
	}
	const response: any = await axios.post(
		requestBase + '/knowledge/entry/query',
		{
			url,
			source: SOURCE_TYPE.LIBRARY
		},
		{
			cancelToken
		}
	);
	return response;
}

export async function batchEntries(url: string): Promise<string[]> {
	const configStore = useConfigStore();
	return await axios.post(
		configStore.url + '/knowledge/page/batchCollect',
		url,
		{ headers: { 'Content-Type': 'text/plain' } }
	);
}
