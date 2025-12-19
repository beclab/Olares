import { defineStore } from 'pinia';
import { useTokenStore } from './token';
import { notifySuccess } from 'src/utils/settings/btNotify';
import { i18n } from 'src/boot/i18n';
import axios from 'axios';
import { useApplicationStore } from './application';

export enum ActionType {
	APPLIED = 'applied',
	APPLYING = 'applying'
}

export interface AclInfo {
	proto: string;
	dst: string[];
}

export type AclState = {
	allow_ssh: boolean;
	state: ActionType;
	appAclList: AclInfo[];
	allPortAppAclList: PortAclInfo[];
	allAppAclList: AppAclInfo[];
	allow_subroutes: boolean;
	subroutes: string[];
};

export enum PortStatus {
	Normal = 0,
	Remove = -1,
	Add = 1
}

export interface PortAclInfo {
	port: string;
	appName: string;
	appOwner: string;
	proto: string;
	status: PortStatus; //0 normal -1 remove 1 add
	appTitle: string;
}

export interface AppAclInfo {
	appName: string;
	appOwner: string;
	proto: string;
	dst: string[];
}

export const useAclStore = defineStore('acl', {
	state: () => {
		return {
			allow_ssh: false,
			state: ActionType.APPLYING,
			appAclList: [],
			allPortAppAclList: [],
			allAppAclList: [],
			subroutes: [],
			allow_subroutes: false
		} as AclState;
	},
	getters: {
		displayPort(state) {
			return state.allPortAppAclList.filter(
				(e) => e.status != PortStatus.Remove
			);
			// return state.allPortAppAclList;
		}
	},
	actions: {
		async getAclStatus() {
			const tokenStore = useTokenStore();

			const data: any = await axios.get(tokenStore.url + '/api/acl/ssh/status');
			this.allow_ssh = data.allow_ssh;
			this.state = data.state;
		},

		async toggleAclStatus(open: boolean) {
			if (this.state === ActionType.APPLYING) {
				return;
			}
			const tokenStore = useTokenStore();
			let requestUrl = '';
			if (open) {
				requestUrl = '/api/acl/ssh/enable';
			} else {
				requestUrl = '/api/acl/ssh/disable';
			}

			await axios.post(tokenStore.url + requestUrl, {});
			notifySuccess(i18n.global.t('success'));
		},
		//jump boot axios
		async getAppAclStatus(name: string) {
			this.appAclList = [];
			const tokenStore = useTokenStore();

			const data: any = await axios.get(
				tokenStore.url + '/api/acl/app/status?name=' + name
			);
			console.log(data);
			if (data.code === 0) {
				this.appAclList = data.data;
			} else {
				this.appAclList = [];
			}
		},
		async updateAppAclStatus(
			acls: {
				proto: string;
				dst: string[];
			}[],
			name: string
		) {
			const tokenStore = useTokenStore();
			await axios.post(tokenStore.url + '/api/acl/app/status', {
				name,
				acls: acls.filter((e) => e.dst.length > 0)
			});
			notifySuccess(i18n.global.t('success'));
			await this.getAllApplicationAcls();
		},
		async deleteAppAclStatus(name: string, proto: string, dst: string) {
			const app = this.allAppAclList.find(
				(e) => e.appName === name && e.proto == proto
			);
			if (!app || !app.dst.includes(dst)) {
				return;
			}
			const dstList = app.dst.filter((e) => e != dst);

			let acls: {
				proto: string;
				dst: string[];
			}[] = [];

			const sameApps = this.allAppAclList
				.filter((e) => e.appName === name && e.proto != proto)
				.map((e) => {
					return {
						proto: e.proto,
						dst: e.dst
					};
				});

			acls = acls.concat(sameApps);
			acls.push({
				proto: proto,
				dst: dstList
			});
			await this.updateAppAclStatus(acls, name);
		},
		async addAppAclStatus(name: string, proto: string, dst: string) {
			const app = this.allAppAclList.find(
				(e) => e.appName === name && e.proto == proto
			);
			let dstList = [dst];
			if (app) {
				if (app.dst.includes(dst)) {
					return;
				}
				dstList = [...app.dst];
				dstList.push(dst);
			}
			let acls: {
				proto: string;
				dst: string[];
			}[] = [];

			const sameApps = this.allAppAclList
				.filter((e) => e.appName === name && e.proto != proto)
				.map((e) => {
					return {
						proto: e.proto,
						dst: e.dst
					};
				});

			acls = acls.concat(sameApps);

			acls.push({
				proto: proto,
				dst: dstList
			});

			await this.updateAppAclStatus(acls, name);
		},

		async appAclSubmit(name: string) {
			const filterAppsPorts = this.allPortAppAclList.filter(
				(e) => e.appName === name
			);
			const acls: {
				proto: string;
				dst: string[];
			}[] = [];

			filterAppsPorts.forEach((port) => {
				if (port.status == PortStatus.Remove) {
					return;
				}
				const acl = acls.find((e) => e.proto == port.proto);
				if (acl) {
					acl.dst.push(port.port);
				} else {
					acls.push({
						proto: port.proto,
						dst: [port.port]
					});
				}
			});
			await this.updateAppAclStatus(acls, name);
		},

		async getAllApplicationAcls() {
			const tokenStore = useTokenStore();
			const appStore = useApplicationStore();
			const data: any = await axios.get(tokenStore.url + '/api/acl/all');
			this.allAppAclList = data;
			const appAcls: PortAclInfo[] = [];
			if (data && data.length > 0) {
				data.forEach(
					(app: {
						dst?: string[];
						appName: string;
						appOwner: string;
						proto: string;
					}) => {
						if (app.dst && app.dst.length > 0) {
							const appInfo = appStore.getApplicationById(app.appName);
							app.dst.forEach((dst: string) => {
								appAcls.push({
									appName: app.appName,
									port: dst,
									appOwner: app.appOwner,
									proto: app.proto,
									status: PortStatus.Normal,
									appTitle: appInfo ? appInfo.title : app.appName
								});
							});
						}
					}
				);
			}
			this.allPortAppAclList = appAcls;
		},
		async getSubroutesStatus() {
			const tokenStore = useTokenStore();
			const data: any = await axios.get(
				tokenStore.url + '/api/acl/subroutes/status',
				{
					timeout: 15 * 1000
				}
			);
			this.subroutes = data;
			this.allow_subroutes = data && data.length > 0;
		},
		async toggleSubroutesStatus(open: boolean) {
			const tokenStore = useTokenStore();
			let requestUrl = '';
			if (open) {
				requestUrl = '/api/acl/subroutes/enable';
			} else {
				requestUrl = '/api/acl/subroutes/disable';
			}
			await axios.post(tokenStore.url + requestUrl, {});
			let retryCount = 0;
			let getStatusSuccess = false;
			while (retryCount < 3 && !getStatusSuccess) {
				try {
					await this.getSubroutesStatus();
					getStatusSuccess = true;
					notifySuccess(i18n.global.t('success'));
				} catch (error) {
					if (retryCount < 3) {
						retryCount++;
					}
					console.log(error);
				}
			}
		}
	}
});
