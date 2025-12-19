import { defineStore } from 'pinia';
import { useUserStore } from './user';
import axios from 'axios';
import { Submission, PrivateJwk, GolbalHost } from '@bytetrade/core';
import { UserItem } from '@didvault/sdk/src/core';
import { v4 as uuidv4 } from 'uuid';
import { getPrivateJWK, getDID } from '../did/did-key';
import { signJWS } from '../layouts/dialog/sign';
import { i18n } from 'src/boot/i18n';

export type CloudState = {
	//
	domain_init: boolean;
	domains: Domain[];
};
export interface TokenData {
	userid: string;
	token: string;
	expired: number;
	emailEnable?: boolean;
	emailMask?: string;
	phoneEnable?: boolean;
	phoneMask?: string;
	gaEnable?: boolean;
}

export interface Domain {
	id: number;
	userid: string;
	domain: string;
	status: string;
	terminusName?: string;
	ethAddress?: string;
	did?: string;
	confirmTransaction?: string;
	resource_id?: string;
	vc?: string;
	fullString?: string;
	createAt?: string;
	updateAt?: string;
	rule_type?: string;
	rule_data?: string;
	txt?: {
		name: string;
		value: string;
	};
	ns?: string;
}

export interface Domain {
	userid: string;
	domain: string;
	status: string;
	terminusName?: string;
	ethAddress?: string;
	did?: string;
	confirmTransaction?: string;
	resource_id?: string;
	vc?: string;
	fullString?: string;
	createAt?: string;
	updateAt?: string;
	rule_type?: string;
	rule_data?: string;
}

export interface DomainUser {
	userid: string;
	domain: string;
	status: string;
	terminusName?: string;
	ethAddress?: string;
	did?: string;
	confirmTransaction?: string;
	resource_id?: string;
	vc?: string;
	fullString?: string;
	createAt?: string;
	updateAt?: string;
	rule_type?: string;
	rule_data?: string;
}

export interface DomainRule {
	ruleType: string;
	ruleData: string;
}

export enum DOMAIN_STATUS {
	WAIT_TXT_RESOLVE = 'WAIT_TXT',
	WAIT_NS_RESOLVE = 'WAIT_NS',
	WAIT_REQUEST_VC = 'WAIT_REQUEST_VC',
	WAIT_SUBMIT_VP = 'WAIT_SUBMIT_VP',
	WAIT_RULE_SET = 'WAIT_RULE_SET',
	BIND = 'BIND',
	BINDING = 'BINDING',
	ALLOCATED = 'ALLOCATED'
}

export function getDomainDesc(status: string) {
	switch (status) {
		case DOMAIN_STATUS.WAIT_TXT_RESOLVE:
			return i18n.global.t('domain.wait_txt_resolve_desc');
		case DOMAIN_STATUS.WAIT_NS_RESOLVE:
			return i18n.global.t('domain.wait_ns_resolve_desc');
		case DOMAIN_STATUS.WAIT_REQUEST_VC:
			return i18n.global.t('domain.wait_request_vc_desc');
		case DOMAIN_STATUS.WAIT_SUBMIT_VP:
			return i18n.global.t('domain.wait_submit_vp_desc');
		case DOMAIN_STATUS.WAIT_RULE_SET:
			return i18n.global.t('wait_rule_set_desc');
		case DOMAIN_STATUS.BIND:
			return i18n.global.t('domain.bind_desc');
		case DOMAIN_STATUS.BINDING:
			return i18n.global.t('domain.bind_desc');
		case DOMAIN_STATUS.ALLOCATED:
			return i18n.global.t('domain.allocated_desc');
	}

	return status;
}

export const useCloudStore = defineStore('cloud', {
	state: () => {
		return {
			url: 'https://cloud.localhost',
			instance: undefined,
			domain_init: false,
			domains: []
		} as CloudState;
	},

	getters: {
		//
	},

	actions: {
		async autoLogin() {
			const userStore = useUserStore();
			if (!userStore.current_user) {
				return null;
			}
			if (!userStore.current_mnemonic) {
				return null;
			}
			if (
				!userStore.current_user.cloud_id ||
				!userStore.current_user.cloud_token ||
				userStore.current_user.cloud_expired < new Date().getTime() - 1000 * 60
			) {
				const did = await getDID(userStore.current_mnemonic.mnemonic);
				const privateJWK: PrivateJwk | undefined = await getPrivateJWK(
					userStore.current_mnemonic.mnemonic
				);
				const body = {
					did: did,
					secret: (await uuidv4()).replaceAll('-', ''),
					time: new Date().getTime()
				};

				const jws = await signJWS(did, body, privateJWK);

				const response: any = await axios.post(
					this.getUrl() + '/v2/user/login',
					{
						id: new Date().getTime(),
						jws,
						did,
						body
					}
				);

				if (response.code != 200 && response.code != 0) {
					throw Error('login failure');
				}

				const result: any = await axios.post(
					this.getUrl() + '/v2/user/activeLogin',
					{
						secret: body.secret
					}
				);
				const loginToken: TokenData = result.data;

				if (result.code == 200) {
					const user: UserItem = userStore.users!.items.get(
						userStore.current_id!
					)!;
					user.cloud_id = loginToken.userid;
					user.cloud_token = loginToken.token;
					user.cloud_expired = loginToken.expired;
					await userStore.users!.items.update(user);
					await userStore.save();
				}
			}
		},
		async getDomainRule(domain: string): Promise<DomainRule> {
			const userStore = useUserStore();

			await this.autoLogin();
			const response: any = await axios.post(
				this.getUrl() + '/v1/domain/get_domain_rule',
				{
					domain: domain,
					userid: userStore.current_user?.cloud_id,
					token: userStore.current_user?.cloud_token
				}
			);
			if (response.code != 200 && response.code != 0) {
				throw Error(response.message);
			}
			return response.data;
		},

		async requestDomainVC(jws: string, domain: string) {
			const userStore = useUserStore();
			if (!userStore.current_user) {
				return null;
			}
			await this.autoLogin();
			const response: any = await axios.post(
				this.getUrl() + '/v1/domain/requestDomainVC',
				{
					jws: jws,
					domain: domain,
					userid: userStore.current_user.cloud_id,
					token: userStore.current_user.cloud_token
				}
			);
			if (response.code != 200 && response.code != 0) {
				throw Error('get google result failure');
			}
			return response.data;
		},
		async verifyDomainPresentation(
			jws: string,
			domain: string
		): Promise<Submission | null> {
			const userStore = useUserStore();
			if (!userStore.current_user) {
				return null;
			}
			await this.autoLogin();
			const response: any = await axios.post(
				this.getUrl() + '/v1/domain/submitDomainVP',
				{
					userid: userStore.current_user.cloud_id,
					token: userStore.current_user.cloud_token,
					domain: domain,
					jws: jws
				}
			);
			if (response.code != 200 && response.code != 0) {
				throw Error(response.data.message || 'Submit Presentation Failure');
			}
			return response.data;
		},
		async selectDomain(domain: string): Promise<Domain | undefined> {
			const userStore = useUserStore();
			if (!userStore.current_user) {
				return;
			}

			if (!this.domain_init) {
				await this.getDomains();
			}

			return this.domains.find((item) => item.domain == domain);
		},
		async getDomains(): Promise<void> {
			const userStore = useUserStore();
			if (!userStore.current_user) {
				return;
			}
			await this.autoLogin();
			const response: any = await axios.post(
				this.getUrl() + '/v1/domain/domains',
				{
					userid: userStore.current_user.cloud_id,
					token: userStore.current_user.cloud_token
				}
			);
			this.domain_init = true;

			this.domains = response.data.domains;
		},

		async postRule(
			domain: string,
			ruleType: string,
			ruleData: string
		): Promise<void> {
			const userStore = useUserStore();
			if (!userStore.current_user) {
				return;
			}
			await this.autoLogin();
			await axios.post(this.getUrl() + '/v1/domain/set_domain_rule', {
				userid: userStore.current_user.cloud_id,
				token: userStore.current_user.cloud_token,
				domain,
				ruleType,
				ruleData
			});

			await this.getDomains();
		},

		async requestNFT(jws: string) {
			const userStore = useUserStore();
			if (!userStore.current_user) {
				return null;
			}
			await this.autoLogin();
			const response: any = await axios.post(
				this.getUrl() + '/v1/domain/requestNFT',
				{
					jws: jws,
					//domain: domain,
					userid: userStore.current_user.cloud_id,
					token: userStore.current_user.cloud_token
				}
			);
			if (response.code != 200 && response.code != 0) {
				throw Error('get google result failure');
			}
			return response.data;
		},

		async verifyNFTPresentation(jws: string): Promise<Submission | null> {
			const userStore = useUserStore();
			if (!userStore.current_user) {
				return null;
			}
			await this.autoLogin();
			const response: any = await axios.post(
				this.getUrl() + '/v1/domain/submitDomainVP',
				{
					userid: userStore.current_user.cloud_id,
					token: userStore.current_user.cloud_token,
					jws: jws
				}
			);
			if (response.code != 200 && response.code != 0) {
				throw Error(response.data.message || 'Submit Presentation Failure');
			}
			return response.data;
		},

		getUrl() {
			const userStore = useUserStore();
			if (userStore.defaultDomain == 'cn') {
				return GolbalHost.OLARES_SPACE_URL.cn;
			}
			return GolbalHost.OLARES_SPACE_URL.en;
		}
	}
});
