<template>
	<DatePickerStyle>
		<el-date-picker
			v-model="dateValue"
			type="datetimerange"
			unlink-panels
			:range-separator="$t('RANGE_SEPARATOR')"
			:start-placeholder="$t('START_DATE')"
			:end-placeholder="$t('END_DATE')"
			:shortcuts="shortcuts"
			:disabled-date="disabledDate"
			:disabled="disabled"
			@visible-change="handleVisibleChange"
		/>
	</DatePickerStyle>
</template>

<script lang="ts" setup>
import { ElDatePicker } from 'element-plus';

import { computed, watch, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import DatePickerStyle from 'src/components/style/DatePickerStyle.vue';

const { t } = useI18n();

const props = defineProps<{
	modelValue: string[];
	disabled?: boolean;
}>();

const emit = defineEmits<{
	'update:modelValue': [value: string[]];
}>();

const internalDateValue = ref<string[]>(props.modelValue);

watch(
	() => props.modelValue,
	(newValue) => {
		internalDateValue.value = newValue;
	}
);

const dateValue = computed({
	get: () => internalDateValue.value,
	set: (value) => {
		internalDateValue.value = value;
	}
});

const handleVisibleChange = (visible: boolean) => {
	if (!visible && internalDateValue.value) {
		const hasChanged =
			JSON.stringify(internalDateValue.value) !==
			JSON.stringify(props.modelValue);
		if (hasChanged) {
			emit('update:modelValue', internalDateValue.value);
		}
	}
};

const shortcuts = computed(() => [
	{
		text: t('LAST_TIME_M', { count: 10 }),
		value: () => {
			const end = new Date();
			const start = new Date();
			start.setTime(start.getTime() - 60 * 1000 * 10);
			return [start, end];
		}
	},
	{
		text: t('LAST_TIME_M', { count: 20 }),
		value: () => {
			const end = new Date();
			const start = new Date();
			start.setTime(start.getTime() - 60 * 1000 * 20);
			return [start, end];
		}
	},
	{
		text: t('LAST_TIME_M', { count: 30 }),
		value: () => {
			const end = new Date();
			const start = new Date();
			start.setTime(start.getTime() - 60 * 1000 * 30);
			return [start, end];
		}
	},
	{
		text: t('LAST_TIME_H', { count: 1 }),
		value: () => {
			const end = new Date();
			const start = new Date();
			start.setTime(start.getTime() - 3600 * 1000 * 1);
			return [start, end];
		}
	},
	{
		text: t('LAST_TIME_H', { count: 2 }),
		value: () => {
			const end = new Date();
			const start = new Date();
			start.setTime(start.getTime() - 3600 * 1000 * 2);
			return [start, end];
		}
	},
	{
		text: t('LAST_TIME_H', { count: 3 }),
		value: () => {
			const end = new Date();
			const start = new Date();
			start.setTime(start.getTime() - 3600 * 1000 * 3);
			return [start, end];
		}
	},
	{
		text: t('LAST_TIME_H', { count: 5 }),
		value: () => {
			const end = new Date();
			const start = new Date();
			start.setTime(start.getTime() - 3600 * 1000 * 5);
			return [start, end];
		}
	},
	{
		text: t('LAST_TIME_H', { count: 8 }),
		value: () => {
			const end = new Date();
			const start = new Date();
			start.setTime(start.getTime() - 3600 * 1000 * 8);
			return [start, end];
		}
	},
	{
		text: t('LAST_TIME_H', { count: 12 }),
		value: () => {
			const end = new Date();
			const start = new Date();
			start.setTime(start.getTime() - 3600 * 1000 * 12);
			return [start, end];
		}
	},
	{
		text: t('LAST_TIME_D', { count: 1 }),
		value: () => {
			const end = new Date();
			const start = new Date();
			start.setTime(start.getTime() - 3600 * 1000 * 24);
			return [start, end];
		}
	},
	{
		text: t('LAST_TIME_D', { count: 3 }),
		value: () => {
			const end = new Date();
			const start = new Date();
			start.setTime(start.getTime() - 3600 * 1000 * 24 * 3);
			return [start, end];
		}
	},
	{
		text: t('LAST_TIME_D', { count: 7 }),
		value: () => {
			const end = new Date();
			const start = new Date();
			start.setTime(start.getTime() - 3600 * 1000 * 24 * 7);
			return [start, end];
		}
	}
]);

const disabledDate = (time: Date) => {
	return time.getTime() >= Date.now();
};
</script>
