import { useDataStore } from 'src/stores/data';
import { axiosInstanceProxy } from 'src/platform/httpProxy';

import { MenuItem } from 'src/utils/contact';
import { SyncRepoItemType, SyncRepoSharedItemType } from './type';
import { useFilesStore } from 'src/stores/files';
import { CommonFetch } from '../../fetch';
import { encodeUrl } from 'src/utils/encode';

export async function instanceAxios(config) {
	const store = useDataStore();
	const baseURL = store.baseURL();
	const instance = axiosInstanceProxy({
		baseURL: baseURL,
		timeout: 10000
	});

	instance.interceptors.request.use(
		(config) => {
			if (config.headers) {
				config.headers['Access-Control-Allow-Origin'] = '*';
				config.headers['Access-Control-Allow-Headers'] =
					'X-Requested-With,Content-Type';
				config.headers['Access-Control-Allow-Methods'] =
					'PUT,POST,GET,DELETE,OPTIONS';
				return config;
			} else {
				return config;
			}
		},
		(error) => {
			return Promise.reject(error);
		}
	);

	return new Promise((resolve, reject) => {
		instance
			.request(config)
			.then((res) => {
				resolve(res);
			})
			.catch((e) => {
				reject(e);
			});
	});
}

export const getFormData = (object) =>
	Object.keys(object).reduce((formData, key) => {
		formData.append(key, object[key]);
		return formData;
	}, new FormData());

export async function createLibrary(name) {
	const parmas = {
		name: name,
		passwd: ''
	};
	const res = await instanceAxios({
		url: 'seahub/api2/repos/?from=web',
		method: 'post',
		data: getFormData(parmas),
		headers: { 'Content-Type': 'application/x-www-form-urlencoded' }
	});

	return res;
}

export async function fileOperate(
	path: string,
	url: string,
	parmas: { operation: string; newname?: string },
	floder: string,
	origin_id: number
) {
	const filesStore = useFilesStore();
	const res = await instanceAxios({
		url: `seahub/${url}/${
			filesStore.activeMenu(origin_id).id
		}/${floder}/?p=${encodeUrl(path)}`,
		method: 'post',
		data: getFormData(parmas),
		headers: { 'Content-Type': 'application/x-www-form-urlencoded' }
	});

	return res;
}

export async function updateFile(
	item,
	content,
	isNative = false,
	origin_id: number
) {
	const filesStore = useFilesStore();
	const res = await CommonFetch.get(
		`seahub/api2/repos/${filesStore.activeMenu(origin_id).id}/update-link/?p=/`,
		{}
	);
	const curPath = item.parentPath + item.name;

	const params = {
		target_file: curPath,
		filename: item.name,
		files_content: content
	};

	const paramsT = {};
	if (isNative) {
		paramsT['reallyContentType'] = 'multipart/form-data';
	}

	const res3 = await instanceAxios({
		url: '/seahub/seafhttp/' + res.slice(res.indexOf('seafhttp/') + 9),
		method: 'post',
		data: getFormData(params),
		headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
		params: paramsT
	});

	return res3;
}

export async function batchDeleteItem(data) {
	const res = await instanceAxios({
		url: 'seahub/api/v2.1/repos/batch-delete-item/',
		method: 'delete',
		data: data,
		headers: { 'Content-Type': 'application/json' }
	});

	return res;
}

export async function downloaFileZip(item: any, origin_id: number) {
	const store = useDataStore();
	const filesStore = useFilesStore();
	const baseURL = store.baseURL();

	const url = `/seahub/api/v2.1/repos/${
		filesStore.activeMenu(origin_id).id
	}/zip-task/`;

	const form = new FormData();
	form.append('parent_dir', item.parentPath);
	form.append('dirents', item.name);

	const res: any = await instanceAxios({
		url: url,
		method: 'post',
		data: form,
		headers: { 'Content-Type': 'application/x-www-form-urlencoded' }
	});

	const zipToken = res.data.zip_token;
	return `${baseURL}/seafhttp/zip/${zipToken}`;
}

export async function downloaFile(item: any, origin_id: number) {
	const store = useDataStore();
	const filesStore = useFilesStore();

	const baseURL = store.baseURL();

	return `${baseURL}/seahub/lib/${
		filesStore.activeMenu(origin_id).id
	}/file${encodeUrl(item.parentPath)}${encodeUrl(item.name)}?dl=1`;
}

export async function batchMoveItem(data) {
	const res = await instanceAxios({
		url: 'seahub/api/v2.1/repos/sync-batch-move-item/',
		method: 'post',
		data: data,
		headers: { 'Content-Type': 'application/json' }
	});

	return res;
}

export async function reRepoName(url, data) {
	const res = await instanceAxios({
		url: url,
		method: 'post',
		data: getFormData(data),
		headers: { 'Content-Type': 'application/x-www-form-urlencoded' }
	});

	return res;
}

export async function deleteRepo(url) {
	const res = await instanceAxios({
		url: url,
		method: 'delete',
		headers: { 'Content-Type': 'application/json' }
	});

	return res;
}

export async function batchCopyItem(data) {
	const res = await instanceAxios({
		url: 'seahub/api/v2.1/repos/sync-batch-copy-item/',
		method: 'post',
		data: data,
		headers: { 'Content-Type': 'application/json' }
	});

	return res;
}

export async function fetchRepo(
	menu: MenuItem
): Promise<SyncRepoItemType[] | SyncRepoSharedItemType[][] | undefined> {
	if (menu != MenuItem.SHAREDWITH && menu != MenuItem.MYLIBRARIES) {
		return undefined;
	}

	if (menu == MenuItem.MYLIBRARIES) {
		return fetchMineRepo();
	} else {
		const repos2 = await fetchtosharedRepo();
		const repos3 = await fetchsharedRepo();
		return [repos2, repos3];
	}
}

export async function fetchMineRepo(): Promise<SyncRepoItemType[]> {
	try {
		const res: any = await CommonFetch.get(
			'/seahub/api/v2.1/repos/?type=mine',
			{}
		);
		const repos: SyncRepoItemType[] = res.repos;
		return repos;
	} catch (error) {
		return [];
	}
}

export async function fetchtosharedRepo(): Promise<SyncRepoSharedItemType[]> {
	try {
		const res2: any = await CommonFetch.get(
			'/seahub/api/v2.1/shared-repos/',
			{}
		);
		const repos2: SyncRepoSharedItemType[] = res2;

		return repos2;
	} catch (error) {
		return [];
	}
}

export async function fetchsharedRepo(): Promise<SyncRepoSharedItemType[]> {
	try {
		const res3: any = await CommonFetch.get(
			'/seahub/api/v2.1/repos/?type=shared',
			{}
		);
		const repos3: SyncRepoSharedItemType[] = res3.repos;

		return repos3;
	} catch (error) {
		return [];
	}
}
