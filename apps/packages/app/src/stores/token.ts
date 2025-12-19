import axios from 'axios';
import { defineStore } from 'pinia';
import queryString from 'query-string';
import { Token, OlaresInfo, DeviceType } from '@bytetrade/core';
import { CurrentView } from 'src/utils/constants';
import { saltedMD5 } from './../utils/salted-md5';

export type RootState = {
	token: Token | null;
	url: string | null;
	user: OlaresInfo;
	currentView: string;
	requestTermiPass: boolean;
	deviceInfo: {
		device: DeviceType;
		isVerticalScreen: boolean;
	};
};

export const useTokenStore = defineStore('token', {
	state: () => {
		return {
			token: null,
			url: null,
			user: {},
			currentView: CurrentView.FIRST_FACTOR,
			requestTermiPass: true,
			deviceInfo: {
				device: DeviceType.DESKTOP,
				isVerticalScreen: false
			}
		} as RootState;
	},
	getters: {
		target_url(): string {
			const name = this.user.olaresId.replace('@', '.');
			const desktopURL = 'https://desktop.' + name;
			const urlParams = new URLSearchParams(window.location.search);
			let targetUrl = urlParams.get('redirect') || urlParams.get('rd');
			if (targetUrl) {
				try {
					targetUrl = decodeURIComponent(targetUrl);
				} catch (e) {
					console.error('Failed to decode redirect URL:', e);
				}
			}
			const targetURL = targetUrl ? targetUrl : desktopURL;
			return targetURL;
		},
		olaresId(): string {
			return this.user.olaresId || this.user.terminusName;
		}
	},
	actions: {
		async loadData() {
			const data: any = await axios.get(
				this.url + '/bfl/info/v1/olares-info',
				{}
			);
			this.user = data;
		},

		async login(username: string, password: string) {
			// const requestMethod = this.urlParams.rm;

			const saltedPassword = saltedMD5(password, {
				osVersion: this.user.osVersion
			});

			const data: Token = await axios.post(
				this.url + '/api/firstfactor',
				{
					username,
					password: saltedPassword,
					keepMeLoggedIn: false,
					// requestMethod,
					targetURL: this.target_url,
					requestTermiPass: this.requestTermiPass
				},
				{
					timeout: 30000
				}
			);
			this.setToken(data);

			return data;
		},

		async apiState(): Promise<boolean> {
			try {
				const data: any = await axios.get(this.url + '/api/state');
				if (data.authentication_level === 2) {
					return true;
				}
				return false;
			} catch (e) {
				return false;
			}
		},

		setToken(new_token: Token) {
			this.token = new_token;

			if (
				this.token &&
				this.token.refresh_token &&
				this.token.refresh_token.length > 0
			) {
				localStorage.setItem('auth_refresh_token', this.token.refresh_token);
			}
		},

		async refresh_token(logout: string | null = null, fa2 = false) {
			if (logout) {
				localStorage.removeItem('auth_refresh_token');
				throw new Error('Logout');
			}

			if (fa2) {
				throw new Error('Need fa2');
			}

			const refresh_token = localStorage.getItem('auth_refresh_token');
			if (!refresh_token) {
				throw new Error('No refresh token found');
			}

			try {
				const data: any = await axios.post(this.url + '/api/refresh', {
					refreshToken: refresh_token
				});

				if (!data || data.length == 0) {
					localStorage.removeItem('auth_refresh_token');
					throw new Error('Invalid response from refresh token endpoint');
				}

				return true;
			} catch (e) {
				localStorage.removeItem('auth_refresh_token');
				throw new Error('Failed to refresh token: ' + e);
			}
		},

		setUrl(new_url: string | null) {
			this.url = new_url;
		},

		async secondFactor(token: string): Promise<Token> {
			const data: Token = await axios.post(
				this.url + '/api/secondfactor/totp',
				{
					targetURL: this.target_url,
					token
				}
			);
			this.setToken(data);
			return data;
		},

		async replaceToDesktopUrl(redirect?: string): Promise<void> {
			if (typeof window === 'undefined') return;

			if (redirect) {
				window.location.replace(redirect);
			} else {
				window.location.replace(this.target_url);
			}
		}
	}
});
