<template>
	<div class="operate-btn q-px-sm row items-center justify-center">
		<q-icon :name="icon" size="16px" v-if="icon" />
		<span class="operate-btn-install">{{ label }}</span>
	</div>
</template>

<script setup lang="ts">
import { defineProps, withDefaults } from 'vue';
interface Props {
	icon: string;
	label: string;
}

withDefaults(defineProps<Props>(), {
	label: ''
});
</script>

<style scoped lang="scss">
.operate-btn {
	height: 32px;
	text-align: center;
	border-radius: 8px;
	border: 1px solid $btn-stroke;
	overflow: hidden;
	color: $ink-1;

	&:hover {
		background: $background-hover;
	}
	&.operate-disabled {
		opacity: 0.5;
	}

	.operate-btn-install {
		line-height: 32px;
		padding: 0 8px;
		color: $ink-1;
		font-size: 12px;
		line-height: 100%;
		text-align: center;
		cursor: pointer;

		.ani-loading {
			animation: rotate 1s linear infinite;
		}
	}
}
</style>
