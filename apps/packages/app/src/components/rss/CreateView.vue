<template>
	<span
		v-if="name"
		class="tag-view row inline justify-start"
		:style="{
			'--Background': border ? borderBackground : selected ? hover : color,
			'--Border': border ? separator : 'transparent',
			'--MaxWidth': maxWidth
		}"
	>
		<span class="tag-text text-caption text-orange-default">
			{{ name }}
		</span>

		<q-icon
			v-if="edit"
			class="tag-close cursor-pointer"
			size="12px"
			name="sym_r_close_small"
			color="grey-5"
			@click="emit('onRemoveClick')"
		/>
	</span>
</template>

<script lang="ts" setup>
import { useColor } from '@bytetrade/ui';

defineProps({
	name: {
		type: String,
		require: true
	},
	edit: {
		type: Boolean,
		default: false
	},
	selected: {
		type: Boolean,
		default: false,
		require: false
	},
	border: {
		type: Boolean,
		default: false,
		require: false
	},
	maxWidth: {
		type: String,
		default: '170px',
		require: false
	}
});

const { color: hover } = useColor('background-2');
const { color } = useColor('background-hover');
const { color: borderBackground } = useColor('background-6');
const { color: separator } = useColor('separator');

const emit = defineEmits(['onRemoveClick']);
</script>

<style lang="scss">
.tag-view {
	height: 20px;
	padding: 4px 8px;
	border-radius: 20px;
	background: var(--Background);
	border: 1px solid var(--Border);

	.tag-text {
		max-width: var(--MaxWidth);
		overflow: hidden !important;
		text-overflow: ellipsis !important;
		text-align: left;
		display: -webkit-box;
		-webkit-line-clamp: 1;
		-webkit-box-orient: vertical;
	}

	.tag-close {
		margin-left: 8px;
		width: 12px;
		height: 12px;
	}
}
</style>
