<template>
	<div
		class="custom-background column items-center"
		:style="{
			border: data.selected ? `1px solid ${color}` : '1px solid transparent'
		}"
	>
		<Handle type="target" :position="Position.Top" />
		<q-img class="custom-image-status" :src="src" />
		<div class="custom-text-label text-overline text-ink-1">{{ label }}</div>
		<Handle type="source" :position="Position.Bottom" />
	</div>
</template>

<script setup lang="ts">
import type { NodeProps } from '@vue-flow/core';
import { NODE_PHASE } from 'src/utils/rss-types';
import { CustomData, CustomEvents } from 'src/utils/nodeUtil';
import { getRequireImage } from 'src/utils/rss-utils';
import { computed } from 'vue';
import { Position } from '@vue-flow/core';
import { useColor } from '@bytetrade/ui';

// props were passed from the slot using `v-bind="customNodeProps"`
const props = defineProps<NodeProps<CustomData, CustomEvents>>();

const src = computed(() => {
	switch (props.data.phase) {
		case NODE_PHASE.RUNNING:
			return getRequireImage('workflow/loading.svg');
		case NODE_PHASE.PENDING:
			return getRequireImage('workflow/waiting.svg');
		case NODE_PHASE.SUCCEEDED:
			return getRequireImage('workflow/success.svg');
		case NODE_PHASE.ERROR:
		case NODE_PHASE.FAILED:
			return getRequireImage('workflow/error.svg');
		default:
			return getRequireImage('workflow/unknown.svg');
	}
});

const { color } = useColor('orange-default');
</script>

<style lang="scss">
.custom-background {
	width: 120px;
	max-height: 82px;
	padding: 12px;
	background-color: $background-1;
	border-radius: 12px;

	.custom-image-status {
		width: 24px;
		height: 24px;
	}

	.custom-text-label {
		margin-top: 8px;
		text-align: center;
		display: -webkit-box;
		-webkit-line-clamp: 2;
		-webkit-box-orient: vertical;
		overflow: hidden;
	}
}
</style>
