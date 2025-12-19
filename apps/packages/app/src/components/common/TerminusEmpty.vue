<template>
	<div class="empty-c-root column items-center">
		<img class="img" :src="getRequireImage(imgPath)" />
		<div class="txt-body3 content-info">
			{{ info }}
		</div>
		<div v-if="$slots.add">
			<slot name="add"></slot>
		</div>
	</div>
</template>

<script setup lang="ts">
import { getRequireImage } from '../../utils/imageUtils';

defineProps({
	imgPath: {
		type: String,
		required: false,
		default: ''
	},
	info: {
		type: String,
		required: false,
		default: 'Empty data'
	}
});
</script>

<style scoped lang="scss">
.empty-c-root {
	width: 100%;

	.img {
		width: 160px;
		height: 160px;
	}

	.content-info {
		max-width: 247px;
		text-align: center;
	}
}
</style>
