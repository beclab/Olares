import { getApplication } from 'src/application/base';
import { t } from 'src/boot/i18n';
import { useAppAbilitiesStore } from 'src/stores/appAbilities';
import { useUserStore } from 'src/stores/user';
import { openUrl } from 'src/utils/bex/tabs';
import { computed } from 'vue';

export const useWiseAbility = () => {
	const userStore = useUserStore();
	const appAbilitiesStore = useAppAbilitiesStore();

	const openWiseInMarket = (appName?: string) => {
		const keyword = appName || appAbilitiesStore.wise.title;
		const url = userStore.getModuleSever(
			'market',
			'https:',
			`/search?keyword=${keyword}`
		);

		getApplication().openUrl(url);
	};

	const missingAbility = computed(() => {
		if (!appAbilitiesStore.wise.running) {
			return {
				app: appAbilitiesStore.wise,
				message: t('bex.wise_not_install')
			};
		}
		return null;
	});

	return {
		openWiseInMarket,
		appAbilitiesStore,
		missingAbility
	};
};
