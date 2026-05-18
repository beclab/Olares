import { BaseWebsocketBean } from '../applications/base';
export class SettingsWebsocketBean extends BaseWebsocketBean {
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

	async onReconnectSuccess() {
		super.onReconnectSuccess();
		this.connections.forEach((port) => {
			port.postMessage({
				type: 'message',
				data: {
					type: 'reconnected'
				}
			});
		});
	}

	otherTypeMethods(data: {
		type: string;
		data: any;
		_port?: MessagePort;
	}): boolean {
		switch (data.type) {
			case 'usersUpdate':
				this.websocket?.send({
					event: 'message',
					data: {
						...data.data,
						eventType: 'usersUpdate'
					}
				});
				return true;

			default:
				return false;
		}
	}
}
