<template>
	<div
		v-if="theme"
		:style="{ fontSize: `${fontSize}px` }"
		class="container-wrapper"
	>
		<div class="preview-container column items-center" :style="backgroundStyle">
			<div
				v-if="props.theme.style === THEME_TYPE.IMAGE"
				class="overlay"
				:style="backgroundOverlay"
			/>

			<div
				class="block-mask"
				:style="{
					borderRadius: '2em',
					background: theme.block.background
				}"
			/>

			<div
				class="link-mask"
				:style="{
					borderRadius: '2em',
					background: theme.link.background
				}"
			/>
		</div>
	</div>
</template>
<script lang="ts" setup>
import { computed, PropType } from 'vue';
import {
	THEME_TYPE,
	getGradientColor,
	IMAGE_FILTER,
	AppearanceTheme
} from '@apps/profile/src/types/User';

const props = defineProps({
	theme: {
		type: Object as PropType<AppearanceTheme>,
		required: true
	},
	fontSize: {
		type: Number,
		default: 5.12
	}
});

const backgroundStyle = computed(() => {
	switch (props.theme.style) {
		case THEME_TYPE.SOLID:
			return { 'background-color': props.theme.background };
		case THEME_TYPE.GRADIENT:
			return {
				background: getGradientColor(
					props.theme.gradientTopColor,
					props.theme.gradientBottomColor
				)
			};
		case THEME_TYPE.IMAGE:
			if (props.theme.useUpload) {
				return {
					background: `url("${props.theme.uploadImg}")`,
					'background-size': 'cover'
				};
			} else {
				return {
					'background-image': `url("/profile/mobile/${props.theme.localImg}")`,
					'background-size': 'cover'
				};
			}
		default:
			return {};
	}
});

const backgroundOverlay = computed(() => {
	switch (props.theme.filter) {
		case IMAGE_FILTER.DARK:
			return {
				background: 'rgba(0, 0, 0, 0.3)'
			};
		case IMAGE_FILTER.Light:
			return {
				background: 'rgba(255, 255, 255, 0.30)'
			};
		default:
			return '';
	}
});
</script>
<style scoped lang="scss">
.block-mask {
	margin-top: 16.75em;
	width: 17.32em;
	height: 8.15em;
}

.link-mask {
	margin-top: 2em;
	width: 17.32em;
	height: 3.05em;
}
</style>
