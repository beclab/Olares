import { BaserWebsocketBeanClass, BaseWebsocketBean } from './base';

import { LarePassWebsocketBean } from './larepass';
// import { VaultWebsocketBean } from './vault';
// import { DashboardWebsocketBean } from './dashboard';
// import { MarketWebsocketBean } from './market';
// import { DesktopWebsocketBean } from './desktop';
// import { SettingsWebsocketBean } from './settings';
import { WebsocketApplicationEnum } from '../interface';

const WebsocketBeanClassRecord: Record<
	WebsocketApplicationEnum,
	BaserWebsocketBeanClass
> = {
	[WebsocketApplicationEnum.LAREPASS]: LarePassWebsocketBean,
	[WebsocketApplicationEnum.LarePass_WISE]: LarePassWebsocketBean
	// [WebsocketApplicationEnum.VAULT]: VaultWebsocketBean,
	// [WebsocketApplicationEnum.DASHBOARD]: DashboardWebsocketBean,
	// [WebsocketApplicationEnum.MARKET]: MarketWebsocketBean,
	// [WebsocketApplicationEnum.DESKTOP]: DesktopWebsocketBean
	// [WebsocketApplicationEnum.SETTINGS]: SettingsWebsocketBean
};

const websocketBeansRecord: Record<
	WebsocketApplicationEnum,
	BaseWebsocketBean | null
> = {
	[WebsocketApplicationEnum.LAREPASS]: null,
	[WebsocketApplicationEnum.LarePass_WISE]: null
	// [WebsocketApplicationEnum.VAULT]: null,
	// [WebsocketApplicationEnum.DASHBOARD]: null,
	// [WebsocketApplicationEnum.MARKET]: null,
	// [WebsocketApplicationEnum.DESKTOP]: null
	// [WebsocketApplicationEnum.SETTINGS]: null
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
