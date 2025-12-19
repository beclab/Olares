import axios from 'axios';
import { defineStore } from 'pinia';
import { useTokenStore } from './token';

export type DataState = {
	data: Record<string, string>;
};

export const useDataStore = defineStore('data', {
	state: () => {
		return {
			data: {}
		} as DataState;
	},

	getters: {
		//
	},

	actions: {
		async get(key: string) {
			const tokenStore = useTokenStore();
			const data: any = await axios.get(
				`${tokenStore.url}/api/datastore/` + key
			);
			this.data[key] = data;
			return data;
		},
		async put(key: string, body: any) {
			const tokenStore = useTokenStore();
			await axios.put(`${tokenStore.url}/api/datastore/` + key, body);

			this.data[key] = body;
		}
	}
});
