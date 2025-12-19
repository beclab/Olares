import { AxiosResponse } from 'axios';
import { api } from '@apps/dashboard/boot/axios';
import {
	SystemIFSResponse,
	SystemIFSParams,
	SystemFanResponse
} from '@apps/dashboard/src/types/network';

export const getSystemIFS = (
	params: SystemIFSParams
): Promise<AxiosResponse<SystemIFSResponse>> => {
	return api.get(`/capi/system/ifs`, {
		params
	});
};
