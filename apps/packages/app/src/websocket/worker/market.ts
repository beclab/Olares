import { BaseWebsocketBean } from '../applications/base';

export class MarketWebsocketBean extends BaseWebsocketBean {
	reconnectMaxNum = -1;
	reconnectGapTime = 3000;
	heartFailNum = 1;

	websocketOnMessage(ev: MessageEvent): void {
		super.websocketOnMessage(ev);
		this.publishAppMessage({
			type: 'ws',
			data: ev.data
		});
	}

	async onReconnectSuccess() {
		super.onReconnectSuccess();
		this.publishAppMessage({ type: 'reconnected' });
	}
}
