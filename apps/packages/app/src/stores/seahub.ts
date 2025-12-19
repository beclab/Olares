import { defineStore } from 'pinia';

export type DataState = {
	repo_id: string;
	repo_name: string;
	file_id: string;
};

export const useSeahubStore = defineStore('seahub', {
	state: () => {
		return {
			repo_id: '',
			repo_name: '',
			file_id: ''
		} as DataState;
	},

	getters: {},

	actions: {
		setRepoId({ id, name }) {
			this.repo_id = id;
			this.repo_name = name;
		}
	}
});
