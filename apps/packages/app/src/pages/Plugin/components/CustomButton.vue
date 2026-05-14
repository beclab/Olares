<template>
	<q-btn v-bind="$attrs" no-caps class="q-py-sm custom-btn-container">
		<div class="row items-center no-wrap q-gutter-x-xs custom-btn-content">
			<q-icon
				:name="iconFont"
				:color="textColor"
				v-if="iconFont"
				:size="iconSize"
			/>
			<q-img
				v-else-if="icon"
				:src="icon"
				:width="iconSize"
				:height="iconSize"
				fit="contain"
				spinner-size="0px"
			/>
			<span class="text-subtitle2" :class="[`text-${textColor}`]">
				<slot name="label" v-if="$slots.label"></slot>
				<span v-else>{{ label }}</span>
			</span>
		</div>
	</q-btn>
</template>

<script setup lang="ts">
interface Props {
	label?: string;
	textColor?: string;
	icon?: string;
	iconSize?: string;
	iconFont?: string;
}
const props = withDefaults(defineProps<Props>(), {
	textColor: 'ink-on-brand-black',
	iconSize: '16px'
});
</script>

<style lang="scss" scoped>
.custom-btn-container.q-btn--outline:before {
	border-color: $separator-2;
}
.custom-btn-content {
	white-space: nowrap;
}
</style>
