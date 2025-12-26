import { BaseWebsocketBean } from '../applications/base';

export class DashboardWebsocketBean extends BaseWebsocketBean {
	websocketOnMessage(ev: MessageEvent): void {
		super.websocketOnMessage(ev);
		this.connections.forEach((port) =>
			port.postMessage({
				type: 'message',
				data: {
					type: 'ws',
					data: ev.data
				}
			})
		);
	}
}
