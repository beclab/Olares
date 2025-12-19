<template>
	<div class="custom-btn q-px-sm row items-center justify-center">
		<q-icon :name="icon" size="16px" v-if="icon && icon.length > 0" />
		<span class="operate-btn-label">{{ label }}</span>
	</div>
</template>

<script setup lang="ts">
defineProps({
	icon: {
		type: String,
		required: false,
		default: ''
	},
	label: {
		type: String,
		required: false,
		default: ''
	}
});
</script>

<style scoped lang="scss">
.custom-btn {
	height: 32px;
	text-align: center;
	border-radius: 8px;
	border: 1px solid $btn-stroke;
	overflow: hidden;
	color: $ink-1;

	&:hover {
		background: $background-hover;
	}
	&.operate-disabled {
		opacity: 0.5;
	}

	.operate-btn-label {
		line-height: 32px;
		padding: 0 8px;
		color: $ink-1;
		font-size: 12px;
		line-height: 100%;
		text-align: center;
		cursor: pointer;
	}
}
</style>
