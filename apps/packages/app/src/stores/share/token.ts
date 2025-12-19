import axios from 'axios';
import { defineStore } from 'pinia';
import { OlaresInfo } from '@bytetrade/core';

export type RootState = {
	url: string | null;
	user: OlaresInfo;
};

export const useTokenStore = defineStore('token', {
	state: () => {
		return {
			url: null,
			user: {}
		} as RootState;
	},
	getters: {},
	actions: {
		async loadData() {
			const data: any = await axios.get(
				this.url + '/bfl/info/v1/olares-info',
				{}
			);
			this.user = data;
		},

		setUrl(new_url: string | null) {
			this.url = new_url;
		}
	}
});
