import { Err, ErrorCode } from './error';
import { toByteArray, fromByteArray, byteLength, isBase64 } from './base64';
import { upgrade, downgrade } from './migrations';
import { BigInteger } from '../vendor/jsbn';
export { bytesToBase32, base32ToBytes } from './base32';

export interface SerializationOptions {
	property: string;
	toProperty: string;
	exclude: boolean;
	arrayDeserializeIndividually: boolean;
	fromRaw: (raw: any) => any;
	toRaw: (val: any, version?: string) => any;
}

function registerSerializationOptions(
	proto: Serializable,
	property: string,
	opts: Partial<SerializationOptions>
) {
	if (!proto.hasOwnProperty('_propertySerializationOptions')) {
		const parentOptions = proto._propertySerializationOptions || [];
		proto._propertySerializationOptions = parentOptions
			? [...parentOptions]
			: [];
	}

	// proto._propertySerializationOptions = proto._propertySerializationOptions.filter(o => o.property === property);

	proto._propertySerializationOptions.unshift(
		Object.assign(
			{
				property,
				toProperty: property,
				exclude: false,
				arrayDeserializeIndividually: true,
				toRaw: () => {},
				fromRaw: () => {}
			},
			opts
		)
	);
}

/**
 * Decorator for defining request handler methods
 */
export function AsBigInteger(toProperty?: string) {
	return (proto: Serializable, prop: string) =>
		registerSerializationOptions(proto, prop, {
			toProperty: toProperty || prop,
			toRaw: (val: BigInteger) => val.toString(),
			fromRaw: (raw: string) => new BigInteger(raw)
		});
}

/**
 * Decorator for defining request handler methods
 */
export function AsSerializable(
	cls: SerializableConstructor | SerializableConstructor[],
	toProperty?: string
) {
	return (proto: Serializable, prop: string) =>
		registerSerializationOptions(proto, prop, {
			toProperty: toProperty || prop,
			toRaw: (val: Serializable, version?: string) => val.toRaw(version),
			fromRaw: (raw: any) => {
				const c = Array.isArray(cls)
					? cls.find((c) => new c().kind === raw.kind)
					: cls;
				return c ? new c().fromRaw(raw) : raw;
			}
		});
}

export function AsBytes(toProperty?: string) {
	return (proto: Serializable, prop: string) =>
		registerSerializationOptions(proto, prop, {
			toProperty: toProperty || prop,
			toRaw: (val: any) => bytesToBase64(val),
			fromRaw: (raw: any) => base64ToBytes(raw)
		});
}

export function AsSet(toProperty?: string) {
	return (proto: Serializable, prop: string) =>
		registerSerializationOptions(proto, prop, {
			toProperty: toProperty || prop,
			arrayDeserializeIndividually: false,
			toRaw: (val: Set<any>) => [...val],
			fromRaw: (raw: any[]) => new Set(raw)
		});
}

export function AsDate(toProperty?: string) {
	return (proto: Serializable, prop: string) =>
		registerSerializationOptions(proto, prop, {
			toProperty: toProperty || prop,
			toRaw: (val: Date) => {
				try {
					return val.toISOString();
				} catch (e) {
					return null;
				}
			},
			fromRaw: (raw: string) => new Date(raw)
		});
}

export function Exclude() {
	return (proto: Serializable, prop: string) =>
		registerSerializationOptions(proto, prop, {
			exclude: true
		});
}

export function Serialize(opts: Partial<SerializationOptions>) {
	return (proto: Serializable, prop: string) =>
		registerSerializationOptions(proto, prop, opts);
}

/**
 * Base class for "serializable" classes, i.e. classes that can be serialized
 * into a plain javascript object, JSON string or byte sequence which can be
 * used for storage or data transfer. Subclasses will generally want to overwrite
 * the [[toRaw]], [[fromRaw]] and [[validate]] methods to account for their
 * specific class structure.
 *
 * @example
 *
 * ```ts
 * class MyClass extends Serializable {
 *      name: string;
 *      parent?: MyClass;
 *      bytes: Uint8Array;
 *
 *      toRaw() {
 *          return {
 *              ...super.toRaw(),
 *              bytes: bytesToBase64(this.bytes)
 *          };
 *      }
 *
 *      fromRaw({ bytes, parent, ...rest }) {
 *          return super.fromRaw({
 *              bytes: base64ToBytes(bytes),
 *              parent: parent && new MyClass().fromRaw(parent),
 *              ...rest
 *          });
 *      }
 *
 *      validate() {
 *          return (
 *              super.validate() &&
 *              typeof this.name === "string" &&
 *              this.bytes instanceof Uint8Array &&
 *              (
 *                  typeof this.parent === "undefined" ||
 *                  this.parent instanceof MyClass
 *              )
 *          )
 *      }
 * }
 * ```
 */
export class Serializable {
	/**
	 * A string representing the objects "type", useful for segmenting storage,
	 * among other things. Defaults to the lowercase class name, but can be
	 * overwritten by subclasses
	 */
	get kind(): string {
		return this.constructor.name.toLowerCase();
	}

	_propertySerializationOptions!: SerializationOptions[];

	/**
	 * This is called during deserialization and should verify that all
	 * properties have been populated with values of the correct type.
	 * Subclasses should implement this method based on their class structure.
	 */
	validate() {
		return true;
	}

	/**
	 * Creates a raw javascript object representation of the class, which
	 * can be used for storage or data transmission. Also handles "downgrading" to previous
	 * versions. Use [[_toRaw]] for subclass-specific behavior.
	 */
	toRaw(version?: string): any {
		let raw = this._toRaw(version);
		raw.kind = this.kind;
		raw = downgrade(this.kind, raw, version);
		return raw;
	}

	/**
	 * Restores propertiers from a raw object of the same form generated by
	 * [[toRaw]]. The base implementation blindly copies over values from the
	 * raw object via `Object.assign` so subclasses should explictly process
	 * any propertyies that need special treatment.
	 *
	 * Also takes are of validation and "upgrading" in case the raw object
	 * has an old version. Use the protected [[_fromRaw]] method to implement
	 * subclass-specific behavior.
	 */
	fromRaw(raw: any): this {
		// raw.kind = raw.kind || this.kind;
		raw = upgrade(this.kind, raw);

		this._fromRaw(raw);

		try {
			if (!this.validate()) {
				console.log('failed to validate', this.kind, raw);
				throw new Err(ErrorCode.ENCODING_ERROR);
			}
		} catch (e) {
			throw new Err(ErrorCode.ENCODING_ERROR);
		}
		return this;
	}

	/**
	 * Returns a JSON serialization of the object
	 */
	toJSON(): string {
		return JSON.stringify(this.toRaw());
	}

	/**
	 * Deserializes the object from a JSON string
	 */
	fromJSON(json: string): this {
		return this.fromRaw(JSON.parse(json));
	}

	/**
	 * Returns a serialization of the object in form of a byte array
	 */
	toBytes(): Uint8Array {
		return stringToBytes(this.toJSON());
	}

	/**
	 * Deserializes the object from a byte array
	 */
	fromBytes(bytes: Uint8Array): this {
		return this.fromJSON(bytesToString(bytes));
	}

	/**
	 * Creates a deep clone of the object
	 */
	clone(): this {
		// @ts-ignore: This causes a typescript warning for some reason but works fine in practice
		return new this.constructor().fromRaw(this.toRaw());
	}

	/**
	 * Transform this object into a raw javascript object used for
	 * serialization.  The default implementation simply copies all iterable
	 * properties not included in the [[exlude]] array and calls [[toRaw]] on
	 * any properties that are themselfes instances of [[Serializable]].  This
	 * method should be overwritten by subclasses if certain properties require
	 * special treatment.
	 */
	protected _toRaw(version: string | undefined): any {
		const raw = {} as any;

		for (const [prop, val] of Object.entries(this)) {
			const opts =
				this._propertySerializationOptions &&
				this._propertySerializationOptions.find(
					(opts) => opts.property === prop
				);

			if (prop.startsWith('_') || (opts && opts.exclude)) {
				continue;
			}

			if (opts && typeof val !== 'undefined' && val !== null) {
				raw[opts.property] = Array.isArray(val)
					? val.map((v) => opts.toRaw(v, version))
					: opts.toRaw(val, version);
			} else {
				raw[prop] = val;
			}
		}

		return raw;
	}

	/**
	 * Restore values from a raw object. The default implementation simply copies over
	 * all iterable properties from the base object. Overwrite this method for properties
	 * that require special treatment
	 */
	protected _fromRaw(raw: any) {
		for (const [prop, val] of Object.entries(raw)) {
			if (prop === 'kind') {
				continue;
			}

			const opts =
				this._propertySerializationOptions &&
				this._propertySerializationOptions.find(
					(opts) => opts.toProperty === prop
				);

			// Skip properties that have no serialization options associated with them
			// and are not explicitly defined as a property on the class
			if (!opts && !this.hasOwnProperty(prop)) {
				continue;
			}

			if (opts && typeof val !== 'undefined' && val !== null) {
				this[opts.property] =
					Array.isArray(val) && opts.arrayDeserializeIndividually
						? val.map((v) => opts.fromRaw(v))
						: opts.fromRaw(val);
			} else {
				this[prop] = val;
			}
		}
	}
}

export type Raw<T extends Serializable> = ReturnType<T['toRaw']>;

/**
 * Generic type representing the constructor of a class extending [[Serializable]]
 */
export type SerializableConstructor = new (...args: any[]) => Serializable;

/**
 * Creates a string from a raw javascript object
 */
export function marshal(obj: object): string {
	try {
		return JSON.stringify(obj);
	} catch (e) {
		throw new Err(ErrorCode.ENCODING_ERROR, e.toString());
	}
}

/**
 * Creates a raw javascript object from a string
 */
export function unmarshal(str: string): any {
	try {
		return JSON.parse(str);
	} catch (e) {
		throw new Err(ErrorCode.ENCODING_ERROR, e.toString());
	}
}

export { isBase64 };

/**
 * Converts a byte array to a base64 string
 */
export function bytesToBase64(inp: Uint8Array, urlSafe = true): string {
	try {
		return fromByteArray(inp, urlSafe);
	} catch (e) {
		throw new Err(ErrorCode.ENCODING_ERROR, e.toString());
	}
}

/**
 * Converts a base64 string to a byte array
 */
export function base64ToBytes(inp: string): Uint8Array {
	try {
		return toByteArray(inp);
	} catch (e) {
		throw new Err(ErrorCode.ENCODING_ERROR, e.toString());
	}
}

/**
 * Converts a utf-8 string to a byte array
 */
export function stringToBytes(str: string): Uint8Array {
	try {
		return new TextEncoder().encode(str);
	} catch (e) {
		throw new Err(ErrorCode.ENCODING_ERROR, e.toString());
	}
}

/**
 * Converts a byte array to an utf-8 string
 */
export function bytesToString(bytes: Uint8Array, encoding = 'utf-8') {
	try {
		return new TextDecoder(encoding).decode(bytes);
	} catch (e) {
		throw new Err(ErrorCode.ENCODING_ERROR, e.toString());
	}
}

/**
 * Converts a utf-8 string to its base64 representation
 */
export function stringToBase64(str: string, urlSafe = true): string {
	const bytes = stringToBytes(str);
	return bytesToBase64(bytes, urlSafe);
}

/**
 * Converts the base64 representation of a utf-a string to it's original representation
 */
export function base64ToString(inp: string): string {
	const bytes = base64ToBytes(inp);
	return bytesToString(bytes);
}

/**
 * Returns the byte length of a base64 string
 */
export function base64ByteLength(inp: string): number {
	return byteLength(inp);
}

/**
 * Converts a hex string to a byte array
 */
export function hexToBytes(str: string): Uint8Array {
	try {
		const bytes = new Uint8Array(str.length / 2);
		for (let i = 0; i < bytes.length; i++) {
			bytes[i] = parseInt(str.substring(i * 2, i * 2 + 2), 16);
		}
		return bytes;
	} catch (e) {
		throw new Err(ErrorCode.ENCODING_ERROR, e.toString());
	}
}

/**
 * Converts a byte array to its hexadecimal representation
 */
export function bytesToHex(bytes: Uint8Array): string {
	try {
		let str = '';
		for (const b of bytes) {
			const s = b.toString(16);
			str += s.length == 1 ? '0' + s : s;
		}
		return str;
	} catch (e) {
		throw new Err(ErrorCode.ENCODING_ERROR, e.toString());
	}
}

/**
 * Converts a base64 string to its hexadecimal representation
 */
export function base64ToHex(b64: string): string {
	return bytesToHex(base64ToBytes(b64));
}

/**
 * Converts a hex string to its base64 representation
 */
export function hexToBase64(hex: string): string {
	return bytesToBase64(hexToBytes(hex));
}

export function numToBytes(num: number) {
	return hexToBytes(num.toString(16).padStart(16, '0'));
}

export function bytesToNum(bytes: Uint8Array) {
	return parseInt(bytesToHex(bytes), 16);
}

/**
 * Concatenates a number of Uint8Arrays to a single array
 */
export function concatBytes(
	arrs: Uint8Array[],
	delimiter?: number
): Uint8Array {
	let length = arrs.reduce((len, arr) => len + arr.length, 0);

	if (typeof delimiter !== 'undefined') {
		length += arrs.length - 1;
	}

	const res = new Uint8Array(length);
	let offset = 0;
	for (const arr of arrs) {
		res.set(arr, offset);
		offset += arr.length;

		if (typeof delimiter !== 'undefined' && offset < length) {
			res.set([delimiter], offset);
			offset++;
		}
	}

	return res;
}

/** Checks two byte arrays for equality */
export function equalBytes(a: Uint8Array, b: Uint8Array): boolean {
	if (a.length !== b.length) {
		return false;
	}

	for (let i = 0; i < a.length; i++) {
		if (a[i] !== b[i]) {
			return false;
		}
	}

	return true;
}
