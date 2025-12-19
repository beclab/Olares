export interface DownloadRecord {
	created_time: string;
	download_app: string;
	download_provdider: string;
	enclosure_id: string;
	entry_id: string;
	file_type: string;
	finished_download_time: string;
	id: number;
	input_extra: string;
	link_type: string;
	name: string;
	output_extra: string;
	path: string;
	progress: string;
	downloaded_bytes: number;
	provider_task_id: string;
	size: number;
	status: string;
	task_user: string;
	update_time: string;
	url: string;
	mimeType?: string;
	startTime?: number;
}
