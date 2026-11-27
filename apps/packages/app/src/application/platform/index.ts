import { TermiPassDeviceInfo } from '@bytetrade/core';
import type { AppPlatform } from '../interface/platform';
import type { NativeAppPlatform } from '../interface/extends/platform';

const browserPlatform = {
	isClient: false,
	vpnAbility: false,
	getDeviceInfo: async () => new TermiPassDeviceInfo({ description: '' }),
	getQuasar: () => undefined
};

let registeredPlatform: AppPlatform | undefined;

export const setAppPlatform = (platform: AppPlatform) => {
	registeredPlatform = platform;
};

export const getAppPlatform = () => {
	return registeredPlatform ?? (browserPlatform as AppPlatform);
};

export const getNativeAppPlatform = () => {
	return getAppPlatform() as NativeAppPlatform;
};
