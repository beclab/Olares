import { date } from 'quasar';
import { MenuItem } from './contact';
import { i18n } from 'src/boot/i18n';
import { DriveType } from './../utils/interface/files';
import { common, filesIsV2 } from './../api';

export function checkSeahub(url: string) {
	// return url.startsWith('/Seahub/')
	const hasSeahub = url.indexOf('/Seahub/');
	if (hasSeahub > -1) {
		return true;
	} else {
		return false;
	}
}

export function isAppData(url: string) {
	const res = url == '/AppData/' || url == '/AppData';
	return res;
}

export function checkAppData(url: string) {
	return (
		url.startsWith('/AppData/') ||
		(url.startsWith('/cache/') && !isAppData(url))
	);
}

export function getAppDataPath(url: string) {
	const res = url.split('/');
	if (res[1] != 'AppData' && res[1] != 'Cache' && res[1] != 'cache') {
		throw Error('Invalid AppData path');
	}
	const node = res[2];
	let path = '';
	for (let i = 3; i < res.length; i++) {
		path = path + '/';
		path = path + res[i];
	}

	return { node, path };
}

export const formatFileModified = (
	modified: number | string | Date,
	format = 'YYYY-MM-DD HH:mm:ss'
) => {
	if (!modified) return '--';
	return date.formatDate(modified, format);
};

export function checkSameName(fileName: string, items: any, index = 0) {
	const hasSameName = items.findIndex((item: any) => {
		return item.name === fileName;
	});

	if (hasSameName > -1) {
		let prefix = fileName;
		if (index > 0) {
			prefix = fileName.slice(0, fileName.indexOf(`(${index})`));
		}

		const filename = `${prefix}(${index + 1})`;

		return checkSameName(filename, items, index + 1);
	} else {
		return fileName;
	}
}

export function disabledClick(path: string) {
	const disabledRightClick = ['/Cache/'];
	if (filesIsV2()) {
		disabledRightClick.push('/Files/External/');
	}
	if (disabledRightClick.includes(path)) {
		return false;
	}
	return true;
}

export function hideHeaderOpt(path: string) {
	const disabledRightClick = ['/Cache/'];

	if (filesIsV2()) {
		disabledRightClick.push('/Files/External/');
	}
	let prefix = path;
	if (!prefix.endsWith('/')) {
		prefix = prefix + '/';
	}
	if (disabledRightClick.includes(prefix)) {
		return false;
	}
	return true;
}

export function deduplicateByField(array, field) {
	return array.reduce((accumulator, current) => {
		const existingIndex = accumulator.findIndex(
			(item) => item[field] === current[field]
		);

		if (existingIndex !== -1) {
			accumulator[existingIndex] = current;
		} else {
			accumulator.push(current);
		}

		return accumulator;
	}, []);
}

export function replaceUrlHost(urlString: string, newUrlString: string) {
	const url = new URL(urlString);
	const newUrl = new URL(newUrlString);
	url.host = newUrl.host;
	return url.toString();
}

export function compareUrlHost(urlString: string, newUrlString: string) {
	const url = new URL(urlString);
	const newUrl = new URL(newUrlString);

	return url.host === newUrl.host;
}

export function translateFolderName(
	path: string,
	name: MenuItem | any,
	isMobileHeader = false
) {
	const driveType = common().formatUrltoDriveType(path) as DriveType;
	const pathParts = path.split('/');

	if (DriveType.Drive == driveType) {
		const subPath = pathParts[isMobileHeader ? 4 : 3];
		if (subPath) return name;

		return Object.values(MenuItem).includes(name)
			? i18n.global.t(`files_menu.${name}`)
			: name;
	} else if (DriveType.External === driveType) {
		const subPath = pathParts[3];
		if (subPath) return name;

		if (isMobileHeader) {
			if (Object.values(MenuItem).includes(name)) {
				return i18n.global.t(`files_menu.${name}`);
			}
		}

		return name;
	} else {
		return name;
	}
}

interface ConversionResult {
	value: number;
	unit: 'B' | 'KB' | 'MB' | 'GB' | 'TB' | 'PB';
}

export function convertBytes(bytes: number): ConversionResult {
	const units: Array<{ unit: ConversionResult['unit']; ratio: number }> = [
		{ unit: 'B', ratio: 1 },
		{ unit: 'KB', ratio: 1024 },
		{ unit: 'MB', ratio: 1024 ** 2 },
		{ unit: 'GB', ratio: 1024 ** 3 },
		{ unit: 'TB', ratio: 1024 ** 4 },
		{ unit: 'PB', ratio: 1024 ** 5 }
	];

	for (let i = units.length - 1; i >= 0; i--) {
		const { unit, ratio } = units[i];
		if (bytes >= ratio) {
			const convertedSize = bytes / ratio;
			return {
				value: Math.round(convertedSize * 100) / 100,
				unit: unit
			};
		}
	}

	return {
		value: Math.round(bytes * 100) / 100,
		unit: 'B'
	};
}

export function convertBytesString(bytes: number): string {
	const data = convertBytes(bytes);
	return data.value ? `${data.value}${data.unit}` : '';
}
