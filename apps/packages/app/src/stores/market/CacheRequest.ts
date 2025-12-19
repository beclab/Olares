import { EventEmitter } from 'events';

export class CacheRequest<T> extends EventEmitter {
	private readonly key: string;
	private readonly requestFn: () => Promise<T>;
	private readonly expire?: number;
	private updateLock: Promise<T | void> | null = null;
	private dataHandler?: (data: T) => void;

	constructor(
		key: string,
		requestFn: () => Promise<T>,
		options: { expire?: number; onData?: (data: T) => void } = {}
	) {
		super();
		this.key = key;
		this.requestFn = requestFn;
		this.expire = options.expire;
		this.dataHandler = options.onData;
	}

	async get(): Promise<T> {
		const cacheStr = localStorage.getItem(this.key);
		const now = Date.now();
		let hasValidCache = false;
		let cacheData: T | null = null;

		if (cacheStr) {
			try {
				const cache = JSON.parse(cacheStr) as { data: T; expire?: number };
				hasValidCache = !cache.expire || now < cache.expire;
				if (hasValidCache) {
					cacheData = cache.data;
				}
			} catch (err) {
				console.error(`parse error（key: ${this.key}）`, err);
				hasValidCache = false;
			}
		}

		if (hasValidCache && cacheData) {
			this.dataHandler?.(cacheData);
			this.updateCache().catch((err) => {
				console.error(`background updateCache error（key: ${this.key}）`, err);
			});
			return cacheData;
		} else {
			try {
				const newData = await this.updateCache();
				if (newData !== undefined) {
					return newData as T;
				}
				const data = this.getCacheData();
				if (data) {
					return data;
				}
				throw new Error(`no data available for key: ${this.key}`);
			} catch (err) {
				console.error(`get data error（key: ${this.key}）`, err);
				throw err;
			}
		}
	}

	async forceRefresh(): Promise<T> {
		if (this.updateLock) {
			await this.updateLock;
		}

		const newData = await this.requestFn();

		if (!this.isValidData(newData)) {
			throw new Error(`invalid data in force refresh（key: ${this.key}）`);
		}

		const oldData = this.getCacheData();
		this.saveToLocalStorage(newData);
		this.emit('requestFinish', newData);

		if (!this.isEqual(oldData, newData)) {
			this.emit('cacheChange', newData, oldData);
		}

		this.dataHandler?.(newData);

		return newData;
	}

	async updateCache(): Promise<T | void> {
		if (this.updateLock) {
			await this.updateLock;
		}

		this.updateLock = (async () => {
			try {
				const oldData = this.getCacheData();
				const newData = await this.requestFn();

				this.emit('requestFinish', newData);

				if (this.isValidData(newData)) {
					if (oldData !== null && !this.isEqual(oldData, newData)) {
						this.saveToLocalStorage(newData);
						this.emit('cacheChange', newData, oldData);
						this.dataHandler?.(newData);
					} else if (oldData === null) {
						this.saveToLocalStorage(newData);
						this.dataHandler?.(newData);
					}
					return newData;
				} else {
					throw new Error(`invalid data in request（key: ${this.key}）`);
				}
			} catch (err) {
				console.error(`update cache failure（key: ${this.key}）`, err);
				const oldData = this.getCacheData();
				if (oldData) {
					return oldData;
				}
				throw err;
			} finally {
				this.updateLock = null;
			}
		})();

		return this.updateLock;
	}

	private isValidData(data: T): boolean {
		if (data === null) {
			return false;
		}
		if (typeof data === 'object' && 'code' in data) {
			return (data as any).code === 200;
		}
		return true;
	}

	getCacheData(): T | null {
		const cacheStr = localStorage.getItem(this.key);
		if (cacheStr) {
			try {
				const cache = JSON.parse(cacheStr) as { data: T; expire?: number };
				return cache.data;
			} catch (err) {
				return null;
			}
		}
		return null;
	}

	clearCache() {
		localStorage.removeItem(this.key);
	}

	private isEqual(oldData: T | null, newData: T): boolean {
		if (oldData === null) return false;
		return JSON.stringify(oldData) === JSON.stringify(newData);
	}

	private saveToLocalStorage(data: T) {
		const cache = {
			data,
			expire: this.expire ? Date.now() + this.expire : undefined
		};
		localStorage.setItem(this.key, JSON.stringify(cache));
	}
}
