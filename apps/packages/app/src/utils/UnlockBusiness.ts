import { app } from '../globals';
import { useUserStore, defaultPassword } from '../stores/user';
import { BusinessAsyncCallback } from './Callback';
import { sendUnlock } from './bexFront';
import { PASSWORD_RULE } from './constants';
import { i18n } from '../boot/i18n';
import { getAppPlatform } from '../application/platform';
import UnlockAppDialog from '../components/unlock/UnlockAppDialog.vue';

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

export async function saveDefaultPassword(callback: BusinessAsyncCallback) {
	const userStore = useUserStore();
	userStore.setPasswordResetedValue(false);
	await savePassword(defaultPassword, callback);
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

		await userStore.unlock(password);

		if (userStore.current_id) {
			await app.load(userStore.current_id);
		}
		await callback.onSuccess(userStore.current_id);
	} catch (error) {
		await userStore.users?.lock();
		callback.onFailure(i18n.global.t('password_error'));
	}
}

export async function unlockByDefaultPwd(callback: BusinessAsyncCallback) {
	const userStore = useUserStore();
	if (userStore.passwordReseted) {
		return;
	}
	await unlockByPwd(defaultPassword, callback);
	userStore.sendSetPassworNotify();
}

export const unlockUserFirstBusiness = async (props: any) => {
	if (!getAppPlatform().getQuasar()) {
		return false;
	}
	return new Promise<boolean>((resolve) => {
		getAppPlatform()
			.getQuasar()
			?.dialog({
				component: UnlockAppDialog,
				componentProps: props
			})
			.onOk(() => {
				resolve(true);
			})
			.onCancel(() => {
				resolve(false);
			});
	});
};
