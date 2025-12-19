import { NormalApplication } from './base';
import { useAppDetailStore } from '@apps/control-hub/stores/AppDetail';
import { useAppList } from '@apps/control-hub/stores/AppList';
import { useMiddlewareStore } from '@apps/control-hub/stores/Middleware';
import { useTerminalStore } from '@apps/control-hub/stores/TerminalStore';
import { Notify } from 'quasar';

export class ControlHubApplication extends NormalApplication {
	applicationName = 'controlHub';
	async appLoadPrepare(data: any): Promise<void> {
		try {
			const middlewareStore = useMiddlewareStore();
			await middlewareStore.getList();
		} catch (error) {
			//
		}
		super.appLoadPrepare(data);
	}

	async appRedirectUrl(): Promise<void> {
		const appDetail = useAppDetailStore();
		const appList = useAppList();
		const terminalStore = useTerminalStore();
		appDetail.init();
		terminalStore.init();

		return appList.init();
	}

	initAxiosIntercepts(): void {
		super.initAxiosIntercepts();
		this.requestIntercepts.push((config) => {
			if (config.headers) {
				config.headers['X-Unauth-Error'] = 'Non-Redirect';
			}
			return config;
		});

		this.responseErrorInterceps = (error: any) => {
			const errorResponse = error.response;
			if (errorResponse?.config.method === 'put') {
				Notify.create({
					type: 'negative',
					caption: `${errorResponse.data.reason} ${errorResponse.data.message} `,
					message: errorResponse.data.status
				});
			}
			throw error;
		};

		this.responseIntercepts.push((response) => {
			return response;
		});
	}
}
