<template>
	<div class="column" style="padding: 0 32px 32px 32px">
		<display-item
			:title="t('base.name')"
			:content="nodeStatus.name"
			:copy="true"
		/>
		<display-item :title="t('base.id')" :content="nodeStatus.id" :copy="true" />
		<display-item :title="t('base.type')" :content="nodeStatus.type" />
		<display-item :title="t('base.phase')" :content="nodeStatus.phase" />
		<display-item
			:title="t('base.start_time')"
			:content="nodeStatus.startedAt ? formattedDate(nodeStatus.startedAt) : ''"
		/>
		<display-item
			v-if="nodeStatus.phase === NODE_PHASE.SUCCEEDED"
			:title="t('base.end_time')"
			:content="
				nodeStatus.finishedAt ? formattedDate(nodeStatus.finishedAt) : ''
			"
		/>
		<display-item
			v-if="nodeStatus.phase === NODE_PHASE.SUCCEEDED"
			:title="t('base.duration')"
			:content="getDuration"
		/>
		<display-item :title="t('base.progress')" :content="nodeStatus.progress" />
		<display-item
			:title="t('base.resources_duration')"
			:content="
				nodeStatus.resourcesDuration?.cpu +
				's(1 cpu),' +
				nodeStatus.resourcesDuration?.memory +
				's(100Mi Memory)'
			"
		/>

		<q-btn
			v-if="nodeStatus.type === 'Pod'"
			class="btn-size-sm q-mt-lg"
			style="width: 74px"
			:label="t('recommendation.logs')"
			color="orange-default"
			outline
			icon="sym_r_assignment"
			no-caps
			@click="getLog()"
		/>
	</div>
</template>

<script lang="ts" setup>
import { date, useQuasar } from 'quasar';
import { PropType, computed } from 'vue';
import { WorkflowDetail, NodeStatus } from 'src/stores/argo';
import { calculateTimeDifference } from 'src/utils/rss-utils';
import WorkflowLogs from './WorkflowLogs.vue';
// import WorkflowManifest from './WorkflowManifest.vue';
// import WorkflowEvents from './WorkflowEvents.vue';
import DisplayItem from './DisplayItem.vue';
import { NODE_PHASE } from 'src/utils/rss-types';
import { useI18n } from 'vue-i18n';

// const argoStore = useArgoStore();
const $q = useQuasar();
const { t } = useI18n();
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

const formattedDate = (datetime: string) => {
	const originalDate = new Date(datetime);
	return (
		date.formatDate(originalDate, 'M/D/YYYY, h:mm A') +
		` (${calculateTimeDifference(datetime, new Date().toISOString())})`
	);
};

const getDuration = computed(() => {
	if (props.nodeStatus.startedAt && props.nodeStatus.finishedAt) {
		return calculateTimeDifference(
			props.nodeStatus.startedAt,
			props.nodeStatus.finishedAt,
			''
		);
	} else {
		return '';
	}
});

async function getLog() {
	$q.dialog({
		component: WorkflowLogs,
		// props forwarded to your custom component
		componentProps: {
			workflow: props.workflow,
			nodeStatus: props.nodeStatus
		}
	}).onOk(async () => {
		//
	});
}

// async function getManifest() {
//   $q.dialog({
//     component: WorkflowManifest,
//     // props forwarded to your custom component
//     componentProps: {
//       workflow: props.workflow,
//       nodeStatus: props.nodeStatus,
//     },
//   }).onOk(async () => {
//     //
//   });
// }
//
// async function getEvents() {
//   $q.dialog({
//     component: WorkflowEvents,
//     // props forwarded to your custom component
//     componentProps: {
//       workflow: props.workflow,
//       nodeStatus: props.nodeStatus,
//     },
//   }).onOk(async () => {
//     //
//   });
// }
</script>
