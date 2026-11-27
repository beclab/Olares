import { defineStore } from 'pinia';
import {
	Bookmark,
	BookmarksDoc,
	BookmarkOpenMode
} from 'src/types/desktop-layout';
import {
	createBookmark as apiCreate,
	deleteBookmark as apiDelete,
	listBookmarks as apiList,
	updateBookmark as apiUpdate,
	CreateBookmarkPayload,
	UpdateBookmarkPayload
} from 'src/api/desktop-bookmark';

const CACHE_KEY = 'bookmarks_cache_v1';

interface State {
	bookmarks: Bookmark[];
	byId: Record<string, Bookmark>;
	version: number;
	updatedAt: number;
	loading: boolean;
	loadedAt: number;
	lastError: string | null;
}

function loadCache(): State | null {
	try {
		const raw = localStorage.getItem(CACHE_KEY);
		return raw ? (JSON.parse(raw) as State) : null;
	} catch (_e) {
		return null;
	}
}

function saveCache(s: State) {
	try {
		localStorage.setItem(
			CACHE_KEY,
			JSON.stringify({
				bookmarks: s.bookmarks,
				byId: s.byId,
				version: s.version,
				updatedAt: s.updatedAt
			})
		);
	} catch (_e) {
		/* ignore */
	}
}

function indexById(list: Bookmark[]): Record<string, Bookmark> {
	const out: Record<string, Bookmark> = {};
	for (const b of list) out[b.id] = b;
	return out;
}

export const useBookmarkStore = defineStore('desktopBookmark', {
	state: (): State => {
		const cached = loadCache();
		if (cached) {
			return {
				bookmarks: cached.bookmarks || [],
				byId: cached.byId || indexById(cached.bookmarks || []),
				version: cached.version || 0,
				updatedAt: cached.updatedAt || 0,
				loading: false,
				loadedAt: 0,
				lastError: null
			};
		}
		return {
			bookmarks: [],
			byId: {},
			version: 0,
			updatedAt: 0,
			loading: false,
			loadedAt: 0,
			lastError: null
		};
	},
	actions: {
		async fetchAll(): Promise<BookmarksDoc | null> {
			this.loading = true;
			try {
				const doc = await apiList();
				this._applyDoc(doc);
				this.loadedAt = Date.now();
				this.lastError = null;
				return doc;
			} catch (e: any) {
				this.lastError = e?.message || 'failed';
				console.error('[bookmark] fetchAll failed', e);
				return null;
			} finally {
				this.loading = false;
			}
		},

		async create(payload: CreateBookmarkPayload): Promise<Bookmark | null> {
			try {
				const bm = await apiCreate(payload);
				const next = [...this.bookmarks, bm];
				this._applyList(next, this.version + 1);
				return bm;
			} catch (e: any) {
				this.lastError = e?.response?.data?.code || e?.message || 'failed';
				console.error('[bookmark] create failed', e);
				return null;
			}
		},

		async update(
			id: string,
			patch: UpdateBookmarkPayload
		): Promise<Bookmark | null> {
			try {
				const bm = await apiUpdate(id, patch);
				const next = this.bookmarks.map((b) => (b.id === id ? bm : b));
				this._applyList(next, this.version + 1);
				return bm;
			} catch (e: any) {
				this.lastError = e?.response?.data?.code || e?.message || 'failed';
				console.error('[bookmark] update failed', e);
				return null;
			}
		},

		async delete(id: string): Promise<boolean> {
			try {
				const r = await apiDelete(id);
				if (r.removed) {
					const next = this.bookmarks.filter((b) => b.id !== id);
					this._applyList(next, this.version + 1);
				}
				return r.removed;
			} catch (e: any) {
				this.lastError = e?.response?.data?.code || e?.message || 'failed';
				console.error('[bookmark] delete failed', e);
				return false;
			}
		},

		// WS push handler. Overwrites local state with the server snapshot, but
		// only when it is newer. Bookmark versions are server-authoritative and
		// monotonic, so a payload whose version is <= ours is a stale/duplicate
		// out-of-order delivery (or an echo of our own optimistic write) and
		// must be ignored — otherwise it could clobber a newer local state.
		applyRemoteBookmarks(payload: { bookmarks: Bookmark[]; version: number }) {
			if (!payload || !Array.isArray(payload.bookmarks)) return;
			const incoming =
				typeof payload.version === 'number' ? payload.version : 0;
			if (incoming <= this.version) return;
			this._applyList(payload.bookmarks, incoming);
		},

		_applyDoc(doc: BookmarksDoc) {
			this.bookmarks = doc.bookmarks || [];
			this.byId = indexById(this.bookmarks);
			this.version = doc.version || 0;
			this.updatedAt = doc.updatedAt || Date.now();
			saveCache(this.$state);
		},

		_applyList(list: Bookmark[], version: number) {
			this.bookmarks = list;
			this.byId = indexById(list);
			this.version = version;
			this.updatedAt = Date.now();
			saveCache(this.$state);
		}
	}
});

export type { BookmarkOpenMode };
