import { defineStore } from 'pinia';
import axios from 'axios';
import { useTerminusStore } from '../terminus';
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
			const terminus = useTerminusStore();
			const tokenStore = useTokenStore();
			const data: any = await axios.get(`${tokenStore.url}/api/system/status`, {
				headers: {
					'X-Signature': terminus.olares_device_id
				}
			});

			if (data && data.code == 0 && data.data) {
				this.olaresInfo = data.data;
			}
			return data;
		},
		async collect_logs() {
			const terminus = useTerminusStore();
			const tokenStore = useTokenStore();
			const data: any = await axios.post(
				`${tokenStore.url}/api/command/collectLogs`,
				{
					systemd: {
						components: ['*'],
						since: '',
						maxLines: 0
					},
					host: {
						dmesg: true,
						network: true
					},
					cluster: {
						info: true
					},
					namespaces: {
						names: ['*']
					}
				}
			);
			return data;
		},
		async collectLogsByRunId(runID: string) {
			const tokenStore = useTokenStore();
			const data: any = await axios.get(
				`${tokenStore.url}/api/command/collectLogs/${runID}`
			);
			return data;
		},
		async getHostsList(): Promise<HostItem[]> {
			const terminus = useTerminusStore();
			const tokenStore = useTokenStore();
			const data: any = await axios.get(
				`${tokenStore.url}/api/system/hosts-file`,
				{
					headers: {
						'X-Signature': terminus.olares_device_id
					}
				}
			);
			return data;
		},
		async updateHostsList(items: HostItem[]) {
			const terminus = useTerminusStore();
			const tokenStore = useTokenStore();
			const data: any = await axios.post(
				`${tokenStore.url}/api/system/hosts-file`,
				{ items: items },
				{
					headers: {
						'X-Signature': terminus.olares_device_id
					}
				}
			);
			console.log(data);
			return data;
		},
		async getRegistryMirrors() {
			const terminus = useTerminusStore();
			const tokenStore = useTokenStore();
			const data: any = await axios.get(
				`${tokenStore.url}/api/system/hosts-file`,
				{
					headers: {
						'X-Signature': terminus.olares_device_id
					}
				}
			);
			return data;
		}
	}
});
