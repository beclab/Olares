<template>
	<q-item
		clickable
		v-close-popup
		class="file-operation-item q-pb-md"
		@click="emit('onItemClick')"
	>
		<div class="file-operation-div row justify-start items-center text-ink-1">
			<q-icon :name="icon" size="20px" />
			<div class="file-operation-text text-body3">{{ label }}</div>
		</div>
	</q-item>
</template>

<script lang="ts" setup>
const props = defineProps({
	icon: String,
	label: String
});

const emit = defineEmits(['onItemClick']);
</script>

<style scoped lang="scss">
.file-operation-item {
	height: 36px;
	margin: 0;
	padding: 0;
	border-radius: 4px;

	.file-operation-div {
		width: 100%;
		height: 100%;
		padding: 8px;

		.file-operation-text {
			margin-left: 8px;
		}
	}
}
</style>
