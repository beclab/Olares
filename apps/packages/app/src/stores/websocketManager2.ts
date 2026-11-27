import { defineStore } from 'pinia';
import { WebSocketStatusEnum } from '@bytetrade/core';
import { getApplication } from 'src/application/base';
import {
	AppMessagePayload,
	BaseWebsocketBean
} from 'src/websocket/applications/base';
import { getWebSocketBean as getApplicationWebSocketBean } from 'src/websocket/applications/beans';
import { getWebSocketBean as getWorkerWebSocketBean } from 'src/websocket/worker/beans';
import {
	traceWebsocket,
	WEBSOCKET_TRACE_TAG
} from 'src/utils/trace/clients/websocketTrace';

export interface WebSocketState {
	worker: SharedWorker | null;
	socketbean: BaseWebsocketBean | null;
	leaderBean: BaseWebsocketBean | null;
	isLeader: boolean;
	leadershipStarted: boolean;
	releaseLock?: () => void;
	leadershipTornDown?: () => void;
	// Periodic liveness ping this tab sends to the worker (every tab, not just
	// the leader) so the worker can prune dead tabs and elect only live ones in
	// worker-mediated mode. Runs for the lifetime of the worker connection.
	clientHeartbeatTimer?: ReturnType<typeof setInterval>;
	// Guards against registering the beforeunload handler more than once, and
	// holds the bound reference so it can be removed on teardown.
	unloadHandlerRegistered?: boolean;
	boundUnloadHandler?: () => void;
	// Monotonic token bumped on every leadership transition (grant attempt,
	// revoke, release). Async leadership setup captures it and aborts if it
	// changed mid-flight, so a stale continuation can never (re)assume a
	// leadership the worker has already moved on from.
	leaderEpoch: number;
	workerWebSocketStatus: WebSocketStatusEnum;
	lastWorkerStatusAt: number;
}

interface ConnectParams {
	url: string;
	login: boolean;
	loginData: any;
	protocols?: string | string[];
	[key: string]: any;
}

function appSkipsSocketEntrance(): boolean {
	const websocketConfig = getApplication().websocketConfig;
	if (!websocketConfig.useShareWorker) {
		return false;
	}
	const bean = getWorkerWebSocketBean(websocketConfig.shareWorkerName);
	return !bean.socketEntrance;
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
			leaderBean: null,
			isLeader: false,
			leadershipStarted: false,
			releaseLock: undefined,
			leadershipTornDown: undefined,
			clientHeartbeatTimer: undefined,
			unloadHandlerRegistered: false,
			boundUnloadHandler: undefined,
			leaderEpoch: 0,
			workerWebSocketStatus: WebSocketStatusEnum.close,
			lastWorkerStatusAt: 0
		} as WebSocketState;
	},

	actions: {
		cleanupSharedWorker(disconnect = false) {
			if (!this.worker) {
				return;
			}

			this.stopClientHeartbeat();

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
			//Do not delete webpackEntryOptions comment, it is used by webpack to name the worker
			this.worker = new SharedWorker(
				/* webpackEntryOptions: { name: "sharedWebsocketWorker" } */
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
				} else if (type === 'do-send') {
					const payload = data?.data !== undefined ? data.data : data;
					const resend = data?.resend ?? false;
					this.leaderBean?.websocket?.send(payload, resend);
				} else if (type === 'do-apply') {
					if (this.isLeader && this.leaderBean) {
						this.leaderBean.otherTypeMethods(data);
					}
				} else if (type === 'do-recreate') {
					if (this.isLeader) {
						const ws = this.leaderBean?.websocket;
						const leaderClosed = !ws || ws.status === WebSocketStatusEnum.close;
						if (event.data.force || leaderClosed) {
							console.log(
								'[ws.manager] recreating leader connection on request from another tab'
							);
							this.createLeaderWebSocket(data);
						}
					}
				} else if (type === 'grant-leader') {
					// Only assume leadership if we can actually open the connection.
					// Otherwise the worker would route sends/do-apply to a leader
					// with no socket. Build params first, then commit. A null return
					// OR a rejection (e.g. getWsLoginData throwing) must both resign,
					// so the worker can re-elect instead of keeping a dead leader.
					const epoch = ++this.leaderEpoch;
					let freshParams: ConnectParams | null = null;
					try {
						freshParams = await this.buildConnectParams();
					} catch (error) {
						console.error(
							'[ws.manager] buildConnectParams failed during leadership grant',
							error
						);
					}
					// If leadership was revoked/released while we were building
					// params, this attempt is stale: abort without opening a socket
					// or publishing, so we never run as a leader the worker dropped.
					if (this.leaderEpoch !== epoch) {
						console.warn(
							'[ws.manager] leadership grant superseded during setup, aborting'
						);
					} else if (freshParams) {
						console.log('[ws.manager] leadership granted by worker');
						this.isLeader = true;
						this.createLeaderWebSocket(freshParams);
						// Signal readiness so the worker flushes work buffered while
						// this tab was setting up as leader.
						this.worker?.port.postMessage({ type: 'leader-ready' });
					} else {
						console.warn(
							'[ws.manager] cannot assume leadership without connect params, resigning'
						);
						this.isLeader = false;
						// Allow a later ensureAlive()/start() to re-claim leadership.
						this.leadershipStarted = false;
						this.worker?.port.postMessage({ type: 'resign-leader' });
					}
				} else if (type === 'leader-revoked') {
					// The worker re-elected another tab (we were presumed dead while
					// still alive, e.g. backgrounded), or invalidated an in-flight
					// grant. Bump the epoch to cancel any pending setup, step down to
					// avoid a second live socket, and allow future re-election.
					console.warn('[ws.manager] leadership revoked by worker');
					this.leaderEpoch++;
					if (this.isLeader) {
						this.isLeader = false;
						this.leaderBean?.websocket?.dispose();
						this.leaderBean = null;
					}
					this.leadershipStarted = false;
				}
			};
		},
		postLeaderStatus(status: WebSocketStatusEnum) {
			this.worker?.port.postMessage({
				type: 'leader-status',
				data: { status }
			});
		},
		postLeaderMessage(message: AppMessagePayload, retain = false) {
			this.worker?.port.postMessage({
				type: 'leader-message',
				data: message,
				retain
			});
		},
		async buildConnectParams(): Promise<ConnectParams | null> {
			const websocketConfig = getApplication().websocketConfig;
			const connectedUrls = getApplication().getWSConnectUrl();
			if (!connectedUrls || connectedUrls.length == 0) {
				console.error('WebSocket URL is empty');
				return null;
			}
			const loginData = await getApplication().getWsLoginData();
			return {
				url: connectedUrls[0],
				loginData,
				login: true,
				...websocketConfig.externalInfo()
			};
		},
		createLeaderWebSocket(params: ConnectParams) {
			const shareWorkerName = getApplication().websocketConfig.shareWorkerName;
			if (!this.leaderBean) {
				this.leaderBean = getWorkerWebSocketBean(shareWorkerName);
			}

			console.log(
				`[ws.manager] createLeaderWebSocket: url=${params.url}, isLeader=${this.isLeader}, bean=${shareWorkerName}`
			);

			this.leaderBean.setMessagePublisher((payload, retain) => {
				this.postLeaderMessage(payload, retain);
			});

			if (this.leaderBean.websocket) {
				this.leaderBean.websocket.dispose();
				this.leaderBean.websocket = null;
			}

			this.leaderBean.initWebSocket(params, () => {
				const status =
					this.leaderBean?.websocket?.status ?? WebSocketStatusEnum.close;
				this.postLeaderStatus(status);
			});

			if (this.leaderBean.websocket) {
				this.postLeaderStatus(WebSocketStatusEnum.load);
			}
		},
		acquireLeadership(params: ConnectParams) {
			const websocketConfig = getApplication().websocketConfig;
			const lockName = `ws-leader-${websocketConfig.shareWorkerName}`;

			if (!navigator.locks) {
				console.warn(
					'[ws.manager] Web Locks API unavailable, using worker-mediated leader election'
				);
				// Do not self-elect: ask the worker to grant leadership to a single
				// tab. We create the WebSocket only when 'grant-leader' arrives,
				// avoiding multiple tabs each opening a connection.
				this.worker?.port.postMessage({ type: 'claim-leader' });
				return;
			}

			void navigator.locks.request(lockName, async () => {
				console.log('[ws.manager] acquired leadership lock');
				this.isLeader = true;
				this.worker?.port.postMessage({ type: 'set-leader' });
				// Install the release resolver synchronously, before any await, so
				// there is never a window where the lock is held but releaseLock is
				// undefined (which would let releaseLeadership() no-op and deadlock).
				const releasePromise = new Promise<void>((resolve) => {
					this.releaseLock = resolve;
				});
				// Rebuild connect params at the moment leadership is acquired so a
				// newly elected leader uses fresh credentials instead of the snapshot
				// captured when start() was first called. Fall back to the captured
				// params if rebuilding fails, so leadership never ends up holding the
				// lock with no socket (which would block re-election and deadlock
				// releaseLeadership()).
				let freshParams: ConnectParams | null = null;
				try {
					freshParams = await this.buildConnectParams();
				} catch (error) {
					console.error(
						'[ws.manager] buildConnectParams failed while acquiring leadership',
						error
					);
				}
				this.createLeaderWebSocket(freshParams ?? params);
				// Tell the worker we can now receive forwarded send/apply work; it
				// buffered anything sent during this setup window and flushes it now.
				this.worker?.port.postMessage({ type: 'leader-ready' });

				await releasePromise;

				console.log('[ws.manager] releasing leadership lock');
				this.leaderBean?.websocket?.dispose();
				this.leaderBean = null;
				this.isLeader = false;
				this.releaseLock = undefined;
				// Signal that teardown has fully completed so callers awaiting
				// releaseLeadership() can safely re-acquire afterwards.
				this.leadershipTornDown?.();
				this.leadershipTornDown = undefined;
			});
		},
		// Releases the held Web Lock and waits until the lock callback has fully
		// torn leadership down (disposed leaderBean, cleared isLeader). Awaiting
		// this before re-acquiring prevents the prior callback from disposing a
		// freshly created connection or clearing isLeader after re-election.
		async releaseLeadership() {
			// Invalidate any in-flight leadership setup before tearing down.
			this.leaderEpoch++;
			if (this.releaseLock) {
				// Web Locks path: resolve the lock wait and await full teardown.
				console.log('[ws.manager] releasing leadership for re-election');
				const tornDown = new Promise<void>((resolve) => {
					this.leadershipTornDown = resolve;
				});
				this.releaseLock();
				await tornDown;
				return;
			}
			// Worker-mediated (claim) path: there is no lock callback to drive
			// teardown, so dispose this tab's leadership directly. Otherwise a
			// recreated worker could elect another tab while this tab keeps
			// isLeader and a live leaderBean socket, yielding duplicate connections
			// and duplicated inbound events.
			if (this.isLeader) {
				console.log(
					'[ws.manager] tearing down worker-mediated leadership for re-election'
				);
				this.isLeader = false;
				this.leaderBean?.websocket?.dispose();
				this.leaderBean = null;
			}
		},
		// Client heartbeat is only used in worker-mediated election mode (no Web
		// Locks). EVERY tab — not just the leader — pings the worker periodically
		// so the worker can tell live tabs from dead ones: a SharedWorker gets no
		// event when a tab closes/crashes, so without this the worker would keep
		// dead ports in its connection set and could elect a dead tab as leader,
		// leaving the app with no working connection. Under Web Locks the OS lock
		// handles liveness, so these pings are simply ignored there.
		startClientHeartbeat() {
			if (this.clientHeartbeatTimer) {
				return;
			}
			// Only needed for worker-mediated election. When the Web Locks API is
			// available (secure contexts, i.e. https/localhost), leadership and
			// liveness are handled by the OS lock and the worker never runs its
			// watchdog, so these heartbeats would be pure overhead with no effect.
			// Skip them there so regular (https) domains are entirely unaffected.
			if (navigator.locks) {
				return;
			}
			// 500ms so a crashed leader is detected fast (worker treats a tab as
			// dead after ~3 missed beats). A foreground tab's timer is not
			// throttled, so healthy leaders never trip this; a backgrounded tab
			// may be throttled and yield leadership to a foreground tab, which is
			// the desired outcome and is handled cleanly via 'leader-revoked'.
			this.clientHeartbeatTimer = setInterval(() => {
				this.worker?.port.postMessage({ type: 'client-heartbeat' });
			}, 500);
		},
		stopClientHeartbeat() {
			if (this.clientHeartbeatTimer) {
				clearInterval(this.clientHeartbeatTimer);
				this.clientHeartbeatTimer = undefined;
			}
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
				await this.releaseLeadership();
				this.cleanupSharedWorker(true);
				this.leadershipStarted = false;
				await this.start();
				return;
			}

			if (this.isClosed() && !appSkipsSocketEntrance()) {
				console.log(
					'[ws.manager] ensureAlive: worker alive but socket closed, restarting'
				);
				await this.start({ skipWorkerAliveCheck: true });
				return;
			}
			console.log('[ws.manager] ensureAlive: worker and socket healthy');
		},
		// Registers a beforeunload handler that disposes on tab close so the
		// worker receives 'disconnect' and can immediately hand leadership to
		// another tab. Idempotent, and also invoked from start() so every app
		// gets it without each having to call appMount() explicitly (previously
		// only some apps did, so their leader tabs never sent 'disconnect' on
		// close and re-election had to wait for the liveness timeout).
		ensureUnloadHandler() {
			if (this.unloadHandlerRegistered) {
				return;
			}
			this.boundUnloadHandler = () => this.dispose();
			window.addEventListener('beforeunload', this.boundUnloadHandler);
			this.unloadHandlerRegistered = true;
		},
		removeUnloadHandler() {
			if (this.boundUnloadHandler) {
				window.removeEventListener('beforeunload', this.boundUnloadHandler);
				this.boundUnloadHandler = undefined;
			}
			this.unloadHandlerRegistered = false;
		},
		appMount() {
			this.ensureUnloadHandler();
		},
		appUnMounted() {
			this.removeUnloadHandler();
		},
		async start(options?: { skipWorkerAliveCheck?: boolean; force?: boolean }) {
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
				const connectParams: ConnectParams = {
					url: connectedUrls[0],
					loginData,
					login: true,
					...websocketConfig.externalInfo()
				};

				if (!this.worker) {
					this.setupSharedWorker(websocketConfig.shareWorkerName);
				} else if (!options?.skipWorkerAliveCheck) {
					const isAlive = await this.checkWorkerAlive();
					if (!isAlive) {
						console.warn('shared websocket worker not responding, recreating');
						await this.releaseLeadership();
						this.cleanupSharedWorker(true);
						this.leadershipStarted = false;
						this.setupSharedWorker(websocketConfig.shareWorkerName);
					}
				}

				this.worker!.port.start();
				this.worker!.port.postMessage({ type: 'register' });
				// Ensure every tab reports liveness and cleans up on close, so the
				// worker can prune dead tabs and elect only live ones (see the
				// worker's watchdog). Both are idempotent.
				this.ensureUnloadHandler();
				this.startClientHeartbeat();

				if (!this.leadershipStarted) {
					this.leadershipStarted = true;
					this.acquireLeadership(connectParams);
				} else if (!appSkipsSocketEntrance()) {
					const force = options?.force ?? false;
					if (force || this.isClosed()) {
						if (this.isLeader) {
							// i18n-ignore-start
							console.log(
								force
									? '[ws.manager] force restart: recreating leader connection with fresh login data'
									: '[ws.manager] leader socket closed, recreating connection'
							);
							// i18n-ignore-end
							this.createLeaderWebSocket(connectParams);
						} else {
							// Follower path: the leader may be backgrounded with a
							// dropped socket, or credentials changed. Ask it to
							// (re)create. With `force` the leader recreates
							// unconditionally; otherwise it only recreates if its own
							// socket is actually closed, so multiple followers polling
							// a closed socket cannot thrash the connection.
							console.log(
								'[ws.manager] asking leader to (re)create connection'
							);
							this.worker!.port.postMessage({
								type: 'recreate-leader',
								data: connectParams,
								force
							});
						}
					}
				}
			} else if (!websocketConfig.useShareWorker) {
				if (this.isConnecting() || this.isConnected()) {
					return;
				}

				if (!this.socketbean) {
					const applicationName = getApplication().applicationName;
					this.socketbean = getApplicationWebSocketBean(applicationName);
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

			this.worker.port.postMessage({ type: 'send', data: { data, resend } });
		},
		restart() {
			this.start({ force: true });
		},

		dispose() {
			if (this.worker) {
				console.log(
					'[ws.manager] dispose shared worker and websocket connection'
				);
				if (this.isLeader) {
					this.leaderBean?.websocket?.dispose();
					this.leaderBean = null;
					this.releaseLock?.();
				}
				this.stopClientHeartbeat();
				this.cleanupSharedWorker(true);
				this.leadershipStarted = false;
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
