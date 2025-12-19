<template>
	<div
		class="row full-width justify-between items-center q-mt-lg"
		style="margin-bottom: -8px"
	>
		<div class="text-subtitle1 text-ink-1">{{ label }}</div>
		<bt-switch
			truthy-track-color="light-green-default"
			:model-value="modelValue"
			@update:model-value="updateModelValue"
			color="light-green-default"
		/>
	</div>
</template>

<script lang="ts" setup>
defineProps({
	label: {
		type: String,
		require: true
	},
	modelValue: {
		type: Boolean,
		require: true
	}
});
const emit = defineEmits(['update:modelValue']);
const updateModelValue = (data: string) => {
	emit('update:modelValue', data);
};
</script>

<style scoped lang="scss"></style>
