import {
	BaserWebsocketBeanClass,
	BaseWebsocketBean
} from '../applications/base';
import { FilesWebsocketBean } from './files';
import { WiseWebsocketBean } from './wise';
import { MarketWebsocketBean } from './market';
import { SettingsWebsocketBean } from './settings';
import { VaultWebsocketBean } from './vault';
import { DesktopWebsocketBean } from './desktop';
import { DashboardWebsocketBean } from './dashboard';
import { StudioWebsocketBean } from './studio';
import { WebsocketSharedWorkerEnum } from '../interface';

const WebsocketBeanClassRecord: Record<
	WebsocketSharedWorkerEnum,
	BaserWebsocketBeanClass
> = {
	[WebsocketSharedWorkerEnum.WISE_NAME]: WiseWebsocketBean,
	[WebsocketSharedWorkerEnum.FILES_NAME]: FilesWebsocketBean,
	[WebsocketSharedWorkerEnum.MARKET_NAME]: MarketWebsocketBean,
	[WebsocketSharedWorkerEnum.SETTINGS_NAME]: SettingsWebsocketBean,
	[WebsocketSharedWorkerEnum.VAULT_NAME]: VaultWebsocketBean,
	[WebsocketSharedWorkerEnum.DASHBOARD_NAME]: DashboardWebsocketBean,
	[WebsocketSharedWorkerEnum.STUDIO_NAME]: StudioWebsocketBean,
	[WebsocketSharedWorkerEnum.DESKTOP_NAME]: DesktopWebsocketBean
};

const websocketBeansRecord: Record<
	WebsocketSharedWorkerEnum,
	BaseWebsocketBean | null
> = {
	[WebsocketSharedWorkerEnum.WISE_NAME]: null,
	[WebsocketSharedWorkerEnum.FILES_NAME]: null,
	[WebsocketSharedWorkerEnum.MARKET_NAME]: null,
	[WebsocketSharedWorkerEnum.SETTINGS_NAME]: null,
	[WebsocketSharedWorkerEnum.VAULT_NAME]: null,
	[WebsocketSharedWorkerEnum.DASHBOARD_NAME]: null,
	[WebsocketSharedWorkerEnum.STUDIO_NAME]: null,
	[WebsocketSharedWorkerEnum.DESKTOP_NAME]: null
};

export function getWebSocketBean(name: string): BaseWebsocketBean {
	if (websocketBeansRecord[name]) {
		return websocketBeansRecord[name];
	}
	const bean = WebsocketBeanClassRecord[name];
	if (bean) {
		const b = new bean();
		websocketBeansRecord[name] = b;
		return b;
	}
	throw new Error('Unknown websocket type');
}
