<template>
	<div class="container-box relative-position">
		<div :class="[`bg-${color}`]" class="container-box-separator"></div>
		<div class="q-pl-sm full-height">
			<slot></slot>
		</div>
	</div>
</template>

<script setup lang="ts">
interface Props {
	color: string;
}

const props = withDefaults(defineProps<Props>(), {
	color: 'ink-3'
});
</script>

<style lang="scss" scoped>
.container-box {
	padding-left: 2px;
	.container-box-separator {
		position: absolute;
		top: 0;
		bottom: 0;
		left: 0;
		width: 2px;
	}
}
</style>
