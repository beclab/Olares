<template>
	<q-item clickable class="bio-share-item justify-between items-center">
		<div class="bio-share-div row justify-start items-center">
			<social-svg
				v-if="platform"
				:platform="platform"
				width="20px"
				height="20px"
			/>
			<q-icon color="ink-1" v-if="icon" :name="icon" size="20px" />
			<div class="bio-share-text">{{ label }}</div>
		</div>
		<q-icon color="grey-5" class="20px" name="sym_r_chevron_right" />
	</q-item>
</template>

<script lang="ts" setup>
import SocialSvg from '@apps/profile/src/components/profile/social/SocialSvg.vue';

defineProps({
	icon: String,
	platform: String,
	label: String,
	close: {
		type: Boolean,
		default: true
	}
});
</script>

<style scoped lang="scss">
.bio-share-item {
	width: 100%;
	height: 44px;
	min-height: 44px;
	margin: 0;
	padding: 8px;

	&:hover {
		background-color: $background-hover;
		border-radius: 8px;
	}

	.bio-share-div {
		width: calc(100% - 20px);
		height: 100%;
		color: $ink-1;

		.bio-share-text {
			font-family: Roboto;
			font-size: 12px;
			font-weight: 500;
			line-height: 16px;
			letter-spacing: 0em;
			text-align: left;
			margin-left: 8px;
		}
	}
}
</style>
