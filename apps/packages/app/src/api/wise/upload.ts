import { useConfigStore } from '../../stores/rss-config';
import axios from 'axios';

export async function uploadCreateEntry(
	path: string,
	name: string,
	file_type: string
) {
	try {
		const configStore = useConfigStore();
		const data = await axios.post(
			configStore.url + '/knowledge/upload/new_upload',
			{
				local_file_path: path,
				local_file_name: name,
				file_type: file_type
			}
		);
		console.log(data);
		return data;
	} catch (e: any) {
		console.log(e.message);
		return undefined;
	}
}

export async function uploadDeleteEntry(id: string, withFile: boolean) {
	try {
		const configStore = useConfigStore();
		const data = await axios.delete(
			configStore.url + '/knowledge/upload/' + id,
			{
				data: {
					file_remove_flag: withFile
				}
			}
		);
		console.log(data);
		return data;
	} catch (e: any) {
		console.log(e.message);
		return undefined;
	}
}
