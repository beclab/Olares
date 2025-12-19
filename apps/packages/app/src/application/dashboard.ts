import { NormalApplication } from './base';
import { useAppDetailStore } from '@apps/dashboard/stores/AppDetail';
import { useAppList } from '@apps/dashboard/src/stores/AppList';
import { bus } from '@apps/dashboard/src/utils/bus';
import { useWebsocketManager2Store } from 'src/stores/websocketManager2';
import { WebsocketSharedWorkerEnum } from 'src/websocket/interface';
import { throttle } from 'lodash';

export class DashboardApplication extends NormalApplication {
	applicationName = 'dashboard';
	async appLoadPrepare(data: any): Promise<void> {
		super.appLoadPrepare(data);
		const socketStore = useWebsocketManager2Store();
		socketStore.start();

		bus.on('app_installation_event', () => {
			this.updateApps();
		});
	}

	async appRedirectUrl(): Promise<void> {
		const appDetail = useAppDetailStore();
		const appList = useAppList();

		appList.getAppList();

		return appDetail.init();
	}

	websocketConfig = {
		useShareWorker: true,
		shareWorkerName: WebsocketSharedWorkerEnum.DASHBOARD_NAME,

		externalInfo() {
			return {};
		},
		responseShareWorkerMessage(data: {
			type: 'ws' | 'reconnected';
			data: any;
		}) {
			try {
				const message = JSON.parse(data.data);
				if (message.notify_type == 'app_state_change') {
					bus.emit('app_installation_event', message);
				}
			} catch (e) {
				console.log('message error');
				console.log(e);
			}
		}
	};

	updateApps = throttle(() => {
		const appList = useAppList();
		appList.getAppList();
	}, 3000);

	initAxiosIntercepts(): void {
		super.initAxiosIntercepts();

		this.requestIntercepts.push((config) => {
			if (config.headers) {
				config.headers['X-Unauth-Error'] = 'Non-Redirect';
			}
			return config;
		});

		this.responseIntercepts.push((response) => {
			return response;
		});
	}
}
