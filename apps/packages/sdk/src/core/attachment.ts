import { Serializable, AsBytes } from './encoding';
import { SimpleContainer } from './container';
import { VaultID } from './vault';
import { AESKeyParams } from './crypto';
import { getCryptoProvider as getProvider } from './platform';
import { Err, ErrorCode } from './error';
import { RequestProgress } from './transport';

export async function readFileAsUint8Array(blob: File): Promise<Uint8Array> {
	return new Promise((resolve, reject) => {
		const reader = new FileReader();

		reader.onload = () => {
			const result = new Uint8Array(reader.result as ArrayBuffer);
			resolve(result);
		};

		reader.onerror = (error) => {
			reader.abort();
			reject(error);
		};

		reader.readAsArrayBuffer(blob);
	});
}

export async function readFileAsArrayBuffer(blob: File): Promise<ArrayBuffer> {
	return new Promise((resolve, reject) => {
		const reader = new FileReader();

		reader.onload = () => {
			resolve(reader.result as ArrayBuffer);
		};

		reader.onerror = (error) => {
			reader.abort();
			reject(error);
		};

		reader.readAsArrayBuffer(blob);
	});
}

export async function readFileAsText(blob: File): Promise<string> {
	return new Promise((resolve, reject) => {
		const reader = new FileReader();

		reader.onload = () => {
			resolve(reader.result as string);
		};

		reader.onerror = (error) => {
			reader.abort();
			reject(error);
		};

		reader.readAsText(blob);
	});
}

function readFileAsDataURL(blob: File): Promise<string> {
	return new Promise((resolve, reject) => {
		const reader = new FileReader();

		reader.onload = () => {
			resolve(reader.result as string);
		};

		reader.onerror = (e) => {
			reader.abort();
			reject(e);
		};

		reader.readAsDataURL(blob);
	});
}

export type AttachmentID = string;

export class AttachmentInfo extends Serializable {
	constructor(vals: Partial<AttachmentInfo> = {}) {
		super();
		Object.assign(this, vals);
	}

	id: AttachmentID = '';
	vault: VaultID = '';
	name = '';
	size = 0;
	type = '';

	@AsBytes()
	key!: Uint8Array;
}

export class Attachment extends SimpleContainer {
	id: AttachmentID = '';
	vault: VaultID = '';
	name = '';
	size = 0;
	type = '';
	uploadProgress?: RequestProgress;
	downloadProgress?: RequestProgress;

	constructor({ key, ...info }: Partial<AttachmentInfo> = {}) {
		super();
		Object.assign(this, {
			_key: key,
			...info
		});
	}

	get info(): AttachmentInfo {
		return new AttachmentInfo({
			id: this.id,
			vault: this.vault,
			name: this.name,
			type: this.type,
			size: this.size,
			key: this._key
		});
	}

	get loaded(): boolean {
		return !!this.encryptedData;
	}

	async fromFile(file: File) {
		this.type = file.type;
		this.size = file.size;
		this.name = file.name;

		const data = await readFileAsUint8Array(file);

		this._key = await getProvider().generateKey({
			algorithm: 'AES',
			keySize: this.encryptionParams.keySize
		} as AESKeyParams);

		await this.setData(data);
		return this;
	}

	async toFile(): Promise<File> {
		const data = await this.getData();
		return new File([data], this.name, { type: this.type });
	}

	async toDataURL(): Promise<string> {
		const file = await this.toFile();
		return readFileAsDataURL(file);
	}

	async toObjectURL(): Promise<string> {
		const file = await this.toFile();
		return URL.createObjectURL(file);
	}

	async toText(): Promise<string> {
		const file = await this.toFile();
		return readFileAsText(file);
	}

	validate() {
		return (
			typeof this.id === 'string' &&
			typeof this.vault === 'string' &&
			typeof this.size === 'number'
		);
	}
}

export interface AttachmentStorage {
	put(a: Attachment): Promise<void>;
	get(vault: VaultID, id: AttachmentID): Promise<Attachment>;
	delete(vault: VaultID, id: AttachmentID): Promise<void>;
	deleteAll(vault: VaultID): Promise<void>;
	getUsage(vault: VaultID): Promise<number>;
}

export class MemoryAttachmentStorage implements AttachmentStorage {
	private _storage = new Map<string, Attachment>();

	async put(a: Attachment): Promise<void> {
		this._storage.set(`${a.vault}_${a.id}`, a);
	}

	async get(vault: VaultID, id: AttachmentID): Promise<Attachment> {
		const att = this._storage.get(`${vault}_${id}`);
		if (!att) {
			throw new Err(ErrorCode.NOT_FOUND);
		}
		return att;
	}

	async delete(vault: VaultID, id: AttachmentID): Promise<void> {
		this._storage.delete(`${vault}_${id}`);
	}

	async deleteAll(vault: VaultID): Promise<void> {
		for (const key of this._storage.keys()) {
			if (key.startsWith(vault)) {
				this._storage.delete(key);
			}
		}
	}

	async getUsage(vault: VaultID): Promise<number> {
		let size = 0;
		for (const [key, att] of this._storage.entries()) {
			if (key.startsWith(vault)) {
				size += att.size;
			}
		}
		return size;
	}
}
