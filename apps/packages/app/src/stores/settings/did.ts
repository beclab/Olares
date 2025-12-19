import { defineStore } from 'pinia';
import { ResolutionResult, GolbalHost } from '@bytetrade/core';
import axios, { AxiosInstance } from 'axios';
import { useTokenStore } from './token';
import {
	getSettingsServerMdnsRequestApi,
	MdnsApiEmum
} from 'src/services/abstractions/mdns/service';
import { notifyFailed } from 'src/utils/notifyRedefinedUtil';
import { i18n } from 'src/boot/i18n';

export type SSIState = {
	// did_instance: AxiosInstance | undefined;
};

export const useDIDStore = defineStore('did', {
	state: () => {
		return {
			// did_instance: undefined
		} as SSIState;
	},
	getters: {},
	actions: {
		async resolve_name_by_did(name: string): Promise<ResolutionResult | null> {
			try {
				const get_name_response = await this.getDidInstanceByName(name).get(
					'/1.0/name/' + name.replace('@', '.')
				);
				if (get_name_response && get_name_response.status == 200) {
					return get_name_response.data;
				}
				notifyFailed(
					i18n.global.t('errors.olares_id_not_exists_on_blockchain')
				);
				return null;
			} catch (err) {
				if (
					err.response &&
					err.response.data.message == 'Failed to resolve DID'
				) {
					notifyFailed(
						i18n.global.t('errors.olares_id_not_exists_on_blockchain')
					);
				} else if (err.response && err.response.data.message) {
					notifyFailed(err.response.data.message);
				} else {
					notifyFailed(err);
				}
				return null;
			}
		},
		async resolve_name(name: string): Promise<ResolutionResult | null> {
			try {
				const get_name_response = await this.getDidInstance().get(
					getSettingsServerMdnsRequestApi(MdnsApiEmum.DID_USER_NAME) +
						'/' +
						name.replace('@', '.')
				);
				if (get_name_response && get_name_response.status == 200) {
					return get_name_response.data;
				}
				notifyFailed(
					i18n.global.t('errors.olares_id_not_exists_on_blockchain')
				);
				return null;
			} catch (err) {
				if (
					err.response &&
					err.response.data.message == 'Failed to resolve DID'
				) {
					notifyFailed(
						i18n.global.t('errors.olares_id_not_exists_on_blockchain')
					);
				} else if (err.response && err.response.data.message) {
					notifyFailed(err.response.data.message);
				} else {
					notifyFailed(err);
				}
				return null;
			}
		},
		getDidInstance() {
			const tokenStore = useTokenStore();
			return axios.create({
				baseURL: tokenStore.url || '',
				timeout: 10000,
				headers: {
					'Content-Type': 'application/json'
				}
			});
		},
		getDidInstanceByName(name: string) {
			const tokenStore = useTokenStore();
			return axios.create({
				baseURL:
					GolbalHost.DID_GATE_URL[GolbalHost.userNameToEnvironment(name)],
				timeout: 10000,
				headers: {
					'Content-Type': 'application/json'
				}
			});
		}
	}
});
