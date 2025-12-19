<template>
	<div
		class="text-center full-width my-empty2-container"
		:class="{ 'my-empty-container-center': center }"
	>
		<span class="empty-content" :class="{ 'empty-content-center': center }">
			{{ $t('NO_DATA_AVAILABLE') }}</span
		>
	</div>
</template>

<script setup lang="ts">
interface Props {
	center?: boolean;
}

const props = withDefaults(defineProps<Props>(), {});
</script>

<style lang="scss" scoped>
.my-empty2-container {
	&.empty-content-center {
		position: absolute;
		top: 50%;
		left: 50%;
		transform: translate(-50%, -50%);
	}
	.empty-content {
		color: $ink-2;
		&.empty-content-center {
			position: absolute;
			top: 50%;
			left: 50%;
			transform: translate(-50%, -50%);
		}
	}
}
</style>
