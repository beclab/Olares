<template>
	<div
		class="header row items-center justify-between q-px-md text-ink-1 q-gutter-x-lg"
	>
		<span>{{ title }}</span>
		<q-btn dense flat color="ink-3" size="xs" v-close-popup>
			<q-icon name="sym_r_close" style="font-size: 16px" />
		</q-btn>
	</div>
</template>

<script setup lang="ts">
interface Props {
	title?: string;
}

withDefaults(defineProps<Props>(), {});
</script>

<style lang="scss" scoped>
.header {
	color: ink-1;
	font-size: 12px;
	line-height: 16px;
	font-weight: 500;
	height: 32px;
	background: $background-3;
}
</style>
