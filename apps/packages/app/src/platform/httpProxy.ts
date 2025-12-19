import { HttpResponse } from '@capacitor/core';

import axios, {
	AxiosInstance,
	AxiosResponse,
	InternalAxiosRequestConfig,
	AxiosRequestConfig,
	Method
} from 'axios';
import { getAppPlatform } from '../application/platform';
import { getApplication } from 'src/application/base';
import { InOfflineText } from '../utils/checkTerminusState';
import { useTermipassStore } from 'src/stores/termipass';
import { TermiPassStatus } from 'src/utils/termipassState';

type RequestInterceptor = (
	config: InternalAxiosRequestConfig
) => InternalAxiosRequestConfig | Promise<InternalAxiosRequestConfig>;
type ResponseInterceptor = (
	response: AxiosResponse
) => AxiosResponse | Promise<AxiosResponse>;

export interface AxiosInstanceWithIntercept extends AxiosInstance {
	interceptRequest: (callback: RequestInterceptor) => void;
	interceptResponse: (callback: ResponseInterceptor) => void;

	requestIntercepts: RequestInterceptor[];
	responseIntercepts: ResponseInterceptor[];
}

const interceptors = {
	request: [] as RequestInterceptor[],
	response: [] as ResponseInterceptor[]
};

const totalHookMethod = [
	'get',
	'delete',
	'head',
	'options',
	'post',
	'put',
	'patch',
	'request'
];

export const axiosProxyHandler: ProxyHandler<AxiosInstanceWithIntercept> = {
	get: function (target, prop: Method) {
		const termipassStore = useTermipassStore();
		if (
			!getAppPlatform().isHookHttpRequest ||
			!totalHookMethod.includes(prop.toLowerCase()) ||
			termipassStore.totalStatus?.status == TermiPassStatus.OfflineMode
		) {
			return Reflect.get(target, prop);
		}
		return formatHookRequest(target, prop);
	},
	async apply(target, thisArg, argumentsList: InternalAxiosRequestConfig[]) {
		const [config] = argumentsList;

		const termipassStore = useTermipassStore();
		if (
			!getAppPlatform().isHookHttpRequest ||
			!totalHookMethod.includes(config.method?.toLowerCase() || 'get') ||
			termipassStore.totalStatus?.status == TermiPassStatus.OfflineMode
		) {
			return Reflect.apply(target, thisArg, argumentsList);
		}

		const reConfig = {
			...target.defaults,
			...config
		};
		return await requestCommonCallBack(
			target,
			config?.url || '',
			reConfig,
			(config?.method as Method) || ('get' as Method)
		);
	}
};

const formatHookRequest = (
	target: AxiosInstanceWithIntercept,
	prop: string
) => {
	if (['get', 'delete', 'head', 'options'].includes(prop)) {
		return async function (url: string, config?: InternalAxiosRequestConfig) {
			const reConfig = {
				...target.defaults,
				...config,
				url,
				method: prop as Method
			};
			return await requestCommonCallBack(
				target,
				url,
				reConfig as InternalAxiosRequestConfig,
				prop as Method
			);
		};
	} else if (['post', 'put', 'patch'].includes(prop)) {
		return async function (
			url: string,
			data?: any,
			config?: InternalAxiosRequestConfig
		) {
			const reConfig = {
				...target.defaults,
				...config,
				url,
				method: prop as Method,
				data
			};
			return await requestCommonCallBack(
				target,
				url,
				reConfig as InternalAxiosRequestConfig,
				prop as Method
			);
		};
	} else if (prop === 'request') {
		return async function (config?: InternalAxiosRequestConfig) {
			const reConfig = {
				...target.defaults,
				...config
			};
			return await requestCommonCallBack(
				target,
				config?.url || '',
				reConfig as InternalAxiosRequestConfig,
				(config?.method as Method) || ('get' as Method)
			);
		};
	}
};

class ErrorResponse {
	response: AxiosResponse<any>;
	message: string;
	constructor(response: AxiosResponse<any>, message: string) {
		this.response = response;
		this.message = message;
	}
	toString(): string {
		return this.message;
	}
}

const requestCommonCallBack = async (
	target: AxiosInstanceWithIntercept,
	url: string,
	config: InternalAxiosRequestConfig,
	method: Method
) => {
	for (const interceptor of interceptors.request) {
		config = (await interceptor(config)) || config;
	}

	for (const interceptor of target.requestIntercepts) {
		config = (await interceptor(config)) || config;
	}

	let fullUrl = '';
	if (url.startsWith('http') || !config.baseURL) {
		fullUrl = new URL(url, config.baseURL).toString();
	} else {
		if (url.startsWith('/') && config.baseURL.endsWith('/')) {
			fullUrl = config.baseURL + url.substring(1);
		} else if (!url.startsWith('/') && !config.baseURL.endsWith('/')) {
			fullUrl = config.baseURL + '/' + url.substring(1);
		} else {
			fullUrl = config.baseURL + url;
		}
	}

	const response: HttpResponse =
		await getAppPlatform().hookCapacitorHttp.request({
			method: method.toUpperCase(),
			url: fullUrl,
			params: config.params,
			data: config.data,
			headers: config.headers,
			responseType: config.responseType as any,
			connectTimeout: config.timeout
		});

	let axiosResponse: AxiosResponse = {
		data: response.data,
		status: response.status,
		statusText: '',
		headers: response.headers,
		config: config,
		request: null
	};

	for (const interceptor of interceptors.response) {
		axiosResponse = (await interceptor(axiosResponse)) || axiosResponse;
	}

	for (const interceptor of target.responseIntercepts) {
		axiosResponse = (await interceptor(axiosResponse)) || axiosResponse;
	}

	if (axiosResponse.status !== 200 && axiosResponse.status !== 201) {
		return Promise.reject(
			new ErrorResponse(
				axiosResponse,
				axiosResponse.data.message
					? `${axiosResponse.data.message}`
					: `Request failed with status code ${axiosResponse.status}`
			)
		);
	}

	return axiosResponse;
};

export const axiosInstanceProxy = (
	config: AxiosRequestConfig,
	selfHost = true
) => {
	if (!config.headers) {
		config.headers = {};
	}

	if (selfHost) {
		config.headers['X-Unauth-Error'] = 'Non-Redirect';
	}

	const instance = axios.create({
		...config
	});

	const termipassStore = useTermipassStore();

	instance.interceptors.request.use((config) => {
		if (termipassStore.totalStatus?.status == TermiPassStatus.OfflineMode) {
			return Promise.reject(new Error(InOfflineText()));
		}
		const application = getApplication();
		if (application.commonRequestIntercepts) {
			config = application.commonRequestIntercepts(config);
		}
		return config;
	});

	instance.interceptors.response.use(
		(response) => {
			const application = getApplication();
			if (
				application.commonResponseIntercepts &&
				application.commonResponseIntercepts(response)
			) {
				throw null;
			}
			return response;
		},
		(error: any) => {
			const application = getApplication();
			if (
				application.tokenInvalidErrorIntercep &&
				application.tokenInvalidErrorIntercep(error)
			) {
				return;
			}
			throw error;
		}
	);

	const instanceProxy = new Proxy(
		instance,
		axiosProxyHandler
	) as AxiosInstanceWithIntercept;

	instanceProxy.requestIntercepts = [];
	instanceProxy.responseIntercepts = [];

	instanceProxy.interceptRequest = (config) => {
		instanceProxy.requestIntercepts.push(config);
	};

	instanceProxy.interceptResponse = (config) => {
		instanceProxy.responseIntercepts.push(config);
	};

	return instanceProxy;
};

export const addAxiosProxyGlobalRequestInterceptor = (
	callback: RequestInterceptor
) => {
	interceptors.request.push(callback);
};

export const addAxiosProxyGlobalResponseInterceptor = (
	callback: ResponseInterceptor
) => {
	interceptors.response.push(callback);
};
