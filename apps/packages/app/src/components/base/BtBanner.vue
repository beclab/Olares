<template>
	<div class="banner-root">
		<!--		@autoplayTimeLeft="onAutoplayTimeLeft"-->
		<swiper
			:centeredSlides="true"
			:autoplay="{
				delay: 2500,
				disableOnInteraction: false
			}"
			:style="{ width: swiperSize + 'px', borderRadius: '12px' }"
			:navigation="false"
			:modules="modules"
			class="banner-swiper"
			@swiper="setSwiperRef"
			@slideChange="onSlideChange"
		>
			<swiper-slide
				class="banner-swiper-slide"
				v-for="(item, index) in size"
				:key="index"
			>
				<slot name="swiper" :index="index" />
			</swiper-slide>
			<div
				v-if="size > 1"
				class="cursor-pointer banner-navigation row justify-start items-center"
			>
				<template v-for="(item, index) in size" :key="item">
					<div
						:class="
							index === currentSlide ? 'navigation-active' : 'navigation-normal'
						"
						@click="changeActive(index)"
					/>
				</template>
			</div>
			<!--			<template #container-end>-->

			<!--				<template v-for="(item, index) in size" :key="item">-->
			<!--					-->
			<!--				</template>-->

			<!--				<div class="autoplay-progress">-->
			<!--					<svg viewBox="0 0 48 48" ref="progressCircle">-->
			<!--						<circle cx="24" cy="24" r="20" />-->
			<!--					</svg>-->
			<!--					<span ref="progressContent"></span>-->
			<!--				</div>-->
			<!--			</template>-->
		</swiper>
	</div>
</template>
<script lang="ts" setup>
import { onBeforeUnmount, onMounted, ref } from 'vue';
// Import Swiper Vue.js components
import { Swiper, SwiperSlide } from 'swiper/vue';

// Import Swiper styles
import 'swiper/css';
import 'swiper/css/navigation';
import 'swiper/css/autoplay';
import { Navigation, Autoplay } from 'swiper/modules';
import { useQuasar } from 'quasar';

defineProps({
	size: {
		type: Number,
		require: true
	}
});
// import required modules
const modules = [Navigation, Autoplay];
const $q = useQuasar();
// const progressCircle = ref<string | null>(null);
// const progressContent = ref<string | null>(null);
const swiperSize = ref();
let swiperRef: any = null;
let resizeTimer: NodeJS.Timeout | null = null;
const currentSlide = ref(0);
// const onAutoplayTimeLeft = (s, time, progress) => {
// 	progressCircle.value.style.setProperty('--progress', 1 - progress);
// 	progressContent.value.textContent = `${Math.ceil(time / 1000)}s`;
// };

const setSwiperRef = (swiper: any) => {
	swiperRef = swiper;
};

const changeActive = (index: number) => {
	swiperRef.slideTo(index);
};

function onSlideChange(swiper) {
	currentSlide.value = swiper.activeIndex;
}

const resize = () => {
	if (resizeTimer) {
		clearTimeout(resizeTimer);
	}
	resizeTimer = setTimeout(function () {
		updateSwiper();
	}, 200);
};

const updateSwiper = () => {
	if ($q.screen.width < 864) {
		swiperSize.value = 800;
	} else if ($q.screen.width < 1024) {
		swiperSize.value = $q.screen.width - 88;
	} else {
		swiperSize.value = $q.screen.width - 240 - 88;
	}
};

onMounted(async () => {
	updateSwiper();
	window.addEventListener('resize', resize);
});

onBeforeUnmount(() => {
	window.removeEventListener('resize', resize);
});
</script>

<style lang="scss" scoped>
.banner-root {
	width: 100%;
	max-width: 100%;
	height: auto;

	.banner-swiper {
		position: relative;
		max-width: calc(100% - 88px);
		height: 100%;

		.banner-swiper-slide {
			max-width: 100%;
			text-align: center;
			font-size: 18px;
			background: #fff;
			display: flex;
			justify-content: center;
			align-items: center;
		}

		.banner-swiper-slide img {
			display: block;
			width: 100%;
			height: 100%;
			object-fit: cover;
		}

		.banner-navigation {
			width: auto;
			height: 20px;
			right: 44px;
			bottom: 32px;
			position: absolute;
			z-index: 1000;

			.navigation-base {
				height: 8px;
				border-radius: 20px;
				margin-right: 8px;
			}

			.navigation-active {
				@extend .navigation-base;
				width: 32px;
				background-color: #ffffff;
			}

			.navigation-normal {
				@extend .navigation-base;
				width: 8px;
				background: rgba(255, 255, 255, 0.4);
			}
		}
	}
}

//.autoplay-progress {
//	position: absolute;
//	right: 16px;
//	bottom: 16px;
//	z-index: 10;
//	width: 48px;
//	height: 48px;
//	display: flex;
//	align-items: center;
//	justify-content: center;
//	font-weight: bold;
//	color: var(--swiper-theme-color);
//}
//
//.autoplay-progress svg {
//	--progress: 0;
//	position: absolute;
//	left: 0;
//	top: 0px;
//	z-index: 10;
//	width: 100%;
//	height: 100%;
//	stroke-width: 4px;
//	stroke: var(--swiper-theme-color);
//	fill: none;
//	stroke-dashoffset: calc(125.6px * (1 - var(--progress)));
//	stroke-dasharray: 125.6;
//	transform: rotate(-90deg);
//}
</style>
