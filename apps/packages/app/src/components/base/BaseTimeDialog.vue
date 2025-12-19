<template>
	<bt-custom-dialog
		ref="CustomRef"
		:title="t('run_backup_at')"
		:skip="false"
		:ok="t('confirm')"
		size="small"
		:platform="deviceStore.platform"
		:cancel="t('cancel')"
		@onSubmit="onConfirm"
		:okDisabled="okDisable"
	>
		<div class="text-body1 text-ink-3 q-mb-xs">
			{{ t('hours') }}
		</div>

		<bt-select-v3 v-model="hourResult" :options="hourOptions" />

		<div class="text-body1 text-ink-3 q-mb-xs q-mt-lg">
			{{ t('minutes') }}
		</div>

		<bt-select-v3 v-model="minuteResult" :options="minuteOptions" />
	</bt-custom-dialog>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import { computed, onMounted, ref } from 'vue';
import { useDeviceStore } from 'src/stores/device';
import BtSelectV3 from '../settings/base/BtSelectV3.vue';
import { SelectorProps } from '../../constant';

const { t } = useI18n();
const CustomRef = ref();
const deviceStore = useDeviceStore();

const props = defineProps({
	hours: {
		type: Number,
		default: 0
	},
	minutes: {
		type: Number,
		default: 0
	},
	time: {
		type: String
	}
});

const padZero = (value: number | string): string => {
	return value.toString().padStart(2, '0');
};

const hourResult = ref();
const minuteResult = ref();
const hourOptions = ref<SelectorProps[]>([]);
const minuteOptions = ref<SelectorProps[]>([]);

onMounted(() => {
	for (let i = 0; i < 24; i++) {
		hourOptions.value.push({
			value: padZero(i),
			label: padZero(i)
		});
	}
	for (let i = 0; i < 60; i++) {
		minuteOptions.value.push({
			value: padZero(i),
			label: padZero(i)
		});
	}

	if (props.time && props.time.includes(':')) {
		const array = props.time.split(':');
		hourResult.value = array[0];
		minuteResult.value = array[1];
	} else {
		hourResult.value = padZero(props.hours);
		minuteResult.value = padZero(props.minutes);
	}
});

const okDisable = computed(() => {
	if (props.time && props.time.includes(':')) {
		return `${hourResult.value}:${minuteResult.value}` == props.time;
	} else {
		return (
			hourResult.value == padZero(props.hours) &&
			minuteResult.value == padZero(props.minutes)
		);
	}
});

const onConfirm = () => {
	CustomRef.value.onDialogOK(`${hourResult.value}:${minuteResult.value}`);
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
