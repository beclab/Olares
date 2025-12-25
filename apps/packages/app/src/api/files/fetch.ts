// import { axiosInstanceProxy } from '../platform/httpProxy';
import axios, { AxiosInstance, AxiosRequestConfig, AxiosResponse } from 'axios';
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
let errorTimeout: any;

const handleError = (message: string) => {
	if (!errorMessages.has(message)) {
		errorMessages.add(message);
		console.error('handleError', message);

		if (message.split(' ').find((code) => code == '403')) {
			BtNotify.show({
				type: NotifyDefinedType.FAILED,
				message: i18n.global.t('access_denied')
			});
		} else {
			BtNotify.show({
				type: NotifyDefinedType.FAILED,
				message: message
			});
		}

		if (errorTimeout) clearTimeout(errorTimeout);

		errorTimeout = setTimeout(() => {
			errorMessages.clear();
		}, 5000);
	}
};

class Fetch {
	private instance: AxiosInstance;

	// private instanceWeb: AxiosInstance;

	private termipassStore: Store<
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

	constructor() {
		this.init();
	}
	public async init(): Promise<void> {
		try {
			// Delayed reference axiosInstanceProxy
			const { axiosInstanceProxy } = await import('../../platform/httpProxy');
			const instance = axiosInstanceProxy({
				...defaultConfig
			});

			instance.interceptRequest((config) => {
				const dataStore = useDataStore();
				config.baseURL = dataStore.baseURL();

				return config;
			});

			this.instance = instance;

			this.instance.interceptors.request.use(
				(config) => {
					const dataStore = useDataStore();
					config.baseURL = dataStore.baseURL();

					if (
						getAppPlatform().isClient &&
						this.termipassStore.totalStatus?.isError !== UserStatusActive.active
					) {
						console.error('Request blocked');
						return Promise.reject(new Error('Request blocked'));
					}

					return config;
				},
				(error) => {
					return Promise.reject(error);
				}
			);

			this.instance.interceptors.response.use(
				(response) => {
					if (
						response &&
						response.status == 200 &&
						response.data.code &&
						response.data.code != 0 &&
						response.data.code != 200 &&
						response.data.code != 300 &&
						response.data.message
					) {
						if (response.data.message) {
							handleError(response.data.message);
						}
					}
					return response;
				},
				(error) => {
					const errorMessage = getRequestErrorMessage(error);
					if (error.message !== 'Request blocked') {
						handleError(errorMessage);
					}

					if (error.message == InOfflineText()) {
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

	public async get<T = any>(
		url: string,
		config?: AxiosRequestConfig
	): Promise<T> {
		this.termipassStore = useTermipassStore();
		const response: AxiosResponse<T> = await this.instance.get(url, config);
		return response.data;
	}

	async post<T = any>(
		url: string,
		data?: any,
		config?: AxiosRequestConfig
	): Promise<AxiosResponse<T>> {
		this.termipassStore = useTermipassStore();
		return this.instance.post(url, data, config);
	}

	async put<T = any>(
		url: string,
		data?: any,
		config?: AxiosRequestConfig
	): Promise<AxiosResponse<T>> {
		return this.instance.put(url, data, config);
	}

	async delete<T = any>(
		url: string,
		config?: AxiosRequestConfig
	): Promise<AxiosResponse<T>> {
		return this.instance.delete(url, config);
	}

	async patch<T = any>(
		url: string,
		data?: any,
		config?: AxiosRequestConfig
	): Promise<AxiosResponse<T>> {
		return this.instance.patch(url, data, config);
	}
}

const CommonFetch = new Fetch();

export { CommonFetch };
