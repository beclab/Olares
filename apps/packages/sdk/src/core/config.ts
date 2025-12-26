import { Serializable, AsSerializable } from './encoding';

/**
 * Generic type representing the constructor of a class extending [[Config]]
 */
export type ConfigConstructor = new (...args: any[]) => Config;

interface ParamDefinition {
	prop: string;
	type: 'string' | 'string[]' | 'number' | 'boolean' | ConfigConstructor;
	secret: boolean;
}

export function ConfigParam(
	type:
		| 'string'
		| 'string[]'
		| 'number'
		| 'boolean'
		| ConfigConstructor = 'string',
	secret = false
) {
	return (proto: Config, prop: string) => {
		if (typeof type === 'function') {
			AsSerializable(type)(proto, prop);
		}
		if (!proto._paramDefinitions) {
			proto._paramDefinitions = [];
		}
		proto._paramDefinitions.push({
			prop,
			type,
			secret
		});
	};
}

export class Config extends Serializable {
	_paramDefinitions!: ParamDefinition[];

	fromEnv(env: { [prop: string]: string }, prefix = 'PL_') {
		for (const { prop, type } of this._paramDefinitions || []) {
			// type is another config object
			if (typeof type === 'function') {
				const newPrefix = `${prefix}${prop
					.replace(/([a-z])([A-Z])/g, '$1_$2')
					.toUpperCase()}_`;
				if (
					!this[prop] &&
					Object.keys(env).some((key) => key.startsWith(newPrefix))
				) {
					this[prop] = new type();
				}
				if (this[prop]) {
					this[prop].fromEnv(env, newPrefix);
				}
				continue;
			}

			const varName = `${prefix}${prop
				.replace(/([a-z])([A-Z])/g, '$1_$2')
				.toUpperCase()}`;

			let str = env[varName];

			if (typeof str === 'undefined') {
				continue;
			} else if (type === 'number') {
				const num = Number(str);
				if (isNaN(num)) {
					throw `Invalid value for var ${varName}: ${str} (should be a number)`;
				}
				this[prop] = num;
			} else if (type === 'boolean') {
				this[prop] = str.toLocaleLowerCase() === 'true';
			} else if (type === 'string[]') {
				this[prop] = str.split(',');
			} else {
				this[prop] = str;
			}
		}

		return this;
	}

	toEnv(prefix = 'PL_', includeUndefined = false) {
		const vars: { [prop: string]: string } = {};

		for (const { prop, type } of this._paramDefinitions || []) {
			// type is another config object
			if (typeof type === 'function') {
				const newPrefix = `${prefix}${prop
					.replace(/([a-z])([A-Z])/g, '$1_$2')
					.toUpperCase()}_`;

				if (!this[prop] && !includeUndefined) {
					continue;
				}

				const subVars =
					this[prop]?.toEnv(newPrefix, includeUndefined) ||
					new type().toEnv(newPrefix, includeUndefined);
				Object.assign(vars, subVars);
				continue;
			}

			const varName = `${prefix}${prop
				.replace(/([a-z])([A-Z])/g, '$1_$2')
				.toUpperCase()}`;

			const val = this[prop];

			if (typeof val === 'undefined' && !includeUndefined) {
				continue;
			}

			switch (type) {
				case 'string[]':
					vars[varName] = val?.join(',') || '';
					break;
				default:
					vars[varName] = val?.toString() || '';
			}
		}

		return vars;
	}

	toRaw(version?: string) {
		const raw = super.toRaw(version);
		for (const { prop, secret } of this._paramDefinitions) {
			if (secret) {
				raw[prop] = '[secret redacted]';
			}
		}
		return raw;
	}
}
