<template>
	<template v-if="isAnimated">
		<div class="container center_animation">
			<div
				class="animation_img background_size"
				:style="{ backgroundImage: `url(${picture})` }"
			></div>
			<div class="center">
				<div class="wrap_wrap">
					<div class="box-1 box-display">
						<div
							class="font_download_img background_size"
							:style="{ backgroundImage: `url(${certificate})` }"
						></div>
						<div
							class="font_download_img background_size"
							:style="{ backgroundImage: `url(${certificate})` }"
						></div>
					</div>
					<div class="box-2 box-display">
						<div
							class="font_download_img background_size"
							:style="{ backgroundImage: `url(${certificate})` }"
						></div>
						<div
							class="font_download_img background_size"
							:style="{ backgroundImage: `url(${certificate})` }"
						></div>
					</div>
				</div>
			</div>
		</div>
	</template>
	<template v-else>
		<div class="container">
			<div
				class="animation_img background_size"
				:style="{ backgroundImage: `url(${picture})` }"
			></div>
			<div class="center">
				<div class="wrap_wrap">
					<div class="box-display">
						<div
							class="font_download_img background_size"
							:style="{ backgroundImage: `url(${certificate})` }"
						></div>
						<div
							class="font_download_img background_size"
							:style="{ backgroundImage: `url(${certificate})` }"
						></div>
					</div>
				</div>
			</div>
		</div>
	</template>
</template>

<script lang="ts">
import { defineComponent } from 'vue';

export default defineComponent({
	name: 'animationPage',
	props: {
		picture: {
			type: String,
			default: ''
		},
		certificate: {
			type: String,
			default: ''
		},
		isAnimated: {
			type: Boolean,
			default: false
		}
	},

	setup() {
		return {
			isFailed: true
		};
	}
});
</script>
<style scoped lang="scss">
.font_download_img {
	margin: 14px;
	width: 32px;
	height: 32px;
}
.container {
	position: relative;
}
.background_size {
	background-position: center;
	background-repeat: no-repeat;
	background-size: contain;
}
.animation_img {
	width: 140px;
	height: 140px;
	position: absolute;
	z-index: 99;
}

.center {
	display: -webkit-box;
	display: -ms-flexbox;
	display: flex;
	-webkit-box-pack: center;
	-ms-flex-pack: center;
	justify-content: center;
	-webkit-box-align: center;
	-ms-flex-align: center;
	align-items: center;
	width: 140px;
	height: 140px;
	border-radius: 100%;
	background-color: $tooltip-color;
	border: 9px solid $tooltip-color;
	overflow: hidden;
}

.wrap_wrap {
	display: flex;
	align-items: center;
	justify-content: center;
}

.box-display {
	display: -webkit-box;
	display: -ms-flexbox;
	display: flex;
	-webkit-box-pack: justify;
	-ms-flex-pack: justify;
	justify-content: space-between;
}

/* animation */

@-webkit-keyframes rowup-1 {
	0% {
		-webkit-transform: translate3d(0, 0, 0);
		transform: translate3d(0, 0, 0);
	}
	100% {
		-webkit-transform: translate3d(50%, 0, 0);
		transform: translate3d(50%, 0, 0);
	}
}

@keyframes rowup-1 {
	0% {
		-webkit-transform: translate3d(0, 0, 0);
		transform: translate3d(0, 0, 0);
	}
	100% {
		-webkit-transform: translate3d(50%, 0, 0);
		transform: translate3d(50%, 0, 0);
	}
}

@-webkit-keyframes rowup-2 {
	0% {
		-webkit-transform: translate3d(0, 0, 0);
		transform: translate3d(0, 0, 0);
	}
	100% {
		-webkit-transform: translate3d(50%, 0, 0);
		transform: translate3d(50%, 0, 0);
	}
}

@keyframes rowup-2 {
	0% {
		-webkit-transform: translate3d(0, 0, 0);
		transform: translate3d(0, 0, 0);
	}
	100% {
		-webkit-transform: translate3d(50%, 0, 0);
		transform: translate3d(50%, 0, 0);
	}
}

@keyframes cloud-loop {
	0% {
		-webkit-transform: translate(0, 15px);
		transform: translate(0, 15px);
	}
	100% {
		-webkit-transform: translate(0, 0);
		transform: translate(0, 0);
	}
}

@-webkit-keyframes shadow-loop {
	0% {
		-webkit-transform: translate(0, -35px) scale(1.15, 0.25);
		transform: translate(0, -35px) scale(1.15, 0.25);
	}
	100% {
		-webkit-transform: translate(0, -35px) scale(1, 0.25);
		transform: translate(0, -35px) scale(1, 0.25);
	}
}

@keyframes shadow-loop {
	0% {
		-webkit-transform: translate(0, -35px) scale(1.15, 0.25);
		transform: translate(0, -35px) scale(1.15, 0.25);
	}
	100% {
		-webkit-transform: translate(0, -35px) scale(1, 0.25);
		transform: translate(0, -35px) scale(1, 0.25);
	}
}

.box-1 {
	-webkit-animation: 1s rowup-1 linear infinite normal;
	animation: 1s rowup-1 linear infinite normal;
}

.box-2 {
	-webkit-animation: 1s rowup-2 linear infinite normal;
	animation: 1s rowup-2 linear infinite normal;
}

.center_animation {
	-webkit-animation-name: cloud-loop;
	animation-name: cloud-loop;
	-webkit-animation-duration: 0.7s;
	animation-duration: 0.7s;
	-webkit-animation-iteration-count: infinite;
	animation-iteration-count: infinite;
	-webkit-animation-timing-function: ease;
	animation-timing-function: ease;
	-webkit-animation-direction: alternate;
	animation-direction: alternate;
}
</style>
