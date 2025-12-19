<template>
	<div class="status">
		<Status
			:type="statusType"
			:name="statusName"
			:total="totalPods"
			:ready="readyPods"
			:flicker="true"
		/>
		<q-icon name="watch_later" v-show="hasHPA" color="grey-5">
			<q-tooltip>
				{{ t('HPA_SET_TIP') }}
			</q-tooltip>
		</q-icon>
	</div>
</template>

<script setup lang="ts">
import { get, isEmpty } from 'lodash';
import { getWorkloadStatus } from '@apps/control-panel-common/src/utils/status';
import { defineProps } from 'vue';
import Status from '../components/Status.vue';
import { useI18n } from 'vue-i18n';
const { t } = useI18n();
interface WorkloadStatusProps {
	data: any;
	module: string;
}

const props = defineProps<WorkloadStatusProps>();

const { status: statusResult } =
	getWorkloadStatus(props.data, props.module) || '';
const statusType = statusResult.startsWith('S2I')
	? statusResult.slice(4)
	: statusResult;
const statusName = t(statusType.toUpperCase());

const totalPods =
	props.module === 'daemonsets'
		? get(props.data, 'status.desiredNumberScheduled', 0)
		: props.data.podNums;
const readyPods =
	props.module === 'daemonsets'
		? get(props.data, 'status.numberAvailable', 0)
		: props.data.readyPodNums;

const hasHPA = !isEmpty(
	get(props.data, 'annotations["kubesphere.io/relatedHPA"]')
);
</script>
