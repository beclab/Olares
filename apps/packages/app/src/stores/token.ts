import axios from 'axios';
import { defineStore } from 'pinia';
import queryString from 'query-string';
import { Token, OlaresInfo, DeviceType } from '@bytetrade/core';
import { CurrentView } from 'src/utils/constants';
import { saltedMD5 } from './../utils/salted-md5';
import { BtNotify, NotifyDefinedType } from '@bytetrade/ui';
import { i18n } from 'src/boot/i18n';
import { getLoginResponseErrorMessage } from 'src/utils/interface/login';

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
			// const desktopURL = 'https://desktop.' + name;
			const urlParams = new URLSearchParams(window.location.search);
			let targetUrl = urlParams.get('redirect') || urlParams.get('rd');
			if (targetUrl) {
				try {
					targetUrl = decodeURIComponent(targetUrl);
				} catch (e) {
					console.error('Failed to decode redirect URL:', e);
				}
			}
			const targetURL = targetUrl ? targetUrl : this.desktopURL;
			return targetURL;
		},
		olaresId(): string {
			return this.user.olaresId || this.user.terminusName;
		},
		islocal(): boolean {
			return window.location.hostname.endsWith('olares.local');
		},
		desktopURL() {
			let url = window.location.protocol + '//' + 'desktop.';
			if (this.islocal) {
				url = url + this.olaresId.split('@')[0] + '.olares.local';
			} else {
				const name = this.olaresId.replace('@', '.');
				url = url + name;
			}
			return url;
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

			try {
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
			} catch (err) {
				//
				if (err.response) {
					const message =
						typeof err.response.data == 'string'
							? err.response.data
							: err.response.data?.message
							? err.response.data.message
							: '';
					if (message && typeof message == 'string') {
						BtNotify.show({
							type: NotifyDefinedType.MESSAGE,
							message: i18n.global.t(getLoginResponseErrorMessage(message))
						});
					}
				} else {
					BtNotify.show({
						type: NotifyDefinedType.MESSAGE,
						message: err.message
					});
				}
				throw err;
			}
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

		async refresh_token(logout: string | null = null) {
			if (logout) {
				localStorage.removeItem('auth_refresh_token');
				throw new Error('Logout');
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
			try {
				const data: Token = await axios.post(
					this.url + '/api/secondfactor/totp',
					{
						targetURL: this.target_url,
						token
					}
				);
				this.setToken(data);
				return data;
			} catch (err) {
				if (err.response) {
					const message =
						typeof err.response.data == 'string'
							? err.response.data
							: err.response.data?.message
							? err.response.data.message
							: '';
					if (message && typeof message == 'string') {
						BtNotify.show({
							type: NotifyDefinedType.MESSAGE,
							message: i18n.global.t(getLoginResponseErrorMessage(message))
						});
					}
				} else {
					BtNotify.show({
						type: NotifyDefinedType.MESSAGE,
						message: err.message
					});
				}
				throw err;
			}
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
