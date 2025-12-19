<template>
	<div
		class="row justify-center items-center"
		:class="selected ? 'tab-item-root-selected' : 'tab-item-root'"
	>
		<q-icon v-if="icon" :name="icon" class="font-icon q-mr-xs" size="16px" />
		<div class="tab-item-title">{{ label }}</div>
	</div>
</template>

<script setup lang="ts">
import { defineProps } from 'vue';

defineProps({
	selected: {
		type: Boolean,
		default: false
	},
	label: {
		type: String,
		default: '',
		require: false
	},
	icon: {
		type: String,
		default: '',
		require: false
	}
});
</script>

<style scoped lang="scss">
.tab-item-root {
	height: 20px;
	width: 112.67px;
	color: $ink-3;
	text-align: center;
	border-radius: 4px;
	background-color: $background-3;
}

.tab-item-root-selected {
	height: 20px;
	width: 112.67px;
	color: $ink-1;
	border-radius: 4px;
	text-align: center;
	background-color: $background-2;
	border: 1px solid $separator-2;
}

.tab-item-title {
	font-family: Roboto;
	font-size: 12px;
	font-weight: 400;
	line-height: 16px;
	letter-spacing: 0em;
	text-align: center;
	text-transform: capitalize;
}
</style>
