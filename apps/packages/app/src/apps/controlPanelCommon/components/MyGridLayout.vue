<template>
	<div class="my-grid-container" ref="gridContainerRef">
		<slot></slot>
		<div v-for="item in fillCount" :key="item"></div>
	</div>
</template>

<script setup lang="ts">
import {
	computed,
	toRefs,
	ref,
	onBeforeUpdate,
	nextTick,
	onMounted
} from 'vue';
import { getCssVar } from 'quasar';
import { isNil } from 'lodash';

const size = {
	xs: '4px',
	sm: '8px',
	md: '12px',
	lg: '20px',
	xl: '32px',
	none: '0px'
};

interface Props {
	colWidth: string;
	gap?: 'xs' | 'sm' | 'md' | 'lg' | 'xl' | 'none';
	gapRow?: 'xs' | 'sm' | 'md' | 'lg' | 'xl' | 'none';
	gapCol?: 'xs' | 'sm' | 'md' | 'lg' | 'xl' | 'none';
	contentLength?: number;
}
const props = withDefaults(defineProps<Props>(), {
	colWidth: '160px',
	gap: 'md'
});

const { colWidth, gap, contentLength } = toRefs(props);
const gridContainerRef = ref();
const gapRowConputed = computed(() => {
	return props.gapRow ? size[props.gapRow] : size[props.gap ?? 'md'];
});

const gapColConputed = computed(() => {
	return props.gapCol ? size[props.gapCol] : size[props.gap ?? 'md'];
});
const fillCount = ref(0);

onMounted(() => {
	updateLayout();
});
onBeforeUpdate(() => {
	updateLayout();
});

const updateLayout = () => {
	if (gridContainerRef.value && !isNil(contentLength.value)) {
		const containerWidth = gridContainerRef.value?.clientWidth ?? 0;
		const itemLength = contentLength.value ?? 0;
		const itemWidth = Number(colWidth.value.replace('px', ''));
		const gapWidth = Number(gapColConputed.value.replace('px', ''));
		const rowLength = Math.floor(
			(containerWidth - gapWidth) / (itemWidth + gapWidth)
		);
		fillCount.value = Math.max(0, rowLength - itemLength);
	}
};
</script>

<style lang="scss" scoped>
.my-grid-container {
	display: grid;
	grid-template-columns: repeat(auto-fit, minmax(v-bind(colWidth), 1fr));
	grid-row-gap: v-bind(gapRowConputed);
	grid-column-gap: v-bind(gapColConputed);
}
</style>
