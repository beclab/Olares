<template>
	<div
		class="reading-progress-root row justify-start"
		:style="{ height: height }"
	>
		<div
			class="progress-gradient-line"
			:style="{ width: Number(props.percentage).toFixed(2) + '%' }"
		/>
	</div>
</template>

<script setup lang="ts">
const props = defineProps({
	percentage: { type: Number, default: 100 },
	height: { type: String, default: '1px' }
});
</script>

<style lang="scss" scoped>
.reading-progress-root {
	width: 100%;
	background: $separator;

	.progress-gradient-line {
		height: 100%;
		background: linear-gradient(to right, $background-3, $orange-default);
		transition: width 0.2s ease;
	}
}
</style>
