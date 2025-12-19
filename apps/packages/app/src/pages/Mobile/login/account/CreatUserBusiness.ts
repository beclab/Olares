import { DIDKey, getDID } from '../../../../did/did-key';
import { UserItem } from '@didvault/sdk/src/core';
import { app } from '../../../../globals';
import { useUserStore } from '../../../../stores/user';
import { getAppPlatform } from 'src/application/platform';
import { i18n } from '../../../../boot/i18n';

export async function createUser() {
	try {
		const userStore = useUserStore();
		if (!(await userStore.importUserPrecheck())) {
			return;
		}
		const mnemonic = await DIDKey.generate();
		if (!mnemonic) {
			throw new Error(i18n.global.t('errors.mnemonic_generate_failure')); //mnemonic generate failure
		}
		const did = await getDID(mnemonic);
		if (!did) {
			throw new Error(i18n.global.t('errors.mnemonic_generate_failure'));
		}

		if (userStore.current_id) {
			const current_user: UserItem = userStore.users!.items.get(
				userStore.current_id
			)!;
			if (current_user && current_user.url) {
				await app.lock();
			}
		}

		const user = await userStore.importUser(did, '', mnemonic);
		if (!user) {
			throw new Error(i18n.global.t('errors.add_user_failed'));
		}
		if (user) {
			await userStore.setCurrentID(user.id);
			await app.load(user.id, getAppPlatform().reconfigAppStateDefaultValue);
			await app.new(user.id, mnemonic);
		}
		return { data: userStore.current_id };
	} catch (e) {
		console.error(e.message);
		return { message: e.message };
	}
}
