<template>
	<div class="labels-viewer">
		<header class="toolbar">
			<div class="title">🏷️ {{ t('debug.labels.title') }}</div>
			<div class="stats">
				{{
					t('debug.labels.totalWithLabels', { count: appsWithLabels.length })
				}}
			</div>
		</header>

		<div class="labels-area">
			<!-- {{ t('debug.labels.groupStats') }} -->
			<div class="label-groups">
				<div class="group-header">{{ t('debug.labels.groupStats') }}</div>
				<div class="group-list">
					<button
						:class="['group-btn', { active: selectedGroup === '' }]"
						@click="selectedGroup = ''"
					>
						{{ t('debug.common.all') }} <span>{{ appsWithLabels.length }}</span>
					</button>
					<button
						v-for="g in labelGroups"
						:key="g.labels"
						:class="['group-btn', { active: selectedGroup === g.labels }]"
						@click="selectedGroup = g.labels"
					>
						<span class="group-labels">
							<span
								v-for="label in g.labelArray"
								:key="label"
								class="label-tag"
								:style="getLabelStyle(label)"
							>
								{{ label }}
							</span>
						</span>
						<span class="group-count">{{ g.count }}</span>
					</button>
				</div>
			</div>

			<!-- {{ t('debug.labels.singleStats') }} -->
			<div class="single-labels">
				<div class="single-header">{{ t('debug.labels.singleStats') }}</div>
				<div class="single-list">
					<div v-for="s in singleLabelStats" :key="s.label" class="single-item">
						<span class="label-tag" :style="getLabelStyle(s.label)">
							{{ s.label }}
						</span>
						<span class="single-count">{{ s.count }}</span>
					</div>
				</div>
			</div>
		</div>

		<div class="result-area">
			<div class="app-list">
				<div class="list-header">
					{{ t('debug.labels.appList') }} <span>{{ filteredApps.length }}</span>
				</div>
				<div class="list-body">
					<div
						v-for="(app, index) in filteredApps"
						:key="app.key"
						:class="['app-item', { active: selectedIndex === index }]"
						@click="selectedIndex = index"
					>
						<div class="app-main">
							<span class="app-name">{{ app.appInfo.app_name }}</span>
							<span class="app-version">{{ app.appInfo.app_version }}</span>
						</div>
						<div class="app-labels">
							<span
								v-for="label in app.appInfo.app_labels"
								:key="label"
								class="label-tag small"
								:style="getLabelStyle(label)"
							>
								{{ label }}
							</span>
						</div>
						<div class="app-title">{{ getTitle(app.appInfo) }}</div>
					</div>
					<div v-if="filteredApps.length === 0" class="empty">
						{{ t('debug.common.noMatchedData') }}
					</div>
				</div>
			</div>

			<div class="app-detail">
				<div v-if="selectedIndex === null" class="placeholder">
					👈 {{ t('debug.labels.selectApp') }}
				</div>
				<div v-else class="detail-content">
					<div class="detail-summary">
						<div class="summary-row">
							<span class="label">{{ t('debug.common.appName') }}</span>
							<span class="value">{{ selectedApp?.appInfo.app_name }}</span>
						</div>
						<div class="summary-row">
							<span class="label">Source</span>
							<span class="value">{{ selectedApp?.sourceName }}</span>
						</div>
						<div class="summary-row">
							<span class="label">{{ t('debug.common.version') }}</span>
							<span class="value">{{ selectedApp?.appInfo.app_version }}</span>
						</div>
						<div class="summary-row">
							<span class="label">Labels</span>
							<span class="value labels">
								<span
									v-for="label in selectedApp?.appInfo.app_labels"
									:key="label"
									class="label-tag"
									:style="getLabelStyle(label)"
								>
									{{ label }}
								</span>
							</span>
						</div>
					</div>

					<div class="detail-tabs">
						<button
							v-for="tab in tabs"
							:key="tab.key"
							:class="['tab-btn', { active: activeTab === tab.key }]"
							@click="activeTab = tab.key"
						>
							{{ tab.icon }} {{ tab.name }}
						</button>
					</div>

					<div class="detail-json">
						<div class="json-header">
							<span>{{ activeTabName }}</span>
							<button class="copy-btn" @click="copyData">
								{{ copied ? '✅' : '📋' }}
							</button>
						</div>
						<pre class="json-body">{{ activeTabData }}</pre>
					</div>
				</div>
			</div>
		</div>
	</div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue';
import { useI18n } from 'vue-i18n';
import { useAppStore } from '../../../stores/market/appStore';

const appStore = useAppStore();
const { t } = useI18n();

const LABEL_CONFIG: Record<
	string,
	{ name: string; bgColor: string; textColor: string }
> = {
	nsfw: { name: 'NSFW', bgColor: '#fecaca', textColor: '#991b1b' },
	remove: { name: 'Remove', bgColor: '#fed7aa', textColor: '#9a3412' },
	suspend: { name: 'Suspend', bgColor: '#fef08a', textColor: '#854d0e' }
};

const DEFAULT_LABEL_STYLE = { bgColor: '#e2e8f0', textColor: '#475569' };

const selectedGroup = ref('');
const selectedIndex = ref<number | null>(null);
const activeTab = ref('simple');
const copied = ref(false);

const tabs = [
	{ key: 'simple', name: 'Simple Info', icon: '📝' },
	{ key: 'status', name: 'Status', icon: '🔄' },
	{ key: 'full', name: 'Full Info', icon: '📦' }
];

const appsWithLabels = computed(() => {
	return Array.from(appStore.appSimpleInfoMap.entries())
		.filter(([key, data]) => {
			const labels = data.app_simple_info?.app_labels;
			return labels && Array.isArray(labels) && labels.length > 0;
		})
		.map(([key, data]) => {
			const [sourceName, application] = key.split('_');
			return {
				key,
				sourceName,
				application,
				appInfo: data.app_simple_info,
				labelsKey: [...data.app_simple_info.app_labels].sort().join(',')
			};
		});
});

// {{ t('debug.labels.groupStats') }}
const labelGroups = computed(() => {
	const groupMap = new Map<string, number>();

	appsWithLabels.value.forEach((app) => {
		const key = app.labelsKey;
		groupMap.set(key, (groupMap.get(key) || 0) + 1);
	});

	return Array.from(groupMap.entries())
		.map(([labels, count]) => ({
			labels,
			labelArray: labels.split(','),
			count
		}))
		.sort((a, b) => b.count - a.count);
});

// {{ t('debug.labels.singleStats') }}
const singleLabelStats = computed(() => {
	const labelMap = new Map<string, number>();

	appsWithLabels.value.forEach((app) => {
		app.appInfo.app_labels.forEach((label: string) => {
			labelMap.set(label, (labelMap.get(label) || 0) + 1);
		});
	});

	return Array.from(labelMap.entries())
		.map(([label, count]) => ({ label, count }))
		.sort((a, b) => b.count - a.count);
});

const filteredApps = computed(() => {
	if (!selectedGroup.value) return appsWithLabels.value;
	return appsWithLabels.value.filter(
		(app) => app.labelsKey === selectedGroup.value
	);
});

const selectedApp = computed(() => {
	if (selectedIndex.value === null) return null;
	return filteredApps.value[selectedIndex.value];
});

const activeTabName = computed(
	() => tabs.find((t) => t.key === activeTab.value)?.name || ''
);

const activeTabData = computed(() => {
	if (!selectedApp.value) return '';
	const { sourceName, application } = selectedApp.value;
	let data = null;

	switch (activeTab.value) {
		case 'simple':
			data = appStore.getAppSimpleInfo(application, sourceName);
			break;
		case 'status':
			data = appStore.getAppStatus(application, sourceName);
			break;
		case 'full':
			data = appStore.getAppFullInfo(application, sourceName);
			break;
	}

	try {
		return JSON.stringify(data, null, 2);
	} catch {
		return String(data);
	}
});

const getLabelStyle = (label: string) => {
	const lower = label.toLowerCase();
	const config = LABEL_CONFIG[lower] || DEFAULT_LABEL_STYLE;
	return {
		backgroundColor: config.bgColor,
		color: config.textColor
	};
};

const getTitle = (appInfo: any) => {
	return appInfo?.app_title?.['zh-CN'] || appInfo?.app_title?.['en-US'] || '-';
};

const copyData = async () => {
	if (!activeTabData.value) return;
	await navigator.clipboard.writeText(activeTabData.value);
	copied.value = true;
	setTimeout(() => (copied.value = false), 1500);
};
</script>

<style scoped>
.labels-viewer {
	height: 100%;
	display: flex;
	flex-direction: column;
	background: #fff;
}

.toolbar {
	height: 48px;
	background: #fff;
	border-bottom: 1px solid #e2e8f0;
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 0 16px;
	flex-shrink: 0;
}

.title {
	font-size: 14px;
	font-weight: 600;
	color: #1e293b;
}

.stats {
	font-size: 12px;
	color: #64748b;
}

.labels-area {
	display: flex;
	gap: 16px;
	padding: 12px 16px;
	background: #f8fafc;
	border-bottom: 1px solid #e2e8f0;
	flex-shrink: 0;
}

.label-groups {
	flex: 1;
}

.group-header,
.single-header {
	font-size: 11px;
	font-weight: 600;
	color: #64748b;
	text-transform: uppercase;
	margin-bottom: 8px;
}

.group-list {
	display: flex;
	flex-wrap: wrap;
	gap: 6px;
}

.group-btn {
	display: flex;
	align-items: center;
	gap: 8px;
	padding: 6px 10px;
	border: 1px solid #e2e8f0;
	border-radius: 6px;
	background: #fff;
	font-size: 11px;
	cursor: pointer;
	transition: all 0.15s;
}

.group-btn:hover {
	border-color: #6366f1;
}

.group-btn.active {
	background: #6366f1;
	color: #fff;
	border-color: #6366f1;
}

.group-labels {
	display: flex;
	gap: 4px;
}

.group-count {
	background: rgba(0, 0, 0, 0.1);
	padding: 2px 6px;
	border-radius: 10px;
	font-size: 10px;
}

.group-btn.active .group-count {
	background: rgba(255, 255, 255, 0.3);
}

.single-labels {
	width: 200px;
	flex-shrink: 0;
}

.single-list {
	display: flex;
	flex-direction: column;
	gap: 4px;
}

.single-item {
	display: flex;
	justify-content: space-between;
	align-items: center;
	padding: 4px 8px;
	background: #fff;
	border-radius: 4px;
	border: 1px solid #e2e8f0;
}

.single-count {
	font-size: 11px;
	font-weight: 600;
	color: #64748b;
}

.label-tag {
	padding: 2px 8px;
	border-radius: 10px;
	font-size: 10px;
	font-weight: 600;
}

.label-tag.small {
	padding: 1px 6px;
	font-size: 9px;
}

.group-btn.active .label-tag {
	background: rgba(255, 255, 255, 0.3) !important;
	color: #fff !important;
}

.result-area {
	flex: 1;
	display: flex;
	min-height: 0;
}

.app-list {
	width: 380px;
	border-right: 1px solid #e2e8f0;
	display: flex;
	flex-direction: column;
	min-height: 0;
}

.list-header {
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

.list-header span {
	color: #6366f1;
}

.list-body {
	flex: 1;
	overflow-y: auto;
	padding: 6px;
}

.list-body::-webkit-scrollbar {
	width: 4px;
}

.list-body::-webkit-scrollbar-thumb {
	background: #cbd5e1;
	border-radius: 2px;
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

.app-main {
	display: flex;
	justify-content: space-between;
	align-items: center;
	margin-bottom: 4px;
}

.app-name {
	font-size: 12px;
	font-weight: 600;
	color: #1e293b;
}

.app-version {
	font-size: 10px;
	color: #94a3b8;
}

.app-labels {
	display: flex;
	gap: 4px;
	margin-bottom: 4px;
}

.app-title {
	font-size: 11px;
	color: #64748b;
	overflow: hidden;
	text-overflow: ellipsis;
	white-space: nowrap;
}

.empty {
	padding: 30px;
	text-align: center;
	color: #94a3b8;
	font-size: 13px;
}

.app-detail {
	flex: 1;
	display: flex;
	flex-direction: column;
	min-height: 0;
	padding: 12px;
	background: #f8fafc;
}

.placeholder {
	flex: 1;
	display: flex;
	align-items: center;
	justify-content: center;
	color: #94a3b8;
	font-size: 13px;
}

.detail-content {
	flex: 1;
	display: flex;
	flex-direction: column;
	gap: 12px;
	min-height: 0;
}

.detail-summary {
	display: grid;
	grid-template-columns: repeat(4, 1fr);
	gap: 10px;
	background: #fff;
	padding: 14px;
	border-radius: 8px;
	border: 1px solid #e2e8f0;
	flex-shrink: 0;
}

.summary-row {
	display: flex;
	flex-direction: column;
	gap: 2px;
}

.summary-row .label {
	font-size: 10px;
	font-weight: 600;
	color: #64748b;
	text-transform: uppercase;
}

.summary-row .value {
	font-size: 12px;
	font-weight: 500;
	color: #1e293b;
	word-break: break-all;
}

.summary-row .value.labels {
	display: flex;
	gap: 4px;
	flex-wrap: wrap;
}

.detail-tabs {
	display: flex;
	gap: 6px;
	flex-shrink: 0;
}

.tab-btn {
	padding: 6px 14px;
	border: 1px solid #e2e8f0;
	border-radius: 6px;
	background: #fff;
	font-size: 11px;
	cursor: pointer;
	transition: all 0.15s;
}

.tab-btn:hover {
	border-color: #6366f1;
}

.tab-btn.active {
	background: #6366f1;
	color: #fff;
	border-color: #6366f1;
}

.detail-json {
	flex: 1;
	display: flex;
	flex-direction: column;
	background: #fff;
	border: 1px solid #e2e8f0;
	border-radius: 8px;
	overflow: hidden;
	min-height: 0;
}

.json-header {
	padding: 8px 12px;
	background: #f8fafc;
	border-bottom: 1px solid #e2e8f0;
	display: flex;
	justify-content: space-between;
	align-items: center;
	font-size: 11px;
	font-weight: 600;
	color: #64748b;
	flex-shrink: 0;
}

.copy-btn {
	padding: 3px 8px;
	border: 1px solid #e2e8f0;
	border-radius: 4px;
	background: #fff;
	font-size: 10px;
	cursor: pointer;
}

.copy-btn:hover {
	border-color: #6366f1;
}

.json-body {
	flex: 1;
	overflow: auto;
	padding: 12px;
	margin: 0;
	font-size: 10px;
	line-height: 1.5;
	color: #334155;
	white-space: pre-wrap;
	word-break: break-word;
	font-family: 'SF Mono', 'Fira Code', monospace;
}

.json-body::-webkit-scrollbar {
	width: 4px;
}

.json-body::-webkit-scrollbar-thumb {
	background: #cbd5e1;
	border-radius: 2px;
}
</style>
