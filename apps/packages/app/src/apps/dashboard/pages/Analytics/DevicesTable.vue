<template>
	<MetricsTable
		:websiteId="websiteId"
		type="device"
		:valueLabel="$t('DEVICE')"
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
