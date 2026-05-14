import { CacheRequestBarrier } from 'src/stores/market/CacheRequestBarrier';
import { useWebsocketManager2Store } from 'src/stores/websocketManager2';
import { WebsocketSharedWorkerEnum } from 'src/websocket/interface';
import { usePreCheckStore } from 'src/stores/market/preCheck';
import { useSettingStore } from 'src/stores/market/setting';
import { useDeviceStore } from 'src/stores/settings/device';
import { BtNotify, NotifyDefinedType } from '@bytetrade/ui';
import { useCenterStore } from 'src/stores/market/center';
import { useAppStore } from 'src/stores/market/appStore';
import { useTokenStore } from 'src/stores/settings/token';
import { useUserStore } from '../stores/settings/user';
import { useMenuStore } from '../stores/market/menu';
import globalConfig from 'src/api/market/config';
import { bus, BUS_EVENT } from '../utils/bus';
import { DeviceType } from '@bytetrade/core';
import { NormalApplication, getApplication } from './base';
import { supportLanguages } from '../i18n';
import { i18n } from 'src/boot/i18n';
import axios from 'axios';

export class MarketApplication extends NormalApplication {
	applicationName = 'market';
	socketStore = useWebsocketManager2Store();
	preCheckStore = usePreCheckStore();
	menuStore = useMenuStore();

	protected shouldEnsureSocketAlive(): boolean {
		return !globalConfig.isOfficial;
	}

	onErrorMessage = (failureMessage: string) => {
		BtNotify.show({
			type: NotifyDefinedType.FAILED,
			message: i18n.global.t(failureMessage)
		});
	};

	refreshData = async () => {
		const appStore = useAppStore();
		const centerStore = useCenterStore();
		await appStore.init();
		const sourceRequest = appStore.initSourceRequest();
		centerStore.init();

		const settingStore = useSettingStore();
		settingStore.init();
		const settingRequest = settingStore.initConfigRequest();

		if (!globalConfig.isOfficial) {
			const preCheckRequest = this.preCheckStore.initSystemRequest();

			const barrier = new CacheRequestBarrier(
				['user', 'setting', 'source'],
				(data, isFromCache) => {
					console.log('CacheRequestBarrier callback');
					if (isFromCache) {
						console.log('CacheRequestBarrier first (from cache)');
						centerStore.loadFirstData();
					} else {
						console.log('CacheRequestBarrier refresh');
						centerStore.fetchNewData(true);
					}
				}
			);
			barrier.addRequest('user', preCheckRequest);
			barrier.addRequest('setting', settingRequest);
			barrier.addRequest('source', sourceRequest);
			const tokenStore = useTokenStore();
			const host = window.location.origin;
			tokenStore.setUrl(host);
			const userStore = useUserStore();
			userStore.get_accounts();
		} else {
			const barrier = new CacheRequestBarrier(
				['setting', 'source'],
				(data, isFromCache) => {
					console.log('CacheRequestBarrier callback');
					if (isFromCache) {
						console.log('CacheRequestBarrier first (from cache)');
						centerStore.loadFirstData();
					} else {
						console.log('CacheRequestBarrier refresh');
						centerStore.fetchNewData(true);
					}
				}
			);
			barrier.addRequest('setting', settingRequest);
			barrier.addRequest('source', sourceRequest);
		}
	};

	initializeApp = async () => {
		const deviceStore = useDeviceStore();
		const menuStore = useMenuStore();
		deviceStore.init(
			(state: { device: DeviceType; isVerticalScreen: boolean }) => {
				if (!deviceStore.isMobile && state.device === DeviceType.MOBILE) {
					menuStore.leftDrawerOpen = false;
				}
			}
		);
		menuStore.leftDrawerOpen = !deviceStore.isMobile;

		if (!globalConfig.isOfficial) {
			this.socketStore.start();
		}

		await this.refreshData();
	};

	async appLoadPrepare(data: any) {
		super.appLoadPrepare(data);
		let terminusLanguage = '';
		const terminusLanguageInfo: any = document.querySelector(
			'meta[name="terminus-language"]'
		);
		if (terminusLanguageInfo && terminusLanguageInfo.content) {
			terminusLanguage = terminusLanguageInfo.content;
		} else {
			terminusLanguage = navigator.language;
		}

		console.log(navigator.language);

		if (terminusLanguage) {
			if (supportLanguages.find((e) => e.value == terminusLanguage)) {
				i18n.global.locale.value = terminusLanguage as any;
			}
		}
	}

	async appMounted(): Promise<void> {
		await super.appMounted();
		await this.initializeApp();
		bus.on(BUS_EVENT.APP_BACKEND_ERROR, this.onErrorMessage);
	}

	async appRedirectUrl(): Promise<void> {}

	async appUnMounted(): Promise<void> {
		await super.appUnMounted();
		bus.off(BUS_EVENT.APP_BACKEND_ERROR, this.onErrorMessage);
	}

	websocketConfig = {
		useShareWorker: true,
		shareWorkerName: WebsocketSharedWorkerEnum.MARKET_NAME,
		externalInfo() {
			return {};
		},
		responseShareWorkerMessage(data: {
			type: 'ws' | 'reconnected';
			data: any;
		}) {
			if (data.type == 'ws') {
				try {
					const body = JSON.parse(data.data);
					if (body) {
						console.log(body);
						const appStore = useAppStore();
						if (body['notify_type'] == 'app_state_change') {
							appStore.updateAppStatusBySocket(body);
						} else if (body['notify_type'] == 'market_system_point') {
							appStore.updateMarketSystemBySocket(body);
						} else if (body['notify_type'] == 'image_state_change') {
							appStore.updateDownloadedImageSizeBySocket(body);
						} else if (body['notify_type'] == 'payment_state_update') {
							appStore.updateLocalStatus(
								body.extensions.app_name,
								body.extensions.source_id,
								{
									status: body.extensions.status,
									data: body?.extensions_obj?.payment_data
								}
							);
						} else if (body['eventType'] == 'usersUpdate') {
							const userStore = useUserStore();
							userStore.get_accounts();
							const preCheckStore = usePreCheckStore();
							const request = preCheckStore.initSystemRequest();
							request.get();
						}
					}
				} catch (e) {
					console.log('message error');
					console.log(e);
				}
			} else if (data.type == 'reconnected') {
				console.log('[market] socket reconnected, refreshing data');
				(getApplication() as MarketApplication).refreshData().catch((err) => {
					console.error('Error refreshing market data after reconnect:', err);
				});
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
			const data = response.data;

			if (data.code == 100001) {
				const config = response.config;
				const tokenStore = useTokenStore();
				return new Promise((resolve) => {
					this.addCallbacks(() => {
						config.headers['X-Authorization'] =
							tokenStore.$state.token?.access_token;
						resolve(axios.request(config));
					});
				});
			}

			if (response.config.url!.indexOf('/api/users') >= 0) {
				return response.data.data;
			}
			return data;
		});
	}

	callbacks: any[] = [];

	addCallbacks(callback: () => void) {
		this.callbacks.push(callback);
	}
}
