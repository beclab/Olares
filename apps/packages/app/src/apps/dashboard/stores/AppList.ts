import { AppListItem } from '@apps/control-panel-common/src/network/network';
import { defineStore } from 'pinia';
import { getAppsList } from '@apps/dashboard/src/network';

interface AppDetailActions {
	getAppList(): void;
}

interface AppDetailState {
	apps: AppListItem[] | [];
}
export const useAppList = defineStore<
	'appList',
	AppDetailState,
	any,
	AppDetailActions
>('appList', {
	state: (): AppDetailState => ({
		apps: []
	}),
	getters: {
		appsWithNamespace: (state) => state.apps.filter((item) => item.entrances)
	},
	actions: {
		async getAppList() {
			return getAppsList().then((res) => {
				this.apps = res.data.data;
			});
		}
	}
});
