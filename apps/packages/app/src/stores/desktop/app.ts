import { defineStore } from 'pinia';
import axios from 'axios';
import {
	DesktopAppInfo,
	DesktopPosition,
	DockerAppInfo
} from '@apps/desktop/type/types';
import { TerminusApp } from '@bytetrade/core';
import { AppStatusChange } from 'src/constant/constants';
import { useTokenStore } from './token';
import { AppInfo } from 'src/utils/interface/applications';
import { SearchCategory } from 'src/utils/interface/search';
import { uninstalledAppState } from 'src/constant/config';
import { MessageQueue } from 'src/utils/desktop/utils';
import { i18n } from 'src/boot/i18n';
import {
	Bookmark,
	BOOKMARK_RAW_PREFIX,
	DockItem,
	DockLayout,
	InitLayout,
	LAYOUT_SCHEMA_VERSION,
	LaunchLayout,
	LaunchpadItem,
	bookmarkIdFromRawId,
	encodeKey,
	isBookmarkRawId,
	keyToRawId,
	parseKey,
	rawIdToKey
} from 'src/types/desktop-layout';
import { useBookmarkStore } from './bookmark';
import { bookmarkAvatarDataUri } from 'src/utils/desktop/bookmarkAvatar';
import {
	getDockLayout as apiGetDock,
	getLaunchLayout as apiGetLaunch,
	orphanCleanup as apiOrphanCleanup,
	putDockLayout as apiPutDock,
	putLaunchLayout as apiPutLaunch
} from 'src/api/desktop-layout';
import {
	Debounced,
	debounce,
	genClientOpId,
	getDeviceId
} from 'src/utils/desktop-debounce';

export interface Service {
	name: string;
	status: string;
	url: string;
}

export const systemApps = [
	'settings',
	'market',
	'dashboard',
	'control-hub',
	'files'
];

export const systemI18nApps = (app: TerminusApp) => {
	if (!systemApps.includes(app.name)) {
		return app.title;
	}
	return i18n.global.t(`system_app.${app.title}`);
};

export const getAppDisplayTitle = (
	app: { name: string; title: string; isSysApp?: boolean },
	t: (key: string) => string
) => {
	return systemApps.includes(app.name) && app.isSysApp
		? t(`system_app.${app.title}`)
		: app.title;
};

export type AppState = {
	DOCKER_APP_SIZE: number;
	DOCKER_APP_GAP: number;
	DOCKER_APP_START_GAP: number;
	DOCKER_APP_END_GAP: number;

	DESKTOP_APP_SIZE: number;
	DESKTOP_APP_X_NUM: number;
	DESKTOP_APP_Y_NUM: number;
	DESKTOP_APP_X_GAP: number;
	DESKTOP_APP_Y_GAP: number;
	DOCKER_APP_START_X_GAP: number;
	DOCKER_APP_END_X_GAP: number;
	DOCKER_APP_START_Y_GAP: number;
	DOCKER_APP_END_Y_GAP: number;

	profile_id: string | undefined;

	dockerApps: DockerAppInfo[];
	launchPadApps: number[][];

	launchpadGridOverride: { x: number; y: number } | null;

	launchpadSearchQuery: string;

	launchpadShakeMode: boolean;

	myApps: AppInfo[];
	myAppsInitialized: boolean;
	desktopApps: DesktopAppInfo[];

	axiosAppInfo: TerminusApp[];

	appMessageQueue: MessageQueue<{
		name: string;
		payload: AppStatusChange;
		isdequeued: boolean;
	}>;

	// Cross-device layout sync. A single shared layout per surface (no
	// form-factor split): PC and mobile read the same linear order and reflow
	// it onto their own grid at render time. Populated by GET on boot, mutated
	// by applyRemote* on WS events, and PUT through the saveDebounced helpers.
	layoutBySurface: {
		launchpad?: LaunchLayout;
		dock?: DockLayout;
	};
	syncedAt: { [k: string]: number };
	pendingOpIds: Set<string>;

	isLayoutDragLocked: boolean;
	pendingDockLayout: { layout: DockLayout; opIds: string[] } | null;
	pendingLaunchLayout: { layout: LaunchLayout; opIds: string[] } | null;
};

const socketMessageTimer: NodeJS.Timeout | null = null;

export const localStoreKeys = {
	LAUNCH_PAD: 'lanuchPad',
	DOCK_PAD: 'dockerApps',
	LAUNCH_LAYOUT_CACHE: 'launch_layout_cache',
	DOCK_LAYOUT_CACHE: 'dock_layout_cache',
	LAYOUT_MIGRATED: 'desktop_layout_migrated_v1'
};

const DEBOUNCE_WAIT_MS = 800;

// Debounce holders are stored outside the Pinia state (Set/Map don't survive
// reactivity well and we don't want to serialize them). A single shared
// layout per surface means one holder each.
let launchDebounce: Debounced<[LaunchLayout]> | null = null;
let dockDebounce: Debounced<[DockLayout]> | null = null;

export const useApplicationStore = defineStore('desktopApp', {
	state: () => {
		return {
			DOCKER_APP_SIZE: 36,
			DOCKER_APP_GAP: 10,
			DOCKER_APP_START_GAP: 12,
			DOCKER_APP_END_GAP: 12,

			DESKTOP_APP_SIZE: 70,
			DESKTOP_APP_X_NUM: 7,
			DESKTOP_APP_Y_NUM: 4,
			DESKTOP_APP_X_GAP: 120,
			DESKTOP_APP_Y_GAP: 75,
			DOCKER_APP_START_X_GAP: 160,
			DOCKER_APP_END_X_GAP: 100,
			DOCKER_APP_START_Y_GAP: 50,
			DOCKER_APP_END_Y_GAP: 100,

			profile_id: undefined,

			dockerApps: [],
			launchPadApps: [[]],
			launchpadGridOverride: null,
			launchpadSearchQuery: '',
			launchpadShakeMode: false,
			desktopApps: [],
			myApps: [],
			myAppsInitialized: false,
			axiosAppInfo: [],
			appMessageQueue: new MessageQueue(),
			layoutBySurface: {},
			syncedAt: {},
			pendingOpIds: new Set<string>(),
			isLayoutDragLocked: false,
			pendingDockLayout: null,
			pendingLaunchLayout: null
		} as AppState;
	},
	actions: {
		async get_default_dock_list(isMobile: boolean) {
			let index = 0;
			const result: DockerAppInfo[] = [];
			// Shared default across PC and mobile (the dock is now a single
			// synced layout). `isMobile` only affects the launchpad icon asset.
			const dock_app_lists = ['files', 'launchpad', 'market', 'settings'];

			for (const dock_app of dock_app_lists) {
				let curApp: any = null;
				if (dock_app == 'launchpad') {
					curApp = {
						id: 'launchpad',
						appid: 'launchpad',
						icon: isMobile
							? './app-icon/launch-icon.png'
							: 'https://app.cdn.olares.com/appstore/launchpad/icon.png',
						name: 'Launchpad',
						title: '',
						target: '',
						state: 'running',
						fatherName: null,
						openMethod: isMobile ? 'window' : 'default'
					};
				} else {
					const dock = this.myApps.find((app) => app.id.startsWith(dock_app));
					if (!dock) {
						continue;
					}

					curApp = this.myApps.find((app) => app.id == dock.id);
				}

				if (curApp) {
					result.push({
						id: 'bdock:' + curApp.id,
						left: 0,
						top:
							index == 0
								? this.DOCKER_APP_START_GAP
								: this.DOCKER_APP_START_GAP +
								  this.DOCKER_APP_SIZE * index +
								  this.DOCKER_APP_GAP * index,
						size: this.DOCKER_APP_SIZE,
						icon: curApp.icon,
						name: curApp.name,
						title: curApp.title,
						namespace: curApp.namespace!,
						owner: curApp.owner!,
						url: curApp.url!,
						is_temp: false,
						show_dot: false
					});
					index++;
				}
			}

			return result;
		},

		async remove_not_exist_apps_on_dock() {
			const bookmarkStore = useBookmarkStore();
			const hasRemoveApp = this.dockerApps.filter((dockerApp) => {
				const rid = this.normalizeAppId(dockerApp.id);
				// Bookmark dock entries resolve against the bookmark store; only
				// drop them when the underlying bookmark no longer exists.
				if (isBookmarkRawId(rid)) {
					return !bookmarkStore.byId[bookmarkIdFromRawId(rid)];
				}
				return !this.myApps.some((app) => 'bdock:' + app.id === dockerApp.id);
			});

			if (hasRemoveApp && hasRemoveApp.length > 0) {
				for (let i = 0; i < hasRemoveApp.length; i++) {
					const app = hasRemoveApp[i];
					this.remove_app_on_docker(app.id, true);
				}
			}
		},

		async get_my_apps_info(isMobile = false, prefetched?: TerminusApp[]) {
			const res = localStorage.getItem('dockerApps');

			if (res) {
				const adjustRes = this.adjustDockerAppPosition(
					JSON.parse(res),
					this.DOCKER_APP_SIZE + this.DOCKER_APP_GAP
				);
				this.dockerApps = adjustRes || JSON.parse(res);
			}

			await this.update_my_apps_info(isMobile, prefetched);

			if (!res) {
				this.dockerApps = await this.get_default_dock_list(isMobile);
				localStorage.setItem('dockerApps', JSON.stringify(this.dockerApps));
			}
		},

		async update_search_result(searchTxt?: string) {
			this.launchpadSearchQuery = String(searchTxt ?? '');
			const data: any = this.myApps;
			const apps: any[] = [];
			for (let i = 0; i < data.length; ++i) {
				const d = data[i];

				if (searchTxt !== undefined) {
					if (
						d.title.toLowerCase().indexOf(searchTxt.toLowerCase()) >= 0 ||
						systemI18nApps(d).toLowerCase().indexOf(searchTxt.toLowerCase()) >=
							0
					) {
						d.type = SearchCategory.Application;
						apps.push(d);
					}
				}
			}
			if (searchTxt !== undefined) {
				const q = searchTxt.toLowerCase();
				for (const bm of this._bookmarksAsAppInfo()) {
					if (bm.title.toLowerCase().indexOf(q) >= 0) {
						apps.push(bm);
					}
				}
			}
			// We already merged matching bookmarks above; don't auto-append again.
			this.relocate_application_place(apps, false);
		},

		async dismiss_search_result() {
			this.launchpadSearchQuery = '';
			this.relocate_application_place(this.myApps);
		},

		async update_my_apps_info(isMobile = false, prefetched?: TerminusApp[]) {
			// if (!tokenStore.token) {
			// 	return;
			// }
			const tokenStore = useTokenStore();

			// Boot folds the myApps list into /server/init; when that prefetched
			// list is handed in we use it verbatim and skip the extra POST. All
			// other callers (reconnect, install/uninstall events) pass nothing
			// and fetch live.
			const data: TerminusApp[] =
				prefetched || (await axios.post(tokenStore.url + '/server/myApps', {}));
			this.axiosAppInfo = data;

			this.myApps = [];

			for (let i = 0; i < data.length; ++i) {
				if (uninstalledAppState(data[i].state, true)) {
					continue;
				}
				if (!data[i].entrances || data[i].entrances.length == 0) {
					continue;
				}
				// if (!data[i].entrances || data[i].entrances.length == 0) {
				// 	this.myApps.push({
				// 		...data[i],
				// 		id: data[i].id,
				// 		appid: data[i].id,
				// 		type: SearchCategory.Application,
				// 		fatherName: null,
				// 		openMethod: 'default',
				// 		isSysApp: data[i].isSysApp,
				// 		fatherState: data[i].state
				// 	});
				// }
				for (let j = 0; j < data[i].entrances.length; ++j) {
					if (data[i].entrances[j].id == 'profile') {
						this.profile_id = data[i].entrances[j].id;
					}
					if (data[i].entrances[j].invisible) {
						// do nothing
					} else {
						const state =
							data[i].state === 'running'
								? data[i].entrances[j].state
								: data[i].state;
						this.myApps.push({
							...data[i],
							state,
							type: SearchCategory.Application,
							id: data[i].entrances[j].id,
							appid: data[i].id,
							name: data[i].entrances[j].name,
							title: data[i].entrances[j].title || data[i].title,
							url: data[i].entrances[j].url || data[i].url,
							icon: data[i].entrances[j].icon || data[i].icon,
							fatherName: data[i].name,
							openMethod: data[i].entrances[j].openMethod || 'default',
							isSysApp: data[i].isSysApp,
							fatherState: data[i].state
						});
					}
				}
			}

			this.relocate_application_place(this.myApps);

			const curApp = {
				id: 'launchpad',
				appid: 'launchpad',
				icon: isMobile
					? './app-icon/launch-icon.png'
					: 'https://app.cdn.olares.com/appstore/launchpad/icon.png',
				name: 'Launchpad',
				title: 'Launchpad',
				target: '',
				//installed: true,
				state: 'running',
				type: SearchCategory.Application,
				fatherName: null,
				openMethod: 'default',
				isSysApp: true,
				requiredGpu: '',
				ports: [],
				entrances: [],
				isClusterScoped: false,
				fatherState: 'running'
			};
			this.myApps.push(curApp);

			this.remove_not_exist_apps_on_dock();

			this.myAppsInitialized = true;
		},

		// Map a bookmark to an AppInfo-shaped record so it can flow through the
		// same launchpad/dock render pipeline as installed apps. The id is
		// namespaced with `bm:` (BOOKMARK_RAW_PREFIX) to avoid colliding with
		// app entrance ids; normalizeAppId/resolveAppForOpen understand it.
		_bookmarkToAppInfo(bm: Bookmark): AppInfo {
			return {
				id: BOOKMARK_RAW_PREFIX + bm.id,
				appid: BOOKMARK_RAW_PREFIX + bm.id,
				icon: bm.icon || bookmarkAvatarDataUri(bm.name),
				title: bm.name,
				name: bm.name,
				target: '',
				url: bm.url,
				state: 'running',
				type: SearchCategory.Application,
				fatherName: null,
				// 'window' bookmarks open in a new browser tab (matches the app
				// 'window' openMethod); 'iframe' bookmarks open inside a desktop
				// window via BasicWindow's 'bookmark-iframe' mode.
				openMethod: bm.openMode === 'window' ? 'window' : 'default',
				isSysApp: false,
				fatherState: 'running',
				isClusterScoped: false
			};
		},

		_bookmarksAsAppInfo(): AppInfo[] {
			const bookmarkStore = useBookmarkStore();
			return bookmarkStore.bookmarks.map((bm) => this._bookmarkToAppInfo(bm));
		},

		// Re-render the live desktop after the bookmark set changes (boot fetch,
		// WS push, create/update/delete). Rebuilds the launchpad grid and
		// re-projects the dock so bookmark icons resolve once their data loads.
		refreshDesktopForBookmarks(): void {
			this.relocate_application_place(this.myApps);
			const dock = this.layoutBySurface.dock;
			if (dock) {
				this._projectDockToDesktop(dock);
			}
		},

		// `includeBookmarks` defaults to true so every existing caller that
		// passes `myApps` now renders bookmarks alongside apps. Search opts out
		// and supplies its own already-filtered list (apps + matching bookmarks).
		relocate_application_place(
			relocate_apps: AppInfo[],
			includeBookmarks = true
		) {
			if (includeBookmarks) {
				relocate_apps = [...relocate_apps, ...this._bookmarksAsAppInfo()];
			}
			this.launchPadApps = [];
			this.desktopApps = [];

			const t: any = {};

			const savedLaunchPad = localStorage.getItem(localStoreKeys.LAUNCH_PAD);
			let re_relocate_apps: AppInfo[] = [];
			if (savedLaunchPad) {
				const copy_relocate_apps: AppInfo[] = JSON.parse(
					JSON.stringify(relocate_apps)
				);
				const apps = JSON.parse(savedLaunchPad);

				for (let i = 0; i < apps.length; ++i) {
					const app = apps[i];
					const index = copy_relocate_apps.findIndex(
						(a) => 'bdesk:' + a.id == app.id
					);
					if (index >= 0) {
						re_relocate_apps.push(copy_relocate_apps[index]);
						copy_relocate_apps.splice(index, 1);
					}
				}
				re_relocate_apps = re_relocate_apps.concat(copy_relocate_apps);
			} else {
				re_relocate_apps = relocate_apps;
			}
			// Drop the synthetic 'launchpad' pseudo-app (it lives at the tail of
			// myApps and is rendered by the dock, not the launchpad grid). It must
			// be removed BEFORE the loop: skipping it mid-loop with `continue`
			// would desync the loop index `i` pushed into launchPadApps from the
			// compacted desktopApps position, yielding undefined slots once any
			// item (e.g. a bookmark) follows it. See bookmark launchpad crash.
			relocate_apps = re_relocate_apps.filter((a) => a.id != 'launchpad');

			const gridX = this.launchpadGridOverride?.x ?? this.DESKTOP_APP_X_NUM;
			const gridY = this.launchpadGridOverride?.y ?? this.DESKTOP_APP_Y_NUM;
			const perPage = gridX * gridY;

			for (let i = 0; i < relocate_apps.length; ++i) {
				t[relocate_apps[i].id] = '1';

				if (i % perPage == 0) {
					this.launchPadApps.push([]);
				}

				const page_num = i % perPage;
				const y = Math.floor(page_num / gridX);
				const x = page_num - y * gridX;

				const d: DesktopAppInfo = {
					id: 'bdesk:' + relocate_apps[i].id,
					appid: 'bdesk:' + relocate_apps[i].appid,
					kind: isBookmarkRawId(relocate_apps[i].id) ? 'bookmark' : 'app',
					index: i,
					page: this.launchPadApps.length - 1,
					page_num: page_num,
					left:
						x == 0
							? this.DOCKER_APP_START_X_GAP
							: this.DOCKER_APP_START_X_GAP +
							  this.DESKTOP_APP_SIZE * x +
							  this.DESKTOP_APP_X_GAP * x,
					top:
						y == 0
							? this.DOCKER_APP_START_Y_GAP
							: this.DOCKER_APP_START_Y_GAP +
							  this.DESKTOP_APP_SIZE * y +
							  this.DESKTOP_APP_Y_GAP * y,
					size: this.DESKTOP_APP_SIZE,
					icon: relocate_apps[i].icon,
					name: relocate_apps[i].name,
					title: relocate_apps[i].title,
					state: relocate_apps[i].state,
					namespace: '',
					owner: '',
					url: '',
					fatherName: relocate_apps[i].fatherName,
					isSysApp: relocate_apps[i].isSysApp,
					fatherState: relocate_apps[i].fatherState,
					isClusterScoped: relocate_apps[i].isClusterScoped
				};

				this.launchPadApps[this.launchPadApps.length - 1].push(i);

				this.desktopApps.push(d);
			}
		},

		resize() {
			for (let i = 0; i < this.desktopApps.length; ++i) {
				const y = Math.floor(
					this.desktopApps[i].page_num / this.DESKTOP_APP_X_NUM
				);
				const x = this.desktopApps[i].page_num - y * this.DESKTOP_APP_X_NUM;
				this.desktopApps[i].left =
					x == 0
						? this.DOCKER_APP_START_X_GAP
						: this.DOCKER_APP_START_X_GAP +
						  this.DESKTOP_APP_SIZE * x +
						  this.DESKTOP_APP_X_GAP * x;
				this.desktopApps[i].top =
					y == 0
						? this.DOCKER_APP_START_Y_GAP
						: this.DOCKER_APP_START_Y_GAP +
						  this.DESKTOP_APP_SIZE * y +
						  this.DESKTOP_APP_Y_GAP * y;
			}
		},

		get_desktop_position(index: number) {
			const page = Math.floor(
				index / (this.DESKTOP_APP_X_NUM * this.DESKTOP_APP_Y_NUM)
			);
			const page_num =
				index % (this.DESKTOP_APP_X_NUM * this.DESKTOP_APP_Y_NUM);
			const y = Math.floor(page_num / this.DESKTOP_APP_X_NUM);
			const x = page_num - y * this.DESKTOP_APP_X_NUM;
			const p: DesktopPosition = {
				index: index,
				page: page,
				page_num: page_num,
				left:
					x == 0
						? this.DOCKER_APP_START_X_GAP
						: this.DOCKER_APP_START_X_GAP +
						  this.DESKTOP_APP_SIZE * x +
						  this.DESKTOP_APP_X_GAP * x,
				top:
					y == 0
						? this.DOCKER_APP_START_Y_GAP
						: this.DOCKER_APP_START_Y_GAP +
						  this.DESKTOP_APP_SIZE * y +
						  this.DESKTOP_APP_Y_GAP * y
			};
			return p;
		},

		find_app(id: string) {
			return this.dockerApps.find((app) => app.id == id);
		},

		normalizeAppId(id: string) {
			let rid = id;
			if (rid.startsWith('bdock:')) {
				rid = rid.substring(6);
			} else if (rid.startsWith('bdesk:')) {
				rid = rid.substring(6);
			}
			return rid;
		},

		resolveAppForOpen(id: string): AppInfo | null {
			const rid = this.normalizeAppId(id);
			if (isBookmarkRawId(rid)) {
				const bookmarkStore = useBookmarkStore();
				const bm = bookmarkStore.byId[bookmarkIdFromRawId(rid)];
				return bm ? this._bookmarkToAppInfo(bm) : null;
			}
			const fromMyApps = this.myApps.find((app) => app.id == rid);
			if (fromMyApps) {
				return fromMyApps;
			}

			if (this.myAppsInitialized) {
				return null;
			}

			const docker = this.dockerApps.find((app) => app.id === 'bdock:' + rid);
			if (!docker) {
				return null;
			}

			return {
				id: rid,
				appid: rid,
				icon: docker.icon,
				title: docker.title,
				name: docker.name,
				target: '',
				namespace: docker.namespace,
				owner: docker.owner,
				url: docker.url,
				state: 'running',
				fatherName: null,
				openMethod: 'default',
				isSysApp: false,
				fatherState: 'running',
				isClusterScoped: false
			};
		},

		is_app_in_docker(id: string) {
			let rid = id;
			if (rid.startsWith('bdesk:')) {
				rid = rid.substring(6);
				rid = 'bdock:' + id;
			} else if (rid.startsWith('bdock:') == false) {
				rid = 'bdock:' + id;
			}
			const t = this.dockerApps.find((app) => app.id == rid);
			return t != undefined;
		},

		async add_app_on_docker_bottom(id: string) {
			await this.add_app_on_docker_at(id, this.dockerApps.length, true, true);
		},

		async add_app_on_docker_at(
			id: string,
			index: number,
			is_temp: boolean,
			show_dot: boolean
		) {
			const a = this.resolveAppForOpen(id);
			if (!a) {
				console.log('app not found ', this.normalizeAppId(id));
				return;
			}
			const fullId = 'bdock:' + a.id;
			if (this.dockerApps.find((app) => app.id === fullId)) {
				return;
			}
			const d: DockerAppInfo = {
				id: fullId,
				left: 0,
				top: 0,
				size: this.DOCKER_APP_SIZE,
				icon: a.icon,
				name: a.name,
				title: a.title,
				namespace: a.namespace,
				owner: a.owner,
				url: a.url,
				is_temp,
				show_dot,
				kind: isBookmarkRawId(a.id) ? 'bookmark' : 'app'
			};
			const ordered = [...this.dockerApps].sort((x, y) => x.top - y.top);
			const clamped = Math.max(0, Math.min(index, ordered.length));
			ordered.splice(clamped, 0, d);
			for (let i = 0; i < ordered.length; i++) {
				ordered[i].top =
					this.DOCKER_APP_START_GAP +
					(this.DOCKER_APP_SIZE + this.DOCKER_APP_GAP) * i;
			}
			this.dockerApps = ordered;

			localStorage.setItem('dockerApps', JSON.stringify(this.dockerApps));
		},

		async add_app_on_docker(
			id: string,
			top: number,
			is_temp: boolean,
			show_dot: boolean
		) {
			console.log('add_app_on_docker is_temp', is_temp);
			if (this.dockerApps.find((app) => app.id === id)) {
				return;
			}

			const a = this.resolveAppForOpen(id);
			if (!a) {
				console.log('app not found ', this.normalizeAppId(id));
				return;
			}
			const d: DockerAppInfo = {
				id: 'bdock:' + a.id,
				left: 0,
				top: top,
				size: this.DOCKER_APP_SIZE,
				icon: a.icon,
				name: a.name,
				title: a.title,
				namespace: a.namespace,
				owner: a.owner,
				url: a.url,
				is_temp,
				show_dot,
				kind: isBookmarkRawId(a.id) ? 'bookmark' : 'app'
			};
			this.dockerApps.push(d);

			localStorage.setItem('dockerApps', JSON.stringify(this.dockerApps));
		},

		async remove_app_on_docker(id: string, isRemove: boolean) {
			const index: number = this.dockerApps.findIndex((app) => app.id == id);

			if (index < 0) return false;
			if (isRemove) {
				const dockerApps = JSON.parse(JSON.stringify(this.dockerApps));
				const targets: any[] = [];
				for (let i = 0; i < dockerApps.length; i++) {
					if (i == index) {
						continue;
					}

					const top =
						dockerApps[i].top > dockerApps[index].top
							? dockerApps[i].top - this.DOCKER_APP_SIZE - this.DOCKER_APP_GAP
							: dockerApps[i].top;

					targets.push({
						...dockerApps[i],
						top
					});
				}
				this.dockerApps = targets;
			} else {
				this.dockerApps.splice(index, 1);
			}

			localStorage.setItem('dockerApps', JSON.stringify(this.dockerApps));
		},

		adjustDockerAppPosition(data: DockerAppInfo[], tolerance: number) {
			data.sort((a, b) => a.top - b.top);

			let hasSort = true;
			for (let i = 1; i < data.length; i++) {
				if (data[i].top - data[i - 1].top !== tolerance) {
					hasSort = false;
					break;
				}
			}

			if (hasSort) {
				if (data[0].top === this.DOCKER_APP_START_GAP) {
					return;
				}
			}

			const adjustedData: DockerAppInfo[] = [];
			let currentTop = this.DOCKER_APP_START_GAP;

			for (const item of data) {
				if (adjustedData.find((adjustedItem) => adjustedItem.id === item.id)) {
					continue;
				}

				item.top !== currentTop && (item.top = currentTop);

				adjustedData.push({ ...item });

				currentTop += tolerance;
			}

			return adjustedData;
		},

		async updateAppInDock(value: DockerAppInfo) {
			const index = this.dockerApps.findIndex((app) => app.id === value.id);
			this.dockerApps[index] = value;
			localStorage.setItem('dockerApps', JSON.stringify(this.dockerApps));
		},

		async uninstall_application(app_name: string | null, data: any = {}) {
			if (!app_name) {
				return;
			}
			const removedIds = this.myApps
				.filter((item) => item.fatherName === app_name)
				.map((item: any) => item.id);
			this.myApps = this.myApps.filter((item) => item.fatherName !== app_name);
			this.relocate_application_place(this.myApps);
			this.remove_not_exist_apps_on_dock();

			const keys = Array.from(new Set(removedIds)).map((id) => 'a:' + id);
			if (keys.length) {
				void this.cleanupOrphanKeys(keys);
			}
		},

		updateOneApplicationState(payload: AppStatusChange) {
			this.appMessageQueue.enqueue({
				name: payload.app_name,
				payload,
				isdequeued: false
			});
			this.startDealMessageTimer();
		},

		async startDealMessageTimer() {
			if (this.appMessageQueue.isEmpty()) {
				return;
			}
			if (this.appMessageQueue.isWorking) {
				return;
			}
			this.appMessageQueue.isWorking = true;
			while (!this.appMessageQueue.isEmpty()) {
				const msg = this.appMessageQueue.dequeue();
				if (msg) {
					await this.dealApplicationEntrancesMessages(msg);
				}
			}
			this.appMessageQueue.isWorking = false;
		},

		async dealApplicationEntrancesMessages(message: {
			name: string;
			payload: AppStatusChange;
		}) {
			const appName = message.name;
			const appStatus = message.payload.app_state_latest?.status;
			if (!appName || !appStatus) {
				return;
			}
			const state = appStatus.state;

			if (uninstalledAppState(state, true)) {
				this.myApps = this.myApps.filter((e) => e.fatherName != appName);
			} else {
				const apps = this.myApps.filter((e) => e.fatherName == appName);
				if (apps.length == 0) {
					await this.update_my_apps_info();
					return;
				}

				const sourceEntrances = appStatus.entranceStatuses ?? [];

				if (sourceEntrances.length == 0) {
					this.myApps = this.myApps.filter((e) => e.fatherName != appName);
					return;
				}

				const visibleEntrances = sourceEntrances.filter(
					(e) => e.invisible == false && !!e.name
				);
				if (visibleEntrances.length === 0) {
					this.myApps = this.myApps.filter((e) => e.fatherName != appName);
					return;
				}

				const getScopedEntranceKey = (entranceName: string) =>
					`${appName}::${entranceName}`;
				const entranceMap = new Map(
					visibleEntrances.map((entrance) => [
						getScopedEntranceKey(entrance.name),
						entrance
					])
				);
				this.myApps = this.myApps
					.filter(
						(app) =>
							app.fatherName !== appName ||
							entranceMap.has(getScopedEntranceKey(app.name))
					)
					.map((app) => {
						if (app.fatherName !== appName) {
							return app;
						}
						const entrance = entranceMap.get(getScopedEntranceKey(app.name));
						if (!entrance) {
							return app;
						}
						return {
							...app,
							state: entrance.state || state,
							fatherState: state
						};
					});
			}
			this.relocate_application_place(this.myApps);
			this.remove_not_exist_apps_on_dock();
		},

		// ---------- Cross-device layout sync (M1+) ----------

		_emptyLaunch(): LaunchLayout {
			return {
				schemaVersion: LAYOUT_SCHEMA_VERSION,
				items: [],
				hidden: [],
				version: 0,
				updatedAt: 0,
				updatedBy: ''
			};
		},

		_emptyDock(): DockLayout {
			return {
				schemaVersion: LAYOUT_SCHEMA_VERSION,
				items: [],
				version: 0,
				updatedAt: 0,
				updatedBy: ''
			};
		},

		_loadCache<T>(key: string): T | null {
			try {
				const raw = localStorage.getItem(key);
				return raw ? (JSON.parse(raw) as T) : null;
			} catch (_e) {
				return null;
			}
		},

		_saveCache(key: string, value: unknown): void {
			try {
				localStorage.setItem(key, JSON.stringify(value));
			} catch (_e) {
				/* quota / private mode — ignore */
			}
		},

		// Apply a remote launch layout coming from GET or a WS push. Skips
		// the apply when the broadcasted appliedOpIds matches an in-flight
		// local op (self-echo suppression).
		// Called by the drag layer around an active drag. When locked, inbound
		// remote projections are stashed (latest-wins) and replayed on unlock.
		setLayoutDragLock(locked: boolean) {
			if (this.isLayoutDragLocked === locked) return;
			this.isLayoutDragLocked = locked;
			if (!locked) {
				const pl = this.pendingLaunchLayout;
				if (pl) {
					this.pendingLaunchLayout = null;
					this.applyRemoteLaunchLayout(pl.layout, pl.opIds);
				}
				const pd = this.pendingDockLayout;
				if (pd) {
					this.pendingDockLayout = null;
					this.applyRemoteDockLayout(pd.layout, pd.opIds);
				}
			}
		},

		// Drop a dock layout stashed during the lock window without replaying it.
		// Used when a drag resolves to a pure restore (revert to the pre-drag
		// local layout with no persist): the live dock is authoritative, so a
		// deferred remote projection must not clobber it on unlock.
		discardPendingDockLayout() {
			this.pendingDockLayout = null;
		},

		applyRemoteLaunchLayout(layout: LaunchLayout, appliedOpIds: string[] = []) {
			if (this.isLayoutDragLocked) {
				this.pendingLaunchLayout = { layout, opIds: appliedOpIds };
				return;
			}
			if (appliedOpIds.some((id) => this.pendingOpIds.has(id))) {
				for (const id of appliedOpIds) this.pendingOpIds.delete(id);
				return;
			}
			this.layoutBySurface.launchpad = layout;
			this._saveCache(localStoreKeys.LAUNCH_LAYOUT_CACHE, layout);
			this.syncedAt['launchpad'] = Date.now();
			// Project the synced order into the live desktop so other devices /
			// tabs actually re-render to match.
			this._projectLaunchToDesktop(layout);
		},

		applyRemoteDockLayout(layout: DockLayout, appliedOpIds: string[] = []) {
			if (this.isLayoutDragLocked) {
				this.pendingDockLayout = { layout, opIds: appliedOpIds };
				return;
			}
			if (appliedOpIds.some((id) => this.pendingOpIds.has(id))) {
				for (const id of appliedOpIds) this.pendingOpIds.delete(id);
				return;
			}
			this.layoutBySurface.dock = layout;
			this._saveCache(localStoreKeys.DOCK_LAYOUT_CACHE, layout);
			this.syncedAt['dock'] = Date.now();
			this._projectDockToDesktop(layout);
		},

		// ---------- Inbound projection (layout -> live desktop render) ----------
		// The launchpad grid renders from desktopApps/launchPadApps, which
		// relocate_application_place rebuilds from localStorage['lanuchPad'].
		// To make a synced/remote layout actually visible we rewrite that
		// ordering source from the layout's app ('a:') and bookmark ('b:')
		// items and re-relocate. Items that don't yet resolve (e.g. a bookmark
		// whose data WS hasn't arrived) are simply skipped by relocate and get
		// picked up on the next refresh once their data lands.
		_projectLaunchToDesktop(layout: LaunchLayout): void {
			if (!layout || !Array.isArray(layout.items)) return;
			// The layout is a pure linear order; relocate_application_place
			// reflows it onto this device's grid (PC fixed 7×4, mobile via
			// launchpadGridOverride), so we only need the ordering here.
			const ordered = [...layout.items].sort((a, b) => a.order - b.order);
			const seen = new Set<string>();
			const saved: { id: string; index: number }[] = [];
			let index = 0;
			for (const it of ordered) {
				const parsed = parseKey(it.key);
				if (!parsed || (parsed.kind !== 'a' && parsed.kind !== 'b')) continue;
				const rawId = keyToRawId(parsed);
				if (seen.has(rawId)) continue;
				seen.add(rawId);
				saved.push({ id: 'bdesk:' + rawId, index: index++ });
			}
			try {
				localStorage.setItem(localStoreKeys.LAUNCH_PAD, JSON.stringify(saved));
			} catch (_e) {
				/* ignore quota */
			}
			// relocate appends any installed app / bookmark not present in the
			// ordering at the tail (newly added items stay visible).
			this.relocate_application_place(this.myApps);
		},

		// Rebuild the live dock from a synced/remote dock layout. Pinned items
		// come from the layout (is_temp=false); currently-open temp apps that
		// aren't pinned are preserved at the tail so running windows keep their
		// dock entry.
		_projectDockToDesktop(layout: DockLayout): void {
			if (!layout || !Array.isArray(layout.items)) return;
			const prevById = new Map(this.dockerApps.map((d) => [d.id, d]));
			const pinned: DockerAppInfo[] = [];
			const seen = new Set<string>();
			const ordered = [...layout.items].sort((a, b) => a.slot - b.slot);
			for (const it of ordered) {
				const parsed = parseKey(it.key);
				if (!parsed || (parsed.kind !== 'a' && parsed.kind !== 'b')) continue;
				const rawId = keyToRawId(parsed);
				if (seen.has(rawId)) continue;
				const a = this.resolveAppForOpen(rawId);
				if (!a) continue;
				seen.add(rawId);
				const prev = prevById.get('bdock:' + a.id);
				pinned.push({
					id: 'bdock:' + a.id,
					left: 0,
					top: 0,
					size: this.DOCKER_APP_SIZE,
					icon: a.icon,
					name: a.name,
					title: a.title,
					namespace: a.namespace || '',
					owner: a.owner || '',
					url: a.url || '',
					is_temp: false,
					show_dot: prev?.show_dot ?? false,
					kind: isBookmarkRawId(rawId) ? 'bookmark' : 'app'
				});
			}
			// Launchpad is now a normal draggable + persisted dock item, so a
			// seeded layout already carries it at its own slot (handled by the
			// loop above). This block is only a backstop for legacy layouts that
			// predate persistence (or were seeded without launchpad): inject it
			// at the front so the icon never vanishes / becomes unreachable. The
			// first dock edit then persists it and it becomes reorderable.
			if (!seen.has('launchpad')) {
				const lp = this.resolveAppForOpen('launchpad');
				if (lp) {
					const prevLp = prevById.get('bdock:launchpad');
					pinned.unshift({
						id: 'bdock:launchpad',
						left: 0,
						top: 0,
						size: this.DOCKER_APP_SIZE,
						icon: lp.icon,
						name: lp.name,
						title: lp.title,
						namespace: lp.namespace || '',
						owner: lp.owner || '',
						url: lp.url || '',
						is_temp: false,
						show_dot: prevLp?.show_dot ?? false
					});
					seen.add('launchpad');
				}
			}
			const temps = this.dockerApps.filter(
				(d) => d.is_temp && !seen.has(this.normalizeAppId(d.id))
			);
			const merged = [...pinned, ...temps];
			for (let i = 0; i < merged.length; i++) {
				merged[i].top =
					this.DOCKER_APP_START_GAP +
					(this.DOCKER_APP_SIZE + this.DOCKER_APP_GAP) * i;
			}
			this.dockerApps = merged;
			try {
				localStorage.setItem('dockerApps', JSON.stringify(this.dockerApps));
			} catch (_e) {
				/* ignore quota */
			}
		},

		// Debounced setters. The component layer mutates layoutBySurface
		// optimistically AND calls these, so UI never waits on the network.
		saveLaunchLayoutDebounced(layout: LaunchLayout): void {
			// A local write is the most recent launchpad truth, so drop any
			// remote layout stashed earlier in this drag's lock window; only
			// remote updates that arrive AFTER this write should win on unlock.
			this.pendingLaunchLayout = null;
			this.layoutBySurface.launchpad = layout;
			this._saveCache(localStoreKeys.LAUNCH_LAYOUT_CACHE, layout);

			if (!launchDebounce) {
				launchDebounce = debounce(async (next: LaunchLayout) => {
					const opId = genClientOpId();
					this.pendingOpIds.add(opId);
					try {
						const r = await apiPutLaunch({
							expectedVersion: next.version,
							layout: {
								schemaVersion: next.schemaVersion,
								items: next.items,
								hidden: next.hidden
							},
							deviceId: getDeviceId(),
							clientOpId: opId
						});
						const current = (r as any).current as LaunchLayout;
						this.layoutBySurface.launchpad = current;
						this._saveCache(localStoreKeys.LAUNCH_LAYOUT_CACHE, current);
						this.syncedAt['launchpad'] = Date.now();
					} catch (e: any) {
						this.pendingOpIds.delete(opId);
						const resp = e?.response;
						if (resp?.status === 409 && resp.data?.current) {
							const current = resp.data.current as LaunchLayout;
							const merged = this._simpleMergeLaunch(current, next);
							this.layoutBySurface.launchpad = merged;
							this.saveLaunchLayoutDebounced(merged);
						} else {
							console.error('[launchpad] save failed', e);
						}
					}
				}, DEBOUNCE_WAIT_MS);
			}
			launchDebounce(layout);
		},

		saveDockLayoutDebounced(layout: DockLayout): void {
			// See saveLaunchLayoutDebounced: a local write supersedes any remote
			// dock layout stashed earlier in this drag's lock window.
			this.pendingDockLayout = null;
			this.layoutBySurface.dock = layout;
			this._saveCache(localStoreKeys.DOCK_LAYOUT_CACHE, layout);

			if (!dockDebounce) {
				dockDebounce = debounce(async (next: DockLayout) => {
					const opId = genClientOpId();
					this.pendingOpIds.add(opId);
					try {
						const r = await apiPutDock({
							layout: {
								schemaVersion: next.schemaVersion,
								items: next.items,
								capacity: next.capacity
							},
							deviceId: getDeviceId(),
							clientOpId: opId
						});
						const current = (r as any).current as DockLayout;
						this.layoutBySurface.dock = current;
						this._saveCache(localStoreKeys.DOCK_LAYOUT_CACHE, current);
						this.syncedAt['dock'] = Date.now();
					} catch (e) {
						this.pendingOpIds.delete(opId);
						console.error('[dock] save failed', e);
					}
				}, DEBOUNCE_WAIT_MS);
			}
			dockDebounce(layout);
		},

		cancelPendingLayoutSaves(): void {
			launchDebounce?.cancel();
			dockDebounce?.cancel();
		},

		// Used by M3 reset: wipes the in-memory layouts, the local layout
		// caches, the legacy LaunchPad migration cache, and the migration
		// flag — so the next syncDesktopLayoutOnBoot rebuilds from a clean
		// slate against the post-DEL server state.
		resetLocalLayoutCaches(): void {
			this.layoutBySurface.launchpad = undefined;
			this.layoutBySurface.dock = undefined;
			// Drop the live dock too, otherwise the legacy 'dockerApps' seed
			// would rebuild the pre-reset dock and the reset wouldn't take.
			this.dockerApps = [];
			this.pendingOpIds.clear();
			this.syncedAt = {};
			try {
				localStorage.removeItem(localStoreKeys.LAUNCH_LAYOUT_CACHE);
				localStorage.removeItem(localStoreKeys.DOCK_LAYOUT_CACHE);
				localStorage.removeItem(localStoreKeys.LAUNCH_PAD);
				localStorage.removeItem(localStoreKeys.DOCK_PAD);
				localStorage.removeItem(localStoreKeys.LAYOUT_MIGRATED);
			} catch (_e) {
				/* ignore */
			}
		},

		// Simple merge for 409: keep keys already on server + append any
		// keys from "ours" that the server doesn't have yet. The server has
		// its own normalize so we don't have to be careful about page/slot.
		_simpleMergeLaunch(server: LaunchLayout, ours: LaunchLayout): LaunchLayout {
			const seen = new Set(server.items.map((it) => it.key));
			const extra = ours.items.filter((it) => !seen.has(it.key));
			const items = [...server.items, ...extra];
			return {
				...server,
				items,
				hidden: server.hidden,
				version: server.version
			};
		},

		// Boot-time fetch + seed. Pulls the single shared layout; if unseeded
		// we seed the server (expectedVersion=0) from local state — preferring
		// a lossless import of the legacy localStorage order (main only ever
		// used localStorage), falling back to myApps order + bookmarks. On a
		// multi-device first boot the loser of the race hits 409 and adopts the
		// server's version (first-boot-wins).
		async syncDesktopLayoutOnBoot(initLayout?: InitLayout): Promise<void> {
			// Stale-while-revalidate: paint the cached layout immediately while
			// the network round-trip resolves.
			const cachedLaunch = this._loadCache<LaunchLayout>(
				localStoreKeys.LAUNCH_LAYOUT_CACHE
			);
			if (cachedLaunch) this.layoutBySurface.launchpad = cachedLaunch;
			const cachedDock = this._loadCache<DockLayout>(
				localStoreKeys.DOCK_LAYOUT_CACHE
			);
			if (cachedDock) this.layoutBySurface.dock = cachedDock;

			try {
				// Boot folds both layout envelopes into /server/init; when that
				// bundle is provided we use it and skip the two GETs. Reconnect /
				// reset callers pass nothing and fetch live. For an unseeded
				// surface the server bundle carries a default layout (computed
				// from myApps) that the seed path persists when there's no legacy
				// localStorage to migrate.
				const [launchEnv, dockEnv] = initLayout
					? [initLayout.launch, initLayout.dock]
					: await Promise.all([apiGetLaunch(), apiGetDock()]);

				if (launchEnv.seeded && launchEnv.layout) {
					this.layoutBySurface.launchpad = launchEnv.layout;
					this._saveCache(localStoreKeys.LAUNCH_LAYOUT_CACHE, launchEnv.layout);
					this._projectLaunchToDesktop(launchEnv.layout);
				} else {
					await this._seedLaunchLayout(launchEnv.layout || undefined);
				}

				if (dockEnv.seeded && dockEnv.layout) {
					this.layoutBySurface.dock = dockEnv.layout;
					this._saveCache(localStoreKeys.DOCK_LAYOUT_CACHE, dockEnv.layout);
					this._projectDockToDesktop(dockEnv.layout);
				} else {
					await this._seedDockLayout(dockEnv.layout || undefined);
				}

				localStorage.setItem(localStoreKeys.LAYOUT_MIGRATED, '1');
			} catch (e) {
				console.warn('[desktop-layout] boot sync failed', e);
			}
		},

		// Seed the shared launchpad layout on first boot (server key unseeded).
		// localStorage['lanuchPad'] is already a linear ordered array on `main`,
		// so we import it losslessly as { key, order }; then append any myApps /
		// bookmarks not yet present; finally fall back to myApps order +
		// bookmarks when no legacy data exists.
		async _seedLaunchLayout(serverDefault?: LaunchLayout): Promise<void> {
			const items: LaunchpadItem[] = [];
			const seen = new Set<string>();

			const legacy = localStorage.getItem(localStoreKeys.LAUNCH_PAD);
			if (legacy) {
				try {
					const apps = JSON.parse(legacy) as Array<{ id: string }>;
					for (let i = 0; i < apps.length; i++) {
						const raw = apps[i]?.id || '';
						const rawId = raw.startsWith('bdesk:') ? raw.slice(6) : raw;
						if (!rawId || seen.has(rawId)) continue;
						seen.add(rawId);
						items.push({ key: rawIdToKey(rawId), order: items.length });
					}
				} catch (_e) {
					/* corrupt legacy json — ignore */
				}
			} else if (serverDefault && serverDefault.items?.length) {
				// No legacy data to migrate: adopt the server-computed default
				// (built from myApps) as the seed instead of recomputing locally.
				for (const it of serverDefault.items) {
					const parsed = parseKey(it.key);
					if (!parsed) continue;
					const rawId = keyToRawId(parsed);
					if (!rawId || seen.has(rawId)) continue;
					seen.add(rawId);
					items.push({ key: it.key, order: items.length });
				}
			}

			// Anything in myApps not yet placed gets appended at the tail.
			for (let i = 0; i < this.myApps.length; i++) {
				const id = (this.myApps[i] as any).id;
				if (!id || seen.has(id)) continue;
				if (id === 'launchpad') continue;
				seen.add(id);
				items.push({ key: encodeKey('a', id), order: items.length });
			}

			// Bookmarks not yet placed are appended too, so a fresh (unseeded)
			// server gets bookmarks on the grid from the first seed.
			for (const bm of this._bookmarksAsAppInfo()) {
				const rawId = bm.id;
				if (seen.has(rawId)) continue;
				seen.add(rawId);
				items.push({ key: rawIdToKey(rawId), order: items.length });
			}

			const seed: LaunchLayout = {
				...this._emptyLaunch(),
				items,
				hidden: []
			};
			try {
				const r = await apiPutLaunch({
					expectedVersion: 0,
					layout: {
						schemaVersion: seed.schemaVersion,
						items: seed.items,
						hidden: seed.hidden
					},
					deviceId: getDeviceId(),
					clientOpId: genClientOpId()
				});
				const current = (r as any).current as LaunchLayout;
				this.layoutBySurface.launchpad = current;
				this._saveCache(localStoreKeys.LAUNCH_LAYOUT_CACHE, current);
				this._projectLaunchToDesktop(current);
			} catch (e: any) {
				// Another device seeded first → adopt the server version.
				if (e?.response?.status === 409 && e.response.data?.current) {
					const current = e.response.data.current as LaunchLayout;
					this.layoutBySurface.launchpad = current;
					this._saveCache(localStoreKeys.LAUNCH_LAYOUT_CACHE, current);
					this._projectLaunchToDesktop(current);
				} else {
					console.warn('[desktop-layout] launch seed upload failed', e);
				}
			}
		},

		async _seedDockLayout(serverDefault?: DockLayout): Promise<void> {
			const items: DockItem[] = [];
			const seen = new Set<string>();
			const legacy = localStorage.getItem('dockerApps');
			if (legacy) {
				try {
					const apps = JSON.parse(legacy) as Array<{
						id: string;
						is_temp?: boolean;
					}>;
					let slot = 0;
					for (const a of apps) {
						if (a.is_temp) continue;
						const raw = a?.id || '';
						const rawId = raw.startsWith('bdock:') ? raw.slice(6) : raw;
						if (!rawId || seen.has(rawId)) continue;
						seen.add(rawId);
						items.push({ key: rawIdToKey(rawId), slot: slot++ });
					}
				} catch (_e) {
					/* ignore */
				}
			}

			// No legacy data (fresh install or post-reset): seed the canonical
			// default dock — files, launchpad, market, settings — so the dock is
			// never empty and launchpad keeps its slot (index 1). launchpad is a
			// first-class, draggable, persisted item now (key 'a:launchpad').
			// Prefer the server-computed default (already built from myApps in
			// the canonical order); fall back to the local get_default_dock_list
			// when none was provided (e.g. the GET-based reconnect/reset path).
			if (items.length === 0) {
				let slot = 0;
				if (serverDefault && serverDefault.items?.length) {
					for (const it of serverDefault.items) {
						const parsed = parseKey(it.key);
						if (!parsed) continue;
						const rawId = keyToRawId(parsed);
						if (!rawId || seen.has(rawId)) continue;
						seen.add(rawId);
						items.push({ key: it.key, slot: slot++ });
					}
				} else {
					const defaults = await this.get_default_dock_list(false);
					for (const a of defaults) {
						const raw = a?.id || '';
						const rawId = raw.startsWith('bdock:') ? raw.slice(6) : raw;
						if (!rawId || seen.has(rawId)) continue;
						seen.add(rawId);
						items.push({ key: rawIdToKey(rawId), slot: slot++ });
					}
				}
			}

			const seed: DockLayout = { ...this._emptyDock(), items };
			try {
				const r = await apiPutDock({
					layout: {
						schemaVersion: seed.schemaVersion,
						items: seed.items,
						capacity: seed.capacity
					},
					deviceId: getDeviceId(),
					clientOpId: genClientOpId()
				});
				const current = (r as any).current as DockLayout;
				this.layoutBySurface.dock = current;
				this._saveCache(localStoreKeys.DOCK_LAYOUT_CACHE, current);
				this._projectDockToDesktop(current);
			} catch (e) {
				console.warn('[desktop-layout] dock seed upload failed', e);
			}
		},

		// Removes layout keys left behind by uninstalled apps.
		async cleanupOrphanKeys(keys: string[]): Promise<void> {
			if (!keys || keys.length === 0) return;
			try {
				await apiOrphanCleanup(keys);
			} catch (e) {
				console.warn('[desktop-layout] orphan cleanup failed', e);
			}
		},

		// Appends b:<id> to the end of the shared launchpad layout. Idempotent:
		// if the key already exists we leave it alone. The server's normalize
		// step will compact and re-number `order`.
		async autoPlaceBookmarkOnLaunchpads(bookmarkId: string): Promise<void> {
			const key = `b:${bookmarkId}`;
			const cur = this.layoutBySurface.launchpad || this._emptyLaunch();
			if (!cur.items.some((it) => it.key === key)) {
				const next: LaunchLayout = {
					...cur,
					items: [...cur.items, { key, order: cur.items.length }]
				};
				this.saveLaunchLayoutDebounced(next);
			}
			// Render the new bookmark on the live grid immediately (the layout
			// PUT above is self-echo suppressed, so we won't get a WS re-project).
			this.refreshDesktopForBookmarks();
		},

		// Ensure every existing bookmark has a persisted slot in the shared
		// launchpad layout. Appends any missing b:<id> keys at the tail so
		// bookmarks become first-class: stably ordered (after apps),
		// cross-device synced, and reorderable — instead of only being rendered
		// via relocate's leftover fallback. Only runs when the layout is
		// already loaded, so we never seed/clobber a layout whose server state
		// hasn't been fetched yet.
		backfillBookmarksOnLaunchpads(): void {
			const cur = this.layoutBySurface.launchpad;
			if (!cur) return;
			const bookmarkStore = useBookmarkStore();
			const existing = new Set(cur.items.map((it) => it.key));
			const added: LaunchpadItem[] = [];
			for (const bm of bookmarkStore.bookmarks) {
				const key = encodeKey('b', bm.id);
				if (existing.has(key)) continue;
				existing.add(key);
				added.push({ key, order: cur.items.length + added.length });
			}
			if (added.length === 0) return;
			const next: LaunchLayout = {
				...cur,
				items: [...cur.items, ...added]
			};
			this.saveLaunchLayoutDebounced(next);
			this.refreshDesktopForBookmarks();
		},

		// Strip a layout key ('a:'/'b:') from the launchpad + dock layout and
		// persist the change. Used when a bookmark is deleted from the desktop.
		_removeKeyFromAllLayouts(key: string): void {
			const launch = this.layoutBySurface.launchpad;
			if (launch && launch.items.some((it) => it.key === key)) {
				this.saveLaunchLayoutDebounced({
					...launch,
					items: launch.items.filter((it) => it.key !== key)
				});
			}
			const dock = this.layoutBySurface.dock;
			if (dock && dock.items.some((it) => it.key === key)) {
				this.saveDockLayoutDebounced({
					...dock,
					items: dock.items.filter((it) => it.key !== key)
				});
			}
		},

		// Desktop-level bookmark delete (shake-mode delete entry). Deletes the
		// bookmark via the bookmark store, removes its key from all layouts,
		// drops it from the live dock, re-renders the grid, and cleans up orphan
		// keys.
		async deleteBookmarkFromDesktop(bookmarkId: string): Promise<boolean> {
			const bookmarkStore = useBookmarkStore();
			const removed = await bookmarkStore.delete(bookmarkId);
			if (!removed) return false;
			const key = encodeKey('b', bookmarkId);
			this._removeKeyFromAllLayouts(key);
			this.remove_app_on_docker(
				'bdock:' + BOOKMARK_RAW_PREFIX + bookmarkId,
				true
			);
			this.refreshDesktopForBookmarks();
			void this.cleanupOrphanKeys([key]);
			return true;
		}
	}
});
