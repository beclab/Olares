<template>
	<div
		class="bt-badgecount-container text-subtitle3 q-px-sm text-ink-2"
		:class="[active ? activeClass : '']"
	>
		<slot></slot>
	</div>
</template>

<script setup lang="ts">
interface Props {
	active: boolean;
	activeClass: string;
}

withDefaults(defineProps<Props>(), {});
</script>

<style lang="scss" scoped>
.bt-badgecount-container {
	display: inline-block;
	background: $background-hover;
	border-radius: 8px;
}
</style>
