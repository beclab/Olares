import { defineStore } from 'pinia';

export const useMenuStore = defineStore('menu', {
	state: () => ({
		dialogShow: false,
		useSafeArea: true
	}),
	actions: {
		changeSafeArea(safeArea: boolean) {
			this.useSafeArea = safeArea;
		}
	}
});
