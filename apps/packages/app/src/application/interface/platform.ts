import { Platform } from '@didvault/sdk/src/core';

import { TabbarItem, HookCapacitorHttpPlugin } from './';
import { PlatformExtension } from '@didvault/sdk/src/core/PlatformExtension';
import { AppState } from '@didvault/sdk/src/core/app';
import { RouteLocationNormalizedLoaded } from 'vue-router';
import { QVueGlobals } from 'quasar';

/**
 * app App life cycle
 */
export interface AppPlatform extends Platform, PlatformExtension {
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

	reconfigAppStateDefaultValue(appState: AppState): void;

	getQuasar(): QVueGlobals | undefined;

	isTabbarDisplay(): boolean;

	userAgent: string;
}
