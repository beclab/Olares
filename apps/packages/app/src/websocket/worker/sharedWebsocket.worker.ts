import { WebSocketStatusEnum } from '@bytetrade/core';
import { getWebSocketBean } from './beans';

declare const self: SharedWorkerGlobalScope;

self.onconnect = (event) => {
	const port = event.ports[0];
	const name = (event.currentTarget as any).name;
	const sharedWebsocket = getWebSocketBean(name);
	sharedWebsocket.connections.add(port);

	port.onmessage = (messageEvent) => {
		const { type, data } = messageEvent.data;

		switch (type) {
			case 'connect':
				sharedWebsocket.initWebSocket(data, () => {
					sharedWebsocket.connections.forEach((port) =>
						port.postMessage({
							type: 'status',
							data: {
								status: sharedWebsocket.websocket?.status
							}
						})
					);
				});
				break;
			case 'send':
				if (
					sharedWebsocket &&
					sharedWebsocket.websocket &&
					sharedWebsocket.websocket &&
					sharedWebsocket.websocket.status === WebSocketStatusEnum.open
				) {
					sharedWebsocket.websocket.send(data.data, data.resend);
				}
				break;
			case 'disconnect':
				// sharedWebsocket.connections = sharedWebsocket.connections.filter(
				// 	(conn) => conn !== port
				// );
				sharedWebsocket.connections.delete(port);
				if (sharedWebsocket.connections.size === 0) {
					if (sharedWebsocket.websocket) {
						sharedWebsocket.websocket!.dispose();
					}
					sharedWebsocket.websocket = null;
				}
				break;
			case 'dispose':
				sharedWebsocket.connections.clear();
				if (sharedWebsocket.websocket) {
					sharedWebsocket.websocket!.dispose();
				}
				sharedWebsocket.websocket = null;
				break;

			case 'status':
				port.postMessage({
					type: 'status',
					data: {
						status: sharedWebsocket.websocket?.status
					}
				});
				break;
			default:
				sharedWebsocket.otherTypeMethods(messageEvent.data);
		}
	};
};
