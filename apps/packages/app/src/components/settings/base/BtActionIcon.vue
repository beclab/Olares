<template>
	<div
		class="bt-custom-icon-root row justify-center items-center"
		:style="{
			'--icon-width': size + 'px',
			'--icon-height': size + 'px'
		}"
	>
		<q-icon :class="iconClasses" :name="name" :size="iconSize + 'px'" />
	</div>
</template>

<script lang="ts" setup>
defineProps({
	name: {
		type: String,
		require: true
	},
	size: {
		type: Number,
		required: false,
		default: 32
	},
	iconSize: {
		type: Number,
		required: false,
		default: 20
	},
	iconClasses: {
		type: String,
		required: false,
		default: 'text-ink-2'
	}
});
</script>

<style scoped lang="scss">
.bt-custom-icon-root {
	width: var(--icon-width);
	height: var(--icon-height);
	border-radius: 4px;
	&:hover {
		// filter: brightness(0.9);
		background-color: $background-3;
	}
}
</style>
