<template>
	<div v-if="!isAuthorized" class="not-found">
		<h1>404</h1>
		<p>Page Not Found</p>
	</div>

	<div v-else class="debug-page">
		<nav class="nav-sidebar">
			<div class="nav-header">🔬 Debug</div>
			<div class="nav-list">
				<div
					v-for="item in navItems"
					:key="item.id"
					:class="['nav-item', { active: activeNav === item.id }]"
					@click="activeNav = item.id"
				>
					<span class="nav-icon">{{ item.icon }}</span>
					<span class="nav-text">{{ item.name }}</span>
				</div>
			</div>

			<div class="flags-panel">
				<div class="flags-header">🧪 Flags</div>
				<div class="flag-item" title="bypassV2Install">
					<div class="flag-label">Allow v2 install</div>
					<bt-switch
						size="xs"
						truthy-track-color="orange-default"
						v-model="bypassV2Install"
					/>
				</div>
				<div v-if="bypassV2Install" class="flag-warning">
					⚠️ v2 install gate is OFF for this browser.
				</div>
				<div class="flag-item" title="bypassV2Upgrade">
					<div class="flag-label">Allow v2 upgrade</div>
					<bt-switch
						size="xs"
						truthy-track-color="orange-default"
						v-model="bypassV2Upgrade"
					/>
				</div>
				<div v-if="bypassV2Upgrade" class="flag-warning">
					⚠️ v2 upgrade gate is OFF for this browser.
				</div>
			</div>
		</nav>

		<div class="content-area">
			<MapViewer v-if="activeNav === 'map'" />
			<AppSearch v-else-if="activeNav === 'search'" />
			<LabelsViewer v-else-if="activeNav === 'labels'" />
			<StatusMonitor v-else-if="activeNav === 'monitor'" />
			<AppEntryDiffViewer v-else-if="activeNav === 'entry-diff'" />
		</div>
	</div>
</template>

<script setup lang="ts">
import {
	readMarketDebugFlags,
	writeMarketDebugFlags
} from '../../../services/market/marketDebugFlags';
import { computed, ref, watch, onMounted } from 'vue';
import { useRoute } from 'vue-router';
import MapViewer from './MapViewer.vue';
import AppSearch from './AppSearch.vue';
import LabelsViewer from './LabelsViewer.vue';
import StatusMonitor from './StatusMonitor.vue';
import AppEntryDiffViewer from './AppEntryDiffViewer.vue';
import { useI18n } from 'vue-i18n';

const route = useRoute();
const { t } = useI18n();
const isAuthorized = ref(false);
const activeNav = ref('monitor');

const bypassV2Install = ref(readMarketDebugFlags().bypassV2Install);
watch(bypassV2Install, (next) => {
	writeMarketDebugFlags({ ...readMarketDebugFlags(), bypassV2Install: next });
});

const bypassV2Upgrade = ref(readMarketDebugFlags().bypassV2Upgrade);
watch(bypassV2Upgrade, (next) => {
	writeMarketDebugFlags({ ...readMarketDebugFlags(), bypassV2Upgrade: next });
});

const navItems = computed(() => [
	{ id: 'monitor', name: t('debug.panel.nav.monitor'), icon: '📡' },
	{ id: 'map', name: t('debug.panel.nav.map'), icon: '📊' },
	{ id: 'search', name: t('debug.panel.nav.search'), icon: '🔍' },
	{ id: 'labels', name: t('debug.panel.nav.labels'), icon: '🏷️' },
	{ id: 'entry-diff', name: t('debug.panel.nav.entryDiff'), icon: '🧪' }
]);

onMounted(() => {
	if (route.query._dm === 'Olares') {
		isAuthorized.value = true;
	}
});
</script>

<style scoped>
* {
	box-sizing: border-box;
	margin: 0;
	padding: 0;
}

.not-found {
	height: 100vh;
	display: flex;
	flex-direction: column;
	align-items: center;
	justify-content: center;
	background: #f1f5f9;
	font-family: system-ui, sans-serif;
}
.not-found h1 {
	font-size: 100px;
	color: #cbd5e1;
}
.not-found p {
	color: #94a3b8;
}

.debug-page {
	height: 100vh;
	display: flex;
	background: #f1f5f9;
	font-family: 'SF Mono', 'Fira Code', monospace;
	overflow: hidden;
}

.nav-sidebar {
	width: 130px;
	background: #1e293b;
	display: flex;
	flex-direction: column;
	flex-shrink: 0;
}
.nav-header {
	padding: 14px 12px;
	font-size: 13px;
	font-weight: 700;
	color: #fff;
	border-bottom: 1px solid #334155;
}
.nav-list {
	flex: 1;
	padding: 8px;
	overflow-y: auto;
}

.flags-panel {
	border-top: 1px solid #334155;
	padding: 8px 10px 10px;
	color: #94a3b8;
	font-size: 10px;
	flex-shrink: 0;
}

.flags-header {
	font-weight: 600;
	text-transform: uppercase;
	letter-spacing: 0.04em;
	color: #cbd5e1;
	margin-bottom: 6px;
	opacity: 0.7;
}

.flag-item {
	display: flex;
	align-items: center;
	justify-content: space-between;
	gap: 6px;
	line-height: 1.3;
	user-select: none;
}

.flag-label {
	color: #cbd5e1;
}

.flag-warning {
	margin-top: 8px;
	padding: 4px 6px;
	background: rgba(245, 158, 11, 0.15);
	color: #fbbf24;
	border-radius: 4px;
	font-size: 9px;
	line-height: 1.3;
}
.nav-item {
	display: flex;
	align-items: center;
	gap: 6px;
	padding: 10px;
	border-radius: 6px;
	cursor: pointer;
	color: #94a3b8;
	font-size: 11px;
	margin-bottom: 4px;
	transition: all 0.15s;
}
.nav-item:hover {
	background: #334155;
	color: #fff;
}
.nav-item.active {
	background: #6366f1;
	color: #fff;
}
.nav-icon {
	font-size: 13px;
}

.content-area {
	flex: 1;
	display: flex;
	flex-direction: column;
	min-height: 0;
	overflow: hidden;
}
</style>
