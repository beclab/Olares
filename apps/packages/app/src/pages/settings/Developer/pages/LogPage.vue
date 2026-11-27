<template>
	<page-title-component :show-back="true" :title="t('advanced_logs_label')" />
	<bt-scroll-area class="nav-height-scroll-area-conf">
		<bt-list>
			<bt-form-item
				:title="t('get_logs')"
				:margin-top="false"
				:chevron-right="false"
				:widthSeparator="collected"
			>
				<div
					class="upgradeNow"
					v-if="collectingStatus === ProcessingState.InProgress"
				>
					<span class="loader" />
					{{ t('collecting') }}
				</div>
				<div
					class="upgradeNow"
					v-if="collectingStatus === ProcessingState.Init || collected"
					@click="handleCollect"
				>
					{{ t('collect') }}
				</div>
			</bt-form-item>

			<bt-form-item
				:title="
					collectingStatus === ProcessingState.Completed
						? t('log_file_storage')
						: t('collect_failed_reason')
				"
				:margin-top="false"
				:chevron-right="false"
				:widthSeparator="false"
				v-if="collected"
			>
				<div
					v-if="collectingStatus === ProcessingState.Completed"
					class="row justify-end items-center"
				>
					<!-- i18n-ignore-start: file system path, not translatable -->
					<div class="text-body2 q-mr-lg" style="flex: 1">
						{{ logPath || '/Home/pod_logs' }}
					</div>
					<!-- i18n-ignore-end -->
					<div class="upgradeNow" @click="openLogFile">
						{{ t('base.open') }}
					</div>
				</div>

				<div
					v-if="
						collectingStatus === ProcessingState.Failed && collectFailedReason
					"
					class="text-body2 q-mr-lg text-negative"
					style="text-align: end; word-break: break-all"
				>
					{{ collectFailedReason }}
				</div>
			</bt-form-item>
		</bt-list>

		<bt-list
			v-if="displayedHistoryRecords.length > 0"
			:label="t('collection_history')"
		>
			<bt-form-item
				v-for="(record, index) in displayedHistoryRecords"
				:key="record.runID"
				:margin-top="false"
				:chevron-right="false"
				:widthSeparator="index !== displayedHistoryRecords.length - 1"
				:min-item-height="72"
			>
				<template #title>
					<div class="column justify-center record-info q-py-md">
						<!-- i18n-ignore-start: file system path, not translatable -->
						<div class="text-body2 text-ink-1 record-path">
							{{ record.path || stateText(record.state) }}
						</div>
						<!-- i18n-ignore-end -->
						<div class="text-body3 text-ink-3 q-mt-xs">
							{{ formatRecordTime(record.createdAt) }}
						</div>
					</div>
				</template>

				<div class="row justify-end items-center no-wrap">
					<div
						class="text-body2 q-mr-md"
						:class="{
							'text-positive': record.state === ProcessingState.Completed,
							'text-negative':
								record.state === ProcessingState.Failed ||
								record.state === ProcessingState.NotFound
						}"
					>
						{{ stateText(record.state) }}
					</div>
					<div
						v-if="record.state === ProcessingState.Completed"
						class="upgradeNow"
						@click="openHistoryLog(record)"
					>
						{{ t('base.open') }}
					</div>
				</div>
			</bt-form-item>
		</bt-list>
	</bt-scroll-area>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import { computed, onMounted, onUnmounted, ref } from 'vue';
import BtFormItem from 'src/components/settings/base/BtFormItem.vue';
import PageTitleComponent from 'src/components/settings/PageTitleComponent.vue';
import { notifyFailed } from 'src/utils/settings/btNotify';
import { useTerminusDStore } from 'src/stores/settings/terminusd';
import { ProcessingState } from 'src/constant/global';
import BtList from 'src/components/settings/base/BtList.vue';
import {
	addCollectLogRun,
	updateCollectLogRun,
	getCollectLogRuns,
	getPendingCollectLogRunId,
	CollectLogRunRecord
} from 'src/utils/settings/logRunDB';
import { format as formatDate } from 'date-fns';
import * as process from 'process';

const terminusDStore = useTerminusDStore();
const { t } = useI18n();

const collectFailedReason = ref('');
const collectingStatus = ref<ProcessingState>(ProcessingState.Init);
const firstLoad = ref(false);
const timer = ref();
const runId = ref('');
const logPath = ref('');
const historyRecords = ref<CollectLogRunRecord[]>([]);
const displayedHistoryRecords = computed(() =>
	historyRecords.value.filter(
		(record) =>
			record.state !== ProcessingState.InProgress &&
			record.state !== ProcessingState.Init
	)
);
const collected = computed(() => {
	return (
		collectingStatus.value === ProcessingState.Completed ||
		collectingStatus.value === ProcessingState.Failed ||
		collectingStatus.value === ProcessingState.NotFound
	);
});

const loadHistoryRecords = async () => {
	historyRecords.value = await getCollectLogRuns();
};

const formatRecordTime = (time: number) => {
	try {
		return formatDate(new Date(time), 'yyyy-MM-dd HH:mm:ss'); // i18n-ignore: date-fns format pattern
	} catch (error) {
		return '';
	}
};

const stateText = (state: ProcessingState) => {
	switch (state) {
		case ProcessingState.Completed:
			return t('collect_state_completed');
		case ProcessingState.Failed:
			return t('collect_state_failed');
		case ProcessingState.NotFound:
			return t('collect_state_not_found');
		case ProcessingState.InProgress:
		default:
			return t('collect_state_in_progress');
	}
};

const buildLogFileUrl = (path: string) => {
	const hostname = window.location.hostname.replace(
		process.env.DEV ? 'test' : 'settings',
		'files'
	);
	const filePath = path ? `/Files/${path}` : '/Files/Home/pod_logs/';
	return window.location.protocol + '//' + hostname + filePath;
};

const handleCollect = async () => {
	try {
		const data = await terminusDStore.collect_logs();
		if (!data?.runID) {
			notifyFailed(t('failed'));
			return;
		}
		runId.value = data.runID;
		await addCollectLogRun(runId.value);
		await loadHistoryRecords();
		collectingStatus.value = ProcessingState.InProgress;
		startQueryTimer();
	} catch (error) {
		notifyFailed(t('failed'));
	}
};

const openLogFile = () => {
	window.open(buildLogFileUrl(logPath.value));
};

const openHistoryLog = (record: CollectLogRunRecord) => {
	window.open(buildLogFileUrl(record.path));
};

const startQueryTimer = async () => {
	await queryLogStatus();
	firstLoad.value = true;
	if (collectingStatus.value === ProcessingState.InProgress) {
		timer.value = setInterval(async () => {
			if (collected.value || !runId.value) {
				window.clearInterval(timer.value);
				return;
			}
			await queryLogStatus();
		}, 3000);
	}
};

const queryLogStatus = async () => {
	if (!runId.value) {
		return;
	}
	try {
		const data = await terminusDStore.collectLogsByRunId(runId.value);
		collectingStatus.value = data.state;
		collectFailedReason.value = data.error || '';
		logPath.value = data.path || '';
		await updateCollectLogRun(runId.value, {
			state: data.state,
			path: data.path || '',
			error: data.error || ''
		});
		await loadHistoryRecords();
	} catch (error) {
		console.log(error);
	}
};

const resumePendingCollect = async () => {
	const pendingRunId = await getPendingCollectLogRunId();
	if (!pendingRunId) {
		return;
	}
	runId.value = pendingRunId;
	collectingStatus.value = ProcessingState.InProgress;
	startQueryTimer();
};

onMounted(async () => {
	await loadHistoryRecords();
	await resumePendingCollect();
});

onUnmounted(() => {
	window.clearInterval(timer.value);
});
</script>

<style scoped lang="scss">
.list_section {
	min-height: 56px;
}

.upgradeNow {
	border: 1px solid $btn-stroke;
	padding: 8px 12px;
	border-radius: 8px;
	cursor: pointer;
	// background-color: $background-3;

	&:hover {
		background: $background-5;
	}
}

.record-info {
	width: 100%;

	.record-path {
		word-break: break-all;
		white-space: normal;
	}
}

.loader {
	width: 1rem;
	height: 1rem;
	color: inherit;
	vertical-align: middle;
	border: 0.2em solid transparent;
	border-top-color: currentcolor;
	border-bottom-color: currentcolor;
	border-radius: 50%;
	position: relative;
	animation: 1s loader-30 linear infinite;
	display: inline-block;

	&:before,
	&:after {
		content: '';
		display: block;
		width: 0;
		height: 0;
		position: absolute;
		border: 0.2em solid transparent;
		border-bottom-color: currentcolor;
	}

	&:before {
		transform: rotate(135deg);
		right: -0.2em;
		top: -0.05em;
	}

	&:after {
		transform: rotate(-45deg);
		left: -0.2em;
		bottom: -0.05em;
	}
}

@keyframes loader-30 {
	0% {
		transform: rotate(0deg);
	}

	100% {
		transform: rotate(360deg);
	}
}
</style>
