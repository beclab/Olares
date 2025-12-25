import { NormalApplication } from './base';
import axios from 'axios';
import { useTokenStore } from '../stores/desktop/token';
import { useAppStore } from '../stores/desktop/app';
import { useUpgradeStore } from '../stores/desktop/upgrade';
import { WebPlatform } from '../utils/desktop/platform';
import { useWebsocketManager2Store } from 'src/stores/websocketManager2';
import { useDevice, onDeviceChange, DeviceType } from '@bytetrade/core';
import { WebsocketSharedWorkerEnum } from 'src/websocket/interface';
import { MessageTopic } from '@bytetrade/core';
import { bus } from 'src/utils/bus';
import { commonInterceptValue } from 'src/utils/response';
import { useNotificationStore } from 'src/stores/desktop/notification';

export class DesktopApplication extends NormalApplication {
	applicationName = 'desktop';

	private refreshTimer: NodeJS.Timer | null = null;

	private initializeApp = async (isNetworkRestored = false) => {
		const appStore = useAppStore();
		const tokenStore = useTokenStore();
		const socketStore = useWebsocketManager2Store();
		appStore.get_my_apps_info(
			tokenStore.deviceInfo.device === DeviceType.MOBILE ? true : false
		);

		socketStore.start();
	};

	private onVisibilityChange = () => {
		const socketStore = useWebsocketManager2Store();
		if (document.visibilityState === 'visible') {
			if (socketStore.isClosed()) {
				this.initializeApp().catch((err) => {
					console.error('Error during app re-initialization:', err);
				});
			}
		}
	};

	private preFetch = async () => {
		const platform = new WebPlatform();
		// const tokenStore = useTokenStore();

		await platform.getDeviceInfo().then(async (deviceInfo) => {
			const loginInfo = await this.getWsLoginData();
			deviceInfo.id = loginInfo.id;
			deviceInfo.sso = loginInfo.token;
			try {
				axios.post('/api/device', deviceInfo);
			} catch (e) {
				console.log(e);
			}
		});
	};

	async appLoadPrepare(data): Promise<void> {
		await super.appLoadPrepare(data);
		const tokenStore = useTokenStore();
		await tokenStore.loadData();
		const notificationStore = useNotificationStore();
		await tokenStore.validateTerminusInfo(
			(currentId: string, lastId: string) => {
				return lastId.length == 0 || currentId == lastId;
			},
			async () => {
				await notificationStore.initDatas();
			},
			async () => {
				notificationStore.deleteAll();
			}
		);
		const upgradeStore = useUpgradeStore();
		upgradeStore.update_upgrade_state_info();
	}

	async appRedirectUrl(): Promise<void> {
		const tokenStore = useTokenStore();
		const { state } = useDevice();
		tokenStore.deviceInfo = state;

		const host = window.location.origin;
		tokenStore.setUrl(host);

		onDeviceChange(
			(state: { device: DeviceType; isVerticalScreen: boolean }) => {
				tokenStore.deviceInfo = state;
			}
		);

		this.preFetch();
	}

	async appMounted(): Promise<void> {
		document.addEventListener('visibilitychange', this.onVisibilityChange);
		const socketStore = useWebsocketManager2Store();
		await this.initializeApp(true);
		socketStore.start();
	}

	async appUnMounted() {
		document.addEventListener('visibilitychange', this.onVisibilityChange);

		const { cleanup } = useDevice();
		cleanup();
		this.refreshTimer && clearInterval(this.refreshTimer);
	}

	websocketConfig = {
		useShareWorker: true,
		shareWorkerName: WebsocketSharedWorkerEnum.DESKTOP_NAME,

		externalInfo() {
			return {};
		},
		responseShareWorkerMessage(data: {
			type: 'ws' | 'reconnected' | 'notification';
			data: any;
		}) {
			if (data.type == 'ws') {
				try {
					const message = JSON.parse(data.data);
					if (message.topic == MessageTopic.Data) {
						if (message.event == 'updateConfig') {
							const tokenStore = useTokenStore();
							tokenStore.config = message.message.data;
						}
					} else {
						if (message.event == 'app_installation_event') {
							bus.emit('app_installation_event', message);
						} else if (message.event == 'system_upgrade_event') {
							bus.emit('system_upgrade_event', message.data);
						} else if (message.event == 'ai') {
							bus.emit('ai', message);
						} else if (message.event == 'ai_message') {
							bus.emit('ai_message', message);
						} else if (message.event == 'intent') {
							if (document.visibilityState === 'visible') {
								bus.emit('intent', message.data);
							}
						} else if (message.event == 'n') {
							bus.emit('notification', message.data);
						} else if (message.event == 'entrance_state_event') {
							bus.emit('entrance_state_event', message);
						} else if (message.notify_type == 'app_state_change') {
							const appStore = useAppStore();
							appStore.updateOneApplicationState(
								message.app_name,
								message.app_state_latest.status.state,
								message.app_state_latest.status.entranceStatuses
							);
						}
					}
				} catch (e) {
					console.log('message error');
					console.log(e);
				}
			} else if (data.type == 'reconnected') {
				const upgradeStore = useUpgradeStore();
				const appStore = useAppStore();
				upgradeStore.update_upgrade_state_info();
				appStore.get_my_apps_info();
			} else if (data.type == 'notification') {
				const notificationStore = useNotificationStore();
				notificationStore.addItem(data.data);
			}
		}
	};

	initAxiosIntercepts(): void {
		super.initAxiosIntercepts();
		this.requestIntercepts.push((config) => {
			config.headers['Access-Control-Allow-Origin'] = '*';
			config.headers['Access-Control-Allow-Headers'] =
				'X-Requested-With,Content-Type';
			config.headers['Access-Control-Allow-Methods'] =
				'PUT,POST,GET,DELETE,OPTIONS';
			config.headers['X-Unauth-Error'] = 'Non-Redirect';

			return config;
		});

		this.responseIntercepts.push((response) => {
			const data = response.data;
			console.log('data ===>', data);

			if (
				!response ||
				(response.status != 200 &&
					response.status != 201 &&
					response.status != 304) ||
				!data
			) {
				throw Error('Network error, please try again later');
			}

			if (data.code === undefined) {
				return data;
			}

			if (data.code == 100001) {
				throw Error(data.message);
			}

			if (typeof data == 'string' && commonInterceptValue.includes(data)) {
				return data;
			}

			if (data.status) {
				if (data.status === 'OK') {
					return data.data;
				}
				throw Error(data.status);
			} else {
				if (data.code != 0) {
					throw Error(data.message);
				}

				return data.data;
			}
		});
	}
}
