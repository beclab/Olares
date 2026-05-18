import { defineStore } from 'pinia';
import { WebSocketStatusEnum } from '@bytetrade/core';
import { getApplication } from 'src/application/base';
import { BaseWebsocketBean } from 'src/websocket/applications/base';
import { getWebSocketBean } from 'src/websocket/applications/beans';
import {
	traceWebsocket,
	WEBSOCKET_TRACE_TAG
} from 'src/utils/trace/clients/websocketTrace';

export interface WebSocketState {
	worker: SharedWorker | null;
	socketbean: BaseWebsocketBean | null;
	workerWebSocketStatus: WebSocketStatusEnum;
	lastWorkerStatusAt: number;
}

function getWebsocketTraceSource() {
	return {
		application: getApplication().applicationName,
		context: window.top === window ? 'top-window' : 'iframe',
		windowName: window.name || 'unnamed',
		path: window.location.pathname
	};
}

export const useWebsocketManager2Store = defineStore('websocketManager2', {
	state: () => {
		return {
			worker: null,
			socketbean: null,
			workerWebSocketStatus: WebSocketStatusEnum.close,
			lastWorkerStatusAt: 0
		} as WebSocketState;
	},

	actions: {
		cleanupSharedWorker(disconnect = false) {
			if (!this.worker) {
				return;
			}

			if (disconnect) {
				this.worker.port.postMessage({ type: 'disconnect' });
			}

			this.worker.port.close();
			this.worker = null;
			this.workerWebSocketStatus = WebSocketStatusEnum.close;
			this.lastWorkerStatusAt = 0;
		},
		setupSharedWorker(shareWorkerName: string) {
			if (this.worker) {
				console.log(
					'[ws.manager] setupSharedWorker skipped: worker already exists'
				);
				return;
			}
			console.log(`[ws.manager] create SharedWorker: name=${shareWorkerName}`);
			this.worker = new SharedWorker(
				new URL(
					'../websocket/worker/sharedWebsocket.worker.ts',
					import.meta.url
				),
				{
					name: shareWorkerName,
					type: 'module'
				}
			);

			this.worker.port.onmessage = async (event) => {
				const { type, data } = event.data;

				const messageRes =
					getApplication().websocketConfig.responseShareWorkerMessage;

				if (getApplication().websocketConfig.console) {
					console.log('worker on message --->', event.data);
				}

				traceWebsocket(WEBSOCKET_TRACE_TAG.WORKER_MESSAGE, {
					source: getWebsocketTraceSource(),
					message: event.data
				});

				if (type === 'status') {
					this.workerWebSocketStatus = data.status;
					this.lastWorkerStatusAt = Date.now();
					console.log(
						`[ws.manager] worker status update: status=${data.status}, at=${this.lastWorkerStatusAt}`
					);
				} else if (type == 'message' && messageRes) {
					messageRes(data);
				}
			};
		},
		async checkWorkerAlive(timeoutMs = 3000) {
			if (!this.worker) {
				console.log('[ws.manager] checkWorkerAlive: worker is null');
				return false;
			}
			console.log(`[ws.manager] checkWorkerAlive: timeout=${timeoutMs}ms`);
			const before = this.lastWorkerStatusAt;
			try {
				this.worker.port.postMessage({ type: 'status' });
			} catch (error) {
				console.error('failed to query worker status', error);
				return false;
			}

			return await new Promise<boolean>((resolve) => {
				const startedAt = Date.now();
				const timer = window.setInterval(() => {
					if (this.lastWorkerStatusAt > before) {
						window.clearInterval(timer);
						console.log('[ws.manager] checkWorkerAlive result: alive');
						resolve(true);
						return;
					}
					if (Date.now() - startedAt >= timeoutMs) {
						window.clearInterval(timer);
						console.warn('[ws.manager] checkWorkerAlive result: timeout');
						resolve(false);
					}
				}, 100);
			});
		},
		async ensureAlive() {
			const websocketConfig = getApplication().websocketConfig;
			console.log(
				`[ws.manager] ensureAlive start: useShareWorker=${
					websocketConfig.useShareWorker
				}, hasWorker=${!!this.worker}, workerStatus=${
					this.workerWebSocketStatus
				}`
			);
			if (!websocketConfig.useShareWorker) {
				console.log('[ws.manager] ensureAlive: non-SharedWorker app, skipping');
				return;
			}

			if (!this.worker) {
				console.log('[ws.manager] ensureAlive: worker missing, restarting');
				await this.start();
				return;
			}

			const isAlive = await this.checkWorkerAlive();
			if (!isAlive) {
				console.log('shared websocket worker not responding, recreating');
				this.cleanupSharedWorker();
				await this.start();
				return;
			}

			if (this.isClosed()) {
				console.log(
					'[ws.manager] ensureAlive: worker alive but socket closed, restarting'
				);
				await this.start({ skipWorkerAliveCheck: true });
				return;
			}
			console.log('[ws.manager] ensureAlive: worker and socket healthy');
		},
		appMount() {
			window.addEventListener('beforeunload', () => this.dispose());
		},
		appUnMounted() {
			window.removeEventListener('beforeunload', () => this.dispose());
		},
		async start(options?: { skipWorkerAliveCheck?: boolean }) {
			const websocketConfig = getApplication().websocketConfig;
			const connectedUrls = getApplication().getWSConnectUrl();
			if (!connectedUrls || connectedUrls.length == 0) {
				console.error('WebSocket URL is empty');
				return;
			}

			const loginData = await getApplication().getWsLoginData();
			console.log(
				`[ws.manager] start called: useShareWorker=${websocketConfig.useShareWorker}, url=${connectedUrls[0]}`
			);

			if (websocketConfig.useShareWorker) {
				console.log('WebSocket URL init');
				if (!this.worker) {
					this.setupSharedWorker(websocketConfig.shareWorkerName);
					console.log('WebSocket URL start');
				} else {
					if (!options?.skipWorkerAliveCheck) {
						const isAlive = await this.checkWorkerAlive();
						if (!isAlive) {
							console.warn(
								'shared websocket worker not responding, recreating'
							);
							this.cleanupSharedWorker();
							this.setupSharedWorker(websocketConfig.shareWorkerName);
						}
					} else {
						console.log(
							'[ws.manager] skip checkWorkerAlive: worker already verified alive'
						);
					}
				}

				const externalInfo = websocketConfig.externalInfo();
				console.log('[ws.manager] post connect to shared worker');
				this.worker?.port.postMessage({
					type: 'connect',
					data: {
						url: connectedUrls[0],
						loginData,
						login: true,
						...externalInfo
					}
				});
			} else if (!websocketConfig.useShareWorker) {
				if (this.isConnecting() || this.isConnected()) {
					return;
				}

				if (!this.socketbean) {
					const applicationName = getApplication().applicationName;
					this.socketbean = getWebSocketBean(applicationName);
				}

				const externalInfo = websocketConfig.externalInfo();
				this.socketbean.initWebSocket(
					{
						url: connectedUrls[0],
						login: true,
						loginData,
						...externalInfo
					},
					() => {}
				);
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
			return this.workerWebSocketStatus === WebSocketStatusEnum.close;
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
				console.log(
					'[ws.manager] dispose shared worker and websocket connection'
				);
				this.cleanupSharedWorker(true);
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
