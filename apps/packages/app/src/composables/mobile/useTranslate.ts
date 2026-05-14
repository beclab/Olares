import { computed, reactive, ref, toRefs } from 'vue';
import {
	GLOBLA_RULE,
	MSG_SAVE_RULE,
	MSG_TRANS_GETRULE,
	MSG_TRANS_PUTRULE,
	MSG_TRANS_TOGGLE,
	OPT_TRANS_ALL,
	OPT_TRANS_OLARES,
	OPT_LANGS_FROM,
	OPT_LANGS_TO
} from 'src/platform/interface/bex/translate/config';
import {
	getCurTab,
	sendBgMsg,
	sendTabMsg
} from 'src/platform/interface/bex/translate/libs/msg';
import { kissLog } from 'src/platform/interface/bex/translate/libs/log';
import {
	saveRule,
	matchRule
} from 'src/platform/interface/bex/translate/libs/rules';
import { getSettingWithDefault } from 'src/platform/interface/bex/translate/libs/storage';
import { BtNotify, NotifyDefinedType } from '@bytetrade/ui';
import { useI18n } from 'vue-i18n';
import { useAppAbilitiesStore } from 'src/stores/appAbilities';
import { useUserStore } from 'src/stores/user';

export function useTranslate() {
	const { t } = useI18n();
	const translatorFormat = OPT_TRANS_ALL.map((item) => ({
		label: item,
		value: item
	}));

	const options = computed(() => {
		const userStore = useUserStore();
		const appAbilitiesStore = useAppAbilitiesStore();

		const translatorList = userStore.current_user?.isLargeVersion12
			? translatorFormat
			: translatorFormat.filter((item) => item.value !== OPT_TRANS_OLARES);

		if (!appAbilitiesStore.translate.running) {
			return translatorList.map((item) => ({
				...item,
				disable: item.value === OPT_TRANS_OLARES
			}));
		}
		return [...translatorList];
	});

	const rule = reactive({
		transOpen: false,
		translator: GLOBLA_RULE.translator,
		fromLang: GLOBLA_RULE.fromLang,
		toLang: GLOBLA_RULE.toLang,
		transOnly: GLOBLA_RULE.transOnly,
		skipLangs: (GLOBLA_RULE.skipLangs || []) as string[]
	});
	const transOpen = ref(false);
	const translator = ref();
	const loading = ref(false);
	const contentScriptReady = ref(false);

	const checkContentScriptReady = async () => {
		try {
			await sendTabMsg(MSG_TRANS_GETRULE);
			contentScriptReady.value = true;
			return true;
		} catch (err) {
			contentScriptReady.value = false;
			return false;
		}
	};

	const handleTransToggle = async () => {
		if (!contentScriptReady.value) {
			const isReady = await checkContentScriptReady();
			if (!isReady) {
				BtNotify.show({
					type: NotifyDefinedType.WARNING,
					message: t('bex.content_script_not_ready')
				});
				return;
			}
		}

		transOpen.value = !transOpen.value;

		try {
			await sendTabMsg(MSG_TRANS_TOGGLE);
		} catch (err) {
			kissLog(err, 'sendTabMsg failed');
			contentScriptReady.value = false;
			BtNotify.show({
				type: NotifyDefinedType.WARNING,
				message: t('bex.translate_failed')
			});
		}
	};

	const handleSaveRule = async () => {
		const isExt = false;
		const tran = false;
		try {
			let href = window.location.href;
			if (!tran) {
				const tab = await getCurTab();
				href = tab.url || '';
			}

			const setting = await getSettingWithDefault();
			const matchedRule = await matchRule(href, setting);

			const newRule = {
				...matchedRule,
				...rule,
				transOpen: JSON.stringify(rule.transOpen),
				pattern:
					matchedRule.pattern && matchedRule.pattern !== '*'
						? matchedRule.pattern
						: href.split('/')[2]
			};
			if (isExt && tran) {
				sendBgMsg(MSG_SAVE_RULE, newRule);
			} else {
				const status = await saveRule(newRule);
			}
		} catch (err) {
			kissLog(err, 'save rule');
			BtNotify.show({
				type: NotifyDefinedType.FAILED,
				message: t('bex.save_failed')
			});
		}
	};

	const getTransRule = async (onlyQuery = true) => {
		if (loading.value) {
			return;
		}

		loading.value = true;
		try {
			const tab = await getCurTab();
			if (!tab) {
				kissLog('No active tab found', 'getTransRule');
				return;
			}

			const currentUrl = tab.url || window.location.href;

			const setting = await getSettingWithDefault();

			const matchedRule = await matchRule(currentUrl, setting);

			rule.transOpen = matchedRule.transOpen === 'true';
			rule.translator = matchedRule.translator;
			rule.fromLang = matchedRule.fromLang || GLOBLA_RULE.fromLang;
			rule.toLang = matchedRule.toLang || GLOBLA_RULE.toLang;
			rule.transOnly = matchedRule.transOnly || GLOBLA_RULE.transOnly;
			rule.skipLangs = (matchedRule.skipLangs ||
				GLOBLA_RULE.skipLangs ||
				[]) as string[];

			let actualTransOpen: boolean | null = null;
			try {
				const response = await sendTabMsg(MSG_TRANS_GETRULE);
				contentScriptReady.value = true;
				if (response?.data?.transOpen !== undefined) {
					actualTransOpen = response.data.transOpen === 'true';
				}
			} catch (err) {
				kissLog(err, 'Failed to get actual translation state');
				contentScriptReady.value = false;
			}

			if (!onlyQuery) {
				const ruleToSend = {
					...matchedRule,
					transOpen:
						actualTransOpen !== null
							? String(actualTransOpen)
							: matchedRule.transOpen
				};
				try {
					await sendTabMsg(MSG_TRANS_PUTRULE, ruleToSend);
					contentScriptReady.value = true;
				} catch (tabError) {
					kissLog(tabError, 'sendTabMsg to content script');
					contentScriptReady.value = false;
				}
			}

			if (actualTransOpen !== null) {
				transOpen.value = actualTransOpen;
			} else {
				transOpen.value = matchedRule.transOpen === 'true';
			}
		} catch (err) {
			kissLog(err, 'query rule');
		} finally {
			loading.value = false;
		}
	};

	const translateHandler = async () => {
		if (transOpen.value !== rule.transOpen) {
			handleTransToggle();
		}
		await sendTabMsg(MSG_TRANS_PUTRULE, {
			transOpen: JSON.stringify(rule.transOpen)
		});
		handleSaveRule();
	};

	const translateHandler2 = async () => {
		await sendTabMsg(MSG_TRANS_PUTRULE, { translator: rule.translator });
		handleSaveRule();
	};

	const handleFieldChange = async (field: string, value: any) => {
		(rule as any)[field] = value;
		await sendTabMsg(MSG_TRANS_PUTRULE, { [field]: value });
		handleSaveRule();
	};

	const handleTransOnlyChange = async (value: boolean) => {
		const newValue = value ? 'true' : 'false';
		rule.transOnly = newValue;

		await sendTabMsg(MSG_TRANS_PUTRULE, {
			transOnly: newValue
		});
		handleSaveRule();
	};

	const fromLangOptions = computed(() => {
		const isOlares = rule.translator === OPT_TRANS_OLARES;
		const enabledLangs = ['auto', 'en', 'zh-CN'];

		return OPT_LANGS_FROM.map(([value, label]) => ({
			value,
			label,
			disable: isOlares && !enabledLangs.includes(value)
		}));
	});

	const toLangOptions = computed(() => {
		const isOlares = rule.translator === OPT_TRANS_OLARES;
		const enabledLangs = ['en', 'zh-CN'];

		return OPT_LANGS_TO.map(([value, label]) => ({
			value,
			label,
			disable: isOlares && !enabledLangs.includes(value)
		}));
	});

	return {
		handleTransToggle,
		translator,
		options,
		translateHandler2,
		transOpen,
		translateHandler,
		loading,
		getTransRule,
		rule,
		handleFieldChange,
		handleTransOnlyChange,
		fromLangOptions,
		toLangOptions,
		contentScriptReady,
		checkContentScriptReady
	};
}
