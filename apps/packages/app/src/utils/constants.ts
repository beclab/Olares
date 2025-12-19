interface SettingType {
	icon: string;
	mode: number;
	name: string;
}

export const SETTING_MENU: Record<string, SettingType> = {
	account: {
		icon: 'person',
		mode: 1,
		name: 'web.account.title'
	},
	security: {
		icon: 'security',
		mode: 2,
		name: 'web.security.title'
	},
	display: {
		icon: 'desktop_windows',
		mode: 3,
		name: 'web.display.title'
	},
	tools: {
		icon: 'construction',
		mode: 4,
		name: 'web.tools.title'
	},
	rebinding: {
		icon: 'person',
		mode: 5,
		name: 'web.rebinding.title'
	}
};

export const PASSWORD_RULE = {
	LENGTH_RULE: '^.{8,32}$',
	LOWERCASE_RULE: '^(?=.*[a-z])',
	UPPERCASE_RULE: '^(?=.*[A-Z])',
	DIGIT_RULE: '^(?=.*[0-9])',
	SYMBOL_RULE: '^(?=.*[@$!%*?&_.])[A-Za-z0-9@$!%*?&_.]+$',
	ALL_RULE:
		'^(?=.*[@$!%*?&_.])(?=.*[A-Z])(?=.*[a-z])(?=.*[0-9])[A-Za-z0-9@$!%*?&_.]{8,32}$'
};

export const SSH_PASSWORD_RULE = {
	LENGTH_RULE: '^.{10,32}$'
};

export enum TERMINUS_VC_TYPE {
	GOOGLE = 'Google',
	TWITTER = 'Twitter',
	FACEBOOK = 'Facebook',
	CHANNEL = 'Channel'
}

export enum ConfirmButtonStatus {
	normal = 1,
	error = 2,
	disable = 3
}

export interface LayoutItem {
	name: string;
	icon: string;
	icon_active: string;
	identify: LayoutMenuIdetify;
	path: string;
	icon_active_dark: string;
	short_name: string;
}

export enum LayoutMenuIdetify {
	FILES = 'files',
	VAULT = 'vault',
	TRANSMISSION = 'transmission',
	SYSTEM_SETTINGS = 'system_settings',
	ACCOUNT_CENTER = 'account_center'
}

export const LayoutMenu: LayoutItem[] = [
	{
		name: 'files.files',
		short_name: 'files.files',
		icon: 'files',
		icon_active: 'files-active',
		identify: LayoutMenuIdetify.FILES,
		path: '/Files/Home/',
		icon_active_dark: 'files-active-dark'
	},

	{
		name: 'Vault',
		short_name: 'Vault',
		icon: 'vault',
		icon_active: 'vault-active',
		identify: LayoutMenuIdetify.VAULT,
		path: '/items',
		icon_active_dark: 'vault-active'
	},
	{
		name: 'transmission.title',
		short_name: 'transmission.short_title',
		icon: 'transfer',
		icon_active: 'transfer-active',
		identify: LayoutMenuIdetify.TRANSMISSION,
		path: '/transmission',
		icon_active_dark: 'transfer-active'
	}
];

const name = 'Files';
const disableExternal = false;
const baseURL = window.location.origin;
const staticURL = '';
const signup = false;
const version = '0.1';
const logoURL = staticURL + '/img/logo.svg';
const noAuth = true;
const authMethod = 'json';
const loginPage = false;
const theme = 'light';
const enableThumbs = true;
const resizePreview = true;
const enableExec = false;
const origin = window.location.origin;
const fileList = [
	'avi',
	'doc',
	'docx',
	'gif',
	'png',
	'jpeg',
	'jpg',
	'mp3',
	'mp4',
	'pdf',
	'txt',
	'xls',
	'xlsx',
	'webp'
];

export enum SortTpe {
	name = 'name',
	size = 'size',
	type = 'type',
	modified = 'modify'
}

export enum CurrentView {
	FIRST_FACTOR = 'FirstFactor',
	SECOND_FACTOR = 'SecondFactor',
	MOBILE_VERIFICATION = 'MobileVerification'
}

export const FloatBg: Record<number, { color1: string; color2: string }> = {
	1: {
		color1: '#93e1ff',
		color2: '#faffd9'
	},
	2: {
		color1: '#ADB1FF',
		color2: '#FFEED4'
	},
	3: {
		color1: '#ADFFDD',
		color2: '#F4FFD4'
	},
	4: {
		color1: '#FFFDC1',
		color2: '#FBFFED'
	},
	5: {
		color1: '#FFEED4',
		color2: '#F4FFD4'
	}
};

export const AccountErrMessage: Record<
	string,
	{ message: string; showLink: boolean }
> = {
	CLOUD_ADMIN: {
		message: 'err_cloud_admin',
		showLink: true
	},
	CLOUD_SUB: {
		message: 'err_cloud_sub',
		showLink: false
	},
	SELFHOSTED_ADMIN: {
		message: 'err_selfhosted_admin',
		showLink: true
	},
	SELFHOSTED_SUB: {
		message: 'err_selfhosted_sub',
		showLink: false
	}
};

export {
	name,
	disableExternal,
	baseURL,
	logoURL,
	signup,
	version,
	noAuth,
	authMethod,
	loginPage,
	theme,
	enableThumbs,
	resizePreview,
	enableExec,
	origin,
	fileList
};
