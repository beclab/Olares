<template>
	<MetricsTable
		:websiteId="websiteId"
		type="os"
		:valueLabel="$t('OS')"
		:loading="loading"
	></MetricsTable>
</template>

<script setup lang="ts">
import MetricsTable from './MetricsTable.vue';

interface Props {
	websiteId: string;
	loading?: boolean;
}

const props = withDefaults(defineProps<Props>(), {});
</script>

<style></style>
