<template>
	<main class="json-area">
		<div v-if="!selectedKey" class="placeholder">
			👈 {{ t('debug.jsonView.selectKey') }}
		</div>

		<div v-else class="json-view">
			<div class="json-header">
				<span class="current-key">{{ selectedKey }}</span>
				<button class="copy-btn" @click="copyValue">
					{{
						copied
							? `✅ ${t('debug.common.copied')}`
							: `📋 ${t('debug.common.copy')}`
					}}
				</button>
			</div>
			<pre class="json-body">{{ formattedValue }}</pre>
		</div>
	</main>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue';
import { useI18n } from 'vue-i18n';

const props = defineProps<{
	selectedKey: string | null;
	value: any;
}>();

const emit = defineEmits<{
	copied: [];
}>();

const { t } = useI18n();
const copied = ref(false);

const formattedValue = computed(() => {
	if (!props.value) return '';
	try {
		return JSON.stringify(props.value, null, 2);
	} catch {
		return String(props.value);
	}
});

const copyValue = async () => {
	if (!formattedValue.value) return;
	try {
		await navigator.clipboard.writeText(formattedValue.value);
		copied.value = true;
		emit('copied');
		setTimeout(() => {
			copied.value = false;
		}, 1500);
	} catch (e) {
		console.error(t('debug.jsonView.copyFailed'), e);
	}
};
</script>

<style scoped>
.json-area {
	flex: 1;
	display: flex;
	flex-direction: column;
	min-height: 0;
	padding: 10px;
	background: #f1f5f9;
}

.placeholder {
	flex: 1;
	display: flex;
	align-items: center;
	justify-content: center;
	color: #94a3b8;
	font-size: 13px;
}

.json-view {
	flex: 1;
	display: flex;
	flex-direction: column;
	background: #fff;
	border-radius: 8px;
	border: 1px solid #e2e8f0;
	overflow: hidden;
	min-height: 0;
}

.json-header {
	padding: 8px 12px;
	border-bottom: 1px solid #e2e8f0;
	display: flex;
	justify-content: space-between;
	align-items: center;
	flex-shrink: 0;
	background: #f8fafc;
}

.current-key {
	font-size: 11px;
	font-weight: 600;
	color: #6366f1;
	word-break: break-all;
	flex: 1;
	margin-right: 10px;
}

.copy-btn {
	padding: 4px 8px;
	border: 1px solid #e2e8f0;
	border-radius: 4px;
	background: #fff;
	font-size: 10px;
	cursor: pointer;
	white-space: nowrap;
	transition: all 0.15s;
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
	min-height: 0;
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
