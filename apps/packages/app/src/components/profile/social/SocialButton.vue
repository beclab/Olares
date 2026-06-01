<template>
	<transition name="fade">
		<div
			v-if="display"
			:class="
				selected
					? 'social-btn-background-selected'
					: 'social-btn-background-normal'
			"
			class="row justify-center items-center"
			@click="onButtonClick"
		>
			<q-icon
				size="24px"
				:name="
					selected
						? `img:/profile/social/selected/${platform}.svg`
						: `img:/profile/social/normal/${platform}.svg`
				"
			/>
			<slot />
		</div>
	</transition>
</template>

<script lang="ts" setup>
import { computed } from 'vue';
import { SocialMap } from '@apps/profile/src/types/SocialProps';
import { useUserStore } from '@apps/profile/src/stores/profileUser';
const userStore = useUserStore();

const props = defineProps({
	platform: {
		type: String,
		required: true
	},
	display: {
		type: Boolean,
		required: true
	}
});

const selected = computed(() => {
	const data = userStore.user?.social?.data;
	if (!userStore.user || !Array.isArray(data)) {
		return false;
	}
	return data.some((s) => s && s.platform === props.platform);
});

const onButtonClick = () => {
	const user = userStore.user;
	if (!user?.social?.data) {
		return;
	}
	const data = user.social.data;
	if (data.some((s) => s && s.platform === props.platform)) {
		user.social.data = data.filter(
			(item) => item && item.platform !== props.platform
		);
		return;
	}
	const config = SocialMap[props.platform];
	if (!config) {
		return;
	}
	user.social.data.push({ ...config });
};
</script>

<style scoped lang="scss">
.social-btn-background-normal {
	width: 40px;
	height: 40px;
	border-radius: 12px;
	border: 0.5px solid $separator-2;
	background: $ink-on-brand;
	cursor: pointer;
	text-decoration: none;
}

.social-btn-background-selected {
	width: 40px;
	height: 40px;
	border-radius: 12px;
	border: 0.5px solid $separator-2;
	background: $background-3;
	cursor: pointer;
	text-decoration: none;
}
</style>
