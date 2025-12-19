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
			:title="title"
			persistent
			full-width
			full-height
			:ok="!readonly"
			:cancel="!readonly"
			v-model="visible2"
			@onSubmit="submit"
			@show="show"
			@hide="hide"
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
			</div>
			<q-inner-loading :showing="loading" style="z-index: 999999">
			</q-inner-loading>
		</Dialog>
	</div>
</template>

<script setup lang="ts">
import { getDetail, updateDetail } from '@apps/control-hub/src/network';
import { computed, ref } from 'vue';
import { VAceEditor } from 'vue3-ace-editor';
import ace from 'ace-builds';
import 'ace-builds/src-noconflict/mode-yaml';
import 'ace-builds/src-noconflict/theme-textmate';
import 'ace-builds/src-noconflict/mode-groovy';
import 'ace-builds/src-noconflict/theme-chaos';
import 'ace-builds/src-noconflict/ext-searchbox';
//@ts-ignore
import workerJsonUrl from 'file-loader?esModule=false!ace-builds/src-noconflict/worker-yaml.js';
import { objectToYaml, yamlToObject } from './yaml';
import { useRoute } from 'vue-router';
import { useQuasar } from 'quasar';
import { ObjectMapper } from '@apps/control-hub/src/utils/object.mapper';
import { get, set } from 'lodash-es';
import { saveAs } from 'file-saver';
import { API_VERSIONS } from '@apps/control-hub/src/utils/constants';
import { cloneDeep, setWith } from 'lodash';
import Dialog from '@apps/control-panel-common/src/components/Dialog/Dialog.vue';

ace.config.setModuleUrl('ace/mode/yaml_worker', workerJsonUrl);
interface Props {
	title?: string;
	name?: string;
	module?: string;
	namespace?: string;
	readonly?: boolean;
}

const emits = defineEmits(['change']);

const props = withDefaults(defineProps<Props>(), {
	readonly: false
});

const data = ref();
const detail = ref();
const visible2 = ref(false);
const aceVisileb = ref(false);
const loading = ref(false);
const loading2 = ref(false);
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
		case 'persistentvolumeclaims':
			return 'persistentvolumeclaims';
		default:
			return 'deployment';
	}
});

const route = useRoute();

const yamlShow = () => {
	visible2.value = true;
};

const yamlHide = () => {
	visible2.value = false;
};

const show = (evt: Event) => {
	fetchData();
	aceVisileb.value = true;
};

const hide = (evt: Event) => {
	aceVisileb.value = false;
	data.value = undefined;
};

const apiVersion = API_VERSIONS[props.module] || '';

const fetchData = () => {
	const { namespace, kind, name } = route.params as Record<string, string>;
	const type = props.module || kind;
	loading.value = true;
	getDetail(apiVersion, {
		namespace: props.namespace || namespace,
		kind: type,
		name: props.name || name
	})
		.then((res) => {
			// eslint-disable-next-line @typescript-eslint/ban-ts-comment
			// @ts-ignore
			detail.value = ObjectMapper[type](res.data);
			data.value = objectToYaml(detail.value._originData);
		})
		.finally(() => {
			loading.value = false;
		});
};

const submit = async () => {
	let newData = data.value;
	set(newData, 'metadata.resourceVersion', detail.value.resourceVersion);
	update(detail, newData);
};

const update = async (
	params: Record<string, any>,
	data: Record<string, any>
) => {
	let newObject = yamlToObject(data, false)[0];
	try {
		loading2.value = true;
		const { namespace, kind, name } = route.params as Record<string, string>;
		const type = props.module || kind;

		const params = {
			namespace: props.namespace || namespace,
			kind: type,
			name: props.name || name
		};
		loading.value = true;
		const { data: result } = await getDetail(apiVersion, params);

		const resourceVersion = get(result, 'metadata.resourceVersion');
		if (resourceVersion) {
			set(newObject, 'metadata.resourceVersion', resourceVersion);
		}
		newObject = objectToYaml(newObject);
		const obj = yamlToObject(newObject, false);
		const { data } = await updateDetail(apiVersion, params, obj[0]);
		yamlHide();
		emits('change');
	} catch {
		//
	}
	loading2.value = false;
	loading.value = false;
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
