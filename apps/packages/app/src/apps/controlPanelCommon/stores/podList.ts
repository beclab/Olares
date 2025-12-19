import { defineStore } from 'pinia';
interface State {
	data: any[];
	loading: boolean;
	websocketUpdatedRandom: undefined | number;
}
export const usePodList = defineStore('podList', {
	state: (): State => ({
		data: [],
		loading: true,
		websocketUpdatedRandom: undefined
	}),

	actions: {
		updateData(data: any, randomKey?: number) {
			this.data = data;
			if (randomKey) {
				this.websocketUpdatedRandom = randomKey;
			}
		},
		updateLoading(loading: boolean) {
			this.loading = loading;
		}
	}
});
