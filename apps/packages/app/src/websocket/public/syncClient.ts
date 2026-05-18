import { SyncMessageType, SyncState } from '../public/syncTypes';

/**
 * Generate unique request ID
 */
function generateRequestId(): string {
	return `sync-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
}

/**
 * Generate unique tab ID
 */
function generateTabId(): string {
	return `tab-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
}

const SYNC_CHANNEL_NAME = 'wise-sync-channel';
const SYNC_LOCK_KEY = 'wise-sync-lock';
const SYNC_STATE_KEY = 'wise-sync-state';

/**
 * Sync Client - Uses BroadcastChannel for cross-tab sync coordination
 */
export class SyncClient {
	private channel: BroadcastChannel | null = null;
	private tabId: string;
	private pendingLockRequests = new Map<
		string,
		{
			resolve: (value: any) => void;
			reject: (reason?: any) => void;
			timeout: ReturnType<typeof setTimeout>;
			syncType: SyncMessageType;
		}
	>();
	private connected = false;

	constructor() {
		this.tabId = generateTabId();
	}

	init(): void {
		if (this.connected) return;

		try {
			this.channel = new BroadcastChannel(SYNC_CHANNEL_NAME);
			this.channel.onmessage = (event) => {
				this.handleMessage(event.data);
			};
			this.connected = true;
			console.log('[SyncClient] Initialized with tabId:', this.tabId);
		} catch (error) {
			console.warn('[SyncClient] Failed to create BroadcastChannel:', error);
		}
	}

	/**
	 * Handle messages from BroadcastChannel
	 */
	private handleMessage(message: { type: string; data: any }): void {
		const { type } = message;

		switch (type) {
			case 'sync_lock_released':
				// Another tab released the lock, check if we're waiting
				this.processNextPendingLock();
				break;
		}
	}

	/**
	 * Process next pending lock request
	 */
	private processNextPendingLock(): void {
		if (this.pendingLockRequests.size > 0) {
			const [requestId, pending] = this.pendingLockRequests
				.entries()
				.next().value;
			const syncState = this.getSyncStateFromStorage();

			if (this.tryAcquireLock(pending.syncType)) {
				clearTimeout(pending.timeout);
				this.pendingLockRequests.delete(requestId);
				pending.resolve({
					granted: true,
					syncTime: this.getSyncTimeForType(pending.syncType, syncState),
					syncState
				});
			}
		}
	}

	private getSyncStateFromStorage(): SyncState {
		try {
			const stored = localStorage.getItem(SYNC_STATE_KEY);
			if (stored) {
				return JSON.parse(stored);
			}
		} catch (e) {
			// Ignore parse errors
		}
		return {
			lastEntrySyncTime: 0,
			lastFeedSyncTime: 0,
			lastLabelSyncTime: 0,
			lastNoteSyncTime: 0,
			lastRemoveSyncTime: 0,
			isSyncing: false,
			currentSyncType: null
		};
	}

	private saveSyncStateToStorage(state: Partial<SyncState>): void {
		try {
			const current = this.getSyncStateFromStorage();
			const updated = { ...current, ...state };
			localStorage.setItem(SYNC_STATE_KEY, JSON.stringify(updated));
		} catch (e) {
			console.warn('[SyncClient] Failed to save sync state:', e);
		}
	}

	private getSyncTimeForType(
		syncType: SyncMessageType,
		state: SyncState
	): number {
		switch (syncType) {
			case SyncMessageType.SYNC_ENTRIES:
				return state.lastEntrySyncTime;
			case SyncMessageType.SYNC_FEEDS:
				return state.lastFeedSyncTime;
			case SyncMessageType.SYNC_LABELS:
				return state.lastLabelSyncTime;
			case SyncMessageType.SYNC_NOTES:
				return state.lastNoteSyncTime;
			case SyncMessageType.SYNC_REMOVE:
				return state.lastRemoveSyncTime;
			default:
				return 0;
		}
	}

	/**
	 * Try to acquire sync lock
	 */
	private tryAcquireLock(syncType: SyncMessageType): boolean {
		try {
			const lockData = localStorage.getItem(SYNC_LOCK_KEY);
			if (lockData) {
				const lock = JSON.parse(lockData);
				// Check if lock is expired (30 seconds timeout)
				if (Date.now() - lock.timestamp < 30000) {
					return false;
				}
			}

			localStorage.setItem(
				SYNC_LOCK_KEY,
				JSON.stringify({
					tabId: this.tabId,
					syncType,
					timestamp: Date.now()
				})
			);
			return true;
		} catch (e) {
			console.warn('[SyncClient] Failed to acquire lock:', e);
			return false;
		}
	}

	/**
	 * Release sync lock
	 */
	private releaseLock(): void {
		try {
			const lockData = localStorage.getItem(SYNC_LOCK_KEY);
			if (lockData) {
				const lock = JSON.parse(lockData);
				if (lock.tabId === this.tabId) {
					localStorage.removeItem(SYNC_LOCK_KEY);

					this.channel?.postMessage({
						type: 'sync_lock_released',
						data: { tabId: this.tabId }
					});
				}
			}
		} catch (e) {
			console.warn('[SyncClient] Failed to release lock:', e);
		}
	}

	async initSync(initialState?: Partial<SyncState>): Promise<SyncState> {
		this.init();

		const currentState = this.getSyncStateFromStorage();

		if (initialState) {
			const mergedState: SyncState = {
				lastEntrySyncTime:
					currentState.lastEntrySyncTime || initialState.lastEntrySyncTime || 0,
				lastFeedSyncTime:
					currentState.lastFeedSyncTime || initialState.lastFeedSyncTime || 0,
				lastLabelSyncTime:
					currentState.lastLabelSyncTime || initialState.lastLabelSyncTime || 0,
				lastNoteSyncTime:
					currentState.lastNoteSyncTime || initialState.lastNoteSyncTime || 0,
				lastRemoveSyncTime:
					currentState.lastRemoveSyncTime ||
					initialState.lastRemoveSyncTime ||
					0,
				isSyncing: false,
				currentSyncType: null
			};
			this.saveSyncStateToStorage(mergedState);
			return mergedState;
		}

		return currentState;
	}

	/**
	 * Request sync lock
	 */
	async requestSyncLock(syncType: SyncMessageType): Promise<{
		granted: boolean;
		syncTime: number;
		syncState: SyncState;
	} | null> {
		this.init();

		const syncState = this.getSyncStateFromStorage();

		if (this.tryAcquireLock(syncType)) {
			return {
				granted: true,
				syncTime: this.getSyncTimeForType(syncType, syncState),
				syncState
			};
		}

		return new Promise((resolve) => {
			const requestId = generateRequestId();
			const timeout = setTimeout(() => {
				this.pendingLockRequests.delete(requestId);
				resolve(null);
			}, 5000);

			this.pendingLockRequests.set(requestId, {
				resolve,
				reject: () => resolve(null),
				timeout,
				syncType
			});
		});
	}

	async updateSyncTime(
		type: 'entry' | 'feed' | 'label' | 'note' | 'remove',
		time: number
	): Promise<SyncState> {
		const state = this.getSyncStateFromStorage();

		switch (type) {
			case 'entry':
				if (time > state.lastEntrySyncTime) {
					state.lastEntrySyncTime = time;
				}
				break;
			case 'feed':
				if (time > state.lastFeedSyncTime) {
					state.lastFeedSyncTime = time;
				}
				break;
			case 'label':
				if (time > state.lastLabelSyncTime) {
					state.lastLabelSyncTime = time;
				}
				break;
			case 'note':
				if (time > state.lastNoteSyncTime) {
					state.lastNoteSyncTime = time;
				}
				break;
			case 'remove':
				if (time > state.lastRemoveSyncTime) {
					state.lastRemoveSyncTime = time;
				}
				break;
		}

		this.saveSyncStateToStorage(state);
		return state;
	}

	syncCompleted(
		syncType: SyncMessageType,
		success: boolean,
		error?: string
	): void {
		this.releaseLock();

		this.channel?.postMessage({
			type: 'sync_completed',
			data: {
				tabId: this.tabId,
				syncType,
				success,
				error
			}
		});
	}

	async resetSync(): Promise<SyncState> {
		const emptyState: SyncState = {
			lastEntrySyncTime: 0,
			lastFeedSyncTime: 0,
			lastLabelSyncTime: 0,
			lastNoteSyncTime: 0,
			lastRemoveSyncTime: 0,
			isSyncing: false,
			currentSyncType: null
		};
		this.saveSyncStateToStorage(emptyState);
		localStorage.removeItem(SYNC_LOCK_KEY);
		return emptyState;
	}

	isConnected(): boolean {
		return this.connected;
	}
}

let syncClientInstance: SyncClient | null = null;

export function getSyncClient(): SyncClient {
	if (!syncClientInstance) {
		syncClientInstance = new SyncClient();
	}
	return syncClientInstance;
}
