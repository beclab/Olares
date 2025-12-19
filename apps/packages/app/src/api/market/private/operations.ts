import axios, { AxiosProgressEvent } from 'axios';
import { useCenterStore } from 'src/stores/market/center';
import { MARKET_SOURCE_OFFICIAL } from 'src/constant/constants';
import { CloneEntrance, UpdateEnvBody } from 'src/constant';

export interface BaseOperationRequest {
	app_name: string;
	source: string;
}

export interface OperationRequest extends BaseOperationRequest {
	version: string;
	sync?: boolean;
	envs?: UpdateEnvBody;
}

export interface CSV2Request extends OperationRequest {
	//only uninstall and only cs v2
	all: boolean;
}

//no version
export interface CloneRequest extends BaseOperationRequest {
	sync?: boolean;
	envs?: UpdateEnvBody;
	//only clone
	title?: string;
	entrances?: CloneEntrance[];
}

export async function installApp(request: OperationRequest): Promise<any> {
	const store = useCenterStore();
	const url = store.appUrl + '/apps/' + request.app_name + '/install';
	return await axios.post(url, request);
}

export async function cloneApp(request: CloneRequest): Promise<any> {
	const store = useCenterStore();
	const url = store.appUrl + '/apps/' + request.app_name + '/clone';
	return await axios.post(url, request);
}

export async function cancelInstalling(
	request: OperationRequest
): Promise<boolean> {
	const store = useCenterStore();
	const url = store.appUrl + '/apps/' + request.app_name + '/install';
	return await axios.delete(url, {
		data: request
	});
}

export async function uninstallApp(request: CSV2Request): Promise<any> {
	const store = useCenterStore();
	const url = store.appUrl + '/apps/' + request.app_name;
	return await axios.delete(url, {
		data: request
	});
}

export async function upgradeApp(request: OperationRequest): Promise<any> {
	const store = useCenterStore();
	const url = store.appUrl + '/apps/' + request.app_name + '/upgrade';
	return await axios.put(url, request);
}

export async function resumeApp(appName: string): Promise<any> {
	const store = useCenterStore();
	const url = store.appUrl + '/apps/resume';
	return await axios.post(url, {
		appName
	});
}

export async function stopApp(appName: string, all: boolean): Promise<any> {
	const store = useCenterStore();
	const url = store.appUrl + '/apps/stop';
	return await axios.post(url, {
		appName,
		all
	});
}

export async function marketLogs(size = 100): Promise<any> {
	const store = useCenterStore();
	const url = store.appUrl + '/logs?size=' + size;
	const { data }: any = await axios.get(url);
	console.log(data);
	return data;
}

export async function uploadLocalPackage(
	file: any,
	onProgress: (progressEvent: AxiosProgressEvent) => void = () => {},
	config: { signal?: AbortSignal } = {}
): Promise<{ data: any; message: string }> {
	try {
		const store = useCenterStore();
		const formData = new FormData();
		formData.append('chart', file);
		formData.append('source', MARKET_SOURCE_OFFICIAL.LOCAL.UPLOAD);
		const url = store.appUrl + '/apps/upload';

		const response: any = await axios.post(url, formData, {
			headers: { 'Content-Type': 'multipart/form-data' },
			onUploadProgress: (progressEvent: AxiosProgressEvent) => {
				onProgress(progressEvent);
			},
			...config
		});

		console.log(response);
		if (response.data) {
			return { data: response.data, message: '' };
		}
		return { data: null, message: response.message ?? 'Upload response error' };
	} catch (e: any) {
		console.error(e.response);
		return {
			data: null,
			message: e.response?.data?.message || 'Upload failed'
		};
	}
}

export async function removeLocalPackage(
	request: OperationRequest
): Promise<any> {
	const store = useCenterStore();
	const url = store.appUrl + '/local-apps/delete';
	const { data }: any = await axios.delete(url, {
		data: {
			app_name: request.app_name,
			app_version: request.version
		}
	});
	console.log(data);
	return true;
}
