import type { RouteLocationNormalizedLoaded } from 'vue-router';
import { common } from 'src/api';
import { useFilesStore, FilesIdType } from 'src/stores/files';
import { DriveType } from 'src/utils/interface/files';

/**
 * Mobile files layout: use `<router-view />` for FileRootPage / FilesRepoPage
 * instead of embedding `FilesPage` directly.
 * Covers `/Files/`, `/files`, `/file` (see routes-files.ts, routes-mobile-common.ts) and `/repo/...`.
 */
export function isFilesMobileShellRoutePath(path: string): boolean {
	if (path.startsWith('/repo')) {
		return true;
	}
	const trimmed = path.replace(/\/+$/, '');
	const base = trimmed.toLowerCase();
	return base === '/files' || base === '/file';
}

/**
 * Sync files store from current route (shared by LayoutMobile / LayoutPc).
 */
export async function initFilesLayoutFromRoute(
	route: Pick<RouteLocationNormalizedLoaded, 'fullPath'>,
	filesStore = useFilesStore()
) {
	let url = route.fullPath;
	let driveType = common().formatUrltoDriveType(url);

	if (driveType === undefined) {
		url = '/Files/Home/';
		driveType = DriveType.Drive;
	} else if (driveType == DriveType.Cache || driveType == DriveType.External) {
		if (driveType == DriveType.Cache && url !== '/Cache/') {
			filesStore.currentNode[FilesIdType.PAGEID] = {
				name: url.split('/')[2],
				master: false
			};
		} else if (driveType == DriveType.External && url !== '/Files/External/') {
			filesStore.currentNode[FilesIdType.PAGEID] = {
				name: url.split('/')[3],
				master: false
			};
		}
	}

	return filesStore.setBrowserUrl(url, driveType);
}
