import { MessageBody } from '@bytetrade/core';
import { RouteLocationNormalizedLoaded } from 'vue-router';
import { common, dataAPIs } from './../api';
import { useFilesStore, FilePath } from './../stores/files';
import { DriveType } from '../utils/interface/files';

export const notifyFiles = (
	message: MessageBody | string,
	route: RouteLocationNormalizedLoaded,
	id: number
) => {
	const filesStore = useFilesStore();
	const body: MessageBody =
		typeof message == 'string' ? JSON.parse(message) : message;

	if (body.event == 'filesUpdate') {
		if (body.message?.data?.code === 0) {
			let path = body.message?.data?.msg;

			path = path.endsWith('/') ? path.slice(0, -1) : path;

			if (path.split('/')[path.split('/').length - 1].startsWith('.')) {
				if (filesStore.filterHiddenDir[id]) return false;
			}
			if (path.startsWith('/data')) {
				path = path.replace('/data', '/Files');
			}
			const dataAPI = dataAPIs(path);
			const cur_path = dataAPI.getPurePath(path);
			const driveType = common().formatUrltoDriveType(path) || DriveType.Drive;
			const key = filesStore.registerUniqueKey(cur_path, driveType, '');

			if (driveType === DriveType.Drive) {
				if (key in filesStore.cached) {
					const split_cur_path = cur_path.split('?');
					const path = new FilePath({
						path: split_cur_path[0],
						param: split_cur_path[1] ? `?${split_cur_path[1]}` : '',
						isDir: true,
						driveType: driveType
					});

					if (cur_path === route.path) {
						filesStore.setFilePath(path, false, false);
					} else {
						filesStore.fetchData(path, key, cur_path);
					}
				}
			}
		}
	}
};
