import { DifyModelInfo } from '../constant';
import axios from 'axios';

export async function get_model_list(): Promise<DifyModelInfo[]> {
	try {
		const data: any = await axios.get('/nitro/model');
		return data ? data : [];
	} catch (e) {
		console.log(e);
		return [];
	}
}

export async function install_model(id: string) {
	await axios.post('/nitro/model/' + encodeURIComponent(id));
}

export async function get_model_info(
	id: string
): Promise<DifyModelInfo | null> {
	try {
		const data: any = await axios.get('/nitro/model/' + encodeURIComponent(id));
		return data ? data : null;
	} catch (e) {
		console.log(e);
		return null;
	}
}

export async function load_model(id: string) {
	await axios.post('/nitro/model/' + encodeURIComponent(id) + '/load');
}

export async function stop_model(id: string) {
	await axios.post('/nitro/model/' + encodeURIComponent(id) + '/stop');
}

export async function delete_model(id: string) {
	await axios.delete('/nitro/model/' + encodeURIComponent(id));
}
