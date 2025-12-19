import { busOn } from 'src/utils/bus';

import { QVueGlobals } from 'quasar';
import { delay } from 'src/utils/utils';

export const registerFilePreviewEvent = async ($q?: QVueGlobals) => {
	if (!$q) {
		return;
	}
	let FilePreViewDialog: any = undefined;
	let FilePreVideoDialog: any = undefined;
	let FileMobilePreVideoDialog: any = undefined;
	let FileMobilePreviewPage: any = undefined;
	let previewInit = false;

	busOn('filesPreviewDisplay', async (isVideo: boolean, origin_id: number) => {
		while (!previewInit) {
			await delay(500);
		}
		if (isVideo) {
			$q.dialog({
				component:
					process.env.PLATFORM === 'MOBILE'
						? FileMobilePreVideoDialog
						: FilePreVideoDialog,
				componentProps: {
					origin_id
				}
			});
		} else {
			$q.dialog({
				component:
					process.env.PLATFORM === 'MOBILE'
						? FileMobilePreviewPage
						: FilePreViewDialog,
				componentProps: {
					origin_id
				}
			});
		}
	});
	if (!previewInit) {
		FilePreViewDialog = (
			await import('../../pages/Files/preview/FilePreViewDialog.vue')
		).default;
		FilePreVideoDialog = (
			await import('../../pages/Files/preview/FilePreVideoDialog.vue')
		).default;

		FileMobilePreVideoDialog = (
			await import('../../pages/Mobile/file/FilePreVideoDialog.vue')
		).default;

		FileMobilePreviewPage = (
			await import('../../pages/Mobile/file/FilePreviewPage.vue')
		).default;
		previewInit = true;
	}
};
