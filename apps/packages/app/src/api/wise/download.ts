import axios, { CancelToken } from 'axios';
import { useConfigStore } from 'src/stores/rss-config';
import { DefaultType, DOWNLOAD_OPERATE, FileInfo } from 'src/utils/rss-types';
import { DownloadRecord } from '../../utils/interface/rss';

export async function queryDownloadFile(
	url: string,
	cancelToken?: CancelToken
): Promise<FileInfo | undefined> {
	try {
		const configStore = useConfigStore();
		const info: FileInfo = await axios.get(
			configStore.url +
				'/knowledge/download/download_file_query?url=' +
				encodeURIComponent(url),
			{
				cancelToken
			}
		);
		console.log(info);
		return info;
	} catch (e: any) {
		console.log(e.message);
		return undefined;
	}
}

export async function downloadFile(
	name: string,
	file_type: string,
	download_url: string,
	path = ''
): Promise<DownloadRecord | undefined> {
	try {
		const configStore = useConfigStore();
		const file: DownloadRecord[] = await axios.post(
			configStore.url + '/knowledge/download/file_download',
			{
				url: download_url,
				name: name,
				file_type: file_type,
				path: path
			}
		);
		console.log(file);
		return file[0];
	} catch (e: any) {
		console.log(e.message);
		return undefined;
	}
}

export async function downloadFileNew(params: {
	name: string;
	download_url: string;
	path: string;
	format_id?: string;
	resolution?: string;
	file_type: string;
}): Promise<DownloadRecord | undefined> {
	try {
		const configStore = useConfigStore();
		const file: DownloadRecord = await axios.post(
			configStore.url + '/knowledge/download/file_download',
			{
				url: params.download_url,
				name: params.name,
				format_id: params.format_id,
				path: params.path,
				resolution: params.resolution,
				file_type: params.file_type
			}
		);
		console.log(file);
		return file;
	} catch (e: any) {
		console.log(e.message);
		return undefined;
	}
}

export async function getDownloadHistory(
	req: DownloadRecordListRequest
): Promise<DownloadRecord[]> {
	try {
		const configStore = useConfigStore();
		const response: any = await axios.get(
			configStore.url + '/knowledge/download/task_query' + req.toString()
		);
		return response ? response.items : [];
	} catch (e: any) {
		console.log(e.message);
		return [];
	}
}

export class DownloadRecordListRequest {
	offset: number;
	limit: number;
	entry_id: string | undefined;
	task_id: string | undefined;
	enclosure_id: string | undefined;

	constructor(
		entry_id?: string,
		task_id?: string,
		enclosure_id?: string,
		offset = 0,
		limit = DefaultType.Limit
	) {
		this.offset = offset;
		this.limit = limit;
		this.entry_id = entry_id;
		this.task_id = task_id;
		this.enclosure_id = enclosure_id;
	}

	toString() {
		let url = '?';
		if (this.entry_id) {
			url = url + 'entry_id=' + this.entry_id + '&';
		} else if (this.task_id) {
			url = url + 'task_id=' + this.task_id + '&';
		} else if (this.enclosure_id) {
			url = url + 'enclosure_id=' + this.enclosure_id + '&';
		}
		return url + 'offset=' + this.offset + '&limit=' + this.limit;
	}
}

export class DownloadRecordOperateRequest {
	task_id: string | undefined;
	opt: DOWNLOAD_OPERATE;
	remove?: boolean;

	constructor(
		task_id: string,
		operate: DOWNLOAD_OPERATE,
		removeFile?: boolean
	) {
		this.task_id = task_id;
		this.opt = operate;
		this.remove = removeFile;
	}

	toString() {
		let result = this.opt + '?task_id=' + this.task_id;
		if (this.opt === DOWNLOAD_OPERATE.REMOVE) {
			result = result + '&remove_flag=' + this.remove;
		}
		return result;
	}
}

export async function taskOperate(
	req: DownloadRecordOperateRequest
): Promise<boolean> {
	try {
		const configStore = useConfigStore();
		const history: DownloadRecord = await axios.get(
			configStore.url + '/knowledge/download/' + req.toString()
		);
		console.log(history);
		return true;
	} catch (e: any) {
		console.log(e.message);
		return false;
	}
}

export async function enclosuresRestart(
	enclosure_id: string
): Promise<boolean> {
	try {
		const configStore = useConfigStore();
		const result = await axios.post(
			configStore.url + '/knowledge/download/enclosures_restart_download',
			{ enclosure_id }
		);
		console.log(result);
		return true;
	} catch (e: any) {
		console.log(e.message);
		return false;
	}
}
