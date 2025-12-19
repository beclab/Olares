import axios from 'axios';

export async function openApplication(intent: any): Promise<boolean> {
	const url = '/server/intent/send';
	const { data }: any = await axios.post(url, { ...intent });
	console.log(data);
	return true;
}
