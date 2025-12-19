import { defineStore } from 'pinia';
import { ShareResult } from 'src/utils/interface/share';
import localStorage from 'localforage/src/localforage';
import share from '../../api/files/v2/common/share';

export type RootState = {
	share: ShareResult | undefined;
	path_id: string | undefined;
	token: string | undefined;
	expiredInfo: {
		status: boolean;
		time: number | undefined;
	};
};

export const useShareStore = defineStore('share', {
	state: () => {
		return {
			share: undefined,
			path_id: undefined,
			token: undefined,
			expiredInfo: {
				status: false,
				time: undefined
			}
		} as RootState;
	},
	getters: {},
	actions: {
		setToken(token: string) {
			if (!this.path_id) {
				return;
			}
			localStorage.setItem(`share_token_${this.path_id}`, token);
			this.token = token;
		},
		async getToken() {
			if (!this.path_id) {
				return;
			}
			const token = await localStorage.getItem<string>(
				`share_token_${this.path_id}`
			);
			return token;
		},
		deleteToken() {
			if (!this.path_id) {
				return;
			}
			localStorage.removeItem(`share_token_${this.path_id}`);
			this.token = undefined;
		},
		async requestShareInfo() {
			if (!this.path_id || !this.token) {
				return;
			}
			const result = await share.getShare(this.path_id, this.token);
			if (result) {
				this.share = result;
			}
		}
	}
});
