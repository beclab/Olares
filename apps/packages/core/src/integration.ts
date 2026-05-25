export enum AccountType {
	Space = 'space',
	Google = 'google',
	Dropbox = 'dropbox',
	OneDrive = 'onedrive',
	AWSS3 = 'awss3',
	Tencent = 'tencent'
}

export interface IntegrationAccountBaseData {
	refresh_token: string;
	access_token: string;
	expires_in: number;
	expires_at: number;
}

export interface GoogleIntegrationAccountData
	extends IntegrationAccountBaseData {
	scope: string;
	id_token: string;
}

export interface SpaceIntegrationAccountData
	extends IntegrationAccountBaseData {
	userid: string;
}

export interface AWSS3IntegrationAccountData
	extends IntegrationAccountBaseData {
	endpoint: string;
	bucket: string;
}

export interface TencentIntegrationAccountData
	extends IntegrationAccountBaseData {
	endpoint: string;
	bucket: string;
}

export interface IntegrationAccountMiniData {
	name: string;
	type: AccountType;
	expires_at: number;
	available: boolean;
	create_at: number;
}
