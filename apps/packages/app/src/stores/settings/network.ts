import { defineStore } from 'pinia';
import { useTokenStore } from './token';
import { notifyFailed } from 'src/utils/settings/btNotify';
import axios from 'axios';
import { useBackgroundStore } from './background';
import { useAdminStore } from './admin';
import { OlaresTunneV2Interface } from 'src/utils/interface/frp';

export interface ReverseProxy {
	frp_server: string;
	frp_port: number;
	frp_auth_method: string;
	frp_auth_token: string;
	ip: string;
	enable_cloudflare_tunnel: boolean;
	enable_frp: boolean;
}

export interface OlaresTunnelInterface {
	name: string;
	host: string;
}

export type NetworkState = {
	reverseProxy?: ReverseProxy;
	olaresTunnels: OlaresTunnelInterface[];
	olaresTunnelsV2: OlaresTunneV2Interface[];
};

export const useNetworkStore = defineStore('network', {
	state: () =>
		({
			reverseProxy: undefined,
			olaresTunnels: [],
			olaresTunnelsV2: []
		} as NetworkState),

	getters: {},

	actions: {
		async configReverseProxy() {
			const tokenStore = useTokenStore();
			try {
				if (this.olaresTunnels.length == 0) {
					await this.getOlaresTunnelsV2();
				}
				const proxyData: any = await axios.get(
					`${tokenStore.url}/api/reverse-proxy`
				);
				this.reverseProxy = proxyData;
			} catch (error) {
				console.log(error);
			}
		},

		async updateReverseProxy(proxy: ReverseProxy) {
			const tokenStore = useTokenStore();
			await axios.post(`${tokenStore.url}/api/reverse-proxy`, proxy);
			return await this.configReverseProxy();
		},

		async getOlaresTunnels() {
			const tokenStore = useTokenStore();
			try {
				const olaresTunnels: any = await axios.get(
					`${tokenStore.url}/api/frp-servers`
				);
				this.olaresTunnels = olaresTunnels;
			} catch (error) {
				console.log(error);
			}
		},
		async getOlaresTunnelsV2() {
			const tokenStore = useTokenStore();
			const adminStore = useAdminStore();
			try {
				const olaresTunnels: any = await axios.post(
					`${tokenStore.url}/api/frp-servers-v2`,
					{ name: adminStore.olaresId }
				);
				this.olaresTunnelsV2 = olaresTunnels;
			} catch (error) {
				console.log(error);
			}
		},
		olaresTunnelsOptions() {
			return this.olaresTunnels.map((item: OlaresTunnelInterface) => ({
				label: item.name,
				value: item.host,
				enable: true
			}));
		},
		olaresTunnelsV2Options() {
			const backgroundStore = useBackgroundStore();
			return this.olaresTunnelsV2.map((item: OlaresTunneV2Interface) => {
				const label = item.name[backgroundStore.locale as string]
					? item.name[backgroundStore.locale as string]
					: item.name['en-US']
					? item.name['en-US']
					: '';
				return {
					label: label,
					value: item.machine.length > 0 ? item.machine[0].host : '',
					enable: true
				};
			});
		}
	}
});
