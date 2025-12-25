<template>
	<div class="row">
		<MetricCard :value="data.pageviews?.length" label="Views" hideComparison />
		<MetricCard
			:value="data.visitors?.length"
			label="Visitors"
			hideComparison
		/>
		<MetricCard :value="data.events?.length" label="Events" hideComparison />
		<MetricCard
			:value="data.countries?.length"
			label="Countries"
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
