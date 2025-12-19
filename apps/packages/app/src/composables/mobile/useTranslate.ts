import { computed, reactive, ref, toRefs } from 'vue';
import {
	GLOBLA_RULE,
	MSG_SAVE_RULE,
	MSG_TRANS_GETRULE,
	MSG_TRANS_PUTRULE,
	MSG_TRANS_TOGGLE,
	OPT_TRANS_ALL,
	OPT_TRANS_OLARES
} from 'src/platform/interface/bex/translate/config';
import {
	getCurTab,
	sendBgMsg,
	sendTabMsg
} from 'src/platform/interface/bex/translate/libs/msg';
import { kissLog } from 'src/platform/interface/bex/translate/libs/log';
import { saveRule } from 'src/platform/interface/bex/translate/libs/rules';
import { BtNotify, NotifyDefinedType } from '@bytetrade/ui';
import { useI18n } from 'vue-i18n';
import { useAppAbilitiesStore } from 'src/stores/appAbilities';
import { useUserStore } from 'src/stores/user';

export function useTranslate() {
	const { t } = useI18n();
	const appAbilitiesStore = useAppAbilitiesStore();
	const translatorFormat = OPT_TRANS_ALL.map((item) => ({
		label: item,
		value: item
	}));

	const options = computed(() => {
		const userStore = useUserStore();

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
		translator: GLOBLA_RULE.translator
		// fromLang: '',
		// toLang: '',
		// textStyle: ''
	});
	const transOpen = ref(false);
	const translator = ref();
	const loading = ref(false);
	// const { fromLang, toLang, textStyle } = toRefs(rule);

	const handleTransToggle = async () => {
		console.log('handleTransToggle');
		transOpen.value = !transOpen.value;
		await sendTabMsg(MSG_TRANS_TOGGLE);
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
			const newRule = {
				...rule,
				transOpen: JSON.stringify(rule.transOpen),
				pattern: href.split('/')[2]
			};
			if (isExt && tran) {
				sendBgMsg(MSG_SAVE_RULE, newRule);
			} else {
				const status = await saveRule(newRule);
				BtNotify.show({
					type: NotifyDefinedType.SUCCESS,
					message: t('bex.save_successful')
				});
			}
		} catch (err) {
			kissLog(err, 'save rule');
			BtNotify.show({
				type: NotifyDefinedType.FAILED,
				message: t('bex.save_failed')
			});
		}
	};

	const getTransRule = async () => {
		loading.value = true;
		try {
			const res = await sendTabMsg(MSG_TRANS_GETRULE);
			if (!res.error) {
				rule.transOpen = res.data.transOpen === 'true';
				rule.translator = res.data.translator;
				transOpen.value = res.data.transOpen === 'true';
			}
		} catch (err) {
			kissLog(err, 'query rule');
		}
		loading.value = false;
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

	return {
		handleTransToggle,
		translator,
		options,
		translateHandler2,
		transOpen,
		translateHandler,
		loading,
		getTransRule,
		rule
	};
}
