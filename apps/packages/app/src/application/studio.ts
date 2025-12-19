import { bus } from '@apps/studio/src/utils/bus';
import { NormalApplication } from './base';
import { useDevelopingApps } from '@apps/studio/stores/app';
import { BtNotify, NotifyDefinedType } from '@bytetrade/ui';
import axios from 'axios';
import { useWebsocketManager2Store } from 'src/stores/websocketManager2';
import { WebsocketSharedWorkerEnum } from 'src/websocket/interface';
import { useDockerStore } from '@apps/studio/stores/docker';
import { replaceLastSubdomain } from 'src/utils/olares-url';

export class StudioApplication extends NormalApplication {
	applicationName = 'studio';
	async appLoadPrepare(data: any): Promise<void> {
		super.appLoadPrepare(data);

		const socketStore = useWebsocketManager2Store();
		socketStore.start();

		const dockerStore = useDockerStore();
		dockerStore.init();

		const appStore = useDevelopingApps();
		const host = window.location.origin;
		appStore.setUrl(host);
	}

	async appRedirectUrl(): Promise<void> {
		const appStore = useDevelopingApps();
		const host = window.location.origin;
		appStore.setUrl(host);
		return new Promise((resolve) => {
			appStore.getApps().then(() => {
				resolve();
			});
		});
	}

	getWSConnectUrl() {
		if (process.env.WS_URL) {
			return [process.env.WS_URL];
		}
		const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
		const newHost = replaceLastSubdomain('settings');
		const ws_url = `${protocol}//${newHost}/ws`;

		return [ws_url];
	}

	websocketConfig = {
		useShareWorker: true,
		shareWorkerName: WebsocketSharedWorkerEnum.STUDIO_NAME,

		externalInfo() {
			return {};
		},
		responseShareWorkerMessage(data: {
			type: 'ws' | 'reconnected';
			data: any;
		}) {
			try {
				const message = JSON.parse(data.data);

				if (message.type == 'app') {
					bus.emit('app_installation_event_studio', message);
				}
			} catch (e) {
				console.log('message error');
				console.log(e);
			}
		}
	};

	initAxiosIntercepts(): void {
		super.initAxiosIntercepts();

		this.requestIntercepts.push((config) => {
			if (config.headers) {
				config.headers['Access-Control-Allow-Origin'] = '*';
				config.headers['Access-Control-Allow-Headers'] =
					'X-Requested-With,Content-Type';
				config.headers['Access-Control-Allow-Methods'] =
					'PUT,POST,GET,DELETE,OPTIONS';
				config.headers['X-Unauth-Error'] = 'Non-Redirect';

				return config;
			} else {
				return config;
			}
		});

		this.responseIntercepts.push((response) => {
			if (!response || response.status != 200 || !response.data) {
				BtNotify.show({
					type: NotifyDefinedType.FAILED,
					message: response.status
				});
				throw Error('Network error, please try again later');
			}

			const res = response.data;

			let urlPath = '';
			try {
				if (response.config.url) {
					if (response.config.url.startsWith('http')) {
						urlPath = new URL(response.config.url).pathname;
					} else {
						urlPath = response.config.url;
					}
				}
			} catch (e) {
				urlPath = response.config.url || '';
			}

			if (
				urlPath.startsWith('/api/') &&
				!urlPath.startsWith('/api/v1/') &&
				!urlPath.startsWith('/api/command/install-app') &&
				res?.code !== 200
			) {
				BtNotify.show({
					type: NotifyDefinedType.FAILED,
					message: res.message
				});
				throw Error(res.message);
			}
			return res.data;
		});

		this.responseErrorInterceps = (error: any) => {
			if (!axios.isCancel(error)) {
				BtNotify.show({
					type: NotifyDefinedType.FAILED,
					message: error.message
				});
			}

			throw Error('Network error, please try again later');
		};
	}
}
