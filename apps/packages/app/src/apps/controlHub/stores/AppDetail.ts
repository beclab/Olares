import { AppDetailResponse } from '@apps/control-panel-common/src/network/network';
import { defineStore } from 'pinia';
import { getAppDetail } from '@apps/control-hub/src/network';
import { GLOBAL_ROLE } from '@apps/control-panel-common/src/constant/user';
import { isArray } from 'lodash';

const NAMESPACE_SHARED = '-shared';
const OS_PROTECTED = 'os-protected';

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
		isAdmin: (state): boolean => state.data.user.globalrole === GLOBAL_ROLE,
		isDemo: (state): boolean =>
			!!process.env.DEMO && !(state.data.user.globalrole === GLOBAL_ROLE),
		systemNamespaces: (state): AppDetailResponse['systemNamespaces'] =>
			state.data.systemNamespaces || []
	},
	actions: {
		async init() {
			return getAppDetail().then((res) => {
				this.data = res.data;
			});
		},

		async setData(data: any) {
			this.data = data;
		},
		hasPermission(data: string | string[]) {
			const value = isArray(data) ? data[0] : data;

			// Some permissions of demo station are higher than ordinary users
			const hasHigherDemoPrivileges = isArray(data) ? data[1] : false;

			if (!value) {
				return !this.isDemo;
			} else {
				return this.isAdmin
					? value.includes(this.data.user.username) ||
							this.systemNamespaces.includes(value) ||
							value.endsWith(NAMESPACE_SHARED) ||
							value === OS_PROTECTED
					: !!process.env.DEMO && !hasHigherDemoPrivileges
					? false
					: value.includes(this.data.user.username);
			}
		}
	}
});
