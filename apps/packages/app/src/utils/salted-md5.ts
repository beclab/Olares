// import CryptoJS from 'crypto-js';
import md5 from 'js-md5';
import { compareOlaresVersion } from '@bytetrade/core';

const defaultSuffix = '@Olares2025';
const applyVersion = '1.12.0-0';

export function saltedMD5(
	password: string,
	options: { suffix?: string; osVersion?: string; compareVersion?: string }
) {
	if (!options.osVersion) {
		return password;
	}
	const {
		suffix = defaultSuffix,
		osVersion,
		compareVersion = applyVersion
	} = options;

	const compare = compareOlaresVersion(osVersion, compareVersion).compare;

	if (compare < 0) {
		return password;
	}

	const saltedPassword = password + suffix;
	const hashedPassword = md5(saltedPassword);
	return hashedPassword;
}

export function verifyMD5(
	password: string,
	storedHash: string,
	options: { suffix?: string; osVersion?: string; compareVersion?: string }
) {
	const hashedPassword = saltedMD5(password, options);

	return hashedPassword === storedHash;
}

export function versionIsDailyBuild(version: string) {
	const versionSplits = version.split('-');
	if (versionSplits.length == 0) {
		return false;
	}
	const isRc = versionSplits[1].startsWith('rc');
	if (isRc) {
		return false;
	}
	return true;
}
