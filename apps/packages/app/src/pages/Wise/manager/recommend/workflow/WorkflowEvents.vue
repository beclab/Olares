<template>
	<q-dialog ref="dialogRef">
		<q-card class="avatar-choose-dialog">
			<div></div>
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
import { useArgoStore, WorkflowDetail, NodeStatus } from 'src/stores/argo';
import axios from 'axios';
import { useI18n } from 'vue-i18n';

const { dialogRef, onDialogCancel, onDialogOK } = useDialogPluginComponent();
const i18n = useI18n();
const argoStore = useArgoStore();

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

const events = ref();

async function query2() {
	const ids = props.nodeStatus.id.split('-');
	const url =
		'/api/v1/stream/events/' +
		argoStore.namespace +
		'?listOptions.fieldSelector=involvedObject.kind=Pod,involvedObject.name=' +
		props.workflow.metadata.name +
		'-' +
		props.nodeStatus.templateName +
		'-' +
		ids[ids.length - 1];

	console.log('events ' + url);
	try {
		axios
			.get(url, {
				headers: { responseType: 'stream', accept: 'text/event-stream' }
			})
			.then((data) => {
				console.log('events');
				console.log(events.value);
				events.value = data;
			});
	} catch (e) {
		console.log(e);
	}
}

async function query1() {
	const url =
		//'/api/v1/stream/events/' +
		'/api/v1/workflow-events/' +
		argoStore.namespace +
		'?listOptions.fieldSelector=metadata.namespace=' +
		argoStore.namespace +
		',metadata.name=' +
		props.workflow.metadata.name;

	console.log('events ' + url);
	try {
		axios
			.get(url, {
				headers: { responseType: 'stream', accept: 'text/event-stream' }
			})
			.then((data) => {
				console.log('events');
				console.log(events.value);
				events.value = data;
			});
	} catch (e) {
		console.log(e);
	}
}

onMounted(async () => {
	await query1();
	await query2();
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
