import { AxiosResponse } from 'axios';
import { api } from '@apps/dashboard/boot/axios';
import {
	GPUNodeList,
	GraphicsDetailsParams,
	GraphicsDetailsResponse,
	GraphicsListParams,
	GraphicsListResponse,
	InstantVectorParams,
	InstantVectorResponse,
	RangeVectorParams,
	RangeVectorResponse,
	SettingsGPUResponse,
	TaskDetailParams,
	TaskDetailResponse,
	TaskListParams,
	TaskListResponse
} from '@apps/dashboard/src/types/gpu';

const apiPrefix = '/hami/api/vgpu';

export const getGraphicsList = (
	params: GraphicsListParams
): Promise<AxiosResponse<GraphicsListResponse>> => {
	return api.post(`${apiPrefix}/v1/gpus`, params);
};

export const getTaskList = (
	params: TaskListParams
): Promise<AxiosResponse<TaskListResponse>> => {
	return api.post(`${apiPrefix}/v1/containers`, params);
};

export const getGraphicsDetails = (
	params: GraphicsDetailsParams
): Promise<AxiosResponse<GraphicsDetailsResponse>> => {
	return api.get(`${apiPrefix}/v1/gpu`, {
		params
	});
};

export const getInstantVector = (
	params: InstantVectorParams
): Promise<AxiosResponse<InstantVectorResponse>> => {
	return api.post(`${apiPrefix}/v1/monitor/query/instant-vector`, params);
};

export const getRangeVector = (
	params: RangeVectorParams
): Promise<AxiosResponse<RangeVectorResponse>> => {
	return api.post(`${apiPrefix}/v1/monitor/query/range-vector`, params);
};

export const getTaskDetail = (
	params: TaskDetailParams
): Promise<AxiosResponse<TaskDetailResponse>> => {
	return api.get(`${apiPrefix}/v1/container`, {
		params
	});
};

export const getGPUNodeList = (
	params: GraphicsListParams
): Promise<AxiosResponse<GPUNodeList>> => {
	return api.post(`${apiPrefix}/v1/nodes`, params);
};

// export const getSettingsGPUList = (): Promise<
// 	AxiosResponse<SettingsGPUResponse>
// > => {
// 	return api.get(`/api/gpu/list`);
// };
