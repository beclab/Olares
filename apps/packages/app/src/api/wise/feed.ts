import axios from 'axios';
import { Feed, SearchFeed, SOURCE_TYPE } from 'src/utils/rss-types';
import { useConfigStore } from 'src/stores/rss-config';

export async function syncFeeds(time: number): Promise<Feed[]> {
	const configStore = useConfigStore();
	return await axios.get(configStore.url + '/knowledge/feed/sync/' + time);
}

export async function syncWaitFeeds(feeds: string[]): Promise<Feed[]> {
	const configStore = useConfigStore();
	return await axios.post(configStore.url + '/knowledge/feed/sync/', feeds);
}

export async function searchFeed(content: string): Promise<SearchFeed[]> {
	const configStore = useConfigStore();
	return await axios.get(
		configStore.url + '/knowledge/feed/search?content=' + content
	);
}

export async function createFeed(
	feed_url: string,
	auto_download: boolean
): Promise<Feed> {
	const configStore = useConfigStore();
	return await axios.post(configStore.url + '/knowledge/feed', {
		source: SOURCE_TYPE.WISE,
		feed_url,
		auto_download: auto_download
	});
}

export async function updateFeed(
	feed_url: string,
	title: string,
	description: string,
	site_url: string,
	auto_download: boolean
): Promise<Feed> {
	const configStore = useConfigStore();
	return await axios.post(configStore.url + '/knowledge/feed', {
		source: SOURCE_TYPE.WISE,
		feed_url,
		title,
		description,
		site_url,
		auto_download
	});
}

export async function deleteFeeds(
	feed_url: string[],
	removeFile: boolean
): Promise<Feed[]> {
	const configStore = useConfigStore();
	return await axios.delete(
		configStore.url + '/knowledge/feed/algorithm/wise',
		{
			data: {
				feed_urls: feed_url,
				remove_flag: removeFile
			}
		}
	);
}

export async function subscribeTrendFeed(feed_id: string) {
	const configStore = useConfigStore();
	return await axios.post(configStore.url + '/knowledge/feed/save', {
		feed_id
	});
}

export async function exportFeedAsOpml() {
	const configStore = useConfigStore();
	const response = await axios.get(
		configStore.url + '/knowledge/opml/download',
		{
			responseType: 'blob'
		}
	);
	return new Blob([response.data], { type: 'text/xml' });
}

export async function importFeedAsOpml(formData) {
	const configStore = useConfigStore();
	return await axios.post(
		configStore.url + '/knowledge/opml/upload',
		formData,
		{
			headers: {
				'Content-Type': 'multipart/form-data'
			}
		}
	);
}
