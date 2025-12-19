import { useCenterStore } from 'src/stores/market/center';
import { MarketSource } from 'src/constant/constants';
import axios from 'axios';
import globalConfig from 'src/api/market/config';
import { GolbalHost } from '@bytetrade/core';

export interface MarketRequest {
	id: string;
	name: string;
	base_url: string;
	type: string;
	description: string;
}

export async function getMarketSource(): Promise<MarketSource[]> {
	if (globalConfig.isOfficial) {
		return [
			{
				base_url: GolbalHost.MARKET_PROVIDER.en,
				description: 'Official market source for app store applications',
				id: 'market.olares',
				is_active: false,
				name: 'market.olares',
				priority: 0,
				type: 'remote',
				updated_at: '0001-01-01T00:00:00Z'
			}
		];
	}
	const store = useCenterStore();
	const url = store.appUrl + '/settings/market-source';
	const { data } = await axios.get(url);
	console.log(data);
	return data ?? [];
}

export async function addMarketSource(
	request: MarketRequest
): Promise<MarketSource[]> {
	const store = useCenterStore();
	const url = store.appUrl + '/settings/market-source';
	const { data } = await axios.post(url, request);
	console.log(data);
	return data;
}

export async function deleteMarketSource(sourceId: string): Promise<any> {
	const store = useCenterStore();
	const url = store.appUrl + `/settings/market-source/${sourceId}`;
	const { data } = await axios.delete(url);
	console.log(data);
	return data;
}
