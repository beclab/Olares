import { defineStore } from 'pinia';
import { useRssStore } from './rss';
import cloneDeep from 'lodash/cloneDeep';
import { MenuType } from '../utils/rss-menu';
import {
	ArticleTopic,
	Entry,
	Feed,
	ImpressionAction,
	SDKQueryRequest,
	SimpleEntry,
	SOURCE_TYPE
} from '../utils/rss-types';
import {
	getRemoteEntryById,
	rssContentQuery,
	sdkSearchFeedsByPath,
	updateEntryReadLater,
	updateImpression
} from '../api/wise';
import { getEntryById } from '../pages/Wise/database/tables/entry';
import { useConfigStore } from './rss-config';

export type ReaderState = {
	readingEntry: Entry | null | undefined;
	readingFeed: Feed | null | undefined;
	hoverEntry: Entry | null | undefined;
	unread: boolean;
	readLater: boolean;
	readLaterLoading: boolean;
	inbox: boolean;
	inboxLoading: boolean;
	subscribed: boolean;
	navigationList: any[];
	subscribedLoadingSet: Set<string>;
	readingIndex: number;
	topicArray: ArticleTopic[];
	readingTopic: ArticleTopic | null;
	realContent: string;
};

export const useReaderStore = defineStore('reader', {
	state: () => {
		return {
			readingEntry: null,
			readingFeed: null,
			hoverEntry: null,
			unread: false,
			readLater: false,
			readLaterLoading: false,
			inbox: false,
			inboxLoading: false,
			drawerFeed: null,
			subscribed: false,
			navigationList: [],
			subscribedLoadingSet: new Set(),
			readingIndex: -1,
			topicArray: [],
			readingTopic: null,
			realContent: ''
		} as ReaderState;
	},

	actions: {
		canPreEntryRoute(entry: Entry): boolean {
			if (!entry) {
				return false;
			}
			const index = this.navigationList
				.filter((e) => !!e.id)
				.findIndex((e) => e.id == entry.id);
			return index > 0;
		},

		getPreEntryId(entry: Entry): string {
			const index = this.navigationList
				.filter((e) => !!e.id)
				.findIndex((e) => e.id == entry.id);
			return this.navigationList[index - 1].id;
		},

		setNavigationList(list: SimpleEntry[] | Entry[]) {
			if (!list) {
				this.navigationList = [];
				return;
			}
			this.navigationList = cloneDeep(list);
		},

		clearNavigationList() {
			this.navigationList = [];
		},

		async setCurrentReadChange(unread: boolean) {
			if (this.readingEntry) {
				const rssStore = useRssStore();
				const entry = this.readingEntry;
				try {
					await rssStore.markEntryUnread([entry.id], unread);
					//update history data
					const configStore = useConfigStore();
					if (configStore.menuChoice.type === MenuType.History) {
						this.readingEntry.unread = unread;
					}
					this.unread = unread;
				} catch (e) {
					console.log(e);
				}
				await this.entryUpdate(entry.id);
			}
		},

		async setCurrentReadLater(readLater: boolean) {
			if (this.readingEntry) {
				const entry = this.readingEntry;
				try {
					if (readLater) {
						this.readLaterLoading = true;
					} else {
						this.inboxLoading = true;
					}
					await updateEntryReadLater({
						entry_ids: [entry.id],
						status: readLater
					});
					this.readLater = readLater;
					if (readLater) {
						this.readLaterLoading = false;
					} else {
						this.inboxLoading = false;
					}
				} catch (e) {
					console.log(e);
					if (readLater) {
						this.readLaterLoading = false;
					} else {
						this.inboxLoading = false;
					}
				}
				const rssStore = useRssStore();
				await rssStore.syncEntries();
				await this.entryUpdate(entry.id);
			}
		},

		async setCurrentSubscribe(removeFile?: boolean) {
			if (this.readingFeed) {
				const feed = this.readingFeed;
				this.subscribedLoadingSet.add(feed.id);
				const rssStore = useRssStore();
				if (this.subscribed) {
					try {
						const configStore = useConfigStore();
						await rssStore.removeFeed(
							feed.feed_url,
							removeFile || configStore.feedRemoveWithFile
						);
						this.subscribed = false;
					} catch (e) {
						console.log(e);
					}
				} else {
					try {
						await rssStore.saveTrendFeedInLocal(feed.id);
						this.subscribed = true;
					} catch (e) {
						console.log(e);
					}
				}
				await this.feedUpdate();
				this.subscribedLoadingSet.delete(feed.id);
			}
		},

		async entryUpdate(entryId: string) {
			const rssStore = useRssStore();
			const localEntry = await getEntryById(entryId);
			this.updateReadingIndex();
			this.readingEntry = await rssStore.getLocalEntryByMenu(entryId);
			if (!this.readingEntry) {
				this.readingEntry = await getRemoteEntryById(entryId);
			}
			this.extractTitles();
			this.hoverEntry = cloneDeep(this.readingEntry);

			if (localEntry) {
				this.readLater =
					localEntry.readlater &&
					localEntry.sources.includes(SOURCE_TYPE.LIBRARY);
				this.inbox =
					!localEntry.readlater &&
					localEntry.sources.includes(SOURCE_TYPE.LIBRARY);
				this.unread = localEntry.unread;
			} else {
				this.readLater = false;
				this.inbox = false;
			}
			this.feedUpdate();
			return this.readingEntry;
		},
		async extractTitles() {
			if (this.readingEntry) {
				const content = this.readingEntry.full_content;
				if (content) {
					const parser = new DOMParser();
					const doc = parser.parseFromString(content, 'text/html');

					const titleElements = doc.querySelectorAll('h1, h2, h3, h4, h5, h6');
					this.topicArray = [];
					this.topicArray.push({
						text: this.readingEntry.title ? this.readingEntry.title : '',
						id: 'wise-article-header-title',
						level: 0,
						jump: false
					});
					Array.from(titleElements).map((element, index) => {
						if (!element.id) {
							element.id = `title-${index}`;
						}
						this.topicArray.push({
							text: element.textContent ? element.textContent.trim() : '',
							id: element.id,
							level: parseInt(element.tagName.charAt(1)),
							jump: false
						});
					});

					this.realContent = doc.body.innerHTML;
					return;
				}
			}

			this.topicArray = [];
			this.realContent = '';
		},

		updateReadingTopic(id: string) {
			if (id && this.topicArray.length > 0) {
				const topic = this.topicArray.find(
					(element: ArticleTopic) => element.id === id
				);
				if (topic) {
					this.readingTopic = topic;
				}
			}
		},

		updateReadingIndex() {
			if (this.readingEntry && this.navigationList.length > 0) {
				this.readingIndex = this.navigationList.findIndex(
					(e) => e.id == this.readingEntry!.id
				);
			}
		},

		async feedUpdate() {
			this.readingFeed = null;
			if (this.readingEntry && this.readingEntry.feed_id) {
				const rssStore = useRssStore();
				this.readingFeed = await rssStore.getLocalFeed(
					this.readingEntry.feed_id
				);
			} else {
				this.readingFeed = null;
			}
			// console.log(this.readingFeed);
			if (this.readingFeed) {
				this.subscribed = this.readingFeed.sources.includes(SOURCE_TYPE.WISE);
			} else {
				this.subscribed = false;
			}
			return this.readingFeed;
		},

		canNextEntryRoute(
			entry: Entry,
			condition: (entry: SimpleEntry) => boolean = () => true
		): boolean {
			if (!entry) {
				return false;
			}
			const index = this.navigationList
				.filter((e) => !!e.id)
				.findIndex((e) => e.id == entry.id);
			if (index < 0 || index > this.navigationList.length - 1) {
				return false;
			}
			if (condition) {
				const index2 = this.navigationList.findIndex((value, i) => {
					return i > index && condition(value);
				});
				if (index2 < 0 || index2 > this.navigationList.length - 1) {
					return false;
				}
			}
			return true;
		},

		getNextEntryId(
			entry: Entry,
			condition: (entry: SimpleEntry) => boolean = () => true
		): string {
			const index = this.navigationList
				.filter((e) => !!e.id)
				.findIndex((e) => e.id == entry.id);
			if (condition) {
				const index2 = this.navigationList.findIndex((value, i) => {
					return i > index && condition(value);
				});
				return this.navigationList[index2].id;
			}
			return this.navigationList[index + 1].id;
		},

		async updateImpression(entry: Entry, action: ImpressionAction) {
			if (entry.impression) {
				updateImpression({ id: entry.impression_id, ...action });
			} else {
				const sources = entry.sources.filter(
					(item) => item !== SOURCE_TYPE.LIBRARY && item !== SOURCE_TYPE.WISE
				);
				updateImpression({
					entry_id: entry.id,
					source: sources.length > 0 ? sources[0] : entry.sources[0],
					...action
				});
			}
		},

		async sdkSearchFeed(path: string) {
			try {
				const request = new SDKQueryRequest({ path });
				const result = await sdkSearchFeedsByPath(request);
				console.log(result);

				return result;
			} catch (error) {
				console.log(error);
			}
		},

		async rssContentQuery(query: string) {
			try {
				return await rssContentQuery(query);
			} catch (error) {
				console.log(error);
			}
		}
	}
});
