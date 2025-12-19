<template>
	<MetricsTable
		:websiteId="websiteId"
		type="browser"
		:valueLabel="$t('BROWSER')"
		:loading="loading"
	></MetricsTable>
</template>

<script setup lang="ts">
import MetricsTable from './MetricsTable.vue';

interface Props {
	websiteId: string;
	loading?: boolean;
}

withDefaults(defineProps<Props>(), {});
</script>

<style></style>
