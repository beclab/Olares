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
					<bt-form-item :width-separator="widgetPrefsStore.showWeight">
						<template v-slot:title>
							<div class="text-subtitle1">
								{{ t('Widget') }}
							</div>
						</template>
						<bt-switch
							size="sm"
							truthy-track-color="blue-default"
							v-model="widgetPrefsStore.showWeight"
							@update:model-value="widgetPrefsStore.update()"
						/>
					</bt-form-item>

					<bt-form-item
						v-if="widgetPrefsStore.showWeight"
						:title="t('Date & Time')"
					>
						<div class="text-body1 text-ink-2">{{ currentDateTime }}</div>
					</bt-form-item>

					<bt-form-item
						v-if="widgetPrefsStore.showWeight"
						:title="t('24-hour format')"
					>
						<bt-switch
							size="sm"
							truthy-track-color="blue-default"
							v-model="widgetPrefsStore.is24HourFormat"
							@update:model-value="widgetPrefsStore.update()"
						/>
					</bt-form-item>

					<bt-form-item
						v-if="widgetPrefsStore.showWeight"
						:title="t('Date format')"
					>
						<bt-select
							v-model="widgetPrefsStore.dateFormat"
							:options="dateFormatOptions"
							@update:modelValue="widgetPrefsStore.update()"
						/>
					</bt-form-item>

					<bt-form-item
						v-if="widgetPrefsStore.showWeight"
						:title="t('Show Dashboard')"
						:width-separator="false"
					>
						<bt-switch
							size="sm"
							truthy-track-color="blue-default"
							v-model="widgetPrefsStore.showDashboard"
							@update:model-value="widgetPrefsStore.update()"
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
						<div class="column justify-end">
							<div class="row justify-end items-center q-mb-lg">
								<BtUploader
									:size="5"
									width="110px"
									height="32px"
									class="q-mr-md"
									:file-guard="imagesUploadFormatGuard"
									fileName="image"
									:accept="IMAGES_UPLOAD_V1_ACCEPT"
									action="/images/upload/v1"
									:parmas="uploadParams"
									@ok="ok"
								>
									<div
										class="upload-image-btn text-body3 text-ink-2 cursor-pointer"
									>
										{{ t('Upload image') }}
									</div>
								</BtUploader>

								<bt-select-v3
									v-if="selectBackgroundMode == BackgroundMode.login"
									width="120px"
									height="32px"
									inputClass="text-body3"
									v-model="backgroundStore.wallpaper.loginStyle"
									:options="imgContentMode"
									@update:modelValue="setLoginStyleMode"
								/>

								<bt-select-v3
									v-if="selectBackgroundMode == BackgroundMode.desktop"
									width="120px"
									height="32px"
									inputClass="text-body3"
									v-model="backgroundStore.wallpaper.desktopStyle"
									:options="imgContentMode"
									@update:modelValue="setDesktopStyleMode"
								/>
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
											:model-value="
												selectBackgroundMode == BackgroundMode.login
											"
											:label="t('login_background')"
											@update:modelValue="
												selectBackgroundMode = BackgroundMode.login
											"
										/>
									</template>
								</wallpaper-image>
							</div>
						</div>
					</div>
					<bt-separator
						style="
							margin-left: 20px;
							margin-right: 20px;
							width: calc(100% - 40px);
						"
					/>

					<div class="select-avatar-list-bg" v-if="uploadDesktopBackgrounds">
						<div class="row items-center justify-start">
							<q-icon
								name="sym_r_imagesmode"
								color="ink-1 q-ml-lg"
								size="20px"
							/>
							<div class="text-subtitle2 text-ink-1 q-ml-xs">
								{{ t('Uploaded images') }}
							</div>
						</div>
						<div class="images-list-bg row justify-start">
							<template
								v-for="(item, index) of uploadDesktopBackgrounds"
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
						</div>
					</div>

					<div class="select-avatar-list-bg" v-if="uploadLoginBackgrounds">
						<div class="row items-center justify-start">
							<q-icon
								name="sym_r_imagesmode"
								color="ink-1 q-ml-lg"
								size="20px"
							/>
							<div class="text-subtitle2 text-ink-1 q-ml-xs">
								{{ t('Uploaded images') }}
							</div>
						</div>
						<div class="images-list-bg row justify-start">
							<template
								v-for="(item, index) of uploadLoginBackgrounds"
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
						</div>
					</div>

					<div class="select-avatar-list-bg">
						<div class="row items-center justify-start">
							<q-icon
								name="sym_r_imagesmode"
								color="ink-1 q-ml-lg"
								size="20px"
							/>
							<div class="text-subtitle2 text-ink-1 q-ml-xs">
								{{ t('Gallery') }}
							</div>
						</div>
						<div class="images-list-bg row justify-start">
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
import BtList from 'src/components/settings/base/BtList.vue';
import BtSelect from 'src/components/settings/base/BtSelect.vue';
import BtFormItem from 'src/components/settings/base/BtFormItem.vue';
import BtSelectV3 from 'src/components/settings/base/BtSelectV3.vue';
import BtSeparator from 'src/components/settings/base/BtSeparator.vue';
import WallpaperImage from 'src/components/settings/WallpaperImage.vue';
import AdaptiveLayout from 'src/components/settings/AdaptiveLayout.vue';
import PageTitleComponent from 'src/components/settings/PageTitleComponent.vue';
import ReminderDialogComponent from 'src/components/settings/ReminderDialogComponent.vue';
import BtCheckBoxComponent from 'src/components/settings/base/BtCheckBoxComponent.vue';
import { useWidgetPreferencesStore } from 'src/stores/settings/widgetPreferences';
import { supportLanguages, SupportLanguageType } from 'src/i18n';
import { ThemeDefinedMode } from '@bytetrade/ui';
import { computed, onMounted, onUnmounted, ref } from 'vue';
import { debounce, useQuasar } from 'quasar';
import { useI18n } from 'vue-i18n';
import {
	BackgroundMode,
	IMG_CONTENT_MODE,
	MENU_TYPE,
	SelectorProps
} from 'src/constant';
import {
	useBackgroundStore,
	themeOptions
} from 'src/stores/settings/background';
import { notifyFailed } from 'src/utils/settings/btNotify';
import {
	createImagesUploadV1FormatGuard,
	IMAGES_UPLOAD_V1_ACCEPT
} from '../../../utils/upload/imagesUploadV1Formats';

const backgroundStore = useBackgroundStore();
const widgetPrefsStore = useWidgetPreferencesStore();
const selectBackgroundMode = ref(BackgroundMode.desktop);

const currentDateTime = ref('');
let dateTimeTimer: ReturnType<typeof setInterval> | null = null;

const updateCurrentDateTime = () => {
	const { date, time } = widgetPrefsStore.formatNow();
	currentDateTime.value = `${date} ${time}`;
};

const dateFormatOptions = [
	{ label: 'YYYY/MM/DD', value: 'YYYY/MM/DD' },
	{ label: 'D/M/YY', value: 'D/M/YY' },
	{ label: 'M/D/YY', value: 'M/D/YY' },
	{ label: 'DD/MM/YYYY', value: 'DD/MM/YYYY' },
	{ label: 'DD.MM.YYYY', value: 'DD.MM.YYYY' },
	{ label: 'DD-MM-YYYY', value: 'DD-MM-YYYY' },
	{ label: 'YYYY.MM.DD', value: 'YYYY.MM.DD' },
	{ label: 'YYYY-MM-DD', value: 'YYYY-MM-DD' },
	{ label: 'YY/MM/DD', value: 'YY/MM/DD' },
	{ label: 'YY-M-D', value: 'YY-M-D' },
	{ label: 'YY.M.D', value: 'YY.M.D' }
];

const { t } = useI18n();

const imagesUploadFormatGuard = createImagesUploadV1FormatGuard(t);

const themeOptionsRef = ref(themeOptions);

const imgContentMode = ref<SelectorProps[]>([
	{
		label: t('Fill'),
		value: IMG_CONTENT_MODE.Fill
	},
	{
		label: t('Stretch'),
		value: IMG_CONTENT_MODE.Stretch
	},
	{
		label: t('Tile'),
		value: IMG_CONTENT_MODE.Tile
	}
]);

const ok = async (response: any) => {
	console.log('ok', response);
	if (response && response.code == 200) {
		if (selectBackgroundMode.value == BackgroundMode.desktop) {
			backgroundStore.upload_desktop_background(response.data.imageUrl);
		} else {
			backgroundStore.upload_login_background(response.data.imageUrl);
		}
	} else {
		notifyFailed(response.message);
	}
};

const setDesktopStyleMode = (mode: IMG_CONTENT_MODE) => {
	backgroundStore.set_desktop_style_mode(mode);
};

const setLoginStyleMode = (mode: IMG_CONTENT_MODE) => {
	backgroundStore.set_login_style_mode(mode);
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
	const isWidgetLoaded = await widgetPrefsStore.getWidget();
	if (!isWidgetLoaded) {
		console.error('Failed to load widget preferences, using local cache.');
	}
	updateCurrentDateTime();
	dateTimeTimer = setInterval(updateCurrentDateTime, 1000);
});

onUnmounted(() => {
	if (dateTimeTimer) {
		clearInterval(dateTimeTimer);
	}
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

const uploadDesktopBackgrounds = computed(() => {
	if (selectBackgroundMode.value == BackgroundMode.desktop) {
		return backgroundStore.wallpaper.upload_desktop_backgrounds;
	} else {
		return 0;
	}
});

const uploadLoginBackgrounds = computed(() => {
	if (selectBackgroundMode.value == BackgroundMode.login) {
		return backgroundStore.wallpaper.upload_login_backgrounds;
	} else {
		return 0;
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

	.images-list-bg {
		width: 100%;
		grid-column-gap: 8px;
		grid-row-gap: 8px;
		padding: 20px;
	}
}

.upload-image-btn {
	border-radius: 8px;
	border: 1px solid $btn-stroke;
	padding: 8px 12px;
}

.radio-class {
	margin-top: 8px;
}

.select-radio-bg {
	padding: 20px;
}
</style>
