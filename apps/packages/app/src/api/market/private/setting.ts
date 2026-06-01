import axios from 'axios';
import globalConfig from 'src/api/market/config';
import { useAppStore } from 'src/stores/market/appStore';

export async function getSettingConfig(): Promise<any> {
	if (globalConfig.isOfficial) {
		return { selected_source: 'market.olares' };
	}
	const store = useAppStore();
	const url = store.appUrl + '/settings/market-settings';
	const { data } = await axios.get(url);
	console.log(data);
	return data;
}

export async function updateSettingConfig(
	nsfw: boolean,
	sourId: string
): Promise<boolean> {
	const store = useAppStore();
	const url = store.appUrl + '/settings/market-settings';
	const { data } = await axios.put(url, {
		nsfw,
		selected_source: sourId
	});
	console.log(data);
	return data;
}
