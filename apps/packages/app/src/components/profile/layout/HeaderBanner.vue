<template>
	<q-responsive
		v-if="header.style === HEADER_STYLE_TYPE.BANNER"
		class="avatar-banner"
		:style="{
			marginTop: previewPc ? '7.5em' : '0',
			borderTopRightRadius: previewPc ? '1.2em' : '0',
			borderTopLeftRadius: previewPc ? '1.2em' : '0'
		}"
		:ratio="2"
	>
		<div class="full-width" :style="bannerStyle" />
	</q-responsive>
</template>

<script lang="ts" setup>
import { computed, PropType } from 'vue';
import { HEADER_STYLE_TYPE, UserHeader } from '@apps/profile/src/types/User';
import { useUserStore } from '@apps/profile/src/stores/profileUser';

const props = defineProps({
	header: {
		type: Object as PropType<UserHeader>,
		required: true
	}
});

const userStore = useUserStore();

const bannerStyle = computed(() => {
	if (props.header.banner) {
		return {
			background: `url("${props.header.banner}")`,
			'background-size': 'cover'
		};
	}
	return {
		'background-image': `url("/profile/banner/banner_default.jpg")`,
		'background-size': 'cover'
	};
});

const previewPc = computed(() => {
	return !userStore.isMobile && process.env.ACTION === 'PREVIEW';
});
</script>

<style scoped lang="scss">
.avatar-banner {
	top: 0;
	position: absolute;
	max-width: 37.5em;
	width: 100%;
	overflow: hidden;
}
</style>
