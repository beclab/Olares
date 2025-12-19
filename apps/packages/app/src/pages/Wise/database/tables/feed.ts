import { sendMessageToWorker } from '../sqliteService';
import { Feed } from '../../../../utils/rss-types';
import { convertIntegerPropertiesToBoolean } from '../utils';

const defaultFeedKeyOrder = [
	'id',
	'sources',
	'feed_url',
	'site_url',
	'title',
	'description',
	'checked_at',
	'next_check_at',
	'etag_header',
	'last_modified_header',
	'parsing_error_message',
	'parsing_error_count',
	'scraper_rules',
	'rewrite_rules',
	'crawler',
	'blocklist_rules',
	'keeplist_rules',
	'urlrewrite_rules',
	'user_agent',
	'cookie',
	'username',
	'password',
	'disabled',
	'ignore_http_cache',
	'allow_self_signed_certificates',
	'fetch_via_proxy',
	'icon_content',
	'icon_type',
	'hide_globally',
	'unread_count',
	'read_count',
	'create_at',
	'updated_at',
	'auto_download'
];

const booleanProperties = [
	'crawler',
	'disabled',
	'ignore_http_cache',
	'allow_self_signed_certificates',
	'hide_globally',
	'fetch_via_proxy',
	'auto_download'
];

function convertFeed(feeds: Feed[]) {
	return feeds.map((filter: any) =>
		convertIntegerPropertiesToBoolean(filter, booleanProperties)
	);
}

function feedToParams(entry: Feed, keysOrder = defaultFeedKeyOrder) {
	if (keysOrder) {
		return keysOrder.map((key) => {
			const value = entry[key];
			if (value === undefined || value === null) {
				return null;
			}
			//format
			if (typeof value === 'object') {
				return JSON.stringify(value);
			}
			if (typeof value === 'function') {
				throw new Error(
					`Invalid value for key "${key}": functions are not allowed`
				);
			}
			return value;
		});
	}
	return Object.values(entry);
}

export function createFeedTable(db: any): string {
	db.exec(`
		CREATE TABLE IF NOT EXISTS feeds (
			id TEXT PRIMARY KEY,
			sources TEXT,
			feed_url TEXT,
			site_url TEXT,
			title TEXT,
			description TEXT,
			checked_at TEXT,
			next_check_at TEXT,
			etag_header TEXT,
			last_modified_header TEXT,
			parsing_error_message TEXT,
			parsing_error_count INTEGER,
			scraper_rules TEXT,
			rewrite_rules TEXT,
			crawler INTEGER,
			blocklist_rules TEXT,
			keeplist_rules TEXT,
			urlrewrite_rules TEXT,
			user_agent TEXT,
			cookie TEXT,
			username TEXT,
			password TEXT,
			disabled INTEGER,
			ignore_http_cache INTEGER,
			allow_self_signed_certificates INTEGER,
			fetch_via_proxy INTEGER,
			icon_content TEXT,
			icon_type TEXT,
			hide_globally INTEGER,
			unread_count INTEGER,
			read_count INTEGER,
			create_at TEXT,
			updated_at TEXT,
			auto_download INTEGER
		);
	`);
	return 'feeds';
}

export async function addOrUpdateFeeds(feeds: Feed[]) {
	if (feeds.length === 0) {
		console.log('Feed update length 0 return');
		return;
	}
	try {
		const params = feeds.map((feed) => feedToParams(feed));
		await sendMessageToWorker('transaction', {
			sql: `
				INSERT INTO feeds (
					id, sources, feed_url, site_url, title, description, checked_at, next_check_at,
					etag_header, last_modified_header, parsing_error_message, parsing_error_count,
					scraper_rules, rewrite_rules, crawler, blocklist_rules, keeplist_rules,
					urlrewrite_rules, user_agent, cookie, username, password, disabled,
					ignore_http_cache, allow_self_signed_certificates, fetch_via_proxy,
					icon_content, icon_type, hide_globally, unread_count, read_count,
					create_at, updated_at, auto_download
				)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
				ON CONFLICT(id) DO UPDATE SET
					sources = excluded.sources,
					feed_url = excluded.feed_url,
					site_url = excluded.site_url,
					title = excluded.title,
					description = excluded.description,
					checked_at = excluded.checked_at,
					next_check_at = excluded.next_check_at,
					etag_header = excluded.etag_header,
					last_modified_header = excluded.last_modified_header,
					parsing_error_message = excluded.parsing_error_message,
					parsing_error_count = excluded.parsing_error_count,
					scraper_rules = excluded.scraper_rules,
					rewrite_rules = excluded.rewrite_rules,
					crawler = excluded.crawler,
					blocklist_rules = excluded.blocklist_rules,
					keeplist_rules = excluded.keeplist_rules,
					urlrewrite_rules = excluded.urlrewrite_rules,
					user_agent = excluded.user_agent,
					cookie = excluded.cookie,
					username = excluded.username,
					password = excluded.password,
					disabled = excluded.disabled,
					ignore_http_cache = excluded.ignore_http_cache,
					allow_self_signed_certificates = excluded.allow_self_signed_certificates,
					fetch_via_proxy = excluded.fetch_via_proxy,
					icon_content = excluded.icon_content,
					icon_type = excluded.icon_type,
					hide_globally = excluded.hide_globally,
					unread_count = excluded.unread_count,
					read_count = excluded.read_count,
					create_at = excluded.create_at,
					updated_at = excluded.updated_at,
					auto_download = excluded.auto_download;
			`,
			params: params
		});
		console.log('Feed added/updated successfully!');
	} catch (error) {
		console.error('Failed to add/update feed:', error);
	}
}

export async function updateFeedById(id: string, feed: Feed) {
	try {
		await sendMessageToWorker('execute', {
			sql: `UPDATE feeds
						SET sources                    = ?,
								feed_url                   = ?,
								site_url                   = ?,
								title                      = ?,
								description                = ?,
								checked_at                 = ?,
								next_check_at              = ?,
								etag_header                = ?,
								last_modified_header       = ?,
								parsing_error_message      = ?,
								parsing_error_count        = ?,
								scraper_rules              = ?,
								rewrite_rules              = ?,
								crawler                    = ?,
								blocklist_rules            = ?,
								keeplist_rules             = ?,
								urlrewrite_rules           = ?,
								user_agent                 = ?,
								cookie                     = ?,
								username                   = ?,
								password                   = ?,
								disabled                   = ?,
								ignore_http_cache          = ?,
								allow_self_signed_certificates = ?,
								fetch_via_proxy            = ?,
								icon_content               = ?,
								icon_type                  = ?,
								hide_globally              = ?,
								unread_count               = ?,
								read_count                 = ?,
								create_at                  = ?,
								updated_at                 = ?,
								auto_download              = ?
						WHERE id = ?;
			`,
			params: feedToParams(feed).splice(1).concat([id])
		});
		console.log('Feed updated successfully!');
	} catch (error) {
		console.error('Failed to update feed:', error);
	}
}

export async function getAllFeeds(): Promise<Feed[]> {
	try {
		const feeds: any = await sendMessageToWorker('query', {
			sql: `SELECT * FROM feeds;`
		});
		const convertList = convertFeed(feeds);
		return convertList.length > 0 ? (convertList as Feed[]) : [];
	} catch (error) {
		console.error('Failed to get all feeds:', error);
		return [];
	}
}

export async function getFeedByUrl(feed_url: string): Promise<Feed | null> {
	try {
		const feeds: any = await sendMessageToWorker('query', {
			sql: `SELECT * FROM feeds WHERE feed_url = ?;`,
			params: [feed_url]
		});
		const convertList = convertFeed(feeds);
		return convertList.length > 0 ? (convertList[0] as Feed) : null;
	} catch (error) {
		console.error('Failed to get feed:', error);
		return null;
	}
}

export async function getFeedById(id: string): Promise<Feed | null> {
	try {
		const feeds: any = await sendMessageToWorker('query', {
			sql: `SELECT * FROM feeds WHERE id = ?;`,
			params: [id]
		});
		const convertList = convertFeed(feeds);
		return convertList.length > 0 ? (convertList[0] as Feed) : null;
	} catch (error) {
		console.error('Failed to get feed:', error);
		return null;
	}
}

export async function removeFeedById(id: string) {
	try {
		await sendMessageToWorker('execute', {
			sql: `DELETE FROM feeds WHERE id = ?;`,
			params: [id]
		});
		console.log('Feed removed successfully!');
	} catch (error) {
		console.error('Failed to remove feed:', error);
	}
}

export async function clearFeeds() {
	try {
		await sendMessageToWorker('execute', {
			sql: `DELETE
						FROM feeds`
		});
		console.log('Feed clear successfully!');
	} catch (error) {
		console.error('Failed to clear feed:', error);
	}
}
