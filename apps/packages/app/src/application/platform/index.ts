import { AppPlatform } from '../interface/platform';
import {
	setPlatform as defaultSetPlatform,
	getPlatform as defaultGetPlatform
} from '@didvault/sdk/src/core';

import { NativeAppPlatform } from '../interface/extends/platform';

export const setAppPlatform = (p: AppPlatform) => {
	defaultSetPlatform(p);
};

export const getAppPlatform = () => {
	return defaultGetPlatform() as AppPlatform;
};

export const getNativeAppPlatform = () => {
	return defaultGetPlatform() as NativeAppPlatform;
};
