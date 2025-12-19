<template>
	<div
		v-if="header"
		:style="{ fontSize: `${fontSize}px`, fontFamily: FONT_TYPE.ROBOTO }"
		class="container-wrapper"
	>
		<div
			class="preview-container column items-center"
			:class="
				header.style === HEADER_STYLE_TYPE.PORTRAIT
					? ''
					: 'preview-container-padding'
			"
			:style="{ background: '#FFFFFF' }"
		>
			<header-banner :header="header" />
			<header-preview
				:header="header"
				:social="social"
				:social-clickable="false"
				text-color="#1F1F1F"
			/>
			<div
				class="classic-block-mask"
				v-if="header.style === HEADER_STYLE_TYPE.CLASSIC"
			/>
			<div
				class="classic-block-mask"
				v-if="header.style === HEADER_STYLE_TYPE.CLASSIC"
			/>
			<div
				class="portrait-block-mask"
				v-if="header.style === HEADER_STYLE_TYPE.PORTRAIT"
			/>
			<div
				class="banner-block-mask"
				v-if="header.style === HEADER_STYLE_TYPE.BANNER"
			/>
			<div
				class="banner-block-mask"
				v-if="header.style === HEADER_STYLE_TYPE.BANNER"
			/>
		</div>
	</div>
</template>
<script lang="ts" setup>
import HeaderPreview from '@apps/profile/src/components/profile/layout/HeaderPreview.vue';
import HeaderBanner from '@apps/profile/src/components/profile/layout/HeaderBanner.vue';
import { PropType } from 'vue';
import {
	HEADER_STYLE_TYPE,
	UserHeader,
	FONT_TYPE,
	UserSocial
} from '@apps/profile/src/types/User';

defineProps({
	header: {
		type: Object as PropType<UserHeader>,
		required: true
	},
	social: {
		type: Object as PropType<UserSocial>,
		required: true
	},
	fontSize: {
		type: Number,
		default: 5.12
	}
});
</script>
<style scoped lang="scss">
.classic-block-mask {
	margin-top: 1.875em;
	width: 18.75em;
	height: 8.2em;
	background: #f6f6f6;
	border-radius: 1.56em;
}
.portrait-block-mask {
	margin-top: 1.875em;
	width: 18.75em;
	height: 7.97em;
	background: #f6f6f6;
	border-radius: 1.56em;
}
.banner-block-mask {
	margin-top: 1.875em;
	width: 18.75em;
	height: 6.56em;
	background: #f6f6f6;
	border-radius: 1.56em;
}
</style>
