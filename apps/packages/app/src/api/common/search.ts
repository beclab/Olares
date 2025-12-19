import { getFileIcon, getFileType } from '@bytetrade/core';
import axios from 'axios';
import { useUserStore } from 'src/stores/user';
import {
	ServiceType,
	TextSearchItem,
	SearchV2Type
} from 'src/utils/interface/search';

import { dataAPIs, filesIsV1, common } from 'src/api/files';
import { driveTypeBySearchPathV2 } from '../files/v2/common/common';
import { dataAPIs as dataAPIsV2 } from 'src/api/files/v2';
import { getApplication } from 'src/application/base';
import { appendPath } from '../files/v2/path';
import { driveTypeByFileTypeAndFileExtend } from 'src/api/files/v2/common/common';
import { DriveType } from 'src/utils/interface/files';

export const seachOSVersionLargeThan12_3 = () => {
	if (getApplication().applicationName === 'desktop') {
		return true;
	}
	const userStore = useUserStore();
	return userStore.current_user?.isLargeVersion12_3 || false;
};

export async function searchInit(
	body: {
		reqid: string;
		keyword: string;
		type: SearchV2Type;
		app: ServiceType;
	},
	offset = 0,
	limit = 20
): Promise<TextSearchItem[]> {
	const res: TextSearchItem[] = await axios.post(
		searchBaseUrl() + '/api/search/init',
		{
			...body,
			offset,
			limit
		}
	);

	return paramSearchResult(res);
}

export const searchMore = async (
	body: {
		reqid: string;
	},
	offset = 0,
	limit = 20
): Promise<TextSearchItem[]> => {
	const res: TextSearchItem[] = await axios.post(
		searchBaseUrl() + '/api/search/more',
		{
			...body,
			offset,
			limit
		}
	);
	return paramSearchResult(res);
};

export const searchCancel = async (reqid: string): Promise<void> => {
	await axios.post(searchBaseUrl() + '/api/search/cancel', { reqid });
};

const paramSearchResult = (res: TextSearchItem[]) => {
	const newRes: TextSearchItem[] = [];

	for (let i = 0; i < res.length; i++) {
		const el = res[i];
		el.fileType = getFileType(el.title);
		el.fileIcon = getFileIcon(el.title);
		if (el.resource_uri) {
			if (filesIsV1()) {
				el.driveType = common().driveTypeBySearchPath(el.resource_uri);
			} else {
				el.driveType = driveTypeBySearchPathV2(el.resource_uri);
			}
			if (el.driveType) {
				if (filesIsV1()) {
					el.isDir = el.resource_uri.endsWith('/');
					const path = dataAPIs(el.driveType).formatSearchPath(el.resource_uri);
					el.path = path;
				} else {
					const meta = dataAPIsV2(el.driveType).pathToFrontendFile(
						appendPath('/', el.resource_uri)
					);
					const path = dataAPIsV2(el.driveType).displayPath({
						isDir: meta.isDir,
						fileExtend: meta.fileExtend,
						path: meta.path,
						fileType: ''
					});
					el.isDir = meta.isDir;
					el.path = path;
				}
			} else {
				el.path = el.resource_uri;
			}
		}

		newRes.push(el);
	}
	return newRes;
};

export const syncSearch = async (
	body: {
		query: string;
	},
	offset = 0,
	limit = 20
) => {
	const res: any = await axios.post(searchBaseUrl() + '/api/search/sync', {
		...body,
		offset,
		limit
	});

	if (res && res.length > 0) {
		const resArr: TextSearchItem[] = [];
		for (let i = 0; i < res.length; i++) {
			const el = res[i];
			const id = `id_${i}`;
			const item = syncSearchItemFormat(id, el, body.query);
			if (item) resArr.push(item);
		}
		return resArr;
	}
	return [];
};

const syncSearchItemFormat = (id: string, data: any, query: string) => {
	const fileType = getFileType(data.title) || 'blob';
	const fileIcon = getFileIcon(data.title) || 'other';

	const driveType = driveTypeByFileTypeAndFileExtend(
		data.file_type,
		data.file_extend
	);

	const highlight = [data.title.replace(query, `<hi>${query}</hi>`)];
	const highlight_field = ['title'];

	let path = data.path;
	const isDir = data.type === 'file' ? false : true;
	if (driveType == DriveType.Sync) {
		path =
			appendPath('/Seahub', data.repo_name, data.path, isDir ? '/' : '') +
			`?id=${data.file_extend}`;
	} else if (driveType == DriveType.Share) {
		path = dataAPIsV2(DriveType.Share).displayPath({
			isDir: isDir,
			fileExtend: data.file_extend,
			path: data.path,
			fileType: ''
		});
	}
	return {
		id: id,
		highlight,
		highlight_field,
		title: data.title,
		fileType: fileType,
		fileIcon: fileIcon || 'other',
		repo_name: data.repo_name,
		path: path,
		isDir: isDir,
		driveType
	};
};

const searchBaseUrl = () => {
	const userStore = useUserStore();
	let url = userStore.getModuleSever('desktop');
	if (process.env.IS_DEV) {
		url = '';
	}
	return url;
};
