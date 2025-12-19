import { getFileIcon } from '@bytetrade/core';

export const NODE_PHASE = {
	PENDING: 'Pending',
	RUNNING: 'Running',
	SUCCEEDED: 'Succeeded',
	SKIPPED: 'Skipped',
	FAILED: 'Failed',
	ERROR: 'Error',
	OMITTED: 'Omitted'
};
export type NodePhase =
	| ''
	| 'Pending'
	| 'Running'
	| 'Succeeded'
	| 'Skipped'
	| 'Failed'
	| 'Error'
	| 'Omitted';

export interface LogEntry {
	content?: string;
	podName?: string;
}

export interface ArtifactLogEntry {
	level: string;
	ts: string;
	msg: string;
	caller: string;
	url: string;
	body: string;
}

export interface RemoveData {
	id: number;
	remove_id: string;
	remove_type: RemoveType;
	bfl_user: string;
	created_at: string;
}

export enum RemoveType {
	Feed = 1,
	Entry = 2,
	Label = 3,
	Note = 4
}

export enum SOURCE_TYPE {
	WISE = 'wise',
	LIBRARY = 'library',
	TERMIPASS = 'termipass'
}

export enum COOKIE_LEVEL {
	REQUIRED = 'required',
	RECOMMEND = 'recommend'
}

export enum CookieStatusCode {
	COOKIE_NOT_UPLOADED = 0,
	COOKIE_EXPIRED = 1,
	COOKIE_UPLOADED = 2
}

export enum DefaultType {
	Limit = 20
}

export interface UrlCheck {
	entry_exist: boolean;
	cookie_require: string;
	cookie_exist: boolean;
	cookieExpired: boolean;
	exist_entry_id?: string;
	is_entry_available: string;
}

export interface FileInfo extends UrlCheck {
	file: string;
	file_type: string;
	download_url: string;
}

export const DRIVER_FILE_PREFIX = '/Files/Home/';

export enum DOWNLOAD_OPERATE {
	PAUSE = 'task_pause',
	RETRY = 'task_continue',
	CANCEL = 'task_cancel',
	REMOVE = 'task_remove'
}

export enum DOWNLOAD_RECORD_STATUS {
	COMPLETE = 'complete',
	ERROR = 'error',
	DOWNLOADING = 'downloading',
	WAITING = 'waiting',
	PAUSED = 'paused',
	CANCELLED = 'cancel',
	REMOVE = 'remove',
	LOSS = 'loss'
}

export enum FILE_TYPE {
	ARTICLE = 'article',
	VIDEO = 'video',
	AUDIO = 'audio',
	PDF = 'pdf',
	EBOOK = 'ebook',
	GENERAL = 'general'
}

export function getFileTypeByName(name: string) {
	console.log(name);
	const fileType = getFileIcon(name);
	console.log(fileType);
	switch (fileType) {
		case 'pdf':
			return FILE_TYPE.PDF;
		case 'video':
			return FILE_TYPE.VIDEO;
		case 'audio':
			return FILE_TYPE.AUDIO;
		case 'epub':
			return FILE_TYPE.EBOOK;
		default:
			return '';
	}
}

export interface Enclosure {
	id: string;
	name: string;
	entry_id: string;
	content: string;
	mime_type: string;
	url: string;
	local_file_path: string;
	download_status: string;
}

export interface Algorithm {
	entry: Entry;
	entry_id: string;
	source: string;
	ranked: boolean;
	score: number;
	impression: number;
	extra: Record<string, string>;
}

export interface UpdateImpression {
	id?: string;
	entry_id?: string;
	source?: string;
	clicked?: boolean;
	stared?: boolean;
	read_finish?: boolean;
	read_time?: number;
}

export interface ImpressionAction {
	clicked?: boolean;
	stared?: boolean;
	read_finish?: boolean;
	read_time?: number;
}

export interface SimpleEntry {
	id: string;
	source: string;
	published_at: number;
	author?: string;
	full_content?: string;
	title?: string;
	url?: string;
	feed_id?: string;
	readlater: boolean;
	unread: boolean;
	image_url?: string;
	crawler: boolean;
	status: string;
	file_type: string;
	local_file_path: string;
	createdAt: string;
	updatedAt: string;
}

export interface Entry {
	id: string;

	algorithms: string[];
	feed_id?: string;
	sources: string[];

	url: string;
	title?: string;
	author?: string;
	full_content?: string;
	raw_content?: string;
	image_url: string;
	last_opened: number;
	progress: number;
	played_time: number;
	remaining_time: number;

	attachment: boolean;
	readlater: boolean;
	crawler: boolean;
	starred: boolean;
	disabled: boolean;
	saved: boolean;
	unread: boolean;
	published_at: number;
	//reading_time: number;
	createdAt: string;
	updatedAt: string;

	source?: string;
	ranked?: boolean;
	score?: number;
	impression?: number;
	impression_id?: string;
	keywords?: string[];
	status: string;

	batch_id?: number;

	local_file_path: string;
	file_type: string;
	extract: boolean;
	language: string;
	download_faiure: boolean;
	extra: EntryExtra;
	__v: string;
	debug_recommend_info: any;

	//frontend
	summary: string;
}

export interface EntryExtra {
	embedding: any[];
	reason_data: RankEntry[];
	reason_type: string;
	prerank_score: number;
}

export enum ENTRY_STATUS {
	Empty = 'empty',
	Waiting = 'waiting',
	Crawling = 'crawling',
	Extracting = 'extracting',
	Extracted = 'extracted',
	Staging = 'staging',
	Completed = 'completed',
	Failed = 'failed'
}

export enum REASON_TYPE {
	KEYWORD = 'KEYWORD',
	ARTICLE = 'ARTICLE'
}

export interface RankEntry {
	id?: string;
	url?: string;
	title?: string;
	keyword?: string;
}

export interface EntriesStatusUpdateRequest {
	entry_ids: string[];
	status: boolean;
}

export interface Feed {
	id: string;
	sources: string[];
	feed_url: string;
	site_url: string;
	title: string;
	description: string;
	checked_at: string;
	next_check_at: string;
	etag_header: string;
	last_modified_header: string;
	parsing_error_message: string;
	parsing_error_count: number;
	scraper_rules: string;
	rewrite_rules: string;
	crawler: boolean;
	blocklist_rules: string;
	keeplist_rules: string;
	urlrewrite_rules: string;
	user_agent: string;
	cookie: string;
	username: string;
	password: string;
	disabled: boolean;
	ignore_http_cache: boolean;
	allow_self_signed_certificates: boolean;
	fetch_via_proxy: boolean;
	icon_content: string;
	icon_type: string;
	hide_globally: boolean;
	unread_count: number;
	read_count: number;
	create_at: string;
	updated_at: string;
	auto_download: boolean;
}

export interface SearchFeed {
	id?: string;
	feed_url: string;
	site_url: string;
	title: string;
	description: string;
	icon_content: string;
	is_subscribed: boolean;
	icon_type: string;
	create_at: string;
	updated_at: string;
}

export interface RecommendAlgorithm {
	id: string;
	title: string;
}

export class FeedQuery {
	offset = 0;

	limit = 1000;

	source?: string;

	constructor(props?: Partial<FeedQuery>) {
		props && Object.assign(this, props);
	}

	public build(): string {
		let url = '';
		if (this.offset) {
			if (url) {
				url += '&';
			} else {
				url += '?';
			}
			url += `offset=${this.offset}`;
		}
		if (this.limit) {
			if (url) {
				url += '&';
			} else {
				url += '?';
			}
			url += `limit=${this.limit}`;
		}
		if (this.source) {
			if (url) {
				url += '&';
			} else {
				url += '?';
			}
			url += `source=${this.source}`;
		}
		return url;
	}
}

export class SDKQueryRequest {
	path?: string;

	constructor(props?: Partial<SDKQueryRequest>) {
		props && Object.assign(this, props);
	}

	public build(): string {
		let url = '';
		if (this.path) {
			if (url) {
				url += '&';
			} else {
				url += '?';
			}
			url += `path=${this.path}`;
		}

		return url;
	}
}

export interface SDKSearchPathResponse {
	atomlink: string;
	item: SDKSearchPathItem[];
	lastBuildDate: string;
	link: string;
	title: string;
	ttl: number;
	updated: string;
	logo?: string;
}

export interface SDKSearchPathItem {
	author: string;
	description: string;
	link: string;
	pubDate: string;
	title: string;
}

export interface CreateEntry {
	id?: string;
	url?: string;
	source?: string;
}

export interface RssContentQueryItem {
	name: string;
	entry_id: number;
	created: string;
	feed_infos: {
		feed_id: number;
		feed_name: string;
		feed_icon: string;
	}[];
	brders: {
		name: string;
		id: number;
	}[];
	docId: string;
	snippet: string;
}

export interface RssContentQuery {
	code: number;
	message: string;
	data: {
		code: number;
		data: {
			count: number;
			offset: number;
			limit: number;
			items: RssContentQueryItem[];
		};
	};
}

export interface Env {
	name: string;
	value?: string;
	valueFrom?: {
		configMapKeyRef: ConfigMapKeyRef;
	};
}

interface ConfigMapKeyRef {
	name: string;
	key: string;
}

export interface Label {
	id: string;
	name: string;
	entries: string[];
	notes: string[];
	deleted: boolean;
	updated_at: string;
}

export interface Note {
	id: string;
	entry_id: string;
	highlight: string;
	content: string;
	start: number;
	length: number;
	deleted: boolean;
	create_at: string;
	updated_at: string;
}

export interface CreateNote {
	entry_id: string;
	highlight: string;
	content: string;
	start: number;
	length: number;
}

export enum THEME_TYPE {
	DARK = 'dark',
	LIGHT = 'light',
	AUTO = 'auto'
}

export enum SORT_TYPE {
	CREATED = 'created',
	UPDATED = 'updated',
	PUBLISHED = 'published'
}
export enum ORDER_TYPE {
	ASC = 'asc',
	DESC = 'desc'
}
export enum SPLIT_TYPE {
	LOCATION = 'location',
	SEEN = 'seen',
	NONE = 'none'
}

export interface FilterInfo {
	id: string;
	icon: string;
	name: string;
	query: string;
	description: string;
	showbadge: boolean;
	pin: boolean;
	serial_no: number;
	system: boolean;
	created_at?: string;
	updated_at?: string;
	sortby?: SORT_TYPE;
	orderby?: ORDER_TYPE;
	splitview?: SPLIT_TYPE;
}

export interface ArticleTopic {
	text: string;
	id: string;
	level: number;
	jump: boolean;
}
