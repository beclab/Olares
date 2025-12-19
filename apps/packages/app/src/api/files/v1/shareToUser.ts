import { fetchURL } from './utils';
import { useDataStore } from '../../../stores/data';
import { useMenuStore } from '../../../stores/files-menu';
import { axiosInstanceProxy } from '../../../platform/httpProxy';

// eslint-disable-next-line @typescript-eslint/no-unused-vars
export async function listSharedItems() {
	const menuStore = useMenuStore();

	const res = await fetchURL(
		`/seahub/api2/repos/${menuStore.shareRepoInfo.id}/dir/shared_items/?p=${
			menuStore.shareRepoInfo.path || '/'
		}&share_type=user`
	);
	return res.data;
}

export async function getUserList() {
	try {
		const res = await fetchURL(
			'/seahub/api/v2.1/admin/users/?page=1&per_page=1000'
		);
		return res.data.data;
	} catch (error) {
		console.error('errrr', error);
	}
}

export async function setSharedItems(username, primary) {
	const store = useDataStore();
	const menuStore = useMenuStore();
	const baseURL = store.baseURL();

	const instance = axiosInstanceProxy({
		baseURL: baseURL,
		timeout: 10000,
		headers: {
			'Access-Control-Allow-Origin': '*',
			'Access-Control-Allow-Headers': 'X-Requested-With,Content-Type',
			'Access-Control-Allow-Methods': 'PUT,POST,GET,DELETE,OPTIONS',
			'Content-Type': 'application/x-www-form-urlencoded'
		}
	});

	const form = new FormData();
	form.append('share_type', 'user');
	form.append('permission', primary);
	for (let i = 0; i < username.length; i++) {
		const element = username[i];
		form.append('username', element);
	}

	const res = await instance.put(
		`seahub/api2/repos/${menuStore.shareRepoInfo.id}/dir/shared_items/?p=${
			menuStore.shareRepoInfo.path || '/'
		}`,
		form,
		{}
	);

	return res;
}

export async function deleteItem(username) {
	const store = useDataStore();
	const menuStore = useMenuStore();
	const baseURL = store.baseURL();

	const instance = axiosInstanceProxy({
		baseURL: baseURL,
		timeout: 10000,
		headers: {
			'Access-Control-Allow-Origin': '*',
			'Access-Control-Allow-Headers': 'X-Requested-With,Content-Type',
			'Access-Control-Allow-Methods': 'PUT,POST,GET,DELETE,OPTIONS'
		}
	});

	const res = await instance.delete(
		`seahub/api2/repos/${menuStore.shareRepoInfo.id}/dir/shared_items/?p=${
			menuStore.shareRepoInfo.path || '/'
		}&share_type=user&username=${username}`
	);

	return res;
}

export async function updateItem(username, primary) {
	const store = useDataStore();
	const menuStore = useMenuStore();
	const baseURL = store.baseURL();

	const instance = axiosInstanceProxy({
		baseURL: baseURL,
		timeout: 10000,
		headers: {
			'Access-Control-Allow-Origin': '*',
			'Access-Control-Allow-Headers': 'X-Requested-With,Content-Type',
			'Access-Control-Allow-Methods': 'PUT,POST,GET,DELETE,OPTIONS'
		}
	});

	const form = new FormData();
	form.append('permission', primary);

	const res = await instance.post(
		`seahub/api2/repos/${menuStore.shareRepoInfo.id}/dir/shared_items/?p=/&share_type=user&username=${username}`,
		form
	);
	return res;
}
