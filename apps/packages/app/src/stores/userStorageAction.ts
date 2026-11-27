import { getUserStorage } from 'src/application/platform/userStorage';

type UserStorageSaveType =
	| 'locale'
	| 'local-user-id'
	| 'current-user-id'
	| 'users'
	| 'openBiometric'
	// Mirrors the `kSecAttrServer` value used when writing the biometric
	// credential. The iOS AutoFill extension reads this from the App Group
	// to locate the keychain item; without it the extension's biometric
	// flow can never resolve to a stored password.
	| 'biometric_server'
	| 'backupList'
	| 'terminusInfos'
	| 'launchCounts'
	| 'passwordReseted'
	| 'defaultDomain'
	| 'transferOnlyWifi'
	| 'upgradeIncludeRC'
	| 'temporaryLockUsers';

export const userModeGetItem = async (key: UserStorageSaveType) => {
	return await getUserStorage().getItem(key);
};

export const userModeSetItem = async (key: UserStorageSaveType, value: any) => {
	await getUserStorage().setItem(key, value);
};

export const userModeRemoveItem = async (key: UserStorageSaveType) => {
	await getUserStorage().removeItem(key);
};
