import { NormalApplication } from './base';
import { useAppDetailStore } from '@apps/control-hub/stores/AppDetail';
import { useAppList } from '@apps/control-hub/stores/AppList';
import { useMiddlewareStore } from '@apps/control-hub/stores/Middleware';
import { useTerminalStore } from '@apps/control-hub/stores/TerminalStore';
import { BtNotify, NotifyDefinedType } from '@bytetrade/ui';
import { i18n } from '../boot/control-hub-i18n';

export class ControlHubApplication extends NormalApplication {
	applicationName = 'controlHub';
	async appLoadPrepare(data: any): Promise<void> {
		try {
			const middlewareStore = useMiddlewareStore();
			await middlewareStore.getList();
		} catch (error) {
			//
		}
		super.appLoadPrepare({ ...data, i18n });
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
				const d = errorResponse.data || {};
				const msg = [d.reason, d.message, d.status]
					.filter(Boolean)
					.join(' ')
					.trim();
				BtNotify.show({
					type: NotifyDefinedType.FAILED,
					message: msg || 'Request failed'
				});
			}
			throw error;
		};

		this.responseIntercepts.push((response) => {
			return response;
		});
	}
}
