// Desktop layout types exchanged with the Settings backend. Keep in sync with
// the server contract by hand.

export type Surface = 'launchpad' | 'dock';

// schemaVersion 2: form-factor dimension removed; the launchpad is a single
// linear order ({ key, order }) shared across PC and mobile and reflowed onto
// each device's grid at render time.
export const LAYOUT_SCHEMA_VERSION = 2;

export interface LaunchpadItem {
	key: string;
	order: number;
}

export interface DockItem {
	key: string;
	slot: number;
}

export interface LaunchLayout {
	schemaVersion: number;
	items: LaunchpadItem[];
	hidden: string[];
	version: number;
	updatedAt: number;
	updatedBy: string;
}

export interface DockLayout {
	schemaVersion: number;
	items: DockItem[];
	capacity?: number;
	version: number;
	updatedAt: number;
	updatedBy: string;
}

export interface LayoutEnvelope<L> {
	seeded: boolean;
	layout: L | null;
	version: number;
}

// Bundled launch + dock envelopes returned inside GET /server/init so the
// desktop boot path can paint both surfaces without separate round-trips.
export interface InitLayout {
	launch: LayoutEnvelope<LaunchLayout>;
	dock: LayoutEnvelope<DockLayout>;
}

// Key-encoding helpers. Mirror server logic: 'a:<id>' for apps, 'b:<id>' for
// bookmarks.
//
// For apps we encode the *entrance id* (AppInfo.id). This is appid-derived and
// therefore stable across a user's devices: for a single-entrance app the
// entrance id IS the appid; for a multi-entrance app it is
// `appid + entranceIndex`.
export type KeyKind = 'a' | 'b';

export function encodeKey(kind: KeyKind, id: string): string {
	return `${kind}:${id}`;
}

export interface ParsedKeyRef {
	kind: string;
	id: string;
}

export function parseKey(key: string): ParsedKeyRef | null {
	if (typeof key !== 'string') return null;
	const colon = key.indexOf(':');
	if (colon <= 0 || colon >= key.length - 1) return null;
	return { kind: key.slice(0, colon), id: key.slice(colon + 1) };
}

// The live desktop render pipeline (relocate_application_place / dockerApps)
// keys every item by a prefix-stripped "raw id" (see normalizeAppId). Apps use
// their entrance id verbatim; bookmarks are namespaced with `bm:` so they can
// flow through the same pipeline without colliding with app ids. These helpers
// convert between a raw id and the persisted layout key ('a:'/'b:').
export const BOOKMARK_RAW_PREFIX = 'bm:';

export function isBookmarkRawId(rawId: string): boolean {
	return rawId.startsWith(BOOKMARK_RAW_PREFIX);
}

export function bookmarkIdFromRawId(rawId: string): string {
	return rawId.slice(BOOKMARK_RAW_PREFIX.length);
}

export function rawIdToKey(rawId: string): string {
	return isBookmarkRawId(rawId)
		? encodeKey('b', bookmarkIdFromRawId(rawId))
		: encodeKey('a', rawId);
}

export function keyToRawId(parsed: ParsedKeyRef): string {
	return parsed.kind === 'b' ? BOOKMARK_RAW_PREFIX + parsed.id : parsed.id;
}

export type BookmarkOpenMode = 'iframe' | 'window';

export interface Bookmark {
	id: string;
	name: string;
	icon: string | null;
	url: string;
	openMode: BookmarkOpenMode;
	createdAt: number;
	updatedAt: number;
}

export interface BookmarksDoc {
	schemaVersion: number;
	bookmarks: Bookmark[];
	version: number;
	updatedAt: number;
}

export type DisplayItemKind = 'app' | 'bookmark' | 'unknown';

export interface DisplayItem {
	kind: DisplayItemKind;
	key: string;
	id: string;
	icon: string;
	name: string;
	title: string;
	url?: string;
	openMethod?: string;
	source?: any;
}
