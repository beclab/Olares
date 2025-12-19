import { NormalApplication } from './base';
import { useWebsocketManager2Store } from 'src/stores/websocketManager2';
import { useAdminStore } from 'src/stores/settings/admin';
import { useTokenStore } from 'src/stores/settings/token';
import { useHeadScaleStore } from 'src/stores/settings/headscale';
import { useApplicationStore } from 'src/stores/settings/application';
import { useAccountStore } from 'src/stores/settings/account';
import { useBackgroundStore } from 'src/stores/settings/background';
import { WebPlatform } from '../utils/settings/platform';
import { languagesShort, supportLanguages, SupportLanguageType } from '../i18n';
import { importFilesStyle } from './utils/files';
import axios from 'axios';
import { useBackupStore } from '../stores/settings/backup';
import { useIntegrationStore } from '../stores/settings/integration';
import { useDeviceStore } from 'src/stores/settings/device';
import { WebsocketSharedWorkerEnum } from 'src/websocket/interface';
import { bus } from 'src/utils/bus';
import qs from 'qs';
import { notifyFailed } from 'src/utils/settings/btNotify';
import { DeviceType } from '@bytetrade/core';
import { useRouter } from 'vue-router';
import { useCookieStore } from 'src/stores/settings/cookie';

export class SettingsApplication extends NormalApplication {
	applicationName = 'settings';

	async appLoadPrepare(data: any): Promise<void> {
		//@ts-ignore
		(() => import('../css/styles.css'))();
		await super.appLoadPrepare(data);

		const tokenStore = useTokenStore();
		const headScaleStore = useHeadScaleStore();
		const backupStore = useBackupStore();
		const integrationStore = useIntegrationStore();

		const host = window.location.origin;
		tokenStore.setUrl(host);
		headScaleStore.setUrl(host + '/headscale');
		if (process.env.VERSIONTAG !== '1.11') {
			await backupStore.init();
		}
		integrationStore.getAccount('all');
	}

	async appMounted(): Promise<void> {
		await super.appMounted();
		const platform = new WebPlatform();
		const adminStore = useAdminStore();
		const deviceStore = useDeviceStore();
		importFilesStyle(deviceStore.isMobile);

		platform.getDeviceInfo().then((deviceInfo) => {
			adminStore.thisDevice = deviceInfo;
		});

		const websocketStore = useWebsocketManager2Store();
		websocketStore.start();
	}

	async appRedirectUrl(): Promise<void> {
		const tokenStore = useTokenStore();
		const applicationStore = useApplicationStore();

		const host = window.location.origin;
		tokenStore.setUrl(host);

		const router = useRouter();
		const adminStore = useAdminStore();
		const accountStore = useAccountStore();
		const backgroundStore = useBackgroundStore();
		backgroundStore.init();

		let terminusLanguage = '';
		const terminusLanguageInfo: any = document.querySelector(
			'meta[name="terminus-language"]'
		);
		if (terminusLanguageInfo && terminusLanguageInfo.content) {
			terminusLanguage = terminusLanguageInfo.content;
		} else {
			terminusLanguage = navigator.language || (navigator as any).userLanguage;
		}

		if (terminusLanguage) {
			if (languagesShort[terminusLanguage]) {
				backgroundStore.updateLanguageLocale(languagesShort[terminusLanguage]);
			} else if (supportLanguages.find((e) => e.value == terminusLanguage)) {
				backgroundStore.updateLanguageLocale(
					terminusLanguage as SupportLanguageType
				);
			}
		}

		const deviceStore = useDeviceStore();
		deviceStore.init(
			(state: { device: DeviceType; isVerticalScreen: boolean }) => {
				console.log(state);
				if (!deviceStore.isMobile && state.device === DeviceType.MOBILE) {
					router.replace('/');
				}
				const backgroundStore = useBackgroundStore();
				backgroundStore.updateBodyBg();
			}
		);

		return axios
			.get(tokenStore.url + '/api/init')
			.then((data: any) => {
				adminStore.terminus = data.terminusInfo;
				if (adminStore.terminus?.olaresId) {
					const cookieStore = useCookieStore();
					cookieStore.init(
						adminStore.terminus.olaresId.split('@')[0],
						window.location.origin,
						true
					);
				}
				adminStore.user = data.userInfo;
				applicationStore.applications = data.applicationData;
				accountStore.secrets = data.secrets;
				adminStore.devices = data.devices;
				backgroundStore.wallpaper = data.wallpaper;
			})
			.then(() => {
				// upgradeStore.checkLastOsVersion();
			});
	}

	async appUnMounted(): Promise<void> {
		await super.appUnMounted();
	}

	websocketConfig = {
		useShareWorker: true,
		shareWorkerName: WebsocketSharedWorkerEnum.SETTINGS_NAME,
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
						const applicationStore = useApplicationStore();
						const backupStore = useBackupStore();
						if (body['notify_type'] == 'app_state_change') {
							applicationStore.updateOneApplicationState(
								body.app_name,
								body.app_state_latest.status.state,
								body.app_state_latest.status.entranceStatuses
							);
							bus.emit('entrance_state_event', body);
						} else if (body.type && body.type === 'backup') {
							backupStore.updateBackupBySocket(body);
							bus.emit('backup_state_event', body);
						} else if (body.type && body.type === 'restore') {
							backupStore.updateRestoreBySocket(body);
							bus.emit('restore_state_event', body);
						} else if (body.type == 'olaresStatusUpdate') {
							bus.emit('olaresStatusUpdate', body.data);
						}
					}
				} catch (e) {
					console.log('message error');
					console.log(e);
				}
			} else if (data.type == 'reconnected') {
				const applicationStore = useApplicationStore();
				applicationStore.get_applications().then(() => {
					bus.emit('entrance_state_event');
				});
			}
		}
	};

	initAxiosIntercepts() {
		super.initAxiosIntercepts();
		this.requestIntercepts.push((config) => {
			const store = useAccountStore();
			if (
				config.url &&
				config.url.indexOf(store.space.url) !== -1 &&
				config.method === 'post'
			) {
				config.data = qs.stringify(config.data || {});
				return config;
			} else {
				config.headers['Access-Control-Allow-Origin'] = '*';
				config.headers['Access-Control-Allow-Headers'] =
					'X-Requested-With,Content-Type';
				config.headers['Access-Control-Allow-Methods'] =
					'PUT,POST,GET,DELETE,OPTIONS';
				config.headers['X-Unauth-Error'] = 'Non-Redirect';
				return config;
			}
		});

		this.responseIntercepts.push((response) => {
			const store = useAccountStore();
			if (
				response.config.url &&
				response.config.url.indexOf(store.space.url) !== -1 &&
				response.config.method === 'post'
			) {
				const data = response.data;

				if (data && data.code === 401) {
					return response;
				}

				if (data.code !== 200) {
					throw new Error(data.message);
				}

				return data;
			} else {
				if (
					!response ||
					(response.status != 200 && response.status != 201) ||
					!response.data
				) {
					notifyFailed('Network error, please try again later');
					throw Error('Network error, please try again later');
				}

				const data = response.data;
				if (data.code == 100001) {
					//store.commit("account/remove");
					if (data.message) {
						notifyFailed('' + data.code + ' ' + data.message);
					}
					// router. push( { path : '/login' });
					throw Error(data.message);
					//return response;
				}

				if (
					response.config.url!.indexOf('kapis') >= 0 ||
					response.config.url!.indexOf('ndbq.ursa-services.bttcdn.com') >= 0 ||
					response.config.url!.indexOf('/api/system/status') >= 0
				) {
					return data;
				} else if (response.config.url!.indexOf('acl/app/status?name=') >= 0) {
					//return app acl not found
					return data;
				} else {
					if (data.code != 0 && data.code != 200 && !data.items) {
						//kapis return used in login history
						//kapis return used in login history
						if (
							data.message &&
							response.config.url!.indexOf('permissions') < 0
						) {
							notifyFailed(
								data.message || 'Something wrong, please try again.'
							);
						}
						//return response;
						throw Error(data.message);
					}

					if (data.code == 0 || data.code == 200) {
						return data.data;
					} else {
						return data; //kapis return used in login history
					}
				}
			}
		});
	}
}
