import axios from 'axios';
import { useUserStore } from 'src/stores/user';
import { AbilityData } from 'src/core/abilities';

export async function getAppAbilities(): Promise<AbilityData> {
	const userStore = useUserStore();
	let url = userStore.getModuleSever('settings');
	if (process.env.IS_DEV) {
		url = '';
	}
	return await axios.get(url + '/api/abilities');
}
