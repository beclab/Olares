import { defineStore } from 'pinia';
import {
	getDownloadHistory,
	DownloadRecordListRequest
} from 'src/api/wise/download';
import {
	DRIVER_FILE_PREFIX,
	getFileTypeByName,
	Enclosure
} from 'src/utils/rss-types';
import { uploadCreateEntry } from 'src/api/wise/upload';
import { useReaderStore } from './rss-reader';
import { fetchNodeList } from 'src/api/files/v2/common/utils';
import { FileNode } from 'src/stores/files';
import { DriveAPI } from 'src/api/files/v2';
import { DriveType } from 'src/utils/interface/files';
import { encodeUrl } from 'src/utils/encode';

export type DataState = {
	enclosureMaps: Record<string, number>;
	nodes: FileNode[];
};

export const useTransferStore = defineStore('rssTransfer', {
	state: () => {
		return {
			enclosureMaps: {},
			nodes: [] as FileNode[]
		} as unknown as DataState;
	},
	actions: {
		async init() {
			if (this.nodes.length == 0) {
				this.nodes = await fetchNodeList();
			}
		},
		getVideoUrl(path, type) {
			const driver = DriveAPI.getAPI(DriveType.Drive);
			return encodeUrl(
				driver.getPreviewURL(
					{
						type,
						path: path,
						modify: true
					},
					'big'
				)
			);
		},
		getDownloadUrl(path = '') {
			const readerStore = useReaderStore();
			let wisePath = '';
			if (path) {
				wisePath = path;
			} else if (readerStore.readingEntry) {
				wisePath = readerStore.readingEntry.local_file_path;
			}
			const driver = DriveAPI.getAPI(DriveType.Drive);
			return driver.getDownloadURL({ path: wisePath }, true);
		},
		async uploadFile(path: string, name: string): Promise<any> {
			const fileType = getFileTypeByName(name);
			let wisePath = path;
			if (wisePath.startsWith(DRIVER_FILE_PREFIX)) {
				wisePath = path.substring(DRIVER_FILE_PREFIX.length);
			}
			return await uploadCreateEntry(wisePath, name, fileType);
		},
		async addEnclosureTasks(array: Enclosure[]) {
			for (let i = 0; i < array.length; i++) {
				const enclosure = array[i];
				if (this.enclosureMaps[enclosure.id]) {
					break;
				}
				getDownloadHistory(
					new DownloadRecordListRequest(undefined, undefined, enclosure.id)
				).then((records) => {
					if (records && records.length > 0 && records[0].enclosure_id) {
						this.enclosureMaps[enclosure.id] = records[0].id;
					}
				});
			}
		}
	}
});
