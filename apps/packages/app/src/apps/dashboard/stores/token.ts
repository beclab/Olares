import { defineStore } from 'pinia';
import { Token } from '../global';
import axios from 'axios';
//const domainPing = require("domain-ping");
//import Ping from 'ping.js';

export type RootState = {
	token: Token | null;
	url: string | null;
};

export const useTokenStore = defineStore('token', {
	state: () => {
		return {
			token: null,
			url: null
		} as RootState;
	},
	getters: {
		// getToken: (state) : (string|null) => {
		//   return state.token;
		// }
	},
	actions: {
		loadData() {
			const res = localStorage.getItem('token');
			if (res) {
				this.token = JSON.parse(res);
			}
		},

		setUrl(new_url: string | null) {
			this.url = new_url;
			if (new_url) {
				localStorage.setItem('url', new_url);
			}
		}
	}
});
