import {
	Storage,
	Storable,
	StorableConstructor,
	StorageListOptions,
	StorageQuery
} from './core/storage';
import { Err, ErrorCode } from './core/error';
// eslint-disable-next-line @typescript-eslint/ban-ts-comment
// @ts-ignore
import localStorage from 'localforage/src/localforage';

export class LocalStorage implements Storage {
	async save(s: Storable) {
		console.log('save ' + s.kind);
		await localStorage.setItem(`${s.kind}_${s.id}`, s.toRaw());
	}

	async saveID(id: string, s: Storable) {
		await localStorage.setItem(`${s.kind}_${id}`, s.toRaw());
	}

	async get<T extends Storable>(cls: T | StorableConstructor<T>, id: string) {
		const s = cls instanceof Storable ? cls : new cls();
		const data = await localStorage.getItem(`${s.kind}_${id}`);
		if (!data) {
			throw new Err(ErrorCode.NOT_FOUND);
		}
		return s.fromRaw(data);
	}

	async delete(s: Storable) {
		await localStorage.removeItem(`${s.kind}_${s.id}`);
	}

	// async deleteID(id: string) {
	// 	await localStorage.removeItem(id);
	// }

	async clear() {
		await localStorage.clear();
	}

	async list<T extends Storable>(
		cls: StorableConstructor<T>,
		// eslint-disable-next-line @typescript-eslint/no-unused-vars
		opts: StorageListOptions
	): Promise<T[]> {
		const s = new cls();
		const tList: T[] = [];
		const keys = (await localStorage.keys()) as string[];
		console.log(keys);
		for (let index = 0; index < keys.length; index++) {
			const key = keys[index];
			if (opts.query && opts.query.op === 'eq') {
				const searchId =
					`_${s.id}` +
					(opts.query.path ? opts.query.path : '') +
					(opts.query.value ? opts.query.value : '');
				if (key.indexOf(searchId) >= 0) {
					const data = await localStorage.getItem(key);
					console.log(data);
					if (data) {
						const item = s.fromRaw(data);
						if (data.kind && item.kind !== data.kind) {
							await localStorage.removeItem(key);
						}
						tList.push(item);
					}
				}
			}
		}

		return tList;
	}

	async count<T extends Storable>(
		// eslint-disable-next-line @typescript-eslint/no-unused-vars
		_cls: StorableConstructor<T>,
		// eslint-disable-next-line @typescript-eslint/no-unused-vars
		_: StorageQuery
	): Promise<number> {
		throw new Err(ErrorCode.NOT_SUPPORTED);
	}
}
