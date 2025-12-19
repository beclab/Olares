import { AxiosResponse } from 'axios';
import { api } from '../boot/axios';
import {
	AppListAllResponse,
	AppListResponse,
	SystemStatusResponse
} from './network';
import { SystemFanResponse } from '@apps/dashboard/src/types/network';

export const getUserList = (): Promise<AxiosResponse<any>> => {
	return api.get('/user-service/api/users');
};

export const getAppsList = (): Promise<AxiosResponse<AppListResponse>> => {
	return api.get('/user-service/api/myapps_v2');
};

export const getAppsListAll = (): Promise<
	AxiosResponse<AppListAllResponse>
> => {
	return api.get('/user-service/api/app/alluser/namespaces');
};

export const getSystemStatus = (): Promise<
	AxiosResponse<SystemStatusResponse>
> => {
	return api.get('/user-service/api/system/status');
};

export const getSystemFan = (): Promise<AxiosResponse<SystemFanResponse>> => {
	return api.get(`/user-service/api/mdns/olares-one/cpu-gpu`);
};
