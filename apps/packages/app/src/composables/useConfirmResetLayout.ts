import { ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { useQuasar } from 'quasar';
import { resetLayout as apiResetLayout } from 'src/api/desktop-layout';
import ConfirmDialog from '@apps/desktop/components/ConfirmDialog2.vue';
import { notifyFailed, notifySuccess } from 'src/utils/settings/btNotify';
import { i18n } from 'src/boot/i18n';

export function useConfirmResetLayout() {
	const { t } = useI18n();
	const $q = useQuasar();
	const resetLoading = ref(false);

	function confirmReset() {
		if (resetLoading.value) {
			return;
		}
		$q.dialog({
			component: ConfirmDialog,
			componentProps: {
				title: t('desktop_layout.reset_confirm_title'),
				warningMessage: t('desktop_layout.reset_confirm_warning'),
				noteMessage: t('desktop_layout.reset_confirm_note'),
				confirmText: t('desktop_layout.reset_layout_button'),
				cancelText: t('buttons.cancel'),
				showCancel: true
			}
		}).onOk(async () => {
			resetLoading.value = true;
			try {
				await apiResetLayout();
				notifySuccess(i18n.global.t('success'));
			} catch (e) {
				console.error('[desktop] reset failed', e);
				notifyFailed(i18n.global.t('desktop_layout.reset_failed'));
			} finally {
				resetLoading.value = false;
			}
		});
	}

	return {
		resetLoading,
		confirmReset
	};
}
