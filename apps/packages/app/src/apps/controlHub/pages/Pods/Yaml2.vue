<template>
	<q-btn @click="yamlShow" outline color="primary" flat>
		<slot></slot>
	</q-btn>
	<q-dialog
		persistent
		full-width
		full-height
		v-model="visible2"
		@show="show"
		@hide="hide"
	>
		<q-card class="dialog-margin-2">
			<q-card-section class="row items-center q-pb-none">
				<div class="text-h6">{{}}</div>
				<q-space />
				<q-btn icon="close" flat round dense v-close-popup />
			</q-card-section>
			<v-ace-editor
				v-if="aceVisileb"
				v-model:value="yamlData"
				lang="yaml"
				theme="textmate"
				readonly
				style="height: calc(100% - 60px)"
				:options="{
					showGutter: true,
					showPrintMargin: true,
					useWorker: true,
					wrap: true
				}"
			/>
		</q-card>
	</q-dialog>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue';
import { VAceEditor } from 'vue3-ace-editor';
import ace from 'ace-builds';
import 'ace-builds/src-noconflict/mode-yaml';
import 'ace-builds/src-noconflict/snippets/yaml';
import 'ace-builds/src-noconflict/theme-textmate';

import 'ace-builds/src-noconflict/mode-groovy';
import 'ace-builds/src-noconflict/theme-chaos';
import 'ace-builds/src-noconflict/keybinding-vscode';
import 'ace-builds/src-noconflict/ext-searchbox';

// eslint-disable-next-line @typescript-eslint/ban-ts-comment
//@ts-ignore
import workerJsonUrl from 'file-loader?esModule=false!ace-builds/src-noconflict/worker-yaml.js';
import { objectToYaml } from './yaml';
ace.config.setModuleUrl('ace/mode/yaml_worker', workerJsonUrl);
// @apps/control-panel-common/src/components/Modals/EditYaml/index.jsx

interface Props {
	value: string;
}
const props = withDefaults(defineProps<Props>(), {});

const data = ref();
const visible2 = ref(false);
const aceVisileb = ref(false);

const yamlData = objectToYaml(props.value);
console.log('yamlData', props.value);
// objectToYaml(props.value);
const yamlShow = () => {
	visible2.value = true;
};

const show = (evt: Event) => {
	aceVisileb.value = true;
};

const hide = (evt: Event) => {
	aceVisileb.value = false;
};
</script>

<style></style>
