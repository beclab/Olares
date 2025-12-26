<template>
	<q-btn v-bind="$attrs" no-caps class="q-py-sm custom-btn-container">
		<span class="text-subtitle2" :class="[`text-${textColor}`]">
			<slot name="label" v-if="$slots.label"></slot>
			<span v-else>{{ label }}</span>
		</span>
	</q-btn>
</template>

<script setup lang="ts">
interface Props {
	label?: string;
	textColor?: string;
}
const props = withDefaults(defineProps<Props>(), {
	textColor: 'ink-on-brand-black'
});
</script>

<style lang="scss" scoped>
.custom-btn-container.q-btn--outline:before {
	border-color: $separator-2;
}
</style>
