import { defineStore } from 'pinia';

interface Resources {
	clusterResource: any;
	activeIndex: number;
	loading: boolean;
}
export const useResourcesStore = defineStore('Resources', {
	state: (): Resources => ({
		clusterResource: [],
		activeIndex: 0,
		loading: false
	}),

	getters: {
		currentData(state) {
			return state.clusterResource[state.activeIndex];
		}
	},
	actions: {}
});
