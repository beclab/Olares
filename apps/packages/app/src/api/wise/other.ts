import axios from 'axios';
import {
	RemoveData,
	RssContentQuery,
	SDKQueryRequest
} from 'src/utils/rss-types';
import { useConfigStore } from 'src/stores/rss-config';

export const rssContentQuery = async (query: string) => {
	const configStore = useConfigStore();
	console.log(query);

	let data: string = await axios.get(
		configStore.url + '/api/rss/contentQuery',
		{
			params: {
				query
			}
		}
	);
	data = data.replace(/[\n]/g, '');
	console.log(data);
	if (data.length === 0) {
		return [];
	}

	const mode = JSON.parse(data) as RssContentQuery;
	console.log(mode);
	if (mode.code === 0 && mode.data && mode.data.data) {
		const dataString = mode.data.data as unknown as string;
		mode.data.data = JSON.parse(dataString);
		console.log(mode);
		return mode.data.data.items;
	}
	return [];
};

export async function getAbiAbility() {
	const configStore = useConfigStore();
	return await axios.get(configStore.url + '/knowledge/page/abilities');
}

export async function sdkSearchFeedsByPath(q: SDKQueryRequest) {
	const configStore = useConfigStore();
	return await axios.get(configStore.sdkUrl + '/rss' + q.build());
}

export async function syncRemove(time: number): Promise<RemoveData[]> {
	const configStore = useConfigStore();
	return await axios.get(configStore.url + '/knowledge/remove/sync/' + time);
}

export async function fetchBlackList(baseUrl: string): Promise<any> {
	return await axios.get(baseUrl + '/knowledge/page/blacklist');
}
