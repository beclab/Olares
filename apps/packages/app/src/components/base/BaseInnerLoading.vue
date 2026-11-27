<template>
	<q-inner-loading v-if="absolute" :showing="showing">
		<template v-slot:default>
			<div
				:style="{ width: width, height: height }"
				class="column justify-center items-center"
			>
				<q-img class="loading-image" :src="srcPath" />
				<div v-if="label" class="text-subtitle2 text-grey-8 q-mt-lg">
					{{ label }}
				</div>
			</div>
		</template>
	</q-inner-loading>
	<div
		v-else-if="showing"
		:style="{ width: width, height: height }"
		class="column justify-center items-center"
	>
		<q-img class="loading-image" :src="srcPath" />
		<div v-if="label" class="text-subtitle2 text-grey-8 q-mt-lg">
			{{ label }}
		</div>
	</div>
</template>

<script lang="ts" setup>
import { ref } from 'vue';
const props = defineProps({
	width: {
		type: String,
		default: '80px'
	},
	height: {
		type: String,
		default: '80px'
	},
	absolute: {
		type: Boolean,
		default: false
	},
	showing: {
		type: Boolean,
		default: true
	},
	label: {
		type: String,
		default: ''
	},
	src: {
		type: String,
		default: ''
	}
});

const srcPath = ref(
	props.src
		? props.src
		: process.env.APPLICATION.toString().toLowerCase() + '/loading_image.svg'
);
</script>

<style scoped lang="scss">
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
	width: 44px;
	height: 44px;
	animation: animate 1.2s linear infinite;
	-webkit-animation: animate 1.2s linear infinite;
}
</style>
