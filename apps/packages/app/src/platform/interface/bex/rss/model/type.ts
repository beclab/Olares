export interface Icon {
	id: number;
	hash: string;
	mime_type: string;
	content: string;
}

export interface Board {
	id: number;
	user_id: number;
	title: string;
	description: string;
}

export interface Category {
	id: number;
	title: string;
	user_id: number;
	hide_globally: boolean;
}

export interface Entry {
	_id: string;

	algorithms: string[];
	feed_id?: string;
	sources: string[];

	url: string;
	title?: string;
	author?: string;
	full_content?: string;
	raw_content?: string;
	image_url: string;

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

	batch_id?: number;

	local_file_path: string;
	file_type: string;
	extract: boolean;
	language: string;
	download_faiure: boolean;
	__v: string;
}

export interface Feed {
	_id: string;
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
	//createAt: string;

	updatedAt: string;

	//local
	isTrend: boolean;
}

export interface FeedCreationRequest {
	feed_url: string;
	source: string;
	user_agent?: string;
	cookie?: string;
	username?: string;
	password?: string;
	crawler?: boolean;
	disabled?: boolean;
	ignore_http_cache?: boolean;
	allow_self_signed_certificates?: boolean;
	fetch_via_proxy?: boolean;
	scraper_rules?: string;
	rewrite_rules?: string;
	blocklist_rules?: string;
	keeplist_rules?: string;
	hide_globally?: boolean;
	urlrewrite_rules?: string;
}

export interface CreateEntry {
	id?: string;
	url?: string;
	source?: string;
}
