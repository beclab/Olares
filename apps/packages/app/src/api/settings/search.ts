import { useTokenStore } from 'src/stores/settings/token';
import axios from 'axios';

export async function getSearchTaskStatus(): Promise<any> {
	const tokenStore = useTokenStore();
	return await axios.get(`${tokenStore.url}/api/search/task/stats/merged`);
}

export async function rebuildSearchTask(): Promise<any> {
	const tokenStore = useTokenStore();
	return await axios.post(`${tokenStore.url}/api/search/task/rebuild`);
}

export async function getExcludePatterns() {
	const tokenStore = useTokenStore();
	return await axios.get(
		`${tokenStore.url}/api/search/monitorsetting/exclude-pattern`
	);
}

export async function addExcludePattern(values: string[]) {
	const tokenStore = useTokenStore();
	return await axios.put(
		`${tokenStore.url}/api/search/monitorsetting/exclude-pattern/part`,
		{
			exclude_pattern: values
		}
	);
}

export async function deleteExcludePattern(values: string[]) {
	const tokenStore = useTokenStore();
	return await axios.delete(
		`${tokenStore.url}/api/search/monitorsetting/exclude-pattern/part`,
		{
			data: {
				exclude_pattern: values
			}
		}
	);
}

export async function getSearchDirectories() {
	const tokenStore = useTokenStore();
	return await axios.get(
		`${tokenStore.url}/api/search/monitorsetting/include-directory/full_content`
	);
}

export async function addSearchDirectories(values: string[]) {
	const tokenStore = useTokenStore();
	return await axios.put(
		`${tokenStore.url}/api/search/monitorsetting/include-directory/full_content/part`,
		{
			include_directory: values
		}
	);
}

export async function deleteSearchDirectories(values: string[]) {
	const tokenStore = useTokenStore();
	return await axios.delete(
		`${tokenStore.url}/api/search/monitorsetting/include-directory/full_content/part`,
		{
			data: {
				include_directory: values
			}
		}
	);
}
