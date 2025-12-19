import { useI18n } from 'vue-i18n';
import { useUserStore } from 'src/stores/user';
import { useQuasar } from 'quasar';
import DialogResetPassword from 'src/pages/Electron/SettingsPage/DialogResetPassword.vue';

export const useTerminusSetSecurityPassword = () => {
	const { t } = useI18n();
	const $q = useQuasar();
	const userStore = useUserStore();

	const changePassword = async () => {
		if (!(await userStore.unlockFirst(undefined, { hide: true }))) {
			return;
		}
		$q.dialog({
			component: DialogResetPassword,
			componentProps: {
				title: userStore.passwordReseted
					? t('settings.changePassword')
					: t('Set up a password'),
				navigation: t('cancel')
			}
		});
	};

	return { t, userStore, changePassword };
};
