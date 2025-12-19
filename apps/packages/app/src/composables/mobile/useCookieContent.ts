import { ref, computed, reactive } from 'vue';
import { useI18n } from 'vue-i18n';
import { useQuasar, date } from 'quasar';
import api from 'axios';
import parser from 'tld-extract';
import { browser } from 'src/platform/interface/bex/browser/target';
import { useUserStore } from '../../stores/user';
import { notifyFailed, notifySuccess } from '../../utils/notifyRedefinedUtil';

export function useCookieContent() {
	const cookieTooltipOption = 'cookieTooltipOption';
	const { t } = useI18n();
	const userStore = useUserStore();
	const $q = useQuasar();

	const message = reactive({
		upload: {
			title: t('cookie_upload.title'),
			content: t('cookie_upload.content'),
			footer: t('cookie_upload.footer'),
			type: 'info',
			show: true
		},
		overwrite: {
			title: t('cookie_overwrite.title'),
			content: t('cookie_overwrite.content'),
			footer: t('cookie_overwrite.footer'),
			type: 'negative',
			show: true
		},
		no_data: {
			title: t('cookie_empty.title'),
			content: t('cookie_empty.content'),
			footer: t('cookie_empty.footer'),
			type: 'negative'
		}
	});

	const options = [
		{ label: t('cookie_action.upload_olares'), value: 'upload' },
		{ label: t('cookie_action.overwrite_browser'), value: 'overwrite' }
	];

	// 初始化本地存储的消息配置
	const initLocalMessage = () => {
		const optionsLocal = localStorage.getItem(cookieTooltipOption);
		if (optionsLocal) {
			const data = JSON.parse(optionsLocal);
			for (const key in message) {
				if (data[key]) {
					message[key] = { ...message[key], show: data[key].show };
				}
			}
		}
	};

	const cookiesList = ref<any>();
	const domain = ref();
	const loading = ref(false);
	const pushLoading = ref(false);
	const uploaded = ref(false);
	const model = ref(options[0]);
	const auto_sync = ref(false);
	const uploadTime = ref();
	const auto_sync_step = ref(10 * 60 * 1000);
	let timer: undefined | NodeJS.Timeout = undefined;

	const current_message = computed(() => message[model.value.value]);
	const hasCookie = computed(
		() => cookiesList.value && cookiesList.value.length > 0
	);

	const messageHandler = () => {
		message[model.value.value] = { ...message[model.value.value], show: false };
		localStorage.setItem(cookieTooltipOption, JSON.stringify(message));
	};

	function getPrimaryDomain(url: string) {
		const urlParser = parser(url);
		return urlParser.domain;
	}

	async function getAllCookies() {
		const [tab] = await browser.tabs.query({
			active: true,
			currentWindow: true
		});
		if (tab?.url) {
			try {
				const url = new URL(tab.url);
				domain.value = url.hostname;
				const rootDomain = getPrimaryDomain(url.origin);
				const cookies = await browser.cookies.getAll({ domain: rootDomain });
				const cookieIgnores = ['auth_token', 'authelia_session'];
				console.log('cookies', cookies);
				const selfDomain = userStore.current_user?.name.split('@').join('.');
				const isOlaresWeb = !selfDomain ? false : tab?.url.includes(selfDomain);
				const urlFilterStr = `.${url.hostname}`;
				const data = cookies.filter((item) => {
					if (isOlaresWeb) {
						return (
							urlFilterStr.includes(item.domain) &&
							!cookieIgnores.includes(item.name)
						);
					} else {
						return urlFilterStr.includes(item.domain);
					}
				});
				const allDomains = data.map((item) => item.domain);
				const uniqueNames = [
					...new Set(allDomains)
				] as unknown as Array<string>;
				const cookieAll = uniqueNames.map((item) => ({
					domain: item,
					records: data.filter((cookie) => cookie.domain === item)
				}));
				const cookieArr = cookieAll.filter((item) => !!item.records.length);
				cookieArr.sort((a, b) => b.domain.length - a.domain.length);
				cookiesList.value = cookieArr;
			} catch {
				// ignore
			}
		}
	}

	const pushCookie = async (showLoading = true) => {
		bexPushCookie(showLoading);
	};

	const bexPushCookie = async (showLoading = true) => {
		if (showLoading) {
			pushLoading.value = true;
		}

		const baseURL = userStore.getModuleSever('settings');
		const host = process.env.NODE_ENV == 'production' ? baseURL : '';

		try {
			await Promise.all(
				cookiesList.value.map((item) => api.post(host + '/api/cookie', item))
			);
			const timeStamp = Date.now();
			uploadTime.value = date.formatDate(timeStamp, 'HH:mm');
			if (showLoading) {
				notifySuccess(
					t('cookie_notify.uploaded_successfully'),
					'collection-cookie'
				);
			}
		} catch (error) {
			notifyFailed(t('cookie_notify.uploaded_failed'), 'collection-cookie');
		}
		pushLoading.value = false;
		uploaded.value = true;
	};

	const autoSyncHandler = () => {
		if (auto_sync.value) {
			getAllCookies();
			pushCookie(false);
			timer = setInterval(() => {
				getAllCookies();
				pushCookie(false);
			}, auto_sync_step.value);
		} else {
			timer && clearInterval(timer);
		}
	};

	return {
		message,
		options,
		cookiesList,
		domain,
		loading,
		pushLoading,
		uploaded,
		model,
		auto_sync,
		uploadTime,
		current_message,
		hasCookie,
		messageHandler,
		getAllCookies,
		pushCookie,
		autoSyncHandler,
		initLocalMessage
	};
}
