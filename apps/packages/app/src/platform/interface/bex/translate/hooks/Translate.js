import { ref, watch } from 'vue';
import { tryDetectLang } from '../libs';
import { apiTranslate } from '../apis';
import { DEFAULT_TRANS_APIS } from '../config';
import { kissLog } from '../libs/log';

export function useTranslate(q, ruleDetaul, setting) {
	const text = ref('');
	const loading = ref(false);
	const sameLang = ref(false);
	const rule = ref(ruleDetaul);

	const translate = async () => {
		console.log('useTranslate', q, rule.value);
		try {
			loading.value = true;

			const {
				translator,
				fromLang,
				toLang,
				detectRemote,
				skipLangs = []
			} = rule.value;
			const { langDetector, transApis } = setting;

			if (!q.replace(/\[(\d+)\]/g, '').trim()) {
				text.value = q;
				sameLang.value = false;
				return;
			}

			const deLang = await tryDetectLang(
				q,
				detectRemote === 'true',
				langDetector
			);

			const normalizeLang = (lang) => {
				if (!lang) return '';
				if (lang === 'zh' || lang === 'zh-Hans') return 'zh-CN';
				if (lang === 'zh-Hant') return 'zh-TW';
				return lang;
			};

			const normalizedDeLang = normalizeLang(deLang);
			const normalizedToLang = normalizeLang(toLang);
			const normalizedSkipLangs = skipLangs.map(normalizeLang);

			if (
				deLang &&
				(normalizedDeLang === normalizedToLang ||
					normalizedSkipLangs.includes(normalizedDeLang))
			) {
				sameLang.value = true;
			} else {
				const [trText, isSame] = await apiTranslate({
					translator,
					text: q,
					fromLang,
					toLang,
					apiSetting: {
						...DEFAULT_TRANS_APIS[translator],
						...(transApis[translator] || {})
					}
				});
				text.value = trText;
				sameLang.value = isSame;
			}
		} catch (err) {
			kissLog(err, 'translate');
			loading.value = false;
		} finally {
			loading.value = false;
		}
	};

	translate();

	watch(
		[
			() => q,
			() => rule.value.translator,
			() => rule.value.fromLang,
			() => rule.value.toLang,
			() => rule.value.detectRemote,
			() => rule.value.skipLangs,
			() => setting.langDetector,
			() => setting.transApis
		],
		translate,
		{ deep: true }
	);

	return {
		text,
		sameLang,
		loading,
		rule
	};
}
