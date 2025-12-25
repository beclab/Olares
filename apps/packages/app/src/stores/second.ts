import { defineStore } from 'pinia';
import { axiosInstanceProxy } from '../platform/httpProxy';

export type SecondVerifyData = {
	show_twofa: boolean;
	twofa_ob: any;
};

export const useSecondVerifyStore = defineStore('secondVerify', {
	state: () => {
		return {
			show_twofa: false,
			twofa_ob: null
		} as SecondVerifyData;
	},
	getters: {},
	actions: {
		clearTwoFAData() {
			this.show_twofa = false;
			this.twofa_ob = null;
		},

		clearTwoFAOb() {
			this.twofa_ob = null;
		},

		setTwoFADialog(data: any) {
			this.show_twofa = true;
			this.twofa_ob = data;
		},

		async cert_secondfactor_totp(
			token: string,
			responseURL: string
		): Promise<any> {
			const baseURL = window.location.origin.replace(/files/g, 'auth');
			const instance = axiosInstanceProxy({
				baseURL: baseURL,
				timeout: 10000
			});

			const data: any = await instance.post(
				'/api/secondfactor/totp',
				{
					targetURL: responseURL,
					token
				},
				{
					withCredentials: true
				}
			);
			return data;
		}
	}
});
