import { AppDetailResponse } from '@apps/control-panel-common/src/network/network';
import { defineStore } from 'pinia';
import { getAppDetail } from '@apps/dashboard/src/network';
import { GLOBAL_ROLE } from '@apps/control-panel-common/src/constant/user';

const initData = {
	clusterRole: '',
	config: {},
	ksConfig: {},
	user: {
		email: '',
		globalrole: '',
		grantedClusters: [],
		lang: '',
		lastLoginTime: '',
		username: '',
		globalRules: {}
	},
	workspaces: [],
	systemNamespaces: []
};
interface AppDetailState {
	data: AppDetailResponse;
}
export const useAppDetailStore = defineStore('appDetail', {
	state: (): AppDetailState => ({
		data: initData
	}),

	getters: {
		user: (state): AppDetailResponse['user'] => state.data.user,
		isAdmin: (state): boolean => state.data.user.globalrole === GLOBAL_ROLE
	},
	actions: {
		async init() {
			return getAppDetail().then((res) => {
				this.data = res.data;
			});
		},

		async setData(data: any) {
			this.data = data;
		}
	}
});
