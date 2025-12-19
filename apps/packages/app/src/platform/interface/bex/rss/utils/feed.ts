import { FeedCreationRequest } from '../model/type';
import { useCollectStore } from 'src/stores/collect';
import axios from 'axios';

export async function queryFeed(url: string) {
	const collectStore = useCollectStore();
	try {
		const response: any = await axios.get(
			collectStore.baseUrl + '/knowledge/feed',
			{
				params: {
					feed_url: url,
					source: 'wise'
				}
			}
		);
		return !!(response && response.count > 0);
	} catch (e) {
		return false;
	}
}

export async function saveFeed(req: FeedCreationRequest) {
	const collectStore = useCollectStore();
	try {
		return await axios.post(collectStore.baseUrl + '/knowledge/feed', req);
	} catch (e) {
		return undefined;
	}
}
