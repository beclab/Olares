<template>
	<div class="progress-button" :style="{ backgroundColor: backgroundColor }">
		<span
			class="progress-text"
			:class="textClass"
			:style="{
				background: `linear-gradient(to right, ${coveredTextColor} ${computedProgress}%, ${defaultTextColor} ${computedProgress}%)`,
				WebkitBackgroundClip: 'text',
				color: 'transparent'
			}"
		>
			{{ buttonText ? buttonText : `${computedProgress}%` }}
		</span>
		<div
			class="progress-bar"
			:style="{
				width: computedProgress + '%',
				backgroundColor: progressBarColor
			}"
			:class="progressBarClass"
		/>
	</div>
</template>

<script lang="ts" setup>
import { computed, ref, watch } from 'vue';

const props = defineProps({
	buttonText: {
		type: String,
		required: false
	},
	textClass: String,
	backgroundColor: {
		type: String,
		required: false,
		default: 'transparent'
	},
	progress: {
		type: String,
		default: '0'
	},
	defaultTextColor: {
		type: String,
		default: '#333'
	},
	coveredTextColor: {
		type: String,
		default: '#fff'
	},
	progressBarColor: {
		type: String,
		default: '#4caf50'
	},
	progressBarClass: {
		type: String,
		default: '#4caf50'
	}
});

let interval: NodeJS.Timer | null = null;
const testValue = ref(0);

const computedProgress = computed(() => {
	try {
		const result = Number(props.progress) + testValue.value;
		return Math.min(100, Math.max(0, result));
	} catch (e) {
		return 0;
	}
});

function startProgress() {
	if (interval) return;

	interval = setInterval(() => {
		if (computedProgress.value < 100) {
			console.log(`${computedProgress.value}%`);
			testValue.value += 1;
		} else {
			if (interval !== null) {
				clearInterval(interval);
			}
			testValue.value = 0;
			interval = null;
		}
	}, 1000);
}
const emit = defineEmits(['onLoadComplete']);
watch(
	() => computedProgress.value,
	() => {
		if (computedProgress.value === 100) {
			emit('onLoadComplete');
		}
	}
);
defineExpose({ startProgress });
</script>

<style lang="scss">
.progress-button {
	position: relative;
	overflow: hidden;
	display: flex;
	justify-content: center;
	align-items: center;
	height: 100%;
	width: 100%;

	.progress-bar {
		position: absolute;
		top: 0;
		left: 0;
		height: 100%;
		width: 0;
		z-index: 1;
		transition: width 0.1s linear;
	}

	.progress-text {
		position: relative;
		z-index: 2;
		background-clip: text;
		-webkit-background-clip: text;
		text-align: center;
		color: transparent;
		width: 100%;
	}
}
</style>
