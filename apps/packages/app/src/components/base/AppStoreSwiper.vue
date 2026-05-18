<template>
	<div
		ref="swiperRootRef"
		:style="{
			'--NavigationOffsite': `${navigationOffsite}px`,
			'--paddingAll': paddingX * 2 + 'px'
		}"
		class="swiper-root row justify-center items-center"
	>
		<!-- custom back button -->
		<div
			class="button-left"
			:class="canPrev ? 'button-cursor' : ''"
			@click="customPrev"
			v-if="!deviceStore.isMobile && dataArray.length > showAppSize"
		>
			<img
				:src="
					canPrev
						? getRequireImage('swiper/swiper_prev_normal.svg')
						: getRequireImage('swiper/swiper_prev_disable.svg')
				"
			/>
		</div>

		<swiper
			:modules="modules"
			:slidesPerView="showAppSize"
			:centeredSlides="false"
			:spaceBetween="20"
			:navigation="false"
			:style="{ width: swiperSize + 'px' }"
			:set-wrapper-size="true"
			class="swiper"
			@slideChange="slideChange"
			@swiper="setSwiperRef"
		>
			<swiper-slide
				style="max-width: 100%"
				v-for="(item, index) in dataArray"
				:key="index"
				:virtualIndex="index"
			>
				<slot name="swiper" :item="item" :index="index" />
			</swiper-slide>
		</swiper>

		<!--  custom forward button  -->
		<div
			class="button-right"
			:class="canNext ? 'button-cursor' : ''"
			@click="customNext"
			v-if="!deviceStore.isMobile && dataArray.length > showAppSize"
		>
			<img
				:src="
					canNext
						? getRequireImage('swiper/swiper_next_normal.svg')
						: getRequireImage('swiper/swiper_next_disable.svg')
				"
			/>
		</div>
	</div>
</template>

<script lang="ts" setup>
import { Navigation, Pagination, Virtual } from 'swiper/modules';
import { onBeforeUnmount, onMounted, PropType, ref, watch } from 'vue';
import { getRequireImage } from '../../utils/imageUtils';
import { Swiper, SwiperSlide } from 'swiper/vue';
import { useQuasar } from 'quasar';
import 'swiper/css';
import 'swiper/css/pagination';
import 'swiper/css/navigation';
import 'swiper/css/virtual';
import { useDeviceStore } from '../../stores/settings/device';

const modules = [Pagination, Navigation, Virtual];
const canNext = ref(false);
const canPrev = ref(false);
const $q = useQuasar();
const swiperSize = ref();
let swiperRef: any = null;
let resizeRafId: number | null = null;
const swiperRootRef = ref<HTMLElement | null>(null);
let rootResizeObserver: ResizeObserver | null = null;

const props = defineProps({
	dataArray: {
		type: Array as PropType<any[]>,
		required: true
	},
	slidesPerView: {
		type: Number,
		default: 0
	},
	initialSlide: {
		type: Number,
		default: 0
	},
	navigationOffsite: {
		type: Number,
		default: 0
	},
	paddingX: {
		type: Number,
		default: 0
	},
	showSize: {
		type: String,
		default: '5,3,2'
	},
	ratio: {
		type: Number,
		default: 0
	},
	maxHeight: {
		type: Number,
		default: 0
	}
});

const deviceStore = useDeviceStore();
const sizeArray = deviceStore.isMobile ? ['1'] : props.showSize.split(',');
const showAppSize = ref(
	deviceStore.isMobile
		? Number(sizeArray[0])
		: $q.screen.lg || $q.screen.xl
		? Number(sizeArray[0])
		: $q.screen.md
		? Number(sizeArray[1])
		: Number(sizeArray[2])
);

onMounted(async () => {
	updateSwiper();
	window.addEventListener('resize', resize);
	rootResizeObserver = new ResizeObserver(() => {
		resize();
	});
	if (swiperRootRef.value) {
		rootResizeObserver.observe(swiperRootRef.value);
	}
});

onBeforeUnmount(() => {
	window.removeEventListener('resize', resize);
	if (resizeRafId !== null) {
		window.cancelAnimationFrame(resizeRafId);
		resizeRafId = null;
	}
	if (rootResizeObserver) {
		rootResizeObserver.disconnect();
		rootResizeObserver = null;
	}
});

const slideChange = () => {
	if (props.dataArray) {
		canPrev.value = swiperRef.activeIndex !== 0;
		canNext.value =
			swiperRef.activeIndex !== props.dataArray.length - showAppSize.value;
	}
};

const resize = () => {
	if (resizeRafId !== null) {
		return;
	}
	resizeRafId = window.requestAnimationFrame(() => {
		resizeRafId = null;
		updateSwiper();
	});
};

const updateSwiper = () => {
	const rootWidth = swiperRootRef.value?.clientWidth || $q.screen.width;
	swiperSize.value = Math.max(rootWidth - props.paddingX * 2, 0);
	// console.log(swiperSize.value);
	if (props.maxHeight && props.ratio) {
		const height = swiperSize.value / props.ratio;
		if (height > props.maxHeight) {
			swiperSize.value = props.maxHeight * props.ratio;
		}
	}
	showAppSize.value =
		props.slidesPerView === 0
			? deviceStore.isMobile
				? Number(sizeArray[0])
				: $q.screen.lg || $q.screen.xl
				? Number(sizeArray[0])
				: $q.screen.md
				? Number(sizeArray[1])
				: Number(sizeArray[2])
			: props.slidesPerView;
	if (swiperRef) {
		swiperRef.update();
		const maxStartIndex = Math.max(
			0,
			(props.dataArray?.length || 0) - showAppSize.value
		);
		if (swiperRef.activeIndex > maxStartIndex) {
			swiperRef.slideTo(maxStartIndex, 0);
		}
	}
	slideChange();
};

watch(
	() => [props.maxHeight, props.ratio],
	(newValue) => {
		if (newValue) {
			updateSwiper();
		}
	}
);

const setSwiperRef = (swiper: any) => {
	swiperRef = swiper;
	slideTo(props.initialSlide);
	slideChange();
};

const slideTo = (index: number) => {
	if (!props.dataArray) {
		return;
	}
	if (index >= props.dataArray.length || index < 0) {
		return;
	}
	swiperRef.slideTo(index);
};

const customNext = () => {
	if (swiperRef) {
		swiperRef.slideNext();
	}
};

const customPrev = () => {
	if (swiperRef) {
		swiperRef.slidePrev();
	}
};

defineExpose({ slideTo });
</script>

<style scoped lang="scss">
.swiper-root {
	width: 100% !important;
	max-width: 100% !important;
	height: auto;
	position: relative;

	.button-left {
		position: absolute;
		top: calc(50% - var(--NavigationOffsite));
		left: 25px;
	}

	.button-right {
		position: absolute;
		top: calc(50% - var(--NavigationOffsite));
		right: 25px;
	}

	.button-cursor {
		cursor: pointer;
	}

	.swiper {
		max-width: calc(100% - var(--paddingAll));
		height: auto;
	}
}
</style>
