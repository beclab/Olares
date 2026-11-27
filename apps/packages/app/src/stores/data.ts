import { defineStore } from 'pinia';
import { useUserStore } from './user';
import { baseURL as fileBaseURL } from '../utils/constants';
import { useFilesStore } from './files';
import { useMenuStore } from './files-menu';
import { getApplication } from 'src/application/base';
import { FileViewMode } from 'src/utils/interface/files';

export type UpdatableUserFields = {
	id?: string;
	viewMode?: FileViewMode;
};

export type DataState = {
	user: {
		id: string;
		viewMode: FileViewMode;
		dateFormat: boolean;
		singleClick: boolean;
		perm: {
			admin: boolean;
			execute: boolean;
			create: boolean;
			rename: boolean;
			modify: boolean;
			delete: boolean;
			share: boolean;
			download: boolean;
		};
	};
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
		isDirty: boolean;
		fullSize: boolean;
		isShow: boolean;
	};
};

export const useDataStore = defineStore('data', {
	state: () => {
		return {
			user: {
				id: '',
				viewMode: 'list',
				dateFormat: false,
				singleClick: false,
				// Default to no permissions so early access (e.g. ListingFiles
				// hotkeys reading user.perm.rename/delete) before setUser() runs
				// can't throw on an undefined `perm`.
				perm: {
					admin: false,
					execute: false,
					create: false,
					rename: false,
					modify: false,
					delete: false,
					share: false,
					download: false
				}
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
				isDirty: false,
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
			this.user = value;
		},

		setJWT(value: any) {
			this.jwt = value;
		},

		updateUser(value: UpdatableUserFields) {
			if (typeof value !== 'object' || value === null) return;

			const updatableFields: (keyof UpdatableUserFields)[] = ['id', 'viewMode'];
			for (const field of updatableFields) {
				if (field in value) {
					Object.assign(this.user, { [field]: value[field] });
				}
			}
		},

		baseURL() {
			const user = useUserStore();
			if (
				getApplication().applicationName == 'larepass' &&
				process.env.NODE_ENV !== 'development'
			) {
				return user.getModuleSever('files');
			}
			return fileBaseURL;
		}
	}
});
