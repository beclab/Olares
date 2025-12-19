import axios, { AxiosResponse } from 'axios';
import { useUserStore } from 'src/stores/user';

const api = axios.create();

export const getAppsList = (): Promise<AxiosResponse<any>> => {
	const userStore = useUserStore();
	const baseURL = userStore.getModuleSever('settings');
	const domain = process.env.NODE_ENV === 'development' ? '' : baseURL;

	return api.get(`${domain}/api/myapps`);
};
