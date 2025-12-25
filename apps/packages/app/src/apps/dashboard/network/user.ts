import { AxiosResponse } from 'axios';
import { api } from '@apps/dashboard/boot/axios';

export const getUserInto = (): Promise<AxiosResponse<any>> => {
	return api.get('/api/profile/init');
};
