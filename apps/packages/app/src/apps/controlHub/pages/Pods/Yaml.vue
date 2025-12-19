<template>
	<q-btn
		no-caps
		@click="yamlShow"
		outline
		color="primary"
		v-if="$slots.default"
	>
		<slot></slot>
	</q-btn>
	<Dialog
		:title="name"
		persistent
		full-width
		full-height
		v-model="visible2"
		@show="show"
		@hide="hide"
	>
		<div class="full-height">
			<div style="height: calc(100%)">
				<div
					style="
						height: calc(100%);
						border-radius: 6px;
						overflow: hidden;
						position: relative;
					"
				>
					<Yaml v-if="aceVisileb" v-model:data="data" readonly></Yaml>
				</div>
			</div>
			<q-inner-loading :showing="loading"> </q-inner-loading>
		</div>
	</Dialog>
</template>

<script setup lang="ts">
import { getYaml } from '@apps/control-hub/src/network';
import { nextTick, onMounted, ref } from 'vue';
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
import { useRoute } from 'vue-router';
import { ObjectMapper } from '@apps/control-hub/src/utils/object.mapper';
import Yaml from '@apps/control-panel-common/src/containers/Yaml.vue';
ace.config.setModuleUrl('ace/mode/yaml_worker', workerJsonUrl);
// @apps/control-panel-common/src/components/Modals/EditYaml/index.jsx
import Dialog from '@apps/control-panel-common/src/components/Dialog/Dialog.vue';

interface Props {
	name?: string;
	namespace?: string;
}
const props = withDefaults(defineProps<Props>(), {});
const route = useRoute();
const loading = ref(false);
const data = ref();
const visible2 = ref(false);
const aceVisileb = ref(false);

const yamlShow = () => {
	visible2.value = true;
	nextTick(() => {
		fetchData();
	});
};

const show = (evt: Event) => {
	aceVisileb.value = true;
};

const hide = (evt: Event) => {
	aceVisileb.value = false;
	data.value = undefined;
};

const fetchData = () => {
	const { namespace, name }: Record<string, any> = route.params;
	loading.value = true;
	const namespaceValue = namespace || props.namespace;
	const nameValue = name || props.name;
	console.log('testaaaaa', namespace, name, namespaceValue, nameValue);

	getYaml(namespaceValue, nameValue)
		.then((res) => {
			const result = res.data;
			const detail = ObjectMapper.pods(result);
			// detail.cluster = cluster
			// if (!isEmpty(detail.volumes)) {
			//   detail.containers.forEach(container => {
			//     if (!isEmpty(container.volumeMounts)) {
			//       container.volumeMounts.forEach(volumeMount => {
			//         const volume = detail.volumes.find(
			//           _volume => _volume.name === volumeMount.name
			//         )
			//         if (!isEmpty(volume)) {
			//           volume.containers = volume.containers || []
			//           volume.containers.push(container)
			//         }
			//       })
			//     }
			//   })
			// }
			data.value = objectToYaml(detail._originData);
		})
		.finally(() => {
			loading.value = false;
		});
};

defineExpose({ show: yamlShow });
</script>

<style></style>
