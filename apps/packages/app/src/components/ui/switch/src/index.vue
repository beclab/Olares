<template>
	<span :style="style">
		<q-toggle v-bind="$attrs" :size="width" />
	</span>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import useColor from '../../utils/useColor';

const props = withDefaults(
	defineProps<{
		/**
		 * size of the Toggle
		 */
		size?: 'md' | 'sm' | 'xs';
		/**
		 * backgroundColor opacity of the Toggle
		 */
		opacity: number;
		/**
		 * track color of the Toggle when truthy
		 */
		truthyTrackColor?: string;
		/**
		 * track color of the Toggle when falsy
		 */
		falsyTrackColor?: string;
	}>(),
	{
		opacity: 1,
		size: 'sm',
		truthyTrackColor: 'orange-default',
		falsyTrackColor: 'background-5'
	}
);

const width = computed(() => {
	if (props.size === 'xs') {
		return (12 + 4) * 2 + 'px';
	} else if (props.size === 'md') {
		return (20 + 4) * 2 + 'px';
	} else {
		return (16 + 4) * 2 + 'px';
	}
});

const style = computed(() => {
	let sizeStyle;
	if (props.size === 'xs') {
		sizeStyle = {
			'--toggle-thumb-size': 12 + 'px',
			'--toggle-height': 12 + 2 * 2 + 'px'
		};
	} else if (props.size === 'md') {
		sizeStyle = {
			'--toggle-thumb-size': 20 + 'px',
			'--toggle-height': 20 + 2 * 2 + 'px'
		};
	} else {
		sizeStyle = {
			'--toggle-thumb-size': 16 + 'px',
			'--toggle-height': 16 + 2 * 2 + 'px'
		};
	}

	return {
		...sizeStyle,
		'--toggle-margin-y': 2 + 'px',
		'--toggle-margin-l': 4 + 'px',
		'--toggle-margin-r': '-' + 4 + 'px',
		'--toggle-radius': 999 + 'px',
		'--toggle-background-opacity': props.opacity,
		'--toggle-truthy-track-color': useColor(props.truthyTrackColor).color.value,
		'--toggle-falsy-track-color': useColor(props.falsyTrackColor).color.value
	};
});
</script>

<script lang="ts">
import { defineComponent } from 'vue';
export default defineComponent({
	name: 'BtSwitch'
});
</script>
<style lang="scss" scoped>
::v-deep(.q-toggle__inner) {
	height: var(--toggle-height);
	line-height: var(--toggle-height);
	padding-top: 0;
	padding-bottom: 0;
}
::v-deep(.q-toggle__inner--falsy .q-toggle__track) {
	height: var(--toggle-height);
	border-radius: var(--toggle-radius);
	background: var(--toggle-falsy-track-color);
	opacity: var(--toggle-background-opacity) !important;
}
::v-deep(.q-toggle__inner--truthy .q-toggle__track) {
	height: var(--toggle-height);
	border-radius: var(--toggle-radius);
	background: var(--toggle-truthy-track-color);
	opacity: var(--toggle-background-opacity) !important;
}
::v-deep(.q-toggle__thumb) {
	height: var(--toggle-thumb-size) !important;
	width: var(--toggle-thumb-size);
	top: var(--toggle-margin-y);
}
::v-deep(.q-toggle__thumb:after) {
	box-shadow: none;
	background: $ink-on-brand;
	filter: drop-shadow(0px 2px 4px rgba(0, 0, 0, 0.1));
	left: var(--toggle-margin-l);
	right: var(--toggle-margin-r);
}

::v-deep(.q-toggle__inner--truthy .q-toggle__thumb:after) {
	right: 0;
	left: 0;
}
::v-deep(.q-toggle__thumb:before) {
	display: none;
}
</style>
