import { BtNotify, NotifyDefinedType } from '@bytetrade/ui';
import { i18n } from 'src/boot/i18n';
export const showNotify = (message: string, type: NotifyDefinedType) => {
	BtNotify.show({
		type,
		message: message
	});
};

export const notifySuccess = (info = i18n.global.t('success')) => {
	showNotify(info, NotifyDefinedType.SUCCESS);
};

export const notifyFailed = (info = i18n.global.t('failed')) => {
	showNotify(info, NotifyDefinedType.FAILED);
};

export const notifyWarning = (info = i18n.global.t('warning')) => {
	showNotify(info, NotifyDefinedType.WARNING);
};

export const notifyMessage = (info: string) => {
	if (!info) {
		return;
	}
	showNotify(info, NotifyDefinedType.MESSAGE);
};
