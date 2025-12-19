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

		<Dialog2
			persistent
			full-width
			full-height
			:ok="!readonly"
			:cancel="!readonly"
			:title="name"
			ref="dialog2Ref"
			@onSubmit="submit"
			:loading="loading2"
		>
			<div
				style="
					height: calc(100%);
					border-radius: 6px;
					overflow: hidden;
					position: relative;
				"
			>
				<!-- <q-btn
            color="primary"
            dense
            rounded
            padding="xs md"
            class="yaml-tool-container"
          >
            <q-uploader v-model="fileList" :auto-upload="false" accept=".yaml">
              <template v-slot:header>
                <q-icon name="upload" />
              </template>
            </q-uploader>
            <q-icon dense name="upload" @click="handleDownload" />
            <q-separator spaced inset vertical dark />
            <q-icon name="download" @click="handleDownload" />
          </q-btn> -->
				<v-ace-editor
					v-if="aceVisileb"
					v-model:value="data"
					lang="yaml"
					theme="chaos"
					:readonly="loading2 || readonly"
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
			</div>
			<q-inner-loading :showing="loading" style="z-index: 999999">
			</q-inner-loading>
		</Dialog2>
	</div>
</template>

<script setup lang="ts">
import {
	getWorkloadsControler,
	getYaml,
	updateWorkloadsControler
} from '@apps/control-hub/src/network';
import { computed, onMounted, ref, watch } from 'vue';
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
import Dialog2 from '@apps/control-panel-common/src/components/Dialog/Dialog.vue';
ace.config.setModuleUrl('ace/mode/yaml_worker', workerJsonUrl);
// src/components/Modals/EditYaml/index.jsx
interface Props {
	name?: string;
	nameKey?: string;
	readonly?: boolean;
}

const emits = defineEmits(['change']);

const $q = useQuasar();

const props = withDefaults(defineProps<Props>(), {
	nameKey: 'name',
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
const dialog2Ref = ref();
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

const yamlShow = () => {
	dialog2Ref.value.show();
	show();
};

const yamlHide = () => {
	dialog2Ref.value.close();
};

const show = () => {
	fetchData();
	aceVisileb.value = true;
};

const hide = () => {
	aceVisileb.value = false;
	data.value = undefined;
};

const fetchData = () => {
	const {
		namespace,
		kind,
		[props.nameKey]: name
	} = route.params as Record<string, string>;
	loading.value = true;
	getWorkloadsControler(namespace, kind, name)
		.then((res) => {
			// eslint-disable-next-line @typescript-eslint/ban-ts-comment
			// @ts-ignore
			detail.value = ObjectMapper[kind](res.data);
			data.value = objectToYaml(detail.value._originData);
		})
		.finally(() => {
			loading.value = false;
		});
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
	newObject: Record<string, any>
) => {
	try {
		loading2.value = true;
		const {
			namespace,
			kind,
			[props.nameKey]: name
		} = route.params as Record<string, string>;
		const { data: result } = await getWorkloadsControler(namespace, kind, name);

		const resourceVersion = get(result, 'metadata.resourceVersion');
		if (resourceVersion) {
			// TODO: fix newObject is string
			set(newObject, 'metadata.resourceVersion', resourceVersion);
		}
		// TODO: console.log('newObject', yamlToObject(newObject, !!detail.value.metadata));
		const obj = yamlToObject(newObject, false);
		const { data } = await updateWorkloadsControler(
			namespace,
			kind,
			name,
			obj[0]
		);
		yamlHide();
		emits('change');
	} catch {
		//
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
