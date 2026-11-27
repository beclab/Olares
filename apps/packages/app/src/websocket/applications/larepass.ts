import { BaseWebsocketBean } from './base';
import { WebSocketStatusEnum } from '@bytetrade/core';
import { DownloadRecord } from 'src/utils/interface/rss';
import { TransferDatabase } from 'src/utils/interface/transferDB';
import { wiseInsertTransferItem, WiseWSType } from '../public/wise';
import { busEmit } from 'src/utils/bus';

export class LarePassWebsocketBean extends BaseWebsocketBean {
	// Retry forever (same as desktop/market beans): the reconnect scheduler only
	// gives up when reconnectMaxNum > 0, so -1 means unlimited reconnection.
	reconnectMaxNum = -1;
	private readonly connectTimeoutMs = 30 * 1000;

	private terminusId = '';
	private db = new TransferDatabase();
	private connectWatchdogTimer: ReturnType<typeof setTimeout> | null = null;
	private lastInitData:
		| {
				terminusId?: string;
				url: string;
				login: boolean;
				loginData: any;
				pongRes: (data: any) => boolean;
		  }
		| undefined;
	private lastStatusUpdate: (() => void) | undefined;

	private clearConnectWatchdog() {
		if (this.connectWatchdogTimer) {
			clearTimeout(this.connectWatchdogTimer);
			this.connectWatchdogTimer = null;
		}
	}

	private startConnectWatchdog() {
		// Arm at most once per connect/reconnect cycle. The base bean fires status
		// callbacks (onerror / onreconnect / onReconnectFailure) roughly every
		// reconnectGapTime (~3s) during a reconnect loop; clearing and restarting
		// the timer on each of those would keep pushing the deadline out so the
		// forced dispose/reconnect would never run. Keep any in-flight timer so the
		// 30s window measures from the first non-open signal. A fresh cycle resets
		// it explicitly via clearConnectWatchdog() in initWebSocket.
		if (this.connectWatchdogTimer) {
			return;
		}
		this.connectWatchdogTimer = setTimeout(() => {
			this.connectWatchdogTimer = null;
			if (
				this.websocket &&
				this.websocket.status !== WebSocketStatusEnum.open &&
				this.lastInitData &&
				this.lastStatusUpdate
			) {
				console.warn(
					`larepass websocket not connected within ${this.connectTimeoutMs}ms, force reconnect`
				);
				this.websocket.dispose();
				// Go through the subclass initWebSocket (not super) so the wrapped
				// status handler and the post-init watchdog are re-applied; otherwise
				// a socket that stays non-open after this forced reconnect would never
				// get another 30s reconnect cycle.
				this.initWebSocket(this.lastInitData, this.lastStatusUpdate);
			}
		}, this.connectTimeoutMs);
	}

	initWebSocket(
		data: {
			terminusId?: string;
			url: string;
			login: boolean;
			loginData: any;
			pongRes: (data: any) => boolean;
		},
		statusUpdate: () => void
	): void {
		this.lastInitData = { ...data };
		this.lastStatusUpdate = statusUpdate;
		this.terminusId = data.terminusId || '';
		// Start each (re)connect cycle with a fresh 30s window.
		this.clearConnectWatchdog();
		const wrappedStatusUpdate = () => {
			statusUpdate();
			if (this.websocket?.status === WebSocketStatusEnum.open) {
				this.clearConnectWatchdog();
			} else {
				// During initial connect / reconnect, keep a 30s watchdog. If the
				// socket remains non-open for too long (e.g. hanging CONNECTING),
				// force a fresh reconnect cycle.
				this.startConnectWatchdog();
			}
		};
		super.initWebSocket(data, wrappedStatusUpdate);
		// Ensure watchdog also starts for the first connect attempt even before
		// any status callback from lower layers.
		this.startConnectWatchdog();
	}

	otherTypeMethods(data: { type: string; data: any }) {
		if (data.type === 'addTaskHistory') {
			this.insertTransferItem(data.data);
			return true;
		}
		return false;
	}

	websocketOnMessage(event: MessageEvent): void {
		super.websocketOnMessage(event);
		try {
			const body: any = JSON.parse(event.data);

			if (body.type === WiseWSType.DOWNLOAD_PROCESS) {
				this.insertTransferItem({
					...body.data,
					progress: `${body.data.percent || 0}`,
					id: body.data.task_id
				});

				if (body.type === 'download_process') {
					busEmit('wiseDownloadProcess', body);
				}
			} else {
				// vault message
				busEmit('receiveMessage', body);
			}
		} catch (e) {
			console.error('message error:', e);
		}
	}

	private insertTransferItem(downloadItem: DownloadRecord) {
		wiseInsertTransferItem(downloadItem, this.db, this.terminusId, (item) => {
			busEmit('CloudTransferUpdate', {
				item,
				data: downloadItem
			});
		});
	}
}
