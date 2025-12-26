const lookup: string[] = [];
const lookupURL: string[] = [];
const revLookup: number[] = [];

const code = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/';
const codeURL =
	'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_';

for (let i = 0, len = code.length; i < len; ++i) {
	lookup[i] = code[i];
	lookupURL[i] = codeURL[i];
	revLookup[code.charCodeAt(i)] = i;
}

// Support decoding URL-safe base64 strings, as Node.js does.
// See: https://en.wikipedia.org/wiki/Base64#URL_applications
revLookup['-'.charCodeAt(0)] = 62;
revLookup['_'.charCodeAt(0)] = 63;

function getLens(b64: string) {
	// Remove trailing padding characters
	const trimmed = b64.replace(/=*$/, '');
	const len = trimmed.length;
	const padLen = (4 - (len % 4)) % 4;
	return [len, padLen];
}

// base64 is 4/3 + up to two characters of the original data
export function byteLength(b64: string): number {
	const lens = getLens(b64);
	return _byteLength(lens[0], lens[1]);
}

function _byteLength(validLen: number, placeHoldersLen: number) {
	return ((validLen + placeHoldersLen) * 3) / 4 - placeHoldersLen;
}

export function isBase64(str: string) {
	for (let i = 0; i < str.length; i++) {
		if (!(typeof revLookup[str.charCodeAt(i)] === 'number')) {
			return false;
		}
	}
	return true;
}

export function toByteArray(b64: string) {
	let tmp;
	const lens = getLens(b64);
	const validLen = lens[0];
	const placeHoldersLen = lens[1];

	const arr = new Uint8Array(_byteLength(validLen, placeHoldersLen));

	let curByte = 0;

	// if there are placeholders, only get up to the last complete 4 chars
	const len = placeHoldersLen > 0 ? validLen - 4 : validLen;

	let i = 0;
	for (; i < len; i += 4) {
		tmp =
			(revLookup[b64.charCodeAt(i)] << 18) |
			(revLookup[b64.charCodeAt(i + 1)] << 12) |
			(revLookup[b64.charCodeAt(i + 2)] << 6) |
			revLookup[b64.charCodeAt(i + 3)];
		arr[curByte++] = (tmp >> 16) & 0xff;
		arr[curByte++] = (tmp >> 8) & 0xff;
		arr[curByte++] = tmp & 0xff;
	}

	if (placeHoldersLen === 2) {
		tmp =
			(revLookup[b64.charCodeAt(i)] << 2) |
			(revLookup[b64.charCodeAt(i + 1)] >> 4);
		arr[curByte++] = tmp & 0xff;
	}

	if (placeHoldersLen === 1) {
		tmp =
			(revLookup[b64.charCodeAt(i)] << 10) |
			(revLookup[b64.charCodeAt(i + 1)] << 4) |
			(revLookup[b64.charCodeAt(i + 2)] >> 2);
		arr[curByte++] = (tmp >> 8) & 0xff;
		arr[curByte++] = tmp & 0xff;
	}

	return arr;
}

function tripletToBase64(num: number, urlSafe = false) {
	const lu = urlSafe ? lookupURL : lookup;
	return (
		lu[(num >> 18) & 0x3f] +
		lu[(num >> 12) & 0x3f] +
		lu[(num >> 6) & 0x3f] +
		lu[num & 0x3f]
	);
}

function encodeChunk(
	uint8: Uint8Array,
	start: number,
	end: number,
	urlSafe = false
) {
	let tmp;
	const output: any = [];
	for (let i = start; i < end; i += 3) {
		tmp =
			((uint8[i] << 16) & 0xff0000) +
			((uint8[i + 1] << 8) & 0xff00) +
			(uint8[i + 2] & 0xff);
		output.push(tripletToBase64(tmp, urlSafe));
	}
	return output.join('');
}

export function fromByteArray(uint8: Uint8Array, urlSafe = false) {
	let tmp;
	const lu = urlSafe ? lookupURL : lookup;
	const padChar = urlSafe ? '' : '=';
	const len = uint8.length;
	const extraBytes = len % 3; // if we have 1 byte left, pad 2 bytes
	const parts: any = [];
	const maxChunkLength = 16383; // must be multiple of 3

	// go through the array every three bytes, we'll deal with trailing stuff later
	for (let i = 0, len2 = len - extraBytes; i < len2; i += maxChunkLength) {
		parts.push(
			encodeChunk(
				uint8,
				i,
				i + maxChunkLength > len2 ? len2 : i + maxChunkLength,
				urlSafe
			)
		);
	}

	// pad the end with zeros, but make sure to not forget the extra bytes
	if (extraBytes === 1) {
		tmp = uint8[len - 1];
		parts.push(lu[tmp >> 2] + lu[(tmp << 4) & 0x3f] + padChar + padChar);
	} else if (extraBytes === 2) {
		tmp = (uint8[len - 2] << 8) + uint8[len - 1];
		parts.push(
			lu[tmp >> 10] +
				lu[(tmp >> 4) & 0x3f] +
				lu[(tmp << 2) & 0x3f] +
				padChar
		);
	}

	return parts.join('');
}
