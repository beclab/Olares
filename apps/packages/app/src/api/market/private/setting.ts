import axios from 'axios';
import { useCenterStore } from 'src/stores/market/center';
import globalConfig from 'src/api/market/config';

export async function setNsfw(nsfw: boolean): Promise<boolean> {
	try {
		const { data }: any = await axios.post('/app-store/v1/settings/nsfw', {
			nsfw: nsfw
		});
		console.log(data);
		return true;
	} catch (e) {
		console.log(e);
		return false;
	}
}

export async function getNsfw(): Promise<boolean> {
	try {
		const { data }: any = await axios.get('/app-store/v1/settings/nsfw', {});
		console.log(data);
		return data.nsfw;
	} catch (e) {
		console.log(e);
		return false;
	}
}

export async function getSettingConfig(): Promise<any> {
	if (globalConfig.isOfficial) {
		return { selected_source: 'market.olares' };
	}
	const store = useCenterStore();
	const url = store.appUrl + '/settings/market-settings';
	const { data } = await axios.get(url);
	console.log(data);
	return data;
}

export async function setMarketSource(sourId: string): Promise<any> {
	const store = useCenterStore();
	const url = store.appUrl + '/settings/market-settings';
	const { data } = await axios.put(url, {
		selected_source: sourId
	});
	console.log(data);
	return data;
}
