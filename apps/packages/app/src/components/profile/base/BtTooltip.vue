<template>
	<q-tooltip
		class="tooltip-normal"
		style="white-space: nowrap"
		:offset="[0, 4]"
	>
		{{ label }}
	</q-tooltip>
</template>

<script lang="ts" setup>
defineProps({
	label: {
		type: String,
		require: false
	}
});
</script>

<style lang="scss" scoped></style>
