import { bus } from './bus';
import { BtNotify, NotifyDefinedType } from '@bytetrade/ui';

export const showNotify = (
	message: string,
	type: NotifyDefinedType,
	group?: string
) => {
	BtNotify.show({
		group,
		type,
		message: message
	});
};

export const notifySuccess = (info = 'Success', group?: string) => {
	showNotify(info, NotifyDefinedType.SUCCESS, group);
};

export const notifyFailed = (info = 'Failed', group?: string) => {
	showNotify(info, NotifyDefinedType.FAILED, group);
};

export const notifyWarning = (info = 'Warning') => {
	showNotify(info, NotifyDefinedType.WARNING);
};

export const notifyWaitingShow = (info = 'Waiting', notify_id?: string) => {
	BtNotify.show({
		type: NotifyDefinedType.LOADING,
		message: info,
		closeTimeout: true,
		notify_id
	});
};

export const notifyHide = (notify_id?: string) => {
	BtNotify.hide({ notify_id });
};

export const notifyProgress = (info = 'Waiting', caption = '0%') => {
	BtNotify.show({
		type: NotifyDefinedType.PROGRESS,
		message: info,
		closeTimeout: true,
		caption
	});
};

const errorMessages = new Set();
let errorTimeout: any;

export const notifyRequestMessageError = (error: any) => {
	const message = getRequestErrorMessage(error);
	if (!errorMessages.has(message)) {
		errorMessages.add(message);
		notifyFailed(message);
		if (errorTimeout) clearTimeout(errorTimeout);
		errorTimeout = setTimeout(() => {
			errorMessages.clear();
		}, 5000);
	}
};

export const getRequestErrorMessage = (error: any) => {
	if (error && error.response) {
		const response = error.response;
		if (response.data && response.data.message) {
			return response.data.message;
		}
		// if (typeof response.data === 'string') {
		// 	return response.data;
		// }
	}
	return error.message ? error.message : error;
};

export const notifyEventBus = bus;
