import { defineStore } from 'pinia';
import { AccessLevel } from 'src/constant/global';
import axios from 'axios';
import { useTokenStore } from './token';
import { i18n } from 'src/boot/i18n';
import { notifySuccess } from 'src/utils/settings/btNotify';

export type RootState = {
	access_level: AccessLevel | null;
};

export const useAuthorityStore = defineStore('authority', {
	state: () => {
		return {
			access_level: {}
		} as RootState;
	},
	getters: {},
	actions: {
		async getIps() {
			const tokenStore = useTokenStore();

			const data: any = await axios.get(
				tokenStore.url + '/api/launcher-acc-policy'
			);
			this.access_level = data;
		},

		async setIps(req: AccessLevel) {
			const tokenStore = useTokenStore();

			const data: any = await axios.post(
				tokenStore.url + '/api/launcher-acc-policy',
				req
			);

			notifySuccess(i18n.global.t('success'));

			return data;
		}
	}
});
