<template>
	<div
		:class="{
			'shake-mode': shake
		}"
	>
		<slot></slot>
	</div>
</template>

<script setup lang="ts">
interface Props {
	shake?: boolean;
}

defineProps<Props>();
</script>

<style lang="scss" scoped>
@keyframes shake {
	0% {
		transform: rotate(0deg);
	}
	25% {
		transform: rotate(-10deg);
	}
	75% {
		transform: rotate(10deg);
	}
	100% {
		transform: rotate(0deg);
	}
}

.shake-mode {
	transform-origin: center center;
	animation: shake 0.2s ease-in-out 3;
	will-change: transform;
}
</style>
