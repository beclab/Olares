import { ref, computed } from 'vue';
import { useI18n } from 'vue-i18n';
import { useAppStore } from 'src/stores/market/appStore';

// ============================================================
// Type definitions
// ============================================================
export interface StatusLog {
	logId: string;
	timestamp: number;
	combinedId: string;
	appName: string;
	source: 'api' | 'socket' | 'local';
	oldState: string | null;
	newState: string;
	oldTime: number | null;
	newTime: number;
	timeDiff?: number;
	isIssue: boolean;
	issueType?: string;
	rawData?: any;
	detailText?: string;
	type: 'normal' | 'rejected' | 'regression';
}

export interface AppMonitorInfo {
	combinedId: string;
	appName: string;
	currentState: string;
	updateCount: number;
	hasIssue: boolean;
}

export interface ImportedSnapshotMeta {
	importedAt: number;
	exportTime?: string;
	fileName?: string;
}

export interface ImportSnapshotResult {
	ok: boolean;
	messageKey?: string;
}

// ============================================================
// Configuration
// ============================================================
const MAX_LOGS = 500;
const MAX_APPS = 100;

// ============================================================
// Composable
// ============================================================
export function useStatusMonitor() {
	const appStore = useAppStore();
	const { t } = useI18n();

	// State
	const isRecording = ref(false);
	const logs = ref<StatusLog[]>([]);
	const appInfoMap = ref(new Map<string, AppMonitorInfo>());
	const dataMode = ref<'live' | 'imported'>('live');
	const importedMeta = ref<ImportedSnapshotMeta | null>(null);

	// Original method references
	let originalSetAppStatus: any = null;
	let originalUpdateAppStatusBySocket: any = null;
	let originalUpdateAppStatus: any = null;
	let currentUpdateSource: 'api' | 'socket' | 'local' = 'api';
	let logSeq = 0;

	// ============================================================
	// Computed properties
	// ============================================================
	const sourceStats = computed(() => {
		let api = 0,
			socket = 0,
			local = 0,
			rejected = 0,
			regression = 0;
		logs.value.forEach((log) => {
			if (log.source === 'api') api++;
			else if (log.source === 'socket') socket++;
			else if (log.source === 'local') local++;
			if (log.type === 'rejected') rejected++;
			if (log.type === 'regression') regression++;
		});
		return { api, socket, local, rejected, regression };
	});

	const monitoredApps = computed(() => {
		return Array.from(appInfoMap.value.values()).sort((a, b) => {
			if (a.hasIssue !== b.hasIssue) return a.hasIssue ? -1 : 1;
			return b.updateCount - a.updateCount;
		});
	});

	const hasIssue = computed(() => sourceStats.value.regression > 0);

	// ============================================================
	// Utilities
	// ============================================================
	const getEffectiveTime = (status: any): number => {
		if (!status) return 0;
		const timeFields = [
			status.status?.updateTime,
			status.status?.statusTime,
			status.status?.lastTransitionTime,
			status.updateTime,
			status.timestamp
		];
		for (const field of timeFields) {
			if (field) {
				const time = new Date(field).getTime();
				if (!isNaN(time) && time > 0) return time;
			}
		}
		return 0;
	};

	const formatTime = (ts: number) => {
		const d = new Date(ts);
		return (
			d.toLocaleTimeString('zh-CN', { hour12: false }) +
			'.' +
			String(d.getMilliseconds()).padStart(3, '0')
		);
	};

	const getSourceIcon = (source: string) => {
		if (source === 'api') return '📡';
		if (source === 'socket') return '🔌';
		return '🖱️';
	};

	const getSourceLabel = (source: string) => {
		if (source === 'api') return `📡 ${t('debug.statusMonitor.api')}`;
		if (source === 'socket') return `🔌 ${t('debug.statusMonitor.socket')}`;
		return `🖱️ ${t('debug.statusMonitor.local')}`;
	};

	const safeStringify = (data: unknown): string => {
		if (data === undefined) {
			return t('debug.statusMonitor.noDetailData');
		}
		try {
			const seen = new WeakSet<object>();
			return JSON.stringify(
				data,
				(_key, value) => {
					if (value && typeof value === 'object') {
						const obj = value as object;
						if (seen.has(obj)) {
							return '[Circular]';
						}
						seen.add(obj);
					}
					return value;
				},
				2
			);
		} catch {
			return t('debug.statusMonitor.serializeFailed');
		}
	};

	const normalizeLogSource = (source: unknown): 'api' | 'socket' | 'local' => {
		if (source === 'socket' || source === 'local') {
			return source;
		}
		return 'api';
	};

	const normalizeLogType = (
		type: unknown
	): 'normal' | 'rejected' | 'regression' => {
		if (type === 'rejected' || type === 'regression') {
			return type;
		}
		return 'normal';
	};

	const rebuildAppInfoMap = (items: StatusLog[]) => {
		const next = new Map<string, AppMonitorInfo>();
		for (const item of items) {
			const exist = next.get(item.combinedId);
			if (!exist) {
				next.set(item.combinedId, {
					combinedId: item.combinedId,
					appName: item.appName,
					currentState: item.newState || 'unknown',
					updateCount: 1,
					hasIssue: item.isIssue
				});
			} else {
				exist.currentState = item.newState || 'unknown';
				exist.updateCount += 1;
				if (item.isIssue) {
					exist.hasIssue = true;
				}
			}
		}
		appInfoMap.value = next;
	};

	// ============================================================
	// Log recording
	// ============================================================
	const addLog = (
		combinedId: string,
		source: 'api' | 'socket' | 'local',
		oldStatus: any,
		newStatus: any,
		includeRawData = false
	) => {
		const appName = combinedId.split('_')[1] || combinedId;
		const oldTime = oldStatus ? getEffectiveTime(oldStatus) : null;
		const newTime = getEffectiveTime(newStatus);

		let type: 'normal' | 'rejected' | 'regression' = 'normal';

		// Detect out-of-order timestamp update
		if (
			oldStatus &&
			oldTime &&
			oldTime > 0 &&
			newTime > 0 &&
			oldTime > newTime
		) {
			type = 'rejected';
		}
		// Detect invalid timestamp overriding valid timestamp
		if (oldStatus && oldTime && oldTime > 0 && newTime <= 0) {
			type = 'regression';
		}

		const rawData = includeRawData
			? { old: oldStatus, new: newStatus }
			: {
					oldState: oldStatus?.status?.state ?? null,
					newState: newStatus?.status?.state ?? 'unknown',
					oldTime,
					newTime
			  };

		const log: StatusLog = {
			logId: `${Date.now()}_${++logSeq}_${combinedId}`,
			timestamp: Date.now(),
			combinedId,
			appName,
			source,
			oldState: oldStatus?.status?.state || null,
			newState: newStatus?.status?.state || 'unknown',
			oldTime,
			newTime,
			timeDiff: oldTime !== null ? newTime - oldTime : undefined,
			isIssue: type !== 'normal',
			issueType:
				type === 'rejected'
					? t('debug.statusMonitor.rejected')
					: type === 'regression'
					? `⚠️ ${t('debug.statusMonitor.regression')}`
					: undefined,
			rawData,
			detailText: safeStringify(rawData),
			type
		};

		// Detect state rollback (completed back to in-progress)
		if (
			type === 'normal' &&
			oldStatus?.status?.state &&
			newStatus?.status?.state
		) {
			const progressStates = [
				'installing',
				'downloading',
				'pending',
				'processing'
			];
			const completedStates = ['running', 'completed', 'installed'];
			const oldLower = oldStatus.status.state.toLowerCase();
			const newLower = newStatus.status.state.toLowerCase();

			if (
				completedStates.some((s) => oldLower.includes(s)) &&
				progressStates.some((s) => newLower.includes(s))
			) {
				log.type = 'regression';
				log.isIssue = true;
				log.issueType = `⚠️ ${t('debug.statusMonitor.regression')}`;
			}
		}

		logs.value.push(log);

		// Limit log size
		if (logs.value.length > MAX_LOGS) {
			logs.value = logs.value.slice(-MAX_LOGS);
		}

		// Update app monitor info
		let appInfo = appInfoMap.value.get(combinedId);
		if (!appInfo) {
			if (appInfoMap.value.size >= MAX_APPS) {
				let minApp = '';
				let minCount = Infinity;
				appInfoMap.value.forEach((info, key) => {
					if (info.updateCount < minCount) {
						minCount = info.updateCount;
						minApp = key;
					}
				});
				if (minApp) appInfoMap.value.delete(minApp);
			}
			appInfo = {
				combinedId,
				appName,
				currentState: '',
				updateCount: 0,
				hasIssue: false
			};
			appInfoMap.value.set(combinedId, appInfo);
		}
		appInfo.currentState = newStatus?.status?.state || 'unknown';
		appInfo.updateCount++;
		if (log.isIssue) appInfo.hasIssue = true;
	};

	const normalizeImportedLog = (raw: any, index: number): StatusLog => {
		const combinedId = String(
			raw?.combinedId ?? raw?.appName ?? `unknown_${index}`
		);
		const appName = String(
			raw?.appName ?? combinedId.split('_')[1] ?? combinedId
		);
		const source = normalizeLogSource(raw?.source);
		const type = normalizeLogType(raw?.type);
		const oldState = raw?.oldState ?? null;
		const newState = String(raw?.newState ?? 'unknown');
		const oldTime =
			raw?.oldTime === null || raw?.oldTime === undefined
				? null
				: Number(raw.oldTime);
		const newTime = Number(raw?.newTime ?? 0);
		const timeDiff =
			raw?.timeDiff === undefined || raw?.timeDiff === null
				? oldTime !== null
					? newTime - oldTime
					: undefined
				: Number(raw.timeDiff);
		const isIssue =
			typeof raw?.isIssue === 'boolean' ? raw.isIssue : type !== 'normal';
		const issueType =
			type === 'rejected'
				? t('debug.statusMonitor.rejected')
				: type === 'regression'
				? `⚠️ ${t('debug.statusMonitor.regression')}`
				: undefined;
		const timestamp = Number(raw?.timestamp ?? Date.now());
		const rawData = raw?.rawData;
		const detailText =
			typeof raw?.detailText === 'string' && raw.detailText.length > 0
				? raw.detailText
				: safeStringify(
						rawData ?? {
							oldState,
							newState,
							oldTime,
							newTime
						}
				  );
		return {
			logId: String(raw?.logId ?? `${timestamp}_${index}_${combinedId}`),
			timestamp,
			combinedId,
			appName,
			source,
			oldState,
			newState,
			oldTime,
			newTime,
			timeDiff,
			isIssue,
			issueType,
			rawData,
			detailText,
			type
		};
	};

	const importSnapshot = (
		payload: any,
		options?: { fileName?: string }
	): ImportSnapshotResult => {
		try {
			const rawLogs = Array.isArray(payload)
				? payload
				: Array.isArray(payload?.logs)
				? payload.logs
				: null;
			if (!rawLogs) {
				return {
					ok: false,
					messageKey: 'debug.statusMonitor.importMissingLogsArray'
				};
			}

			if (isRecording.value) {
				stopRecording();
			}

			const normalized = rawLogs.map((item, index) =>
				normalizeImportedLog(item, index)
			);
			logs.value = normalized;
			rebuildAppInfoMap(normalized);
			dataMode.value = 'imported';
			importedMeta.value = {
				importedAt: Date.now(),
				exportTime:
					typeof payload?.exportTime === 'string'
						? payload.exportTime
						: undefined,
				fileName: options?.fileName
			};

			return { ok: true };
		} catch (error: any) {
			console.error('Failed to import status monitor snapshot', error);
			return { ok: false, messageKey: 'debug.statusMonitor.importFailed' };
		}
	};

	// ============================================================
	// Hook methods
	// ============================================================
	const hookSetAppStatus = (includeRawData = false) => {
		originalSetAppStatus = appStore.setAppStatus.bind(appStore);

		(appStore as any).setAppStatus = function (
			combinedId: string,
			appStatus: any
		) {
			const oldStatus = appStore.appStatusMap.get(combinedId);
			addLog(
				combinedId,
				currentUpdateSource,
				oldStatus,
				appStatus,
				includeRawData
			);
			return originalSetAppStatus(combinedId, appStatus);
		};
	};

	const hookSocketUpdate = () => {
		originalUpdateAppStatusBySocket =
			appStore.updateAppStatusBySocket.bind(appStore);

		(appStore as any).updateAppStatusBySocket = function (data: any) {
			currentUpdateSource = 'socket';
			try {
				return originalUpdateAppStatusBySocket(data);
			} finally {
				currentUpdateSource = 'api';
			}
		};
	};

	const hookLocalUpdate = () => {
		originalUpdateAppStatus = appStore.updateAppStatus.bind(appStore);

		(appStore as any).updateAppStatus = function (
			appId: string,
			sourceId: string,
			newState: string,
			opType?: string
		) {
			currentUpdateSource = 'local';
			try {
				return originalUpdateAppStatus(appId, sourceId, newState, opType);
			} finally {
				currentUpdateSource = 'api';
			}
		};
	};

	const restoreOriginalMethods = () => {
		if (originalSetAppStatus) {
			(appStore as any).setAppStatus = originalSetAppStatus;
			originalSetAppStatus = null;
		}
		if (originalUpdateAppStatusBySocket) {
			(appStore as any).updateAppStatusBySocket =
				originalUpdateAppStatusBySocket;
			originalUpdateAppStatusBySocket = null;
		}
		if (originalUpdateAppStatus) {
			(appStore as any).updateAppStatus = originalUpdateAppStatus;
			originalUpdateAppStatus = null;
		}
	};

	// ============================================================
	// Public actions
	// ============================================================
	const startRecording = (includeRawData = false) => {
		if (isRecording.value) return;
		if (dataMode.value === 'imported') {
			logs.value = [];
			appInfoMap.value.clear();
			importedMeta.value = null;
			dataMode.value = 'live';
		}
		hookSetAppStatus(includeRawData);
		hookSocketUpdate();
		hookLocalUpdate();
		isRecording.value = true;
	};

	const stopRecording = () => {
		if (!isRecording.value) return;
		restoreOriginalMethods();
		isRecording.value = false;
	};

	const toggleRecording = (includeRawData = false) => {
		if (isRecording.value) {
			stopRecording();
		} else {
			startRecording(includeRawData);
		}
	};

	const clearLogs = () => {
		logs.value = [];
		appInfoMap.value.clear();
		if (dataMode.value === 'imported') {
			dataMode.value = 'live';
			importedMeta.value = null;
		}
	};

	const exportLogs = () => {
		const data = {
			exportTime: new Date().toISOString(),
			stats: sourceStats.value,
			logs: logs.value
		};
		const blob = new Blob([JSON.stringify(data, null, 2)], {
			type: 'application/json'
		});
		const url = URL.createObjectURL(blob);
		const a = document.createElement('a');
		a.href = url;
		a.download = `status-monitor-${Date.now()}.json`;
		a.click();
		URL.revokeObjectURL(url);
	};

	// Cleanup on component unmount
	const cleanup = () => {
		restoreOriginalMethods();
	};

	return {
		// State
		isRecording,
		logs,
		appInfoMap,
		monitoredApps,
		sourceStats,
		hasIssue,
		dataMode,
		importedMeta,

		// Actions
		startRecording,
		stopRecording,
		toggleRecording,
		clearLogs,
		exportLogs,
		importSnapshot,
		cleanup,

		// Utilities
		formatTime,
		getSourceIcon,
		getSourceLabel,
		getEffectiveTime
	};
}
