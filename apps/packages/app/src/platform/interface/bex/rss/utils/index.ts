import { FileInfo } from 'src/utils/rss-types';
import { DownloadRecord } from 'src/utils/interface/rss';

export type Message =
	| { type: 'getAllRSS'; tabId: number }
	| { type: 'setPageRSS'; feeds: any }
	| {
			type: 'addPageRSS';
			feed: { url: string; title: string; image: string };
	  };

export const rsshubDomain = '{rsshubDomain}';
export const rsshubReplaceDomain = 'http://127.0.0.1:3010/rss';

export interface DownloadFileRecord {
	file: FileInfo;
	record?: DownloadRecord;
	title: string;
}

export { DownloadRecord, FileInfo };
