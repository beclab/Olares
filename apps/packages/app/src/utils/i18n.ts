import { i18n } from 'src/boot/i18n';

export function _t(value: string, params: Record<string, string>) {
	let str = i18n.global.t(value, params);
	for (const key in params) {
		if (Object.prototype.hasOwnProperty.call(params, key)) {
			str = str.replace(`{${key}}`, params[key]);
		}
	}
	return str;
}
