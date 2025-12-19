import axios from 'axios';
import { defineStore } from 'pinia';
import { useTokenStore } from './token';
import { ApplicationInfo } from '../global';

export const useApplicationStore = defineStore('application', {
	state: () => ({
		applications: [] as ApplicationInfo[]
	}),

	getters: {},

	actions: {
		getApplicationById(name: string): ApplicationInfo | undefined {
			return this.applications.find((u) => u.name === name);
		},

		removeApplicationById(name: string) {
			const userIndex = this.applications.findIndex((u) => u.name === name);
			if (userIndex < 0) {
				return;
			}
			this.applications.splice(userIndex, 1);
		}
	}
});
