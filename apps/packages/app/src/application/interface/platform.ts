import type { TermiPassDeviceInfo } from '@bytetrade/core';
import { TabbarItem, HookCapacitorHttpPlugin } from './';
import { RouteLocationNormalizedLoaded } from 'vue-router';
import { QVueGlobals } from 'quasar';

export interface UserStorage {
	getItem(key: string): Promise<any>;
	setItem(key: string, value: unknown): Promise<void>;
	removeItem(key: string): Promise<void>;
}

/**
 * app App life cycle
 */
export interface AppPlatform {
	userStorage?: UserStorage;

	/******** app start *************/
	/**
	 * app.vue start
	 * @param data params
	 */
	appLoadPrepare(data: any): Promise<void>;

	/**
	 * app.vue mounted
	 */
	appMounted(): Promise<void>;

	/**
	 * app.vue unmounted
	 */
	appUnMounted(): Promise<void>;

	/**
	 * app.vue
	 * @param redirect
	 */
	appRedirectUrl(
		redirect: any,
		currentRoute: RouteLocationNormalizedLoaded
	): Promise<void>;

	/******** app home page (mobile) *************/

	tabbarItems: TabbarItem[];

	homeMounted(): Promise<void>;

	homeUnMounted(): Promise<void>;

	/**
	 * hook http request
	 */
	isHookHttpRequest: boolean;

	/**
	 * mobile http hook plugin
	 */
	hookCapacitorHttp: HookCapacitorHttpPlugin;

	/**
	 * server hook
	 */
	hookServerHttp: boolean;

	isMobile: boolean;

	isDesktop: boolean;

	isPad: boolean;

	isClient: boolean;

	getFirebaseToken(): Promise<string>;

	getTailscaleId(): Promise<string>;

	getDeviceInfo(): Promise<TermiPassDeviceInfo>;

	getQuasar(): QVueGlobals | undefined;

	isTabbarDisplay(): boolean;

	userAgent: string;

	socialKeys: {
		facebook: {
			appId: string;
			clientToken: string;
		};
		google: {
			webClientId: string;
			iOSClientId: string;
			androidClientId: string;
		};
	};

	vpnAbility: boolean;
}
