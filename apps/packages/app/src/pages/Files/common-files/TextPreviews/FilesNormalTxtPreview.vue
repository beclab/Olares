<template>
	<code-editor
		v-if="hasContent"
		:value="text"
		:name="name"
		:readonly="true"
		:show-gutter="false"
		hide-cursor
		:options="previewOptions"
		class="info-content"
	/>
	<div v-else class="info-empty column items-center justify-center">
		<div class="text-body3 text-ink-3">{{ t('files.no_content') }}</div>
	</div>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';
import CodeEditor from './CodeEditor.vue';

const { t } = useI18n();

const props = defineProps({
	content: {
		type: [String, Object],
		default: '',
		required: false
	},
	name: {
		type: String,
		default: '',
		required: false
	}
});

const text = computed(() => {
	const value = props.content;
	if (value == null) return '';
	return typeof value === 'object'
		? JSON.stringify(value, null, 2)
		: String(value);
});

const hasContent = computed(() => text.value.trim().length > 0);

const previewOptions = {
	highlightActiveLine: false,
	highlightGutterLine: false,
	highlightSelectedWord: false,
	displayIndentGuides: false
};
</script>

<style scoped lang="scss">
.info-content {
	width: 100%;
	height: 100%;
}

.info-empty {
	width: 100%;
	height: 100%;
}
</style>
