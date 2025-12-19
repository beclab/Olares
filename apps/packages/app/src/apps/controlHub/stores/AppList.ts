import { AppListAllResponse } from '@apps/control-panel-common/src/network/network';
import { defineStore } from 'pinia';
import { getAppsListAll } from '@apps/control-hub/src/network';

interface AppDetailState {
	data: AppListAllResponse['data'];
}
export const useAppList = defineStore('appList', {
	state: (): AppDetailState => ({
		data: {}
	}),
	getters: {},
	actions: {
		async init() {
			return getAppsListAll().then((res) => {
				this.data = res.data.data;
			});
		}
	}
});
