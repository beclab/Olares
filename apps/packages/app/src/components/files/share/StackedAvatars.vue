<template>
	<div class="stacked-avatars row items-center justify-between">
		<div
			v-for="index in visibleCount"
			:key="index"
			class="avatar"
			:style="{
				zIndex: index + 1,
				width: `${itemSize}px`,
				height: `${itemSize}px`,
				right: `${
					(visibleCount - index - (remainingCount > 0 ? -1 : 0)) *
					(itemSize * 0.55)
				}px`
			}"
		>
			<slot name="content" :data="index" />
		</div>
		<div
			v-if="remainingCount > 0"
			class="avatar bg-yellow-soft text-ink-1 text-overline"
			:style="{
				width: `${itemSize}px`,
				height: `${itemSize}px`,
				zIndex: visibleCount + 1,
				right: '0px'
			}"
		>
			+{{ remainingCount }}
		</div>
	</div>
</template>

<script setup lang="ts">
import { computed } from 'vue';

const props = defineProps({
	avatarsLength: {
		type: Number,
		required: true,
		default: 0
	},
	maxVisible: {
		type: Number,
		default: 3
	},
	itemSize: {
		type: Number,
		default: 20
	}
});

const visibleCount = computed(() => {
	return props.avatarsLength >= props.maxVisible
		? props.maxVisible
		: props.avatarsLength;
});

const remainingCount = computed(() => {
	return Math.max(0, props.avatarsLength - props.maxVisible);
});
</script>

<style scoped lang="scss">
.stacked-avatars {
	position: relative;
	padding-right: 20px;
	width: calc(100% - 30px);
}

.avatar {
	position: absolute;
	border-radius: 50%;
	border: 1px solid $background-1;
	overflow: hidden;
	display: flex;
	align-items: center;
	justify-content: center;
	box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
	transition: transform 0.3s ease;
}

.avatar img {
	width: 100%;
	height: 100%;
	object-fit: cover;
}
</style>
