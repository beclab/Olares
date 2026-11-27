import {
	AxiosInstance,
	AxiosRequestConfig,
	AxiosResponse,
	InternalAxiosRequestConfig
} from 'axios';
import { BtNotify, NotifyDefinedType } from '@bytetrade/ui';
import { busEmit, NetworkErrorMode } from 'src/utils/bus';
import {
	InOfflineText,
	UserStatusActive
} from '../../utils/checkTerminusState';
import { getAppPlatform } from '../../application/platform';

import { useDataStore } from '../../stores/data';
import {
	TermiPassState,
	TermiPassStateInfo,
	useTermipassStore
} from 'src/stores/termipass';
import { Store } from 'pinia';
import { i18n } from 'src/boot/i18n';
import { getRequestErrorMessage } from 'src/utils/notifyRedefinedUtil';
import { useUserStore } from 'src/stores/user';

const defaultConfig: AxiosRequestConfig = {
	baseURL: '',
	timeout: 1800000,
	headers: {
		'Access-Control-Allow-Origin': '*',
		'Access-Control-Allow-Headers': 'X-Requested-With,Content-Type',
		'Access-Control-Allow-Methods': 'PUT,POST,GET,DELETE,OPTIONS,PATCH',
		'Content-Type': 'application/json',
		'X-Unauth-Error': 'Non-Redirect'
	}
};

const errorMessages = new Set();
let errorTimeout: ReturnType<typeof setTimeout> | undefined;

const handleError = (message: string) => {
	if (errorMessages.has(message)) {
		return;
	}
	errorMessages.add(message);
	console.error('handleError', message);

	if (message.split(' ').includes('403')) {
		BtNotify.show({
			type: NotifyDefinedType.FAILED,
			message: i18n.global.t('access_denied')
		});
	} else {
		BtNotify.show({
			type: NotifyDefinedType.FAILED,
			message
		});
	}

	if (errorTimeout) clearTimeout(errorTimeout);
	errorTimeout = setTimeout(() => {
		errorMessages.clear();
	}, 5000);
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

type TermipassStoreType = Store<
	'termipass',
	TermiPassState,
	{
		isLocal(): boolean;
		isP2P(): boolean;
		isDER(): boolean;
		isDirect(): boolean;
		totalStatus(): TermiPassStateInfo | undefined;
	}
>;

const assertRequestNotBlocked = (termipassStore: TermipassStoreType) => {
	if (
		getAppPlatform().isClient &&
		termipassStore.totalStatus?.isError !== UserStatusActive.active
	) {
		console.error('Request blocked');
		throw new Error('Request blocked');
	}
};

/**
 * Apply baseURL, blocking and the auth header to a request config.
 * Used by both axios native interceptor and the proxy interceptor.
 */
const applyRequestConfig = (
	config: InternalAxiosRequestConfig,
	termipassStore: TermipassStoreType
): InternalAxiosRequestConfig => {
	const dataStore = useDataStore();
	config.baseURL = dataStore.baseURL();

	assertRequestNotBlocked(termipassStore);

	if (!config.headers) {
		return config;
	}

	const userStore = useUserStore();
	if (!userStore.current_id) {
		return config;
	}
	const user = userStore.users!.items.get(userStore.current_id);
	if (!user || !user.access_token) {
		return config;
	}
	config.headers['X-Authorization'] = user.access_token;
	return config;
};

/**
 * Inspect a 2xx response and surface business-level errors to the user.
 * The backend returns `{ code, message }` where `0`/`200`/`300` are success.
 *
 * NOTE: loose equality is intentional here – the backend may return numeric
 * codes as strings (e.g. `"200"`); strict equality would treat those as
 * errors and trigger a false notification.
 */
const inspectBusinessError = (
	response: AxiosResponse,
	options: { notify?: boolean; throwError?: boolean } = {}
) => {
	const notify = options.notify !== false;
	const code = response?.data?.code;
	const responseMessage =
		response?.data?.message || response?.data?.msg || response?.data?.error;
	const isBusinessError =
		response &&
		response.status === 200 &&
		code !== undefined &&
		code !== null &&
		code !== '' &&
		code != 0 &&
		code != 200 &&
		code != 300 &&
		(responseMessage || options.throwError);
	if (isBusinessError) {
		const message =
			responseMessage || `Request failed with business code ${code}`;
		if (notify && responseMessage) {
			handleError(responseMessage);
		}
		if (options.throwError) {
			throw new ErrorResponse(response, message);
		}
	}
};

class Fetch {
	private instance!: AxiosInstance;

	private termipassStore!: TermipassStoreType;

	constructor() {
		this.init();
	}

	public async init(): Promise<void> {
		try {
			const { axiosInstanceProxy } = await import('../../platform/httpProxy');
			const instance = axiosInstanceProxy({ ...defaultConfig });

			const requestInterceptor = (config: InternalAxiosRequestConfig) => {
				try {
					return applyRequestConfig(config, this.termipassStore);
				} catch (error) {
					return Promise.reject(error);
				}
			};

			const proxyResponseInterceptor = (response: AxiosResponse) => {
				const isNoToast = response?.config?.noToast === true;
				inspectBusinessError(response, {
					notify: !isNoToast,
					throwError: isNoToast
				});
				if (isNoToast) {
					return response;
				}
				if (
					response &&
					response.status !== 200 &&
					response.status !== 201 &&
					response.request &&
					(response.request.method === 'put' ||
						response.request.method === 'PUT' ||
						response.request.method === 'post' ||
						response.request.method === 'POST')
				) {
					const responseError = new ErrorResponse(
						response,
						response.data?.message
							? `${response.data.message}`
							: `Request failed with status code ${response.status}`
					);
					// i18n-ignore-start
					if (responseError.message !== 'Request blocked') {
						handleError(getRequestErrorMessage(responseError));
					}
					// i18n-ignore-end
				}

				return response;
			};

			instance.interceptRequest(requestInterceptor);
			instance.interceptResponse(proxyResponseInterceptor);

			this.instance = instance;

			this.instance.interceptors.request.use(requestInterceptor, (error) =>
				Promise.reject(error)
			);

			this.instance.interceptors.response.use(
				(response) => {
					const isNoToast = response?.config?.noToast === true;
					inspectBusinessError(response, {
						notify: !isNoToast,
						throwError: isNoToast
					});
					return response;
				},
				(error) => {
					const isNoToast = error.response?.config?.noToast === true;
					if (isNoToast) {
						throw error;
					}

					const errorMessage = getRequestErrorMessage(error);
					// i18n-ignore-start
					if (error.message !== 'Request blocked') {
						handleError(errorMessage);
					}

					if (error.message === InOfflineText()) {
						throw error;
					}
					busEmit('network_error', {
						type: NetworkErrorMode.file,
						error: error.message
					});
					// i18n-ignore-end
					throw new Error(errorMessage);
				}
			);
		} catch (error) {
			console.error('Error fetching data:', error);
		}
	}

	private ensureStore() {
		if (!this.termipassStore) {
			this.termipassStore = useTermipassStore();
		}
	}

	public async get<T = any>(
		url: string,
		config?: AxiosRequestConfig
	): Promise<T> {
		this.ensureStore();
		const response: AxiosResponse<T> = await this.instance.get(url, config);
		return response.data;
	}

	async getStream(url: string, config?: RequestInit): Promise<Response> {
		this.ensureStore();
		assertRequestNotBlocked(this.termipassStore);
		const dataStore = useDataStore();
		const userStore = useUserStore();
		const baseURL = dataStore.baseURL();
		const fullUrl = url.startsWith('http')
			? url
			: new URL(url, baseURL).toString();
		const headers = new Headers(config?.headers || {});

		let credentials = config?.credentials;
		if (
			userStore.current_user?.access_token &&
			!headers.has('X-Authorization')
		) {
			headers.set('X-Authorization', userStore.current_user.access_token);
			credentials = credentials = config?.credentials || 'omit';
		}

		const response = await fetch(fullUrl, {
			...config,
			headers,
			credentials: credentials
		});

		return response;
	}

	async post<T = any>(
		url: string,
		data?: any,
		config?: AxiosRequestConfig
	): Promise<AxiosResponse<T>> {
		this.ensureStore();
		return this.instance.post(url, data, config);
	}

	async put<T = any>(
		url: string,
		data?: any,
		config?: AxiosRequestConfig
	): Promise<AxiosResponse<T>> {
		this.ensureStore();
		return this.instance.put(url, data, config);
	}

	async delete<T = any>(
		url: string,
		config?: AxiosRequestConfig
	): Promise<AxiosResponse<T>> {
		this.ensureStore();
		return this.instance.delete(url, config);
	}

	async patch<T = any>(
		url: string,
		data?: any,
		config?: AxiosRequestConfig
	): Promise<AxiosResponse<T>> {
		this.ensureStore();
		return this.instance.patch(url, data, config);
	}
}

const CommonFetch = new Fetch();

export { CommonFetch };
