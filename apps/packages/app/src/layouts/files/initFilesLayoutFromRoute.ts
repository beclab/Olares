import type { RouteLocationNormalizedLoaded } from 'vue-router';
import { common } from 'src/api';
import { useFilesStore, FilesIdType } from 'src/stores/files';
import { DriveType } from 'src/utils/interface/files';
import { canonicalizeFilesPath } from 'src/utils/canonicalizeFilesPath';

/**
 * Mobile files layout: use `<router-view />` for FileRootPage / FilesRepoPage
 * instead of embedding `FilesPage` directly.
 * Covers `/Files/`, `/files`, `/file` (see routes-files.ts, routes-mobile-common.ts) and `/repo/...`.
 */
export function isFilesMobileShellRoutePath(path: string): boolean {
	const trimmed = path.replace(/\/+$/, '');
	const normalizedPath = trimmed.toLowerCase();
	if (normalizedPath.startsWith('/repo')) {
		return true;
	}
	if (normalizedPath.startsWith('/transfer')) {
		return true;
	}
	return normalizedPath === '/files' || normalizedPath === '/file';
}

/**
 * Whether the files store already reflects the given route full path.
 * Used to tell apart in-app navigation (store updated before the route) from
 * system back/forward navigation (only the route changed via popstate).
 */
export function isRouteSyncedToFilesStore(
	fullPath: string,
	filesStore = useFilesStore()
): boolean {
	const current = filesStore.currentPath[FilesIdType.PAGEID];
	if (!current || !current.path) {
		return false;
	}
	const storeFullPath = `${current.path}${current.param || ''}`;
	return (
		canonicalizeFilesPath(storeFullPath) === canonicalizeFilesPath(fullPath)
	);
}

/**
 * Sync files store from current route (shared by LayoutMobile / LayoutPc).
 *
 * `isPrev`/`skipRouter` let callers re-sync the store from the route without
 * pushing a new back-stack entry or triggering another router navigation,
 * which is required when reacting to system back/forward (popstate).
 *
 * `fromHistory` additionally reconciles `backStack` with the restored location
 * (see `syncBrowserHistoryUrl`) so the in-app back button stays consistent
 * after a browser back/forward.
 */
export async function initFilesLayoutFromRoute(
	route: Pick<RouteLocationNormalizedLoaded, 'fullPath'>,
	filesStore = useFilesStore(),
	options: {
		isPrev?: boolean;
		skipRouter?: boolean;
		fromHistory?: boolean;
	} = {}
) {
	const { isPrev = true, skipRouter = false, fromHistory = false } = options;
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

	if (fromHistory) {
		return filesStore.syncBrowserHistoryUrl(url, driveType, FilesIdType.PAGEID);
	}

	return filesStore.setBrowserUrl(
		url,
		driveType,
		isPrev,
		FilesIdType.PAGEID,
		skipRouter
	);
}
