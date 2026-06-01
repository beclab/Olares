import { useTokenStore } from 'src/stores/settings/token';
import axios from 'axios';

export async function getSMBAccountList(): Promise<any> {
	const tokenStore = useTokenStore();
	return await axios.get(`${tokenStore.url}/api/files/smb_share_user/`);
}

export async function addSMBAccount(
	user: string,
	password: string
): Promise<any> {
	const tokenStore = useTokenStore();
	return await axios.post(`${tokenStore.url}/api/files/smb_share_user/`, {
		user,
		password
	});
}

export async function deleteSMBAccount(users: string[]): Promise<any> {
	const tokenStore = useTokenStore();
	return await axios.delete(`${tokenStore.url}/api/files/smb_share_user/`, {
		data: {
			users
		}
	});
}
