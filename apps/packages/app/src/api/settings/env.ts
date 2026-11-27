import axios from 'axios';
import { BaseEnv, UpdateEnvBody } from 'src/constant';
import { useTokenStore } from 'src/stores/settings/token';

/**
 * Wraps a fetcher with in-flight de-duplication plus a short-lived cache.
 * Concurrent callers (e.g. many EnvVarInput instances mounting at once) share
 * one network request, and repeated reads within the TTL reuse the result.
 */
function createCachedFetcher<T>(
	fetcher: () => Promise<T>,
	ttlMs: number
): () => Promise<T> {
	let cache: { value: T; expireAt: number } | null = null;
	let inFlight: Promise<T> | null = null;

	return (): Promise<T> => {
		if (cache && cache.expireAt > Date.now()) {
			return Promise.resolve(cache.value);
		}
		if (inFlight) {
			return inFlight;
		}
		inFlight = fetcher()
			.then((value) => {
				cache = { value, expireAt: Date.now() + ttlMs };
				return value;
			})
			.finally(() => {
				inFlight = null;
			});
		return inFlight;
	};
}

const ENV_LIST_CACHE_TTL_MS = 5000;

const systemEnvFetcher = createCachedFetcher<BaseEnv[]>(() => {
	const tokenStore = useTokenStore();
	return axios.get(`${tokenStore.url}/api/env/systemenvs`);
}, ENV_LIST_CACHE_TTL_MS);

const userEnvFetcher = createCachedFetcher<BaseEnv[]>(() => {
	const tokenStore = useTokenStore();
	return axios.get(`${tokenStore.url}/api/env/userenvs`);
}, ENV_LIST_CACHE_TTL_MS);

export async function getAppEnv(appName: string): Promise<BaseEnv[]> {
	const tokenStore = useTokenStore();
	return await axios.get(`${tokenStore.url}/api/env/apps/${appName}/env`);
}

export async function updateAppEnv(
	appName: string,
	body: UpdateEnvBody
): Promise<any> {
	const tokenStore = useTokenStore();
	return await axios.put(`${tokenStore.url}/api/env/apps/${appName}/env`, body);
}

export async function getSystemEnvList(): Promise<BaseEnv[]> {
	return systemEnvFetcher();
}

export async function updateSystemEnv(body: UpdateEnvBody): Promise<BaseEnv> {
	const tokenStore = useTokenStore();
	return await axios.put(`${tokenStore.url}/api/env/systemenvs`, body);
}

export async function getUserEnvList(): Promise<BaseEnv[]> {
	return userEnvFetcher();
}

export async function addUserEnv(env: BaseEnv) {
	const tokenStore = useTokenStore();
	return await axios.post(`${tokenStore.url}/api/env/userenvs`, env);
}

export async function updateUserEnv(body: UpdateEnvBody): Promise<BaseEnv> {
	const tokenStore = useTokenStore();
	return await axios.put(`${tokenStore.url}/api/env/userenvs`, body);
}

export async function deleteUserEnv(envName: string): Promise<BaseEnv> {
	const tokenStore = useTokenStore();
	return await axios.delete(`${tokenStore.url}/api/env/userenvs/${envName}`);
}

export async function remoteOptionsProxy(endpoint: string): Promise<any> {
	return await axios.post(`/api/env/appenv/remoteOptions`, {
		endpoint
	});
}
