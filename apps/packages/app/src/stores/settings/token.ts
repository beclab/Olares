import { defineStore } from 'pinia';
import { Token } from 'src/constant/constants';

export type RootState = {
	url: string | null;
	token: Token | null;
};

export const useTokenStore = defineStore('token', {
	state: () => {
		return {
			url: null,
			token: null
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
