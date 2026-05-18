import { watch } from 'vue';
import { useI18n } from 'vue-i18n';

export function useLocaleWatch(callback: () => void) {
	const { locale } = useI18n();
	watch(locale, callback);
}
