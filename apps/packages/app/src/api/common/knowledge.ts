import axios, { AxiosResponse, CancelToken } from 'axios';
import { useConfigStore } from 'src/stores/rss-config';
import { CollectInfo } from 'src/types/commonApi';
import { useUserStore } from 'src/stores/user';

const rawAxios = axios.create();

export async function getCollectInfo(
	url: string,
	cancelToken?: CancelToken
): Promise<any> {
	const configStore = useConfigStore();
	const params = { url };

	try {
		const response = await rawAxios.get(
			configStore.url + '/knowledge/page/collect_info',
			{
				params,
				cancelToken,
				headers: {
					'X-Authorization': getAuthToken()
				}
			}
		);
		return response.data;
	} catch (error: any) {
		if (axios.isCancel(error)) {
			throw error;
		}
		throw error;
	}
}

function getAuthToken(): string {
	const userStore = useUserStore();
	if (!userStore.current_id) {
		return '';
	}
	const user = userStore.users?.items.get(userStore.current_id);
	return user?.access_token || '';
}
