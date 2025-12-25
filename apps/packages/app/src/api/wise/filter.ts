import { Entry, FilterInfo } from 'src/utils/rss-types';
import { useConfigStore } from 'src/stores/rss-config';
import { QueryParamsBuilder } from './queryParamsBuilder';
import axios from 'axios';

export async function syncFilterEntries(
	query: string,
	lastTime: number
): Promise<Entry[]> {
	const configStore = useConfigStore();
	const builder = new QueryParamsBuilder({
		params: { query, lastTime }
	});
	return await axios.get(
		configStore.url + '/knowledge/filter/entryQuery' + builder.build()
	);
}

export async function getFilterList(): Promise<any> {
	const configStore = useConfigStore();
	const builder = new QueryParamsBuilder();
	const data: any = await axios.get(
		configStore.url + '/knowledge/filter' + builder.build()
	);
	return data && data.items ? data.items : [];
}

export async function addFilter(
	name: string,
	description: string,
	query: string,
	sortby: string,
	orderby: string,
	splitview: string
): Promise<FilterInfo> {
	const configStore = useConfigStore();
	return await axios.post(configStore.url + '/knowledge/filter', {
		name,
		description,
		query,
		sortby,
		orderby,
		splitview
	});
}

export async function updateFilter(info: FilterInfo): Promise<FilterInfo> {
	const configStore = useConfigStore();
	const { id, ...request } = info;
	return await axios.put(configStore.url + '/knowledge/filter/' + id, {
		...request
	});
}

export async function deleteFilter(id: string): Promise<FilterInfo[]> {
	const configStore = useConfigStore();
	return await axios.delete(configStore.url + '/knowledge/filter/' + id);
}
