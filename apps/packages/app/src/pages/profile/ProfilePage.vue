<template>
	<edit-container v-if="userStore.user">
		<structured-title
			:primary-title="t('profile.header_style')"
			:secondary-items="headerStyleSecondaryItems"
		>
			<template v-slot:primary>
				<grid-picker-group
					column-gap="28px"
					v-model="userStore.user.header.style"
				>
					<picker-component
						:width="userStore.isMobile ? '88px' : '120px'"
						:height="userStore.isMobile ? '152px' : '216px'"
						:value="HEADER_STYLE_TYPE.CLASSIC"
						:label="t('profile.classic')"
						:border="true"
					>
						<template v-slot:default="{ fontSize }">
							<header-preset-component
								:header="classicHeader"
								:social="presetSocial"
								:font-size="fontSize"
							/>
						</template>
						<template v-slot:label="{ selected, label }">
							<check-box-component :model-value="selected" :label="label" />
						</template>
					</picker-component>
					<picker-component
						:width="userStore.isMobile ? '88px' : '120px'"
						:height="userStore.isMobile ? '152px' : '216px'"
						:value="HEADER_STYLE_TYPE.PORTRAIT"
						:label="t('profile.portrait')"
						:border="true"
					>
						<template v-slot:default="{ fontSize }">
							<header-preset-component
								:header="portraitHeader"
								:social="presetSocial"
								:font-size="fontSize"
							/>
						</template>
						<template v-slot:label="{ selected, label }">
							<check-box-component :model-value="selected" :label="label" />
						</template>
					</picker-component>
					<picker-component
						:width="userStore.isMobile ? '88px' : '120px'"
						:height="userStore.isMobile ? '152px' : '216px'"
						:label="t('profile.banner')"
						:value="HEADER_STYLE_TYPE.BANNER"
						:border="true"
					>
						<template v-slot:default="{ fontSize }">
							<header-preset-component
								:header="bannerHeader"
								:social="presetSocial"
								:font-size="fontSize"
							/>
						</template>
						<template v-slot:label="{ selected, label }">
							<check-box-component :model-value="selected" :label="label" />
						</template>
					</picker-component>
				</grid-picker-group>
			</template>

			<template v-slot:secondary-0>
				<upload-component
					v-model:img-url="userStore.user.header.banner"
					placeholder="profile/banner/banner_default.jpg"
					width="100%"
					height="220px"
				/>
			</template>
		</structured-title>

		<structured-title
			:primary-title="t('profile.profile_picture')"
			:primary-description="t('profile.upload_avatar_desc')"
			:secondary-items="pictureSecondaryItems"
		>
			<template v-slot:primary>
				<bio-avatar :size="100" :icon-size="32" />
			</template>

			<template v-slot:secondary-0>
				<grid-picker-group
					:grid="userStore.isMobile"
					:columns="2"
					v-model="userStore.user.header.profileShape"
				>
					<picker-component
						:label="t('base.circular')"
						:auto="!userStore.isMobile"
						:value="PROFILE_SHAPE_TYPE.CIRCULAR"
					>
						<template v-slot:default="{ color }">
							<div
								:style="{ background: color }"
								class="profile_shape_circle"
							/>
						</template>
						<template v-slot:label="{ selected, label }">
							<check-box-component :model-value="selected" :label="label" />
						</template>
					</picker-component>
					<picker-component
						:label="t('base.square')"
						:auto="!userStore.isMobile"
						:class="!userStore.isMobile ? 'q-ml-lg' : ''"
						:value="PROFILE_SHAPE_TYPE.SQUARE"
					>
						<template v-slot:default="{ color }">
							<div
								:style="{ background: color }"
								class="profile_shape_square"
							/>
						</template>
						<template v-slot:label="{ selected, label }">
							<check-box-component :model-value="selected" :label="label" />
						</template>
					</picker-component>
				</grid-picker-group>

				<slider-component
					v-if="userStore.user.header.style !== HEADER_STYLE_TYPE.PORTRAIT"
					:label="t('profile.profile_picture_size')"
					v-model="userStore.user.header.profileSize"
					:min="64"
					:max="200"
					unit="px"
				/>

				<switch-component
					v-if="userStore.user.header.style !== HEADER_STYLE_TYPE.PORTRAIT"
					:label="t('profile.profile_picture_outline')"
					v-model="userStore.user.header.profileOutline"
				/>
			</template>
		</structured-title>

		<structured-title
			:primary-title="t('profile.header_text')"
			:secondary-items="[{ title: t('base.size'), visible: true }]"
		>
			<template v-slot:primary>
				<div class="column">
					<div class="edit-label full-width">
						{{ t('base.name') }}
					</div>
					<edit-view
						v-if="userStore.user"
						class="label-width q-mt-xs"
						:placeholder="t('profile.your_name')"
						v-model="userStore.user.header.nickName"
					/>
					<div class="edit-label full-width q-mt-md">
						{{ t('profile.bio') }}
					</div>
					<edit-view
						v-if="userStore.user"
						height="96px"
						class="label-width q-mt-xs"
						:placeholder="t('profile.a_bit_about_you')"
						v-model="userStore.user.header.description"
					/>
				</div>
			</template>

			<template v-slot:secondary-0>
				<grid-picker-group
					:grid="false"
					v-model="userStore.user.header.textSize"
				>
					<picker-component text="S" :value="SIZE_TYPE.SMALL" />
					<picker-component
						class="q-ml-md"
						text="M"
						:value="SIZE_TYPE.MEDIUM"
					/>
					<picker-component
						class="q-ml-md"
						text="L"
						:value="SIZE_TYPE.LARGER"
					/>
				</grid-picker-group>
			</template>
		</structured-title>

		<structured-title
			v-if="userStore.user.header.style === HEADER_STYLE_TYPE.CLASSIC"
			:primary-title="t('profile.header_format')"
		>
			<template v-slot:primary>
				<grid-picker-group columns="2" v-model="userStore.user.header.format">
					<picker-component
						:width="userStore.isMobile ? '157px' : '210px'"
						:height="userStore.isMobile ? '80px' : '100px'"
						:value="HEADER_FORMAT_TYPE.COLUMN"
					>
						<template v-slot:default="{ color }">
							<format-svg :color="color" :column="true" />
						</template>
					</picker-component>
					<picker-component
						:width="userStore.isMobile ? '157px' : '210px'"
						:height="userStore.isMobile ? '80px' : '100px'"
						:value="HEADER_FORMAT_TYPE.ROW"
					>
						<template v-slot:default="{ color }">
							<format-svg :color="color" :column="false" />
						</template>
					</picker-component>
				</grid-picker-group>
			</template>
		</structured-title>
	</edit-container>
</template>

<script lang="ts" setup>
import HeaderPresetComponent from '@apps/profile/src/components/profile/layout/preset/HeaderPresetComponent.vue';
import CheckBoxComponent from '@apps/profile/src/components/profile/base/CheckBoxComponent.vue';
import GridPickerGroup from '@apps/profile/src/components/profile/base/GridPickerGroup.vue';
import PickerComponent from '@apps/profile/src/components/profile/base/PickerComponent.vue';
import StructuredTitle from '@apps/profile/src/components/profile/base/StructuredTitle.vue';
import SliderComponent from '@apps/profile/src/components/profile/base/SliderComponent.vue';
import SwitchComponent from '@apps/profile/src/components/profile/base/SwitchComponent.vue';
import UploadComponent from '@apps/profile/src/components/profile/base/UploadComponent.vue';
import EditContainer from '@apps/profile/src/pages/profile/EditContainer.vue';
import FormatSvg from '@apps/profile/src/components/profile/base/FormatSvg.vue';
import EditView from '@apps/profile/src/components/profile/base/EditView.vue';
import BioAvatar from '@apps/profile/src/components/profile/base/BioAvatar.vue';
import {
	HEADER_FORMAT_TYPE,
	HEADER_STYLE_TYPE,
	PROFILE_SHAPE_TYPE,
	SIZE_TYPE,
	UserHeader
} from '@apps/profile/src/types/User';
import { SOCIAL_TYPE } from '@apps/profile/src/types/SocialProps';
import { useUserStore } from '@apps/profile/src/stores/profileUser';
import { computed, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import _ from 'lodash';

const userStore = useUserStore();
const { t } = useI18n();
const classicHeader = ref();
const portraitHeader = ref();
const bannerHeader = ref();
const presetSocial = ref({
	data: [
		{
			platform: SOCIAL_TYPE.INSTAGRAM,
			url: '',
			username: ''
		},
		{
			platform: SOCIAL_TYPE.TIKTOK,
			url: '',
			username: ''
		},
		{
			platform: SOCIAL_TYPE.TWITTER,
			url: '',
			username: ''
		}
	],
	size: SIZE_TYPE.MEDIUM
});

const setCommonHeader = (header: UserHeader) => {
	header.profileOutline = false;
	header.profileSize = 100;
	header.textSize = SIZE_TYPE.SMALL;
	header.description = '';
};

watch(
	() => userStore.user?.header,
	(newValue) => {
		classicHeader.value = _.cloneDeep(newValue);
		portraitHeader.value = _.cloneDeep(newValue);
		bannerHeader.value = _.cloneDeep(newValue);
		classicHeader.value.style = HEADER_STYLE_TYPE.CLASSIC;
		portraitHeader.value.style = HEADER_STYLE_TYPE.PORTRAIT;
		bannerHeader.value.style = HEADER_STYLE_TYPE.BANNER;
		setCommonHeader(classicHeader.value);
		setCommonHeader(portraitHeader.value);
		setCommonHeader(bannerHeader.value);
	},
	{
		deep: true,
		immediate: true
	}
);

const headerStyleSecondaryItems = computed(() => {
	if (!userStore.user) {
		return [];
	}

	return [
		{
			title: t('profile.banner'),
			description: t('profile.upload_avatar_desc'),
			visible: userStore.user.header.style === HEADER_STYLE_TYPE.BANNER
		}
	];
});

const pictureSecondaryItems = computed(() => {
	if (!userStore.user) {
		return [];
	}
	return [
		{
			title: t('profile.profile_picture_shape'),
			visible: userStore.user.header.style !== HEADER_STYLE_TYPE.PORTRAIT
		}
	];
});
</script>
<style scoped lang="scss"></style>
