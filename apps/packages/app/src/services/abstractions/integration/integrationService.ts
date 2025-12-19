import {
	AccountType,
	IntegrationAccountBaseData,
	SpaceIntegrationAccountData,
	AWSS3IntegrationAccountData,
	TencentIntegrationAccountData,
	IntegrationAccountMiniData
} from '@bytetrade/core';

export interface GoogleIntegrationAccountData
	extends IntegrationAccountBaseData {
	scope: string;
	id_token: string;
	client_id: string;
}

export interface IntegrationAccount {
	name: string;
	type: AccountType;
	raw_data: IntegrationAccountBaseData;
}

export interface GoogleIntegrationAccount extends IntegrationAccount {
	raw_data: GoogleIntegrationAccountData;
}

export interface SpaceIntegrationAccount extends IntegrationAccount {
	raw_data: SpaceIntegrationAccountData;
}

export interface AWSS3IntegrationAccount extends IntegrationAccount {
	raw_data: AWSS3IntegrationAccountData;
}

export interface TencentIntegrationAccount extends IntegrationAccount {
	raw_data: TencentIntegrationAccountData;
}

export interface IntegrationAuthResult {
	status: boolean;
	account?: IntegrationAccount;
	message: string;
	addMode: AccountAddMode;
}

export interface IntegrationAccountInfoDetail {
	name: string;
	icon: string;
}

export interface IntegrationAccountInfo {
	type: AccountType;
	// name: string;
	detail: IntegrationAccountInfoDetail;
}

export interface IntegrationScopesDetail {
	icon: string;
	introduce: string;
}

export interface IntegrationPermissions {
	title: string;
	scopes: IntegrationScopesDetail[];
}

export enum AccountAddMode {
	common = 1,
	direct = 2
}

export interface IntegrationWebSupportAuth {
	status: boolean;
	message: string;
}

export abstract class OperateIntegrationAuth<T extends IntegrationAccount> {
	type: AccountType;
	addMode: AccountAddMode;
	abstract signIn(options: any): Promise<T | undefined>;
	abstract permissions(): Promise<IntegrationPermissions>;

	abstract webSupport(): Promise<IntegrationWebSupportAuth>;
	abstract detailPath(account: IntegrationAccountMiniData): string;
}

export interface IntegrationService {
	supportAuthList: IntegrationAccountInfo[];
	getAccountByType(
		request_type: AccountType
	): IntegrationAccountInfo | undefined;
	requestIntegrationAuth(
		request_type: AccountType,
		options: any
	): Promise<IntegrationAuthResult>;
	getInstanceByType(
		request_type: AccountType
	): OperateIntegrationAuth<IntegrationAccount> | undefined;
}
