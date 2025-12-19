import { defineStore } from 'pinia';
import { getMiddlewareListAll } from '../network';
import { MiddlewareItem } from '@apps/control-panel-common/network/middleware';
import { Locker } from '@apps/dashboard/types/main';

interface State {
	list: MiddlewareItem[];
	loading: boolean;
	locker: Locker;
}

export const useMiddlewareStore = defineStore('middleware', {
	state: (): State => ({
		list: [],
		loading: false,
		locker: undefined
	}),
	getters: {},
	actions: {
		init() {
			this.getList(false);
			this.getList(true);
		},
		async getList(autofresh = false) {
			if (!autofresh) {
				this.loading = true;
			}
			try {
				const res = await getMiddlewareListAll();
				this.list = res.data.data || [];
				this.refresh();
			} catch (error) {
				this.loading = false;
				this.list = [];
			}
			this.loading = false;
		},
		refresh() {
			this.clearLocker();
			this.locker = setTimeout(() => {
				this.getList(true);
			}, 5000);
		},
		clearLocker() {
			this.locker && clearTimeout(this.locker);
		}
	}
});
