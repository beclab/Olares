<template>
	<page-title-component
		:show-back="false"
		:title="t(`home_menus.${MENU_TYPE.Appearance.toLowerCase()}`)"
	/>

	<bt-scroll-area class="nav-height-scroll-area-conf text-ink-1">
		<AdaptiveLayout>
			<template v-slot:pc>
				<bt-list first>
					<bt-form-item :margin-top="false" :width-separator="false">
						<template v-slot:title>
							<div class="text-subtitle1">
								{{ t('language') }}
							</div>
						</template>
						<bt-select
							v-model="currentLanguage"
							:options="languages"
							@update:modelValue="languageUpdate"
						/>
					</bt-form-item>
				</bt-list>

				<bt-list>
					<div class="row justify-between select-radio-bg">
						<div class="text-subtitle1 text-ink-1">
							{{ t('theme') }}
						</div>
						<div class="row">
							<wallpaper-image
								class="q-mr-xs"
								:width="166"
								:border-radius="8"
								src="settings/theme/light.jpg"
								:selected="backgroundStore.theme == ThemeDefinedMode.LIGHT"
								@click="themeUpdate(ThemeDefinedMode.LIGHT)"
							>
								<template v-slot:legend>
									<bt-check-box-component
										class="q-mt-md"
										:model-value="
											backgroundStore.theme == ThemeDefinedMode.LIGHT
										"
										:label="t(themeOptionsRef[0].label)"
										@update:modelValue="
											backgroundStore.theme = ThemeDefinedMode.LIGHT
										"
									/>
								</template>
							</wallpaper-image>

							<wallpaper-image
								:width="166"
								:border-radius="8"
								style=""
								src="settings/theme/dark.jpg"
								:selected="backgroundStore.theme == ThemeDefinedMode.DARK"
								@click="themeUpdate(ThemeDefinedMode.DARK)"
							>
								<template v-slot:legend>
									<bt-check-box-component
										class="q-mt-md"
										:model-value="
											backgroundStore.theme == ThemeDefinedMode.DARK
										"
										:label="t(themeOptionsRef[1].label)"
										@update:modelValue="
											backgroundStore.theme = ThemeDefinedMode.DARK
										"
									/>
								</template>
							</wallpaper-image>
						</div>
					</div>
				</bt-list>
			</template>
			<template v-slot:mobile>
				<div
					class="mobile-items-list"
					style="padding-bottom: 4px; padding-top: 4px"
				>
					<bt-form-item
						:title="t('language')"
						:margin-top="false"
						:width-separator="false"
					>
						<bt-select
							v-model="currentLanguage"
							:options="languages"
							@update:modelValue="languageUpdate"
						/>
					</bt-form-item>
				</div>
				<div class="text-subtitle2-m text-ink-1 q-mt-lg q-mb-sm">
					{{ t('theme') }}
				</div>
				<div
					class="row mobile-items-list items-center justify-center"
					style="height: 212px"
				>
					<wallpaper-image
						class="q-mr-xl"
						:width="72"
						src="settings/theme/mobile_light.png"
						:border-width="2"
						:border-radius="12"
						:selected="backgroundStore.theme == ThemeDefinedMode.LIGHT"
						@click="themeUpdate(ThemeDefinedMode.LIGHT)"
					>
						<template v-slot:legend>
							<bt-check-box-component
								class="q-mt-md"
								:model-value="backgroundStore.theme == ThemeDefinedMode.LIGHT"
								:label="t(themeOptionsRef[0].label)"
								@update:modelValue="
									backgroundStore.theme = ThemeDefinedMode.LIGHT
								"
							/>
						</template>
					</wallpaper-image>

					<wallpaper-image
						:width="72"
						:border-width="2"
						:border-radius="12"
						src="settings/theme/mobile_dark.png"
						:selected="backgroundStore.theme == ThemeDefinedMode.DARK"
						@click="themeUpdate(ThemeDefinedMode.DARK)"
					>
						<template v-slot:legend>
							<bt-check-box-component
								class="q-mt-md"
								:model-value="backgroundStore.theme == ThemeDefinedMode.DARK"
								:label="t(themeOptionsRef[1].label)"
								@update:modelValue="
									backgroundStore.theme = ThemeDefinedMode.DARK
								"
							/>
						</template>
					</wallpaper-image>
				</div>
			</template>
		</AdaptiveLayout>
		<AdaptiveLayout>
			<template v-slot:pc>
				<bt-list class="q-mb-lg">
					<div class="row justify-between select-radio-bg">
						<div class="text-subtitle1 text-ink-1">
							{{ t('wallpaper') }}
						</div>
						<div class="row">
							<wallpaper-image
								:width="166"
								class="q-mr-xs"
								:src="desktopImgUrl"
								:border-radius="8"
								:selected="selectBackgroundMode == BackgroundMode.desktop"
								@click="selectBackgroundMode = BackgroundMode.desktop"
							>
								<template v-slot:legend>
									<bt-check-box-component
										class="q-mt-md"
										:model-value="
											selectBackgroundMode == BackgroundMode.desktop
										"
										:label="t('desktop_background')"
										@update:modelValue="
											selectBackgroundMode = BackgroundMode.desktop
										"
									/>
								</template>
							</wallpaper-image>

							<wallpaper-image
								:width="166"
								style=""
								:src="loginImgUrl.replace('/bg/', '/login/')"
								:border-radius="8"
								:selected="selectBackgroundMode == BackgroundMode.login"
								@click="selectBackgroundMode = BackgroundMode.login"
							>
								<template v-slot:legend>
									<bt-check-box-component
										class="q-mt-md"
										:model-value="selectBackgroundMode == BackgroundMode.login"
										:label="t('login_background')"
										@update:modelValue="
											selectBackgroundMode = BackgroundMode.login
										"
									/>
								</template>
							</wallpaper-image>
						</div>
					</div>
					<bt-separator
						style="
							margin-left: 20px;
							margin-right: 20px;
							width: calc(100% - 40px);
						"
					/>

					<div class="select-avatar-list-bg">
						<div class="select-avatar-title-bg row items-center justify-start">
							<q-icon
								name="sym_r_imagesmode"
								color="ink-1 q-ml-lg"
								size="20px"
							/>
							<div class="text-subtitle2 select-avatar-title">
								{{ t('pictures') }}
							</div>
						</div>
						<div class="images-list-bg row justify-start">
							<BtUploader
								:size="5"
								width="98px"
								height="58px"
								fileName="image"
								accept=".jpg, image/*"
								action="/images/upload/v1"
								:parmas="uploadParams"
								@ok="ok"
								@fail="fail"
							>
								<wallpaper-image
									:width="92"
									:padding="2"
									:src="
										$q.dark.isActive
											? 'upload_default_dark.svg'
											: 'upload_default.svg'
									"
									:selected="false"
								/>
							</BtUploader>

							<template
								v-for="(item, index) of uploadBackgrounds"
								:key="`bg` + index"
							>
								<wallpaper-image
									v-if="!!item"
									:width="92"
									:padding="2"
									:src="
										item.replace(
											'/resources/Home/Pictures',
											'/api/preview/drive/Home/Pictures'
										) + '?auth=&inline=true&size=big'
									"
									:selected="selectedImgUrl.value === item"
									:deleteEnable="true"
									@deleteI="deletePicture(item)"
									@click="onSelectPicture(item)"
								/>
							</template>

							<template v-for="index in picturesCount" :key="`paper` + index">
								<wallpaper-image
									:width="92"
									:src="
										selectBackgroundMode == BackgroundMode.desktop
											? `settings/bg/${index - 1}.jpg`
											: `settings/login/${index - 1}.jpg`
									"
									:padding="2"
									:selected="
										selectedImgUrl.value === `/settings/bg/${index - 1}.jpg`
									"
									@click="onSelectPicture(`/bg/${index - 1}.jpg`)"
								/>
							</template>
						</div>
					</div>
				</bt-list>
			</template>
		</AdaptiveLayout>
	</bt-scroll-area>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { BackgroundMode, MENU_TYPE } from 'src/constant';
import {
	useBackgroundStore,
	themeOptions
} from 'src/stores/settings/background';
import WallpaperImage from 'src/components/settings/WallpaperImage.vue';
import PageTitleComponent from 'src/components/settings/PageTitleComponent.vue';
import AdaptiveLayout from 'src/components/settings/AdaptiveLayout.vue';
import BtSelect from 'src/components/settings/base/BtSelect.vue';
import BtFormItem from 'src/components/settings/base/BtFormItem.vue';
import { supportLanguages, SupportLanguageType } from 'src/i18n';
import ReminderDialogComponent from 'src/components/settings/ReminderDialogComponent.vue';
import BtSeparator from 'src/components/settings/base/BtSeparator.vue';
import BtCheckBoxComponent from 'src/components/settings/base/BtCheckBoxComponent.vue';
import { debounce, useQuasar } from 'quasar';
import { ThemeDefinedMode } from '@bytetrade/ui';
import { useI18n } from 'vue-i18n';
import BtList from 'src/components/settings/base/BtList.vue';

const backgroundStore = useBackgroundStore();
const selectBackgroundMode = ref(BackgroundMode.desktop);

const { t } = useI18n();

const themeOptionsRef = ref(themeOptions);

const ok = async (response: any) => {
	if (selectBackgroundMode.value == BackgroundMode.desktop) {
		backgroundStore.upload_desktop_background(response.data.imageUrl);
	} else {
		backgroundStore.upload_login_background(response.data.imageUrl);
	}
};

const fail = (response: unknown) => {
	console.log('fail', response);
};

const setDesktopBackground = debounce(async function (item: string) {
	backgroundStore.set_desktop_background(item);
}, 500);

const setLoginBackground = debounce(async function (item: string) {
	backgroundStore.set_login_background(item);
}, 500);

const onSelectPicture = async function (item: string) {
	if (selectBackgroundMode.value == BackgroundMode.desktop) {
		backgroundStore.wallpaper.desktop = item;
		setDesktopBackground(item);
	} else {
		backgroundStore.wallpaper.login = item;
		setLoginBackground(item);
	}
};

const deletePicture = async (item: string) => {
	if (selectBackgroundMode.value == BackgroundMode.desktop) {
		await backgroundStore.delete_desktop_background(item);
	} else {
		await backgroundStore.delete_login_background(item);
	}
	backgroundStore.get_wallpaper();
};

onMounted(async () => {
	backgroundStore.get_wallpaper();
});

const desktopImgUrl = computed(() => {
	return backgroundStore.wallpaper.desktop.startsWith('http')
		? backgroundStore.wallpaper.desktop
		: '/settings' + backgroundStore.wallpaper.desktop;
});
const loginImgUrl = computed(() => {
	return backgroundStore.wallpaper.login.startsWith('http')
		? backgroundStore.wallpaper.login
		: '/settings' + backgroundStore.wallpaper.login;
});

const selectedImgUrl = computed(() => {
	if (selectBackgroundMode.value == BackgroundMode.desktop) {
		return desktopImgUrl;
	} else {
		return loginImgUrl;
	}
});

const picturesCount = computed(() => {
	if (selectBackgroundMode.value == BackgroundMode.desktop) {
		return 28;
	} else {
		return 29;
	}
});

const uploadBackgrounds = computed(() => {
	if (selectBackgroundMode.value == BackgroundMode.desktop) {
		return backgroundStore.wallpaper.upload_desktop_backgrounds;
	} else {
		return backgroundStore.wallpaper.upload_login_backgrounds;
	}
});

const uploadParams = computed(() => {
	if (selectBackgroundMode.value == BackgroundMode.login) {
		return {
			policy: 'public'
		};
	}
	return {};
});

const themeUpdate = (theme: ThemeDefinedMode) => {
	backgroundStore.themeUpdate(theme);
};

const languages = ref(supportLanguages);

const currentLanguage = ref(backgroundStore.locale);
let lastLanguage = backgroundStore.locale;
const $q = useQuasar();

const languageUpdate = (language: SupportLanguageType) => {
	if (backgroundStore.locale == language) {
		return;
	}
	const languageItem = supportLanguages.find((e) => e.value == language);
	if (!languageItem) {
		return;
	}
	$q.dialog({
		component: ReminderDialogComponent,
		componentProps: {
			title: t('Switch language'),
			message: t(
				'Are you sure you need to switch the system language to {language}?',
				{
					language: languageItem.label
				}
			),
			useCancel: true,
			confirmText: t('confirm'),
			cancelText: t('cancel')
		}
	})
		.onOk(async () => {
			await backgroundStore.requestUpdateLanguage(language);
			lastLanguage = backgroundStore.locale;
			currentLanguage.value = backgroundStore.locale;
		})
		.onCancel(() => {
			currentLanguage.value = lastLanguage;
		});
};
</script>

<style scoped lang="scss">
.select-avatar-list-bg {
	margin-top: 16px;

	.select-avatar-title-bg {
		.select-avatar-title {
			margin-left: 4px;
		}
	}

	.images-list-bg {
		width: 100%;
		grid-column-gap: 8px;
		grid-row-gap: 8px;
		padding: 20px;
	}
}

.radio-class {
	margin-top: 8px;
}

.select-radio-bg {
	padding: 20px;
}
</style>
