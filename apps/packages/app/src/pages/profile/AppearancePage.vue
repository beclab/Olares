<template>
	<edit-container v-if="userStore.user">
		<structured-title
			:primary-title="t('appearance.theme')"
			:secondary-items="themeSecondaryItems"
		>
			<template v-slot:primary>
				<grid-picker-group v-model="themeType">
					<picker-component :value="THEME_TYPE.SOLID">
						<template v-slot:default="{ color, selected }">
							<div
								:style="{ color: selected ? color : link1.color.value }"
								class="row q-py-md"
							>
								<q-icon name="sym_r_palette" size="20px" />
								<div class="q-ml-sm text-subtitle2">
									{{ t('appearance.solid') }}
								</div>
							</div>
						</template>
					</picker-component>

					<picker-component :value="THEME_TYPE.GRADIENT">
						<template v-slot:default="{ color, selected }">
							<div
								:style="{ color: selected ? color : link1.color.value }"
								class="row q-py-md"
							>
								<q-icon name="sym_r_gradient" size="20px" />
								<div class="q-ml-sm text-subtitle2">
									{{ t('appearance.gradient') }}
								</div>
							</div>
						</template>
					</picker-component>

					<picker-component :value="THEME_TYPE.IMAGE">
						<template v-slot:default="{ color, selected }">
							<div
								:style="{ color: selected ? color : link1.color.value }"
								class="row q-py-md"
							>
								<q-icon name="sym_r_imagesmode" size="20px" />
								<div class="q-ml-sm text-subtitle2">
									{{ t('base.image') }}
								</div>
							</div>
						</template>
					</picker-component>
				</grid-picker-group>

				<BtUploader
					v-if="themeType === THEME_TYPE.IMAGE"
					width="100%"
					height="40px"
					:size="5"
					fileName="image"
					accept=".jpg, .jpeg, .png, .gif, image/*"
					action="/images/upload/v1"
					@ok="ok"
					@loading="update"
					@fail="fail"
				>
					<bio-button
						v-if="!loading"
						class="q-mt-lg text-body1"
						size="24px"
						:label="t('appearance.upload_an_image')"
						icon="sym_r_add"
						width="100%"
					/>
					<bio-button
						v-else
						class="q-mt-lg text-body1"
						size="24px"
						disabled
						width="100%"
						:label="t('appearance.uploading')"
					/>
				</BtUploader>
			</template>

			<template v-slot:secondary-0>
				<grid-picker-group
					column-gap="28px"
					v-model="userStore.user.appearance.theme.filter"
				>
					<picker-component
						:width="userStore.isMobile ? '88px' : '120px'"
						:height="userStore.isMobile ? '152px' : '216px'"
						:value="IMAGE_FILTER.NONE"
						:label="t('base.none')"
						:border="true"
					>
						<template v-slot:default="{ fontSize }">
							<background-preset-component
								:theme="noneFilter"
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
						:value="IMAGE_FILTER.DARK"
						:label="t('appearance.dark')"
						:border="true"
					>
						<template v-slot:default="{ fontSize }">
							<background-preset-component
								:theme="darkFilter"
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
						:label="t('appearance.light')"
						:value="IMAGE_FILTER.Light"
						:border="true"
					>
						<template v-slot:default="{ fontSize }">
							<background-preset-component
								:theme="lightFilter"
								:font-size="fontSize"
							/>
						</template>
						<template v-slot:label="{ selected, label }">
							<check-box-component :model-value="selected" :label="label" />
						</template>
					</picker-component>
				</grid-picker-group>
			</template>

			<template v-slot:secondary-1>
				<grid-picker-group
					:columns="4"
					column-gap="12px"
					row-gap="12px"
					v-if="themeType === THEME_TYPE.SOLID"
					:model-value="userStore.user.appearance.theme.preset"
				>
					<template v-for="item in BACKGROUND_COLOR_PRESET" :key="item.preset">
						<picker-component
							:width="userStore.isMobile ? '77px' : '92px'"
							:height="userStore.isMobile ? '128px' : '160px'"
							:value="item.preset"
							@on-selected="onThemeSelected(item)"
							:border="true"
						>
							<template v-slot:default="{ fontSize }">
								<background-preset-component
									:font-size="fontSize"
									:theme="item"
								/>
							</template>
						</picker-component>
					</template>
				</grid-picker-group>

				<grid-picker-group
					:columns="4"
					column-gap="12px"
					row-gap="12px"
					v-if="themeType === THEME_TYPE.GRADIENT"
					:model-value="userStore.user.appearance.theme.preset"
				>
					<template
						v-for="item in BACKGROUND_GRADIENT_PRESET"
						:key="item.preset"
					>
						<picker-component
							:width="userStore.isMobile ? '77px' : '92px'"
							:height="userStore.isMobile ? '128px' : '160px'"
							:value="item.preset"
							@on-selected="onThemeSelected(item)"
							:border="true"
						>
							<template v-slot:default="{ fontSize }">
								<background-preset-component
									:font-size="fontSize"
									:theme="item"
								/>
							</template>
						</picker-component>
					</template>
				</grid-picker-group>

				<grid-picker-group
					:columns="4"
					column-gap="12px"
					row-gap="12px"
					v-if="themeType === THEME_TYPE.IMAGE"
					:model-value="userStore.user.appearance.theme.preset"
				>
					<template v-for="item in BACKGROUND_IMAGE_PRESET" :key="item.preset">
						<picker-component
							:width="userStore.isMobile ? '77px' : '92px'"
							:height="userStore.isMobile ? '128px' : '160px'"
							:value="item.preset"
							@on-selected="onThemeSelected(item)"
							:border="true"
						>
							<template v-slot:default="{ fontSize }">
								<background-preset-component
									:font-size="fontSize"
									:theme="item"
								/>
							</template>
						</picker-component>
					</template>
				</grid-picker-group>
			</template>

			<template v-slot:secondary-2>
				<div class="full-width column">
					<color-picker-v2
						v-if="userStore.user.appearance.theme.style === THEME_TYPE.SOLID"
						:label="t('appearance.background')"
						v-model="userStore.user.appearance.theme.background"
						@update:modelValue="onColorChange"
					/>
					<color-picker-v2
						v-if="userStore.user.appearance.theme.style === THEME_TYPE.GRADIENT"
						:label="t('appearance.background_color_1')"
						v-model="userStore.user.appearance.theme.gradientTopColor"
						@update:modelValue="onColorChange"
					/>
					<color-picker-v2
						v-if="userStore.user.appearance.theme.style === THEME_TYPE.GRADIENT"
						:label="t('appearance.background_color_2')"
						v-model="userStore.user.appearance.theme.gradientBottomColor"
						@update:modelValue="onColorChange"
					/>
					<color-picker-v2
						:label="t('appearance.header_text_icons')"
						v-model="userStore.user.appearance.theme.header.textColor"
						@update:modelValue="onColorChange"
					/>
				</div>
			</template>

			<template v-slot:secondary-3>
				<div class="full-width column">
					<color-picker-v2
						:label="t('appearance.background')"
						v-model="userStore.user.appearance.theme.link.background"
						@update:modelValue="onColorChange"
					/>
					<color-picker-v2
						:label="t('appearance.text')"
						v-model="userStore.user.appearance.theme.link.textColor"
						@update:modelValue="onColorChange"
					/>
				</div>
			</template>
			<template v-slot:secondary-4>
				<div class="full-width column">
					<color-picker-v2
						:label="t('appearance.background')"
						v-model="userStore.user.appearance.theme.block.background"
						@update:modelValue="onColorChange"
					/>
					<color-picker-v2
						:label="t('appearance.text')"
						v-model="userStore.user.appearance.theme.block.textColor"
						@update:modelValue="onColorChange"
					/>
				</div>
			</template>
		</structured-title>

		<structured-title :primary-title="t('appearance.block_style')">
			<template v-slot:primary>
				<grid-picker-group v-model="userStore.user.appearance.block.style">
					<picker-component
						:label="t('base.square')"
						:value="BLOCK_STYLE_TYPE.SQUARE"
					>
						<template v-slot:default="{ color }">
							<div :style="{ background: color }" class="shape_square" />
						</template>
						<template v-slot:label="{ selected, label }">
							<check-box-component :model-value="selected" :label="label" />
						</template>
					</picker-component>
					<picker-component
						:label="t('base.round')"
						:value="BLOCK_STYLE_TYPE.ROUND"
					>
						<template v-slot:default="{ color }">
							<div :style="{ background: color }" class="shape_round" />
						</template>
						<template v-slot:label="{ selected, label }">
							<check-box-component :model-value="selected" :label="label" />
						</template>
					</picker-component>
					<picker-component
						:label="t('base.custom')"
						:value="BLOCK_STYLE_TYPE.CUSTOM"
					>
						<template v-slot:default="{ color }">
							<div :style="{ background: color }" class="shape_square" />
						</template>
						<template v-slot:label="{ selected, label }">
							<check-box-component :model-value="selected" :label="label" />
						</template>
					</picker-component>
				</grid-picker-group>

				<slider-component
					v-if="
						userStore.user.appearance.block.style === BLOCK_STYLE_TYPE.CUSTOM
					"
					:label="t('appearance.corner_radius')"
					v-model="userStore.user.appearance.block.cornerRadius"
					:min="4"
					:max="20"
					unit="px"
				/>

				<switch-component
					:label="t('appearance.block_shadow')"
					v-model="userStore.user.appearance.block.shadow"
				/>

				<switch-component
					:label="t('appearance.block_outline')"
					v-model="userStore.user.appearance.block.outline"
				/>

				<slider-component
					:label="t('appearance.block_transparency')"
					v-model="userStore.user.appearance.block.transparency"
					:min="0"
					:max="100"
					unit="%"
				/>
			</template>
		</structured-title>

		<structured-title :primary-title="t('appearance.font')">
			<template v-slot:primary>
				<grid-picker-group v-model="userStore.user.appearance.font">
					<template v-for="item in FONT_ARRAY" :key="item">
						<picker-component :value="item">
							<template v-slot:default="{ color, selected }">
								<div
									class="q-py-md text-font"
									:style="{
										color: selected ? color : link1.color.value,
										fontFamily: item
									}"
								>
									{{ item }}
								</div>
							</template>
						</picker-component>
					</template>
				</grid-picker-group>
			</template>
		</structured-title>
	</edit-container>
</template>

<script lang="ts" setup>
import BackgroundPresetComponent from '@apps/profile/src/components/profile/layout/preset/BackgroundPresetComponent.vue';
import CheckBoxComponent from '@apps/profile/src/components/profile/base/CheckBoxComponent.vue';
import StructuredTitle from '@apps/profile/src/components/profile/base/StructuredTitle.vue';
import PickerComponent from '@apps/profile/src/components/profile/base/PickerComponent.vue';
import GridPickerGroup from '@apps/profile/src/components/profile/base/GridPickerGroup.vue';
import SwitchComponent from '@apps/profile/src/components/profile/base/SwitchComponent.vue';
import SliderComponent from '@apps/profile/src/components/profile/base/SliderComponent.vue';
import ColorPickerV2 from '@apps/profile/src/components/profile/design/ColorPickerV2.vue';
import EditContainer from '@apps/profile/src/pages/profile/EditContainer.vue';
import BioButton from '@apps/profile/src/components/profile/base/BioButton.vue';
import {
	AppearanceTheme,
	BLOCK_STYLE_TYPE,
	IMAGE_FILTER,
	THEME_TYPE
} from '@apps/profile/src/types/User';
import {
	BACKGROUND_COLOR_PRESET,
	BACKGROUND_GRADIENT_PRESET,
	BACKGROUND_IMAGE_PRESET,
	FONT_ARRAY
} from '@apps/profile/src/types/Preset';
import { BtNotify, NotifyDefinedType, useColor } from '@bytetrade/ui';
import { computed, onMounted, ref, watch } from 'vue';
import { useUserStore } from '@apps/profile/src/stores/profileUser';
import { useI18n } from 'vue-i18n';
import _ from 'lodash';

const userStore = useUserStore();
const { t } = useI18n();
const link1 = useColor('link-1');
const noneFilter = ref();
const darkFilter = ref();
const lightFilter = ref();
const themeType = ref(THEME_TYPE.SOLID);
const loading = ref(false);

onMounted(() => {
	if (userStore.user) {
		console.log('test', userStore.user.appearance.theme.style);
		themeType.value = userStore.user.appearance.theme.style;
	}
});

const themeSecondaryItems = computed(() => {
	if (!userStore.user) {
		return [];
	}

	let backgroundTitle = '';

	switch (themeType.value) {
		case THEME_TYPE.SOLID:
			backgroundTitle = t('appearance.solid_color_theme');
			break;
		case THEME_TYPE.GRADIENT:
			backgroundTitle = t('appearance.gradient_theme');
			break;
		case THEME_TYPE.IMAGE:
			backgroundTitle = t('appearance.image_theme');
			break;
	}

	const data = [
		{
			title: t('appearance.filter'),
			visible:
				themeType.value === THEME_TYPE.IMAGE &&
				userStore.user.appearance.theme.style === THEME_TYPE.IMAGE
		},
		{
			title: backgroundTitle,
			visible: true
		},
		{
			title: t('appearance.page'),
			visible: true
		},
		{
			title: t('appearance.link_block'),
			visible: true
		},
		{
			title: t('appearance.block'),
			visible: true
		}
	];
	console.log(data);
	return data;
});

watch(
	() => userStore.user?.appearance.theme,
	(newValue) => {
		noneFilter.value = _.cloneDeep(newValue);
		darkFilter.value = _.cloneDeep(newValue);
		lightFilter.value = _.cloneDeep(newValue);
		noneFilter.value.filter = IMAGE_FILTER.NONE;
		darkFilter.value.filter = IMAGE_FILTER.DARK;
		lightFilter.value.filter = IMAGE_FILTER.Light;
	},
	{
		deep: true,
		immediate: true
	}
);

const onThemeSelected = (item: AppearanceTheme) => {
	if (userStore.user && item) {
		console.log(item);
		console.log(_.cloneDeep(item));
		userStore.user.appearance.theme = _.cloneDeep(item);
	}
};

const ok = (response: { code: string; data: any; message: any }) => {
	loading.value = false;
	if (userStore.user) {
		userStore.user.appearance.theme.style = THEME_TYPE.IMAGE;
		userStore.user.appearance.theme.preset = '';
		userStore.user.appearance.theme.useUpload = true;
		userStore.user.appearance.theme.uploadImg = response.data.imageUrl;
	}
};

const fail = (response: { code: string; data: any; message: any }) => {
	loading.value = false;
	BtNotify.show({
		type: NotifyDefinedType.FAILED,
		message: response.message
	});
};

const onColorChange = () => {
	if (userStore.user) {
		userStore.user.appearance.theme.preset = '';
	}
};

const update = (status: boolean) => {
	loading.value = status;
};
</script>
<style lang="scss"></style>
