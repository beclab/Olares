import { getFileTypeI18nKey } from '@bytetrade/core';
import { useFilesStore } from 'src/stores/files';
import { getextension } from 'src/utils/utils';
import { i18n } from 'src/boot/i18n';
import { DriveType } from 'src/utils/interface/files';
import { displayPath } from './utils';

export function formatData(data, url, origin_id) {
	const filesStore = useFilesStore();
	data.origin = DriveType.Common;
	data.driveType = DriveType.Common;
	data.path = displayPath(data);

	let curItems = data.items || [];
	if (filesStore.filterHiddenDir[origin_id]) {
		curItems = curItems.filter((item) => !item.name.startsWith('.'));
	}

	curItems.map((el, index) => {
		const extension = getextension(el.name);
		el.index = index;

		el.oPath = el.path;
		el.oParentPath = el.path.substring(
			0,
			el.path.length - el.name.length - (el.path.endsWith('/') ? 1 : 0)
		);
		el.path = displayPath(el);
		el.driveType = DriveType.Common;
		el.extension = extension;
		el.modified = new Date(el.modified).getTime();
		el.type = el.isDir
			? i18n.global.t('files.folders')
			: i18n.global.t(getFileTypeI18nKey(el.name));
	});

	data.items = curItems;

	return data;
}
