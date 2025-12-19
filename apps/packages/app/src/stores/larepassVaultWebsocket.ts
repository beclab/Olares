import { defineStore } from 'pinia';
import { getApplication } from 'src/application/base';
import { WebSocketStatusEnum } from '@bytetrade/core';
import { BaseWebsocketBean } from 'src/websocket/applications/base';
import { getWebSocketBean } from 'src/websocket/applications/beans';
export interface WebSocketState {
	socketbean: BaseWebsocketBean | null;
}

export const useVaultSocketStore = defineStore('vaultWebSocket', {
	state: () => {
		return {
			connectedUrl: '',
			socketbean: null
		} as WebSocketState;
	},

	actions: {
		async start() {
			const websockConfig = getApplication().websocketConfig;
			const connectedUrl = getApplication().getWSConnectUrl();
			if (!connectedUrl || connectedUrl.length == 1) {
				// console.error('WebSocket URL is empty');
				return;
			}
			const loginData = await getApplication().getWsLoginData();
			if (this.isConnecting() || this.isConnected()) {
				console.log(
					'socket Starting..., socket status' +
						this.socketbean?.websocket?.status
				);
				return;
			}

			if (!this.socketbean) {
				const applicationName = 'vault';
				const bean = getWebSocketBean(applicationName);
				this.socketbean = bean;
				const externalInfo = websockConfig.externalInfo();
				this.socketbean.initWebSocket(
					{
						url: connectedUrl[1],
						login: true,
						loginData,
						...externalInfo
					},
					() => {}
				);
			} else {
				this.socketbean.websocket?.start();
			}
		},
		isConnecting() {
			if (!this.socketbean?.websocket) {
				return false;
			}
			return this.socketbean?.websocket.status == WebSocketStatusEnum.load;
		},
		isConnected() {
			if (!this.socketbean?.websocket) {
				return false;
			}
			return this.socketbean?.websocket.status == WebSocketStatusEnum.open;
		},
		isClosed() {
			if (!this.socketbean) {
				return true;
			}
		},
		send(data: any, resend = false) {
			if (!this.socketbean?.websocket) {
				return;
			}
			return this.socketbean?.websocket!.send(data, resend);
		},
		restart() {
			this.start();
		},

		dispose() {
			if (this.socketbean && this.socketbean?.websocket) {
				this.socketbean.websocket?.dispose();
				this.socketbean.websocket = null;
			}
		}
	}
});
