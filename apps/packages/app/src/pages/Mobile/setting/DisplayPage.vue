<template>
	<terminus-title-bar :title="t('General')" />
	<TerminusScrollArea class="display-root">
		<template v-slot:content>
			<div class="q-mt-lg q-px-lg">
				<q-item class="q-pa-none">
					<q-item-section>
						<terminus-item :clickable="false" :borderRadius="12">
							<template v-slot:title>
								<div class="text-subtitle2 security-root__title">
									{{ t('language') }}
								</div>
							</template>
							<template v-slot:side>
								<bt-select
									v-model="currentLanguage"
									:options="supportLanguages"
									:offset="[40, 10]"
									@update:modelValue="updateLocale"
								/>
							</template>
						</terminus-item>
					</q-item-section>
				</q-item>

				<q-item class="q-pa-none q-mt-md">
					<q-item-section>
						<terminus-item :clickable="false" :borderRadius="12">
							<template v-slot:title>
								<div class="text-subtitle2 security-root__title">
									{{ t('settings.themes.follow_system_theme') }}
								</div>
							</template>
							<template v-slot:side>
								<bt-switch
									size="sm"
									truthy-track-color="light-blue-default"
									v-model="isThemeAuto"
									@update:model-value="changeAutoTheme"
								/>
							</template>
						</terminus-item>
						<div class="q-mt-sm text-ink-3">
							{{
								t(
									"After being selected, LarePass will follow the device's system settings to switch theme modes"
								)
							}}
						</div>
						<div
							class="q-mt-xl row items-center justify-center theme-select"
							:style="isThemeAuto ? 'opacity: 0.4' : ''"
						>
							<div
								class="theme-item-common q-mr-xl"
								@click="updateTheme(ThemeDefinedMode.LIGHT)"
							>
								<div
									class="image-bg"
									:class="
										isThemeLight ? 'theme-item-select' : 'theme-item-normal'
									"
								>
									<q-img
										src="../../../assets/setting/mobile-theme-light.svg"
										class="image"
									/>
								</div>
								<div class="content row items-center justify-center q-pl-md">
									<q-radio
										dense
										v-model="deviceStore.theme"
										:val="ThemeDefinedMode.LIGHT"
										:label="t('settings.themes.light')"
										color="yellow-default"
										@update:model-value="updateTheme(ThemeDefinedMode.LIGHT)"
									/>
								</div>
							</div>
							<div
								class="theme-item-common"
								@click="updateTheme(ThemeDefinedMode.DARK)"
							>
								<div
									:class="
										isThemeDark ? 'theme-item-select' : 'theme-item-normal'
									"
									class="image-bg"
								>
									<q-img
										src="../../../assets/setting/mobile-theme-dark.svg"
										class="image"
									/>
								</div>
								<div class="content row items-center justify-center q-pl-md">
									<q-radio
										v-model="deviceStore.theme"
										:val="ThemeDefinedMode.DARK"
										:label="t('settings.themes.dark')"
										color="yellow-default"
										dense
										@update:model-value="updateTheme(ThemeDefinedMode.DARK)"
									/>
								</div>
							</div>
						</div>
					</q-item-section>
				</q-item>
				<terminus-item
					v-if="!isBex"
					class="q-mt-md"
					:clickable="false"
					icon-name="sym_r_wifi"
					:wholePictureSize="20"
				>
					<template v-slot:title>
						<div class="text-subtitle2 security-root__title">
							{{ t('Only transfer files over wifi') }}
						</div>
					</template>
					<template v-slot:side>
						<bt-switch
							size="sm"
							truthy-track-color="light-blue-default"
							v-model="transferOnlyWifiStatus"
							@update:model-value="updateTransferOnlyWifiStatus"
						/>
					</template>
				</terminus-item>

				<terminus-item
					class="q-mt-md"
					icon-name="sym_r_book_4"
					:wholePictureSize="20"
					@click="openPrivacyPolicy"
				>
					<template v-slot:title>
						<div class="text-subtitle1">
							{{ t('Privacy Policy') }}
						</div>
					</template>
					<template v-slot:side>
						<div class="row items-center justify-end">
							<q-icon
								name="sym_r_keyboard_arrow_right"
								size="20px"
								color="ink-3"
							/>
						</div>
					</template>
				</terminus-item>

				<terminus-item
					class="q-mt-md"
					icon-name="sym_r_handshake"
					:wholePictureSize="20"
					@click="openServiceAgreement"
				>
					<template v-slot:title>
						<div class="text-subtitle1">
							{{ t('Service Agreement') }}
						</div>
					</template>
					<template v-slot:side>
						<div class="row items-center justify-end">
							<q-icon
								name="sym_r_keyboard_arrow_right"
								size="20px"
								color="ink-3"
							/>
						</div>
					</template>
				</terminus-item>

				<terminus-item
					img-bg-classes="bg-background-3"
					class="q-mt-md"
					icon-name="sym_r_ballot"
					:whole-picture-size="32"
					v-if="$q.platform.is.nativeMobile"
				>
					<template v-slot:title>
						<div class="text-subtitle1">
							{{ t('version') }}
						</div>
					</template>
					<template v-slot:side>
						<div class="row items-center justify-end">
							<div class="app-version text-body2 text-ink-2 q-pr-xs">
								{{ appVersion }}
							</div>
						</div>
					</template>
				</terminus-item>

				<q-item v-if="isBex" class="q-pa-none q-mt-md">
					<BexBadge></BexBadge>
				</q-item>
			</div>
		</template>
	</TerminusScrollArea>
</template>

<script lang="ts" setup>
import TerminusTitleBar from '../../../components/common/TerminusTitleBar.vue';
import { useI18n } from 'vue-i18n';
import { useUserStore } from '../../../stores/user';
import { supportLanguages, SupportLanguageType } from '../../../i18n';
import { computed, ref } from 'vue';
import { ThemeDefinedMode } from '@bytetrade/ui';
import { useDeviceStore } from '../../../stores/device';
import TerminusItem from '../../../components/common/TerminusItem.vue';
import BtSelect from '../../../components/base/BtSelect.vue';
import { i18n } from '../../../boot/i18n';
import BexBadge from '../../../components/common/BexBadge.vue';
import TerminusScrollArea from '../../../components/common/TerminusScrollArea.vue';
import { getPlatform } from '@didvault/sdk/src/core';

import { appServices } from '../../../utils/platform';
import { useQuasar } from 'quasar';

const { t } = useI18n();
const userStore = useUserStore();

const isBex = ref(process.env.IS_BEX || process.env.DEV_PLATFORM_BEX);

const deviceStore = useDeviceStore();

const $q = useQuasar();

const currentLanguage = ref(userStore.locale || i18n.global.locale.value);

const updateLocale = async (language: SupportLanguageType) => {
	if (language) {
		await userStore.updateLanguageLocale(language);
	}
};

const isThemeAuto = ref(deviceStore.theme == ThemeDefinedMode.AUTO);

const changeAutoTheme = (value: boolean) => {
	if (value) {
		updateTheme(ThemeDefinedMode.AUTO);
	} else {
		updateTheme(ThemeDefinedMode.LIGHT);
	}
};

const isThemeDark = computed(function () {
	return deviceStore.theme == ThemeDefinedMode.DARK;
});

const isThemeLight = computed(function () {
	return deviceStore.theme == ThemeDefinedMode.LIGHT;
});
const updateTheme = (theme: ThemeDefinedMode) => {
	if (theme != ThemeDefinedMode.AUTO) {
		isThemeAuto.value = false;
	}
	deviceStore.setTheme(theme);
};

const transferOnlyWifiStatus = ref<boolean>(userStore.transferOnlyWifi);

const updateTransferOnlyWifiStatus = async () => {
	userStore.updateTransferOnlyWifiStatus(!userStore.transferOnlyWifi);
	transferOnlyWifiStatus.value = userStore.transferOnlyWifi;
};

const openServiceAgreement = () => {
	window.open(appServices().serviceAgreement);
};

const openPrivacyPolicy = () => {
	window.open(appServices().privacyPolicy);
};

const appVersion = ref('');

const configVersion = async () => {
	if (!$q.platform.is.nativeMobile) {
		return;
	}
	appVersion.value = (await getPlatform().getDeviceInfo()).appVersion;
};
configVersion();
</script>

<style scoped lang="scss">
.display-root {
	width: 100%;
	height: calc(100% - 56px);

	.item-centent {
		border-radius: 12px;
		width: 100%;
		text-align: left;

		.checkbox-content {
			width: 100%;
			height: 30px;
			.checkbox-common {
				width: 16px;
				height: 16px;
				margin-right: 10px;
				border-radius: 4px;
			}

			.checkbox-unselect {
				border: 1px solid $separator-2;
			}

			.checkbox-selected-yellow {
				background: $yellow-default;
			}
			.checkbox-selected-green {
				background: $positive;
			}
		}
		.adminBtn {
			border: 1px solid $yellow;
			background-color: $yellow-1;
			display: inline-block;
			color: $ink-2;
			padding: 6px 12px;
			border-radius: 8px;
			cursor: pointer;

			&:hover {
				background-color: $yellow-3;
			}
		}

		.lock-slider {
			height: 60px;
			transition: height 0.5s;
			min-height: 0 !important;
			padding-top: 0px !important;
			padding-bottom: 0px !important;
		}

		.hideSlider {
			height: 0 !important;
		}
	}
	.theme-select {
		width: 100%;

		.theme-item-common {
			// height: 144px;
			width: min(calc(50% - 10px), 130px);

			overflow: hidden;
			.image {
				border-radius: 12px;
				width: 100%;
			}
			.theme-item-select {
				border: 2px solid $yellow-default;
			}

			.theme-item-normal {
				border: 2px solid transparent;
			}

			.image-bg {
				width: 100%;
				height: 100%;
				padding: 4px;
				border-radius: 16px;
			}
		}

		.content {
			width: 100%;
			height: 44px;
		}
	}
}

.module-sub-title {
	text-align: left;
	color: $prompt-message;
	text-transform: capitalize;
}
</style>
