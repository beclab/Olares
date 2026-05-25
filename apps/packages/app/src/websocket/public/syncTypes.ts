export enum SyncMessageType {
	SYNC_ENTRIES = 'sync_entries',
	SYNC_FEEDS = 'sync_feeds',
	SYNC_LABELS = 'sync_labels',
	SYNC_NOTES = 'sync_notes',
	SYNC_REMOVE = 'sync_remove'
}

export interface SyncState {
	lastEntrySyncTime: number;
	lastFeedSyncTime: number;
	lastLabelSyncTime: number;
	lastNoteSyncTime: number;
	lastRemoveSyncTime: number;
	isSyncing: boolean;
	currentSyncType: SyncMessageType | null;
}
