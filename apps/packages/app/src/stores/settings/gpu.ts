import { defineStore } from 'pinia';
import axios from 'axios';
import { useTokenStore } from './token';
import { VRAMMode } from '../../constant';
import { notifyFailed, notifySuccess } from 'src/utils/notifyRedefinedUtil';
import { i18n } from 'src/boot/i18n';

export interface GPUInfo {
	nodeName: string;
	id: string;
	count: number;
	devmem: number;
	devcore: number;
	type: string;
	mode: string;
	health: boolean;
	//   "sharemode": "1",
	sharemode: VRAMMode;
	apps?: [
		{
			appName: string;
			memory?: number;
		}
	];
	memoryAvailable?: number;
	memoryAllocated?: number;
	index?: number;
}

export const useGPUStore = defineStore('gpu', {
	state: () => ({
		managed_memory: false,
		gpuList: [] as GPUInfo[]
	}),

	getters: {},

	actions: {
		async getGpuList() {
			const tokenStore = useTokenStore();
			try {
				const list: any = await axios.get(`${tokenStore.url}/api/gpu/list`);
				if (list) {
					this.gpuList = list;
				}
			} catch (error) {
				console.log(error);
			}
		},
		async setGpuMode(mode: VRAMMode, id: string) {
			const tokenStore = useTokenStore();
			try {
				await axios.post(`${tokenStore.url}/api/gpu/mode`, {
					id,
					mode
				});
				notifySuccess(i18n.global.t('successful'));
				this.getGpuList();
			} catch (error) {
				console.log(error);
			}
		},
		async updateApplication(
			mode: VRAMMode,
			id: string,
			appName: string,
			memory?: number
		) {
			const tokenStore = useTokenStore();
			try {
				await axios.post(`${tokenStore.url}/api/gpu/update/app`, {
					id,
					mode,
					appName,
					memory
				});
				notifySuccess(i18n.global.t('successful'));
				await this.getGpuList();
			} catch (error) {
				console.log(error);
			}
		},
		async unbindApplication(mode: VRAMMode, id: string, appName: string) {
			const tokenStore = useTokenStore();
			try {
				await axios.post(`${tokenStore.url}/api/gpu/unassign/app`, {
					id,
					mode,
					appName
				});
				notifySuccess(i18n.global.t('successful'));
				await this.getGpuList();
			} catch (error) {
				console.log(error);
			}
		},
		async switchToVRAM(
			id: string,
			appName: string,
			unassign: { id: string }[] = [],
			memory?: string
		) {
			const tokenStore = useTokenStore();
			try {
				await axios.put(`${tokenStore.url}/api/gpu/assignments/bulk`, {
					appName,
					unassign,
					assign: [
						{
							id,
							memory
						}
					]
				});
				notifySuccess(i18n.global.t('successful'));
				await this.getGpuList();
			} catch (error) {
				console.log(error);
			}
		}
	}
});
