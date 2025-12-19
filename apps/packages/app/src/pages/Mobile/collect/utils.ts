import { Feed, UrlCheck } from 'src/utils/rss-types';

export interface BaseCollectInfo {
	id?: string;
	title: string;
	url: string;
	image: string;
}

export enum PageStatus {
	none,
	error,
	loading,
	cancel,
	success
}

export interface PageInfo extends BaseCollectInfo {
	status: PageStatus;
	progress?: any; // 0-1
}

export enum RssStatus {
	none,
	added,
	removed
}

export interface RssInfo extends BaseCollectInfo {
	status?: RssStatus;
	feed?: Feed;
	cookie?: UrlCheck;
}
export interface DownloadProgress {
	id: string;
	url: string;
	fileName: string;
	status: string;
	total: number;
	download: number;
	//local data
	isLoading: boolean;
}

export enum FILE_TYPE {
	ARTICLE = 'article',
	VIDEO = 'video',
	AUDIO = 'audio',
	PDF = 'pdf',
	EBOOK = 'ebook',
	GENERAL = 'general'
}
