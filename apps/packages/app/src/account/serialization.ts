import { Base64 } from 'js-base64';
import { compareOlaresVersion } from '@bytetrade/core';
import { Err, ErrorCode } from './error';

const WIRE_VERSION = '4.0.0';

export function bytesToBase64(bytes: Uint8Array): string {
	return Base64.fromUint8Array(bytes, true);
}

export function base64ToBytes(value: string): Uint8Array {
	return Base64.toUint8Array(value);
}

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
	if (
		!Object.prototype.hasOwnProperty.call(
			proto,
			'_propertySerializationOptions'
		)
	) {
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

	toRaw(version?: string): any {
		return { ...this._toRaw(version), kind: this.kind, version: WIRE_VERSION };
	}

	fromRaw(raw: any): this {
		// These Web records have no field migrations. Older records may omit kind and version.
		if (
			typeof raw.version === 'string' &&
			compareOlaresVersion(raw.version, WIRE_VERSION).compare > 0
		) {
			throw new Err(ErrorCode.UNSUPPORTED_VERSION);
		}

		this._fromRaw(raw);

		try {
			if (!this.validate()) {
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
		return new TextEncoder().encode(this.toJSON());
	}

	/**
	 * Deserializes the object from a byte array
	 */
	fromBytes(bytes: Uint8Array): this {
		return this.fromJSON(new TextDecoder().decode(bytes));
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
			if (!opts && !Object.prototype.hasOwnProperty.call(this, prop)) {
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
