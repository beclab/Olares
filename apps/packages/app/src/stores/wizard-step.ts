import { defineStore } from 'pinia';
import axios from 'axios';
import { i18n } from '../boot/i18n';
import { Token, Encoder, OlaresInfo, GolbalHost } from '@bytetrade/core';
import { saltedMD5 } from './../utils/salted-md5';
import { WizardInfo } from 'src/utils/interface/wizard';
import { OlaresTunneV2Interface } from 'src/utils/interface/frp';

// export interface SystemOption {
// 	location: string;
// 	language: string;
// }

export interface CloudflareNetworkOption {
	enable_tunnel: boolean;
	external_ip: string | null;
}

// export interface WizardInfo {
// 	step: number;

// 	username: string | null;
// 	password: string | null;
// 	//access_token: string;
// 	url: string;
// 	// lang: string | null;
// 	// timezone: string | null;

// 	//did: string | null;

// 	system: SystemOption;
// 	network: CloudflareNetworkOption;
// }

export type RootState = {
	token: Token | null;
	url: string | null;
	user: OlaresInfo;
	pingResult: boolean;
	raw_login_loading: boolean;
	wizard: WizardInfo;
	frpList: OlaresTunneV2Interface[];
	selectedFrp: string;
};

function stringToIntHash(str: string, lowerbound: number, upperbound: number) {
	if (!str) {
		return lowerbound;
	}

	let result = 0;
	for (let i = 0; i < str.length; i++) {
		result = result + str.charCodeAt(i);
	}

	if (!lowerbound) lowerbound = 0;
	if (!upperbound) upperbound = 500;

	return (result % (upperbound - lowerbound)) + lowerbound;
}

export const useTokenStore = defineStore('token', {
	state: () => {
		return {
			token: null,
			url: null,
			user: {},
			pingResult: false,
			raw_login_loading: false,
			wizard: {},
			frpList: [] as OlaresTunneV2Interface[],
			selectedFrp: ''
		} as RootState;
	},
	getters: {
		step(): number {
			return this.wizard.step;
		},
		olaresId(): string {
			return this.user?.olaresId || this.user?.terminusName || '';
		},
		get_terminus_url(): string | undefined {
			if (!this.user) {
				return undefined;
			}

			const res = 'https://desktop.' + this.user.olaresId.replace('@', '.');
			return res;
		},
		get_auth_url(): string | undefined {
			if (!this.user) {
				return undefined;
			}

			const res = 'https://auth.' + this.user.olaresId.replace('@', '.');
			return res;
		},
		avatar_url(): string {
			if (!this.user || !this.user.olaresId) {
				return 'https://app.cdn.olares.com/avatar3/1.png';
			}

			if (!this.user.avatar) {
				const id = stringToIntHash(this.user.olaresId, 1, 36);

				return `https://app.cdn.olares.com/avatar3/${id}.png`;
			}

			if (this.user.avatar.startsWith('http')) {
				return this.user.avatar;
			} else {
				const re = new RegExp('^[1-3]?[0-9]\\.png');
				if (re.test(this.user.avatar)) {
					console.log('re test true');

					return 'https://app.cdn.olares.com/avatar3/' + this.user.avatar;
				} else {
					try {
						const vp = JSON.parse(this.user.avatar);
						console.log(vp);
						if (vp) {
							const vcstr = Encoder.bytesToString(
								Encoder.base64UrlToBytes(
									vp.verifiableCredential![0].split('.')[1]
								)
							);
							console.log(vcstr);
							const vc = JSON.parse(vcstr);
							console.log(vc);
							console.log(vc.vc.credentialSubject.image);
							let imageUrl = vc.vc.credentialSubject.image;
							if (imageUrl.startsWith('ipfs://')) {
								imageUrl = imageUrl.replace(
									'ipfs://',
									'https://gateway.ipfs.io/ipfs/'
								);
							}

							console.log(imageUrl);
							return imageUrl;
						} else {
							return 'https://app.cdn.olares.com/avatar3/1.png';
						}
					} catch (e) {
						console.log(e);
						return 'https://app.cdn.olares.com/avatar3/1.png';
					}
				}
			}
		}
	},

	actions: {
		async loadData(isDefault = false) {
			try {
				const data: any = await axios.get(
					this.url + '/bfl/info/v1/olares-info',
					{}
				);
				if (data.wizardStatus) {
					this.user = data;
				} else {
					console.error(data);
				}
			} catch (e) {
				if (isDefault && e.message == 'Default') {
					this.user.wizardStatus = 'wait_reset_password';
				}
				console.log(e);
			}
		},

		async loadIP() {
			try {
				const data: any = await axios.get(this.url + '/bfl/backend/v1/ip');
				const external = data.masterExternalIP;
				this.wizard.network.external_ip = external;
			} catch (e) {
				console.log(e);
			}
		},

		async getFrpList() {
			if (this.frpList.length > 0) {
				return;
			}
			const frpBaseUrl =
				GolbalHost.FRP_LIST_URL[
					GolbalHost.userNameToEnvironment(this.olaresId)
				];
			// const url = frpBaseUrl.endsWith('/')
			// 	? frpBaseUrl + 'v2/servers'
			// 	: frpBaseUrl + '/v2/servers';
			// console.log('frpUrl ===>', url);
			// const axios = axios.
			const instance = axios.create({
				baseURL: frpBaseUrl,
				timeout: 1000 * 10,
				headers: {}
			});
			const response = await instance.post('/v2/servers', {
				name: this.olaresId
			});
			if (response.status == 200) {
				this.frpList = response.data;
				if (this.frpList.length > 0) {
					this.selectedFrp = this.frpList[0].machine[0].host;
				}
			}
		},

		async ping2(): Promise<boolean> {
			if (!this.get_auth_url) {
				return false;
			}
			try {
				const data: any = await axios.get(
					this.get_auth_url + '/bfl/info/v1/olares-info'
				);
				if (data.wizardStatus) {
					this.user = data;
					return true;
				} else {
					console.error(data);
					return false;
				}
			} catch (e) {
				this.pingResult = false;
				return false;
			}
		},

		async raw_login(username: string, password: string) {
			const local_username = username.split('@')[0];

			const saltedPassword = saltedMD5(password, {
				osVersion: this.user.osVersion
			});

			const data: Token = await axios.post(this.url + '/api/firstfactor', {
				username: local_username,
				password: saltedPassword,
				keepMeLoggedIn: false,
				requestMethod: 'POST',
				targetURL: '',
				acceptCookie: false
			});

			this.setToken(data);
		},

		setToken(new_token: Token | null) {
			if (new_token) {
				this.token = new_token;
			} else {
				this.token = null;
			}
		},

		setUrl(new_url: string | null) {
			this.url = new_url;
		},
		async loadWizard() {
			this.setWizard({
				step: 1,
				username: '',
				password: '',
				url: '',
				system: {
					language: i18n.global.locale.value,
					location: 'Singapore',
					theme: 'light',
					frp: {
						host: '',
						jws: ''
					}
				},
				network: {
					enable_tunnel: false,
					external_ip: null
				}
			});
		},

		setWizard(wizard: WizardInfo) {
			this.wizard = wizard;
		},

		setStep(step: number) {
			this.wizard.step = step;
			this.setWizard(this.wizard);
		},

		olaresTunnelsV2Options() {
			return this.frpList.map((item: OlaresTunneV2Interface) => {
				const label = item.name[i18n.global.locale.value]
					? item.name[i18n.global.locale.value]
					: item.name['en-US']
					? item.name['en-US']
					: '';
				return {
					label: label,
					value: item.machine.length > 0 ? item.machine[0].host : '',
					enable: true
				};
			});
		}
	}
});
