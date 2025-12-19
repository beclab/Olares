import { defineStore } from 'pinia';
import axios from 'axios';
import { Cookies } from 'quasar';
import { OlaresInfo, DefaultOlaresInfo } from '@bytetrade/core';
import { v4 as uuidv4 } from 'uuid';
import { Encoder, DeviceType } from '@bytetrade/core';

export interface DesktopConfig {
	bg: string;
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
			}
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
			} catch (e) {
				return e;
			}
		},

		setUrl(new_url: string | null) {
			this.url = new_url;
		},

		getAuthURL() {
			const name = this.olaresId.replace('@', '.');

			const url = 'https://auth.' + name;
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
