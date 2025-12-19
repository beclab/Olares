import { numToBytes, bytesToNum, stringToBytes } from './encoding';
import { HMACParams } from './crypto';
import { getCryptoProvider as getProvider } from './platform';
import { base32ToBytes } from './encoding';
import { bytesToBase32 } from './base32';

export interface HOTPOpts {
	digits: number;
	hash: 'SHA-1' | 'SHA-256';
}

export interface TOTPOpts extends HOTPOpts {
	interval: number;
}

export interface TOTPValidationOpts extends TOTPOpts {
	window: number;
}

function getToken(hmac: Uint8Array, digits: number = 6): string {
	const offset = hmac[hmac.length - 1] & 0xf;
	const bin = new Uint8Array([
		hmac[offset] & 0x7f,
		hmac[offset + 1],
		hmac[offset + 2],
		hmac[offset + 3]
	]);
	const num = bytesToNum(bin);
	return (num % 10 ** digits).toString().padStart(digits, '0');
}

export async function hotp(
	secret: Uint8Array,
	counter: number,
	{ hash, digits }: HOTPOpts = { digits: 6, hash: 'SHA-1' }
) {
	const hmac = await getProvider().sign(
		secret,
		numToBytes(counter),
		new HMACParams({ hash, keySize: secret.length * 8 })
	);
	return getToken(hmac, digits);
}

export async function totp(
	secret: Uint8Array,
	time: number = Date.now(),
	{ interval, ...opts }: TOTPOpts = { interval: 30, digits: 6, hash: 'SHA-1' }
) {
	const counter = getCounter(time, { interval });
	return hotp(secret, counter, opts);
}

export async function validateHotp(
	secret: Uint8Array,
	token: string,
	counter: number,
	{ interval, window, ...opts }: TOTPValidationOpts = {
		interval: 30,
		digits: 6,
		hash: 'SHA-1',
		window: 1
	}
) {
	counter = Math.floor(counter);
	let matchFound = false;
	for (let c = counter - window; c <= counter + window; c++) {
		const t = await hotp(secret, c, opts);
		matchFound =
			matchFound ||
			(await getProvider().timingSafeEqual(
				stringToBytes(t),
				stringToBytes(token)
			));
	}
	return matchFound;
}

export function getCounter(
	time: number = Date.now(),
	{ interval = 30 }: { interval?: number } = {}
) {
	return Math.floor(time / interval / 1000);
}

export function parseURL(data: string) {
	const url = new URL(data);
	const params = new URLSearchParams(url.search);
	const secret = params.get('secret');

	if (!secret || !base32ToBytes(secret).length) {
		throw 'Invalid secret';
	}

	return { secret };
}

export async function generateSecret() {
	const bytes = await getProvider().randomBytes(16);
	return bytesToBase32(bytes);
}

export function generateURL({
	secret,
	account,
	issuer = 'Padloc',
	type = 'totp',
	interval = 30,
	digits = 6,
	hash = 'SHA-1'
}: {
	secret: string;
	account: string;
	issuer?: string;
	type?: 'hotp' | 'totp';
} & Partial<TOTPOpts>) {
	const params = new URLSearchParams();
	params.set('secret', secret);
	params.set('issuer', issuer);
	params.set('digits', digits.toString());
	params.set('algorithm', hash.replace('-', ''));
	params.set('period', interval.toString());
	return `otpauth://${type}/${encodeURIComponent(
		issuer
	)}:${encodeURIComponent(account)}?${params.toString()}`;
}
