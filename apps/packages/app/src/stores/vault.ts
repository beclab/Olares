import { defineStore } from 'pinia';

export type DataState = {
	editing_item: any;
};

export const useVaultStore = defineStore('vault', {
	state: () => {
		return {
			editing_item: null
		} as DataState;
	},

	getters: {},

	actions: {}
});
