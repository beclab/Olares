<template>
	<bt-custom-dialog
		ref="CustomRef"
		:title="t('edit_backup')"
		:skip="false"
		:ok="t('confirm')"
		size="medium"
		:platform="deviceStore.platform"
		:cancel="t('cancel')"
		:ok-loading="isLoading"
		@onSubmit="onConfirm"
		:okDisabled="okDisable"
	>
		<div class="text-body1 text-ink-3 q-mb-xs">
			{{ t('snapshot_frequency') }}
		</div>

		<bt-select-v3 v-model="frequency" :options="frequencyOptions" />

		<div
			class="text-body1 text-ink-3 q-mb-xs q-mt-lg"
			v-if="frequency !== BackupFrequency.Daily"
		>
			{{ t('run_backup_at') }}
		</div>

		<bt-select-v3
			v-model="monthDay"
			:options="monthOption"
			v-if="frequency == BackupFrequency.Monthly"
		/>

		<bt-select-v3
			v-model="weekDay"
			:options="weekOption"
			v-if="frequency == BackupFrequency.Weekly"
		/>

		<div class="text-body1 text-ink-3 q-mb-xs q-mt-lg">
			{{ t('times_of_day') }}
		</div>

		<div class="date-of-day row justify-between items-center">
			<div class="text-body1 text-ink-2 q-ml-md">
				{{ time }}
			</div>

			<q-icon
				size="20px"
				name="sym_r_access_time"
				color="ink-1"
				class="time-clock"
				@click="onTimeDialog"
			/>
		</div>
	</bt-custom-dialog>
</template>

<script setup lang="ts">
import {
	frequencyOptions,
	weekOption,
	monthOption,
	BackupPolicy
} from '../../../../constant';
import { useI18n } from 'vue-i18n';
import { BackupFrequency } from '@bytetrade/core';
import { useDeviceStore } from 'src/stores/device';
import { timestampToTime } from './FormatBackupTime';
import { computed, onMounted, PropType, ref } from 'vue';
import { useBackupStore } from '../../../../stores/settings/backup';
import BtSelectV3 from '../../../../components/settings/base/BtSelectV3.vue';
import BaseTimeDialog from '../../../../components/base/BaseTimeDialog.vue';
import { useQuasar } from 'quasar';

const { t } = useI18n();
const CustomRef = ref();
const deviceStore = useDeviceStore();
const backupStore = useBackupStore();
const isLoading = ref(false);

const props = defineProps({
	backupId: {
		type: String,
		required: true
	},
	policy: {
		type: Object as PropType<BackupPolicy>,
		required: true
	}
});

const frequency = ref<BackupFrequency | null>(null);
const weekDay = ref('');
const monthDay = ref('');
const time = ref('');
const $q = useQuasar();

onMounted(() => {
	console.log(props.policy);
	frequency.value = props.policy?.snapshotFrequency;
	weekDay.value = props.policy?.dayOfWeek !== 0 ? props.policy?.dayOfWeek : 1;
	monthDay.value =
		props.policy?.dateOfMonth !== 0 ? props.policy?.dateOfMonth : 1;
	const realTime = timestampToTime(Number(props.policy.timespanOfDay));
	console.log(realTime);
	time.value = realTime;
});

const onTimeDialog = () => {
	$q.dialog({
		component: BaseTimeDialog,
		componentProps: {
			time: time.value
		}
	}).onOk((data) => {
		time.value = data;
	});
};

const okDisable = computed(() => {
	if (frequency.value === BackupFrequency.Daily) {
		return (
			frequency.value === props.policy?.snapshotFrequency &&
			time.value === timestampToTime(Number(props.policy.timespanOfDay))
		);
	} else if (frequency.value === BackupFrequency.Weekly) {
		return (
			frequency.value === props.policy?.snapshotFrequency &&
			weekDay.value === props.policy?.dayOfWeek &&
			time.value === timestampToTime(Number(props.policy.timespanOfDay))
		);
	} else if (frequency.value === BackupFrequency.Monthly) {
		return (
			frequency.value === props.policy?.snapshotFrequency &&
			monthDay.value === props.policy?.dateOfMonth &&
			time.value === timestampToTime(Number(props.policy.timespanOfDay))
		);
	}
	return true;
});

const onConfirm = () => {
	isLoading.value = true;
	backupStore
		.updateBackupPlan(props.backupId, {
			snapshotFrequency: frequency.value,
			timesOfDay: time.value,
			dayOfWeek: weekDay.value,
			dateOfMonth: monthDay.value
		})
		.then(async () => {
			CustomRef.value.onDialogOK();
		})
		.catch((e) => {
			console.error(e);
		})
		.finally(() => {
			isLoading.value = false;
		});
};
</script>

<style scoped lang="scss">
.date-of-day {
	height: 36px;
	color: $ink-2;
	border-radius: 8px;
	border: 1px solid $input-stroke;
	width: var(--selectedWidth);

	.time-clock {
		border-radius: 4px;
		margin-right: 8px;
	}
}
</style>
