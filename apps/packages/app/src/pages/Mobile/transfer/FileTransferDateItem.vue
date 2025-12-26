<template>
	<q-item class="date-item q-pa-none text-ink-2" dense clickable>
		{{ isToday(date) ? t('transmission.today') : date }}
	</q-item>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import { date as quasarDate } from 'quasar';

defineProps({
	date: {
		type: String,
		required: true
	}
});

const { t } = useI18n();

const isToday = (startTime: string): boolean => {
	const timeStamp = Date.now();
	const formattedString = quasarDate.formatDate(timeStamp, 'YYYY-MM-DD');

	if (formattedString === startTime) {
		return true;
	} else {
		return false;
	}
};
</script>

<style scoped lang="scss">
.date-item {
	width: 100%;
	line-height: 32px;
	font-size: 16px;
	font-weight: 700;
	margin-top: 8px;
}
</style>
