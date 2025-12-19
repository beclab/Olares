import { defineStore } from 'pinia';
import { useUserStore } from './user';
import { i18n } from 'src/boot/i18n';
import {
	BaseCollectInfo,
	DownloadProgress,
	RssInfo,
	RssStatus
} from '../pages/Mobile/collect/utils';
import { saveEntry, urlCheck } from '../api/wise';
import { notifyFailed, notifySuccess } from '../utils/notifyRedefinedUtil';
import { binaryInsert } from 'src/utils/utils';
import axios, { CancelToken, CancelTokenSource } from 'axios';
import {
	deleteFeeds,
	downloadFile,
	DownloadRecordListRequest,
	getDownloadHistory
} from '../api/wise';
import {
	FileInfo,
	DOWNLOAD_RECORD_STATUS,
	COOKIE_LEVEL
} from '../utils/rss-types';
import { DownloadRecord } from 'src/utils/interface/rss';
import { CompareDownloadRecord } from '../utils/rss-utils';
import { useAppAbilitiesStore } from './appAbilities';
import { BtNotify, NotifyDefinedType } from '@bytetrade/ui';
import { useCollectSiteStore } from './collect-site';
import { DownloadFileRecord } from 'src/platform/interface/bex/rss/utils';
import { saveFeed } from 'src/platform/interface/bex/rss/utils/feed';

export type DataState = {
	baseUrl: string;
	pagesList: RssInfo[];
	rssList: RssInfo[];
	filesDownloadHistoryList: DownloadRecord[];
	waitingQueue: DownloadProgress[];
	queryTimer: any;
	filesList: DownloadFileRecord[];
	loading: boolean;
	currentRequestId: number | null;
};
const CancelToken = axios.CancelToken;
let pageListSource: undefined | CancelTokenSource = undefined;

export const useCollectStore = defineStore('collect', {
	state: () => {
		return {
			baseUrl: '',
			pagesList: [],
			rssList: [],
			filesDownloadHistoryList: [],
			waitingQueue: [],
			downloadTasks: {},
			filesList: [],
			queryTimer: null,
			loading: false,
			currentRequestId: null
		} as DataState;
	},
	getters: {
		me_name(): string | undefined {
			const userStore = useUserStore();
			if (userStore.current_user) {
				return userStore.current_user.name;
			} else {
				return undefined;
			}
		}
	},
	actions: {
		async setList(pagesList: BaseCollectInfo[]) {
			const appAbilitiesStore = useAppAbilitiesStore();
			this.baseUrl = appAbilitiesStore.getAppDomain('wise');
			this.pagesList = pagesList;
			this.setPageListStatus();
		},
		async setRssList(list: RssInfo[]) {
			this.rssList = list;
		},
		async deleteRss(feed_url: string[]) {
			try {
				//temp false
				await deleteFeeds(feed_url, false);
				this.rssList = this.rssList.map((item) => {
					if (item.url === feed_url[0]) {
						return { ...item, status: RssStatus.none };
					}
					return item;
				});
				notifySuccess(
					i18n.global.t('remove_feed_success'),
					'collection-feed-success'
				);
			} catch (e) {
				notifyFailed(
					i18n.global.t('remove_feed_fail') + ' ' + e?.message,
					'collection-feed-fail'
				);
			}
		},
		getRssCollectTarget(items, url) {
			return items.find((e) => e.url === url);
		},
		setPageListStatus() {
			const requestId = Date.now();
			this.currentRequestId = requestId;
			pageListSource && pageListSource.cancel();
			pageListSource = CancelToken.source();
			const pageList: RssInfo[] = [];
			const collectSiteStore = useCollectSiteStore();

			for (const page of this.pagesList) {
				const data = page as RssInfo;
				this.loading = true;
				urlCheck(data.url, pageListSource.token)
					.then((response) => {
						if (response) {
							data.status = response?.entry_exist
								? RssStatus.added
								: RssStatus.none;
							data.cookie = response;

							try {
								const cookieRequire =
									response.cookie_require as COOKIE_LEVEL.RECOMMEND;
								const cookieExist = response.cookie_exist;
								const cookieExpired = response.cookieExpired;
								const is_entry_available = response.is_entry_available;
								collectSiteStore.data.cookie = {
									...collectSiteStore.cookie,
									cookieRequire,
									cookieExist,
									cookieExpired,
									is_entry_available
								};
							} catch (error) {
								//
							}
						} else {
							data.status = RssStatus.none;
						}
						pageList.push({ ...data, id: response?.exist_entry_id });
						this.pagesList = pageList;
					})
					.catch((error) => {
						if (!axios.isCancel(error)) {
							BtNotify.show({
								type: NotifyDefinedType.FAILED,
								message: error
							});
						}
					})
					.finally(() => {
						if (this.currentRequestId === requestId) {
							this.loading = false;
						}
					});
			}
		},
		setFileList(filesList: FileInfo[]) {
			this.filesList = [];
			for (const file of filesList) {
				const record = this.getDownloadRecordByUrl(file.download_url);
				this.filesList.push({
					file,
					record,
					title: file.file
				});
			}
		},
		async addFeed(item: RssInfo) {
			try {
				await saveFeed({
					feed_url: item.url,
					source: 'wise'
				});
				item.status = RssStatus.added;
				notifySuccess(
					i18n.global.t('add_feed_success'),
					'collection-feed-success'
				);
			} catch (e) {
				console.log(e.message);
				item.status = RssStatus.removed;
				notifyFailed(
					i18n.global.t('add_feed_fail') + e.message,
					'collection-feed-fail'
				);
			}
		},

		async addEntry(item: RssInfo) {
			try {
				const data = await saveEntry(
					[
						{
							url: item.url,
							source: 'library'
						}
					],
					this.baseUrl
				);
				item.status = RssStatus.added;
				const target = data[0];
				if (target) {
					item.id = target.id;
				}
				notifySuccess(
					i18n.global.t('add_page_success'),
					'collection-page-success'
				);
			} catch (e) {
				console.log(e.message);
				item.status = RssStatus.removed;
				notifyFailed(
					i18n.global.t('add_page_fail') + e.message,
					'collection-page-fail'
				);
			}
		},

		async addDownloadRecord(info: FileInfo) {
			const record = await downloadFile(
				info.file,
				info.file_type,
				info.download_url
			);
			if (record) {
				const index = this.filesDownloadHistoryList.findIndex(
					(l) => l.id == record.id
				);
				if (index >= 0) {
					this.filesDownloadHistoryList.splice(index, 1, record);
				} else {
					binaryInsert<DownloadRecord>(
						this.filesDownloadHistoryList,
						record,
						CompareDownloadRecord
					);
				}
				const currentRecordIndex = this.filesList.findIndex(
					(file) =>
						(!file.record && file.file.download_url == record.url) ||
						(file.record && file.record.id == record.id)
				);
				if (currentRecordIndex >= 0) {
					this.filesList.splice(index, 1, {
						...this.filesList[currentRecordIndex],
						record
					});
				}
				this.startInterval();
			}
			return !!record;
		},

		startInterval() {
			if (this.queryTimer) {
				return;
			}
			const needQuery = (list: DownloadRecord[]) => {
				const downloading = list.filter(
					(item) =>
						item.status === DOWNLOAD_RECORD_STATUS.WAITING ||
						item.status === DOWNLOAD_RECORD_STATUS.DOWNLOADING
				);

				return downloading.length !== 0;
			};
			if (!needQuery(this.filesDownloadHistoryList)) {
				return;
			}
			this.queryTimer = setInterval(async () => {
				const list = await this.getDownloadHistory(
					0,
					this.filesDownloadHistoryList.length
				);
				if (!needQuery(list)) {
					clearInterval(this.queryTimer);
					this.queryTimer = null;
					console.log('clear');
					return;
				}
			}, 5000);
		},
		async getDownloadHistory(offset: number, limit: number) {
			const req = new DownloadRecordListRequest(
				undefined,
				undefined,
				undefined,
				offset,
				limit
			);
			const list = await getDownloadHistory(req);
			if (list) {
				list.forEach((record) => {
					const index = this.filesDownloadHistoryList.findIndex(
						(l) => l.id == record.id
					);

					if (index >= 0) {
						this.filesDownloadHistoryList.splice(index, 1, record);
					} else {
						binaryInsert<DownloadRecord>(
							this.filesDownloadHistoryList,
							record,
							CompareDownloadRecord
						);
					}

					const currentRecordIndex = this.filesList.findIndex(
						(file) =>
							(!file.record && file.file.download_url == record.url) ||
							(file.record && file.record.id == record.id)
					);
					if (currentRecordIndex >= 0) {
						this.filesList.splice(index, 1, {
							...this.filesList[currentRecordIndex],
							record
						});
					}
				});
			}
			return list ? list : [];
		},
		getDownloadRecordByUrl(url: string) {
			const item = this.filesDownloadHistoryList.find((l) => l.url == url);
			return item;
		},
		updateCookieStatus() {
			if (this.pagesList[0]?.cookie) {
				this.pagesList[0].cookie.cookieExpired = false;
				this.pagesList[0].cookie.cookie_exist = true;
			}
		},
		resetCookieStatus() {
			if (this.pagesList[0]?.cookie) {
				this.pagesList[0].cookie.cookieExpired = false;
				this.pagesList[0].cookie.cookie_exist = false;
			}
		}
	}
});
