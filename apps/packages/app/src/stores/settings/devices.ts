import { defineStore } from 'pinia';
import axios from 'axios';
import { TermiPassDeviceInfo, SSOToken } from '@bytetrade/core';
import { useTokenStore } from './token';

export type DevicesState = {
	devices: TermiPassDeviceInfo[];
	thisDevice: TermiPassDeviceInfo | null;
};

export interface SSOResult {
	sso: SSOToken;
	termiPass: TermiPassDeviceInfo;
}

export const useDevicesStore = defineStore('settingsDevices', {
	state: () => {
		return {
			devices: [],
			thisDevice: null
		} as DevicesState;
	},

	actions: {
		async get_vault_devices(): Promise<any> {
			const tokenStore = useTokenStore();
			const data: any = await axios.get(tokenStore.url + '/api/device/vault');
			return data;
		},
		async get_sso(): Promise<SSOResult[]> {
			const tokenStore = useTokenStore();
			const data: SSOResult[] = await axios.get(
				tokenStore.url + '/api/device/sso'
			);
			return data;
		},
		async revoke_token(token: SSOResult) {
			const tokenStore = useTokenStore();
			await axios.delete(
				`${tokenStore.url}/api/device/sso/${token.termiPass.sso}`
			);
		}
	}
});
