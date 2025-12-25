import { defineStore } from 'pinia';
import { getCollectInfo } from 'src/api/common/knowledge';
import { AbilityData } from 'src/core/abilities';
import {
	CollectInfo,
	DownloadItem,
	DownloadStatusEnum
} from 'src/types/commonApi';
import { useRssStore } from './rss';
import { BtNotify, NotifyDefinedType } from '@bytetrade/ui';
import { t } from 'src/boot/i18n';
import { COOKIE_LEVEL } from 'src/utils/rss-types';
import { BaseSiteCardProps } from 'src/components/collection/collect';
import { downloadFileNew } from 'src/api/wise/download';
import { busOn } from 'src/utils/bus';
import axios, { CancelTokenSource } from 'axios';
const CancelToken = axios.CancelToken;
let CollectSearchCancelTokenSource: undefined | CancelTokenSource = undefined;

interface SiteInfo {
	title: string;
	icon: string;
	url: string;
}

interface AppAbilitiesStore {
	data: CollectInfo & { web_site_info: SiteInfo };
	loading: boolean;
	collectLoading: boolean;
	currentRequestId: number | null;
	url?: string;
	cache: Map<string, CollectInfo & { web_site_info: SiteInfo }>;
}

const defaultData = {
	is_download_available: '',
	is_entry_available: '',
	is_feed_available: '',
	feed: [],
	download: {
		list: [],
		source: '',
		title: '',
		thumbnail: ''
	},
	entry: {
		exist: false,
		collect: false,
		url: '',
		title: '',
		file_type: '',
		thumbnail: '',
		entry: []
	},
	web_site_info: {
		title: '',
		url: '',
		icon: ''
	},
	cookie: {
		cookieRequire: COOKIE_LEVEL.RECOMMEND,
		cookieExist: false,
		cookieExpired: false,
		is_entry_available: ''
	}
};

export const useCollectSiteStore = defineStore('collectSite', {
	state: (): AppAbilitiesStore => ({
		data: { ...defaultData },
		loading: false,
		collectLoading: false,
		currentRequestId: null,
		url: '',
		cache: new Map()
	}),
	getters: {
		cookie: (state) => state.data.cookie || defaultData.cookie,
		feed: (state) => state.data.feed || defaultData.feed,
		download: (state) => {
			const data = state.data.download || defaultData.download;
			return data.list.map((item) => ({
				title: data.title,
				icon: data.thumbnail,
				url: item.download_url,
				...item
			}));
		},
		entry: (state) => state.data.entry || defaultData.entry,
		web_site_info: (state) =>
			state.data.web_site_info || defaultData.web_site_info,
		hasCache: (state) =>
			process.env.PLATFORM_BEX_ALL && state.url
				? state.cache.has(state.url)
				: false
	},
	actions: {
		syncCache() {
			if (process.env.PLATFORM_BEX_ALL && this.url) {
				this.cache.set(this.url, { ...this.data });
			}
		},
		async init() {
			busOn('wiseDownloadProcess', (message) => {
				const data = message.data;
				const target = this.getDownloadByTaskId(data.task_id);
				if (target) {
					target.file = data.name;
					if (data.percent && data.percent >= 100) {
						target.download_status = DownloadStatusEnum.COMPLETE;
						target.is_exist = true;
					} else {
						target.percent = data.percent;
					}
					this.syncCache();
				}
			});
		},
		async reset() {
			this.data = { ...defaultData };
			this.loading = false;
			this.collectLoading = false;
			this.url = undefined;
		},
		async search(url: string) {
			if (!url) {
				return;
			}
			this.url = url;

			if (process.env.PLATFORM_BEX_ALL) {
				const cachedData = this.cache.get(url);
				if (cachedData) {
					this.data = { ...cachedData };
					this.loading = false;
				} else {
					this.reset();
				}
			} else {
				this.reset();
			}

			this.loading = true;
			const requestId = Date.now();
			this.currentRequestId = requestId;
			try {
				CollectSearchCancelTokenSource &&
					CollectSearchCancelTokenSource.cancel();
				CollectSearchCancelTokenSource = CancelToken.source();

				const res = await getCollectInfo(
					url,
					CollectSearchCancelTokenSource.token
				);
				const web_site_info = {
					url: url,
					title: '',
					icon: ''
				};
				const newData = { ...res, web_site_info };

				this.data = newData;
				if (process.env.PLATFORM_BEX_ALL) {
					this.cache.set(url, newData);
				}
			} catch (error) {
				if (!axios.isCancel(error) && process.env.PLATFORM_BEX_ALL) {
					BtNotify.show({
						type: NotifyDefinedType.FAILED,
						message: error
					});
				}
			} finally {
				if (this.currentRequestId === requestId) {
					this.loading = false;
				}
			}
		},

		updateCookieStatus() {
			try {
				this.data.cookie.cookieExist = true;
				this.data.cookie.cookieExpired = false;
				this.syncCache();
			} catch (error) {
				//
			}
		},

		updateEntry(data: { title: string; url: string; thumbnail: string }) {
			this.data.entry = { ...this.data.entry, ...data };
			this.syncCache();
		},

		getFeedTarget(feed_url: string) {
			const index = this.data.feed.findIndex(
				(item) => item.feed_url === feed_url
			);
			return this.data.feed[index];
		},

		getDownloadById(id: string) {
			const index = this.data.download.list.findIndex((item) => item.id === id);
			return this.data.download.list[index];
		},

		getDownloadByTaskId(task_id: number) {
			const index = this.data.download.list.findIndex(
				(item) => item.task_id === task_id
			);
			return this.data.download.list[index];
		},

		async addFeed(feed_url: string) {
			const rssStore = useRssStore();

			const IS_BEX = process.env.IS_BEX || process.env.DEV_PLATFORM_BEX;
			const addFeed =
				process.env.APPLICATION === 'WISE'
					? rssStore.addFeed
					: rssStore.addFeedOnly;

			const index = this.data.feed.findIndex(
				(item) => item.feed_url === feed_url
			);
			const target = this.getFeedTarget(feed_url);
			if (target) {
				target.loading = true;
			}
			try {
				const res = await addFeed(feed_url);

				if (target) {
					target.is_subscribed = true;
					target.id = res?.id || '';
				}

				BtNotify.show({
					type: NotifyDefinedType.SUCCESS,
					message: t('dialog.add_rss_feed_success')
				});
				if (process.env.APPLICATION === 'WISE') {
					rssStore.syncEntries();
				}
				this.syncCache();
			} catch (error) {
				BtNotify.show({
					type: NotifyDefinedType.FAILED,
					message: t('dialog.add_rss_feed_failed')
				});
			}

			if (target) {
				target.loading = false;
				this.syncCache();
			}
		},

		async addCollect(url: string) {
			const rssStore = useRssStore();
			this.collectLoading = true;
			try {
				const addEntryFn =
					process.env.APPLICATION === 'WISE'
						? rssStore.addEntry
						: rssStore.addEntryOnlyRequest;
				const res = await addEntryFn(url);

				this.data.entry.exist = true;
				this.data.entry.exist_entry_id = res[0].id;
				this.syncCache();
				BtNotify.show({
					type: NotifyDefinedType.SUCCESS,
					message: t('dialog.add_article_success')
				});
				if (process.env.APPLICATION === 'WISE') {
					rssStore.syncEntries();
				}
			} catch (error) {
				BtNotify.show({
					type: NotifyDefinedType.FAILED,
					message: t('dialog.add_article_failed')
				});
			}
			this.collectLoading = false;
		},
		async downloadFile(item: DownloadItem & BaseSiteCardProps['data']) {
			const path = 'Downloads/';
			const resolution = item.resolution;
			const format_id = item.id;
			const params = {
				download_url: item.download_url,
				name: item.file,
				path,
				format_id,
				resolution,
				file_type: item.file_type
			};
			const target = this.getDownloadById(item.id);
			target.loading = true;
			try {
				const result = await downloadFileNew(params);
				target.download_status = DownloadStatusEnum.DOWNLOADING;
				target.task_id = result?.id;
				this.syncCache();
			} catch (error) {
				//
			}
			target.loading = true;
			this.syncCache();
		}
	}
});
