import { defineStore } from 'pinia';
import axios from 'axios';
import { date } from 'quasar';
import { getFileType, getFileIcon } from '@bytetrade/core';
import { format } from '../utils/format';
import { useUserStore } from './user';
import { CommonFetch } from './../api';
import { seahub } from './../api';
import {
	SearchCategory,
	ServiceParamsType,
	ServiceType,
	TextSearchItem,
	SearchV2Type
} from 'src/utils/interface/search';
import { dataAPIs, filesIsV1 } from 'src/api/files';
import { common } from 'src/api/files';
import { useWebsocketManager2Store } from './websocketManager2';
import { AppInfo } from 'src/utils/interface/applications';
import { getApplication } from 'src/application/base';
import { useAppStore } from './desktop/app';
import { useTokenStore } from './token';
import { driveTypeBySearchMeta } from 'src/api/files/v2/common/common';
import { dataAPIs as dataAPIsV2 } from 'src/api/files/v2';
import { APP_STATUS } from 'src/constant/constants';
import { useAppAbilitiesStore } from './appAbilities';
import {
	searchInit,
	seachOSVersionLargeThan12_3,
	searchCancel,
	searchMore,
	syncSearch
} from 'src/api/common/search';

import { v4 as uuidv4 } from 'uuid';
import { uninstalledAppState } from 'src/constant/config';

const { humanStorageSize } = format;

interface SearchItemType {
	name: string;
	title: string;
	icon: string;
	type: SearchCategory;
	id: string;
	serviceType: ServiceType;
}

interface ChatMsg {
	text: string;
	messageId: string;
	conversationId: string;
	model: string;
	name: string;
}

export type SearchState = {
	can_input: boolean;
	waiting: boolean;
	conversationId: string | null;
	searchList: SearchItemType[];
	filesSearchItem: SearchItemType;
	filesV2SearchItem: SearchItemType;
	wiseSearchItem: SearchItemType;
	chatList: ChatMsg[];
	appList: AppInfo[];
	searchId: string | undefined;
	limit: number;
};

export const useSearchStore = defineStore('search', {
	state: () => {
		return {
			can_input: true,
			waiting: false,
			conversationId: null,
			chatList: [],
			filesSearchItem: {
				name: 'Files Search',
				title: 'Drive',
				icon: './app-icon/drive.svg',
				type: 'Command',
				id: 'drive',
				serviceType: ServiceType.Files
			},
			filesV2SearchItem: {
				name: 'Files Search',
				title: 'Drive',
				icon: './app-icon/drive.svg',
				type: 'Command',
				id: 'drive',
				serviceType: ServiceType.FilesV2
			},
			wiseSearchItem: {
				name: 'Wise Search',
				title: 'Wise',
				icon: './app-icon/wise.svg',
				type: 'Command',
				id: 'wise',
				serviceType: ServiceType.Knowledge
			},
			searchList: [
				// {
				// 	name: 'Files Search',
				// 	title: 'Sync',
				// 	icon: './app-icon/sync.svg',
				// 	type: 'Command',
				// 	id: 'sync',
				// 	serviceType: ServiceType.Sync
				// }
				// {
				// 	name: 'Google Drive',
				// 	title: 'Files Search',
				// 	icon: './app-icon/gddrive.svg',
				// 	type: 'Command',
				// 	id: 'gddrive',
				// 	serviceType: ServiceType.Sync
				// },
				// {
				// 	name: 'Dropbox',
				// 	title: 'Files Search',
				// 	icon: './app-icon/dropbox.svg',
				// 	type: 'Command',
				// 	id: 'dropbox',
				// 	serviceType: ServiceType.Sync
				// },
			],
			appList: [],
			searchId: undefined,
			limit: 20
		} as SearchState;
	},
	getters: {},
	actions: {
		getCommand() {
			let myApps: AppInfo[] = [];
			let wiseInstalled = false;
			if (getApplication().applicationName == 'desktop') {
				const appStore = useAppStore();
				myApps = appStore.myApps;
				if (
					myApps.find(
						(e) =>
							(e.fatherName == 'wise' || e.name == 'wise') &&
							e.state == APP_STATUS.RUNNING
					)
				) {
					wiseInstalled = true;
				}
			} else {
				myApps = this.appList;
				const abilitiesStore = useAppAbilitiesStore();
				wiseInstalled =
					abilitiesStore.wise.running && abilitiesStore.wise.id.length > 0;
			}
			const apps: any[] = [];
			myApps.map((item) => {
				item.type = SearchCategory.Application;
				if (item.id !== 'launchpad') {
					apps.push(item);
				}
			});

			let res = [...this.searchList];
			if (seachOSVersionLargeThan12_3()) {
				res.push({
					name: 'Files Search',
					title: 'Sync',
					icon: './app-icon/sync.svg',
					type: SearchCategory.Command,
					id: 'sync',
					serviceType: ServiceType.Sync
				});
			}
			if (wiseInstalled) {
				res.push(this.wiseSearchItem);
			}
			res = [...res, ...apps];

			// , ...myApps
			if (filesIsV1()) {
				res.splice(0, 0, this.filesSearchItem);
			} else {
				res.splice(0, 0, this.filesV2SearchItem);
			}

			console.log('getCommandres', res);
			return res;
		},

		async getAppList() {
			try {
				const data: any = await CommonFetch.post(
					this.getBaseurl() + '/server/myApps',
					{}
				);

				const myApps: AppInfo[] = [];

				for (let i = 0; i < data.data.data.length; i++) {
					const app = data.data.data[i];
					if (uninstalledAppState(app.state)) {
						continue;
					}
					if (!app.entrances || app.entrances.length == 0) {
						continue;
					}

					for (let j = 0; j < app.entrances.length; ++j) {
						const entrance = app.entrances[j];

						if (!entrance.invisible) {
							const state =
								app.state === 'running' ? entrance.state : app.state;
							myApps.push({
								...app,
								state,
								type: SearchCategory.Application,
								id: entrance.id,
								appid: app.id,
								name: entrance.name,
								title: entrance.title || app.title,
								url: entrance.url || app.url,
								icon: entrance.icon || app.icon,
								fatherName: app.name,
								openMethod: entrance.openMethod || 'default',
								isSysApp: app.isSysApp,
								fatherState: app.state
							});
						}
					}
				}
				this.appList = myApps;
			} catch (error) {
				return [];
			}
		},

		async getFiles(query = '') {
			const res: any = await axios.post(this.getBaseurl() + '/server/query', {
				query,
				limit: this.limit
			});
			if (typeof res === 'object' && res.length === 0) {
				return res;
			} else {
				const items = res.items;

				items.map((item: any) => {
					item.fileType = getFileType(item.name);
					item.fileIcon = getFileIcon(item.name);
					item.size = humanStorageSize(item.size);
					item.created = date.formatDate(
						item.created * 1000,
						'MMM Do YYYY, HH:mm:ss'
					);
					item.modified = date.formatDate(
						item.modified * 1000,
						'MMM Do YYYY, HH:mm:ss'
					);
				});
				return items;
			}
		},

		async getContent(
			query = '',
			serviceType: ServiceType
		): Promise<TextSearchItem[]> {
			const params: ServiceParamsType = {
				query,
				serviceType: serviceType,
				limit: this.limit,
				offset: 0
			};

			if (serviceType === ServiceType.Sync) {
				if (seachOSVersionLargeThan12_3()) {
					return await syncSearch({ query });
				}
				const repo_list = await this.getSyncRepoList();
				const res = await this.fetchSyncData(params, repo_list, query);
				return res;
			}

			if (seachOSVersionLargeThan12_3()) {
				if (this.searchId) {
					searchCancel(this.searchId);
				}
				this.searchId = uuidv4();
				return await searchInit(
					{
						reqid: this.searchId || '',
						keyword: query,
						type: SearchV2Type.AGGREGATE,
						app: serviceType
					},
					0,
					this.limit
				);
			}

			return await this.fetchData(params);
		},

		async fetchData(params: ServiceParamsType) {
			const res: TextSearchItem[] = await axios.post(
				this.getBaseurl() + '/server/search',
				params
			);
			const newRes: TextSearchItem[] = [];

			for (let i = 0; i < res.length; i++) {
				const el = res[i];
				el.fileType = getFileType(el.title);
				el.fileIcon = getFileIcon(el.title);
				if (el.resource_uri) {
					if (filesIsV1()) {
						el.driveType = common().driveTypeBySearchPath(el.resource_uri);
					} else {
						el.driveType = driveTypeBySearchMeta({
							fileType: el.meta?.file_type as any,
							fileExtend: el.meta?.extend
						});
					}
					if (el.driveType) {
						if (filesIsV1()) {
							el.isDir = el.resource_uri.endsWith('/');
							const path = dataAPIs(el.driveType).formatSearchPath(
								el.resource_uri
							);
							el.path = path;
						} else {
							if (el.meta) {
								if (el.meta.is_dir) {
									el.isDir = el.meta.is_dir;
								}
								const path = dataAPIsV2(el.driveType).displayPath({
									isDir: el.meta.is_dir,
									fileExtend: el.meta.extend,
									path: el.meta.path,
									fileType: el.meta.file_type
								});
								el.path = path;
							}
						}
					} else {
						el.path = el.resource_uri;
					}
				}

				newRes.push(el);
			}

			return newRes;
		},

		async fetchSyncData(
			params: ServiceParamsType,
			repo_list: any[],
			query: string
		) {
			const results: TextSearchItem[] = [];
			for (let j = 0; j < repo_list.length; j++) {
				const repo_item = repo_list[j];
				try {
					params.repo_id = repo_item.repo_id;
					const res: any = await axios.post(
						this.getBaseurl() + '/server/search',
						params
					);

					if (res && res.length > 0) {
						const resArr: TextSearchItem[] = [];
						for (let i = 0; i < res.length; i++) {
							const el = res[i];
							el.repo_id = repo_item.repo_id;
							el.repo_name = repo_item.repo_name;
							const id = `id_${j}_${i}`;
							resArr.push(this.formatSyncToSearch(id, el, query));
						}

						results.push(...resArr);
					}
				} catch (error) {
					console.log('error', error);
				}
			}
			return results;
		},

		formatSyncToSearch(id: string, data: any, query: string) {
			const lastIndex = data.path.lastIndexOf('/');
			const path = lastIndex !== -1 ? data.path.slice(0, lastIndex) : data.path;
			const name =
				lastIndex !== -1 ? data.path.slice(lastIndex + 1) : data.path;
			const fileType = getFileType(name) || 'blob';
			const fileIcon = getFileIcon(name);

			const searchRes: TextSearchItem = {
				id: id,
				highlight: name.replace(query, `<hi>${query}</hi>`),
				highlight_field: '',
				title: name,
				fileType: fileType,
				fileIcon: fileIcon || 'other',
				repo_id: data.repo_id,
				path: '/' + data.repo_name + path,
				// meta: {
				// 	updated: new Date(data.mtime).getTime() / 1000
				// },
				isDir: data.type === 'file' ? false : true
			};
			return searchRes;
		},

		async getSyncRepoList() {
			if (getApplication().applicationName == 'desktop') {
				const tokenStore = useTokenStore();
				const res: any = await axios.get(
					tokenStore.url + '/seahub/api/v2.1/repos/?type=mine'
				);
				return res.repos;
			}
			const res: any = await seahub().fetchMineRepo();
			return res;
		},

		async getPath(resource_uri: string) {
			if (resource_uri.startsWith('/data')) {
				const splitUri = resource_uri.split('/');
				const indexUri = splitUri.findIndex(
					(item) => item.indexOf('pvc-') > -1
				);
				let res = '';
				for (let i = indexUri + 1; i < splitUri.length - 1; i++) {
					const el = splitUri[i];
					res += '/' + el;
				}
				return res;
			} else {
				return resource_uri;
			}
		},
		async sendChat(
			conversationId: string,
			message = '',
			path = '',
			type = 'basic'
		) {
			const qMessage = {
				text: message,
				messageId: '',
				conversationId: '',
				model: '',
				name: 'Me'
			};
			this.chatList.push(qMessage);
			this.can_input = false;
			this.waiting = true;
			this.conversationId = null;

			const socketStore = useWebsocketManager2Store();
			socketStore.send({
				event: 'ai',
				data: {
					message,
					conversationId,
					path,
					type
				}
			});
		},
		initChat() {
			this.chatList = [];
			this.can_input = true;
			this.waiting = false;
		},
		getBaseurl() {
			const userStore = useUserStore();
			return process.env.DEV ? '' : userStore.getModuleSever('desktop');
		},
		async searchMore(offset: number): Promise<TextSearchItem[]> {
			return await searchMore(
				{
					reqid: this.searchId || ''
				},
				offset,
				this.limit
			);
		},
		async cancelSearch() {
			if (!!this.searchId) {
				searchCancel(this.searchId);
				this.searchId = undefined;
			}
		}
	}
});
