<template>
	<div
		id="logoLoadingToast"
		v-show="show"
		:style="{ background: backgroundColor }"
	>
		<div class="loading">
			<slot name="logo"></slot>
		</div>
		<p class="loadingContent" v-if="text" :style="{ color: textColor }">
			{{ text }}<span class="loadingTxt">{{ loadingTxt }}</span>
		</p>
	</div>
</template>

<script lang="ts">
import { defineComponent, ref, onMounted, onUnmounted } from 'vue';
export default defineComponent({
	name: 'BtLogoLoading',
	props: {
		show: {
			type: Boolean,
			default: false
		},
		text: {
			type: String,
			default: 'Loading...'
		},
		color: {
			type: String,
			default: '#1d953f'
		},
		backgroundColor: {
			type: String,
			default: 'rgba(0, 0, 0, 0.4)'
		},
		textColor: {
			type: String,
			default: '#1f1f1f'
		}
	},

	setup() {
		const timer = ref();
		const loadingTxt = ref('...');

		const startTimer = () => {
			let text = '...';
			timer.value = window.setInterval(() => {
				if (text === '...') {
					text = '.';
				} else {
					text = text + '.';
				}
				loadingTxt.value = text;
			}, 500);
		};

		onMounted(() => {
			startTimer();
		});

		onUnmounted(() => {
			window.clearInterval(timer.value);
		});

		return {
			loadingTxt
		};
	}
});
</script>

<style lang="scss" scoped>
li {
	list-style-type: none;
}
#logoLoadingToast {
	width: 100%;
	height: 100%;
	display: flex;
	flex-direction: column;
	align-items: center;
	justify-content: center;
	position: relative;

	.loadingContent {
		font-size: 18px;
		font-weight: 800;
		margin: 0;

		.loadingTxt {
			display: inline-block;
			text-align: left;
			min-width: 20px;
		}
	}
}
.loading {
	width: 100px;
	height: 100px;
	display: flex;
	align-items: center;
	justify-content: center;

	animation: rotate 2s infinite ease-in-out;

	@keyframes rotate {
		100% {
			transform: rotatey(360deg);
		}
	}
}
</style>
