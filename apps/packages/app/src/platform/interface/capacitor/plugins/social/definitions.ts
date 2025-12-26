import {
	AppleProviderResponse,
	FacebookLoginResponse,
	GoogleLoginResponseOnline,
	LoginOptions,
	TwitterLoginResponse
} from '@capgo/capacitor-social-login/dist/esm/definitions';
import { SocialLoginPlugin } from '@capgo/capacitor-social-login';
export * from './definitions';

export interface LarePassGoogleLoginResponseOffline {
	serverAuthCode: string;
	profile?: {
		email: string | null;
		familyName: string | null;
		givenName: string | null;
		id: string | null;
		name: string | null;
		imageUrl: string | null;
	};
	responseType: 'offline';
}
export type LarePassGoogleLoginResponse =
	| GoogleLoginResponseOnline
	| LarePassGoogleLoginResponseOffline;

export interface LarePassSocialLoginPlugin extends SocialLoginPlugin {
	/**
	 * Login with the selected provider
	 * @description login with the selected provider
	 */
	login<T extends LoginOptions['provider']>(
		options: Extract<
			LoginOptions,
			{
				provider: T;
			}
		>
	): Promise<{
		provider: T;
		result: LarePassProviderResponseMap[T];
	}>;
}

export type LarePassProviderResponseMap = {
	facebook: FacebookLoginResponse;
	google: LarePassGoogleLoginResponse;
	apple: AppleProviderResponse;
	twitter: TwitterLoginResponse;
};
