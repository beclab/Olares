<template>
	<div class="row">
		<MetricCard
			:value="data.pageviews?.length"
			:label="$t('VIEWS')"
			hideComparison
		/>
		<MetricCard
			:value="data.visitors?.length"
			:label="$t('VISITORS')"
			hideComparison
		/>
		<MetricCard
			:value="data.events?.length"
			:label="$t('EVENTS')"
			hideComparison
		/>
		<MetricCard
			:value="data.countries?.length"
			:label="$t('COUNTRIES')"
			hideComparison
		/>
	</div>
</template>

<script setup lang="ts">
import MetricCard from './MetricCard.vue';
interface Props {
	data: any;
}

withDefaults(defineProps<Props>(), {});
</script>

<style></style>
