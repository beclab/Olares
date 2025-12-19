/* eslint-disable @typescript-eslint/no-unused-vars */
import {
	Err,
	ErrorCode,
	RequestProgress,
	base64ToString,
	marshal,
	unmarshal
} from '@didvault/sdk/src/core';
import { Sender, Request, Response } from '@didvault/sdk/src/core/transport';
import { axiosInstanceProxy } from './httpProxy';
import { useUserStore } from 'src/stores/user';
import { getAppPlatform } from '../application/platform';
import axios, { AxiosProgressEvent, InternalAxiosRequestConfig } from 'axios';
const httpHookList = ['createAttachment', 'getAttachment'];

export class PlatformAjaxSender implements Sender {
	constructor(public url: string) {}

	async send(req: Request, progress?: RequestProgress): Promise<Response> {
		const body = marshal(req.toRaw());

		const config = {
			headers: {
				'Content-Type': 'application/json',
				Accept: 'application/json'
			},
			responseType: 'blob',
			onUploadProgress: (progressEvent: AxiosProgressEvent) => {
				if (progress) {
					progress.uploadProgress = {
						total: progressEvent.total || 0,
						loaded: progressEvent.loaded
					};
				}
			},
			onDownloadProgress: (progressEvent: AxiosProgressEvent) => {
				if (progress) {
					progress.downloadProgress = {
						total: progressEvent.total || 0,
						loaded: progressEvent.loaded
					};
				}
			}
		} as any;
		let instance = axios.create(config);
		if (getAppPlatform().hookServerHttp && !httpHookList.includes(req.method)) {
			const instance1 = axiosInstanceProxy(config, false);
			instance1.interceptResponse((res) => {
				if (typeof res.data === 'string') {
					const jsonString = base64ToString(res.data);
					res.data = new Blob([jsonString], {
						type: 'application/json'
					});
				} else {
					res.data = new Blob([marshal(res.data)], {
						type: 'application/json'
					});
				}
				return res;
			});
			instance1.interceptRequest((config) => {
				return this.formatConfig(config);
			});
			instance = instance1;
		}

		instance.interceptors.request.use((config) => {
			return this.formatConfig(config);
		});
		try {
			const response = await instance.post(this.url, body);
			let res = await response.data.text();
			res = unmarshal(res);
			return new Response().fromRaw(res);
		} catch (e) {
			if (e.response) {
				if (
					e.response.status == 525 ||
					e.response.status == 530 ||
					e.response.status == 522 ||
					e.response.status > 1000
				) {
					throw new Err(ErrorCode.SERVER_NOT_EXIST, e.message, { error: e });
				}
				if (e.response.status == 401) {
					throw new Err(ErrorCode.SERVER_AUTH_FAILED, e.message, { error: e });
				}
				if (e.response.status == 400 || e.response.status == 459) {
					throw new Err(ErrorCode.TOKE_INVILID, e.message, { error: e });
				}
			}
			throw new Err(ErrorCode.SERVER_ERROR, e.message, { error: e });
		}
	}

	private formatConfig(config: InternalAxiosRequestConfig) {
		if (!config.headers) {
			config.headers = {
				'Content-Type': 'application/json',
				Accept: 'application/json'
			} as any;
		}

		if (config.headers) {
			const userStore = useUserStore();
			if (!userStore.current_id) {
				return config;
			}
			const user = userStore.users!.items.get(userStore.current_id);
			if (!user) {
				return config;
			}
			if (!user.access_token) {
				return config;
			}
			config.headers['X-Authorization'] = user.access_token;
		}
		return config;
	}
}
