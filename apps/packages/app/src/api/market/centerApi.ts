import axios from 'axios';
import {
	AppFullInfoLatest,
	MarketAppRequest,
	MarketData
} from 'src/constant/constants';
import { useCenterStore } from 'src/stores/market/center';

export async function fetchMarketData(): Promise<MarketData | null> {
	try {
		const store = useCenterStore();
		const url = store.appUrl + '/market/data';
		const { data } = await axios.get<MarketData>(url);
		console.log(data);

		return data;
	} catch (e) {
		console.error(e);
		return null;
	}
}

export async function getMarketDataHash(): Promise<{ hash: string }> {
	try {
		const store = useCenterStore();
		const url = store.appUrl + '/market/hash';
		const { data } = await axios.get<{ hash: string }>(url);
		return data;
	} catch (e) {
		console.error(e);
		return { hash: '' };
	}
}

export async function getMarketApps(
	apps: MarketAppRequest[]
): Promise<AppFullInfoLatest[]> {
	const store = useCenterStore();
	const url = store.appUrl + '/apps';

	const { data }: any = await axios.post<AppFullInfoLatest[]>(url, {
		apps
	});

	return data;
}

export async function getSystemStatus(): Promise<any> {
	try {
		const store = useCenterStore();
		const url = store.appUrl + '/settings/system-status';

		const { data } = await axios.get(url);

		return data;
	} catch (e: any) {
		console.error(e);
		return null;
	}
}

export async function getMarketState(): Promise<MarketData | null> {
	try {
		const store = useCenterStore();
		const url = store.appUrl + '/market/state';

		const { data }: any = await axios.get<MarketData>(url);

		return data;
	} catch (e) {
		console.error(e);
		return null;
	}
}
