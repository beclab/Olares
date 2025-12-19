import { getPlatform } from './platform';
// eslint-disable-next-line @typescript-eslint/ban-ts-comment
// @ts-ignore
import localStorage from 'localforage/src/localforage';

export interface UserStorage {
	// eslint-disable-next-line @typescript-eslint/no-explicit-any
	getItem(key: string): Promise<any>;
	// eslint-disable-next-line @typescript-eslint/no-explicit-any
	setItem(key: string, value: any): Promise<void>;

	removeItem(key: string): Promise<void>;
}

export interface PlatformExtension {
	userStorage: UserStorage;
}

export class PlatformExtensionUserStorageTools {
	// eslint-disable-next-line @typescript-eslint/no-explicit-any
	static async getItem(key: string): Promise<any> {
		const platform = getPlatform();
		console.log('get item');
		if (
			// eslint-disable-next-line @typescript-eslint/no-explicit-any
			(platform as any).userStorage &&
			// eslint-disable-next-line @typescript-eslint/no-explicit-any
			(platform as any).userStorage.getItem
		) {
			try {
				// eslint-disable-next-line @typescript-eslint/no-explicit-any
				return await (platform as any).userStorage.getItem(key);
			} catch (error) {
				console.log(error);
			}
		}
		return await localStorage.getItem(key);
	}
	// eslint-disable-next-line @typescript-eslint/no-explicit-any
	static async setItem(key: string, value: any): Promise<void> {
		const platform = getPlatform();
		if (
			// eslint-disable-next-line @typescript-eslint/no-explicit-any
			(platform as any).userStorage &&
			// eslint-disable-next-line @typescript-eslint/no-explicit-any
			(platform as any).userStorage.setItem
		) {
			try {
				// eslint-disable-next-line @typescript-eslint/no-explicit-any
				return await (platform as any).userStorage.setItem(key, value);
			} catch (error) {
				console.log(error);
			}
		}
		await localStorage.setItem(key, value);
	}

	// eslint-disable-next-line @typescript-eslint/no-explicit-any
	static async removeItem(key: string): Promise<any> {
		const platform = getPlatform();
		if (
			// eslint-disable-next-line @typescript-eslint/no-explicit-any
			(platform as any).userStorage &&
			// eslint-disable-next-line @typescript-eslint/no-explicit-any
			(platform as any).userStorage.removeItem
		) {
			try {
				// eslint-disable-next-line @typescript-eslint/no-explicit-any
				return await (platform as any).userStorage.removeItem(key);
			} catch (error) {
				console.log(error);
			}
		}
		return await localStorage.removeItem(key);
	}
}
