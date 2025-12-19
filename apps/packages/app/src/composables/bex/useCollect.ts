import { useCollectStore } from 'src/stores/collect';
import { RssInfo, RssStatus } from 'src/pages/Mobile/collect/utils';
import { useUserStore } from 'src/stores/user';
import { useQuasar } from 'quasar';
import { computed, ref } from 'vue';
import { openUrl, getCurrentTabInfo } from 'src/utils/bex/tabs';
import { useAppAbilitiesStore } from 'src/stores/appAbilities';
import { browser } from 'src/platform/interface/bex/browser/target';
import {
	URL_VALID_STATUS,
	UrlValidationResult,
	validateUrlWithReasonAsync
} from 'src/utils/url2';
import { COOKIE_LEVEL, CookieStatusCode } from 'src/utils/rss-types';
const validateDefault = { valid: false };

export function useCollect() {
	const collectStore = useCollectStore();
	const $q = useQuasar();
	const userStore = useUserStore();
	const appAbilitiesStore = useAppAbilitiesStore();
	const validate = ref<UrlValidationResult>({ ...validateDefault });

	const onSaveEntry = async (item: RssInfo) => {
		if (item.status !== RssStatus.none) {
			return;
		}
		$q.loading.show();
		await collectStore.addEntry(item);
		$q.loading.hide();
	};

	const wiseUrl = computed(() =>
		userStore.getModuleSever(
			appAbilitiesStore.wise.id,
			'https:',
			`/history/${item.value.id}`
		)
	);

	const openWise = () => {
		if (!wiseUrl.value) {
			return;
		}
		openUrl(wiseUrl.value);
	};

	async function setData(tab) {
		validate.value = await validateUrlWithReasonAsync(tab?.url);
		if (
			!validate.value.valid &&
			validate.value?.status !== URL_VALID_STATUS.BLOCKED
		) {
			return;
		}
		collectStore.setList([
			{
				title: tab.title,
				url: tab.url,
				image: tab.favIconUrl
			}
		]);
	}

	async function init() {
		const tab = await getCurrentTabInfo();
		setData(tab);
	}

	const item = computed(() => collectStore.pagesList[0]);

	const handleTabInfo = (tab) => {
		if (tab) {
			setData(tab);
		}
	};
	const handleActivated = async (activeInfo) => {
		const tab = await browser.tabs.get(activeInfo.tabId);
		handleTabInfo(tab);
	};

	const handleUpdated = (tabId, changeInfo, tab) => {
		if (changeInfo.url || changeInfo.title) {
			handleTabInfo(tab);
		}
	};

	const cookieStatusCode = computed(() => {
		try {
			if (!item.value.cookie?.cookie_exist) {
				return CookieStatusCode.COOKIE_NOT_UPLOADED;
			} else if (item.value.cookie?.cookieExpired) {
				return CookieStatusCode.COOKIE_EXPIRED;
			} else {
				return CookieStatusCode.COOKIE_UPLOADED;
			}
		} catch (error) {
			return CookieStatusCode.COOKIE_NOT_UPLOADED;
		}
	});

	return {
		collectStore,
		RssStatus,
		onSaveEntry,
		openWise,
		item,
		setData,
		init,
		handleActivated,
		handleUpdated,
		appAbilitiesStore,
		validate,
		cookieStatusCode
	};
}
