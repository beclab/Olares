<template>
	<div class="layoutHeader row items-center justify-start">
		<NavigationComponent :origin_id="origin_id" />
		<BreadcrumbsComponent :origin_id="origin_id" />
	</div>
</template>

<script lang="ts" setup>
import BreadcrumbsComponent from './../../components/files/BreadcrumbsComponent.vue';
import NavigationComponent from './../../components/files/NavigationComponent.vue';

defineProps({
	origin_id: {
		type: Number,
		required: false
	}
});
</script>

<style lang="scss" scoped>
.layoutHeader {
	width: 100%;
	border-bottom: 1px solid $separator;
}
</style>
