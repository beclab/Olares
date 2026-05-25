import { supportLanguages, SupportLanguageType } from 'src/i18n';
import { CacheRequest } from 'src/stores/market/CacheRequest';
import globalConfig from 'src/api/market/config';
import { i18n } from 'src/boot/i18n';
import { defineStore } from 'pinia';
import {
	getSettingConfig,
	updateSettingConfig
} from 'src/api/market/private/setting';

export const useSettingStore = defineStore('setting', {
	state: () => ({
		initialized: false,
		reload: false,
		nsfw: false,
		currentLanguage: '' as SupportLanguageType,
		lastLanguage: '' as SupportLanguageType,
		marketSourceId: ''
	}),
	actions: {
		init() {
			const storedLang = localStorage.getItem(
				'language'
			) as SupportLanguageType | null;
			this.currentLanguage =
				storedLang && this.isValidLanguage(storedLang)
					? storedLang
					: this.getLanguage();
			this.languageUpdate(this.currentLanguage, false);
		},
		initConfigRequest() {
			return new CacheRequest('cache_market_setting', getSettingConfig, {
				onData: (data) => {
					this.marketSourceId = data.selected_source;
					this.nsfw = data.nsfw;
					this.initialized = true;
				}
			});
		},
		async setNsfw(status: boolean) {
			const result: any = await updateSettingConfig(
				status,
				this.marketSourceId
			);
			if (result) {
				this.nsfw = result.nsfw;
				this.reload = true;
			}
			return result;
		},
		async setMarketSourceId(sourceId: string) {
			const result: any = await updateSettingConfig(this.nsfw, sourceId);
			if (result) {
				this.marketSourceId = result.selected_source;
				this.reload = true;
			}
			return result;
		},
		languageUpdate(language: SupportLanguageType, save = true) {
			if (this.lastLanguage == language) {
				return;
			}

			if (!globalConfig.isOfficial) {
				return;
			}

			const languageItem = supportLanguages.find((e) => e.value == language);
			if (!languageItem || !language) {
				return;
			}

			i18n.global.locale.value = language;
			this.lastLanguage = language;
			if (save) {
				localStorage.setItem('language', language);
			}
		},
		getLanguage() {
			const lang = navigator.language;
			if (lang.startsWith('zh')) {
				return 'zh-CN';
			} else {
				return 'en-US';
			}
		},
		isValidLanguage(lang: string): boolean {
			return supportLanguages.some((item) => item.value === lang);
		}
	}
});
