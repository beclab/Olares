import axios from 'axios';
import { defineStore } from 'pinia';
import { useTokenStore } from './token';
import { i18n } from 'src/boot/i18n';
import { notifySuccess } from 'src/utils/settings/btNotify';

export const useSecretStore = defineStore('secret', {
	state: () => ({}),

	getters: {},

	actions: {
		async checkSecretPermission(app_name: string | undefined) {
			if (!app_name) {
				return null;
			}
			const tokenStore = useTokenStore();

			const data: any = await axios.get(
				`${tokenStore.url}/admin/permission/${app_name}`
			);
			return data;
		},

		async getSecrets(app_name: string | undefined) {
			if (!app_name) {
				return null;
			}
			const tokenStore = useTokenStore();

			const data: any = await axios.get(
				`${tokenStore.url}/admin/secret/${app_name}`
			);
			return data;
		},

		async newSecret(app_name: string | undefined, key: string, value: string) {
			if (!app_name) {
				return false;
			}
			const tokenStore = useTokenStore();
			const data: any = await axios.post(
				`${tokenStore.url}/admin/secret/${app_name}`,
				{
					Key: key,
					Value: value
				}
			);
			notifySuccess(i18n.global.t('success'));
			return data;
		},
		async updateSecret(
			app_name: string | undefined,
			key: string,
			value: string
		) {
			if (!app_name) {
				return false;
			}
			const tokenStore = useTokenStore();
			const data: any = await axios.put(
				`${tokenStore.url}/admin/secret/${app_name}`,
				{
					Key: key,
					Value: value
				}
			);

			notifySuccess(i18n.global.t('success'));

			return data;
		},
		async deleteSecret(app_name: string | undefined, key: string) {
			if (!app_name) {
				return false;
			}
			const tokenStore = useTokenStore();
			const data: any = await axios.delete(
				`${tokenStore.url}/admin/secret/${app_name}`,
				{
					data: {
						Key: key
					}
				}
			);
			notifySuccess(i18n.global.t('success'));
			return data;
		}
	}
});
