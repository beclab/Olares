import { COOKIE_LEVEL } from 'src/utils/rss-types';
import { Entry } from 'src/utils/rss-types';
export enum DownloadStatusEnum {
	NOT_DOWNLOADED = 'notDownloaded',
	DOWNLOADING = 'downloading',
	COMPLETE = 'complete',
	ERROR = 'error',
	WAITING = 'waiting',
	REMOVE = 'remove',
	PAUSED = 'paused',
	CANCEL = 'cancel'
}
export interface FeedItem {
	id: string;
	title: string;
	feed_url: string;
	site_url: string;
	icon_type: string;
	icon_content: string;
	description: string;
	is_subscribed: boolean;
	loading?: boolean;
}

export interface DownloadItem {
	download_url: string;
	ext: string;
	file: string;
	file_type: string;
	filesize: number;
	id: string;
	resolution: string;
	tbr: number;
	download_status: `${DownloadStatusEnum}`;
	is_exist: boolean;
	task_id?: number;
	loading?: boolean;
	percent?: number;
	created_time?: string;
}

export interface CollectEntry {
	exist: boolean;
	collect: boolean;
	url: string;
	title: string;
	file_type: string;
	thumbnail: string;
	exist_entry_id?: string;
}

export interface DownloadInfoItem {
	title: string;
	thumbnail: string;
	list: DownloadItem[];
}
export interface CollectInfo {
	is_download_available: string;
	is_entry_available: string;
	is_feed_available: string;
	entry_plugin_dependency?: string[];
	download_plugin_dependency?: string[];
	feed_plugin_dependency?: string[];
	feed: FeedItem[];
	download: {
		source: string;
		info: DownloadInfoItem[];
	};
	entry: CollectEntry;
	cookie: {
		cookieRequire: `${COOKIE_LEVEL}`;
		cookieExist: boolean;
		cookieExpired: boolean;
		is_entry_available: string;
	};
}
