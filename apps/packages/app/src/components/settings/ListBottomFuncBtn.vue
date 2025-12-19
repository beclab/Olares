<template>
	<div
		class="func-btn-bg row items-center justify-end clickable-view"
		:style="{
			'--height': height
		}"
	>
		<div class="func-btn" @click="funcOnClick">
			<div
				class="text-body3 title row items-center"
				:class="deviceStore.isMobile ? 'text-body2-m' : 'text-body3'"
			>
				{{ title }}
			</div>
		</div>
	</div>
</template>

<script setup lang="ts">
import { useDeviceStore } from 'src/stores/settings/device';
defineProps({
	title: {
		type: String,
		required: false,
		default: ''
	},
	height: {
		type: Number,
		required: false,
		default: 40
	}
});

const emit = defineEmits(['funcClick']);

const deviceStore = useDeviceStore();

const funcOnClick = () => {
	emit('funcClick');
};
</script>

<style scoped lang="scss">
.func-btn-bg {
	.func-btn {
		height: var(--height);
		border-radius: 8px;
		border: 1px solid $btn-stroke;
		width: auto;
		padding: 8px 12px;

		.title {
			height: 100%;
			text-align: center;
			color: $ink-2;
		}
	}

	.func-btn:hover {
		background-color: $background-hover;
	}
}
</style>
