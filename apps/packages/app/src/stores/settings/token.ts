import { defineStore } from 'pinia';

export type RootState = {
	url: string | null;
};

export const useTokenStore = defineStore('token', {
	state: () => {
		return {
			url: null
		} as RootState;
	},
	getters: {},
	actions: {
		setUrl(new_url: string | null) {
			this.url = new_url;
			if (new_url) {
				localStorage.setItem('url', new_url);
			}
		}
	}
});
