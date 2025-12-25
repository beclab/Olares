import { getMsauth, setMsauth } from './storage';
import { URL_MICROSOFT_AUTH } from '../config';
import { fetchHandle } from './fetch';
import { kissLog } from './log';

const parseMSToken = (token) => {
	try {
		return JSON.parse(atob(token.split('.')[1])).exp;
	} catch (err) {
		kissLog(err, 'parseMSToken');
	}
	return 0;
};

/**
 * Closure cache token to reduce storage queries
 * @returns
 */
const _msAuth = () => {
	let { token, exp } = {};

	return async () => {
		// Query memory cache
		const now = Date.now();
		if (token && exp * 1000 > now + 1000) {
			return [token, exp];
		}

		// Query storage cache
		const res = await getMsauth();
		token = res?.token;
		exp = res?.exp;
		if (token && exp * 1000 > now + 1000) {
			return [token, exp];
		}

		// Cache is not available or expired, query API
		token = await fetchHandle({ input: URL_MICROSOFT_AUTH });
		exp = parseMSToken(token);
		await setMsauth({ token, exp });
		return [token, exp];
	};
};

export const msAuth = _msAuth();
