import { useFilesStore, FileItem } from 'src/stores/files';
import { getextension } from 'src/utils/utils';
import { encodeUrl } from 'src/utils/encode';
import { DriveType } from 'src/utils/interface/files';
import { appendPath } from '../path';
import { displayPath } from './utils';

export function formatAppData(
	node: string,
	data: any,
	url: string,
	origin_id: number
) {
	const filesStore = useFilesStore();
	data.origin = DriveType.Cache;
	data.driveType = DriveType.Cache;
	data.url = `/Files${url}`;
	data.path = displayPath(data);
	if (data.isDir) {
		if (!data.url.endsWith('/')) data.url += '/';
	}

	let curItems: FileItem[] = data.items || [];
	if (filesStore.filterHiddenDir[origin_id]) {
		curItems = data.items.filter((item) => !item.name.startsWith('.'));
	}

	curItems.map((el, index) => {
		const extension = getextension(el.name);
		el.index = index;
		el.driveType = DriveType.Cache;
		el.extension = extension;
		el.modified = new Date(el.modified).getTime();

		el.oPath = el.path;
		el.oParentPath = el.path.substring(
			0,
			el.path.length - el.name.length - (el.path.endsWith('/') ? 1 : 0)
		);
		el.path = displayPath(el);
	});

	data.items = curItems;

	return data;
}
