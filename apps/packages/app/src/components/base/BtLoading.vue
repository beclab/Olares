<template>
	<q-img
		v-if="loading"
		class="loading-image"
		:style="{
			width: size,
			height: size
		}"
		:src="getRequireImage('loading_image.svg')"
	/>
</template>

<script lang="ts" setup>
import { getRequireImage } from 'src/utils/rss-utils';

defineProps({
	loading: {
		type: Boolean,
		require: true,
		default: false
	},
	size: {
		type: String,
		default: '16px'
	}
});
</script>

<style lang="scss" scoped>
@keyframes animate {
	0% {
		transform: rotate(0deg);
		-webkit-transform: rotate(0deg);
	}
	100% {
		transform: rotate(360deg);
		-webkit-transform: rotate(360deg);
	}
}
@-webkit-keyframes animate {
	0% {
		transform: rotate(0deg);
		-webkit-transform: rotate(0deg);
	}
	100% {
		transform: rotate(360deg);
		-webkit-transform: rotate(360deg);
	}
}

.loading-image {
	animation: animate 1.2s linear infinite;
	-webkit-animation: animate 1.2s linear infinite;
}
</style>
