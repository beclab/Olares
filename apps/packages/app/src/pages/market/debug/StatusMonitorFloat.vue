<template>
	<div
		v-if="!isExpanded"
		class="float-trigger"
		:class="{ 'has-issue': hasIssue, 'is-recording': isRecording }"
		@click="isExpanded = true"
	>
		<span class="trigger-icon">{{ isRecording ? '🔴' : '📡' }}</span>
		<span v-if="isRecording" class="trigger-count">{{ logs.length }}</span>
	</div>

	<div v-if="isExpanded" class="float-panel" :style="panelStyle">
		<div class="panel-header" @mousedown="startDrag">
			<span class="header-title">📡 {{ t('debug.statusFloat.title') }}</span>
			<div class="header-actions">
				<button
					:class="['action-btn', { active: isRecording }]"
					@click.stop="toggleRecording"
				>
					{{ isRecording ? '⏹' : '▶' }}
				</button>
				<button class="action-btn" @click.stop="clearLogs">🗑</button>
				<button class="action-btn" @click.stop="isExpanded = false">✕</button>
			</div>
		</div>

		<div class="panel-stats">
			<span class="stat">📱 {{ appCount }}</span>
			<span class="stat">📡 {{ sourceStats.api }}</span>
			<span class="stat">🔌 {{ sourceStats.socket }}</span>
			<span class="stat">🖱 {{ sourceStats.local }}</span>
			<span v-if="sourceStats.regression > 0" class="stat error"
				>⚠️ {{ sourceStats.regression }}</span
			>
		</div>

		<div class="panel-logs" ref="logsContainer">
			<div
				v-for="(log, index) in recentLogs"
				:key="index"
				:class="['log-row', log.type, log.source]"
			>
				<span class="log-time">{{ formatTime(log.timestamp) }}</span>
				<span class="log-source">{{ getSourceIcon(log.source) }}</span>
				<span class="log-app">{{ log.appName }}</span>
				<span class="log-state"
					>{{ log.oldState || '?' }} → {{ log.newState }}</span
				>
			</div>
			<div v-if="logs.length === 0" class="empty-hint">
				{{
					isRecording
						? t('debug.statusMonitor.waitingStateChange')
						: t('debug.statusFloat.clickToStart')
				}}
			</div>
		</div>

		<div class="panel-footer">
			<span class="footer-hint">{{
				t('debug.statusFloat.dragHint', { count: logs.length })
			}}</span>
			<button class="open-full-btn" @click="openFullPage">
				{{ t('debug.statusFloat.openFullPage') }}
			</button>
		</div>
	</div>
</template>

<script setup lang="ts">
import { ref, computed, onUnmounted } from 'vue';
import { useI18n } from 'vue-i18n';
import { useStatusMonitor } from './useStatusMonitor';

const {
	isRecording,
	logs,
	monitoredApps,
	sourceStats,
	hasIssue,
	toggleRecording,
	clearLogs,
	cleanup,
	getSourceIcon
} = useStatusMonitor();
const { t } = useI18n();

const isExpanded = ref(false);
const logsContainer = ref<HTMLElement | null>(null);

const panelPosition = ref({ x: window.innerWidth - 420, y: 80 });
const isDragging = ref(false);
const dragOffset = ref({ x: 0, y: 0 });

const panelStyle = computed(() => ({
	left: `${panelPosition.value.x}px`,
	top: `${panelPosition.value.y}px`
}));

const VISIBLE_LOGS = 50;
const recentLogs = computed(() => logs.value.slice(-VISIBLE_LOGS).reverse());

const appCount = computed(() => monitoredApps.value.length);

const openFullPage = () => {
	window.open('/dev?_dm=Olares', '_blank');
};

const startDrag = (e: MouseEvent) => {
	isDragging.value = true;
	dragOffset.value = {
		x: e.clientX - panelPosition.value.x,
		y: e.clientY - panelPosition.value.y
	};
	document.addEventListener('mousemove', onDrag);
	document.addEventListener('mouseup', stopDrag);
};

const onDrag = (e: MouseEvent) => {
	if (!isDragging.value) return;
	panelPosition.value = {
		x: Math.max(
			0,
			Math.min(window.innerWidth - 400, e.clientX - dragOffset.value.x)
		),
		y: Math.max(
			0,
			Math.min(window.innerHeight - 300, e.clientY - dragOffset.value.y)
		)
	};
};

const stopDrag = () => {
	isDragging.value = false;
	document.removeEventListener('mousemove', onDrag);
	document.removeEventListener('mouseup', stopDrag);
};

const formatTime = (ts: number) => {
	const d = new Date(ts);
	return d.toLocaleTimeString('en-GB', { hour12: false }).slice(-8);
};

onUnmounted(() => {
	cleanup();
});
</script>

<style scoped>
.float-trigger {
	position: fixed;
	bottom: 20px;
	right: 20px;
	width: 48px;
	height: 48px;
	background: #1e293b;
	border-radius: 50%;
	display: flex;
	align-items: center;
	justify-content: center;
	cursor: pointer;
	box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
	z-index: 9999;
	transition: all 0.2s;
}

.float-trigger:hover {
	transform: scale(1.1);
	background: #334155;
}

.float-trigger.is-recording {
	animation: pulse 1.5s infinite;
}

.float-trigger.has-issue {
	background: #ef4444;
}

@keyframes pulse {
	0%,
	100% {
		box-shadow: 0 0 0 0 rgba(239, 68, 68, 0.4);
	}
	50% {
		box-shadow: 0 0 0 10px rgba(239, 68, 68, 0);
	}
}

.trigger-icon {
	font-size: 20px;
}

.trigger-count {
	position: absolute;
	top: -4px;
	right: -4px;
	background: #ef4444;
	color: #fff;
	font-size: 10px;
	font-weight: 600;
	padding: 2px 6px;
	border-radius: 10px;
	min-width: 18px;
	text-align: center;
}

.float-panel {
	position: fixed;
	width: 400px;
	height: 380px;
	background: #fff;
	border-radius: 12px;
	box-shadow: 0 8px 32px rgba(0, 0, 0, 0.2);
	z-index: 10000;
	display: flex;
	flex-direction: column;
	overflow: hidden;
	font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
}

.panel-header {
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 10px 12px;
	background: linear-gradient(135deg, #1e293b 0%, #334155 100%);
	cursor: move;
	user-select: none;
}

.header-title {
	font-size: 13px;
	font-weight: 600;
	color: #fff;
}

.header-actions {
	display: flex;
	gap: 4px;
}

.action-btn {
	width: 24px;
	height: 24px;
	border: none;
	border-radius: 4px;
	background: rgba(255, 255, 255, 0.1);
	color: #fff;
	cursor: pointer;
	font-size: 12px;
	display: flex;
	align-items: center;
	justify-content: center;
}

.action-btn:hover {
	background: rgba(255, 255, 255, 0.2);
}

.action-btn.active {
	background: #ef4444;
}

.panel-stats {
	display: flex;
	gap: 12px;
	padding: 8px 12px;
	background: #f8fafc;
	border-bottom: 1px solid #e2e8f0;
	font-size: 11px;
}

.stat {
	color: #64748b;
}

.stat.error {
	color: #ef4444;
	font-weight: 600;
}

.panel-logs {
	flex: 1;
	overflow-y: auto;
	padding: 6px;
	font-size: 11px;
}

.log-row {
	display: flex;
	align-items: center;
	gap: 6px;
	padding: 4px 6px;
	border-radius: 4px;
	margin-bottom: 2px;
}

.log-row.regression {
	background: #fee2e2;
}

.log-row.rejected {
	background: #fef3c7;
}

.log-time {
	color: #94a3b8;
	font-family: monospace;
	font-size: 10px;
	flex-shrink: 0;
}

.log-source {
	flex-shrink: 0;
}

.log-app {
	font-weight: 600;
	color: #1e293b;
	flex-shrink: 0;
	max-width: 100px;
	overflow: hidden;
	text-overflow: ellipsis;
	white-space: nowrap;
}

.log-state {
	color: #64748b;
	flex: 1;
	overflow: hidden;
	text-overflow: ellipsis;
	white-space: nowrap;
}

.empty-hint {
	text-align: center;
	color: #94a3b8;
	padding: 20px;
}

.panel-footer {
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 8px 12px;
	background: #f8fafc;
	border-top: 1px solid #e2e8f0;
}

.footer-hint {
	font-size: 10px;
	color: #94a3b8;
}

.open-full-btn {
	padding: 4px 10px;
	border: 1px solid #6366f1;
	border-radius: 4px;
	background: #fff;
	color: #6366f1;
	font-size: 10px;
	cursor: pointer;
}

.open-full-btn:hover {
	background: #6366f1;
	color: #fff;
}

.panel-logs::-webkit-scrollbar {
	width: 4px;
}

.panel-logs::-webkit-scrollbar-thumb {
	background: #cbd5e1;
	border-radius: 2px;
}
</style>
