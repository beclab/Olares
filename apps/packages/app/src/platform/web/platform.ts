import { SubAppPlatform } from '../subAppPlatform';

import { app } from '../../globals';
import { useUserStore } from '../../stores/user';
import {
	homeMounted as commonHomeMounted,
	homeUnMounted as commonHomeUnMounted
} from '../homeLayoutCommon';
import {
	AuthPurpose,
	AuthType,
	DeviceInfo,
	Err,
	ErrorCode,
	getPlatform as defaultGetPlatform
} from '@didvault/sdk/src/core';
import axios from 'axios';
import { MessageTopic, OlaresInfo, TerminusInfo } from '@bytetrade/core';
import { useMenuStore } from 'src/stores/menu';
import { Router } from 'vue-router';
import TerminusDesktopTipDialog from '../../components/dialog/TerminusDesktopTipDialog.vue';

import { QVueGlobals } from 'quasar';
import { i18n } from '../../boot/i18n';
import { _authenticate } from '@didvault/sdk/src/authenticate';
import { busOff, busOn } from 'src/utils/bus';

export const getWebPlatform = () => {
	return defaultGetPlatform() as WebPlatform;
};

export class WebPlatform extends SubAppPlatform {
	router: Router | undefined;
	quasar: QVueGlobals | undefined;
	// deal more dialog
	dealInvalidSession = false;

	async appLoadPrepare(data: any): Promise<void> {
		super.appLoadPrepare(data);
		if (data.router) {
			this.router = data.router;
		}

		if (data.quasar) {
			this.quasar = data.quasar;
		}

		app.load(undefined);

		busOn('receiveMessage', (message: any | string) => {
			const body: any =
				typeof message == 'string' ? JSON.parse(message) : message;
			if (body.topic == MessageTopic.Data) {
				if (
					body.event == 'vault.account.update' ||
					body.event == 'vault.org.update'
				) {
					if (!app.state.locked) {
						app.synchronize();
					}
				}
			} else if (
				body.eventType == 'vault.account.update' ||
				body.eventType == 'vault.org.update'
			) {
				if (!app.state.locked) {
					app.synchronize();
				}
			}
		});
	}
	async appMounted(): Promise<void> {
		super.appMounted();
		//window.addEventListener('load', onload);
	}
	async appUnMounted(): Promise<void> {
		super.appUnMounted();
		//window.removeEventListener('load', onload);
	}

	async appRedirectUrl(redirect: any): Promise<void> {
		const userStore = useUserStore();
		return this.getTerminusInfo()
			.then((data) => {
				if (data.olaresId && data.wizardStatus == 'completed') {
					return userStore.load().then(() => {
						userStore.currentUserSaveTerminusInfo(data);
						if (userStore.isBooted) {
							redirect({ path: '/unlock' });
						} else {
							redirect({ path: '/setUnlockPassword' });
						}
					});
				} else {
					userStore.currentUserSaveTerminusInfo(data);
					redirect({ path: '/binding' });
				}
			})
			.catch(() => {
				redirect({ path: '/error' });
			});
	}

	stateUpdate() {
		const menuStore = useMenuStore();
		menuStore.updateMenuInfo();

		const userStore = useUserStore();

		if (app.state._errors.length > 0) {
			const INVALID_SESSION = app.state._errors.find(
				(e) => e.code == ErrorCode.INVALID_SESSION
			)
				? true
				: false;

			if (INVALID_SESSION && !getWebPlatform().dealInvalidSession) {
				getWebPlatform().dealInvalidSession = true;
				getWebPlatform()
					.quasar?.dialog({
						component: TerminusDesktopTipDialog,
						componentProps: {
							title: i18n.global.t('reconnect_olares'),
							message: i18n.global.t('reconnect_olares_message'),
							confirmBtnTitle: i18n.global.t('ok'),
							cancelBtnTitle: i18n.global.t('cancel'),
							showCancel: true
						}
					})
					.onOk(async () => {
						getWebPlatform().quasar?.loading.show();
						try {
							await app.clearSession();
							const authRes = await _authenticate({
								did: userStore.current_user!.local_name,
								type: AuthType.SSI,
								purpose: AuthPurpose.Login
							});
							if (authRes == null) {
								throw new Err(
									ErrorCode.AUTHENTICATION_FAILED,
									i18n.global.t('errors.authentication_failed')
								);
							}
							await app.login({
								did: authRes.did,
								password: userStore.current_mnemonic!.mnemonic,
								authToken: authRes.token
							});
						} catch (e) {
							console.error(e);
						} finally {
							getWebPlatform().quasar?.loading.hide();
						}
					})
					.onDismiss(() => {
						getWebPlatform().dealInvalidSession = false;
					});
			}
			app.state._errors = [];
		}

		if (app.state.locked) {
			getWebPlatform().router?.push({
				path: '/unlock'
			});
		}
	}

	async homeMounted(): Promise<void> {
		commonHomeMounted();
		busOn('appSubscribe', this.stateUpdate);
	}

	async homeUnMounted(): Promise<void> {
		commonHomeUnMounted();
		busOff('appSubscribe', this.stateUpdate);
	}

	async getDeviceInfo(): Promise<DeviceInfo> {
		const userStore = useUserStore();
		const info = await super.getDeviceInfo();
		if (userStore.locale) {
			info.locale = userStore.locale.split('-')[0];
		} else {
			info.locale = i18n.global.locale.value.split('-')[0];
		}
		return info;
	}

	private async getTerminusInfo(
		terminus_url: string | null = null
	): Promise<OlaresInfo> {
		let baseUrl = terminus_url || window.location.origin;

		if (process.env.NODE_ENV === 'development') {
			baseUrl = '';
		}

		const data: any = await axios.get(baseUrl + '/bfl/info/v1/olares-info');

		return data;
	}
}
