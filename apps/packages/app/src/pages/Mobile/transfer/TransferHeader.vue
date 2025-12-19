<template>
	<div class="transfer-header row items-center justify-between">
		<div class="text-subtitle2 text-ink-3">
			{{ title }}
		</div>
		<slot name="buttons" />
	</div>
</template>

<script setup lang="ts">
defineProps({
	title: {
		type: String,
		default: '',
		required: false
	}
});
</script>

<style scoped lang="scss">
.transfer-header {
	width: 100%;
	height: 20px;
}
</style>
