import { defineStore } from 'pinia';
import { useUserStore } from './user';
import { baseURL as fileBaseURL } from '../utils/constants';
import { useFilesStore } from './files';
import { useMenuStore } from './files-menu';

export type DataState = {
	user: any;
	req: any;
	jwt: string;
	progress: number;
	loading: boolean;
	reload: boolean;
	show: any;
	showShell: boolean;
	showPadPopup: boolean;
	showConfirm: any;
	preview: {
		isEditEnable: boolean;
		isEditing: boolean;
		isSaving: boolean;
		fullSize: boolean;
		isShow: boolean;
	};
};

export const useDataStore = defineStore('data', {
	state: () => {
		return {
			user: {
				viewMode: 'list'
			},
			req: {},
			jwt: '',
			progress: 0,
			loading: false,
			show: null,
			showShell: false,
			showPadPopup: false,
			showConfirm: null,
			preview: {
				isEditEnable: false,
				isEditing: false,
				isSaving: false,
				fullSize: false,
				isShow: false
			}
		} as DataState;
	},

	getters: {
		isLogged(state) {
			return state.user !== null;
		}
	},

	actions: {
		isFiles(route: any) {
			return (
				!this.loading &&
				(route.name === 'Files' ||
					route.name === 'Application' ||
					route.name === 'Seahub')
			);
		},

		closeHovers() {
			const fileStore = useFilesStore();
			if (this.showPadPopup) {
				fileStore.resetSelected();
			}
			this.show = null;
			this.showConfirm = null;
			this.showPadPopup = false;
			fileStore.shareRepoInfo = undefined;
		},

		toggleShell() {
			this.showShell = !this.showShell;
		},

		showHover(value: any) {
			if (typeof value !== 'object') {
				this.show = value;
				return;
			}
			this.show = value.prompt;
			this.showConfirm = value.confirm;
		},

		setUser(value: any) {
			if (value === null) {
				this.user = null;
				return;
			}
			this.user = value;
		},

		setJWT(value: any) {
			this.jwt = value;
		},

		updateUser(value: any) {
			if (typeof value !== 'object') return;
			for (const field in value) {
				if (field === 'id') {
					this.user.id = value[field];
				}
				if (field === 'viewMode') {
					this.user.viewMode = value[field];
				}
			}
		},

		baseURL() {
			const user = useUserStore();

			if (process.env.NODE_ENV === 'development') {
				return fileBaseURL;
			} else {
				if (
					process.env.PLATFORM == 'DESKTOP' ||
					process.env.PLATFORM == 'MOBILE'
				) {
					const baseURL = user.getModuleSever('files');
					return baseURL;
				} else {
					return fileBaseURL;
				}
			}
		}
	}
});
