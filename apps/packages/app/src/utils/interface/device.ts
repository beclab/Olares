import { isChromeExtension } from '../bex/link';

export const languageStorageKey = 'locale';
export const userCurrentIdStorageKey = 'current_id';
export const themeStorageKey = 'theme';
export const appInitKey = 'app_init';

export enum TARGET_ORIGIN {
	OPTION_PAGE = 'OPTION_PAGE',
	SIDE_PANEL = 'SIDE_PANEL'
}

export const getCurrentOrigin = () =>
	isChromeExtension() ? TARGET_ORIGIN.OPTION_PAGE : TARGET_ORIGIN.SIDE_PANEL;
