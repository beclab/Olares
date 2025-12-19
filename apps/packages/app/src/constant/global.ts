export interface Token {
	access_token: string;
	token_type: string;
	refresh_token: string;
	expires_in: number;
	expires_at: number;
}

export interface UserInfo {
	created_user: string;
	is_ephemeral: boolean;
	name: string;
	owner_role: string;
	terminusName: string;
	wizard_complete: boolean;
	zone: string;

	imgContentMode: 'fill' | undefined;
}

export interface AccountInfo {
	uid: string;
	name: string;
	display_name: string;
	description: string;
	email: string;
	state: string;
	last_login_time: number;
	creation_timestamp: number;
	terminusName: string;
	wizard_complete: boolean;
	roles: string[];
	memory_limit: string;
	cpu_limit: string;

	avatar?: string;
}

export interface CreateAccountRequest {
	name: string;
	description?: string;
	email?: string;
	display_name?: string;
	owner_role: string;
	password: string;
	memory_limit: string;
	cpu_limit: string;
}

export interface UpdateAccountQuotaRequest {
	memory_limit: string;
	cpu_limit: string;
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
	Deleted = 'Deleted',
	Failed = 'Failed'
}

export interface AccountAddress {
	wizard: string;
	doc: string;
}

export interface AccountModifyStatus {
	name: string;
	status: AccountStatus;
	address: AccountAddress;
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
export interface AccessLevel {
	access_level: number | null;
	allow_cidrs: undefined | string[] | null;
	auth_policy: string | null;
}
export interface NewVersion {
	current_version: string;
	is_new: boolean;
	new_version: string;
}

export interface BackupInfo {
	backupConfigName: string;
	completionTimestamp: string;
	creationTimestamp: string;
	expiration: string;
	name: string;
	phase: string;
	s3Config: S3ConfigInfo;
	size: number;
	uid: string;
}
export interface S3ConfigInfo {
	bucket: string;
	prefix: string;
	provider: string;
	region: string;
}

export interface UsersParam {
	limit?: number;
	page?: number;
	sortBy?: string; //createTime
	globalrole?: string;
	email?: string;
	user?: string;
}

export enum SetupDomainStatus {
	None = 'None',
	Default = 'Default',
	ThirdLevel = 'ThirdLevel',
	ThirdParty = 'ThirdParty'
}

export interface SetupDomain {
	third_level_domain: string;
	third_party_domain: string;
	cname_status: string;
	cname_target: string;
	cname_target_status: string;
	cert?: string;
	key?: string;
}

export interface AuthLevel {
	authorization_level: string;
}

export interface Permission {
	dataType: string;
	group: string;
	ops: string[];
	version: string;
}

export interface AppPermission {
	app: string;
	owner: string;
	permissions: Permission[];
}

export interface ProviderRegisterApi {
	name: string;
	uri: string;
}

export interface PermissionProviderRegister {
	dataType: string;
	deployment: string;
	endpoint: string;
	group: string;
	kind: string;
	namespace: string;
	description: string;
	opApis: ProviderRegisterApi[];
	version: string;
}

export enum ProcessingState {
	Init = '',
	Completed = 'completed',
	Failed = 'failed',
	InProgress = 'in-progress'
}

export interface ApplicationCustonDomain {
	name: string;
	host: string;
	port: number;
	icon: string;

	title: string;
	authLevel: string;
	app_name: string;
	custom_domain: string;
}
