import { TokenData, useCloudStore } from 'src/stores/cloud';
import {
	AccountAddMode,
	OperateIntegrationAuth,
	SpaceIntegrationAccount
} from '../abstractions/integration/integrationService';
import { useUserStore } from 'src/stores/user';
import { uid } from 'quasar';
import { getDID, getPrivateJWK } from 'src/did/did-key';
import { PrivateJwk } from '@bytetrade/core';
import { signJWS } from 'src/layouts/dialog/sign';
import axios from 'axios';
import { Loading } from 'quasar';
import { AccountType } from '@bytetrade/core';
import { i18n } from 'src/boot/i18n';
import { useDeviceStore } from 'src/stores/settings/device';

export class SpaceAuthService extends OperateIntegrationAuth<SpaceIntegrationAccount> {
	type = AccountType.Space;
	addMode = AccountAddMode.common;
	async signIn(): Promise<SpaceIntegrationAccount> {
		return new Promise(async (resolve, reject) => {
			Loading.hide();
			const cloudStore = useCloudStore();
			const userStore = useUserStore();

			const time = new Date().getTime();
			const secret = uid().replace(/-/g, '');
			const sign_body = {
				did: userStore.current_user?.id,
				secret,
				time
			};

			if (!(await userStore.unlockFirst())) {
				reject('Need login');
				return;
			}
			Loading.show();

			const did = await getDID(userStore.current_mnemonic?.mnemonic);
			const privateJWK: PrivateJwk | undefined = await getPrivateJWK(
				userStore.current_mnemonic?.mnemonic
			);

			const jws = await signJWS(did, sign_body, privateJWK);

			const postData = {
				id: secret,
				jws,
				did,
				body: sign_body
			};
			try {
				const loginUrl = cloudStore.getUrl() + '/v2/user/login';
				await axios.post(loginUrl, postData);
				const activeLogin = cloudStore.getUrl() + '/v2/user/activeLogin';
				const response: any = await axios.post(activeLogin, {
					secret
				});

				const loginToken: TokenData = response.data;
				if (response.code == 200) {
					resolve({
						name: userStore.current_user?.name || did,
						type: AccountType.Space,
						raw_data: {
							refresh_token: loginToken.token,
							access_token: loginToken.token,
							expires_in: 30 * 60 * 1000,
							expires_at: Math.trunc(loginToken.expired),
							userid: did
						}
					});
				} else {
					reject('Login fail');
				}
			} catch (error) {
				reject(error);
			}
		});
	}
	async permissions() {
		return {
			title: '',
			scopes: []
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

	detailPath() {
		const deviceStore = useDeviceStore();
		if (deviceStore.isMobile) {
			return '';
		}
		return '/integration/detail/space';
	}
}
