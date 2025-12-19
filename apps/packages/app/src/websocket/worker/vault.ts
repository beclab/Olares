import { BaseWebsocketBean } from '../applications/base';
// import { busEmit } from 'src/utils/bus';
export class VaultWebsocketBean extends BaseWebsocketBean {
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
