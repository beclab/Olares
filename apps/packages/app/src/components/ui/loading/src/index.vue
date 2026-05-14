<template>
	<div id="loadingToast" v-show="show" :style="{ background: backgroundColor }">
		<div class="loading" ref="loading">
			<div
				class="line"
				v-for="(line, index) in 5"
				:key="index"
				:style="{ background: color }"
			></div>
		</div>
		<p class="loadingContent" v-if="text" :style="{ color: textColor }">
			{{ text }}
		</p>
	</div>
</template>

<script lang="ts">
import { defineComponent } from 'vue';
export default defineComponent({
	name: 'BtLoading',
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
	}
});
</script>

<style lang="scss" scoped>
li {
	list-style-type: none;
}
#loadingToast {
	width: 100%;
	height: 100%;
	display: flex;
	flex-direction: column;
	align-items: center;
	justify-content: center;

	.loadingContent {
		font-size: 20px;
		font-weight: 800;
		margin: 0;
	}
}
.loading {
	width: 100px;
	height: 100px;
	background: transparent;
	display: flex;
	align-items: center;
	justify-content: center;
	.line {
		width: 5px;
		height: 40px;
		margin: 0 5px;
		background: #acd8e4;
		border-top-left-radius: 5px;
		border-top-right-radius: 5px;
		border-bottom-left-radius: 5px;
		border-bottom-right-radius: 5px;
	}
	.line:nth-of-type(1) {
		animation: transiteAnimate 0.8s 0s infinite;
	}
	.line:nth-of-type(2) {
		animation: transiteAnimate 0.8s 0.1s infinite;
	}
	.line:nth-of-type(3) {
		animation: transiteAnimate 0.8s 0.2s infinite;
	}
	.line:nth-of-type(4) {
		animation: transiteAnimate 0.8s 0.3s infinite;
	}
	.line:nth-of-type(5) {
		animation: transiteAnimate 0.8s 0.4s infinite;
	}
	.line:nth-of-type(6) {
		animation: transiteAnimate 0.8s 0.5s infinite;
	}
	@keyframes transiteAnimate {
		0% {
			height: 60px;
		}
		50% {
			height: 20px;
		}
		100% {
			height: 60px;
		}
	}
}
</style>
