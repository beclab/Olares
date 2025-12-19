<template>
	<v-ace-editor
		v-model:value="data"
		:lang="lang"
		theme="chaos"
		:readonly="readonly"
		style="height: 100%"
		wrap
		v-show="data"
		:options="{
			showGutter: true,
			showPrintMargin: false,
			useWorker: false,
			keyboardHandler: 'vscode',
			wrapEnabled: true,
			tabSize: 2,
			wrap: true
		}"
	/>
</template>

<script setup lang="ts">
import { useQuasar } from 'quasar';
import { getYaml } from '@apps/control-panel-common/src/network';
import { onMounted, ref, toRef, toRefs } from 'vue';
import { VAceEditor } from 'vue3-ace-editor';
import ace from 'ace-builds';
import 'ace-builds/src-noconflict/mode-yaml';
import 'ace-builds/src-noconflict/snippets/yaml';
import 'ace-builds/src-noconflict/theme-textmate';
import 'ace-builds/src-noconflict/theme-chaos';

import 'ace-builds/src-noconflict/mode-groovy';
import 'ace-builds/src-noconflict/keybinding-vscode';
import 'ace-builds/src-noconflict/ext-searchbox';
import 'ace-builds/src-noconflict/mode-javascript';
import 'ace-builds/src-noconflict/snippets/javascript';

// eslint-disable-next-line @typescript-eslint/ban-ts-comment
//@ts-ignore
import workerJsonUrl from 'file-loader?esModule=false!ace-builds/src-noconflict/worker-yaml.js';
ace.config.setModuleUrl('ace/mode/yaml_worker', workerJsonUrl);
// src/components/Modals/EditYaml/index.jsx

interface Props {
	data: string;
	readonly?: boolean;
	theme?: 'textmate' | 'chaos';
	lang?: 'yaml' | 'javascript';
}
const props = withDefaults(defineProps<Props>(), {
	readonly: true,
	theme: 'chaos',
	lang: 'yaml'
});

const $q = useQuasar();

const { data } = toRefs(props);
</script>

<style></style>
