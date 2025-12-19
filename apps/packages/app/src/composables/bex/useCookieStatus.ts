import { computed } from 'vue';
import { useI18n } from 'vue-i18n';
import cookieUploadIconLight from 'src/assets/plugin/cookie-upload.svg';
import cookieExpiredfromIconLight from 'src/assets/plugin/cookie-expired-light.svg';
import cookieUploadedIconLight from 'src/assets/plugin/cookie-uploaded-white.svg';
import cookieUploadIconDark from 'src/assets/plugin/cookie-upload-dark.svg';
import cookieExpiredfromIconDark from 'src/assets/plugin/cookie-expired-dark.svg';
import cookieUploadedIconDark from 'src/assets/plugin/cookie-uploaded-dark.svg';
import { COOKIE_LEVEL, CookieStatusCode } from 'src/utils/rss-types';
import { useBrowserCookieStore } from 'src/stores/settings/browserCookie';
import { useCollectSiteStore } from 'src/stores/collect-site';
import { replaceOriginDomain } from 'src/utils/url2';
import { getApplication } from 'src/application/base';
import { useUserStore } from 'src/stores/user';
import { openUrl } from 'src/utils/bex/tabs';
import { useCollect } from 'src/composables/bex/useCollect';

export function useCookieStatus() {
	const { t } = useI18n();
	const browserCookieStore = useBrowserCookieStore();
	const collectSiteStore = useCollectSiteStore();

	const { validate } = useCollect();

	const cookieUploadIcon = process.env.PLATFORM_BEX_ALL
		? cookieUploadIconLight
		: cookieUploadIconDark;

	const cookieUploadedIcon = process.env.PLATFORM_BEX_ALL
		? cookieUploadedIconLight
		: cookieUploadedIconDark;

	const cookieExpiredfromIcon = process.env.PLATFORM_BEX_ALL
		? cookieExpiredfromIconLight
		: cookieExpiredfromIconDark;

	const cookieStatusCode = computed(() => {
		try {
			if (collectSiteStore.cookie?.cookieRequire === COOKIE_LEVEL.REQUIRED) {
				if (!collectSiteStore.cookie?.cookieExist) {
					return CookieStatusCode.COOKIE_NOT_UPLOADED;
				} else if (collectSiteStore.cookie?.cookieExpired) {
					return CookieStatusCode.COOKIE_EXPIRED;
				} else {
					return CookieStatusCode.COOKIE_UPLOADED;
				}
			}
			return CookieStatusCode.COOKIE_NOT_UPLOADED;
		} catch (error) {
			return CookieStatusCode.COOKIE_NOT_UPLOADED;
		}
	});

	const cookieRequire = computed(
		() =>
			cookieStatusCode.value < CookieStatusCode.COOKIE_UPLOADED &&
			collectSiteStore.cookie?.cookieRequire === COOKIE_LEVEL.REQUIRED
	);

	const cookieIcon = computed(() => {
		const icons = [cookieUploadIcon, cookieExpiredfromIcon, cookieUploadedIcon];
		const tooltip = [t('upload_cookie_info'), t('cookie_expired_reupload'), ''];

		return {
			icon: icons[cookieStatusCode.value],
			tooltip: tooltip[cookieStatusCode.value]
		};
	});

	const ytdlpRequire = computed(() => {
		if (!!collectSiteStore.data.cookie.is_entry_available) {
			return {
				icon: '',
				tooltip: collectSiteStore.data.cookie.is_entry_available
			};
		}
		return null;
	});

	const pushLoading = computed(() => browserCookieStore.pushLoading);

	const openSettingsCookieManager = () => {
		let url = '';
		const settingsKey = 'settings';

		if (getApplication().platform && getApplication().platform?.isClient) {
			const userStore = useUserStore();
			url = userStore.getModuleSever(
				settingsKey,
				'https:',
				`/integration/cookie`
			);
		} else {
			const origin = replaceOriginDomain(location.origin, settingsKey, true);
			url = `${origin}/integration/cookie`;
		}

		openUrl(url);
	};

	const cookieHandler = () => {
		if (process.env.PLATFORM_BEX_ALL) {
			browserCookieStore.pushCookie();
		} else {
			openSettingsCookieManager();
		}
	};

	return {
		cookieStatusCode,
		cookieRequire,
		cookieIcon,
		pushLoading,
		cookieHandler,
		openSettingsCookieManager,
		collectSiteStore,
		ytdlpRequire
	};
}
