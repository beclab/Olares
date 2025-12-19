import { fetchBlackList } from 'src/api/wise/other';
import { defineStore } from 'pinia';
import { parseBlacklistData } from 'src/utils/url2';

export type BlackState = {
	base_url: string;
	loading: boolean;
	initialized: boolean;
	parsedBlacklistRules: Array<{
		pattern: RegExp;
	}>;
};

let initPromise: Promise<void> | null = null;

export const useBlacklistStore = defineStore('black', {
	state: () => {
		return {
			base_url: '',
			loading: false,
			initialized: false,
			parsedBlacklistRules: []
		} as BlackState;
	},

	actions: {
		async init(base_url = '') {
			this.base_url = base_url;

			const loadRemoteBlacklist = async () => {
				this.loading = true;
				try {
					const data = await fetchBlackList(base_url);
					if (data) {
						this.parsedBlacklistRules = parseBlacklistData(data);
						console.log('parsedBlacklistRules---', this.parsedBlacklistRules);
					}
				} catch (e) {
					this.parsedBlacklistRules = [];
				} finally {
					this.loading = false;
					this.initialized = true;
				}
			};

			if (!initPromise) {
				initPromise = loadRemoteBlacklist();
			}

			if (typeof requestIdleCallback !== 'undefined') {
				requestIdleCallback(() => initPromise, { timeout: 2000 });
			} else {
				setTimeout(() => initPromise, 100);
			}
		},

		async ensureInitialized(): Promise<void> {
			if (this.initialized) {
				return;
			}

			if (initPromise) {
				await initPromise;
				return;
			}

			this.loading = true;
			try {
				const data = await fetchBlackList(this.base_url);
				if (data) {
					this.parsedBlacklistRules = parseBlacklistData(data);
				}
			} catch (e) {
				this.parsedBlacklistRules = [];
			} finally {
				this.loading = false;
				this.initialized = true;
			}
		}
	}
});
