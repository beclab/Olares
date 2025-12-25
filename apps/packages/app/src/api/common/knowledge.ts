import axios, { AxiosResponse, CancelToken } from 'axios';
import { useConfigStore } from 'src/stores/rss-config';
import { CollectInfo } from 'src/types/commonApi';
export async function getCollectInfo(
	url: string,
	cancelToken?: CancelToken
): Promise<CollectInfo> {
	const configStore = useConfigStore();
	const params = { url };
	return await axios.get(configStore.url + '/knowledge/page/collect_info', {
		params,
		cancelToken
	});
}
