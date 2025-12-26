import { Platform } from 'quasar';

interface StorageInterface {
	get(key: string): Promise<any>;
	set(key: string, value: any): Promise<void>;
	remove(key: string): Promise<void>;
	clear(): Promise<void>;
}

class WebStorage implements StorageInterface {
	async get(key: string): Promise<any> {
		const value = localStorage.getItem(key);
		return value ? JSON.parse(value) : null;
	}

	async set(key: string, value: any): Promise<void> {
		localStorage.setItem(key, JSON.stringify(value));
	}

	async remove(key: string): Promise<void> {
		localStorage.removeItem(key);
	}

	async clear(): Promise<void> {
		localStorage.clear();
	}
}

class BexStorage implements StorageInterface {
	async get(key: string): Promise<any> {
		return new Promise((resolve) => {
			chrome.storage.local.get(key, (result) => {
				resolve(result[key] || null);
			});
		});
	}

	async set(key: string, value: any): Promise<void> {
		return new Promise((resolve) => {
			chrome.storage.local.set({ [key]: value }, () => {
				resolve();
			});
		});
	}

	async remove(key: string): Promise<void> {
		return new Promise((resolve) => {
			chrome.storage.local.remove(key, () => {
				resolve();
			});
		});
	}

	async clear(): Promise<void> {
		return new Promise((resolve) => {
			chrome.storage.local.clear(() => {
				resolve();
			});
		});
	}
}

export const storage: StorageInterface = Platform.is.bex
	? new BexStorage()
	: new WebStorage();
