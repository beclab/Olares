<template>
	<div class="status-monitor">
		<header class="toolbar">
			<div class="title">📡 {{ t('debug.statusMonitor.title') }}</div>
			<div class="controls">
				<button
					:class="['btn', { active: isRecording }]"
					@click="toggleRecording"
				>
					{{
						isRecording
							? `⏹️ ${t('debug.statusMonitor.stop')}`
							: `▶️ ${t('debug.statusMonitor.start')}`
					}}
				</button>
				<button class="btn" @click="openImportPicker">
					📤 {{ t('debug.statusMonitor.import') }}
				</button>
				<input
					ref="importInputRef"
					type="file"
					accept=".json,application/json"
					class="hidden-file-input"
					@change="onImportFileChange"
				/>
				<input
					v-model="filterApp"
					class="filter-input"
					:placeholder="t('debug.statusMonitor.filterPlaceholder')"
				/>
				<button class="btn" @click="clearLogs">
					🗑️ {{ t('debug.statusMonitor.clear') }}
				</button>
				<button class="btn" @click="exportLogs">
					📥 {{ t('debug.statusMonitor.export') }}
				</button>
			</div>
		</header>

		<div
			v-if="importNotice"
			class="import-notice"
			:class="`import-notice--${importNotice.type}`"
		>
			<span>{{ importNotice.message }}</span>
			<button class="import-notice-close" @click="importNotice = null">
				✕
			</button>
		</div>

		<div v-if="dataMode === 'imported'" class="import-banner">
			<div class="import-banner-title">
				{{ t('debug.statusMonitor.importReadonlyMode') }}
			</div>
			<div class="import-banner-meta">
				<span v-if="importedMeta?.fileName"
					>{{ t('debug.statusMonitor.fileLabel')
					}}{{ importedMeta.fileName }}</span
				>
				<span v-if="importedMeta?.exportTime"
					>{{ t('debug.statusMonitor.exportTimeLabel')
					}}{{ importedMeta.exportTime }}</span
				>
				<span
					>{{ t('debug.statusMonitor.importTimeLabel')
					}}{{ formatTime(importedMeta?.importedAt || Date.now()) }}</span
				>
			</div>
		</div>

		<div class="stats-bar">
			<div class="stat-group">
				<span class="stat-group-label">{{
					t('debug.statusMonitor.appCount')
				}}</span>
				<span class="stat-group-value">{{ monitoredApps.length }}</span>
			</div>
			<div class="stat-divider"></div>
			<div class="stat-group">
				<span class="stat-group-label">{{
					t('debug.statusMonitor.callCount')
				}}</span>
				<div class="stat-items">
					<div class="stat-chip">
						<span class="chip-label">{{ t('debug.statusMonitor.total') }}</span>
						<span class="chip-value">{{ logs.length }}</span>
					</div>
					<div class="stat-chip api">
						<span class="chip-label"
							>📡 {{ t('debug.statusMonitor.api') }}</span
						>
						<span class="chip-value">{{ sourceStats.api }}</span>
					</div>
					<div class="stat-chip socket">
						<span class="chip-label"
							>🔌 {{ t('debug.statusMonitor.socket') }}</span
						>
						<span class="chip-value">{{ sourceStats.socket }}</span>
					</div>
					<div class="stat-chip local">
						<span class="chip-label"
							>🖱️ {{ t('debug.statusMonitor.local') }}</span
						>
						<span class="chip-value">{{ sourceStats.local }}</span>
					</div>
				</div>
			</div>
			<div
				class="stat-divider"
				v-if="sourceStats.rejected > 0 || sourceStats.regression > 0"
			></div>
			<div
				class="stat-group"
				v-if="sourceStats.rejected > 0 || sourceStats.regression > 0"
			>
				<span class="stat-group-label">{{
					t('debug.statusMonitor.anomaly')
				}}</span>
				<div class="stat-items">
					<button
						v-if="sourceStats.rejected > 0"
						class="stat-chip warning clickable"
						:class="{ active: issueFilter === 'rejected' }"
						@click="toggleIssueFilter('rejected')"
					>
						<span class="chip-label">{{
							t('debug.statusMonitor.rejected')
						}}</span>
						<span class="chip-value">{{ sourceStats.rejected }}</span>
					</button>
					<button
						v-if="sourceStats.regression > 0"
						class="stat-chip error clickable"
						:class="{ active: issueFilter === 'regression' }"
						@click="toggleIssueFilter('regression')"
					>
						<span class="chip-label"
							>⚠️ {{ t('debug.statusMonitor.regression') }}</span
						>
						<span class="chip-value">{{ sourceStats.regression }}</span>
					</button>
					<button
						v-if="issueFilter !== 'all'"
						class="stat-chip clickable"
						@click="issueFilter = 'all'"
					>
						<span class="chip-label">{{
							t('debug.statusMonitor.clearFilter')
						}}</span>
					</button>
				</div>
			</div>
		</div>

		<div class="main-content">
			<div class="app-list-panel">
				<div class="panel-header">
					{{ t('debug.statusMonitor.monitoredApps') }}
					<span>{{ monitoredApps.length }}</span>
				</div>
				<div class="app-list-body">
					<div
						v-for="app in monitoredApps"
						:key="app.combinedId"
						:class="[
							'app-item',
							{
								active: selectedApp === app.combinedId,
								'has-issue': app.hasIssue
							}
						]"
						@click="selectedApp = app.combinedId"
					>
						<div class="app-name">{{ app.appName }}</div>
						<div class="app-meta">
							<span
								class="state-badge"
								:class="getStateClass(app.currentState)"
							>
								{{ app.currentState }}
							</span>
							<span class="update-count">{{
								t('debug.statusMonitor.updatesCount', {
									count: app.updateCount
								})
							}}</span>
						</div>
					</div>
					<div v-if="monitoredApps.length === 0" class="empty">
						{{
							isRecording
								? t('debug.statusMonitor.waitingStateChange')
								: t('debug.statusMonitor.clickStartMonitor')
						}}
					</div>
				</div>
			</div>

			<div class="log-panel">
				<div class="panel-header">
					<div class="log-panel-title">
						{{ t('debug.statusMonitor.stateTransitionLog') }}
						<span v-if="selectedApp">- {{ selectedApp }}</span>
					</div>
					<div class="log-header-actions">
						<button
							v-if="selectedApp"
							class="tiny-btn"
							@click="onlySelectedAppLogs = !onlySelectedAppLogs"
						>
							{{
								onlySelectedAppLogs
									? t('debug.statusMonitor.viewAll')
									: t('debug.statusMonitor.onlySelectedApp')
							}}
						</button>
					</div>
				</div>
				<div class="log-body">
					<div v-if="selectedLog" class="selected-log-card">
						<div class="selected-log-title">
							{{ t('debug.statusMonitor.currentSelectedDetail') }}
						</div>
						<div class="selected-log-meta">
							<span>{{ selectedLog.appName }}</span>
							<span>{{ formatTime(selectedLog.timestamp) }}</span>
							<span
								>{{ selectedLog.oldState || 'N/A' }} ->
								{{ selectedLog.newState }}</span
							>
						</div>
						<div class="compare-grid">
							<div class="compare-col compare-col-old">
								<div class="compare-title">OLD</div>
								<div
									class="compare-scroll"
									:ref="
										(el) =>
											setComparePaneRef(
												`selected-${selectedLog?.logId ?? 'none'}`,
												'old',
												el as HTMLElement | null
											)
									"
									@scroll="
										onCompareScroll(
											`selected-${selectedLog?.logId ?? 'none'}`,
											'old',
											$event
										)
									"
								>
									<div
										v-for="(row, idx) in getCompareRows(selectedLog)"
										:key="`selected-old-${selectedLog.logId}-${idx}`"
										class="compare-line"
										:class="{ changed: row.changed }"
									>
										{{ row.old || ' ' }}
									</div>
								</div>
							</div>
							<div class="compare-col compare-col-new">
								<div class="compare-title">NEW</div>
								<div
									class="compare-scroll"
									:ref="
										(el) =>
											setComparePaneRef(
												`selected-${selectedLog?.logId ?? 'none'}`,
												'new',
												el as HTMLElement | null
											)
									"
									@scroll="
										onCompareScroll(
											`selected-${selectedLog?.logId ?? 'none'}`,
											'new',
											$event
										)
									"
								>
									<div
										v-for="(row, idx) in getCompareRows(selectedLog)"
										:key="`selected-new-${selectedLog.logId}-${idx}`"
										class="compare-line"
										:class="{ changed: row.changed }"
									>
										{{ row.new || ' ' }}
									</div>
								</div>
							</div>
						</div>
					</div>
					<div
						v-for="log in filteredLogs"
						:key="log.logId"
						:class="[
							'log-item',
							log.type,
							{
								'is-issue': log.isIssue,
								'selected-log': selectedLogId === log.logId,
								'selected-app': !!selectedApp && selectedApp === log.combinedId
							}
						]"
						@click="selectLog(log.logId)"
					>
						<div class="log-time">{{ formatTime(log.timestamp) }}</div>
						<div class="log-main">
							<div class="log-header">
								<span class="log-app">{{ log.appName }}</span>
								<span :class="['log-source', log.source]">
									{{
										log.source === 'api'
											? `📡 ${t('debug.statusMonitor.api')}`
											: log.source === 'socket'
											? `🔌 ${t('debug.statusMonitor.socket')}`
											: `🖱️ ${t('debug.statusMonitor.local')}`
									}}
								</span>
								<span v-if="log.isIssue" class="issue-badge">
									{{ log.issueType }}
								</span>
							</div>
							<div class="log-transition">
								<span
									class="state-badge small"
									:class="getStateClass(log.oldState)"
								>
									{{ log.oldState || 'N/A' }}
								</span>
								<span class="arrow">→</span>
								<span
									class="state-badge small"
									:class="getStateClass(log.newState)"
								>
									{{ log.newState }}
								</span>
							</div>
							<div class="log-times">
								<span
									>{{ t('debug.statusMonitor.oldTime') }}
									{{ log.oldTime || 'N/A' }}</span
								>
								<span
									>{{ t('debug.statusMonitor.newTime') }}
									{{ log.newTime }}</span
								>
								<span
									v-if="log.timeDiff !== undefined"
									:class="{ 'time-negative': log.timeDiff < 0 }"
								>
									{{ t('debug.statusMonitor.diff') }}
									{{ log.timeDiff > 0 ? '+' : '' }}{{ log.timeDiff }}ms
								</span>
							</div>
						</div>
						<button class="expand-btn" @click.stop="toggleExpand(log.logId)">
							{{
								expandedLogs.has(log.logId)
									? t('debug.statusMonitor.collapse')
									: t('debug.statusMonitor.detail')
							}}
						</button>
						<div v-if="expandedLogs.has(log.logId)" class="log-detail">
							<div class="compare-grid">
								<div class="compare-col compare-col-old">
									<div class="compare-title">OLD</div>
									<div
										class="compare-scroll"
										:ref="
											(el) =>
												setComparePaneRef(
													`log-${log.logId}`,
													'old',
													el as HTMLElement | null
												)
										"
										@scroll="onCompareScroll(`log-${log.logId}`, 'old', $event)"
									>
										<div
											v-for="(row, idx) in getCompareRows(log)"
											:key="`log-old-${log.logId}-${idx}`"
											class="compare-line"
											:class="{ changed: row.changed }"
										>
											{{ row.old || ' ' }}
										</div>
									</div>
								</div>
								<div class="compare-col compare-col-new">
									<div class="compare-title">NEW</div>
									<div
										class="compare-scroll"
										:ref="
											(el) =>
												setComparePaneRef(
													`log-${log.logId}`,
													'new',
													el as HTMLElement | null
												)
										"
										@scroll="onCompareScroll(`log-${log.logId}`, 'new', $event)"
									>
										<div
											v-for="(row, idx) in getCompareRows(log)"
											:key="`log-new-${log.logId}-${idx}`"
											class="compare-line"
											:class="{ changed: row.changed }"
										>
											{{ row.new || ' ' }}
										</div>
									</div>
								</div>
							</div>
						</div>
					</div>
					<div v-if="filteredLogs.length === 0" class="empty">
						{{
							selectedApp
								? t('debug.statusMonitor.noLogsForApp')
								: t('debug.statusMonitor.noLogs')
						}}
					</div>
				</div>
			</div>
		</div>
	</div>
</template>

<script setup lang="ts">
import { ref, computed, onUnmounted } from 'vue';
import { useI18n } from 'vue-i18n';
import { useStatusMonitor, StatusLog } from './useStatusMonitor';

const {
	isRecording,
	logs,
	monitoredApps: baseMonitoredApps,
	sourceStats,
	toggleRecording: baseToggleRecording,
	importSnapshot,
	dataMode,
	importedMeta,
	clearLogs: baseClearLogs,
	exportLogs,
	cleanup,
	formatTime
} = useStatusMonitor();
const { t } = useI18n();

const filterApp = ref('');
const selectedApp = ref<string | null>(null);
const selectedLogId = ref<string | null>(null);
const expandedLogs = ref(new Set<string>());
const issueFilter = ref<'all' | 'rejected' | 'regression'>('all');
const onlySelectedAppLogs = ref(false);
const importInputRef = ref<HTMLInputElement | null>(null);
const importNotice = ref<{
	type: 'success' | 'error' | 'info';
	message: string;
} | null>(null);
let importNoticeTimer: ReturnType<typeof setTimeout> | null = null;

const monitoredApps = computed(() => {
	let apps = baseMonitoredApps.value;
	if (filterApp.value) {
		const filter = filterApp.value.toLowerCase();
		apps = apps.filter(
			(app) =>
				app.appName.toLowerCase().includes(filter) ||
				app.combinedId.toLowerCase().includes(filter)
		);
	}
	return apps;
});

const filteredLogs = computed(() => {
	let result = logs.value;

	if (issueFilter.value !== 'all') {
		result = result.filter((log) => log.type === issueFilter.value);
	}

	if (onlySelectedAppLogs.value && selectedApp.value) {
		result = result.filter((log) => log.combinedId === selectedApp.value);
	}

	if (filterApp.value) {
		const filter = filterApp.value.toLowerCase();
		result = result.filter(
			(log) =>
				log.appName.toLowerCase().includes(filter) ||
				log.combinedId.toLowerCase().includes(filter)
		);
	}
	return result.slice().reverse();
});

const selectedLog = computed<StatusLog | null>(() => {
	if (!selectedLogId.value) {
		return null;
	}
	return logs.value.find((log) => log.logId === selectedLogId.value) || null;
});

const toggleRecording = () => baseToggleRecording(true);

const openImportPicker = () => {
	importInputRef.value?.click();
};

const showImportNotice = (
	type: 'success' | 'error' | 'info',
	message: string,
	autoHide = true
) => {
	importNotice.value = { type, message };
	if (importNoticeTimer) {
		clearTimeout(importNoticeTimer);
		importNoticeTimer = null;
	}
	if (autoHide) {
		importNoticeTimer = setTimeout(() => {
			importNotice.value = null;
			importNoticeTimer = null;
		}, 5000);
	}
};

const onImportFileChange = async (event: Event) => {
	const input = event.target as HTMLInputElement;
	const file = input.files?.[0];
	if (!file) {
		return;
	}

	try {
		const text = await file.text();
		const payload = JSON.parse(text);
		const result = importSnapshot(payload, { fileName: file.name });
		if (!result.ok) {
			showImportNotice(
				'error',
				t(result.messageKey || 'debug.statusMonitor.importFailed')
			);
			return;
		}
		expandedLogs.value.clear();
		selectedApp.value = null;
		selectedLogId.value = null;
		issueFilter.value = 'all';
		onlySelectedAppLogs.value = false;
		const count = Array.isArray(payload?.logs)
			? payload.logs.length
			: Array.isArray(payload)
			? payload.length
			: logs.value.length;
		showImportNotice(
			'success',
			t('debug.statusMonitor.importSuccess', { count })
		);
	} catch (error) {
		showImportNotice(
			'error',
			t('debug.statusMonitor.importSnapshotFailed', {
				message:
					(error as Error)?.message || t('debug.statusMonitor.checkFileFormat')
			})
		);
	} finally {
		input.value = '';
	}
};

const clearLogs = () => {
	baseClearLogs();
	expandedLogs.value.clear();
	selectedApp.value = null;
	selectedLogId.value = null;
	issueFilter.value = 'all';
	onlySelectedAppLogs.value = false;
};

const toggleExpand = (logId: string) => {
	if (expandedLogs.value.has(logId)) {
		expandedLogs.value.delete(logId);
	} else {
		expandedLogs.value.add(logId);
	}
};

const selectLog = (logId: string) => {
	selectedLogId.value = logId;
};

const toggleIssueFilter = (type: 'rejected' | 'regression') => {
	issueFilter.value = issueFilter.value === type ? 'all' : type;
};

const toPrettyJson = (value: unknown): string => {
	if (value === undefined) return 'N/A';
	if (value === null) return 'null';
	try {
		return JSON.stringify(value, null, 2);
	} catch {
		return String(value);
	}
};

const getOldDetailText = (log: StatusLog): string => {
	const oldVal = (log.rawData as any)?.old;
	if (oldVal !== undefined) {
		return toPrettyJson(oldVal);
	}
	return toPrettyJson({
		state: log.oldState,
		time: log.oldTime
	});
};

const getNewDetailText = (log: StatusLog): string => {
	const newVal = (log.rawData as any)?.new;
	if (newVal !== undefined) {
		return toPrettyJson(newVal);
	}
	return toPrettyJson({
		state: log.newState,
		time: log.newTime
	});
};

type CompareRow = {
	old: string;
	new: string;
	changed: boolean;
};

const compareRowsCache = new Map<string, CompareRow[]>();
const comparePaneRefs = new Map<
	string,
	{ old?: HTMLElement | null; new?: HTMLElement | null }
>();
let isSyncingCompareScroll = false;

const buildCompareRows = (oldText: string, newText: string): CompareRow[] => {
	const oldLines = oldText.split('\n');
	const newLines = newText.split('\n');
	const len = Math.max(oldLines.length, newLines.length);
	const rows: CompareRow[] = [];
	for (let i = 0; i < len; i++) {
		const oldLine = oldLines[i] ?? '';
		const newLine = newLines[i] ?? '';
		rows.push({
			old: oldLine,
			new: newLine,
			changed: oldLine !== newLine
		});
	}
	return rows;
};

const getCompareRows = (log: StatusLog): CompareRow[] => {
	const oldText = getOldDetailText(log);
	const newText = getNewDetailText(log);
	const cacheKey = `${log.logId}__${oldText.length}__${newText.length}`;
	const cached = compareRowsCache.get(cacheKey);
	if (cached) {
		return cached;
	}
	const rows = buildCompareRows(oldText, newText);
	compareRowsCache.set(cacheKey, rows);
	return rows;
};

const setComparePaneRef = (
	key: string,
	side: 'old' | 'new',
	el: HTMLElement | null
) => {
	const pair = comparePaneRefs.get(key) || {};
	pair[side] = el;
	comparePaneRefs.set(key, pair);
};

const onCompareScroll = (key: string, side: 'old' | 'new', ev: Event) => {
	if (isSyncingCompareScroll) {
		return;
	}
	const pair = comparePaneRefs.get(key);
	if (!pair) {
		return;
	}
	const source = ev.target as HTMLElement;
	const target = side === 'old' ? pair.new : pair.old;
	if (!source || !target) {
		return;
	}
	isSyncingCompareScroll = true;
	const sourceMax = Math.max(source.scrollHeight - source.clientHeight, 1);
	const targetMax = Math.max(target.scrollHeight - target.clientHeight, 1);
	target.scrollTop = (source.scrollTop / sourceMax) * targetMax;
	target.scrollLeft = source.scrollLeft;
	requestAnimationFrame(() => {
		isSyncingCompareScroll = false;
	});
};

const getStateClass = (state: string | null) => {
	if (!state) return 'unknown';
	const lower = state.toLowerCase();
	if (lower.includes('running') || lower.includes('completed'))
		return 'success';
	if (lower.includes('fail') || lower.includes('error')) return 'error';
	if (
		lower.includes('install') ||
		lower.includes('download') ||
		lower.includes('pending')
	)
		return 'progress';
	if (lower.includes('uninstall')) return 'uninstall';
	return 'default';
};

onUnmounted(() => {
	if (importNoticeTimer) {
		clearTimeout(importNoticeTimer);
		importNoticeTimer = null;
	}
	cleanup();
});
</script>

<style scoped>
.status-monitor {
	height: 100%;
	display: flex;
	flex-direction: column;
	background: #fff;
	font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
}

.toolbar {
	height: 52px;
	background: linear-gradient(135deg, #1e293b 0%, #334155 100%);
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 0 16px;
	flex-shrink: 0;
}

.title {
	font-size: 15px;
	font-weight: 600;
	color: #fff;
}

.controls {
	display: flex;
	gap: 8px;
	align-items: center;
}

.hidden-file-input {
	display: none;
}

.filter-input {
	padding: 6px 12px;
	border: 1px solid rgba(255, 255, 255, 0.2);
	border-radius: 6px;
	background: rgba(255, 255, 255, 0.1);
	color: #fff;
	font-size: 12px;
	width: 180px;
}

.filter-input::placeholder {
	color: rgba(255, 255, 255, 0.5);
}

.btn {
	padding: 6px 12px;
	border: 1px solid rgba(255, 255, 255, 0.3);
	border-radius: 6px;
	background: rgba(255, 255, 255, 0.1);
	color: #fff;
	font-size: 12px;
	cursor: pointer;
	transition: all 0.15s;
}

.btn:hover {
	background: rgba(255, 255, 255, 0.2);
}

.btn.active {
	background: #ef4444;
	border-color: #ef4444;
}

.import-notice {
	padding: 8px 16px;
	display: flex;
	align-items: center;
	justify-content: space-between;
	gap: 12px;
	font-size: 12px;
	border-bottom: 1px solid;
}

.import-notice--success {
	background: #ecfdf5;
	color: #065f46;
	border-color: #6ee7b7;
}

.import-notice--error {
	background: #fef2f2;
	color: #991b1b;
	border-color: #fca5a5;
}

.import-notice--info {
	background: #eff6ff;
	color: #1e40af;
	border-color: #93c5fd;
}

.import-notice-close {
	border: none;
	background: transparent;
	color: inherit;
	font-size: 12px;
	cursor: pointer;
}

.import-banner {
	padding: 8px 16px;
	background: #fff7ed;
	border-bottom: 1px solid #fed7aa;
	display: flex;
	flex-direction: column;
	gap: 4px;
}

.import-banner-title {
	font-size: 12px;
	font-weight: 700;
	color: #9a3412;
}

.import-banner-meta {
	display: flex;
	flex-wrap: wrap;
	gap: 12px;
	font-size: 11px;
	color: #9a3412;
}

.stats-bar {
	display: flex;
	align-items: center;
	gap: 16px;
	padding: 10px 16px;
	background: #f8fafc;
	border-bottom: 1px solid #e2e8f0;
	flex-shrink: 0;
}

.stat-group {
	display: flex;
	align-items: center;
	gap: 8px;
}

.stat-group-label {
	font-size: 10px;
	font-weight: 600;
	color: #94a3b8;
	text-transform: uppercase;
}

.stat-group-value {
	font-size: 18px;
	font-weight: 700;
	color: #1e293b;
}

.stat-divider {
	width: 1px;
	height: 24px;
	background: #e2e8f0;
}

.stat-items {
	display: flex;
	gap: 6px;
}

.stat-chip {
	display: flex;
	align-items: center;
	gap: 4px;
	padding: 4px 8px;
	background: #e2e8f0;
	border-radius: 12px;
	font-size: 11px;
	border: 1px solid transparent;
}

.stat-chip.clickable {
	cursor: pointer;
}

.stat-chip.clickable.active {
	border-color: #6366f1;
	box-shadow: 0 0 0 1px rgba(99, 102, 241, 0.2);
}

.chip-label {
	color: #64748b;
}

.chip-value {
	font-weight: 600;
	color: #1e293b;
}

.stat-chip.api {
	background: #dbeafe;
}

.stat-chip.api .chip-value {
	color: #1e40af;
}

.stat-chip.socket {
	background: #f3e8ff;
}

.stat-chip.socket .chip-value {
	color: #7c3aed;
}

.stat-chip.local {
	background: #d1fae5;
}

.stat-chip.local .chip-value {
	color: #065f46;
}

.stat-chip.warning {
	background: #fef3c7;
}

.stat-chip.warning .chip-value {
	color: #b45309;
}

.stat-chip.error {
	background: #fee2e2;
}

.stat-chip.error .chip-value {
	color: #dc2626;
}

.main-content {
	flex: 1;
	display: flex;
	min-height: 0;
}

.app-list-panel {
	width: 280px;
	border-right: 1px solid #e2e8f0;
	display: flex;
	flex-direction: column;
	min-height: 0;
}

.panel-header {
	padding: 10px 14px;
	font-size: 11px;
	font-weight: 600;
	color: #64748b;
	text-transform: uppercase;
	border-bottom: 1px solid #e2e8f0;
	display: flex;
	justify-content: space-between;
	flex-shrink: 0;
}

.panel-header span {
	color: #6366f1;
}

.log-panel-title {
	display: flex;
	align-items: center;
	gap: 6px;
}

.log-header-actions {
	display: flex;
	align-items: center;
	gap: 8px;
}

.tiny-btn {
	padding: 2px 8px;
	border: 1px solid #d1d5db;
	border-radius: 6px;
	background: #ffffff;
	color: #334155;
	font-size: 10px;
	cursor: pointer;
}

.tiny-btn:hover {
	border-color: #6366f1;
	color: #4338ca;
}

.app-list-body {
	flex: 1;
	overflow-y: auto;
	padding: 6px;
}

.app-item {
	padding: 10px 12px;
	border-radius: 6px;
	cursor: pointer;
	margin-bottom: 4px;
	border: 1px solid transparent;
	transition: all 0.15s;
}

.app-item:hover {
	background: #f1f5f9;
}

.app-item.active {
	background: #eef2ff;
	border-color: #6366f1;
}

.app-item.has-issue {
	border-left: 3px solid #ef4444;
}

.app-name {
	font-size: 12px;
	font-weight: 600;
	color: #1e293b;
	margin-bottom: 4px;
}

.app-meta {
	display: flex;
	align-items: center;
	gap: 8px;
}

.update-count {
	font-size: 10px;
	color: #94a3b8;
}

.state-badge {
	padding: 2px 8px;
	border-radius: 10px;
	font-size: 10px;
	font-weight: 600;
}

.state-badge.small {
	padding: 1px 6px;
	font-size: 9px;
}

.state-badge.success {
	background: #dcfce7;
	color: #166534;
}

.state-badge.error {
	background: #fee2e2;
	color: #991b1b;
}

.state-badge.progress {
	background: #dbeafe;
	color: #1e40af;
}

.state-badge.uninstall {
	background: #f3e8ff;
	color: #7c3aed;
}

.state-badge.default {
	background: #e2e8f0;
	color: #475569;
}

.state-badge.unknown {
	background: #f1f5f9;
	color: #94a3b8;
}

.log-panel {
	flex: 1;
	display: flex;
	flex-direction: column;
	min-height: 0;
	background: #f8fafc;
}

.log-body {
	flex: 1;
	overflow-y: auto;
	padding: 8px;
}

.selected-log-card {
	margin-bottom: 10px;
	padding: 10px;
	border: 1px solid #6366f1;
	background: #eef2ff;
	border-radius: 8px;
}

.selected-log-title {
	font-size: 11px;
	font-weight: 700;
	color: #3730a3;
	margin-bottom: 6px;
}

.selected-log-meta {
	display: flex;
	flex-wrap: wrap;
	gap: 8px;
	font-size: 10px;
	color: #475569;
	margin-bottom: 8px;
}

.selected-log-card pre {
	display: none;
}

.log-item {
	background: #fff;
	border: 1px solid #e2e8f0;
	border-radius: 8px;
	padding: 10px 12px;
	margin-bottom: 6px;
	position: relative;
	cursor: pointer;
}

.log-item.rejected {
	border-left: 3px solid #f59e0b;
	background: #fffbeb;
}

.log-item.regression,
.log-item.is-issue {
	border-left: 3px solid #ef4444;
	background: #fef2f2;
}

.log-item.selected-app {
	outline: 1px dashed #6366f1;
}

.log-item.selected-log {
	border-color: #6366f1;
	box-shadow: 0 0 0 1px rgba(99, 102, 241, 0.25);
}

.log-time {
	font-size: 10px;
	color: #94a3b8;
	font-family: 'SF Mono', monospace;
	margin-bottom: 6px;
}

.log-header {
	display: flex;
	align-items: center;
	gap: 8px;
	margin-bottom: 6px;
}

.log-app {
	font-size: 12px;
	font-weight: 600;
	color: #1e293b;
}

.log-source {
	font-size: 10px;
	padding: 2px 6px;
	border-radius: 4px;
}

.log-source.api {
	background: #dbeafe;
	color: #1e40af;
}

.log-source.socket {
	background: #f3e8ff;
	color: #7c3aed;
}

.log-source.local {
	background: #d1fae5;
	color: #065f46;
}

.issue-badge {
	font-size: 10px;
	padding: 2px 6px;
	border-radius: 4px;
	background: #ef4444;
	color: #fff;
	font-weight: 600;
}

.log-transition {
	display: flex;
	align-items: center;
	gap: 8px;
	margin-bottom: 6px;
}

.arrow {
	color: #94a3b8;
	font-size: 12px;
}

.log-times {
	display: flex;
	gap: 16px;
	font-size: 10px;
	color: #64748b;
	font-family: 'SF Mono', monospace;
}

.time-negative {
	color: #ef4444;
	font-weight: 600;
}

.expand-btn {
	position: absolute;
	top: 10px;
	right: 10px;
	padding: 2px 8px;
	border: 1px solid #e2e8f0;
	border-radius: 4px;
	background: #fff;
	font-size: 10px;
	cursor: pointer;
}

.expand-btn:hover {
	border-color: #6366f1;
}

.log-detail {
	margin-top: 10px;
	padding: 10px;
	background: #f8fafc;
	border-radius: 6px;
	border: 1px solid #e2e8f0;
}

.log-detail pre {
	display: none;
}

.compare-grid {
	display: grid;
	grid-template-columns: 1fr 1fr;
	gap: 10px;
}

.compare-col {
	border: 1px solid #e2e8f0;
	border-radius: 6px;
	background: #ffffff;
	overflow: hidden;
	display: flex;
	flex-direction: column;
}

.compare-col-old {
	border-left: 3px solid #f59e0b;
}

.compare-col-new {
	border-left: 3px solid #22c55e;
}

.compare-title {
	padding: 6px 8px;
	font-size: 10px;
	font-weight: 700;
	letter-spacing: 0.04em;
	color: #334155;
	background: #f8fafc;
	border-bottom: 1px solid #e2e8f0;
}

.compare-scroll {
	flex: 1;
	overflow: auto;
	max-height: 380px;
	background: #ffffff;
}

.compare-line {
	min-height: 18px;
	padding: 0 8px;
	font-size: 10px;
	line-height: 18px;
	font-family: 'SF Mono', 'Fira Code', monospace;
	color: #0f172a;
	white-space: pre;
	border-bottom: 1px solid #f1f5f9;
}

.compare-line.changed {
	background: #fff7ed;
}

@media (max-width: 1100px) {
	.compare-grid {
		grid-template-columns: 1fr;
	}
}

.empty {
	padding: 30px;
	text-align: center;
	color: #94a3b8;
	font-size: 13px;
}

.app-list-body::-webkit-scrollbar,
.log-body::-webkit-scrollbar,
.log-detail pre::-webkit-scrollbar {
	width: 4px;
}

.app-list-body::-webkit-scrollbar-thumb,
.log-body::-webkit-scrollbar-thumb,
.log-detail pre::-webkit-scrollbar-thumb {
	background: #cbd5e1;
	border-radius: 2px;
}
</style>
