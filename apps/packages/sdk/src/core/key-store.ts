import { AsBytes, AsDate, Serializable } from './encoding';
import { Storable } from './storage';
import { uuid } from './util';

export class KeyStoreEntryInfo extends Serializable {
	id = '';

	@AsDate()
	created!: Date;

	accountId = '';

	authenticatorId = '';

	constructor(init: Partial<KeyStoreEntry> = {}) {
		super();
		Object.assign(this, init);
	}
}

export class KeyStoreEntry extends Storable {
	get kind(): string {
		return 'keystoreentry';
	}

	id = '';

	@AsDate()
	created!: Date;

	accountId = '';

	authenticatorId = '';

	@AsBytes()
	data!: Uint8Array;

	get info() {
		return new KeyStoreEntryInfo({
			id: this.id,
			created: this.created,
			accountId: this.accountId,
			authenticatorId: this.authenticatorId
		});
	}

	constructor(init: Partial<KeyStoreEntry> = {}) {
		super();
		Object.assign(this, init);
	}

	async init() {
		this.id = await uuid();
		this.created = new Date();
	}
}

// export interface KeyStore {
//     set(value: string): Promise<string>;
//     get(id: string): Promise<string>;
//     delete(id: string): Promise<void>;
// }

// export class MemoryKeyStore {
//     private _data = new Map<string, string>();
//     async set(value: string) {
//         const id = await uuid();
//         this._data.set(id, value);
//         return id;
//     }

//     async get(id: string) {
//         return this._data.get(id);
//     }

//     async delete(id: string) {
//         this._data.delete(id);
//     }
// }
