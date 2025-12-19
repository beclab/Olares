<template>
	<div class="step-scroll-container" ref="containerRef">
		<div class="scroll-wrapper" ref="wrapperRef">
			<div class="scroll-content" ref="contentRef">
				<slot></slot>
			</div>
		</div>
		<div
			class="scroll-btn left-btn btn-show bg-background-1"
			:class="{ disabled: !canScrollLeft }"
			@click="scrollLeft"
			v-show="showButtons"
		>
			<q-icon name="sym_r_chevron_backward" size="16px" color="ink-3" />
		</div>
		<div
			class="scroll-btn right-btn btn-show bg-background-1"
			:class="{ disabled: !canScrollRight }"
			@click="scrollRight"
			v-show="showButtons"
		>
			<q-icon
				name="sym_r_chevron_backward"
				class="arrow-right"
				size="16px"
				color="ink-3"
			/>
		</div>
	</div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue';
import circleLeftIcon from 'src/assets/plugin/circle-left.svg';
import circleRightIcon from 'src/assets/plugin/circle-right.svg';

const props = defineProps({
	step: {
		type: Number,
		default: 100
	}
});

const containerRef = ref<HTMLElement | null>(null);
const wrapperRef = ref<HTMLElement | null>(null);
const contentRef = ref<HTMLElement | null>(null);

const canScrollLeft = ref(false);
const canScrollRight = ref(false);
const showButtons = ref(false);

const currentScroll = ref(0);

const SCROLL_THRESHOLD = 1;

function updateScrollButtons() {
	if (!wrapperRef.value || !contentRef.value) return;

	const { scrollLeft, scrollWidth, clientWidth } = wrapperRef.value;
	currentScroll.value = scrollLeft;

	canScrollLeft.value = scrollLeft > 0;
	canScrollRight.value =
		Math.ceil(scrollLeft + clientWidth + SCROLL_THRESHOLD) < scrollWidth;
	showButtons.value = scrollWidth > clientWidth;
}

const observer = new ResizeObserver(() => {
	updateScrollButtons();
});

function scrollLeft() {
	if (!wrapperRef.value || !canScrollLeft.value) return;
	const newScroll = Math.max(currentScroll.value - props.step, 0);
	wrapperRef.value.scrollTo({
		left: newScroll,
		behavior: 'smooth'
	});
}

function scrollRight() {
	if (!wrapperRef.value || !canScrollRight.value) return;
	const newScroll = currentScroll.value + props.step;
	wrapperRef.value.scrollTo({
		left: newScroll,
		behavior: 'smooth'
	});
}

onMounted(() => {
	updateScrollButtons();
	wrapperRef.value?.addEventListener('scroll', updateScrollButtons);
	if (contentRef.value) {
		observer.observe(contentRef.value);
	}
});

onUnmounted(() => {
	wrapperRef.value?.removeEventListener('scroll', updateScrollButtons);
	observer.disconnect();
});
</script>

<style scoped lang="scss">
.step-scroll-container {
	position: relative;
	width: 100%;
	&:hover .scroll-btn.btn-show {
		display: flex;
	}
}

.scroll-wrapper {
	overflow-x: scroll;
	scrollbar-width: none;
	-ms-overflow-style: none;
}

.scroll-wrapper::-webkit-scrollbar {
	display: none;
}

.scroll-content {
	display: flex;
	flex-wrap: nowrap;
	width: fit-content;
}

.scroll-btn {
	width: 24px;
	height: 24px;
	position: absolute;
	top: 50%;
	transform: translate(0, -50%);
	border-radius: 50%;
	background: rgba(255, 255, 255, 0.9);
	cursor: pointer;
	display: flex;
	align-items: center;
	justify-content: center;
	z-index: 1;
	display: none;
	border: 1px solid $separator;
	.arrow-right {
		transform: rotate(180deg);
	}
}

.scroll-btn.disabled {
	opacity: 0.5;
	cursor: not-allowed;
}

.left-btn {
	left: 0;
	transform: translate(-50%, -50%);
}

.right-btn {
	right: 0;
	transform: translate(50%, -50%);
}

.scroll-btn:hover:not(.disabled) {
	background: rgba(255, 255, 255, 1);
	box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}
</style>
