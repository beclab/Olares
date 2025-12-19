import { MenuChoice, MenuType, TabType } from '../utils/rss-menu';
import { THEME_TYPE } from 'src/utils/rss-types';
import { useReaderStore } from './rss-reader';
import { defineStore } from 'pinia';
import { Dark } from 'quasar';
import {
	ARTICLE_TOPIC_OPEN,
	ENTRY_REMOVE_WITH_FILE,
	FEED_REMOVE_WITH_FILE,
	RECOMMENDATION_OPEN,
	RIGHT_DRAWER_OPEN,
	STORAGE_FALSE,
	STORAGE_TRUE,
	TASK_REMOVE_WITH_FILE,
	THEME_SETTING,
	TREND_REMOVE_NOT_NOTIFY,
	UPLOAD_COOKIES_OPEN,
	UPLOAD_LINKS_OPEN
} from '../utils/localStorageConstant';
import { useRssStore } from 'src/stores/rss';
import { nextTick } from 'vue';

export type DataState = {
	menuInited: boolean;
	url: string;
	sdkUrl: string;
	menuChoice: MenuChoice;
	userTabs: Map<string, TabType[] | string[] | number[]>;
	lastSelectedTabs: Map<string, TabType | number | string>;
	leftDrawerOpen: boolean;
	articleTopicOpen: boolean;
	rightDrawerOpen: boolean;
	recommendationOpen: boolean;
	uploadCookiesOpen: boolean;
	uploadLinksOpen: boolean;
	trendRemoveNotNotify: boolean;
	entryRemoveWithFile: boolean;
	feedRemoveWithFile: boolean;
	taskRemoveWithFile: boolean;
	readingFontSize: number;
	readingFontFamily: string;
	readingContentPending: number;
	themeSetting: string;
};

export const useConfigStore = defineStore('config', {
	state: () => {
		return {
			menuInited: false,
			url: '',
			sdkUrl: '',
			menuChoice: {
				type: MenuType.Trend,
				tab: TabType.Empty,
				params: {}
			},
			userTabs: new Map(),
			lastSelectedTabs: new Map(),
			leftDrawerOpen: false,
			//settings
			articleTopicOpen: false,
			rightDrawerOpen: false,
			recommendationOpen: false,
			uploadCookiesOpen: false,
			uploadLinksOpen: false,
			trendRemoveNotNotify: false,
			entryRemoveWithFile: true,
			feedRemoveWithFile: true,
			taskRemoveWithFile: true,
			readingFontSize: 16,
			readingFontFamily: '',
			readingContentPending: 80,
			themeSetting: THEME_TYPE.AUTO
		} as unknown as DataState;
	},
	actions: {
		init(new_url: string, new_sdk: string) {
			this.url = new_url;
			this.sdkUrl = new_sdk;
			const themeLocal = localStorage.getItem(THEME_SETTING);
			if (themeLocal) {
				this.themeSetting = themeLocal;
			}
			this.setUserTabs(MenuType.Transmission, [
				TabType.Download,
				TabType.Upload
			]);
			const rssStore = useRssStore();
			this.setUserTabs(
				MenuType.Trend,
				rssStore.support_algorithm.map((item) => item.id)
			);
		},
		getModuleSever(module: string, protocol = 'https:', suffix = '') {
			let url = protocol + '//';
			const parts = window.location.hostname.split('.');
			if (parts.length > 1 && module && module.length > 0) {
				parts[0] = module;
				const processedHostname = parts.join('.');
				url = url + processedHostname + suffix;
			} else {
				url = url + module + window.location.hostname + suffix;
			}
			console.log('baseUrl ===>', url);
			return url;
		},
		setMenuType(type: MenuType | string, params = {}) {
			this.menuChoice.type = type;
			const last = this.lastSelectedTabs.get(type);
			if (last) {
				this.menuChoice.tab = last;
			} else {
				const tabs = this.userTabs.get(type);
				if (tabs && tabs.length > 0) {
					this.menuChoice.tab = tabs[0];
				} else {
					this.menuChoice.tab = TabType.Empty;
				}
			}
			this.menuChoice.params = params;
			const readerStore = useReaderStore();
			readerStore.readingIndex = -1;
			console.log('======> menu', type);
			console.log('======> last tab', this.lastSelectedTabs);
			console.log('======> tab', this.menuChoice.tab);
		},

		getOnceMenuParams() {
			const data = this.menuChoice.params;
			this.menuChoice.params = '';
			return data;
		},

		setMenuTab(tab: TabType | string) {
			this.lastSelectedTabs.set(this.menuChoice.type, tab);
			this.menuChoice.tab = tab;
			const readerStore = useReaderStore();
			readerStore.readingIndex = -1;
			console.log('======> last tab', this.lastSelectedTabs);
			console.log('======> tab', this.menuChoice.tab);
		},
		setUserTabs(
			type: MenuType | string,
			tabs: TabType[] | string[] | number[]
		) {
			this.userTabs.set(type, tabs);
			const last = this.lastSelectedTabs.get(type);
			if (!last || tabs.findIndex((tab: any) => tab === last) === -1) {
				if (tabs.length > 0) {
					this.lastSelectedTabs.set(type, tabs[0]);
				} else {
					this.lastSelectedTabs.set(type, TabType.Empty);
				}
			}
		},
		nextTab() {
			const tab = this.lastSelectedTabs.get(this.menuChoice.type);
			const userTabs = this.userTabs.get(this.menuChoice.type);
			if (!tab || !userTabs || userTabs?.length === 0) {
				return;
			}
			const index = userTabs.findIndex((item) => item === tab);
			if (index + 1 < userTabs.length) {
				this.setMenuTab(userTabs[index + 1] as string);
			} else {
				this.setMenuTab(userTabs[0] as string);
			}
		},
		preTab() {
			const tab = this.lastSelectedTabs.get(this.menuChoice.type);
			const userTabs = this.userTabs.get(this.menuChoice.type);
			if (!tab || !userTabs || userTabs?.length === 0) {
				return;
			}

			const index = userTabs.findIndex((item) => item === tab);
			if (index - 1 > -1) {
				this.setMenuTab(userTabs[index - 1] as string);
			} else {
				this.setMenuTab(userTabs[userTabs.length - 1] as string);
			}
		},
		load() {
			if (localStorage.getItem(RIGHT_DRAWER_OPEN)) {
				this.rightDrawerOpen =
					localStorage.getItem(RIGHT_DRAWER_OPEN) === STORAGE_TRUE;
			}
			if (localStorage.getItem(ARTICLE_TOPIC_OPEN)) {
				this.articleTopicOpen =
					localStorage.getItem(ARTICLE_TOPIC_OPEN) === STORAGE_TRUE;
			}
			if (localStorage.getItem(ENTRY_REMOVE_WITH_FILE)) {
				this.entryRemoveWithFile =
					localStorage.getItem(ENTRY_REMOVE_WITH_FILE) === STORAGE_TRUE;
			}
			if (localStorage.getItem(TASK_REMOVE_WITH_FILE)) {
				this.taskRemoveWithFile =
					localStorage.getItem(TASK_REMOVE_WITH_FILE) === STORAGE_TRUE;
			}
			if (localStorage.getItem(FEED_REMOVE_WITH_FILE)) {
				this.feedRemoveWithFile =
					localStorage.getItem(FEED_REMOVE_WITH_FILE) === STORAGE_TRUE;
			}
			if (localStorage.getItem(TREND_REMOVE_NOT_NOTIFY)) {
				this.trendRemoveNotNotify =
					localStorage.getItem(TREND_REMOVE_NOT_NOTIFY) === STORAGE_FALSE;
			}
			if (localStorage.getItem(RECOMMENDATION_OPEN)) {
				this.recommendationOpen =
					localStorage.getItem(RECOMMENDATION_OPEN) === STORAGE_TRUE;
			}
			if (localStorage.getItem(UPLOAD_COOKIES_OPEN)) {
				this.uploadCookiesOpen =
					localStorage.getItem(UPLOAD_COOKIES_OPEN) === STORAGE_TRUE;
			}
			if (localStorage.getItem(UPLOAD_LINKS_OPEN)) {
				this.uploadLinksOpen =
					localStorage.getItem(UPLOAD_LINKS_OPEN) === STORAGE_TRUE;
			}
		},
		clear() {
			localStorage.removeItem(RIGHT_DRAWER_OPEN);
			localStorage.removeItem(ARTICLE_TOPIC_OPEN);
			localStorage.removeItem(ENTRY_REMOVE_WITH_FILE);
			localStorage.removeItem(TASK_REMOVE_WITH_FILE);
			localStorage.removeItem(FEED_REMOVE_WITH_FILE);
			localStorage.removeItem(TREND_REMOVE_NOT_NOTIFY);
			localStorage.removeItem(RECOMMENDATION_OPEN);
		},
		setThemeSetting(type: string, save = true) {
			if (!type) {
				return;
			}
			this.themeSetting = type;
			console.log(this.themeSetting);
			if (type === THEME_TYPE.AUTO) {
				Dark.set(Dark.mode);
			} else {
				Dark.set(type === THEME_TYPE.DARK);
			}
			if (save) {
				localStorage.setItem(THEME_SETTING, type);
			}
		},
		setRightDrawerOpen(open: boolean) {
			this.rightDrawerOpen = open;
			localStorage.setItem(
				RIGHT_DRAWER_OPEN,
				open ? STORAGE_TRUE : STORAGE_FALSE
			);
		},
		setArticleTopicOpen(open: boolean) {
			this.articleTopicOpen = open;
			localStorage.setItem(
				ARTICLE_TOPIC_OPEN,
				open ? STORAGE_TRUE : STORAGE_FALSE
			);
		},
		setEntryRemoveWithFile(open: boolean) {
			this.entryRemoveWithFile = open;
			localStorage.setItem(
				ENTRY_REMOVE_WITH_FILE,
				open ? STORAGE_TRUE : STORAGE_FALSE
			);
		},
		setTrendRemoveNotNotify(open: boolean) {
			this.trendRemoveNotNotify = open;
			localStorage.setItem(
				TREND_REMOVE_NOT_NOTIFY,
				open ? STORAGE_TRUE : STORAGE_FALSE
			);
		},
		setFeedRemoveWithFile(open: boolean) {
			this.feedRemoveWithFile = open;
			localStorage.setItem(
				FEED_REMOVE_WITH_FILE,
				open ? STORAGE_TRUE : STORAGE_FALSE
			);
		},
		setTaskRemoveWithFile(open: boolean) {
			this.taskRemoveWithFile = open;
			localStorage.setItem(
				TASK_REMOVE_WITH_FILE,
				open ? STORAGE_TRUE : STORAGE_FALSE
			);
		},
		setRecommendationOpen(open: boolean) {
			this.recommendationOpen = open;
			localStorage.setItem(
				RECOMMENDATION_OPEN,
				open ? STORAGE_TRUE : STORAGE_FALSE
			);
		},
		setUploadCookiesOpen(open: boolean) {
			this.uploadCookiesOpen = open;
			localStorage.setItem(
				UPLOAD_COOKIES_OPEN,
				open ? STORAGE_TRUE : STORAGE_FALSE
			);
		},
		setUploadLinksOpen(open: boolean) {
			this.uploadLinksOpen = open;
			localStorage.setItem(
				UPLOAD_LINKS_OPEN,
				open ? STORAGE_TRUE : STORAGE_FALSE
			);
		}
	}
});
