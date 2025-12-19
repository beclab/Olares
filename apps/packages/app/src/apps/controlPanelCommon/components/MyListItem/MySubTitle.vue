<template>
	<div class="sub-title ellipsis full-width">
		<template v-if="$slots.default">
			<slot></slot>
		</template>
		<template v-else>
			{{ title }}
		</template>
	</div>
</template>

<script setup lang="ts">
interface Props {
	title?: string;
}

withDefaults(defineProps<Props>(), {});
</script>
<style lang="scss" scoped>
.sub-title {
	overflow: hidden;
	color: $ink-3;
	font-size: 10px;
	line-height: 12px;
}
</style>
