<template>
	<q-img
		v-if="loading"
		class="loading-image"
		:style="{
			width: size,
			height: size
		}"
		:src="srcPath"
	/>
</template>

<script lang="ts" setup>
import { ref } from 'vue';

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

const srcPath = ref(
	process.env.APPLICATION?.toString().toLowerCase() + '/loading_image.svg'
);
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
