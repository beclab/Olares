import { MenuItem } from 'src/utils/contact';
import { getParams } from 'src/utils/utils';
import {
	DriveType,
	supportDriveTypes,
	ActiveMenuType
} from 'src/utils/interface/files';
import { useFilesStore } from 'src/stores/files';
import { appendPath } from '../path';
import { useDataStore } from 'src/stores/data';
import { FileType } from './utils';
import { getApplication } from 'src/application/base';
import { useShareStore } from 'src/stores/share/share';

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
	} else if (href.startsWith('/Share')) {
		if (getApplication().applicationName == 'share') {
			return DriveType.PublicShare;
		}
		return DriveType.Share;
	} else if (href.startsWith('/sharable-link')) {
		return DriveType.PublicShare;
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
	} else if (href.startsWith('/Share')) {
		return {
			label: MenuItem.SHARE,
			id: MenuItem.SHARE,
			driveType: DriveType.Share
		};
	}
	// else if (href.startsWith('/ShareWith')) {
	// 	return {
	// 		label: MenuItem.SHAREWITHME,
	// 		id: MenuItem.SHAREWITHME,
	// 		driveType: DriveType.ShareWithMe
	// 	};
	// }
	else {
		const label = decodeURIComponent(href).split('/')[2];
		return {
			label: label,
			id: label,
			driveType: DriveType.Drive
		};
	}
}

export function filterPcvPath(path: string, position = 1): string {
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

export const driveTypeBySearchPath = (url: string): DriveType | undefined => {
	if (!url) {
		return undefined;
	}
	const lowerUrl = url.toLowerCase();
	if (lowerUrl.startsWith('drive') || lowerUrl.startsWith('/drive')) {
		return DriveType.Drive;
	} else if (lowerUrl.startsWith('hdd') || lowerUrl.startsWith('/hdd')) {
		return DriveType.External;
	} else if (lowerUrl.startsWith('data') || lowerUrl.startsWith('/data')) {
		return DriveType.Data;
	} else if (lowerUrl.startsWith('cache') || lowerUrl.startsWith('/cache')) {
		return DriveType.Cache;
	} else {
		return undefined;
	}
};

export const driveTypeBySearchPathV2 = (path: string): DriveType => {
	if (!path) {
		return DriveType.Drive;
	}

	if (path.startsWith('drive/Home')) {
		return DriveType.Drive;
	} else if (path.startsWith('drive/Data')) {
		return DriveType.Data;
	} else if (path.startsWith('cache/')) {
		return DriveType.Cache;
	} else if (path.startsWith('share')) {
		return DriveType.Share;
	}
	return DriveType.Drive;
};

export const driveTypeByFileTypeAndFileExtend = (
	fileType: string,
	fileExtend: string
) => {
	if (!supportDriveTypes.includes(fileType as DriveType)) {
		return DriveType.Drive;
	}

	if (fileType == DriveType.Drive) {
		if (fileExtend == 'Data') {
			return DriveType.Data;
		}
		return DriveType.Drive;
	}
	return fileType as DriveType;
};

export const driveTypeBySearchMeta = (meta?: {
	fileType?: FileType;
	fileExtend?: string;
}): DriveType | undefined => {
	if (!meta || !meta.fileType || !meta.fileExtend) {
		return DriveType.Drive;
	}

	if (meta.fileType == 'drive') {
		if (meta.fileExtend == 'Home') {
			return DriveType.Drive;
		} else if (meta.fileExtend == 'Data') {
			return DriveType.Data;
		}
	} else if (meta.fileType == 'external') {
		return DriveType.External;
	} else if (meta.fileType == 'cache') {
		return DriveType.Cache;
	}

	return DriveType.Drive;
};

export const displayConnectServer = (path: string) => {
	const filesStore = useFilesStore();
	return filesStore.nodes
		.map((e) => {
			return appendPath('/Files/External/', e.name, '/');
		})
		.includes(path);
};

export const isShareRootPage = (path: string) => {
	// return ['/ShareWith/', '/ShareBy/'].includes(path);
	return ['/Share/'].includes(path);
};

export const videoPlayUrl = (raw: string, node: string) => {
	const dataStore = useDataStore();

	if (getApplication().applicationName == 'share') {
		const shareStore = useShareStore();
		console.log('shareStore.share ===>', shareStore.share);

		node = shareStore.share?.node || '';
	}

	return `${dataStore.baseURL()}/videos/${
		node && node.length ? node + '/' : ''
	}?PlayPath=${raw}`;
};
