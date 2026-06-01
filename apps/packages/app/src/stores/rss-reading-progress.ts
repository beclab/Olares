import { defineStore } from 'pinia';
import { MenuType } from '../utils/rss-menu';
import { useConfigStore } from './rss-config';
import { useRssStore } from './rss';
import { FILE_TYPE } from '../utils/rss-types';
import { updateReadProgress } from '../api/wise';
import {
	getEntryById,
	updateEntryById
} from '../pages/Wise/database/tables/entry';
import { useReaderStore } from './rss-reader';

export const useReadingProgressStore = defineStore('readingProgress', {
	state: () => ({
		entryId: '',
		total: 0,
		totalWords: 0,
		current: 0,
		isLoaded: false,
		beginReadTime: 0,
		finished: false,
		intervalId: null as any
	}),

	getters: {
		progressPercentage: (state) => {
			if (!state.isLoaded) {
				return 0;
			}
			if (state.total === 0) {
				return 1;
			}
			return (state.current / state.total) * 100;
		}
	},

	actions: {
		checkReadingProgress() {
			if (!this.isLoaded) {
				return;
			}
			const readerStore = useReaderStore();
			if (!readerStore.readingEntry) {
				return;
			}

			let readPercentage: number;
			switch ((readerStore.readingEntry as any)?.file_type) {
				case FILE_TYPE.ARTICLE:
				case FILE_TYPE.EBOOK:
					readPercentage = 0.98;
					break;
				case FILE_TYPE.VIDEO:
				case FILE_TYPE.AUDIO:
					readPercentage = 0.95;
					break;
				default:
					readPercentage = 1;
					break;
			}

			if (this.current >= this.total * readPercentage) {
				if (this.finished) {
					return;
				}

				this.finished = true;
				this.markAsRead();

				if (readerStore.readingEntry) {
					readerStore.updateImpression(readerStore.readingEntry, {
						read_finish: true,
						read_time: Date.now() - this.beginReadTime
					});
				}
			}
		},

		markAsRead() {
			const readerStore = useReaderStore();
			const configStore = useConfigStore();
			const rssStore = useRssStore();
			if (readerStore.readingEntry && readerStore.readingEntry.unread) {
				readerStore.readingEntry.unread = false;
				if (configStore.menuChoice.type === MenuType.Trend) {
					rssStore.setRecommendUnread(readerStore.readingEntry.id, false);
				} else {
					rssStore.markEntryUnread([readerStore.readingEntry.id], false);
				}
			}
		},

		onreset(entryId: string) {
			this.entryId = entryId;
			this.total = 0;
			this.totalWords = 0;
			this.current = 0;
			this.isLoaded = false;
			this.beginReadTime = 0;
			this.finished = false;
		},

		setTotalProgress(total: number, totalWords = 0) {
			this.total = total;
			this.totalWords = totalWords;
			this.isLoaded = true;

			this.checkReadingProgress();
		},

		updateProgress(current: number) {
			this.current = current;

			this.checkReadingProgress();
		},
		startReading() {
			this.beginReadTime = Date.now();
			this._startAutoSave();
		},
		stopReading() {
			console.log('readingProgress: stopReading...');
			this._stopAutoSave();
			this._requestReadingProgress();
		},
		_startAutoSave() {
			if (this.intervalId) return;

			this.intervalId = setInterval(() => {
				this._requestReadingProgress();
			}, 10000);
		},
		_stopAutoSave() {
			if (this.intervalId) {
				clearInterval(this.intervalId);
				this.intervalId = null;
			}
		},
		async _requestReadingProgress() {
			if (!this.isLoaded) {
				return;
			}

			// Capture entryId before any async to avoid saving for wrong entry when switching
			const capturedEntryId = this.entryId;

			const entry = await getEntryById(capturedEntryId);

			if (!entry) {
				console.log('_requestReadingProgress entry null return');
				return;
			}

			// Re-read progress AFTER await: PDF child may call updateProgress() in its
			// onBeforeUnmount after our stopReading(), so store can have newer value here
			if (this.entryId !== capturedEntryId) {
				return;
			}
			const progress = this.progressPercentage;
			const current = this.current;
			const total = this.total;
			const totalWords = this.totalWords;

			if (entry.progress == progress) {
				console.log('_requestReadingProgress not need update');
				return;
			}

			entry.progress = progress;
			entry.played_time = current;

			switch (entry.file_type) {
				case FILE_TYPE.ARTICLE:
					entry.remaining_time = (totalWords * (1 - current)) / 200;
					break;
				case FILE_TYPE.EBOOK:
				case FILE_TYPE.PDF:
					// PDF/EBOOK: remaining pages (assuming 1 min per page)
					entry.remaining_time = total - current;
					break;
				case FILE_TYPE.VIDEO:
				case FILE_TYPE.AUDIO:
					entry.remaining_time = (total - current) / 60;
					break;
			}

			updateEntryById(entry.id, entry);

			updateReadProgress(
				capturedEntryId,
				progress.toString(),
				current.toString(),
				entry.remaining_time.toFixed(0).toString()
			)
				.then((data) => {
					console.log('need update == !!!', data);
				})
				.catch((error) => {
					console.error('Failed to update reading progress:', error);
				});
		}
	}
});
