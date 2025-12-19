import { defineStore } from 'pinia';

export const useSplitMenu = defineStore('splitMenu', {
	state: () => ({
		type: ''
	}),
	getters: {
		doubleCount: (state) => state.type
	},
	actions: {
		changeStatus(type: string) {
			this.type = type;
		}
	}
});
