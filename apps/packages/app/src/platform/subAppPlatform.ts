import { AppPlatform } from '../application/interface/platform';
import { WebPlatform } from './platform';

import localStorage from 'localforage/src/localforage';
import { TabbarItem } from '../application/interface/index';

import { CapacitorHttp } from '@capacitor/core';
import { AppState } from '@didvault/sdk/src/core/app';
import { RouteLocationNormalizedLoaded } from 'vue-router';
import { QVueGlobals } from 'quasar';

export class SubAppPlatform extends WebPlatform implements AppPlatform {
	async appLoadPrepare(_data: any): Promise<void> {
		// commonAppLoadPrepare(this, data);
	}
	async appMounted(): Promise<void> {
		// commonAppMounted(this);
	}
	async appUnMounted(): Promise<void> {
		// commonUnMounted(this);
	}
	async appRedirectUrl(
		_redirect: any,
		_currentRoute: RouteLocationNormalizedLoaded
	): Promise<void> {
		// throw new Error('Method not implemented.');
	}
	tabbarItems = [] as TabbarItem[];

	async homeMounted(): Promise<void> {
		// throw new Error('Method not implemented.');
	}
	async homeUnMounted(): Promise<void> {
		// throw new Error('Method not implemented.');
	}

	userStorage = localStorage;

	isHookHttpRequest = false;

	hookServerHttp = true;

	get hookCapacitorHttp() {
		return CapacitorHttp;
	}

	isMobile = false;

	isDesktop = false;

	isPad = false;
	isClient = false;

	async getFirebaseToken() {
		return '';
	}

	async getTailscaleId() {
		return '';
	}

	getQuasar(): QVueGlobals | undefined {
		return undefined;
	}

	reconfigAppStateDefaultValue(_appState: AppState) {}

	isTabbarDisplay() {
		return true;
	}

	userAgent = navigator.userAgent;

	socialKeys = {
		facebook: {
			appId: '',
			clientToken: ''
		},
		google: {
			webClientId: '',
			iOSClientId: '',
			androidClientId: ''
		}
	};
}
