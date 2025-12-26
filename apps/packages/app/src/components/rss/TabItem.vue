<template>
	<div
		class="row justify-start items-center"
		@click="itemOnClick"
		:style="{
			'--borderBottom': hideBorder ? 'transparent' : orangeDefault
		}"
		:class="
			index !== 0
				? curIndex === index
					? 'tab-item-root-selected'
					: 'tab-item-root'
				: selected
				? 'tab-item-root-selected'
				: 'tab-item-root'
		"
	>
		<q-icon
			style="margin-right: 4px"
			v-if="name"
			:name="name"
			class="font-icon"
			size="16px"
		/>
		<div
			class="text-subtitle1"
			:class="
				index !== 0
					? curIndex === index
						? 'tab-item-title-selected'
						: 'tab-item-title'
					: selected
					? 'tab-item-title-selected'
					: 'tab-item-title'
			"
		>
			{{ title }}
		</div>
	</div>
</template>

<script setup lang="ts">
import '../../css/page.scss';

const prop = defineProps({
	selected: {
		type: Boolean,
		default: false
	},
	curIndex: {
		type: Number,
		default: 0
	},
	index: {
		type: Number,
		default: 0
	},
	title: {
		type: String,
		default: '',
		require: false
	},
	name: {
		type: String,
		default: '',
		require: false
	},
	transparent: {
		type: Boolean,
		default: false
	},
	hideBorder: {
		type: Boolean,
		default: false
	}
});

import { useColor } from '@bytetrade/ui';
const { color: orangeDefault } = useColor('orange-default');

const emit = defineEmits(['OnItemClick']);

const itemOnClick = () => {
	emit('OnItemClick', prop.index);
};
</script>

<style scoped lang="scss">
.tab-item-root {
	height: 56px;
	padding-left: 12px;
	padding-right: 12px;
	width: auto;
	background-color: transparent;
	color: $ink-2;
}

.tab-item-root-selected {
	height: 56px;
	padding-left: 12px;
	padding-right: 12px;
	width: auto;
	background-color: transparent;
	color: $orange-default;
}

.tab-item-title {
	height: 100%;
	text-align: center;
	padding-top: 16px;
	border-bottom: 2px solid transparent;
}

.tab-item-title-selected {
	height: 100%;
	text-align: center;
	padding-top: 16px;
	border-bottom: 2px solid var(--borderBottom);
}
</style>
