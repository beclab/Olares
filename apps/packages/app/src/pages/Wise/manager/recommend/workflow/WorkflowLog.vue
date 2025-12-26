<template>
	<q-dialog ref="dialogRef">
		<q-card class="avatar-choose-dialog">
			<div>
				{{ log }}
			</div>
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
import { LogEntry } from '/@/types';

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

const log = ref();

onMounted(async () => {
	const ids = props.nodeStatus.id.split('-');
	const url =
		'/api/v1/workflows/' +
		argoStore.namespace +
		'/' +
		props.workflow.metadata.name +
		'/log' +
		//'?listOptions.fieldSelector=involvedObject.kind=Pod,involvedObject.name=' +
		'?logOptions.container=main&grep=&logOptions.follow=true&podName=' +
		props.workflow.metadata.name +
		'-' +
		props.nodeStatus.templateName +
		'-' +
		ids[ids.length - 1];

	log.value = await axios
		.get(url, {
			headers: { responseType: 'stream', accept: 'text/event-stream' }
		})
		.then((data) => {
			const array = data
				.split('\n')
				.filter((block) => block.trim() !== '')
				.map((block) => {
					try {
						console.log(block);
						const index = block.indexOf('{');
						const json = block.substring(index);
						const logEntry = JSON.parse(json).result as LogEntry;
						try {
							logEntry.content = JSON.parse(logEntry.content);
						} catch (e) {
							console.log(block);
						}
						return JSON.stringify(logEntry, null, 2);
					} catch (e) {
						console.log(block);
					}
				});
			if (array && array.length > 0) {
				// logsData.value = array.join('\n');
			}
		});
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
