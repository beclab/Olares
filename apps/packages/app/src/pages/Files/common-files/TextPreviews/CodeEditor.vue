<template>
	<v-ace-editor
		:value="value"
		:lang="editorLang"
		:theme="editorTheme"
		:readonly="readonly"
		:options="editorOptions"
		:class="{ 'code-editor--hide-cursor': hideCursor }"
		@update:value="(val: string) => emit('update:value', val)"
		@init="onInit"
	/>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { useQuasar } from 'quasar';
import { VAceEditor } from 'vue3-ace-editor';
import ace from 'ace-builds';
import 'ace-builds/webpack-resolver';
import 'ace-builds/src-noconflict/theme-chrome';
import 'ace-builds/src-noconflict/theme-tomorrow_night';
import 'ace-builds/src-noconflict/ext-modelist';
import 'ace-builds/src-noconflict/ext-language_tools';
import type { Ace } from 'ace-builds';

const props = defineProps({
	value: {
		type: String,
		default: ''
	},
	// File name used to detect the syntax highlighting language.
	name: {
		type: String,
		default: ''
	},
	readonly: {
		type: Boolean,
		default: false
	},
	showGutter: {
		type: Boolean,
		default: true
	},
	autocomplete: {
		type: Boolean,
		default: false
	},
	hideCursor: {
		type: Boolean,
		default: false
	},
	// Extra Ace options merged on top of (and overriding) the defaults.
	options: {
		type: Object,
		default: () => ({})
	}
});

const emit = defineEmits<{
	(e: 'update:value', value: string): void;
	(e: 'init', editor: Ace.Editor): void;
}>();

const $q = useQuasar();
const modelist = ace.require('ace/ext/modelist');

const editorLang = computed(() => {
	const mode: string = modelist.getModeForPath(props.name || '').mode;
	return mode.replace('ace/mode/', '');
});

const editorTheme = computed(() =>
	$q.dark.isActive ? 'tomorrow_night' : 'chrome'
);

const editorOptions = computed(() => ({
	showGutter: props.showGutter,
	showPrintMargin: false,
	useWorker: false,
	wrapEnabled: true,
	tabSize: 2,
	wrap: true,
	enableBasicAutocompletion: props.autocomplete,
	enableLiveAutocompletion: props.autocomplete,
	enableSnippets: props.autocomplete,
	...props.options
}));

const onInit = (editor: Ace.Editor) => {
	emit('init', editor);
};
</script>

<style scoped lang="scss">
.code-editor--hide-cursor {
	:deep(.ace_cursor) {
		display: none !important;
	}

	:deep(.ace_cursor-layer) {
		display: none !important;
	}
}
</style>
