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

	myApps: AppInfo[];
	desktopApps: DesktopAppInfo[];

	axiosAppInfo: TerminusApp[];

	appMessageQueue: MessageQueue<{
		name: string;
		payload: AppStatusChange;
		isdequeued: boolean;
	}>;
};

const socketMessageTimer: NodeJS.Timeout | null = null;

export const localStoreKeys = {
	LAUNCH_PAD: 'lanuchPad'
};

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
			desktopApps: [],
			myApps: [],
			axiosAppInfo: [],
			appMessageQueue: new MessageQueue()
		} as AppState;
	},
	actions: {
		async get_default_dock_list(isMobile: boolean) {
			let index = 0;
			const result: DockerAppInfo[] = [];
			const dock_app_lists = isMobile
				? ['settings', 'files', 'launchpad', 'wise', 'vault']
				: ['files', 'launchpad', 'market', 'settings'];

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
			const hasRemoveApp = this.dockerApps.filter(
				(dockerApp) =>
					!this.myApps.some((app) => 'bdock:' + app.id === dockerApp.id)
			);

			if (hasRemoveApp && hasRemoveApp.length > 0) {
				for (let i = 0; i < hasRemoveApp.length; i++) {
					const app = hasRemoveApp[i];
					this.remove_app_on_docker(app.id, true);
				}
			}
		},

		async get_my_apps_info(isMobile = false) {
			const res = localStorage.getItem('dockerApps');

			if (res) {
				const adjustRes = this.adjustDockerAppPosition(
					JSON.parse(res),
					this.DOCKER_APP_SIZE + this.DOCKER_APP_GAP
				);
				this.dockerApps = adjustRes || JSON.parse(res);
			}

			await this.update_my_apps_info(isMobile);

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
			this.relocate_application_place(apps);
		},

		async dismiss_search_result() {
			this.launchpadSearchQuery = '';
			this.relocate_application_place(this.myApps);
		},

		async update_my_apps_info(isMobile = false) {
			// if (!tokenStore.token) {
			// 	return;
			// }
			const tokenStore = useTokenStore();

			const data: TerminusApp[] = await axios.post(
				tokenStore.url + '/server/myApps',
				{}
			);
			this.axiosAppInfo = data;

			this.myApps = [];

			for (let i = 0; i < data.length; ++i) {
				if (uninstalledAppState(data[i].state)) {
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
		},

		relocate_application_place(relocate_apps: AppInfo[]) {
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
			relocate_apps = re_relocate_apps;

			const gridX = this.launchpadGridOverride?.x ?? this.DESKTOP_APP_X_NUM;
			const gridY = this.launchpadGridOverride?.y ?? this.DESKTOP_APP_Y_NUM;
			const perPage = gridX * gridY;

			for (let i = 0; i < relocate_apps.length; ++i) {
				if (relocate_apps[i].id == 'launchpad') {
					continue;
				}
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
			const index = this.dockerApps.length;
			const top =
				index == 0
					? this.DOCKER_APP_START_GAP
					: this.DOCKER_APP_START_GAP +
					  this.DOCKER_APP_SIZE * index +
					  this.DOCKER_APP_GAP * index;
			this.add_app_on_docker(id, top, true, true);
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

			let rid = id;
			if (rid.startsWith('bdock:')) {
				rid = rid.substring(6);
			} else if (rid.startsWith('bdesk:')) {
				rid = rid.substring(6);
			}

			const a: any = this.myApps.find((app) => app.id == rid);

			if (!a) {
				console.log('app not found ', rid);
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
				show_dot
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
			const tokenStore = useTokenStore();

			this.myApps = this.myApps.filter((item) => item.fatherName !== app_name);
			this.relocate_application_place(this.myApps);
			await axios.post(tokenStore.url + '/api/app/uninstall/' + app_name, data);
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

			if (uninstalledAppState(state)) {
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
		}
	}
});
