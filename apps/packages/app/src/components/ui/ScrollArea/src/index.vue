<template>
	<q-scroll-area
		ref="scrollRef"
		v-bind="$attrs"
		:thumb-style="
			$attrs['thumb-style'] ? $attrs['thumb-style'] : scrollBarStyle.thumbStyle
		"
		@scroll="scroll"
	>
		<slot />
	</q-scroll-area>
</template>

<script setup lang="ts">
import { ref } from 'vue';

const scrollRef = ref();

const scrollBarStyle = ref({
	contentStyle: {},
	contentActiveStyle: {},
	thumbStyle: {
		right: '2px',
		borderRadius: '3px',
		backgroundColor: '#BCBDBE',
		width: '6px',
		height: '6px',
		opacity: '1'
	}
});

const emits = defineEmits(['scroll']);

const scroll = (info: any) => {
	emits('scroll', info);
};

const getScrollTarget = () => {
	return scrollRef.value.getScrollTarget();
};

const getScroll = () => {
	return scrollRef.value.getScroll();
};

const getScrollPosition = () => {
	return scrollRef.value.getScrollPosition();
};

const getScrollPercentage = (info: any) => {
	return scrollRef.value.getScrollPercentage();
};
const setScrollPosition = (axis: string, offset: number, duration = 300) => {
	return scrollRef.value.setScrollPosition(axis, offset, duration);
};
const setScrollPercentage = (axis: string, offset: number, duration = 300) => {
	return scrollRef.value.setScrollPercentage(axis, offset, duration);
};

defineExpose({
	getScrollTarget,
	getScroll,
	getScrollPercentage,
	getScrollPosition,
	setScrollPosition,
	setScrollPercentage
});
</script>

<script lang="ts">
import { defineComponent } from 'vue';
export default defineComponent({
	name: 'BtScrollArea'
});
</script>
