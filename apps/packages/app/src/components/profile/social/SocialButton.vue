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
	if (!userStore.user || userStore.user?.social.data.length < 0) {
		return false;
	}
	const index = userStore.user?.social.data.findIndex(
		(social) => social.platform === props.platform
	);
	return index >= 0;
});

const onButtonClick = () => {
	if (userStore.user) {
		if (selected.value) {
			userStore.user.social.data = userStore.user.social.data.filter(
				(item) => item.platform !== props.platform
			);
		} else {
			userStore.user.social.data.push(SocialMap[props.platform]);
		}
	}
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
