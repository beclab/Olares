export interface Token {
	username?: string;
	password?: string;
	access_token: string;
	refresh_token: string;
	fa2: boolean;
	redirect: string;
	session_id: string;
}

export interface SpaceTokenData {
	userid: string;
	token: string;
	expired: number;
	emailEnable?: boolean;
	emailMask?: string;
	phoneEnable?: boolean;
	phoneMask?: string;
	gaEnable?: boolean;
}

export interface SpaceSaveData {
	email: string;
	userid: string;
	token: string;
	expired: number;
}

export interface Secret {
	name: string;
	value: string;
}
