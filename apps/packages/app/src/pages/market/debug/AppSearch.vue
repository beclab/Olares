<template>
	<div class="app-search">
		<header class="toolbar">
			<div class="title">🔍 {{ t('debug.appSearch.title') }}</div>
		</header>

		<div class="search-area">
			<div class="search-form">
				<div class="form-group">
					<label>{{ t('debug.appSearch.keyword') }}</label>
					<input
						v-model="searchContent"
						type="text"
						:placeholder="t('debug.appSearch.placeholder')"
						@input="handleSearch"
					/>
				</div>
				<div class="search-info">
					<span v-if="searchContent">{{ foundText }}</span>
				</div>
			</div>
		</div>

		<div class="result-area">
			<div v-if="!searchContent" class="placeholder">
				{{ t('debug.appSearch.emptyKeyword') }}
			</div>

			<div v-else-if="results.length === 0" class="no-result">
				❌ {{ t('debug.appSearch.noMatch') }}
			</div>

			<div v-else class="result-content">
				<div class="result-list">
					<div class="list-header">
						{{ t('debug.appSearch.resultList') }}
						<span>{{ results.length }}</span>
					</div>
					<div class="list-body">
						<div
							v-for="(item, index) in results"
							:key="item.key"
							:class="['result-item', { active: selectedIndex === index }]"
							@click="selectResult(index)"
						>
							<div class="item-main">
								<span class="item-name">{{ item.appInfo.app_name }}</span>
								<span class="item-weight">{{ item.weight }}</span>
							</div>
							<div class="item-sub">
								<span class="item-title">{{ getTitle(item.appInfo) }}</span>
								<span class="item-source">{{ item.sourceName }}</span>
							</div>
						</div>
					</div>
				</div>

				<div class="result-detail">
					<div v-if="selectedIndex === null" class="placeholder">
						👈 {{ t('debug.appSearch.selectHint') }}
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
								<span class="value">{{
									selectedApp?.appInfo.app_version
								}}</span>
							</div>
							<div class="summary-row">
								<span class="label">{{ t('debug.appSearch.weight') }}</span>
								<span class="value weight">{{ selectedApp?.weight }}</span>
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
	</div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useCenterStore } from '../../../stores/market/center';
import { useAppStore } from '../../../stores/market/appStore';

const centerStore = useCenterStore();
const appStore = useAppStore();
const { t } = useI18n();

const searchContent = ref('');
const results = ref<any[]>([]);
const selectedIndex = ref<number | null>(null);
const activeTab = ref('simple');
const copied = ref(false);

const tabs = [
	{ key: 'simple', name: 'Simple Info', icon: '📝' },
	{ key: 'status', name: 'Status', icon: '🔄' },
	{ key: 'full', name: 'Full Info', icon: '📦' }
];

const selectedApp = computed(() => {
	if (selectedIndex.value === null) return null;
	return results.value[selectedIndex.value];
});

const activeTabName = computed(
	() => tabs.find((t) => t.key === activeTab.value)?.name || ''
);
const foundText = computed(() =>
	t('debug.appSearch.found', { count: results.value.length })
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

const calculateWeight = (appInfo: any, searchText: string) => {
	if (!searchText) return 0;
	const lowerSearch = searchText.toLowerCase();
	let weight = 0;

	const WEIGHTS = {
		nameExact: 150,
		nameInclude: 50,
		titleExact: 100,
		titleInclude: 20,
		descExact: 15,
		descInclude: 5,
		fullDescExact: 10,
		fullDescInclude: 3
	};

	const appName = appInfo?.app_name?.toLowerCase();
	if (appName === lowerSearch) {
		weight += WEIGHTS.nameExact;
	} else if (appName?.includes(lowerSearch)) {
		weight += WEIGHTS.nameInclude;
	}

	let titleMatched = false;
	const titles = [
		appInfo?.app_title?.['en-US']?.toLowerCase(),
		appInfo?.app_title?.['zh-CN']?.toLowerCase()
	].filter(Boolean);

	if (titles.some((title) => title === lowerSearch)) {
		weight += WEIGHTS.titleExact;
		titleMatched = true;
	} else if (
		!titleMatched &&
		titles.some((title) => title?.includes(lowerSearch))
	) {
		weight += WEIGHTS.titleInclude;
	}

	let descMatched = false;
	const descs = [
		appInfo?.app_description?.['en-US']?.toLowerCase(),
		appInfo?.app_description?.['zh-CN']?.toLowerCase()
	].filter(Boolean);

	if (descs.some((desc) => desc === lowerSearch)) {
		weight += WEIGHTS.descExact;
		descMatched = true;
	} else if (
		!descMatched &&
		descs.some((desc) => desc?.includes(lowerSearch))
	) {
		weight += WEIGHTS.descInclude;
	}

	let fullDescMatched = false;
	const fullDescs = [
		appInfo?.fullDescription?.['en-US']?.toLowerCase(),
		appInfo?.fullDescription?.['zh-CN']?.toLowerCase()
	].filter(Boolean);

	if (fullDescs.some((fullDesc) => fullDesc === lowerSearch)) {
		weight += WEIGHTS.fullDescExact;
		fullDescMatched = true;
	} else if (
		!fullDescMatched &&
		fullDescs.some((fullDesc) => fullDesc?.includes(lowerSearch))
	) {
		weight += WEIGHTS.fullDescInclude;
	}

	return weight;
};

const search = () => {
	if (!searchContent.value) {
		results.value = [];
		selectedIndex.value = null;
		return;
	}

	results.value = Array.from(appStore.appSimpleInfoMap.entries())
		.map(([key, appInfo]) => {
			const weight = calculateWeight(
				appInfo.app_simple_info,
				searchContent.value
			);
			return {
				key,
				appInfo: appInfo.app_simple_info,
				weight
			};
		})
		.filter(({ weight }) => weight > 0)
		.sort((a, b) => b.weight - a.weight)
		.map(({ key, appInfo, weight }) => {
			const [sourceName, application] = key.split('_');
			return { key, sourceName, application, appInfo, weight };
		});

	selectedIndex.value = null;
};

let searchTimer: ReturnType<typeof setTimeout> | null = null;
const handleSearch = () => {
	if (searchTimer) clearTimeout(searchTimer);
	searchTimer = setTimeout(search, 300);
};

const selectResult = (index: number) => {
	selectedIndex.value = index;
	activeTab.value = 'simple';
};

const getTitle = (appInfo: any) => {
	return (
		appInfo?.app_title?.['zh-CN'] ||
		appInfo?.app_title?.['en-US'] ||
		appInfo?.title ||
		'-'
	);
};

const copyData = async () => {
	if (!activeTabData.value) return;
	await navigator.clipboard.writeText(activeTabData.value);
	copied.value = true;
	setTimeout(() => (copied.value = false), 1500);
};

watch(
	() => appStore.appSimpleInfoMap,
	() => {
		if (searchContent.value) search();
	},
	{ deep: true }
);
</script>

<style scoped>
.app-search {
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
	padding: 0 16px;
	flex-shrink: 0;
}
.title {
	font-size: 14px;
	font-weight: 600;
	color: #1e293b;
}

.search-area {
	padding: 16px;
	background: #f8fafc;
	border-bottom: 1px solid #e2e8f0;
	flex-shrink: 0;
}
.search-form {
	display: flex;
	gap: 16px;
	align-items: flex-end;
}
.form-group {
	display: flex;
	flex-direction: column;
	gap: 4px;
	flex: 1;
	max-width: 400px;
}
.form-group label {
	font-size: 11px;
	font-weight: 600;
	color: #64748b;
	text-transform: uppercase;
}
.form-group input {
	padding: 10px 14px;
	border: 1px solid #e2e8f0;
	border-radius: 8px;
	font-size: 14px;
	outline: none;
}
.form-group input:focus {
	border-color: #6366f1;
}
.search-info {
	font-size: 13px;
	color: #64748b;
}

.result-area {
	flex: 1;
	display: flex;
	min-height: 0;
}

.placeholder,
.no-result {
	flex: 1;
	display: flex;
	align-items: center;
	justify-content: center;
	color: #94a3b8;
	font-size: 14px;
}
.no-result {
	color: #ef4444;
}

.result-content {
	flex: 1;
	display: flex;
	min-height: 0;
}

.result-list {
	width: 350px;
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

.result-item {
	padding: 10px 12px;
	border-radius: 6px;
	cursor: pointer;
	margin-bottom: 4px;
	border: 1px solid transparent;
	transition: all 0.15s;
}
.result-item:hover {
	background: #f1f5f9;
}
.result-item.active {
	background: #eef2ff;
	border-color: #6366f1;
}
.item-main {
	display: flex;
	justify-content: space-between;
	align-items: center;
	margin-bottom: 4px;
}
.item-name {
	font-size: 12px;
	font-weight: 600;
	color: #1e293b;
}
.item-weight {
	font-size: 10px;
	background: #10b981;
	color: #fff;
	padding: 2px 6px;
	border-radius: 10px;
}
.item-sub {
	display: flex;
	justify-content: space-between;
	font-size: 11px;
	color: #64748b;
}
.item-title {
	flex: 1;
	overflow: hidden;
	text-overflow: ellipsis;
	white-space: nowrap;
	margin-right: 8px;
}
.item-source {
	color: #94a3b8;
}

.result-detail {
	flex: 1;
	display: flex;
	flex-direction: column;
	min-height: 0;
	padding: 12px;
	background: #f8fafc;
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
.summary-row .value.weight {
	color: #10b981;
	font-weight: 700;
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
