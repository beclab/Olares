import { getAppPlatform } from 'src/application/platform';

type UserStorageSaveType =
	| 'locale'
	| 'local-user-id'
	| 'current-user-id'
	| 'users'
	| 'openBiometric'
	| 'backupList'
	| 'terminusInfos'
	| 'launchCounts'
	| 'passwordReseted'
	| 'defaultDomain'
	| 'transferOnlyWifi'
	| 'upgradeIncludeRC';

export const userModeGetItem = async (key: UserStorageSaveType) => {
	return await getAppPlatform().userStorage.getItem(key);
};

export const userModeSetItem = async (key: UserStorageSaveType, value: any) => {
	await getAppPlatform().userStorage.setItem(key, value);
};

export const userModeRemoveItem = async (key: UserStorageSaveType) => {
	await getAppPlatform().userStorage.removeItem(key);
};
