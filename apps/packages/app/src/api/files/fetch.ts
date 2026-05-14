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

/**
 * Apply baseURL, blocking and X-Authorization to a request config.
 * Used by both axios native interceptor and the proxy interceptor.
 */
const applyRequestConfig = (
	config: InternalAxiosRequestConfig,
	termipassStore: TermipassStoreType
): InternalAxiosRequestConfig => {
	const dataStore = useDataStore();
	config.baseURL = dataStore.baseURL();

	if (
		getAppPlatform().isClient &&
		termipassStore.totalStatus?.isError !== UserStatusActive.active
	) {
		console.error('Request blocked');
		throw new Error('Request blocked');
	}

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
const inspectBusinessError = (response: AxiosResponse) => {
	if (
		response &&
		response.status === 200 &&
		response.data?.code &&
		response.data.code != 0 &&
		response.data.code != 200 &&
		response.data.code != 300 &&
		response.data?.message
	) {
		handleError(response.data.message);
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
				inspectBusinessError(response);
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
					if (responseError.message !== 'Request blocked') {
						handleError(getRequestErrorMessage(responseError));
					}
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
					inspectBusinessError(response);
					return response;
				},
				(error) => {
					const errorMessage = getRequestErrorMessage(error);
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
