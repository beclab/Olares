import { defineStore } from 'pinia';
import { ClientSchema } from '../globals';
import {
	GetResponseResponse,
	PresentationDefinition,
	ResolutionResult,
	Submission,
	GolbalHost
} from '@bytetrade/core';
import { axiosInstanceProxy } from '../platform/httpProxy';
// import { AxiosInstance } from 'axios';
import { useUserStore } from './user';
import { useDeviceStore } from './device';
import { getApplication } from 'src/application/base';

export type SSIState = {
	// did_url: string | undefined;
	// vc_url: string | undefined;
};

export const useSSIStore = defineStore('did', {
	state: () => {
		return {} as SSIState;
	},
	getters: {},
	actions: {
		async pre_did_register(jws: string) {
			const data = await this.didInstance().post('/create_local', {
				jws
			});

			if (!data.data || data.data.code != 0) {
				throw new Error(
					data.data.message ? data.data.message : 'create local faild'
				);
			}
		},
		async get_name_by_did(did: string): Promise<string | null | undefined> {
			try {
				const get_did_response = await this.didInstance().get(
					'/get_name_by_did/' + did
				);

				if (get_did_response && get_did_response.status == 200) {
					if (get_did_response.data.code == 0) {
						return get_did_response.data.data;
					} else {
						return null;
					}
				}
				return null;
			} catch (err) {
				return undefined;
			}
		},
		async resolve_did(did: string): Promise<ResolutionResult | null> {
			try {
				const get_did_response = await this.didInstance().get(
					'/1.0/identifiers/' + did
				);

				if (get_did_response && get_did_response.status == 200) {
					return get_did_response.data;
				}
				return null;
			} catch (err) {
				return null;
			}
		},
		async resolve_name(name: string): Promise<ResolutionResult | null> {
			try {
				const get_name_response = await this.didInstance().get(
					'/1.0/name/' + name.replace('@', '.')
				);
				if (get_name_response && get_name_response.status == 200) {
					return get_name_response.data;
				}
				return null;
			} catch (err) {
				return null;
			}
		},
		async get_application_schema(
			type: string
		): Promise<ClientSchema | undefined> {
			const response: any = await this.vcInstance().get(
				'/get_application_schema/' + type
			);
			if (
				(response.status != 200 && response.status != 201) ||
				response.data.code != 0
			) {
				return undefined;
			}
			return response.data.data;
		},
		async get_application_response(
			id: string
		): Promise<GetResponseResponse | undefined> {
			const response: any = await this.vcInstance().get(
				'/get_application_response/' + id
			);
			if (
				(response.status != 200 && response.status != 201) ||
				response.data.code != 0
			) {
				return undefined;
			}
			return response.data.data;
		},
		async get_presentation_definition(
			type: string
		): Promise<PresentationDefinition | undefined> {
			const response = await this.vcInstance().get(
				'/get_presentation_definition/' + type
			);
			if (response.status != 200 || response.data.code != 0) {
				return undefined;
			}
			return response.data.data;
		},
		async submit_presentation(
			jws: string,
			address: string,
			domain: string | null
		): Promise<Submission> {
			const obj: any = {
				jws,
				address
			};
			if (domain) {
				obj.domain = domain;
			}

			const response: any = await this.vcInstance().post(
				'/submit_presentation',
				obj
			);

			if (
				(response.status != 201 && response.status != 200) ||
				response.data.code != 0
			) {
				throw Error(response.data.message || 'Submit Presentation Failure');
			}
			return response.data.data;
		},

		getDidUrl() {
			if (getApplication() && getApplication().ssiRule) {
				const application = getApplication();
				const rule = application.ssiRule!();
				return GolbalHost.DID_GATE_URL[rule] || GolbalHost.DID_GATE_URL.en;
			}
			const userStore = useUserStore();
			if (userStore.defaultDomain == 'cn') {
				return GolbalHost.DID_GATE_URL.cn;
			}
			return GolbalHost.DID_GATE_URL.en;
		},
		didInstance() {
			const instance = axiosInstanceProxy(
				{
					baseURL: this.getDidUrl(),
					timeout: 10000,
					headers: {
						'Content-Type': 'application/json'
					}
				},
				false
			);

			instance.interceptRequest((config) => {
				if (!config.headers) {
					config.headers = {
						'Content-Type': 'application/json',
						Accept: 'application/json'
					} as any;
				}
				const deviceStore = useDeviceStore();
				config.headers['User-Agent'] = deviceStore.getUserAgent();
				return config;
			});

			return instance;
		},
		getVCUrl() {
			if (getApplication() && getApplication().ssiRule) {
				const application = getApplication();
				const rule = application.ssiRule!();
				return GolbalHost.VC[rule] || GolbalHost.VC.en;
			}
			const userStore = useUserStore();
			if (userStore.defaultDomain == 'cn') {
				return GolbalHost.VC.cn;
			}
			return GolbalHost.VC.en;
		},
		vcInstance() {
			const instance = axiosInstanceProxy(
				{
					baseURL: this.getVCUrl(),
					timeout: 1000 * 300,
					headers: {
						'Content-Type': 'application/json'
					}
				},
				false
			);

			instance.interceptRequest((config) => {
				if (!config.headers) {
					config.headers = {
						'Content-Type': 'application/json',
						Accept: 'application/json'
					} as any;
				}
				const deviceStore = useDeviceStore();
				config.headers['User-Agent'] = deviceStore.getUserAgent();
				return config;
			});

			return instance;
		}
	}
});
