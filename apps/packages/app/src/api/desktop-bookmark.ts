import axios from 'axios';
import { useTokenStore } from 'src/stores/desktop/token';
import {
	Bookmark,
	BookmarkOpenMode,
	BookmarksDoc
} from 'src/types/desktop-layout';

function baseUrl(): string {
	const tokenStore = useTokenStore();
	return tokenStore.url || '';
}

export interface CreateBookmarkPayload {
	name: string;
	url: string;
	openMode: BookmarkOpenMode;
	icon?: string;
}

export interface UpdateBookmarkPayload {
	name?: string;
	url?: string;
	openMode?: BookmarkOpenMode;
	icon?: string | null;
}

export interface FaviconResult {
	dataUri: string;
	contentType: string;
}

export async function listBookmarks(): Promise<BookmarksDoc> {
	return await axios.get(`${baseUrl()}/api/desktop/bookmarks`);
}

export async function createBookmark(
	body: CreateBookmarkPayload
): Promise<Bookmark> {
	return await axios.post(`${baseUrl()}/api/desktop/bookmarks`, body);
}

export async function updateBookmark(
	id: string,
	body: UpdateBookmarkPayload
): Promise<Bookmark> {
	return await axios.put(
		`${baseUrl()}/api/desktop/bookmarks/${encodeURIComponent(id)}`,
		body
	);
}

export async function deleteBookmark(
	id: string
): Promise<{ removed: boolean }> {
	return await axios.delete(
		`${baseUrl()}/api/desktop/bookmarks/${encodeURIComponent(id)}`
	);
}

export async function resolveFavicon(url: string): Promise<FaviconResult> {
	return await axios.get(`${baseUrl()}/api/desktop/bookmarks/favicon`, {
		params: { url }
	});
}
