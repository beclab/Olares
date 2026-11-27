<template>
	<div class="map-viewer">
		<header class="toolbar">
			<div class="title">📊 {{ t('debug.mapViewer.title') }}</div>
			<div class="actions">
				<button class="btn" @click="refreshData">
					🔄 {{ t('debug.common.refresh') }}
				</button>
				<button class="btn primary" @click="exportAll">
					📥 {{ t('debug.common.export') }}
				</button>
			</div>
		</header>

		<div class="main">
			<DebugSidebar
				:map-options="mapOptions"
				:selected-map="selectedMap"
				:selected-prefix="selectedPrefix"
				:search-query="searchQuery"
				:selected-key="selectedKey"
				:all-keys="allKeys"
				:get-map-size="getMapSize"
				@select-map="selectMap"
				@select-key="selectKey"
				@update:selected-prefix="selectedPrefix = $event"
				@update:search-query="searchQuery = $event"
			/>

			<DebugJsonView :selected-key="selectedKey" :value="selectedValue" />
		</div>
	</div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue';
import { useI18n } from 'vue-i18n';
import { useAppStore } from '../../../stores/market/appStore';
import DebugSidebar from './DebugSidebar.vue';
import DebugJsonView from './DebugJsonView.vue';

const appStore = useAppStore();
const { t } = useI18n();

const selectedMap = ref('localStatusMap');
const selectedPrefix = ref('');
const searchQuery = ref('');
const selectedKey = ref<string | null>(null);

const mapOptions = [
	{ key: 'localStatusMap', name: 'Local', icon: '📍' },
	{ key: 'appSimpleInfoMap', name: 'Simple', icon: '📝' },
	{ key: 'appStatusMap', name: 'Status', icon: '🔄' },
	{ key: 'appFullInfoMap', name: 'Full', icon: '📦' }
];

const mapRegistry = computed(() => ({
	localStatusMap: appStore.localStatusMap,
	appSimpleInfoMap: appStore.appSimpleInfoMap,
	appStatusMap: appStore.appStatusMap,
	appFullInfoMap: appStore.appFullInfoMap
}));

const currentMap = computed(
	() =>
		mapRegistry.value[selectedMap.value as keyof typeof mapRegistry.value] ||
		new Map()
);

const allKeys = computed(() => Array.from(currentMap.value.keys()));

const selectedValue = computed(() => {
	if (!selectedKey.value) return null;
	return currentMap.value.get(selectedKey.value);
});

const getMapSize = (key: string) =>
	mapRegistry.value[key as keyof typeof mapRegistry.value]?.size || 0;

const selectMap = (key: string) => {
	selectedMap.value = key;
	selectedPrefix.value = '';
	selectedKey.value = null;
	searchQuery.value = '';
};

const selectKey = (key: string) => {
	selectedKey.value = key;
};

const refreshData = () => {
	const cur = selectedMap.value;
	selectedMap.value = '';
	setTimeout(() => (selectedMap.value = cur), 0);
};

const exportAll = () => {
	const data: Record<string, any> = {};
	currentMap.value.forEach((v, k) => (data[k] = v));
	const blob = new Blob([JSON.stringify(data, null, 2)], {
		type: 'application/json'
	});
	const a = document.createElement('a');
	a.href = URL.createObjectURL(blob);
	a.download = `${selectedMap.value}_${Date.now()}.json`;
	a.click();
};
</script>

<style scoped>
.map-viewer {
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

.actions {
	display: flex;
	gap: 8px;
}

.btn {
	padding: 6px 12px;
	border: 1px solid #e2e8f0;
	border-radius: 6px;
	background: #fff;
	font-size: 11px;
	cursor: pointer;
	transition: all 0.15s;
}

.btn:hover {
	border-color: #6366f1;
	color: #6366f1;
}

.btn.primary {
	background: #6366f1;
	color: #fff;
	border-color: #6366f1;
}

.btn.primary:hover {
	background: #4f46e5;
}

.main {
	flex: 1;
	display: flex;
	min-height: 0;
}
</style>
