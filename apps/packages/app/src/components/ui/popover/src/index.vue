<template>
	<div
		class="relative-position popup-inner"
		ref="popupContainerRef"
		@click.stop
	>
		<div @click="referenceClickHandler">
			<slot name="reference"></slot>
		</div>
		<div
			ref="popupRef"
			class="relative-position bt-popup-container"
			:style="anchorsStyle"
		>
			<div class="content absolute" v-if="visible" :style="styleObj">
				<ScrollArea
					:style="{ height: scrollAreaHeight, width: scrollAreaWidth }"
				>
					<div ref="contentRef">
						<slot></slot>
					</div>
				</ScrollArea>
			</div>
		</div>
	</div>
</template>

<script setup lang="ts">
import {
	ref,
	onMounted,
	defineProps,
	computed,
	withDefaults,
	watch
} from 'vue';
import { verticalObj, horizontalObj, bus } from './popover';
import ScrollArea from '../../ScrollArea/src/index.vue';

const emits = defineEmits(['update:modelValue']);
const visible = ref(false);

interface Props {
	/**
	 * Allows the menu to match at least the full width of its target
	 */
	fit?: boolean;
	/**
	 * Allows the menu to cover its target. When used, the 'self' and 'fit' props are no longer effective
	 */
	cover?: boolean;
	/**
	 * The maximum height of the menu; Size in CSS units, including unit name
	 */
	maxHeight?: string | null | undefined;
	/**
	 * The maximum width of the menu; Size in CSS units, including unit name
	 */
	maxWidth?: string | null | undefined;
	anchor?:
		| 'top left'
		| 'top middle'
		| 'top right'
		| 'top start'
		| 'top end'
		| 'center left'
		| 'center middle'
		| 'center right'
		| 'center start'
		| 'center end'
		| 'bottom left'
		| 'bottom middle'
		| 'bottom right'
		| 'bottom start'
		| 'bottom end'
		| undefined;
	/**
	 * Two values setting the menu's own position relative to its target
	 */
	self?:
		| 'top left'
		| 'top middle'
		| 'top right'
		| 'top start'
		| 'top end'
		| 'center left'
		| 'center middle'
		| 'center right'
		| 'center start'
		| 'center end'
		| 'bottom left'
		| 'bottom middle'
		| 'bottom right'
		| 'bottom start'
		| 'bottom end'
		| undefined;
	/**
	 * An array of two numbers to offset the menu horizontally and vertically in pixels
	 */
	offset?: readonly any[] | undefined;
	/**
	 * Model of the component defining shown/hidden state; Either use this property (along with a listener for 'update:model-value' event) OR use v-model directive
	 */
	modelValue?: boolean;
}
const props = withDefaults(defineProps<Props>(), {
	modelValue: undefined,
	maxHeight: '999999px',
	maxWidth: '999999px',
	offset: [0, 0] as any
});

const popupContainerRef = ref();
const popupRef = ref();
const contentRef = ref();

const contentRect = computed(() => {
	if (contentRef.value) {
		return contentRef.value.getBoundingClientRect();
	}
	return { width: 0, height: 0 };
});

const scrollAreaHeight = computed(
	() => `min(${contentRect.value.height}px, ${props.maxHeight})`
);

const scrollAreaWidth = computed(() => {
	const widthTemp =
		props.fit || props.cover ? '100%' : `${contentRect.value.width}px`;
	return `min(${widthTemp}, ${props.maxWidth})`;
});

const offsetPX = computed(() => props.offset.map((item) => `${item}px`));

const anchorsStyle = computed(() => {
	if (props.cover) {
		return {
			top: offsetYFormat('0px'),
			width: '100%',
			left: offsetXFormat('0px')
		};
	} else if (props.fit) {
		return {
			top: offsetYFormat('100%'),
			width: '100%',
			left: offsetXFormat('0px')
		};
	} else if (props.anchor) {
		const anchors = props.anchor.split(' ');
		return {
			top: offsetYFormat(verticalObj[anchors[0]]),
			left: offsetXFormat(horizontalObj[anchors[1]])
		};
	}
	return {
		top: offsetYFormat('100%'),
		left: offsetXFormat('0px')
	};
});

const styleObj = computed(() => {
	if (props.fit || props.cover) {
		return { width: '100%' };
	} else if (props.self) {
		const selfs = props.self.split(' ');
		return {
			transform: `translate(-${horizontalObj[selfs[1]]}, -${
				verticalObj[selfs[0]]
			})`
		};
	}
	return {};
});

function isBoolean(value: any) {
	return typeof value === 'boolean';
}

function offsetXFormat(value: string) {
	return `calc(${offsetPX.value[0]} + ${value})`;
}

function offsetYFormat(value: string) {
	return `calc(${offsetPX.value[1]} + ${value})`;
}

function referenceClickHandler() {
	if (!isBoolean(props.modelValue)) {
		visible.value = !visible.value;
	}
}

function close() {
	if (!isBoolean(props.modelValue)) {
		visible.value = false;
	}
}

onMounted(() => {
	bus.emit('bt-popover-close', close);
	document.body.addEventListener(
		'click',
		(el) => {
			if (
				!isBoolean(props.modelValue) &&
				!popupContainerRef.value?.contains(el.target)
			) {
				close();
			}
		},
		true
	);
});
watch(
	() => props.modelValue,
	(newValue) => {
		if (isBoolean(newValue) && newValue != visible.value) {
			visible.value = newValue;
		}
	},
	{
		immediate: true
	}
);
</script>

<script lang="ts">
import { defineComponent } from 'vue';
export default defineComponent({
	name: 'BtPopover'
});
</script>

<style lang="scss" scoped>
.popup-inner {
	display: inline-block;
}
.bt-popup-container {
	position: absolute;
	top: 100%;
	z-index: 9999;
	width: 100%;
	.content {
		border-radius: 4px;
		background: $background-2;
		box-shadow: 0px 4px 10px 0px rgba(0, 0, 0, 0.2);
		top: 0px;
		overflow: hidden;
	}
}
</style>
