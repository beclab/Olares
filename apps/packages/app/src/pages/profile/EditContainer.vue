<template>
	<q-scroll-area class="design-scroll" :thumb-style="scrollBarStyle.thumbStyle">
		<div class="design-area column">
			<slot />
		</div>
	</q-scroll-area>
</template>

<script lang="ts" setup>
import { scrollBarStyle } from 'src/utils/helper';
</script>

<style scoped lang="scss">
.design-scroll {
	width: 100%;
	height: 100%;

	.design-area {
		width: 100%;
		height: 100%;
		padding-left: var(--page-editor-padding);
		padding-right: var(--page-editor-padding);
		padding-bottom: var(--page-editor-padding-bottom);
	}
}
</style>
