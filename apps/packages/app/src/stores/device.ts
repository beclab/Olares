import { defineStore } from 'pinia';
import { useUserStore } from './user';
import { useScaleStore } from './scale';
import { getAppPlatform } from '../application/platform';
import { DeviceInfo } from '@didvault/sdk/src/core';
import { app } from 'src/globals';
import { TermiPassDeviceInfo } from '@bytetrade/core';
import { ThemeDefinedMode } from '@bytetrade/ui';
import { ThemePlugin } from 'src/plugins/theme';
import { Dark } from 'quasar';
import { Platform } from 'quasar';
import { SupportLanguageType } from 'src/i18n';
import { busEmit } from 'src/utils/bus';
import { userModeSetItem } from './userStorageAction';
import { useChannelBexPost } from 'src/platform/interface/bex/front/interface';
import {
	languageStorageKey,
	themeStorageKey
} from 'src/utils/interface/device';

export type DeviceStoreState = {
	networkOnLine: boolean;
	isLandscape: boolean;
	theme: ThemeDefinedMode;
	isScaning: boolean;
	isInEditor: boolean;
	connectType: 'wifi' | 'cellular' | 'none';
	isMobile: boolean;
};

export const useDeviceStore = defineStore('device', {
	state: () => {
		return {
			networkOnLine: navigator.onLine,
			isLandscape: true,
			theme: ThemeDefinedMode.LIGHT,
			isInEditor: false,
			connectType: 'wifi',
			isScaning: false,
			isMobile: false
		} as DeviceStoreState;
	},

	getters: {},

	actions: {
		async init() {
			this.theme = (await ThemePlugin.get()).theme;
		},
		async getTermiPassInfo() {
			const userStore = useUserStore();

			const scaleStore = useScaleStore();

			const platform = getAppPlatform();

			const device_info: DeviceInfo = await platform.getDeviceInfo();

			const info = new TermiPassDeviceInfo(device_info);

			if (userStore.id) {
				info.id = userStore.id;
			}

			info.tailScaled = scaleStore.isOn;

			info.client_type = 'larePass';

			info.firebase_token = await platform.getFirebaseToken();

			info.tailScale_id = await platform.getTailscaleId();

			info.srp_id =
				app.authInfo && app.authInfo.sessions.length > 0
					? app.authInfo.sessions[0].id
					: '';

			return info;
		},

		compareTerminuPassInfo(
			oldValue: TermiPassDeviceInfo,
			newValue: TermiPassDeviceInfo
		) {
			return (
				// oldValue.termiPassID == newValue.termiPassID &&
				oldValue.platform == newValue.platform &&
				oldValue.osVersion == newValue.osVersion &&
				oldValue.id == newValue.id &&
				oldValue.appVersion == newValue.appVersion &&
				oldValue.vendorVersion == newValue.vendorVersion &&
				oldValue.userAgent == newValue.userAgent &&
				oldValue.locale == newValue.locale &&
				oldValue.manufacturer == newValue.manufacturer &&
				oldValue.model == newValue.model &&
				oldValue.browser == newValue.browser &&
				oldValue.browserVersion == newValue.browserVersion &&
				oldValue.description == newValue.description &&
				oldValue.runtime == newValue.runtime &&
				oldValue.tailScaled == newValue.tailScaled &&
				oldValue.tailScale_id == newValue.tailScale_id &&
				oldValue.sso == newValue.sso &&
				oldValue.srp_id == newValue.srp_id &&
				oldValue.createTime == newValue.createTime &&
				oldValue.lastSeenTime == newValue.lastSeenTime &&
				oldValue.lastIp == newValue.lastIp &&
				oldValue.client_type == newValue.client_type &&
				oldValue.firebase_token == newValue.firebase_token
			);
		},
		setTheme(theme: ThemeDefinedMode) {
			this.theme = theme;
			ThemePlugin.set({
				theme
			});
			this.updateTheme();
			useChannelBexPost<ThemeDefinedMode>(themeStorageKey, theme);
		},

		async updateTheme() {
			if (this.theme == ThemeDefinedMode.AUTO) {
				Dark.set('auto');
			} else {
				Dark.set(this.theme == ThemeDefinedMode.DARK);
			}
		},
		async setLanguage(locale: SupportLanguageType) {
			await userModeSetItem(languageStorageKey, locale);
			busEmit('LanguageUpdate', locale);
			useChannelBexPost<SupportLanguageType>(languageStorageKey, locale);
		},
		getUserAgent() {
			if (
				process.env.APPLICATION !== 'FILES' &&
				process.env.APPLICATION !== 'WISE'
			) {
				return getAppPlatform().userAgent;
			}
			return navigator.userAgent;
		},
		transferWifiEnable() {
			if (!Platform.is.nativeMobile) {
				return true;
			}
			const userStore = useUserStore();
			return (
				this.connectType == 'wifi' ||
				(this.connectType == 'cellular' && !userStore.transferOnlyWifi)
			);
		},
		transferEnable() {
			return this.networkOnLine && this.transferWifiEnable();
		}
	}
});
