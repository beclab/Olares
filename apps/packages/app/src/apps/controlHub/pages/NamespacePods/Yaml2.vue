<template>
	<div class="yaml-container">
		<q-btn
			v-if="$slots.default"
			no-caps
			@click="yamlShow"
			flat
			outline
			color="primary"
		>
			<slot></slot>
		</q-btn>
		<Dialog
			:title="name"
			persistent
			full-width
			full-height
			:ok="true"
			:cancel="true"
			v-model="visible2"
			@show="show"
			@hide="hide"
			@onSubmit="submit"
		>
			<div
				style="
					height: calc(100%);
					border-radius: 6px;
					overflow: hidden;
					position: relative;
				"
			>
				<v-ace-editor
					v-if="aceVisileb"
					v-model:value="data"
					lang="yaml"
					theme="chaos"
					:readonly="readonly || loading2"
					style="height: calc(100%)"
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
				<q-inner-loading :showing="loading" style="z-index: 999999">
				</q-inner-loading>
			</div>
		</Dialog>
	</div>
</template>

<script setup lang="ts">
import {
	getCustomresourceItem,
	getWorkloadsControler,
	getYaml,
	updateCustomResources,
	updateWorkloadsControler
} from '@apps/control-hub/src/network';
import { computed, onMounted, ref, watch, watchEffect } from 'vue';
import { VAceEditor } from 'vue3-ace-editor';
import ace from 'ace-builds';
import 'ace-builds/src-noconflict/mode-yaml';
// import 'ace-builds/src-noconflict/snippets/yaml';
// import 'ace-builds/src-noconflict/theme-textmate';

import 'ace-builds/src-noconflict/theme-textmate';
import 'ace-builds/src-noconflict/mode-groovy';
import 'ace-builds/src-noconflict/theme-chaos';
// import 'ace-builds/src-noconflict/keybinding-vscode';
import 'ace-builds/src-noconflict/ext-searchbox';

// eslint-disable-next-line @typescript-eslint/ban-ts-comment
//@ts-ignore
import workerJsonUrl from 'file-loader?esModule=false!ace-builds/src-noconflict/worker-yaml.js';
import { objectToYaml, yamlToObject } from './yaml';
import { useRoute } from 'vue-router';
import { Loading, Notify, useQuasar } from 'quasar';
import { ObjectMapper } from '@apps/control-hub/src/utils/object.mapper';
import { t } from '@apps/control-hub/src/boot/i18n';
import { get, set } from 'lodash-es';
import { saveAs } from 'file-saver';
import Dialog from '@apps/control-panel-common/src/components/Dialog/Dialog.vue';
ace.config.setModuleUrl('ace/mode/yaml_worker', workerJsonUrl);
// src/components/Modals/EditYaml/index.jsx
interface Props {
	name: string;
	apiVersion: string;
	namespace: string;
	originData: any;
	readonly?: boolean;
}

const emits = defineEmits(['change', 'show', 'hide']);

const $q = useQuasar();

const props = withDefaults(defineProps<Props>(), {
	readonly: false
});

const data = ref();
const detail = ref();
const visible2 = ref(false);
const aceVisileb = ref(false);
const loading = ref(false);
const loading2 = ref(false);
const theme = ref('chaos');
const fileList = ref();
const mode = computed(() => {
	const { kind } = route.params as Record<string, string>;

	switch (kind) {
		case 'deployments':
			return 'deployment';
		case 'statefulsets':
			return 'statefulset';
		case 'daemonsets':
			return 'daemonset';
		default:
			return 'deployment';
	}
});

const route = useRoute();
const footerHeight = computed(() => (props.readonly ? '0px' : '56px'));

const yamlShow = () => {
	visible2.value = true;
};

const yamlHide = () => {
	visible2.value = false;
};

const show = (evt: Event) => {
	emits('show');
	aceVisileb.value = true;
};

const hide = (evt: Event) => {
	emits('hide');
	aceVisileb.value = false;
	data.value = undefined;
};

const saveAsFile = (text = '', fileName = 'default.txt') => {
	const blob = new Blob([text], { type: 'text/plain;charset=utf-8' });
	saveAs(blob, fileName);
};

const handleDownload = () => {
	let value = data.value;
	let template;
	if (value.metadata) {
		template = value;
	} else {
		const values: any = Object.values(value);
		if (values && values[0] && values[0].metadata) {
			template = values[0];
		}
	}
	const name = get(template, 'metadata.name', 'default');
	const namespace = get(template, 'metadata.namespace', '');
	const kind = get(template, 'kind', '').toLowerCase();
	const fileName = [name, namespace, kind].filter((item) => item).join('.');
	saveAsFile(value, `${fileName}.${mode.value}.yaml`);
};

const submit = async () => {
	let newData = data.value;
	set(newData, 'metadata.resourceVersion', detail.value.resourceVersion);
	update(detail, newData);
};

const update = async (
	params: Record<string, any>,
	ObjectData: Record<string, any>
) => {
	try {
		loading2.value = true;
		const { module: paramModule } = route.params as Record<string, string>;
		const { module: metaModule } = route.meta as Record<string, string>;
		const module = paramModule || metaModule;
		const { data: result } = await getCustomresourceItem(
			props.apiVersion,
			props.namespace,
			module,
			props.name
		);
		let newObject = yamlToObject(ObjectData, false)[0];
		const resourceVersion = get(result, 'metadata.resourceVersion');

		if (resourceVersion) {
			set(newObject, 'metadata.resourceVersion', resourceVersion);
			console.log('resourceVersion', newObject);
		}

		// TODO: console.log('newObject', yamlToObject(newObject, !!detail.value.metadata));
		const { data } = await updateCustomResources(
			props.namespace,
			module,
			props.name,
			props.apiVersion,
			newObject
		);
		$q.notify({
			type: 'positive',
			message: t('CHANGED_SUCCESSFULLY')
		});
		yamlHide();
		emits('change');
	} catch (error) {
		$q.notify({
			type: 'negative',
			message: JSON.stringify(error)
		});
	}
	loading2.value = false;
	loading.value = false;
};

const handleFileAdded = (file: any) => {
	if (file.type === 'application/x-yaml') {
		fileList.value = [file];
	} else {
		fileList.value = [];
	}
};

const handleFileFailed = (err: any, file: any) => {
	fileList.value = [];
};

watchEffect(() => {
	detail.value = ObjectMapper['customresourcedefinitions'](props.originData);
	data.value = objectToYaml(detail.value._originData);
});

defineExpose({
	show: yamlShow
});
</script>

<style lang="scss" scoped>
.yaml-container {
	font-family: 'Roboto';
}
.yaml-tool-container {
	position: absolute;
	top: 8px;
	right: 8px;
	z-index: 1;
}
</style>
