import { MenuItem } from 'src/utils/contact';
import { getParams } from 'src/utils/utils';
import { DriveType, ActiveMenuType } from 'src/utils/interface/files';
import { useDataStore } from 'src/stores/data';

export function formatUrltoDriveType(href: string): DriveType | undefined {
	if (!href) {
		return undefined;
	}
	if (href.startsWith('/Files')) {
		if (href.startsWith('/Files/Externa')) {
			return DriveType.External;
		}
		return DriveType.Drive;
	} else if (href.startsWith('/Seahub') || href.startsWith('/repo')) {
		return DriveType.Sync;
	} else if (href.startsWith('/Data')) {
		return DriveType.Data;
	} else if (href.startsWith('/Cache')) {
		return DriveType.Cache;
	} else if (href.startsWith('/Drive/google')) {
		return DriveType.GoogleDrive;
	} else if (href.startsWith('/Drive/dropbox')) {
		return DriveType.Dropbox;
	} else if (href.startsWith('/Drive/awss3')) {
		return DriveType.Awss3;
	} else if (href.startsWith('/Drive/tencent')) {
		return DriveType.Tencent;
	} else {
		return undefined;
	}
}

export function formatUrltoActiveMenu(href: string): ActiveMenuType {
	if (href.startsWith('/Files/Home')) {
		const label = decodeURIComponent(href).split('/')[3] || MenuItem.HOME;

		const isHome = href === '/Files/Home/';
		return {
			label: isHome ? MenuItem.HOME : label,
			id: isHome ? MenuItem.HOME : label,
			driveType: DriveType.Drive
		};
	} else if (href.startsWith('/Files/External')) {
		const label = decodeURIComponent(href).split('/')[2];
		return {
			label: label,
			id: label,
			driveType: DriveType.External
		};
	} else if (href.startsWith('/Seahub')) {
		const label = decodeURIComponent(href).split('/')[2];
		const splitUrl = href.split('?');
		const repo_id = getParams(splitUrl.length > 1 ? splitUrl[1] : href, 'id');

		return {
			label: label,
			id: repo_id,
			driveType: DriveType.Sync,
			params: '?' + href.split('?')[1]
		};
	} else if (href.startsWith('/Data')) {
		// console.log(label);
		return {
			label: MenuItem.DATA,
			id: MenuItem.DATA,
			driveType: DriveType.Data
		};
	} else if (href.startsWith('/Cache')) {
		return {
			label: MenuItem.CACHE,
			id: MenuItem.CACHE,
			driveType: DriveType.Cache
		};
	} else if (href.startsWith('/Drive/google')) {
		const splitHref = href.split('/')[3];
		return {
			label: splitHref,
			id: splitHref,
			driveType: DriveType.GoogleDrive
		};
	} else if (href.startsWith('/Drive/dropbox')) {
		const splitHref = href.split('/')[3];
		return {
			label: splitHref,
			id: splitHref,
			driveType: DriveType.Dropbox
		};
	} else if (href.startsWith('/Drive/awss3')) {
		const splitHref = href.split('/')[3];
		return {
			label: splitHref,
			id: splitHref,
			driveType: DriveType.Awss3
		};
	} else if (href.startsWith('/Drive/tencent')) {
		const splitHref = href.split('/')[3];
		return {
			label: splitHref,
			id: splitHref,
			driveType: DriveType.Tencent
		};
	} else {
		const label = decodeURIComponent(href).split('/')[2];
		return {
			label: label,
			id: label,
			driveType: DriveType.Drive
		};
	}
}

export function filterPcvPath(path: string): string {
	const splitPath = path.split('/');
	const newPathArr: string[] = [];
	for (let i = 0; i < splitPath.length; i++) {
		const path_1 = splitPath[i];
		if (path_1.indexOf('pvc-') <= -1) {
			newPathArr.push(path_1);
		}
	}
	return newPathArr.join('/');
}

export function filterPcvPath2(path: string, position = 1): string {
	const splitPath = path.split('/');
	if (
		splitPath &&
		splitPath.length > position + 1 &&
		splitPath[position].indexOf('pvc-') > -1
	) {
		splitPath.splice(position, 1);
		return splitPath.join('/');
	} else {
		return path;
	}
}

export const displayConnectServer = (path: string) => {
	return path == '/Files/External/';
};

export const videoPlayUrl = (raw: string, node: string) => {
	const dataStore = useDataStore();
	return `${dataStore.baseURL()}/videos/play?PlayPath=${raw}`;
};

export const driveTypeBySearchPath = (url: string): DriveType | undefined => {
	if (!url) {
		return undefined;
	}
	// const lowerUrl = url.toLowerCase();
	// if (lowerUrl.startsWith('drive') || lowerUrl.startsWith('/drive')) {
	// 	return DriveType.Drive;
	// } else if (lowerUrl.startsWith('hdd') || lowerUrl.startsWith('/hdd')) {
	// 	return DriveType.External;
	// } else if (lowerUrl.startsWith('data') || lowerUrl.startsWith('/data')) {
	// 	return DriveType.Data;
	// } else if (lowerUrl.startsWith('cache') || lowerUrl.startsWith('/cache')) {
	// 	return DriveType.Cache;
	// } else {
	// 	return undefined;
	// }
	return DriveType.Drive;
};

export const isShareRootPage = (path: string) => {
	return ['/Share/'].includes(path);
};
