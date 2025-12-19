<template>
	<div
		class="row items-center justify-center item"
		:style="{
			width: `${size}px`,
			height: `${size}px`
		}"
		:class="`bg-${colors.bg}`"
	>
		<q-icon name="sym_r_person" :color="colors.color" :size="`${innerSize}px`">
			<q-tooltip>
				{{ name }}
			</q-tooltip>
		</q-icon>
	</div>
</template>

<script setup lang="ts">
import { stringToIntHash } from 'src/utils/format';
import { computed } from 'vue';

const props = defineProps({
	iconColor: {
		type: String,
		required: false,
		default: ''
	},
	bgColor: {
		type: String,
		required: false,
		default: ''
	},
	size: {
		type: Number,
		required: false,
		default: 32
	},
	innerSize: {
		type: Number,
		required: false,
		default: 24
	},
	name: {
		type: String,
		required: false,
		default: ''
	}
});

const id = props.name.length > 0 ? stringToIntHash(props.name, 0, 6) : 0;

const colorsArray = [
	{
		bg: 'blue-soft',
		color: 'blue-default'
	},
	{
		bg: 'orange-soft',
		color: 'orange-default'
	},
	{
		bg: 'green-soft',
		color: 'green-default'
	},
	{
		bg: 'yellow-soft',
		color: 'yellow-default'
	},
	{
		bg: 'red-soft',
		color: 'red-default'
	},
	{
		bg: 'light-blue-soft',
		color: 'light-blue-default'
	},
	{
		bg: 'teal-soft',
		color: 'teal-default'
	}
];

const colors = computed(() => {
	const co = id >= colorsArray.length ? colorsArray[0] : colorsArray[id];
	return {
		bg: !!props.bgColor ? props.bgColor : co.bg,
		color: !!props.iconColor ? props.iconColor : co.color
	};
});
</script>

<style scoped lang="scss">
.item {
	border-radius: 50%;
	overflow: hidden;
	cursor: pointer;
}
</style>
