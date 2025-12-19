export interface Token {
	access_token: string;
	token_type: string;
	refresh_token: string;
	expires_in: number;
	expires_at: number;
}

export interface UserInfo {
	name: string;
	did: string;
	owner_role: string;

	avatar: string;
	imgContentMode: 'fill' | undefined;
	// | 'contain' | 'cover' | 'none' | 'scale-down' | undefined
}

export interface AccountInfo {
	uid: string;
	state: string;
	roles: string[];
	name: string;
	email: string;
	display_name: string;
	did: string;
	description: string;
	creation_timestamp: number;

	avatar: string;
}

export interface CreateAccountRequest {
	description: string;
	email: string;
	display_name: string;
	name: string;
	owner_role: string;
	password: string;
}

export interface RestAccountPassword {
	username: string;
	current_password: string;
	password: string;
}

export enum AccountStatus {
	Created = 'Created',
	Creating = 'Creating',
	Deleting = 'Deleting',
	Deleted = 'Deleted'
}

export interface AccountAddress {
	desktop: string;
	doc: string;
}

export interface AccountModifyStatus {
	name: string;
	status: AccountStatus;
	address: AccountAddress;
}

export interface ApplicationInfo {
	deployment: string;
	icon: string;
	id: string;
	name: string;
	namespace: string;
	owner: string;
	target: string;
	title: string;
	url: string;
}

export interface NotificationEmailConfig {
	config: {
		email: {
			from: string;
			smartHost: {
				host: string;
				port: number;
			};
			authUsername: string;
			authPassword: {
				value: string;
			};
			requireTLS: boolean;
		};
	};

	receiver: {
		email: {
			enabled: boolean;
			to: string[];
		};
	};
}
