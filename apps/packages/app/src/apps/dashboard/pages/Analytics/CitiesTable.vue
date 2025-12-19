<template>
	<MetricsTable
		:websiteId="websiteId"
		type="city"
		:valueLabel="$t('CITY')"
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
