import axios from 'axios';
import { useConfigStore } from 'src/stores/rss-config';
import { Entry, RecommendAlgorithm } from 'src/utils/rss-types';

// export async function getRecommendAlgorithmList(): Promise<
// 	RecommendAlgorithm[]
// > {
// 	const configStore = useConfigStore();
// 	return await axios.get(
// 		configStore.url + '/knowledge/algorithm/recommend/methods'
// 	);
// }

export async function getRecommendEntryList(
	source: string,
	pageSize: number
): Promise<Entry[]> {
	const configStore = useConfigStore();
	return await axios.get(
		configStore.url +
			'/knowledge/algorithm/recommend/' +
			source +
			'/' +
			pageSize
	);
}

export async function fireImpression(
	source: string,
	entryId: string
): Promise<void> {
	const configStore = useConfigStore();
	return await axios.post(
		configStore.url + '/knowledge/impression/fire/' + source + '/' + entryId
	);
}
