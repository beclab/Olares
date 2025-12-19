import { AppListItem } from '@apps/control-panel-common/src/network/network';
import { defineStore } from 'pinia';
import { getAppsList } from 'src/api/bex';
import { storage } from 'src/utils/bex/storage';
import desktopIcon from 'src/assets/plugin/desktop.svg';
import { openUrl } from 'src/utils/bex/tabs';
import { useUserStore } from 'src/stores/user';
import { uninstalledAppState } from 'src/constant/config';

const getUserDid = () => {
	const userStore = useUserStore();
	const userInfo = userStore.terminusInfo();
	return userInfo.did;
};

const updateAppsIds = async (data) => {
	const localData = await storage.get('foregroundAppsIds');
	const storageData = {
		...localData,
		[getUserDid()]: data
	};
	storage.set('foregroundAppsIds', storageData);
};

interface AppDetailState {
	allApps: AppListItem[];
	allAppIds: string[];
	foregroundAppsIds: string[];
	foregroundAppsIdsCache: string[];
	loading: boolean;
}

const olaresDesktop = (domain: string) => {
	const url = `desktop.${domain}`;
	return {
		id: 'desktop',
		name: 'desktop',
		namespace: '',
		deployment: '',
		owner: '',
		url,
		icon: desktopIcon,
		title: 'Desktop',
		target: '',
		entrances: [
			{
				id: 'desktop',
				name: 'desktop',
				title: 'Desktop',
				url,
				icon: desktopIcon,
				invisible: false
			}
		],
		state: 'running',
		isSysApp: true,
		isClusterScoped: false
	};
};

const others = (): AppListItem[] => {
	const userStore = useUserStore();
	const userInfo = userStore.terminusInfo();
	const domain = (userInfo?.olaresId || userInfo?.terminusName)?.replace(
		'@',
		'.'
	);
	const desktop = olaresDesktop(domain);
	return [desktop];
};

const defualtForegroundAppIds = ['desktop', 'files', 'market', 'vault'];

export const useAppsStore = defineStore('appsStore', {
	state: (): AppDetailState => ({
		allApps: [],
		allAppIds: [],
		foregroundAppsIds: [],
		foregroundAppsIdsCache: [],
		loading: false
	}),
	getters: {
		foregroundApps: (state) =>
			state.foregroundAppsIds.map((item) => {
				const target = state.allApps.find((app) => app.id === item);
				return target!;
			}),

		backgroundApps: (state) =>
			state.allApps
				.filter((item) => !state.foregroundAppsIds.includes(item.id))
				.map((item) => ({
					...item,
					title: item.title || item.entrances?.[0]?.title || ''
				}))
	},
	actions: {
		async init() {
			const localData = await storage.get('foregroundAppsIds');
			let appIds = defualtForegroundAppIds;
			try {
				appIds = Object.values(localData[getUserDid()]);
			} catch (error) {
				console.error('Failed to get foreground app IDs from storage', error);
			}
			try {
				if (this.allApps.length <= others().length) {
					this.loading = true;
					const res = await getAppsList();
					const items = getRunningApps(res.data.data);
					this.allApps = others().concat(items);
					this.allAppIds = this.allApps.map((item) => item.id);
					this.foregroundAppsIds = appIds.filter((item) =>
						this.allAppIds.includes(item)
					);
					this.foregroundAppsIdsCache = [...this.foregroundAppsIds];
				}
			} catch (error) {
				console.error('Failed to fetch app list', error);
			}

			this.loading = false;
		},
		resetAppList() {
			this.allApps = [...others()];
		},
		deleteApp(id: string) {
			this.foregroundAppsIds = this.foregroundAppsIds.filter(
				(item) => item !== id
			);
			updateAppsIds(this.foregroundAppsIds);
		},
		async addApp(id: string) {
			this.foregroundAppsIds.push(id);
			updateAppsIds(this.foregroundAppsIds);
		},
		async updateForegroundAppsIds(id: string) {
			const targetId = this.foregroundAppsIds.find((item) => item === id);
			if (targetId) {
				this.deleteApp(id);
			} else {
				this.addApp(id);
			}
		},
		async sortApp(ids: string[]) {
			this.foregroundAppsIds = [...ids];
			updateAppsIds(this.foregroundAppsIds);
		},
		async appActionCancel() {
			this.foregroundAppsIds = [...this.foregroundAppsIdsCache];
			updateAppsIds(this.foregroundAppsIds);
		},
		async appActionSave() {
			this.foregroundAppsIdsCache = [...this.foregroundAppsIds];
			updateAppsIds(this.foregroundAppsIds);
		},
		openUrl: openUrl
	}
});

function getRunningApps(items) {
	const data = items;
	const list: any[] = [];
	for (let i = 0; i < data.length; ++i) {
		if (uninstalledAppState(data[i].state)) {
			continue;
		}

		if (!data[i].entrances || data[i].entrances.length == 0) {
			continue;
		} else {
			for (let j = 0; j < data[i].entrances.length; ++j) {
				if (data[i].entrances[j].invisible) {
					// do nothing
				} else {
					const state =
						data[i].state === 'running'
							? data[i].entrances[j].state
							: data[i].state;
					list.push({
						...data[i].entrances[j],
						state
					});
				}
			}
		}
	}
	return list;
}
