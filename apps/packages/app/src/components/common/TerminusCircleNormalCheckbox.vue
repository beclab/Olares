<template>
	<div
		:style="{
			width: `${size}px`,
			height: `${size}px`,
			border: `1px solid ${effectiveBorderColor}`,
			borderRadius: '50%'
		}"
	></div>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { useColor } from '@bytetrade/ui';

const props = defineProps({
	modelValue: {
		type: Boolean,
		default: false
	},
	size: {
		type: Number,
		default: 16
	},
	borderColor: {
		type: String,
		default: undefined
	}
});

// Resolve the fallback color inside the component's setup context so that
// `useQuasar()` / `watchEffect()` (used by `useColor`) actually have an
// active effect scope. Putting this expression directly in a `defineProps`
// default would evaluate it at module-load time, returning either a stale
// snapshot or `undefined` (rendering `border: 1px solid undefined`).
const fallbackBorderColor = useColor('background-5').color;
const effectiveBorderColor = computed(
	() => props.borderColor ?? fallbackBorderColor.value
);
</script>

<style scoped lang="scss"></style>
