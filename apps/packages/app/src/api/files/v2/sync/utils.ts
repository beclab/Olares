import { useDataStore } from 'src/stores/data';
import { MenuItem } from 'src/utils/contact';
import { SyncRepoItemType, SyncRepoSharedItemType } from './type';
import { CommonFetch } from '../../fetch';
import { encodeUrl } from 'src/utils/encode';
import { CommonUrlApiType, commonUrlPrefix } from '../common/utils';
import { appendPath } from '../path';

export function formatUrl(url: string, repoId: string) {
	const newUrl = syncRemovePrefix(url);
	return syncCommonUrl('resources', newUrl, repoId);
}

export async function createLibrary(name: string) {
	return await CommonFetch.post(commonUrlPrefix('repos'), undefined, {
		params: {
			repoName: name
		}
	});
}

export async function renameRepo(params: {
	destination: string;
	repoId: string;
}) {
	return await CommonFetch.patch(commonUrlPrefix('repos'), undefined, {
		params
	});
}

export async function updateFile(
	path: string,
	repoId: string,
	content: string
) {
	CommonFetch.put(syncCommonUrl('resources', path, repoId), content, {
		headers: {
			'Content-Type': 'text/plain'
		}
	});
}

export async function remove(url: string, repoId: string) {
	return CommonFetch.delete(formatUrl(url, repoId));
}

export async function createDir(url: string, repoId: string) {
	return CommonFetch.post(formatUrl(url, repoId));
}

export function downloaFile(item: any, repo_id: string) {
	const store = useDataStore();

	const baseURL = store.baseURL();

	const downloadUrl =
		baseURL +
		syncCommonUrl(
			'raw',
			appendPath(
				encodeUrl(item.parentPath),
				encodeUrl(item.name),
				item.isDir ? '/' : ''
			),
			repo_id
		);

	return downloadUrl;
}

export async function deleteRepo(params: { repoId: string }) {
	return await CommonFetch.delete(commonUrlPrefix('repos'), {
		params
	});
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
		const res: any = await CommonFetch.get(commonUrlPrefix('repos'));
		const repos: SyncRepoItemType[] = res.repos;
		return repos;
	} catch (error) {
		return [];
	}
}

export async function fetchtosharedRepo(): Promise<SyncRepoSharedItemType[]> {
	try {
		const res2: any = await CommonFetch.get(commonUrlPrefix('repos'), {
			params: {
				type: 'share_to_me'
			}
		});
		if (res2 && res2.repos) {
			const repos: SyncRepoSharedItemType[] = res2.repos;
			return repos;
		}
		return [];
	} catch (error) {
		return [];
	}
}

export async function fetchsharedRepo(): Promise<SyncRepoSharedItemType[]> {
	try {
		const res3: any = await CommonFetch.get(commonUrlPrefix('repos'), {
			params: {
				type: 'shared'
			}
		});
		if (res3 && res3.repos) {
			const repos: SyncRepoSharedItemType[] = res3.repos;
			return repos;
		}
		return [];
	} catch (error) {
		return [];
	}
}

export function syncRemovePrefix(url: string) {
	url = syncRemoveHomePrefix(url);
	return url;
}

export function syncRemoveHomePrefix(url: string) {
	if (!url.startsWith('/sync/')) {
		return url;
	}
	return url.slice(5);
}

export const syncCommonUrl = (
	type: CommonUrlApiType,
	path: string,
	repoId: string
) => {
	return commonUrlPrefix(type) + appendPath('sync', repoId, path);
};

export const formatPathtoUrl = (path: string, repoId: string) => {
	path = syncRemovePrefix(path);
	return appendPath('/sync', repoId, path);
};

export const displayPath = (
	file: {
		isDir: boolean;
		fileType?: string;
		parent_dir: string;
		name: string;
	},
	repoName: string
) => {
	return appendPath(
		'/Seahub/',
		repoName,
		encodeUrl(file.parent_dir),
		file.isDir ? file.name : '',
		'/'
	);
};
