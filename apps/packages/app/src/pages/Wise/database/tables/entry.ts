import { Entry } from 'src/utils/rss-types';
import { sendMessageToWorker } from '../sqliteService';
import { convertIntegerPropertiesToBoolean } from '../utils';

const defaultEntryKeyOrder = [
	'id',
	'algorithms',
	'feed_id',
	'sources',
	'url',
	'title',
	'author',
	'full_content',
	'summary',
	'raw_content',
	'image_url',
	'last_opened',
	'attachment',
	'readlater',
	'crawler',
	'status',
	'starred',
	'disabled',
	'saved',
	'progress',
	'remaining_time',
	'played_time',
	'unread',
	'published_at',
	'createdAt',
	'updatedAt',
	'source',
	'ranked',
	'score',
	'impression',
	'impression_id',
	'keywords',
	'batch_id',
	'local_file_path',
	'file_type',
	'extract',
	'language',
	'download_faiure',
	'extra',
	'__v',
	'debug_recommend_info'
];

const booleanProperties = [
	'attachment',
	'readlater',
	'crawler',
	'starred',
	'disabled',
	'saved',
	'unread',
	'ranked',
	'extract',
	'download_faiure'
];

function convertEntry(feeds: Entry[]) {
	return feeds.map((filter: any) =>
		convertIntegerPropertiesToBoolean(filter, booleanProperties)
	);
}

function entryToParams(entry: Entry, keysOrder = defaultEntryKeyOrder) {
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

export function createEntryTable(db: any): string {
	db.exec(`
		CREATE TABLE IF NOT EXISTS entries
		(
			id
			TEXT
			PRIMARY
			KEY,
			algorithms
			TEXT,
			feed_id
			TEXT,
			sources
			TEXT,
			url
			TEXT,
			title
			TEXT,
			author
			TEXT,
			full_content
			TEXT,
			summary
		  TEXT,
			raw_content
			TEXT,
			image_url
			TEXT,
			last_opened
			INTEGER,
			attachment
			INTEGER,
			readlater
			INTEGER,
			crawler
			INTEGER,
			status
			TEXT,
			starred
			INTEGER,
			disabled
			INTEGER,
			saved
			INTEGER,
			progress
			INTEGER,
			remaining_time
			INTEGER,
			played_time,
			INTEGER,
			unread
			INTEGER,
			published_at
			INTEGER,
			createdAt
			TEXT,
			updatedAt
			TEXT,
			source
			TEXT,
			ranked
			INTEGER,
			score
			REAL,
			impression
			INTEGER,
			impression_id
			TEXT,
			keywords
			TEXT,
			batch_id
			INTEGER,
			local_file_path
			TEXT,
			file_type
			TEXT,
			extract
			INTEGER,
			language
			TEXT,
			download_faiure
			INTEGER,
			extra
			TEXT,
			__v
			TEXT,
			debug_recommend_info
			TEXT
		);
	`);
	return 'entries';
}

export async function addOrUpdateEntries(entries: Entry[]) {
	if (entries.length === 0) {
		console.log('Entry update length 0 return');
		return;
	}
	try {
		const params = entries.map((item) => entryToParams(item));
		console.log(params);
		console.log(
			'Param types:',
			params.map((param) => typeof param)
		);
		await sendMessageToWorker('transaction', {
			sql: `INSERT INTO entries (id, algorithms, feed_id, sources, url, title, author, full_content, summary, raw_content,
																 image_url, last_opened, attachment, readlater, crawler,status, starred, disabled, saved,
                                 progress, remaining_time, played_time, unread, published_at, createdAt, updatedAt,
																 source, ranked, score, impression, impression_id, keywords, batch_id, local_file_path,
																 file_type, extract, language, download_faiure, extra, __v, debug_recommend_info)
						VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?,?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?,
						        ?, ?, ?, ?,?, ?) ON CONFLICT(id) DO
			UPDATE SET
				algorithms = excluded.algorithms,
				feed_id = excluded.feed_id,
				sources = excluded.sources,
				url = excluded.url,
				title = excluded.title,
				author = excluded.author,
				full_content = excluded.full_content,
				summary = excluded.summary,
				raw_content = excluded.raw_content,
				image_url = excluded.image_url,
				last_opened = excluded.last_opened,
				attachment = excluded.attachment,
				readlater = excluded.readlater,
				crawler = excluded.crawler,
				status = excluded.status,
				starred = excluded.starred,
				disabled = excluded.disabled,
				saved = excluded.saved,
				progress = excluded.progress,
				remaining_time = excluded.remaining_time,
				played_time = excluded.played_time,
				unread = excluded.unread,
				published_at = excluded.published_at,
				createdAt = excluded.createdAt,
				updatedAt = excluded.updatedAt,
				source = excluded.source,
				ranked = excluded.ranked,
				score = excluded.score,
				impression = excluded.impression,
				impression_id = excluded.impression_id,
				keywords = excluded.keywords,
				batch_id = excluded.batch_id,
				local_file_path = excluded.local_file_path,
				file_type = excluded.file_type,
				extract = excluded.extract,
				language = excluded.language,
				download_faiure = excluded.download_faiure,
				extra = excluded.extra,
				__v = excluded.__v,
				debug_recommend_info = excluded.debug_recommend_info;`,
			params: params
		});
		console.log('Entry added successfully!');
	} catch (error) {
		console.error('Failed to add entry:', error);
	}
}

export async function removeEntryById(id: string) {
	try {
		await sendMessageToWorker('execute', {
			sql: `DELETE
						FROM entries
						WHERE id = ?;`,
			params: [id]
		});
		console.log('remove entry id:', id);
	} catch (error) {
		console.error('Failed to get entry:', error);
	}
}

export async function removeEntryByUrl(url: string) {
	try {
		await sendMessageToWorker('execute', {
			sql: `DELETE
						FROM entries
						WHERE url = ?;`,
			params: [url]
		});
		console.log('remove entry url:', url);
	} catch (error) {
		console.error('Failed to get entry:', error);
	}
}

export async function updateEntryById(id: string, entry: Entry) {
	try {
		await sendMessageToWorker('execute', {
			sql: `UPDATE entries
						SET algorithms           = ?,
								feed_id              = ?,
								sources              = ?,
								url                  = ?,
								title                = ?,
								author               = ?,
								full_content         = ?,
								summary              = ?,
								raw_content          = ?,
								image_url            = ?,
								last_opened          = ?,
								attachment           = ?,
								readlater            = ?,
								crawler              = ?,
							  status               = ?,
								starred              = ?,
								disabled             = ?,
								saved                = ?,
								progress             = ?,
								remaining_time       = ?,
								played_time          = ?,
								unread               = ?,
								published_at         = ?,
								createdAt            = ?,
								updatedAt            = ?,
								source               = ?,
								ranked               = ?,
								score                = ?,
								impression           = ?,
								impression_id        = ?,
								keywords             = ?,
								batch_id             = ?,
								local_file_path      = ?,
								file_type            = ?,
								extract              = ?,
								language             = ?,
								download_faiure      = ?,
								extra                = ?,
								__v                  = ?,
								debug_recommend_info = ?
						WHERE id = ?;
			`,
			params: entryToParams(entry).splice(1).concat([id])
		});
		console.log('Entry update successfully!');
	} catch (error) {
		console.error('Failed to update entry:', error);
	}
}

export async function getEntryById(id: string): Promise<Entry | null> {
	try {
		const entries: any = await sendMessageToWorker('query', {
			sql: `SELECT entries.*
						FROM entries
						WHERE id = ?;`,
			params: [id]
		});
		// console.info(`Query entry ${id} successfully:`, entry);
		const coverList = convertEntry(entries);
		return coverList.length > 0 ? (coverList[0] as Entry) : null;
	} catch (error) {
		console.error('Failed to get entry:', error);
		return null;
	}
}

export async function getEntriesByFeedId(feedId: string): Promise<Entry[]> {
	try {
		const entries: any = await sendMessageToWorker('query', {
			sql: `SELECT entries.*
						FROM entries
						WHERE feed_id = ?;`,
			params: [feedId]
		});
		const coverList = convertEntry(entries);
		console.info(`Query entries feedId ${feedId} successfully:`, entries);
		return coverList.length > 0 ? (coverList as Entry[]) : [];
	} catch (error) {
		console.error('Failed to get entries:', error);
		return [];
	}
}

export async function clearEntries() {
	try {
		await sendMessageToWorker('execute', {
			sql: `DELETE
						FROM entries`
		});
		console.log('Entry clear successfully!');
	} catch (error) {
		console.error('Failed to clear entry:', error);
	}
}
