<template>
	<MetricsTable
		:websiteId="websiteId"
		type="url"
		:label="$t('PAGES')"
		:valueLabel="t('VIEWS')"
		:loading="loading"
	></MetricsTable>
</template>

<script setup lang="ts">
import MetricsTable from './MetricsTable.vue';
import { t } from '@apps/dashboard/src/boot/i18n';

interface Props {
	websiteId: string;
	loading?: boolean;
}

withDefaults(defineProps<Props>(), {});
</script>

<style></style>
