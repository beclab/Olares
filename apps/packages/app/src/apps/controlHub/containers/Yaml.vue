<template>
	<v-ace-editor
		v-model:value="data"
		lang="yaml"
		:theme="$q.dark.isActive ? 'chaos' : 'textmate'"
		readonly
		style="height: 100%"
		:options="{
			showGutter: true,
			showPrintMargin: false,
			useWorker: true,
			keyboardHandler: 'vscode',
			wrapEnabled: true,
			tabSize: 2,
			wrap: true
		}"
	/>
</template>

<script setup lang="ts">
import { getYaml } from '@apps/control-hub/src/network';
import { useQuasar } from 'quasar';
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

// eslint-disable-next-line @typescript-eslint/ban-ts-comment
//@ts-ignore
import workerJsonUrl from 'file-loader?esModule=false!ace-builds/src-noconflict/worker-yaml.js';
ace.config.setModuleUrl('ace/mode/yaml_worker', workerJsonUrl);
// src/components/Modals/EditYaml/index.jsx

interface Props {
	data: string;
}
const props = withDefaults(defineProps<Props>(), {});
const $q = useQuasar();
const { data } = toRefs(props);
</script>
