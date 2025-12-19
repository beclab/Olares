//import { Serializable } from "./encoding";
//import { Serializable, stringToBytes, AsBytes, AsSerializable, AsDate, bytesToBase64 } from "./encoding";
import { Serializable, AsSerializable } from './encoding';
import { Storable } from './storage';

export enum SecertType {
	None = 'none',
	PrivateKey = 'private_key',
	Mnemonic = 'mnemonic',
	Exchange = 'exchange',
	Website = 'website'
}

export class SecertData extends Serializable {
	exchange?: string = undefined;
	account?: string = undefined;
	subaccount?: string = undefined;
	key?: string = undefined;
	secret?: string = undefined;
	passphrase?: string = undefined;
	mnemonic?: string = undefined;
	private_key?: string = undefined;
	curve?: string = undefined;
	address?: string = undefined;

	constructor(props?: Partial<SecertData>) {
		super();
		props && Object.assign(this, props);
	}
}

export enum HostType {
	None = 'none',
	ONLY_SAVE_NAME = 'only_save_name',
	SAVE_SECERT = 'save_secert'
}

export class Secert extends Storable {
	secert_type: SecertType = SecertType.None;
	host_type: HostType = HostType.None;
	name: string = '';
	id: string = '';

	@AsSerializable(SecertData)
	data?: SecertData;

	constructor(props?: Partial<Secert>) {
		super();
		Object.assign(this, props);
	}

	match(st: SecertType[], ht: HostType[]): boolean {
		// let m  = 0;
		let s1 = st.find((a) => a != SecertType.None && a == this.secert_type);
		if (s1) {
			return true;
		}

		let s2 = ht.find((a) => a != HostType.None && a == this.host_type);
		if (s2) {
			return true;
		}

		return false;

		// if( ht && ht != HostType.None) {
		//   if( ht != this.host_type ) {
		//     return false;
		//   }
		//   m++;
		// }

		// return m > 0;
	}
}
