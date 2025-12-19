import axios from 'axios';
import { BaseEnv, EnvDetail, UpdateEnvBody } from 'src/constant';
import { useTokenStore } from 'src/stores/settings/token';

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
	const tokenStore = useTokenStore();
	return await axios.get(`${tokenStore.url}/api/env/systemenvs`);
}

export async function updateSystemEnv(body: UpdateEnvBody): Promise<BaseEnv> {
	const tokenStore = useTokenStore();
	return await axios.put(`${tokenStore.url}/api/env/systemenvs`, body);
}

export async function getUserEnvList(): Promise<BaseEnv[]> {
	const tokenStore = useTokenStore();
	return await axios.get(`${tokenStore.url}/api/env/userenvs`);
}

export async function updateUserEnv(body: UpdateEnvBody): Promise<BaseEnv> {
	const tokenStore = useTokenStore();
	return await axios.put(`${tokenStore.url}/api/env/userenvs`, body);
}

export async function remoteOptionsProxy(endpoint: string): Promise<any> {
	return await axios.post(`/api/env/appenv/remoteOptions`, {
		endpoint
	});
}
