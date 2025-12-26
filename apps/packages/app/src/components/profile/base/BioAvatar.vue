<template>
	<div
		class="cursor-pointer"
		:style="{
			position: 'relative',
			width: size + 'px',
			height: size + 'px',
			'--avatarWidth': size + 'px',
			'--iconSize': iconSize + 'px'
		}"
		@click="openDialog"
	>
		<profile-terminus-avatar
			class="profile-avatar"
			v-if="userStore.info"
			@mouseover="setHover(true)"
			@mouseleave="setHover(false)"
			:show-image-border="false"
			:info="userStore.info"
			:size="size + 'px'"
		>
			<div
				v-if="hover && showHover"
				class="avatar-hover-background column justify-center items-center cursor-pointer"
				@click="openDialog"
			>
				<div class="avatar-hover-edit">{{ t('profile.edit') }}</div>
			</div>
		</profile-terminus-avatar>

		<div
			v-else
			class="profile-avatar-default column justify-center items-center"
			@click="openDialog"
		>
			<q-icon name="sym_r_account_circle" size="20px" />
			<span class="profile-avatar-desc">{{ t('profile.avatar') }}</span>
		</div>

		<q-img
			class="avatar-photo"
			:src="getRequireImage('profile/avatar_photo.svg')"
		/>
	</div>
</template>

<script setup lang="ts">
import ProfileTerminusAvatar from '../avatar/ProfileTerminusAvatar.vue';
import AvatarChooseDialog from '../avatar/AvatarChooseDialog.vue';
import { getRequireImage } from 'src/utils/helper';
import { useUserStore } from '@apps/profile/src/stores/profileUser';
import { useRouter } from 'vue-router';
import { useQuasar } from 'quasar';
import { useI18n } from 'vue-i18n';
import { ref } from 'vue';

const userStore = useUserStore();
const hover = ref(false);
const router = useRouter();
const { t } = useI18n();
const $q = useQuasar();

defineProps({
	size: {
		type: Number,
		require: true
	},
	showHover: {
		type: Boolean,
		default: false
	},
	iconSize: {
		type: Number,
		require: true
	}
});

async function openDialog() {
	if (userStore.isMobile) {
		router.push({
			path: '/avatar'
		});
	} else {
		$q.dialog({
			component: AvatarChooseDialog
		})
			.onOk((data) => {
				if (!userStore.user) {
					return;
				}
				userStore.user.header.avatarUrl = data;
			})
			.onCancel(() => {
				console.log('Cancel');
			})
			.onDismiss(() => {
				console.log('Called on OK or Cancel');
			});
	}
}

const setHover = (isHover: boolean) => {
	hover.value = isHover;
};
</script>

<style scoped lang="scss">
.profile-avatar {
	overflow: hidden;
	position: relative;

	.avatar-hover-background {
		background: linear-gradient(0deg, #1f1814 0%, rgba(31, 24, 20, 0) 100%);
		width: var(--avatarWidth);
		height: var(--iconSize);
		opacity: 0.8;
		position: absolute;
		bottom: 0;

		.avatar-hover-edit {
			font-family: Roboto;
			font-size: 12px;
			font-weight: 400;
			line-height: 16px;
			letter-spacing: 0em;
			text-align: center;
			color: var(--Basic-White, #fff);
		}
	}
}

.profile-avatar-default {
	@extend .profile-avatar;
	border: 1px solid var(--border, #ebebeb);
	color: var(--grey-05, #adadad);

	.profile-avatar-desc {
		font-family: Roboto;
		font-size: 12px;
		font-weight: 400;
		line-height: 16px;
		letter-spacing: 0em;
		text-align: center;
		margin-top: 6px;
	}
}

.avatar-photo {
	width: var(--iconSize);
	height: var(--iconSize);
	position: absolute;
	bottom: 0;
	right: 0;
	stroke: $background-1;
	stroke-width: 2px;
	fill: $ink-on-brand-black;
}
</style>
