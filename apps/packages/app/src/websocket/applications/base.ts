import { WebSocketBean, WebSocketStatusEnum } from '@bytetrade/core';

export class BaseWebsocketBean {
	websocket: WebSocketBean | null = null;
	token: string;
	connections = new Set<MessagePort>();

	reconnectMaxNum = 5;
	reconnectGapTime = 3000;
	heartFailNum = 10;

	initWebSocket(
		data: {
			url: string;
			login: boolean;
			loginData: {
				token: string;
			};
		},
		statusUpdate: () => void
	): void {
		if (this.websocket) {
			if (
				this.websocket.status === WebSocketStatusEnum.open &&
				data.loginData.token == this.token
			) {
				return;
			}
			this.websocket.dispose();
		}
		this.token = data.loginData.token;

		this.websocket = new WebSocketBean({
			url: data.url,
			needReconnect: true,
			heartSend: JSON.stringify({
				event: 'ping'
			}),
			heartGet: JSON.stringify({
				event: 'pong'
			}),
			heartRes: this.getWsPongRes,
			heartFailNum: this.heartFailNum,
			reconnectMaxNum: this.reconnectMaxNum,
			reconnectGapTime: this.reconnectGapTime,
			onopen: async () => {
				if (data.login) {
					this.websocket?.send({
						event: 'login',
						data: data.loginData
					});
				}
				statusUpdate();
			},
			onmessage: (event) => {
				// console.log('Message received:', event.data);
				this.websocketOnMessage(event);
			},
			onerror: () => {
				console.log('socket error');
				statusUpdate();
			},
			onreconnect: () => {
				console.log('socket start reconnect');
				statusUpdate();
			},
			onReconnectFailure: async () => {
				console.log('socket fail reconnect');
				statusUpdate();
			},
			onReconnectSuccess: async () => {
				this.onReconnectSuccess();
			}
		});
		this.websocket.start();
		statusUpdate();
	}

	otherTypeMethods(_data: { type: string; data: any }): boolean {
		return false;
	}

	websocketOnMessage(event: MessageEvent): void {
		console.log('log event data ===>', event.data);
	}

	async onReconnectSuccess() {
		console.log('socket success reconnect');
	}

	getWsPongRes(data: any) {
		if (typeof data == 'string') {
			return JSON.parse(data).event === 'pong';
		}
		if (typeof data == 'object') {
			return data.event == 'pong';
		}
		return false;
	}
}

export type BaserWebsocketBeanClass = new () => BaseWebsocketBean;
