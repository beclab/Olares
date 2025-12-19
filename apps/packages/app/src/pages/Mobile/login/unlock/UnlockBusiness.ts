import { UserItem, MnemonicItem } from '@didvault/sdk/src/core';
import { app, setSenderUrl } from '../../../../globals';
import { useUserStore } from '../../../../stores/user';
import { BusinessAsyncCallback } from '../../../../utils/Callback';
import { sendUnlock } from '../../../../utils/bexFront';
import { PASSWORD_RULE } from '../../../../utils/constants';
import { i18n } from '../../../../boot/i18n';

export async function savePassword(
	password: string,
	callback: BusinessAsyncCallback
) {
	try {
		if (!password) {
			throw new Error(i18n.global.t('password_not_empty'));
		}

		const allRule = new RegExp(PASSWORD_RULE.ALL_RULE);

		if (!allRule.test(password)) {
			throw new Error(i18n.global.t('password_not_meet_rules'));
		}

		const userStore = useUserStore();
		await userStore.create(password);
		userStore.password = password;
		sendUnlock();
		callback.onSuccess(password);
	} catch (e) {
		console.error(e);
		callback.onFailure(e.message);
	}
}

export async function unlockByPwd(
	password: string,
	callback: BusinessAsyncCallback
) {
	const userStore = useUserStore();
	try {
		if (!password) {
			throw new Error(i18n.global.t('password_not_empty'));
		}

		// await userStore.users!.unlock(password);
		// userStore.password = password;
		// await userStore.save();
		await userStore.unlock(password);

		if (userStore.current_id) {
			const user: UserItem = userStore.users!.items.get(userStore.current_id!)!;
			if (user.name) {
				setSenderUrl({
					url: user.vault_url
				});
			}

			const mnemonic: MnemonicItem = userStore.users!.mnemonics.get(
				userStore.current_id!
			)!;

			await app.load(userStore.current_id!);
			await app.unlock(mnemonic.mnemonic);
		}
		await callback.onSuccess(userStore.current_id);
	} catch (error) {
		userStore.users?.lock();
		callback.onFailure(i18n.global.t('password_error'));
	}
}
