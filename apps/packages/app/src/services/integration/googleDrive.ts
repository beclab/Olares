import {
	AccountAddMode,
	GoogleIntegrationAccount,
	OperateIntegrationAuth
} from '../abstractions/integration/integrationService';
import { getAppPlatform } from 'src/application/platform';
import { axiosInstanceProxy } from 'src/platform/httpProxy';
import { i18n } from 'src/boot/i18n';
import { AccountType, IntegrationAccountMiniData } from '@bytetrade/core';
import { LarePassSocialLogin } from 'src/platform/interface/capacitor/plugins/social';

export class GoogleAuthService extends OperateIntegrationAuth<GoogleIntegrationAccount> {
	type = AccountType.Google;
	addMode = AccountAddMode.common;
	async signIn(): Promise<GoogleIntegrationAccount> {
		const scopes = [
			'https://www.googleapis.com/auth/drive',
			'https://www.googleapis.com/auth/drive.file'
		];
		await LarePassSocialLogin.initialize({
			google: {
				webClientId:
					'xxxx',
				iOSClientId:
					'xxxx',
				mode: getAppPlatform().getQuasar()?.platform.is?.android
					? 'offline'
					: 'online' // replaces grantOfflineAccess
			}
		});
		if (getAppPlatform().getQuasar()?.platform.is?.ios) {
			await LarePassSocialLogin.logout({
				provider: 'google'
			});
		}
		const googleResponse: any = await LarePassSocialLogin.login({
			provider: 'google',
			options: {
				scopes: scopes,
				forceRefreshToken: true // if you need refresh token
			}
		});
		let response;

		let clientId = '';

		if (getAppPlatform().getQuasar()?.platform.is?.android) {
			clientId =
				'xxx';
			response = await axiosInstanceProxy(
				{
					baseURL: 'https://cloud-api.jointerminus.com/',
					timeout: 10000,
					headers: {
						'Content-Type': 'application/json'
					}
				},
				false
			).post('/v1/common/google/token', {
				code: googleResponse.result.serverAuthCode
			});
			console.log(response);
			if (!response || response.data.code !== 200 || !response.data.data) {
				throw new Error(
					'Exchange authorization code error ' + response.data.data
						? response.data.message
						: ''
				);
			}
		} else {
			clientId =
				'xxxx';
		}

		const result = {
			name: googleResponse.result.profile.email,
			type: AccountType.Google,
			raw_data: {
				access_token: response
					? response.data.data.accessToken
					: googleResponse.result.accessToken.token,
				refresh_token: response
					? response.data.data.refreshToken
					: googleResponse.result.accessToken.refreshToken || '',
				expires_at: Date.now() + 30 * 60 * 1000,
				expires_in: 30 * 60 * 1000,
				scope: scopes.join(','),
				id_token: response
					? response.data.data.idToken
					: googleResponse.result.idToken,
				client_id: clientId
			}
		};
		return result;
	}
	async permissions() {
		return {
			title: i18n.global.t(
				'Your Google account grants us the following permissions:'
			),
			scopes: [
				{
					introduce: i18n.global.t('See your profile info'),
					icon: 'sym_r_account_circle'
				},
				{
					introduce: i18n.global.t(
						'See, edit, create, and delete all of your Google Drive files'
					),
					icon: 'sym_r_cloud'
				},
				{
					introduce: i18n.global.t(
						'See, edit, share, and permanently delete all the calendars you can access using Google Calendar'
					),
					icon: 'sym_r_calendar_today'
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
