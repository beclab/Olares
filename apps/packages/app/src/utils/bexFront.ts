import { stringToBase64 } from '@didvault/sdk/src/core';
import { useBexStore } from '../stores/bex';
import { useUserStore } from '../stores/user';
import { busEmit } from './bus';
// import { sendExtensionMessage } from '../extension/autofill2/utils/sendMessage';
export const sendUnlock = () => {
	if (process.env.PLATFORM !== 'BEX') {
		return;
	}
	const bex = useBexStore();
	const user = useUserStore();

	bex.controller.sendUnlocked(stringToBase64(user.password as string));
};

export const sendLock = () => {
	if (process.env.PLATFORM !== 'BEX') {
		return;
	}
	const bex = useBexStore();
	bex.controller.sendLocked();
};

export const bexVaultUpdate = () => {
	if (process.env.PLATFORM !== 'BEX') {
		return;
	}
	busEmit('updateVaultComplete');
};

export const autofillById = (id: string) => {
	if (process.env.PLATFORM !== 'BEX') {
		return;
	}
	const bex = useBexStore();
	bex.controller.autofillById(id);
};
