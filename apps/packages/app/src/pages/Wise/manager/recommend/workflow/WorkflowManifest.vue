<template>
	<q-dialog ref="dialogRef">
		<q-card class="avatar-choose-dialog">
			<pre>
      {{ displayYamlFromJson(entry) }}
      </pre>

			<pre>
      {{ displayYamlFromJson(template) }}
      </pre>
			<div class="row justify-end items-center" style="margin-top: 10px">
				<q-btn
					class="q-mr-sm btn-size-xs"
					:label="i18n.t('base.cancel')"
					color="orange-6"
					outline
					no-caps
					@click="onDialogCancel()"
				/>
				<q-btn
					class="q-mr-md btn-size-xs"
					:label="i18n.t('base.ok')"
					color="orange-6"
					no-caps
					@click="onOKClick()"
				/>
			</div>
		</q-card>
	</q-dialog>
</template>

<script lang="ts" setup>
import { useDialogPluginComponent } from 'quasar';
import { ref, onMounted, PropType } from 'vue';
import { WorkflowDetail, NodeStatus } from 'src/stores/argo';
import * as yaml from 'js-yaml';
import { useI18n } from 'vue-i18n';

const i18n = useI18n();

function displayYamlFromJson(jsonData: any): string {
	try {
		const yamlData = yaml.dump(jsonData, { indent: 4 });
		console.log(yamlData);
		return yamlData;
	} catch (error) {
		return 'Error converting JSON to YAML';
	}
}

const { dialogRef, onDialogCancel, onDialogOK } = useDialogPluginComponent();

const props = defineProps({
	workflow: {
		type: Object as PropType<WorkflowDetail>,
		required: true
	},
	nodeStatus: {
		type: Object as PropType<NodeStatus>,
		required: true
	}
});

async function onOKClick() {
	onDialogOK();
}

const entry = ref();
const template = ref();

onMounted(async () => {
	template.value = props.workflow.spec.templates.find(
		(te) => te.name == props.nodeStatus.templateName
	);
	console.log(template.value);

	entry.value = props.workflow.spec.templates.find(
		(te) => te.name == props.workflow.spec.entrypoint
	);
	// const ids = props.nodeStatus.id.split('-');
	// const url =
	//   //'/api/v1/stream/events/' +
	//   '/api/v1/workflows/' +
	//   argoStore.namespace +
	//   '/' +
	//   props.workflow.metadata.name +
	//   '/log' +
	//   //'?listOptions.fieldSelector=involvedObject.kind=Pod,involvedObject.name=' +
	//   '?logOptions.container=main&grep=&logOptions.follow=true&podName=' +
	//   props.workflow.metadata.name +
	//   '-' +
	//   props.nodeStatus.templateName +
	//   '-' +
	//   ids[ids.length - 1];
	// log.value = await axios.get(url, {
	//   headers: { responseType: 'stream', accept: 'text/event-stream' },
	// });
});
</script>
<style lang="scss">
.avatar-choose-dialog {
	.loading-view {
		width: 100%;
		height: 100%;
		position: absolute;
	}
}

.q-dialog__inner--minimized > div {
	max-width: 800px;
}
</style>
