import {
	AccountAddMode,
	IntegrationAccount,
	OperateIntegrationAuth
} from '../abstractions/integration/integrationService';
import { i18n } from 'src/boot/i18n';

import { AccountType, IntegrationAccountMiniData } from '@bytetrade/core';
import { DropboxAuth } from 'src/platform/interface/capacitor/plugins/dropbox';

export class DropboxAuthService extends OperateIntegrationAuth<IntegrationAccount> {
	type = AccountType.Dropbox;
	addMode = AccountAddMode.common;
	async signIn(): Promise<IntegrationAccount> {
		const dropboxSignInResponse = await DropboxAuth.signIn();
		return {
			name: dropboxSignInResponse.uid,
			type: AccountType.Dropbox,
			raw_data: {
				access_token: dropboxSignInResponse.accessToken,
				refresh_token: dropboxSignInResponse.refreshToken || '',
				expires_at: dropboxSignInResponse.tokenExpirationTimestamp
					? Math.trunc(dropboxSignInResponse.tokenExpirationTimestamp * 1000)
					: Date.now() + 30 * 60 * 1000,
				expires_in: dropboxSignInResponse.tokenExpirationTimestamp
					? Math.trunc(
							dropboxSignInResponse.tokenExpirationTimestamp * 1000 - Date.now()
					  )
					: 30 * 60 * 1000
			}
		};
	}
	async permissions() {
		return {
			title: i18n.global.t(
				'Your Dropbox account grants us the following permissions:'
			),
			scopes: [
				{
					introduce: i18n.global.t('See your profile info'),
					icon: 'sym_r_account_circle'
				},
				{
					introduce: i18n.global.t(
						'See, edit, create, and delete all of your Dropbox files'
					),
					icon: 'sym_r_cloud'
				}
			]
		};
	}

	async webSupport() {
		return {
			status: false,
			message: i18n.global.t(
				'Due to some restrictions, we do not support binding this type of account in Settings. Please use TermiPass mobile app to complete the account authorization and binding.'
			)
		};
	}
	detailPath(account: IntegrationAccountMiniData) {
		return '/integration/common/detail/' + account.type + '/' + account.name;
	}
}
