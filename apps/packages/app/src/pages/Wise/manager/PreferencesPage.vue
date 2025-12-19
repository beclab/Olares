<template>
	<div class="wise-page-root bg-color-white column justify-start">
		<title-bar>
			<template v-slot:before>
				<bt-breadcrumbs
					:title="t('main.preferences')"
					icon="sym_r_settings_applications"
					margin="44px"
				/>
			</template>
		</title-bar>
		<bt-scroll-area class="preferences-scroll-area">
			<div class="preferences-page column justify-start">
				<div class="text-ink-1 text-h6">{{ t('preferences.theme') }}</div>
				<bt-check-box
					:model-value="themeModel === THEME_TYPE.AUTO"
					@update:model-value="themeModelUpdate"
					style="padding: 0"
					class="q-mt-md"
					:label="t('settings.themes.follow_system_theme')"
				/>
				<div class="text-body3 text-ink-3">
					{{ t('preferences.follow_system_theme_desc') }}
				</div>
				<div class="q-mt-md row justify-between full-width">
					<theme-selector
						image="rss/theme/light.svg"
						:model="THEME_TYPE.LIGHT"
						v-model="themeModel"
						:label="t('settings.themes.light')"
					/>
					<theme-selector
						image="rss/theme/dark.svg"
						:model="THEME_TYPE.DARK"
						v-model="themeModel"
						:label="t('settings.themes.dark')"
					/>
				</div>

				<div class="text-ink-1 text-h6 q-mt-xl">
					{{ t('preferences.import_or_export') }}
				</div>
				<div class="text-body3 text-ink-3 q-mt-md">
					{{ t('main.feeds') }}
				</div>
				<div class="row justify-start q-mt-xs">
					<input
						ref="input"
						v-show="false"
						accept=".opml,.xml,.txt"
						type="file"
						@change="handleFileChange"
					/>
					<request-btn
						:label="t('preferences.import_feeds_opml')"
						:loading="importLoading"
						@request="onImport"
						class="q-mr-lg"
					/>

					<request-btn
						:label="t('preferences.export_feeds_opml')"
						:loading="exportLoading"
						@request="onExport"
					/>
				</div>

				<!--			<div class="text-body3 text-ink-3 q-mt-lg">-->
				<!--				{{ t('main.library') }}-->
				<!--			</div>-->
				<!--			<div class="row justify-start q-mt-xs">-->
				<!--				<div class="rss-click-button text-body3 text-orange-default q-mr-md">-->
				<!--					{{ t('preferences.import_library_csv') }}-->
				<!--				</div>-->
				<!--				<div class="rss-click-button text-body3 text-orange-default">-->
				<!--					{{ t('preferences.export_library_csv') }}-->
				<!--				</div>-->
				<!--			</div>-->

				<!--				<div class="text-ink-1 text-h6 q-mt-xl">-->
				<!--					{{ t('main.recommendations') }}-->
				<!--				</div>-->
				<!--				<bt-check-box-->
				<!--					:model-value="configStore.recommendationOpen"-->
				<!--					@update:model-value="recommendationUpdate"-->
				<!--					style="padding: 0"-->
				<!--					class="q-mt-md"-->
				<!--					:label="t('preferences.Enable local news recommendation (beta)')"-->
				<!--				/>-->
				<!--				<div class="text-body3 text-ink-3">-->
				<!--					{{-->
				<!--						t(-->
				<!--							'preferences.When enabled, Wise will provide trending news for you'-->
				<!--						)-->
				<!--					}}-->
				<!--				</div>-->

				<div class="text-ink-1 text-h6 q-mt-xl">
					{{ t('preferences.Upload settings') }}
				</div>
				<bt-check-box
					:model-value="configStore.uploadLinksOpen"
					@update:model-value="uploadLinksUpdate"
					style="padding: 0"
					class="q-mt-md"
					:label="t('preferences.Enable batch link upload')"
				/>
				<bt-check-box
					:model-value="configStore.uploadCookiesOpen"
					@update:model-value="uploadCookiesUpdate"
					style="padding: 0"
					:label="t('preferences.Enable batch cookie upload')"
				/>

				<div class="text-ink-1 text-h6 q-mt-xl">
					{{ t('preferences.storage') }}
				</div>
				<div class="text-body3 text-ink-3 q-mt-md">
					{{ t('preferences.clear_local_data_and_sync_remote_data') }}
				</div>
				<request-btn
					class="q-mt-xs"
					:label="t('preferences.resynchronize')"
					:loading="clearLoading"
					@request="onClear"
				/>

				<div class="text-ink-1 text-h6 q-mt-xl">
					{{ t('about') }}
				</div>
				<div class="text-ink-2 text-body2 q-mt-md">
					{{ t('preferences.current_version', { version: versionRef }) }}
				</div>
			</div>
		</bt-scroll-area>
	</div>
</template>

<script setup lang="ts">
import { exportFeedAsOpml, importFeedAsOpml } from '../../../api/wise';
import ThemeSelector from '../../../components/rss/ThemeSelector.vue';
import BtBreadcrumbs from '../../../components/base/BtBreadcrumbs.vue';
import BtCheckBox from '../../../components/rss/BtCheckBox.vue';
import RequestBtn from '../../../components/rss/RequestBtn.vue';
import TitleBar from '../../../components/rss/TitleBar.vue';
import { useConfigStore } from '../../../stores/rss-config';
import { BtNotify, NotifyDefinedType } from '@bytetrade/ui';
import { THEME_TYPE } from '../../../utils/rss-types';
import { useRssStore } from '../../../stores/rss';
import { useI18n } from 'vue-i18n';
import { ref, watch } from 'vue';
import { date } from 'quasar';
import { TERMINUS_ID } from '../../../utils/localStorageConstant';
import { sendMessageToWorker } from '../database/sqliteService';
import { useTerminusStore } from '../../../stores/terminus';

const { t } = useI18n();
const input = ref();
const versionRef = ref(process.env.APP_VERSION);
const configStore = useConfigStore();
const terminusStore = useTerminusStore();
const rssStore = useRssStore();
const themeModel = ref(configStore.themeSetting);
const themeModelUpdate = (status: boolean) => {
	if (status) {
		themeModel.value = THEME_TYPE.AUTO;
	}
};

watch(
	() => themeModel.value,
	() => {
		configStore.setThemeSetting(themeModel.value);
	}
);

const recommendationUpdate = (status: boolean) => {
	configStore.setRecommendationOpen(status);
};

const uploadLinksUpdate = (status: boolean) => {
	configStore.setUploadLinksOpen(status);
};

const uploadCookiesUpdate = (status: boolean) => {
	configStore.setUploadCookiesOpen(status);
};

const importLoading = ref(false);
const exportLoading = ref(false);
const handleFileChange = (event) => {
	const file = event.target.files[0];
	if (file) {
		const formData = new FormData();
		formData.append('file', file);
		importLoading.value = true;
		importFeedAsOpml(formData)
			.then(() => {
				rssStore.syncFeeds();
				BtNotify.show({
					type: NotifyDefinedType.SUCCESS,
					message: t('preferences.import_feeds_success')
				});
			})
			.catch(() => {
				BtNotify.show({
					type: NotifyDefinedType.FAILED,
					message: t('preferences.import_feeds_failed')
				});
			})
			.finally(() => {
				importLoading.value = false;
			});
	}
};

const onImport = () => {
	input.value.click();
};
const onExport = () => {
	exportLoading.value = true;
	exportFeedAsOpml()
		.then((blob) => {
			const url = window.URL.createObjectURL(blob);
			const link = document.createElement('a');
			link.href = url;
			const name = terminusStore.terminusInfo
				? '_' + terminusStore.olaresId
				: '';
			link.download =
				date.formatDate(Date.now(), 'YYYY-MM-DD HH:mm:ss') +
				name +
				'_feed_list.opml';
			link.click();
			window.URL.revokeObjectURL(url);
		})
		.catch(() => {
			BtNotify.show({
				type: NotifyDefinedType.SUCCESS,
				message: t('preferences.export_feeds_failed')
			});
		})
		.finally(() => {
			exportLoading.value = false;
		});
};

const clearLoading = ref(false);
const onClear = () => {
	clearLoading.value = true;
	sendMessageToWorker('close')
		.then(() => {
			localStorage.removeItem(TERMINUS_ID);
			location.reload();
		})
		.catch((e: any) => {
			BtNotify.show({
				type: NotifyDefinedType.FAILED,
				message: `Clear error, please refresh manually. Error: ${e?.message}`
			});
		});
};
</script>

<style scoped lang="scss">
.preferences-scroll-area {
	width: 100%;
	height: calc(100vh - 56px);

	.preferences-page {
		padding: 20px 44px;
		width: 528px;
	}
}
</style>
