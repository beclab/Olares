import axios from 'axios';
import { DownloadProgress } from 'src/pages/Mobile/collect/utils';
import { useCollectStore } from 'src/stores/collect';
import { DownloadRecord } from '.';
import { DownloadRecordListRequest } from 'src/api/wise/download';

export async function queryPdfEntry(url: string) {
	const collectStore = useCollectStore();
	try {
		const response: any = await axios.post(
			collectStore.baseUrl + '/knowledge/entry/query',
			{
				url,
				source: 'termipass',
				file_type: 'pdf'
			}
		);
		return !!(response && response.count > 0);
	} catch (e) {
		return false;
	}
}

export async function downloadPdf(
	url: string,
	filename: string
): Promise<any | null> {
	try {
		const collectStore = useCollectStore();
		const entry = await axios.post(
			collectStore.baseUrl + '/knowledge/pdf/download',
			{
				url,
				filename
			}
		);
		return entry;
	} catch (e: any) {
		return null;
	}
}

export async function getDownloadPdfProgress(
	id: string
): Promise<DownloadProgress | null> {
	try {
		const collectStore = useCollectStore();
		const progress: DownloadProgress = await axios.get(
			collectStore.baseUrl + '/knowledge/pdf/download/progress/' + id
		);
		return progress;
	} catch (e: any) {
		return null;
	}
}

export async function getDownloadHistory(
	req: DownloadRecordListRequest
): Promise<DownloadRecord[]> {
	try {
		const collectStore = useCollectStore();
		const history: DownloadRecord[] = await axios.get(
			collectStore.baseUrl + '/knowledge/download/task_query' + req.toString()
		);
		console.log(history);
		return history;
	} catch (e: any) {
		console.log(e.message);
		return [];
	}
}
