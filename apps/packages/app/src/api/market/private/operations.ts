import axios, { AxiosProgressEvent } from 'axios';
import { MARKET_SOURCE_OFFICIAL } from 'src/constant/constants';
import { CloneEntrance, UpdateEnvBody } from 'src/constant';
import { ComputeBindingItem, ComputeType } from 'src/constant/compute';
import { useAppStore } from 'src/stores/market/appStore';
import { i18n } from 'src/boot/i18n';

export interface BaseOperationRequest {
	app_name: string;
	source: string;
}

export interface OperationRequest extends BaseOperationRequest {
	version: string;
	sync?: boolean;
	envs?: UpdateEnvBody;
	selectedGpuType?: ComputeType;
}

export interface StopRequest extends BaseOperationRequest {
	//only uninstall and only cs v2
	all?: boolean;
}

export interface UninstallRequest extends BaseOperationRequest {
	version: string;
	//only uninstall and only cs v2
	all?: boolean;
	sync?: boolean;
	deleteData: boolean;
}

//no version
export interface CloneRequest extends BaseOperationRequest {
	sync?: boolean;
	envs?: UpdateEnvBody;
	selectedGpuType?: ComputeType;
	//only clone
	title?: string;
	entrances?: CloneEntrance[];
	// true when creating an instance from a template app (no installable body)
	templateClone?: boolean;
}

//no version, shared by resume and restart
export interface Resume2RestartRequest extends BaseOperationRequest {
	computeBinding?: ComputeBindingItem[];
}

export async function installApp(request: OperationRequest): Promise<any> {
	const store = useAppStore();
	const url = store.appUrl + '/apps/' + request.app_name + '/install';
	return await axios.post(url, request);
}

export async function cloneApp(request: CloneRequest): Promise<any> {
	const store = useAppStore();
	const url = store.appUrl + '/apps/' + request.app_name + '/clone';
	return await axios.post(url, request);
}

export async function cancelInstalling(
	request: OperationRequest
): Promise<boolean> {
	const store = useAppStore();
	const url = store.appUrl + '/apps/' + request.app_name + '/install';
	return await axios.delete(url, {
		data: request
	});
}

export async function uninstallApp(request: UninstallRequest): Promise<any> {
	const store = useAppStore();
	const url = store.appUrl + '/apps/' + request.app_name;
	return await axios.delete(url, {
		data: request
	});
}

export async function upgradeApp(request: OperationRequest): Promise<any> {
	const store = useAppStore();
	const url = store.appUrl + '/apps/' + request.app_name + '/upgrade';
	return await axios.put(url, request);
}

export async function resumeApp(request: Resume2RestartRequest): Promise<any> {
	const store = useAppStore();
	const url = store.appUrl + '/apps/resume';
	return await axios.post(url, {
		...request
	});
}

export async function validateResumeCompute(
	request: Resume2RestartRequest
): Promise<any> {
	const store = useAppStore();
	const url = store.appUrl + '/apps/' + request.app_name + '/bindings/validate';
	return await axios.post(url, {
		...request
	});
}

export async function restartApp(request: Resume2RestartRequest): Promise<any> {
	const store = useAppStore();
	const url = store.appUrl + '/apps/restart';
	return await axios.post(url, {
		...request
	});
}

export async function stopApp(request: StopRequest): Promise<any> {
	const store = useAppStore();
	const url = store.appUrl + '/apps/stop';
	return await axios.post(url, {
		...request
	});
}

export async function marketLogs(size = 100): Promise<any> {
	const store = useAppStore();
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
		const store = useAppStore();
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
		return {
			data: null,
			message: response.message ?? i18n.global.t('Upload response error')
		};
	} catch (e: any) {
		console.error(e.response);
		return {
			data: null,
			message: e.response?.data?.message || i18n.global.t('Upload failed')
		};
	}
}

export async function removeLocalPackage(
	request: OperationRequest
): Promise<any> {
	const store = useAppStore();
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
