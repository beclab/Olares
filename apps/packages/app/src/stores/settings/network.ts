import { defineStore } from 'pinia';
import { useTokenStore } from './token';
import { notifyFailed } from 'src/utils/settings/btNotify';
import axios from 'axios';
import { useBackgroundStore } from './background';
import { useAdminStore } from './admin';
import { OlaresTunneV2Interface } from 'src/utils/interface/frp';
import { useTerminusDStore } from './terminusd';

export interface ReverseProxy {
	mode: 'none' | 'public-ip' | 'frp';
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
	proxyStatus?: { phase: string; message?: string };
	hasProxyToken: boolean;
	defaultDomain?: {
		domain: string;
		username?: string;
		status: { phase: string; message?: string };
	};
	olaresTunnels: OlaresTunnelInterface[];
	olaresTunnelsV2: OlaresTunneV2Interface[];
};

export const useNetworkStore = defineStore('network', {
	state: () =>
		({
			reverseProxy: undefined,
			hasProxyToken: false,
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
					`${tokenStore.url}/api/network/reverse-proxy`
				);
				this.reverseProxy = {
					...proxyData.config,
					frp_server: proxyData.config.frp_server || '',
					frp_port: proxyData.config.frp_port || 7000,
					frp_auth_method: proxyData.config.frp_auth_method || 'jws',
					frp_auth_token: '',
					ip: proxyData.entrance?.public_ip || '',
					enable_frp: proxyData.config.mode === 'frp',
					enable_cloudflare_tunnel: false
				};
				this.proxyStatus = proxyData.status;
				this.hasProxyToken = proxyData.has_auth_token;
			} catch (error) {
				console.log(error);
			}
		},

		async updateReverseProxy(proxy: ReverseProxy) {
			const tokenStore = useTokenStore();
			await axios.post(`${tokenStore.url}/api/network/reverse-proxy`, {
				mode: proxy.mode,
				...(proxy.mode === 'frp'
					? {
							frp_server: proxy.frp_server,
							frp_port: proxy.frp_port,
							frp_auth_method: proxy.frp_auth_method || 'jws',
							frp_auth_token: proxy.frp_auth_token
					  }
					: {})
			});
			return await this.configReverseProxy();
		},

		async getDefaultDomain() {
			const tokenStore = useTokenStore();
			this.defaultDomain = (await axios.get(
				`${tokenStore.url}/api/network/domain`
			)) as any;
			return this.defaultDomain;
		},
		async bindDefaultDomain(domain: string) {
			const tokenStore = useTokenStore();
			this.defaultDomain = (await axios.post(
				`${tokenStore.url}/api/network/domain`,
				{ domain }
			)) as any;
			return this.defaultDomain;
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
			const terminusDStore = useTerminusDStore();
			try {
				if (!terminusDStore.olaresInfo) {
					await terminusDStore.system_status();
				}
				const olaresTunnels: any = await axios.post(
					`${tokenStore.url}/api/frp-servers-v2`,
					{
						name: adminStore.olaresId,
						device_name: terminusDStore.olaresInfo?.device_name
					}
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
