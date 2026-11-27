<template>
	<aside class="sidebar">
		<div class="map-row">
			<button
				v-for="map in mapOptions"
				:key="map.key"
				:class="['map-btn', { active: selectedMap === map.key }]"
				@click="emit('selectMap', map.key)"
			>
				{{ map.icon }} {{ map.name }}
				<span class="count">{{ getMapSize(map.key) }}</span>
			</button>
		</div>

		<div class="prefix-row" v-if="prefixList.length > 1">
			<button
				:class="['prefix-btn', { active: selectedPrefix === '' }]"
				@click="emit('update:selectedPrefix', '')"
			>
				{{ t('debug.common.all') }} <span>{{ allKeys.length }}</span>
			</button>
			<button
				v-for="p in prefixList"
				:key="p.prefix"
				:class="['prefix-btn', { active: selectedPrefix === p.prefix }]"
				@click="emit('update:selectedPrefix', p.prefix)"
			>
				{{ p.prefix }} <span>{{ p.count }}</span>
			</button>
		</div>

		<div class="search-row">
			<input
				:value="searchQuery"
				@input="
					emit('update:searchQuery', ($event.target as HTMLInputElement).value)
				"
				type="text"
				:placeholder="`🔍 ${t('debug.sidebar.searchPlaceholder')}`"
			/>
			<button
				v-if="searchQuery"
				class="clear"
				@click="emit('update:searchQuery', '')"
			>
				✕
			</button>
		</div>

		<div class="keys-panel">
			<div class="keys-header">
				Keys <span>{{ filteredKeys.length }} / {{ allKeys.length }}</span>
			</div>
			<div class="keys-list">
				<div
					v-for="key in filteredKeys"
					:key="key"
					:class="['key-item', { active: selectedKey === key }]"
					@click="emit('selectKey', key)"
				>
					{{ key }}
				</div>
				<div v-if="filteredKeys.length === 0" class="empty">
					{{ t('debug.sidebar.noMatchedKey') }}
				</div>
			</div>
		</div>
	</aside>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';

interface MapOption {
	key: string;
	name: string;
	icon: string;
}

const props = defineProps<{
	mapOptions: MapOption[];
	selectedMap: string;
	selectedPrefix: string;
	searchQuery: string;
	selectedKey: string | null;
	allKeys: string[];
	getMapSize: (key: string) => number;
}>();

const { t } = useI18n();

const emit = defineEmits<{
	selectMap: [key: string];
	selectKey: [key: string];
	'update:selectedPrefix': [value: string];
	'update:searchQuery': [value: string];
}>();

const prefixList = computed(() => {
	const prefixMap = new Map<string, number>();
	props.allKeys.forEach((key) => {
		const prefix = key.split('_')[0];
		if (prefix) prefixMap.set(prefix, (prefixMap.get(prefix) || 0) + 1);
	});
	return Array.from(prefixMap.entries())
		.map(([prefix, count]) => ({ prefix, count }))
		.sort((a, b) => b.count - a.count);
});

const filteredKeys = computed(() => {
	let keys = props.allKeys;
	if (props.selectedPrefix) {
		keys = keys.filter((k) => k.split('_')[0] === props.selectedPrefix);
	}
	if (props.searchQuery) {
		const q = props.searchQuery.toLowerCase();
		keys = keys.filter((k) => k.toLowerCase().includes(q));
	}
	return keys;
});
</script>

<style scoped>
.sidebar {
	width: 400px;
	background: #f8fafc;
	border-right: 1px solid #e2e8f0;
	display: flex;
	flex-direction: column;
	min-height: 0;
}

.map-row {
	display: flex;
	gap: 6px;
	padding: 10px;
	border-bottom: 1px solid #e2e8f0;
	flex-shrink: 0;
}

.map-btn {
	flex: 1;
	padding: 8px 6px;
	border: 1px solid #e2e8f0;
	border-radius: 6px;
	background: #fff;
	font-size: 10px;
	cursor: pointer;
	display: flex;
	align-items: center;
	justify-content: center;
	gap: 3px;
	transition: all 0.15s;
}

.map-btn:hover {
	border-color: #6366f1;
}

.map-btn.active {
	background: #6366f1;
	color: #fff;
	border-color: #6366f1;
}

.map-btn .count {
	background: rgba(0, 0, 0, 0.1);
	padding: 1px 4px;
	border-radius: 6px;
	font-size: 9px;
}

.map-btn.active .count {
	background: rgba(255, 255, 255, 0.3);
}

.prefix-row {
	display: flex;
	gap: 5px;
	padding: 8px 10px;
	border-bottom: 1px solid #e2e8f0;
	overflow-x: auto;
	flex-shrink: 0;
}

.prefix-row::-webkit-scrollbar {
	height: 3px;
}

.prefix-row::-webkit-scrollbar-thumb {
	background: #cbd5e1;
	border-radius: 2px;
}

.prefix-btn {
	padding: 5px 8px;
	border: 1px solid #e2e8f0;
	border-radius: 12px;
	background: #fff;
	font-size: 10px;
	cursor: pointer;
	white-space: nowrap;
	display: flex;
	align-items: center;
	gap: 3px;
	flex-shrink: 0;
	transition: all 0.15s;
}

.prefix-btn:hover {
	border-color: #10b981;
	background: #f0fdf4;
}

.prefix-btn.active {
	background: #10b981;
	color: #fff;
	border-color: #10b981;
}

.prefix-btn span {
	background: rgba(0, 0, 0, 0.08);
	padding: 1px 4px;
	border-radius: 6px;
	font-size: 9px;
}

.prefix-btn.active span {
	background: rgba(255, 255, 255, 0.3);
}

.search-row {
	position: relative;
	padding: 8px 10px;
	border-bottom: 1px solid #e2e8f0;
	flex-shrink: 0;
}

.search-row input {
	width: 100%;
	padding: 7px 28px 7px 10px;
	border: 1px solid #e2e8f0;
	border-radius: 5px;
	font-size: 11px;
	outline: none;
	box-sizing: border-box;
}

.search-row input:focus {
	border-color: #6366f1;
}

.search-row .clear {
	position: absolute;
	right: 16px;
	top: 50%;
	transform: translateY(-50%);
	border: none;
	background: #e2e8f0;
	width: 16px;
	height: 16px;
	border-radius: 50%;
	cursor: pointer;
	font-size: 9px;
	display: flex;
	align-items: center;
	justify-content: center;
}

.search-row .clear:hover {
	background: #f87171;
	color: #fff;
}

.keys-panel {
	flex: 1;
	display: flex;
	flex-direction: column;
	min-height: 0;
}

.keys-header {
	padding: 8px 10px;
	font-size: 10px;
	font-weight: 600;
	color: #64748b;
	text-transform: uppercase;
	display: flex;
	justify-content: space-between;
	border-bottom: 1px solid #f1f5f9;
	flex-shrink: 0;
}

.keys-header span {
	color: #94a3b8;
	font-weight: 400;
}

.keys-list {
	flex: 1;
	overflow-y: auto;
	padding: 5px;
}

.keys-list::-webkit-scrollbar {
	width: 4px;
}

.keys-list::-webkit-scrollbar-thumb {
	background: #cbd5e1;
	border-radius: 2px;
}

.key-item {
	padding: 7px 8px;
	border-radius: 4px;
	cursor: pointer;
	font-size: 11px;
	color: #334155;
	word-break: break-all;
	margin-bottom: 1px;
	transition: background 0.1s;
}

.key-item:hover {
	background: #e2e8f0;
}

.key-item.active {
	background: #6366f1;
	color: #fff;
}

.empty {
	padding: 20px;
	text-align: center;
	color: #94a3b8;
	font-size: 12px;
}
</style>
