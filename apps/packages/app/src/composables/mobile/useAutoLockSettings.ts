import { reactive, ref } from 'vue';
import { app } from 'src/globals';
import { i18n } from 'src/boot/i18n';

const customTimeValue = -1;
export const lockTimeOptions = [
	{ value: 10, label: i18n.global.t('10min') },
	{ value: 30, label: i18n.global.t('30min') },
	{ value: 60, label: i18n.global.t('1hour') },
	{ value: 4 * 60, label: i18n.global.t('4hours') },
	{ value: 24 * 60, label: i18n.global.t('1day') },
	{ value: customTimeValue, label: i18n.global.t('custom') }
];

export function useAutoLockSettings() {
	const settings = reactive({
		autoLock: app.settings.autoLock,
		lockTime: app.settings.autoLockDelay
	});
	let lockTimeSelectCache = lockTimeOptions[0].value;
	const defaultTime = lockTimeOptions.find(
		(item) => item.value === app.settings.autoLockDelay
	);
	const lockTimeSelect = ref(defaultTime ? defaultTime.value : customTimeValue);

	const changeAutoLock = (value: boolean) => {
		settings.autoLock = value;
		app.setSettings({ autoLock: value });
	};

	const changeAutoLockDelay = (value: number) => {
		if (value < 0) {
			settings.lockTime = lockTimeSelectCache;
		} else {
			lockTimeSelectCache = value;
			app.setSettings({ autoLockDelay: value });
		}
	};

	return {
		settings,
		changeAutoLock,
		changeAutoLockDelay,
		lockTimeSelect
	};
}
