import { Err, ErrorCode } from './error';

const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ234567';

export function bytesToBase32(arr: Uint8Array) {
	let bits = 0;
	let value = 0;
	let str = '';

	for (let i = 0; i < arr.length; i++) {
		value = (value << 8) | arr[i];
		bits += 8;

		while (bits >= 5) {
			str += chars[(value >>> (bits - 5)) & 31];
			bits -= 5;
		}
	}

	if (bits > 0) {
		str += chars[(value << (5 - bits)) & 31];
	}

	return str;
}

export function base32ToBytes(str: string) {
	const strUpp = str.toUpperCase();
	const arr = new Uint8Array(((str.length * 5) / 8) | 0);

	let bits = 0;
	let value = 0;
	let index = 0;

	for (let i = 0; i < strUpp.length; i++) {
		const idx = chars.indexOf(strUpp[i]);

		if (idx === -1) {
			throw new Err(
				ErrorCode.ENCODING_ERROR,
				`Invalid Base32 character found: ${strUpp[i]}`
			);
		}

		value = (value << 5) | idx;
		bits += 5;

		if (bits >= 8) {
			arr[index++] = (value >>> (bits - 8)) & 255;
			bits -= 8;
		}
	}

	return arr;
}
