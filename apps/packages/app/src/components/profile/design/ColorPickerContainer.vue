<template>
	<div class="color-picker-container column justify-start items-start">
		<span class="color-picker-container__title">{{ title }}</span>
		<slot />
	</div>
</template>

<script lang="ts" setup>
defineProps({
	title: {
		type: String,
		required: true
	}
});
</script>

<style scoped lang="scss">
.color-picker-container {
	width: 100%;
	height: auto;
	padding: 12px;
	border-radius: 8px;
	margin-top: 12px;
	border: 1px solid $separator-2;

	&__title {
		overflow: hidden;
		color: $ink-1;
		text-overflow: ellipsis;
		font-family: Roboto;
		font-size: 12px;
		font-weight: 500;
		line-height: 16px;
		letter-spacing: 0em;
		text-align: left;
	}
}
</style>
