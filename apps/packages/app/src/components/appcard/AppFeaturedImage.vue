<template>
	<div
		v-if="skeleton"
		:style="{ '--width': width > 0 ? `${width}px` : '100%' }"
		class="workflow-featured-image-skeleton-box"
	>
		<q-skeleton class="workflow-featured-image-skeleton" />
	</div>

	<q-img
		v-else
		:src="src ? src : getRequireImage('icon_background.svg')"
		:style="{ '--width': width > 0 ? `${width}px` : '100%' }"
		class="workflow-featured-image"
		:ratio="16 / 10"
	>
		<template v-slot:loading>
			<q-skeleton class="workflow-featured-image" />
		</template>
		<div v-if="!src" class="workflow-icon-mask row justify-center items-center">
			<slot />
		</div>
	</q-img>
</template>

<script lang="ts" setup>
import { getRequireImage } from '@apps/market/src/utils/imageUtils';

defineProps({
	src: {
		type: String,
		require: true
	},
	skeleton: {
		type: Boolean,
		default: false
	},
	width: {
		type: Number,
		default: 0
	}
});
</script>

<style scoped lang="scss">
.workflow-featured-image-skeleton-box {
	width: var(--width);
	padding-top: 62.5%;
	position: relative;

	.workflow-featured-image-skeleton {
		position: absolute;
		border-radius: 8px;
		top: 0;
		width: 100%;
		height: 100%;
	}
}

.workflow-featured-image {
	width: var(--width);
	height: 100%;
	border-radius: 12px;

	.workflow-icon-mask {
		background: transparent;
		width: 100%;
		height: 100%;
	}
}
</style>
