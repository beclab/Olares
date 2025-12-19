import { CacheRequestBarrier } from 'src/stores/market/CacheRequestBarrier';
import { useWebsocketManager2Store } from 'src/stores/websocketManager2';
import { WebsocketSharedWorkerEnum } from 'src/websocket/interface';
import { useSettingStore } from 'src/stores/market/setting';
import { useDeviceStore } from 'src/stores/settings/device';
import { BtNotify, NotifyDefinedType } from '@bytetrade/ui';
import { useCenterStore } from 'src/stores/market/center';
import { useTokenStore } from 'src/stores/market/token';
import { bus, BUS_EVENT, busEmit } from '../utils/bus';
import { useUserStore } from '../stores/market/user';
import { useMenuStore } from '../stores/market/menu';
import globalConfig from 'src/api/market/config';
import { DeviceType } from '@bytetrade/core';
import { NormalApplication } from './base';
import { supportLanguages } from '../i18n';
import { i18n } from 'src/boot/i18n';
import axios from 'axios';

export class MarketApplication extends NormalApplication {
	applicationName = 'market';
	socketStore = useWebsocketManager2Store();
	userStore = useUserStore();
	menuStore = useMenuStore();

	onErrorMessage = (failureMessage: string) => {
		BtNotify.show({
			type: NotifyDefinedType.FAILED,
			message: failureMessage
		});
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
		const centerStore = useCenterStore();
		if (!globalConfig.isOfficial) {
			const userRequest = this.userStore.init();
			const settingStore = useSettingStore();
			const settingRequest = settingStore.init();
			const centerRequest = await centerStore.init();
			this.socketStore.start();

			const barrier = new CacheRequestBarrier(
				['user', 'setting', 'center'],
				(data, isFirstLoad) => {
					console.log('CacheRequestBarrier callback');
					if (isFirstLoad) {
						console.log('CacheRequestBarrier first');
						centerStore.loadFirstData();
					} else {
						console.log('CacheRequestBarrier refresh');
						centerStore.fetchNewData(true);
					}
				}
			);
			barrier.addRequest('user', userRequest);
			barrier.addRequest('setting', settingRequest);
			barrier.addRequest('center', centerRequest);
		} else {
			const settingStore = useSettingStore();
			const settingRequest = settingStore.init();
			const centerRequest = await centerStore.init();
			const barrier = new CacheRequestBarrier(
				['setting', 'center'],
				(data, isFirstLoad) => {
					console.log('CacheRequestBarrier callback');
					if (isFirstLoad) {
						console.log('CacheRequestBarrier first');
						centerStore.loadFirstData();
					} else {
						console.log('CacheRequestBarrier refresh');
						centerStore.fetchNewData(true);
					}
				}
			);
			barrier.addRequest('setting', settingRequest);
			barrier.addRequest('center', centerRequest);
		}
	};

	onVisibilityChange = () => {
		if (document.visibilityState === 'visible') {
			console.log('Page is active');
			if (this.socketStore.isClosed()) {
				this.initializeApp().catch((err) => {
					console.error('Error during app re-initialization:', err);
				});
			}
		} else {
			console.log('Page is in background');
		}
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
		await this.initializeApp();
		bus.on(BUS_EVENT.APP_BACKEND_ERROR, this.onErrorMessage);
		document.addEventListener('visibilitychange', this.onVisibilityChange);
	}

	async appRedirectUrl(): Promise<void> {}

	async appUnMounted(): Promise<void> {
		await super.appUnMounted();
		bus.off(BUS_EVENT.APP_BACKEND_ERROR, this.onErrorMessage);
		document.removeEventListener('visibilitychange', this.onVisibilityChange);
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
						const center = useCenterStore();
						if (body['notify_type'] == 'app_state_change') {
							center.updateAppStatusBySocket(body);
						} else if (body['notify_type'] == 'market_system_point') {
							center.updateMarketSystemBySocket(body);
						} else if (body['notify_type'] == 'image_state_change') {
							center.updateDownloadedImageSizeBySocket(body);
						} else if (body['notify_type'] == 'payment_state_update') {
							center.updateLocalStatus(
								body.extensions.app_name,
								body.extensions.source_id,
								{
									status: body.extensions.status,
									data: body?.extensions_obj?.payment_data
								}
							);
						}
					}
				} catch (e) {
					console.log('message error');
					console.log(e);
				}
			} else if (data.type == 'reconnected') {
				const userStore = useUserStore();
				const centerStore = useCenterStore();
				if (!globalConfig.isOfficial) {
					const request = userStore.init();
					request.get();
					centerStore.fetchNewData();
				}
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
			return data;
		});
	}

	callbacks: any[] = [];

	addCallbacks(callback: () => void) {
		this.callbacks.push(callback);
	}
}
