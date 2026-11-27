<template>
	<div
		ref="buttonRef"
		class="add-file"
		:style="buttonStyle"
		@pointerdown="onPointerDown"
		@click="addFile"
	>
		<img src="./../../../assets/add_file.svg" alt="add_file" />
	</div>
</template>

<script lang="ts" setup>
import { computed, nextTick, onBeforeUnmount, onMounted, ref } from 'vue';

const emits = defineEmits(['addFile']);

const BUTTON_SIZE = 44;
const DEFAULT_RIGHT = 32;
const DEFAULT_BOTTOM = 110;
const EDGE_PADDING = 12;
const DRAG_THRESHOLD = 6;

const buttonRef = ref<HTMLElement>();
const position = ref({ x: 0, y: 0 });
const startPointer = ref({ x: 0, y: 0 });
const startPosition = ref({ x: 0, y: 0 });
const isDragging = ref(false);
const moved = ref(false);

const buttonStyle = computed(() => ({
	left: `${position.value.x}px`,
	top: `${position.value.y}px`
}));

const clampPosition = (x: number, y: number) => {
	const maxX = window.innerWidth - BUTTON_SIZE - EDGE_PADDING;
	const maxY = window.innerHeight - BUTTON_SIZE - EDGE_PADDING;
	return {
		x: Math.min(Math.max(x, EDGE_PADDING), maxX),
		y: Math.min(Math.max(y, EDGE_PADDING), maxY)
	};
};

const setDefaultPosition = () => {
	position.value = clampPosition(
		window.innerWidth - BUTTON_SIZE - DEFAULT_RIGHT,
		window.innerHeight - BUTTON_SIZE - DEFAULT_BOTTOM
	);
};

const addFile = () => {
	if (moved.value) {
		moved.value = false;
		return;
	}
	emits('addFile');
};

const onPointerDown = (event: PointerEvent) => {
	isDragging.value = true;
	moved.value = false;
	startPointer.value = {
		x: event.clientX,
		y: event.clientY
	};
	startPosition.value = {
		...position.value
	};
	buttonRef.value?.setPointerCapture(event.pointerId);
};

const onPointerMove = (event: PointerEvent) => {
	if (!isDragging.value) {
		return;
	}

	const deltaX = event.clientX - startPointer.value.x;
	const deltaY = event.clientY - startPointer.value.y;
	if (Math.abs(deltaX) > DRAG_THRESHOLD || Math.abs(deltaY) > DRAG_THRESHOLD) {
		moved.value = true;
	}

	position.value = clampPosition(
		startPosition.value.x + deltaX,
		startPosition.value.y + deltaY
	);
};

const onPointerUp = () => {
	isDragging.value = false;
};

const onResize = () => {
	position.value = clampPosition(position.value.x, position.value.y);
};

onMounted(() => {
	nextTick(setDefaultPosition);
	window.addEventListener('pointermove', onPointerMove);
	window.addEventListener('pointerup', onPointerUp);
	window.addEventListener('pointercancel', onPointerUp);
	window.addEventListener('resize', onResize);
});

onBeforeUnmount(() => {
	window.removeEventListener('pointermove', onPointerMove);
	window.removeEventListener('pointerup', onPointerUp);
	window.removeEventListener('pointercancel', onPointerUp);
	window.removeEventListener('resize', onResize);
});
</script>

<style lang="scss" scoped>
.add-file {
	position: fixed;
	width: 44px;
	height: 44px;
	background-color: #ffeb3b;
	border-radius: 22px;
	display: flex;
	align-items: center;
	justify-content: center;
	z-index: 10;
	touch-action: none;
	user-select: none;
	// box-shadow: 0px 0px 24px 0 rgba($color: #51aeff, $alpha: 0.8);
	img {
		width: 16px;
		height: 16px;
	}
}
</style>
