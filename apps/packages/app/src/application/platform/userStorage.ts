import localForage from 'localforage';
import { getAppPlatform } from '.';
import type { UserStorage } from '../interface/platform';

const browserUserStorage: UserStorage = {
	getItem: (key) => localForage.getItem(key),
	async setItem(key, value) {
		await localForage.setItem(key, value);
	},
	removeItem: (key) => localForage.removeItem(key)
};

export function getUserStorage(): UserStorage {
	return getAppPlatform().userStorage ?? browserUserStorage;
}
