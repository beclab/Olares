import { VaultItem, Tag, AuditType } from './core/item';
import { Vault, VaultID } from './core/vault';

export interface ListItem {
	item: VaultItem;
	vault: Vault;
	section?: string;
	firstInSection?: boolean;
	lastInSection?: boolean;
	warning?: boolean;
	selectedEnable?: boolean;
}

export interface ItemsFilter {
	vault?: VaultID;
	tag?: Tag;
	favorites?: boolean;
	attachments?: boolean;
	recent?: boolean;
	host?: boolean;
	report?: AuditType;
}

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
}

export interface LoginRequest {
	username: string;
	password: string;
}

export interface DIDAccount {
	did: string;
	masterPassword: string;
	url: string;
}

export enum ItemHeaderModel {
	SEARCH,
	DEFAULT,
	CHECKBOX
}
