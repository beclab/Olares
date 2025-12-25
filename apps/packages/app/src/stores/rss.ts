import { defineStore } from 'pinia';
import {
	CreateNote,
	DefaultType,
	Entry,
	Feed,
	Label,
	Note,
	RecommendAlgorithm,
	RemoveData,
	RemoveType,
	SimpleEntry,
	SOURCE_TYPE
} from 'src/utils/rss-types';
import { MenuType } from 'src/utils/rss-menu';
import {
	createFeed,
	createLabel,
	createNote,
	deleteFeeds,
	deleteLabel,
	deleteNote,
	getRecommendEntryList,
	removeEntry,
	removeEntrySource,
	removeLabelOnEntry,
	removeLabelOnNote,
	saveEntry,
	setLabelOnEntry,
	setLabelOnNote,
	subscribeTrendFeed,
	syncEntries,
	syncFeeds,
	syncLabels,
	syncNotes,
	syncRemove,
	syncWaitFeeds,
	updateEntryReadLater,
	updateEntryUnread,
	updateLabel,
	updateNote
} from 'src/api/wise';
import { useConfigStore } from './rss-config';
import {
	binaryInsert,
	CompareFeed,
	CompareRecentlyEntry,
	extractHtml
} from 'src/utils/rss-utils';
import { busEmit } from '../utils/bus';
import {
	DATABASE_VERSION,
	SYNC_ENTRY_TIME,
	SYNC_FEED_TIME,
	SYNC_LABEL_TIME,
	SYNC_NOTE_TIME,
	SYNC_REMOVE_TIME,
	WISE_DATABASE
} from '../utils/localStorageConstant';
import {
	addOrUpdateEntries,
	clearEntries,
	getEntriesByFeedId,
	getEntryById,
	removeEntryById,
	removeEntryByUrl,
	updateEntryById
} from '../pages/Wise/database/tables/entry';
import {
	addOrUpdateFeeds,
	clearFeeds,
	getAllFeeds,
	getFeedById,
	getFeedByUrl,
	removeFeedById,
	updateFeedById
} from '../pages/Wise/database/tables/feed';
import {
	addOrUpdateNotes,
	clearNotes,
	getAllNotes,
	removeNoteById
} from '../pages/Wise/database/tables/note';
import {
	addOrUpdateLabels,
	clearLabels,
	getAllLabels,
	removeLabelById
} from '../pages/Wise/database/tables/label';
import { sendMessageToWorker } from '../pages/Wise/database/sqliteService';
import { useTransfer2Store } from './transfer2';
import { clearFilters } from '../pages/Wise/database/tables/view';
import { i18n } from 'src/boot/i18n';
import { useBlacklistStore } from 'src/stores/settings/blacklist';

export type DataState = {
	loading: boolean;
	support_algorithm: RecommendAlgorithm[];
	show_recommends: SimpleEntry[];
	recommends: Record<string, Entry>;
	recentlyList: Entry[];

	lastEntrySyncTime: number;
	lastFeedSyncTime: number;
	feeds: Feed[];
	wait_sync_feeds: string[];
	lastLabelSyncTime: number;
	labels: Label[];
	lastNoteSyncTime: number;
	notes: Note[];
	lastRemoveSyncTime: number;
};

export const useRssStore = defineStore('rss', {
	state: () => {
		return {
			loading: false,
			lastEntrySyncTime: 0,
			lastFeedSyncTime: 0,
			lastLabelSyncTime: 0,
			lastNoteSyncTime: 0,
			lastRemoveSyncTime: 0,
			recentlyList: [],
			wait_sync_feeds: [],
			show_recommends: [],
			recommends: {},
			support_algorithm: [
				{
					id: 'r4world',
					title: 'World'
				},
				{
					id: 'r4tech',
					title: 'Tech'
				},
				{
					id: 'r4sport',
					title: 'Sport'
				},
				{
					id: 'r4business',
					title: 'Business'
				},
				{
					id: 'r4entertainment',
					title: 'Entertainment'
				}
			],
			feeds: [],
			labels: [],
			notes: []
		} as unknown as DataState;
	},
	actions: {
		async load() {
			this.loading = true;
			console.log('load');
			await sendMessageToWorker('init');

			if (localStorage.getItem(SYNC_FEED_TIME)) {
				this.lastFeedSyncTime = Number(localStorage.getItem(SYNC_FEED_TIME));
			}
			if (localStorage.getItem(SYNC_ENTRY_TIME)) {
				this.lastEntrySyncTime = Number(localStorage.getItem(SYNC_ENTRY_TIME));
			}
			if (localStorage.getItem(SYNC_REMOVE_TIME)) {
				this.lastRemoveSyncTime = Number(
					localStorage.getItem(SYNC_REMOVE_TIME)
				);
			}
			if (localStorage.getItem(SYNC_LABEL_TIME)) {
				this.lastLabelSyncTime = Number(localStorage.getItem(SYNC_LABEL_TIME));
			}
			if (localStorage.getItem(SYNC_NOTE_TIME)) {
				this.lastNoteSyncTime = Number(localStorage.getItem(SYNC_NOTE_TIME));
			}

			console.log('load entry:', this.lastEntrySyncTime);
			console.log('load feed:', this.lastFeedSyncTime);
			console.log('load labels:', this.lastLabelSyncTime);
			console.log('load notes:', this.lastNoteSyncTime);
			console.log('load remove:', this.lastRemoveSyncTime);

			const feedList = await getAllFeeds();
			feedList.forEach((feed) => {
				this._addLocalFeed(feed);
			});

			const labelList = await getAllLabels();
			labelList.forEach((label) => {
				this._addLocalLabel(label);
			});

			const noteList = await getAllNotes();
			noteList.forEach((note) => {
				this._addLocalNote(note);
			});

			this.loading = false;
		},
		async clear() {
			await sendMessageToWorker('init');
			await clearEntries();
			await clearFeeds();
			await clearNotes();
			await clearLabels();
			await clearFilters();
			await sendMessageToWorker('close');
			const transferStore = useTransfer2Store();
			await transferStore.clearTransferData();
			localStorage.removeItem(SYNC_FEED_TIME);
			localStorage.removeItem(SYNC_ENTRY_TIME);
			localStorage.removeItem(SYNC_LABEL_TIME);
			localStorage.removeItem(SYNC_NOTE_TIME);
			localStorage.removeItem(SYNC_REMOVE_TIME);
			localStorage.setItem(WISE_DATABASE, DATABASE_VERSION);
			await sendMessageToWorker('init');
		},
		async sync() {
			await this.syncFeeds();
			await this.syncEntries();
			await this.syncLabels();
			await this.syncNotes();
			await this.syncRemoveData();
		},

		async syncEntries() {
			let data: Entry[] = [];
			do {
				data = await syncEntries(this.lastEntrySyncTime);
				await addOrUpdateEntries(
					data.map((item) => {
						return { ...item, summary: extractHtml(item) };
					})
				);

				for (const entry of data) {
					const time = new Date(entry.updatedAt).getTime();
					if (time > this.lastEntrySyncTime) {
						this.lastEntrySyncTime = time;
						localStorage.setItem(
							SYNC_ENTRY_TIME,
							this.lastEntrySyncTime.toString()
						);
					}
				}
			} while (data.length >= 100);
		},
		async syncFeeds() {
			let data: Feed[] = [];
			do {
				data = await syncFeeds(this.lastFeedSyncTime);
				await addOrUpdateFeeds(data);

				for (const feed of data) {
					const time = new Date(feed.updated_at).getTime();
					if (time > this.lastFeedSyncTime) {
						this.lastFeedSyncTime = time;
						localStorage.setItem(
							SYNC_FEED_TIME,
							this.lastFeedSyncTime.toString()
						);
					}

					this._addLocalFeed(feed);
				}
			} while (data.length >= 100);
		},
		async syncLabels() {
			let data: Label[] = [];
			do {
				data = await syncLabels(this.lastLabelSyncTime);
				await addOrUpdateLabels(data);

				for (const label of data) {
					const time = new Date(label.updated_at).getTime();
					if (time > this.lastLabelSyncTime) {
						this.lastLabelSyncTime = time;
						localStorage.setItem(
							SYNC_LABEL_TIME,
							this.lastLabelSyncTime.toString()
						);
					}

					this._addLocalLabel(label);
				}
			} while (data.length >= 100);
		},
		async syncNotes() {
			let data: Note[] = [];
			do {
				data = await syncNotes(this.lastNoteSyncTime);
				await addOrUpdateNotes(data);

				for (const note of data) {
					const time = new Date(note.updated_at).getTime();
					if (time > this.lastNoteSyncTime) {
						this.lastNoteSyncTime = time;
						localStorage.setItem(
							SYNC_NOTE_TIME,
							this.lastNoteSyncTime.toString()
						);
					}

					this._addLocalNote(note);
				}
			} while (data.length >= 100);
		},
		async syncRemoveData() {
			let data: RemoveData[] = [];
			do {
				data = await syncRemove(this.lastRemoveSyncTime);

				for (const remove of data) {
					const time = new Date(remove.created_at).getTime();
					if (time > this.lastRemoveSyncTime) {
						this.lastRemoveSyncTime = time;
						localStorage.setItem(
							SYNC_REMOVE_TIME,
							this.lastRemoveSyncTime.toString()
						);
					}

					switch (remove.remove_type) {
						case RemoveType.Entry:
							removeEntryById(remove.remove_id);
							break;
						case RemoveType.Feed:
							this.feeds = this.feeds.filter(
								(feed) => feed.id !== String(remove.id)
							);
							this.wait_sync_feeds = this.wait_sync_feeds.filter(
								(feed) => feed !== String(remove.id)
							);
							removeFeedById(remove.remove_id);
							break;
						case RemoveType.Label:
							this.labels = this.labels.filter(
								(label) => label.id !== String(remove.id)
							);
							removeLabelById(remove.remove_id);
							break;
						case RemoveType.Note:
							this.notes = this.notes.filter(
								(note) => note.id !== String(remove.id)
							);
							removeNoteById(remove.remove_id);
							break;
					}
				}
			} while (data.length >= 100);
		},
		async syncWaitFeeds() {
			if (this.wait_sync_feeds.length === 0) {
				return;
			}

			console.log('sync_wait_feeds start:', this.wait_sync_feeds.length);

			try {
				const feeds = this.wait_sync_feeds.slice(0, 100);
				const data: Feed[] = await syncWaitFeeds(feeds);
				console.log(data);

				this.wait_sync_feeds = this.wait_sync_feeds.filter((id) => {
					return !data.some((feed) => String(feed.id) === String(id));
				});

				await addOrUpdateFeeds(data);
				data.map(async (feed) => {
					this._addLocalFeed(feed);
				});

				if (data.length > 0) {
					busEmit('feedUpdate');
				}

				console.log('sync_wait_feeds end:', this.wait_sync_feeds.length);
			} catch (err) {
				console.error('sync_wait_feeds error:', err);
			}
		},

		async removeSourceEntries(
			source: string,
			entries: SimpleEntry[],
			removeFile: boolean
		): Promise<boolean> {
			try {
				const urls = entries
					.map((item) => (item.url ? item.url : ''))
					.filter((item) => !!item);
				await removeEntrySource(source, urls, removeFile);
				for (let i = 0; i < entries.length; i++) {
					this._removeLocalEntrySource(source, entries[i]);
				}
				return true;
			} catch (e) {
				console.log(e);
				return false;
			}
		},

		async _removeLocalEntrySource(source: string, entry: SimpleEntry) {
			const localEntry = await getEntryById(entry.id);
			if (!localEntry) {
				return;
			}
			const index2 = localEntry.sources.findIndex((item) => item === source);
			if (index2 >= 0) {
				localEntry.sources.splice(index2, 1);
			}
			if (localEntry.sources.length == 0) {
				// console.log('local entry remove', localEntry);
				await removeEntryById(entry.id);
			} else {
				console.log('local entry update', localEntry);
				// this.entryStore?.setItem<Entry>(entry.id, localEntry);
			}
		},

		async removeEntry(url: string, removeFile: boolean): Promise<boolean> {
			try {
				await removeEntry([url], removeFile);
				await removeEntryByUrl(url);
				return true;
			} catch (e) {
				console.log(e);
				return false;
			}
		},

		getLocalRecommend(id: string): Entry | null {
			return this.recommends[id];
		},

		async getLocalEntryByMenu(id: string): Promise<Entry | null | undefined> {
			let entry: Entry | null | undefined = null;
			const configStore = useConfigStore();
			if (configStore.menuChoice.type === MenuType.Trend) {
				entry = this.getLocalRecommend(id);
			} else if (configStore.menuChoice.type === MenuType.History) {
				entry = this.getRecentlyEntry(id);
			} else {
				entry = await getEntryById(id);
			}
			return entry;
		},

		async setRecommendUnread(id: string, unread: boolean) {
			const recommend = this.recommends[id];
			if (recommend) {
				recommend.unread = unread;
			}
			this.show_recommends.forEach((item) => {
				if (item.id === id) {
					item.unread = unread;
				}
			});
		},

		async markEntryUnread(entry_ids: string[], status: boolean) {
			try {
				await updateEntryUnread({
					entry_ids,
					status
				});
				await this.syncEntries();
			} catch (e) {
				console.log(e);
			}
		},

		async markEntryReadLater(entry_ids: string[], status: boolean) {
			await updateEntryReadLater({
				entry_ids,
				status
			});
			await this.sync();
		},

		// async getSupportAlgorithm(): Promise<void> {
		// 	try {
		// 		const data = await getRecommendAlgorithmList();
		// 		console.log('get_support_algorithm');
		// 		console.log(data);
		//
		// 		this.support_algorithm = data;
		// 		if (this.support_algorithm.length > 0) {
		// 			const configStore = useConfigStore();
		// 			configStore.setUserTabs(
		// 				MenuType.Trend,
		// 				this.support_algorithm.map((item) => item.id)
		// 			);
		// 		}
		// 	} catch (error: any) {
		// 		console.log(error);
		// 	}
		// },

		addRecentlyEntry(entry: Entry) {
			const index = this.recentlyList.findIndex((l) => l.id == entry.id);
			if (index >= 0) {
				this.recentlyList.splice(index, 1);
				this.recentlyList.splice(0, 0, entry);
			} else {
				binaryInsert<Entry>(this.recentlyList, entry, CompareRecentlyEntry);
			}
		},

		getRecentlyEntry(id: string) {
			return this.recentlyList.find((l) => l.id == id);
		},

		async getRecommendList(
			source: string,
			pageSize = DefaultType.Limit
		): Promise<{ data: Entry[]; message: string }> {
			try {
				const data: Entry[] = await getRecommendEntryList(source, pageSize);
				console.log('get_recommendList');
				console.log(data);
				this.setLocalRecommendEntry(source, data);

				return { data, message: '' };
			} catch (error: any) {
				console.log(error);
				return { data: [], message: error.message };
			}
		},

		setLocalRecommendEntry(source: string, data: Entry[]) {
			for (const entry of data) {
				//this.recommends.push(item);
				const le: SimpleEntry = {
					id: entry.id,
					source,
					createdAt: entry.createdAt,
					updatedAt: entry.updatedAt,
					title: entry.title || '',
					url: entry.url,
					readlater: entry.readlater,
					local_file_path: entry.local_file_path || '',
					unread: entry.unread,
					published_at: entry.published_at,
					author: entry.author || '',
					image_url: entry.image_url || '',
					feed_id: entry.feed_id || '',
					full_content: entry.full_content || '',
					crawler: entry.crawler,
					status: entry.status,
					file_type: entry.file_type || ''
				};
				const index = this.show_recommends.indexOf(le);
				if (index >= 0) {
					this.show_recommends.splice(index, 1, le);
				} else {
					this.show_recommends.push(le);
				}
				this.recommends[entry.id] = entry;
			}
		},

		async getLocalFeed(id: string): Promise<Feed | null> {
			const feed = await getFeedById(id);
			if (feed) {
				return feed;
			}
			const feedId = this.wait_sync_feeds.find((feed_id) => feed_id == id);
			if (!feedId) {
				console.log('push feedID', id);
				this.wait_sync_feeds.push(id);
			}
			return null;
		},

		async addFeed(feed_url: string, auto_download = true) {
			const feed: Feed = await this.addFeedOnly(feed_url, auto_download);
			await updateFeedById(feed.id, feed);
			await this.syncFeeds();
		},

		async addFeedOnly(feed_url: string, auto_download = false) {
			try {
				const feed: Feed = await createFeed(feed_url, auto_download);
				return Promise.resolve(feed);
			} catch (error) {
				return Promise.reject(error);
			}
		},

		async addEntryOnlyRequest(entry_url: string): Promise<Entry[]> {
			const entries: Entry[] = await saveEntry([
				{
					url: entry_url,
					source: SOURCE_TYPE.LIBRARY
				}
			]);
			return entries;
		},

		async addEntry(entry_url: string, onlyRequest = false) {
			const entries: Entry[] = await this.addEntryOnlyRequest(entry_url);

			if (onlyRequest) {
				return entries;
			}
			for (let i = 0; i < entries.length; i++) {
				await updateEntryById(entries[i].id, {
					...entries[i],
					summary: extractHtml(entries[i])
				});
			}
			await this.syncEntries();
			return entries;
		},

		async removeFeed(feed_url: string, removeFile: boolean) {
			try {
				await deleteFeeds([feed_url], removeFile);
				const feed = await getFeedByUrl(feed_url);

				if (feed) {
					console.log('==> feed', feed);
					const feedId = feed.id;
					if (feed.sources?.includes(SOURCE_TYPE.WISE)) {
						console.log('==>', feed.title);
						console.log('==>', feed.sources);
						feed.sources = feed.sources.filter((item) => {
							return item !== SOURCE_TYPE.WISE;
						});
						this._addLocalFeed(feed);
						if (feed.sources.length === 0) {
							console.log('==> feed remove');
							await removeFeedById(feedId);
							delete this.feeds[feedId];
						} else {
							console.log('==> feed update', feed);
							await updateFeedById(feedId, feed);
						}
					}

					const entries: Entry[] = await getEntriesByFeedId(feedId);
					console.log('==>', entries);

					for (const entry of entries) {
						if (entry.sources?.includes(SOURCE_TYPE.WISE)) {
							console.log('==>', entry.title);
							console.log('==>', entry.sources);
							entry.sources = entry.sources.filter(
								(source) => source !== SOURCE_TYPE.WISE
							);
							const entryId = entry.id;
							if (entry.sources.length == 0) {
								console.log('==> entry remove');
								await removeEntryById(entryId);
							} else {
								console.log('==> entry update', entry);
								await updateEntryById(entryId, entry);
							}
						}
					}
				}

				await this.sync();
			} catch (e) {
				console.log(e);
			}
		},

		async saveTrendFeedInLocal(feed_id: string) {
			await subscribeTrendFeed(feed_id);
			await this.syncFeeds();
		},

		getEntryLabels(entry: Entry | undefined | SimpleEntry): Label[] {
			if (entry && entry.id) {
				return this.labels.filter((item) => {
					if (item.entries) {
						return item.entries.includes(entry.id);
					}
				});
			} else {
				return [];
			}
		},

		getEntryNote(
			entry: Entry | undefined | SimpleEntry
		): Note | null | undefined {
			if (entry && entry.id) {
				return this.notes.find((item) => {
					return item.entry_id === entry.id && !item.deleted;
				});
			} else {
				return null;
			}
		},

		_addLocalFeed(feed: Feed) {
			const index = this.feeds.findIndex((item) => item.id == feed.id);
			if (index >= 0) {
				this.feeds[index] = feed;
			} else {
				binaryInsert(this.feeds, feed, CompareFeed);
			}
		},

		_addLocalLabel(label: Label) {
			const index = this.labels.findIndex((item) => item.id == label.id);
			if (index >= 0) {
				if (label.deleted) {
					this.labels.splice(index, 1);
				} else {
					this.labels[index] = label;
				}
			} else {
				if (!label.deleted) {
					this.labels.push(label);
				}
			}
		},

		_addLocalNote(note: Note) {
			const index = this.notes.findIndex((item) => item.id == note.id);
			if (index >= 0) {
				if (note.deleted) {
					this.notes.splice(index, 1);
				} else {
					this.notes[index] = note;
				}
			} else {
				if (!note.deleted) {
					this.notes.push(note);
				}
			}
		},

		async addNote(note: CreateNote) {
			const data: Note = await createNote(note);
			console.log('add_note');
			console.log(data);
			await this.syncNotes();
			return data;
		},

		async updateNote(note: Note) {
			const data: Note = await updateNote(note);
			console.log('update_note');
			console.log(data);
			await this.syncNotes();
			return data;
		},

		async removeNote(id: string) {
			const data: Note = await deleteNote(id);
			console.log('remove_note');
			console.log(data);
			await removeNoteById(id);
			await this.syncNotes();
			return data;
		},

		async addLabel(name: string) {
			const data: Label = await createLabel(name);
			console.log('add_label');
			console.log(data);
			await this.syncLabels();
			return data;
		},

		async updateLabel(label: Label) {
			const data: Label = await updateLabel(label);
			console.log('update_label');
			console.log(data);
			await this.syncLabels();
			return data;
		},

		async removeLabel(id: string) {
			const data: Label = await deleteLabel(id);
			console.log('remove_label');
			console.log(data);
			await removeLabelById(id);
			await this.syncLabels();
			return data;
		},

		async setLabelOnEntry(label_id: string, entry_id: string) {
			const data: Label = await setLabelOnEntry(label_id, entry_id);
			console.log('setLabelOnEntry');
			console.log(data);
			await this.syncLabels();
			return data;
		},

		async removeLabelOnEntry(label_id: string, entry_id: string) {
			const data: Label = await removeLabelOnEntry(label_id, entry_id);
			console.log('removeLabelOnEntry');
			console.log(data);
			await this.syncLabels();
			return data;
		},

		async setLabelOnNote(label_id: string, note_id: string) {
			const data: Label = await setLabelOnNote(label_id, note_id);
			console.log('setLabelOnNote');
			console.log(data);
			await this.syncLabels();
			return data;
		},

		async removeLabelOnNote(label_id: string, note_id: string) {
			const data: Label = await removeLabelOnNote(label_id, note_id);
			console.log('removeLabelOnNote');
			console.log(data);
			await this.syncLabels();
			return data;
		}
	}
});
