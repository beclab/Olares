import { BaseWebsocketBean } from '../applications/base';

export class DashboardWebsocketBean extends BaseWebsocketBean {
	websocketOnMessage(ev: MessageEvent): void {
		super.websocketOnMessage(ev);
		this.publishAppMessage({
			type: 'ws',
			data: ev.data
		});
	}
}
