<template>
	<div
		v-if="header.style === HEADER_STYLE_TYPE.CLASSIC"
		class="justify-center items-center"
		:class="
			header.format
				? 'classic-header-format-' +
				  header.format.toLowerCase() +
				  ' ' +
				  header.format.toLowerCase()
				: 'classic-header-format-column column'
		"
	>
		<profile-terminus-avatar
			class="avatar-picture"
			:info="userStore.info"
			:show-border="header.profileOutline"
			:border-color="textColor"
			:size="header.profileSize / 16 + 'em'"
			:radius="
				header.profileShape === PROFILE_SHAPE_TYPE.CIRCULAR ? '50%' : '0.75em'
			"
		/>
		<div
			class="header-text column"
			:style="{
				'--avatarSize': header.profileSize / 16 + 'em'
			}"
		>
			<div
				v-if="header.nickName"
				:class="
					header.textSize
						? 'nickname-' + header.textSize.toLowerCase()
						: 'nickname-medium'
				"
				:style="{ color: textColor }"
			>
				{{ header.nickName }}
			</div>
			<div
				v-if="header.description"
				:class="
					header.textSize
						? 'description-' + header.textSize.toLowerCase()
						: 'description-medium'
				"
				:style="{ color: textColor }"
			>
				{{ header.description }}
			</div>
		</div>
		<social-preview
			:style="{
				marginTop: header.format === HEADER_FORMAT_TYPE.COLUMN ? '' : '1.25em'
			}"
			:social="social"
			:color="textColor"
			:clickable="socialClickable"
			:justify="header.format === HEADER_FORMAT_TYPE.COLUMN ? 'center' : 'left'"
		/>
	</div>

	<div
		v-if="header.style === HEADER_STYLE_TYPE.PORTRAIT"
		class="portrait-header-format-column column justify-center items-center"
	>
		<profile-terminus-avatar
			:info="userStore.info"
			:border-color="textColor"
			:show-border="false"
			size="100%"
			:mask="true"
			:radius="previewPc ? '0.75em' : ''"
			:style="{
				marginTop: previewPc ? '7em' : '',
				maxWidth: previewPc ? '23.43em' : ''
			}"
		/>
		<div class="header-text column">
			<div
				v-if="header.nickName"
				:class="
					header.textSize
						? 'nickname-' + header.textSize.toLowerCase()
						: 'nickname-medium'
				"
				:style="{ color: textColor }"
			>
				{{ header.nickName }}
			</div>
			<div
				v-if="header.description"
				:class="
					header.textSize
						? 'description-' + header.textSize.toLowerCase()
						: 'description-medium'
				"
				:style="{ color: textColor }"
			>
				{{ header.description }}
			</div>
		</div>
		<social-preview
			:social="social"
			:clickable="socialClickable"
			:color="textColor"
		/>
	</div>

	<div
		v-if="header.style === HEADER_STYLE_TYPE.BANNER"
		class="banner-header-format-column column justify-center items-center"
	>
		<profile-terminus-avatar
			class="avatar-picture"
			:border-color="textColor"
			:style="{
				'--avatarMarginTop': `calc(50% - ${
					header.profileSize / 16 / 2 + 2 + 'em'
				} + ${previewPc ? '7.5em' : '0em'})`
			}"
			:info="userStore.info"
			:show-border="header.profileOutline"
			:size="header.profileSize / 16 + 'em'"
			:radius="
				header.profileShape === PROFILE_SHAPE_TYPE.CIRCULAR ? '50%' : '0.75em'
			"
		/>
		<div class="header-text column">
			<div
				v-if="header.nickName"
				:class="
					header.textSize
						? 'nickname-' + header.textSize.toLowerCase()
						: 'nickname-medium'
				"
				:style="{ color: textColor }"
			>
				{{ header.nickName }}
			</div>
			<div
				v-if="header.description"
				:class="
					header.textSize
						? 'description-' + header.textSize.toLowerCase()
						: 'description-medium'
				"
				:style="{ color: textColor }"
			>
				{{ header.description }}
			</div>
		</div>
		<social-preview
			:social="social"
			:clickable="socialClickable"
			:color="textColor"
		/>
	</div>
</template>
<script lang="ts" setup>
import { computed, PropType } from 'vue';
import {
	HEADER_STYLE_TYPE,
	PROFILE_SHAPE_TYPE,
	HEADER_FORMAT_TYPE,
	UserHeader,
	UserSocial
} from '@apps/profile/src/types/User';
import { useUserStore } from '@apps/profile/src/stores/profileUser';
import ProfileTerminusAvatar from '../avatar/ProfileTerminusAvatar.vue';
import SocialPreview from '@apps/profile/src/components/profile/layout/SocialPreview.vue';

defineProps({
	header: {
		type: Object as PropType<UserHeader>,
		required: true
	},
	social: {
		type: Object as PropType<UserSocial>,
		required: true
	},
	textColor: {
		type: String,
		require: true
	},
	socialClickable: {
		type: Boolean,
		default: true
	}
});

const userStore = useUserStore();

const previewPc = computed(() => {
	return !userStore.isMobile && process.env.ACTION === 'PREVIEW';
});
</script>
<style scoped lang="scss">
.classic-header-format-column {
	width: 100%;

	.avatar-picture {
		margin-top: 1.5em;
		margin-bottom: 1.25em;
	}

	.header-text {
		width: 100%;

		.nickname {
			max-width: 100%;
			overflow: hidden;
			text-align: center;
			text-overflow: ellipsis;
			display: -webkit-box;
			-webkit-box-orient: vertical;
			-webkit-line-clamp: 1;
		}

		.nickname-small {
			@extend .nickname;
			font-size: 1.25em;
			font-style: normal;
			font-weight: 700;
		}

		.nickname-medium {
			@extend .nickname;
			font-size: 1.5em;
			font-style: normal;
			font-weight: 700;
		}

		.nickname-larger {
			@extend .nickname;
			font-size: 1.75em;
			font-style: normal;
			font-weight: 700;
		}

		.description {
			max-width: 100%;
			overflow: hidden;
			text-align: center;
			white-space: pre-wrap;
		}

		.description-small {
			@extend .description;
			font-size: 1em;
			font-style: normal;
			font-weight: 400;
		}

		.description-medium {
			@extend .description;
			font-size: 1.125em;
			font-style: normal;
			font-weight: 400;
		}

		.description-larger {
			@extend .description;
			font-size: 1.25em;
			font-style: normal;
			font-weight: 400;
		}
	}
}

.classic-header-format-row {
	width: 100%;

	.header-text {
		width: calc(100% - 0.75em - var(--avatarSize));
		margin-left: 0.75em;

		.nickname {
			max-width: 100%;
			overflow: hidden;
			text-align: left;
			text-overflow: ellipsis;
			display: -webkit-box;
			-webkit-box-orient: vertical;
			-webkit-line-clamp: 1;
		}

		.nickname-small {
			@extend .nickname;
			font-size: 1.25em;
			font-style: normal;
			font-weight: 700;
		}

		.nickname-medium {
			@extend .nickname;
			font-size: 1.5em;
			font-style: normal;
			font-weight: 700;
		}

		.nickname-larger {
			@extend .nickname;
			font-size: 1.75em;
			font-style: normal;
			font-weight: 700;
		}

		.description {
			max-width: 100%;
			overflow: hidden;
			text-align: left;
			white-space: pre-wrap;
		}

		.description-small {
			@extend .description;
			font-size: 1em;
			font-style: normal;
			font-weight: 400;
		}

		.description-medium {
			@extend .description;
			font-size: 1.125em;
			font-style: normal;
			font-weight: 400;
		}

		.description-larger {
			@extend .description;
			font-size: 1.25em;
			font-style: normal;
			font-weight: 400;
		}
	}
}

.portrait-header-format-column {
	width: 100%;

	.avatar-picture {
		margin-bottom: 0;
	}

	.header-text {
		width: 100%;

		.nickname {
			max-width: 100%;
			overflow: hidden;
			text-align: center;
			text-overflow: ellipsis;
			display: -webkit-box;
			-webkit-box-orient: vertical;
			-webkit-line-clamp: 1;
		}

		.nickname-small {
			@extend .nickname;
			font-size: 1.25em;
			font-style: normal;
			font-weight: 700;
			line-height: 1.75em;
		}

		.nickname-medium {
			@extend .nickname;
			font-size: 1.5em;
			font-style: normal;
			font-weight: 700;
			line-height: 2em;
		}

		.nickname-larger {
			@extend .nickname;
			font-size: 1.75em;
			font-style: normal;
			font-weight: 700;
		}

		.description {
			max-width: 100%;
			overflow: hidden;
			text-align: center;
			white-space: pre-wrap;
		}

		.description-small {
			@extend .description;
			font-size: 1em;
			font-style: normal;
			font-weight: 400;
		}

		.description-medium {
			@extend .description;
			font-size: 1.125em;
			font-style: normal;
			font-weight: 400;
		}

		.description-larger {
			@extend .description;
			font-size: 1.25em;
			font-style: normal;
			font-weight: 400;
		}
	}
}

.banner-header-format-column {
	width: 100%;

	.avatar-picture {
		margin-top: var(--avatarMarginTop);
		margin-bottom: 1.25em;
	}

	.header-text {
		width: 100%;

		.nickname {
			max-width: 100%;
			overflow: hidden;
			text-align: center;
			text-overflow: ellipsis;
			display: -webkit-box;
			-webkit-box-orient: vertical;
			-webkit-line-clamp: 1;
		}

		.nickname-small {
			@extend .nickname;
			font-size: 1.25em;
			font-style: normal;
			font-weight: 700;
		}

		.nickname-medium {
			@extend .nickname;
			font-size: 1.5em;
			font-style: normal;
			font-weight: 700;
		}

		.nickname-larger {
			@extend .nickname;
			font-size: 1.75em;
			font-style: normal;
			font-weight: 700;
		}

		.description {
			max-width: 100%;
			overflow: hidden;
			text-align: center;
			white-space: pre-wrap;
		}

		.description-small {
			@extend .description;
			font-size: 1em;
			font-style: normal;
			font-weight: 400;
		}

		.description-medium {
			@extend .description;
			font-size: 1.125em;
			font-style: normal;
			font-weight: 400;
		}

		.description-larger {
			@extend .description;
			font-size: 1.25em;
			font-style: normal;
			font-weight: 400;
		}
	}
}
</style>
