import { defineStore } from 'pinia';
import { api } from '@apps/dashboard/boot/axios';

interface OlaresInfoState {
	olaresd: string;
	terminusd: string;
}

const buildOlaresInfoUrl = (): string => {
	const parts = window.location.hostname.split('.');
	const host =
		parts.length > 1 ? parts.slice(1).join('.') : window.location.hostname;
	return `${window.location.protocol}//${host}/api/olares-info`;
};

export const useOlaresInfoStore = defineStore('dashboardOlaresInfo', {
	state: (): OlaresInfoState => ({
		olaresd: '0',
		terminusd: '0'
	}),

	getters: {
		isOlaresd(): boolean {
			return this.olaresd === '1' || this.terminusd === '1';
		}
	},

	actions: {
		async fetchOlaresInfo() {
			try {
				const res = await api.get(buildOlaresInfoUrl(), {});
				const info = res?.data?.data ?? {};
				this.olaresd = info.olaresd ?? '0';
				this.terminusd = info.terminusd ?? '0';
			} catch (error) {
				this.olaresd = '0';
				this.terminusd = '0';
			}
		}
	}
});
