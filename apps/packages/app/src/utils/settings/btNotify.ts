import { BtNotify, NotifyDefinedType } from '@bytetrade/ui';

export const showNotify = (message: string, type: NotifyDefinedType) => {
	BtNotify.show({
		type,
		message: message
	});
};

export const notifySuccess = (info = 'Success') => {
	showNotify(info, NotifyDefinedType.SUCCESS);
};

export const notifyFailed = (info = 'Failed') => {
	showNotify(info, NotifyDefinedType.FAILED);
};

export const notifyWarning = (info = 'Warning') => {
	showNotify(info, NotifyDefinedType.WARNING);
};
