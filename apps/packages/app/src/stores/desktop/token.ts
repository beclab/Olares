import { defineStore } from 'pinia';
import axios from 'axios';
import { Cookies } from 'quasar';
import { OlaresInfo, DefaultOlaresInfo } from '@bytetrade/core';
import { v4 as uuidv4 } from 'uuid';
import { Encoder, DeviceType } from '@bytetrade/core';
import { AccountInfo } from 'src/constant/global';
import { OLARES_ROLE } from 'src/constant';
import { useWidgetPreferencesStore } from 'src/stores/settings/widgetPreferences';
import { InitLayout } from 'src/types/desktop-layout';

export interface DesktopConfig {
	bg: string;
	style: string;
}

export const TERMINUS_ID = 'terminusId';

export type RootState = {
	url: string | null;
	config: DesktopConfig;
	terminus: OlaresInfo;
	// id: string | null;
	deviceInfo: {
		device: DeviceType;
		isVerticalScreen: boolean;
	};
	users: AccountInfo[];
	// Desktop boot payload folded into /server/init: the raw myApps list the
	// client resolves layout keys against, and the bundled launch/dock layout
	// envelopes. Consumed once by the boot path so it skips the separate
	// /server/myApps + launch-layout + dock-layout requests.
	initMyApps: any[];
	initLayout: InitLayout | null;
	// In-flight (or resolved) loadData promise, so the boot path can await the
	// same /server/init round-trip appLoadPrepare already kicked off instead of
	// racing it.
	loadDataPromise: Promise<void> | null;
};

export const useTokenStore = defineStore('token', {
	state: () => {
		return {
			url: '',
			config: {},
			terminus: DefaultOlaresInfo,
			// id: null,
			deviceInfo: {
				device: DeviceType.DESKTOP,
				isVerticalScreen: false
			},
			users: [] as AccountInfo[],
			initMyApps: [] as any[],
			initLayout: null,
			loadDataPromise: null
		} as RootState;
	},
	getters: {
		olaresId(): string {
			return this.terminus.olaresId || this.terminus.terminusName;
		},
		olares_device_id(): string {
			return this.terminus.id || this.terminus.terminusId;
		},
		islocal(): boolean {
			return window.location.hostname.endsWith('olares.local');
		},
		isAdmin(): boolean {
			if (this.users.length == 0) {
				return false;
			}
			const user = this.users.find((e) => e.terminusName == this.olaresId);
			if (!user || user.roles.length == 0) {
				return false;
			}
			return (
				user.roles[0] == OLARES_ROLE.ADMIN || user.roles[0] == OLARES_ROLE.OWNER
			);
		},
		activeUsers(): AccountInfo[] {
			return this.users.filter((account) => account.wizard_complete);
		}
	},
	actions: {
		async loadData() {
			// this.id = localStorage.getItem('desktop-id');
			// if (!this.id) {
			// 	this.id = uuidv4();
			// 	localStorage.setItem('desktop-id', this.id || '');
			// }
			const data: any = await axios.get(this.url + '/server/init', {});
			this.terminus = data.terminus;
			this.config = data.config;
			this.users = data.users || [];
			// Stash the folded desktop boot payload for the boot path to consume.
			this.initMyApps = Array.isArray(data.apps) ? data.apps : [];
			this.initLayout = data.layout || null;
			const widgetStore = useWidgetPreferencesStore();
			widgetStore.save(data.config.widget);
			// this.loadUsers();
		},

		// Memoized loadData so appLoadPrepare and the boot initializeApp share a
		// single /server/init round-trip instead of racing two of them.
		async ensureLoaded(): Promise<void> {
			if (!this.loadDataPromise) {
				this.loadDataPromise = this.loadData();
			}
			return this.loadDataPromise;
		},

		async loadUsers() {
			const data: any = await axios.get(this.url + '/api/users', {});
			this.users = data;
		},

		async updateDesktopConfig(config: any) {
			const data: any = await axios.post(
				this.url + '/server/updateConfig',
				config
			);

			this.config = data;
		},

		async logout() {
			try {
				await axios.post(this.url + '/api/logout');
				Cookies.remove('auth_token');
				// Drop the shared desktop-layout caches so the next account
				// doesn't inherit this one's launchpad/dock order.
				try {
					localStorage.removeItem('launch_layout_cache');
					localStorage.removeItem('dock_layout_cache');
				} catch (_e) {
					/* ignore quota / private mode */
				}
			} catch (e) {
				return e;
			}
		},

		setUrl(new_url: string | null) {
			this.url = new_url;
		},

		getAuthURL() {
			let url = window.location.protocol + '//' + 'auth.';
			if (this.islocal) {
				url = url + this.olaresId.split('@')[0] + '.olares.local';
			} else {
				const name = this.olaresId.replace('@', '.');
				url = url + name;
			}
			return url;
		},

		getAppLocalUrl(id: string) {
			return id + '.' + this.olaresId.split('@')[0] + '.' + 'olares.local';
		},

		async validateTerminusInfo(
			customValidator: (currentId: string, lastId: string) => boolean = (
				currentId,
				lastId
			) => currentId === (lastId ?? ''),
			onSuccess = () => {},
			onFailure = () => {}
		) {
			if (!this.terminus || !this.olares_device_id) {
				await onFailure();
				return;
			}

			const currentId = this.olares_device_id;
			const lastId = localStorage.getItem(TERMINUS_ID) ?? '';

			const isValid = customValidator(currentId, lastId);

			if (isValid) {
				if (!lastId) {
					localStorage.setItem(TERMINUS_ID, currentId);
				}
				await onSuccess();
			} else {
				localStorage.setItem(TERMINUS_ID, currentId);
				await onFailure();
			}
		}
	}
});
