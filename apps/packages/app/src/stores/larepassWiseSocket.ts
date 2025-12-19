import { defineStore } from 'pinia';
import { WebSocketStatusEnum } from '@bytetrade/core';
import { getApplication } from 'src/application/base';
import { BaseWebsocketBean } from 'src/websocket/applications/base';
import { getWebSocketBean } from 'src/websocket/applications/beans';
import { useAppAbilitiesStore } from './appAbilities';
import { useUserStore } from './user';
import { WebsocketApplicationEnum } from 'src/websocket/interface';

export interface WebSocketState {
	socketbean: BaseWebsocketBean | null;
	workerWebSocketStatus: WebSocketStatusEnum;
}

export const useLarePassWiseSocket = defineStore('larepassWiseSocket', {
	state: () => {
		return {
			worker: null,
			socketbean: null,
			workerWebSocketStatus: WebSocketStatusEnum.close
		} as WebSocketState;
	},

	actions: {
		async start() {
			const abilityStore = useAppAbilitiesStore();
			if (!abilityStore.wise.running || !abilityStore.wise.id) {
				return;
			}

			if (this.isConnecting() || this.isConnected()) {
				return;
			}
			const loginData = await getApplication().getWsLoginData();
			const userStore = useUserStore();
			const url = userStore.getModuleSever(abilityStore.wise.id, 'wss:', '/ws');
			const websockConfig = getApplication().websocketConfig;
			if (!this.socketbean) {
				const applicationName = WebsocketApplicationEnum.LarePass_WISE;
				const bean = getWebSocketBean(applicationName);
				this.socketbean = bean;
				const externalInfo = websockConfig.externalInfo();
				this.socketbean.initWebSocket(
					{
						url: url,
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
			if (!this.socketbean?.websocket) {
				return false;
			}
			return this.socketbean.websocket?.status == WebSocketStatusEnum.close;
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
				this.socketbean = null;
			}
		},

		apply(type: string, data: any) {
			if (this.socketbean)
				this.socketbean.otherTypeMethods({
					type,
					data
				});
		}
	}
});
