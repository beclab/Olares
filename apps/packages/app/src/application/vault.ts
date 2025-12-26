import { walletService } from 'src/wallet';
import { NormalApplication } from './base';
import platform from './platform/vault';
import { setPlatform } from '@didvault/sdk/src/core';
import { useCloudStore } from 'src/stores/cloud';
import { useUserStore } from 'src/stores/user';
import { busEmit, NetworkErrorMode } from 'src/utils/bus';
import { WebsocketSharedWorkerEnum } from 'src/websocket/interface';
import { registerTranslateMethod } from '@didvault/sdk/src/util';
import { i18n } from 'src/boot/i18n';
import { GolbalHost } from '@bytetrade/core';

export class VaultApplication extends NormalApplication {
	applicationName = 'vault';
	platform = new platform.WebPlatform();
	ssiRule = () => {
		const userStore = useUserStore();
		return GolbalHost.userNameToEnvironment(
			userStore.terminusInfo().olaresId || userStore.terminusInfo().terminusName
		);
	};

	async appLoadPrepare(data: any): Promise<void> {
		//@ts-ignore
		(() => import('../css/styles.css'))();
		registerTranslateMethod(i18n.global.t, 'locale.');
		super.appLoadPrepare(data);
		setPlatform(this.platform);
		walletService.load();
		await this.platform.appLoadPrepare(data);
	}
	async appMounted(): Promise<void> {
		await super.appMounted();
		await this.platform.appMounted();
	}
	async appUnMounted(): Promise<void> {
		await super.appMounted();
		await this.platform.appUnMounted();
	}
	async appRedirectUrl(redirect: any): Promise<void> {
		if (this.platform) {
			await this.platform.appRedirectUrl(redirect);
		}
	}

	getWSConnectUrl() {
		if (process.env.IS_PC_TEST) {
			return ['ws://localhost:5300'];
		}
		const userStore = useUserStore();
		if (!userStore.connected) {
			return [];
		}
		return [userStore.getModuleSever('vault', 'wss:', '/ws')];
	}

	websocketConfig = {
		useShareWorker: true,
		shareWorkerName: WebsocketSharedWorkerEnum.VAULT_NAME,

		externalInfo() {
			return {};
		},
		responseShareWorkerMessage(data: {
			type: 'ws' | 'reconnected';
			data: any;
		}) {
			console.log('data ===>', data);
			try {
				const body: any = JSON.parse(data.data);
				busEmit('receiveMessage', body);
			} catch (e) {
				console.error('message error:', e);
			}
		}
	};

	initAxiosIntercepts(): void {
		super.initAxiosIntercepts();
		this.requestIntercepts.push((config) => {
			config.headers['X-Unauth-Error'] = 'Non-Redirect';
			return config;
		});

		this.responseIntercepts.push((response) => {
			if (
				!response ||
				(response.status != 200 && response.status != 201) ||
				!response.data
			) {
				throw Error('Network error, please try again later');
			}

			const cloudStore = useCloudStore();
			if (
				response.config.url &&
				response.config.url.indexOf(cloudStore.getUrl()) !== -1 &&
				response.config.method === 'post'
			) {
				const data = response.data;

				if (data.code === 506) {
					//cloudStore.removeToken();

					// router.push({ path: '/login' });
					return response;
				}

				if (data && data.code === 401) {
					return response;
				}

				if (data.code == 200 || data.code == 0) {
					return data;
				}

				throw new Error(data.message);
			} else {
				const data = response.data;
				if (data.code == 100001) {
					busEmit('network_error', {
						type: NetworkErrorMode.axois,
						error: 'token expired'
					});
					//store.commit("account/remove");

					// router.push({ path: '/login' });
					throw Error(data.message);
					//return response;
				}

				if (data.code != 0) {
					throw Error(data.message);
				}

				return data.data;
			}
		});
	}
}
