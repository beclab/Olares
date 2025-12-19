import { defineStore } from 'pinia';
import axios from 'axios';
import { useAdminStore } from './admin';
import { useTokenStore } from './token';
import { HostItem } from '../../constant';
import { TerminusStatus } from 'src/services/abstractions/mdns/service';

export type SpaceState = {
	olaresInfo: TerminusStatus | undefined;
	commandData: any;
};

export const useTerminusDStore = defineStore('terminusd', {
	state: () => {
		return {
			olaresInfo: undefined
		} as SpaceState;
	},

	actions: {
		async system_status() {
			const admin = useAdminStore();
			const tokenStore = useTokenStore();
			const data: any = await axios.get(`${tokenStore.url}/api/system/status`, {
				headers: {
					'X-Signature': admin.olares_device_id
				}
			});

			if (data && data.code == 0 && data.data) {
				this.olaresInfo = data.data;
			}
			return data;
		},
		async collect_logs() {
			const admin = useAdminStore();
			const tokenStore = useTokenStore();
			const data: any = await axios.post(
				`${tokenStore.url}/api/command/collectLogs`,
				null,
				{
					headers: {
						'X-Signature': admin.olares_device_id
					}
				}
			);
			return data;
		},
		async getHostsList(): Promise<HostItem[]> {
			const admin = useAdminStore();
			const tokenStore = useTokenStore();
			const data: any = await axios.get(
				`${tokenStore.url}/api/system/hosts-file`,
				{
					headers: {
						'X-Signature': admin.olares_device_id
					}
				}
			);
			return data;
		},
		async updateHostsList(items: HostItem[]) {
			const admin = useAdminStore();
			const tokenStore = useTokenStore();
			const data: any = await axios.post(
				`${tokenStore.url}/api/system/hosts-file`,
				{ items: items },
				{
					headers: {
						'X-Signature': admin.olares_device_id
					}
				}
			);
			console.log(data);
			return data;
		},
		async getRegistryMirrors() {
			const admin = useAdminStore();
			const tokenStore = useTokenStore();
			const data: any = await axios.get(
				`${tokenStore.url}/api/system/hosts-file`,
				{
					headers: {
						'X-Signature': admin.olares_device_id
					}
				}
			);
			return data;
		}
	}
});
