import { get } from 'lodash';
import { defineStore } from 'pinia';
import axios from 'axios';
import {
	ApplicationInfo,
	OpenApplicationConfig
} from '@apps/studio/src/types/core';

export type DevelopingApps = {
	url: string;
	apps: ApplicationInfo[];
	current_app: ApplicationInfo | undefined;
};

export const useDevelopingApps = defineStore('studio-app', {
	state() {
		return {
			url: '',
			apps: [],
			current_app: undefined
		} as DevelopingApps;
	},
	getters: {
		isDeployed: (state) =>
			state.current_app
				? get(state.current_app, 'state', '') === 'deployed'
				: false
	},
	actions: {
		setUrl(new_url: string) {
			this.url = new_url;
			if (new_url) {
				localStorage.setItem('url', new_url);
			}
		},
		async getApps() {
			try {
				this.apps = await axios.get(this.url + '/api/command/list-app');
			} catch (e) {
				console.log(e);
			}
		},

		async openApplication(config: OpenApplicationConfig) {
			try {
				const data: any = await axios.post(
					this.url + '/api/command/open-application',
					config
				);
				return data.appId;
			} catch (e) {
				console.log(e);
			}
		},

		setHideErrorFooter(appId: number, hide: boolean) {
			const app = this.apps.find((app) => app.id === appId);
			if (app) {
				app.hideErrorFooter = hide;
			}
			if (this.current_app && this.current_app.id === appId) {
				this.current_app.hideErrorFooter = hide;
			}
		}
	}
});
