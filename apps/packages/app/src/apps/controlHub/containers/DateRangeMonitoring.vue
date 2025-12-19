<template>
	<QSectionStyle>
		<q-select
			v-model="selectValue"
			:options="options"
			dense
			outlined
			@update:model-value="change"
		/>
	</QSectionStyle>
</template>

<script lang="ts">
import { t } from 'src/boot/control-hub-i18n';

export interface DateRangeItem {
	label: string;
	value: number;
}

export const options: DateRangeItem[] = [
	{
		label: t('LAST_TIME_H', { count: 1 }),
		value: 3600
	},
	{
		label: t('LAST_TIME_H', { count: 2 }),
		value: 3600 * 2
	},
	{
		label: t('LAST_TIME_H', { count: 3 }),
		value: 3600 * 3
	},
	{
		label: t('LAST_TIME_H', { count: 5 }),
		value: 3600 * 5
	},
	{
		label: t('LAST_TIME_H', { count: 8 }),
		value: 3600 * 8
	},
	{
		label: t('LAST_TIME_H', { count: 12 }),
		value: 3600 * 12
	},
	{
		label: t('LAST_TIME_D', { count: 1 }),
		value: 3600 * 24
	}
];
</script>

<script setup lang="ts">
import { ref, toRefs } from 'vue';
import QSectionStyle from '@apps/control-hub/src/components/QSectionStyle.vue';
import { useI18n } from 'vue-i18n';

const selectValue = ref<DateRangeItem>(options[0]);

interface Props {
	defaultValue?: DateRangeItem;
}

const emit = defineEmits<{
	(e: 'change', data: DateRangeItem): void;
}>();

const props = withDefaults(defineProps<Props>(), {});
const { defaultValue } = toRefs(props);
if (defaultValue?.value) {
	selectValue.value = defaultValue.value;
}

const change = (value: DateRangeItem) => {
	emit('change', value);
};
</script>

<style></style>
