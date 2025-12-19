import { defineStore } from 'pinia';
import { WebSocketStatusEnum } from '@bytetrade/core';
import { getApplication } from 'src/application/base';
import { BaseWebsocketBean } from 'src/websocket/applications/base';
import { getWebSocketBean } from 'src/websocket/applications/beans';

export interface WebSocketState {
	worker: SharedWorker | null;
	socketbean: BaseWebsocketBean | null;
	workerWebSocketStatus: WebSocketStatusEnum;
}

export const useWebsocketManager2Store = defineStore('websocketManager2', {
	state: () => {
		return {
			worker: null,
			socketbean: null,
			workerWebSocketStatus: WebSocketStatusEnum.close
		} as WebSocketState;
	},

	actions: {
		appMount() {
			window.addEventListener('beforeunload', () => this.dispose());
		},
		appUnMounted() {
			window.removeEventListener('beforeunload', () => this.dispose());
		},
		async start() {
			const websockConfig = getApplication().websocketConfig;
			const connectedUrls = getApplication().getWSConnectUrl();
			if (!connectedUrls || connectedUrls.length == 0) {
				console.error('WebSocket URL is empty');
				return;
			}

			const loginData = await getApplication().getWsLoginData();

			if (websockConfig.useShareWorker && !this.worker) {
				console.log('WebSocket URL init');
				this.worker = new SharedWorker(
					new URL(
						'../websocket/worker/sharedWebsocket.worker.ts',
						import.meta.url
					),
					{
						name: websockConfig.shareWorkerName,
						type: 'module'
					}
				);
				console.log('WebSocket URL start');

				this.worker.port.onmessage = async (event) => {
					const { type, data } = event.data;

					const messageRes =
						getApplication().websocketConfig.responseShareWorkerMessage;

					if (type === 'status') {
						this.workerWebSocketStatus = data.status;
					} else if (type == 'message' && messageRes) {
						messageRes(data);
					}
				};
				const externalInfo = websockConfig.externalInfo();
				this.worker?.port.postMessage({
					type: 'connect',
					data: {
						url: connectedUrls[0],
						loginData,
						login: true,
						...externalInfo
					}
				});
			} else if (!websockConfig.useShareWorker) {
				if (this.isConnecting() || this.isConnected()) {
					console.log(
						'socket Starting..., socket status' +
							this.socketbean?.websocket?.status
					);
					return;
				}

				if (!this.socketbean) {
					const applicationName = getApplication().applicationName;
					const bean = getWebSocketBean(applicationName);
					this.socketbean = bean;
					const externalInfo = websockConfig.externalInfo();
					this.socketbean.initWebSocket(
						{
							url: connectedUrls[0],
							login: true,
							loginData,
							...externalInfo
						},
						() => {}
					);
				} else {
					this.socketbean.websocket?.start();
				}
			}
		},
		isConnecting() {
			if (!this.worker) {
				if (!this.socketbean?.websocket) {
					return false;
				}
				return this.socketbean?.websocket.status == WebSocketStatusEnum.load;
			}
			return this.workerWebSocketStatus === WebSocketStatusEnum.load;
		},
		isConnected() {
			if (!this.worker) {
				// return false;
				if (!this.socketbean?.websocket) {
					return false;
				}
				return this.socketbean?.websocket.status == WebSocketStatusEnum.open;
			}

			return this.workerWebSocketStatus == WebSocketStatusEnum.open;
		},
		isClosed() {
			if (!this.worker) {
				if (!this.socketbean) {
					return true;
				}
				return this.socketbean.websocket?.status == WebSocketStatusEnum.close;
			}
		},
		send(data: any, resend = false) {
			if (!this.worker) {
				if (!this.socketbean?.websocket) {
					return;
				}
				return this.socketbean?.websocket!.send(data, resend);
			}

			this.worker.port.postMessage({ type: 'send', data });
		},
		restart() {
			this.start();
		},

		dispose() {
			if (this.worker) {
				this.worker.port.postMessage({ type: 'disconnect' });
				window.removeEventListener('beforeunload', () => this.dispose());
			} else {
				if (this.socketbean && this.socketbean?.websocket) {
					this.socketbean.websocket?.dispose();
					this.socketbean.websocket = null;
					this.socketbean = null;
				}
			}
		},

		apply(type: string, data: any) {
			if (this.worker) {
				this.worker.port.postMessage({
					type,
					data
				});
			} else {
				if (this.socketbean)
					this.socketbean.otherTypeMethods({
						type,
						data
					});
			}
		}
	}
});
