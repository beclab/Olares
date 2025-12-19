<template>
	<div
		:style="backgroundStyle"
		class="bio-profile-picture row justify-center items-center"
	>
		<div class="bio-profile-card column items-center">
			<!-- <q-avatar>
        <q-img
          class="bio-profile-avatar"
          alt="name"
          :src="userData.avatarUrl"
        />
      </q-avatar> -->
			<div class="bio-profile-background column justify-center items-center">
				<q-avatar size="48px">
					<TerminusAvatar :info="userStore.info" :size="48" />
				</q-avatar>
				<div class="bio-profile-username">
					{{ name }}
				</div>
				<div class="bio-profile-url">
					{{ url }}
				</div>
			</div>
		</div>
	</div>
</template>

<script lang="ts" setup>
import { computed } from 'vue';
import { useUserStore } from '@apps/profile/src/stores/profileUser';
import { getGradientColor, THEME_TYPE } from '@apps/profile/src/types/User';
const userStore = useUserStore();

defineProps({
	name: String,
	url: String
});

const backgroundStyle = computed(() => {
	if (!userStore.user) {
		return {};
	}
	let backgroundStyle = {};
	switch (userStore.user.appearance.theme.style) {
		case THEME_TYPE.SOLID:
			backgroundStyle = {
				backgroundColor: userStore.user.appearance.theme.background
			};
			break;
		case THEME_TYPE.GRADIENT:
			backgroundStyle = {
				background: getGradientColor(
					userStore.user.appearance.theme.gradientTopColor,
					userStore.user.appearance.theme.gradientBottomColor
				)
			};
			break;
		case THEME_TYPE.IMAGE:
			if (userStore.user.appearance.theme.useUpload) {
				backgroundStyle = {
					background: `url("${userStore.user.appearance.theme.uploadImg}")`,
					backgroundSize: 'cover'
				};
			} else {
				backgroundStyle = {
					backgroundImage: `url("/profile/mobile/${userStore.user.appearance.theme.localImg}")`,
					backgroundSize: 'cover'
				};
			}
			break;
	}
	return backgroundStyle;
});
</script>

<style scoped lang="scss">
.bio-profile-picture {
	width: 236px;
	position: relative;
	height: 360px;
	border-radius: 12px;
	backdrop-filter: blur(10px);

	.bio-profile-card {
		width: 236px;
		padding: 20px;
		border-radius: 8px;

		.bio-profile-background {
			width: 200px;
			height: 144px;
			background-color: $background-1;
			border-radius: 8px;

			.bio-profile-avatar {
				width: 48px;
				height: 48px;
			}

			.bio-profile-username {
				font-family: Roboto;
				font-size: 14px;
				margin-top: 4px;
				font-weight: 500;
				line-height: 20px;
				letter-spacing: 0em;
				text-align: center;
				color: $ink-1;
			}

			.bio-profile-url {
				font-family: Roboto;
				padding: 6px 8px;
				margin-top: 4px;
				font-size: 10px;
				font-weight: 400;
				line-height: 12px;
				letter-spacing: 0em;
				text-align: center;
				border: 1px solid $separator;
				border-radius: 8px;
				word-break: break-word;
			}
		}
	}
}

.bio-profile-picture::before {
	content: '';
	position: absolute;
	top: 0;
	left: 0;
	right: 0;
	bottom: 0;
	backdrop-filter: blur(10px); /* 模糊效果的强度，可以根据需要调整 */
	border-radius: inherit;
	z-index: -1; /* 将伪元素置于底部，避免遮挡内容 */
}
</style>
