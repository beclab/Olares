<template>
	<div
		:style="{
			'--NavigationOffsite': `${navigationOffsite}px`,
			'--paddingAll': paddingX * 2 + 'px'
		}"
		class="swiper-root row justify-center items-center"
	>
		<!-- custom back button -->
		<div
			class="button-left"
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
import {
	computed,
	onBeforeUnmount,
	onMounted,
	PropType,
	ref,
	watch
} from 'vue';
import { getRequireImage } from '../../utils/imageUtils';
import { Swiper, SwiperSlide } from 'swiper/vue';
import { useQuasar } from 'quasar';
import 'swiper/css';
import 'swiper/css/pagination';
import 'swiper/css/navigation';
import 'swiper/css/virtual';
import { useI18n } from 'vue-i18n';
import { useRoute } from 'vue-router';
import { useDeviceStore } from '../../stores/settings/device';

const modules = [Pagination, Navigation, Virtual];
const canNext = ref(false);
const canPrev = ref(false);
const $q = useQuasar();
const swiperSize = ref();
let swiperRef: any = null;
const route = useRoute();
let resizeTimer: NodeJS.Timeout | null = null;

const props = defineProps({
	dataArray: {
		type: Object as PropType<any[]>,
		require: true
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
});

onBeforeUnmount(() => {
	window.removeEventListener('resize', resize);
});

const showMenu = computed(() => {
	const menuParam = route.query.showMenu;
	return menuParam === undefined ? true : menuParam === 'true';
});

const slideChange = () => {
	if (props.dataArray) {
		canPrev.value = swiperRef.activeIndex !== 0;
		canNext.value =
			swiperRef.activeIndex !== props.dataArray.length - showAppSize.value;
	}
};

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
		swiperSize.value = $q.screen.width - props.paddingX * 2;
	} else {
		swiperSize.value = $q.screen.width - (showMenu.value ? 240 : 0);
		props.paddingX * 2;
	}
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
// 自定义上一个按钮点击事件
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

	.swiper {
		max-width: calc(100% - var(--paddingAll));
		height: auto;
	}
}
</style>
