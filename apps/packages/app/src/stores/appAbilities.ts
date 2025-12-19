import { defineStore } from 'pinia';
import { AbilityData } from 'src/core/abilities';
import { getAppAbilities } from 'src/api/settings/ability';
import { useUserStore } from './user';
import { busEmit } from 'src/utils/bus';

interface AppAbilitiesStore {
	data: AbilityData;
	loading: boolean;
}
const WISE_NAME = 'wise';
const TRANSLATE_NAME = 'translate';
const YTDLP_NAME = 'ytdlp';
type APP_KEYS = typeof WISE_NAME | typeof TRANSLATE_NAME | typeof YTDLP_NAME;

const defaultData = {
	vault: false,
	wise: {
		running: false,
		id: '',
		url: '',
		name: WISE_NAME,
		title: 'Wise'
	},
	translate: {
		running: false,
		id: '',
		url: '',
		name: TRANSLATE_NAME,
		title: 'MTranServer'
	},
	ytdlp: {
		running: false,
		id: '',
		url: '',
		name: YTDLP_NAME,
		title: 'YT-DLP'
	}
};

export const useAppAbilitiesStore = defineStore('appAbilities', {
	state: (): AppAbilitiesStore => ({
		data: { ...defaultData },
		loading: false
	}),
	getters: {
		vault: (state) => state.data?.vault || defaultData.vault,
		wise: (state) => ({ ...defaultData.wise, ...state.data?.wise }),
		translate: (state) => ({
			...defaultData.translate,
			...state.data?.translate
		}),
		ytdlp: (state) => ({ ...defaultData.ytdlp, ...state.data?.ytdlp })
	},
	actions: {
		async init() {
			this.data = { ...defaultData };
			return this.refreshAppAbilities();
		},
		async refreshAppAbilities() {
			this.loading = true;
			try {
				const res = await getAppAbilities();
				this.data = res;
				this.busEmitAbilityUpdate();
			} catch (error) {
				console.error('getAppAbilities err:', error);
				return Promise.reject(error);
			}
			this.loading = false;
			return Promise.resolve(this.data);
		},
		async busEmitAbilityUpdate() {
			busEmit('appAbilitiesUpdate');
		},
		getAppDomain(key: APP_KEYS) {
			const userStore = useUserStore();
			return userStore.getModuleSever(this.data[key].id);
		}
	}
});
