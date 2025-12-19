import { OlaresInfo, TerminusInfo, Token } from '@bytetrade/core';
import { axiosInstanceProxy } from 'src/platform/httpProxy';
import { i18n } from '../boot/i18n';
import { saltedMD5 } from './salted-md5';
// import axios from 'axios';

export const onFirstFactor = async (
	baseURL: string,
	terminus_name: string,
	osUser: string,
	osPwd: string,
	acceptCookie = true,
	needTwoFactor = false,
	olaresVersion?: string
): Promise<Token> => {
	const instanceProxy = axiosInstanceProxy({
		baseURL: baseURL,
		timeout: 1000 * 10,
		headers: {
			'Content-Type': 'application/json'
		},
		withCredentials: true
	});

	let targetURL =
		'https://vault.' + terminus_name.replace('@', '.') + '/' + 'server';

	if (needTwoFactor) {
		targetURL = 'https://desktop.' + terminus_name.replace('@', '.') + '/';
	}

	try {
		const response = await instanceProxy.post(
			'/api/firstfactor',
			{
				username: osUser,
				password: passwordAddSort(osPwd, olaresVersion),
				keepMeLoggedIn: false,
				requestMethod: 'POST',
				targetURL,
				acceptCookie
			},
			{
				params: {
					hideCookie: true
				}
			}
		);

		if (!response || response.status != 200 || !response.data) {
			throw Error('Network error, please try again later');
		}
		if (response.data.status != 'OK') {
			throw new Error('Password Error');
		}
		const token = response.data.data;
		if (!token) {
			throw new Error('Password Error');
		}

		return token;
	} catch (error) {
		if (error.response) {
			if (error.response.data && error.response.data.message) {
				throw new Error(error.response.data.message);
			}
			throw new Error(error.message);
		}
		throw error;
	}
};

export interface SSOTokenRaw {
	exp: number;
	iat: number;
	iss: string;
	sub: string;
	token_type: string;
	username: string;
	extra: {
		uninitialized: string[];
	};
}

export const refresh_token = async (
	baseURL: string,
	refreshToken: string,
	access_token: string
) => {
	const instanceProxy = axiosInstanceProxy({
		baseURL: baseURL,
		timeout: 1000 * 10,
		headers: {
			'Content-Type': 'application/json',
			'X-Authorization': access_token
		},
		withCredentials: true
	});

	const response = await instanceProxy.post('/api/refresh', {
		refreshToken
	});

	if (!response || response.status != 200 || !response.data) {
		throw Error(i18n.global.t('errors.network_error_please_try_again_later'));
	}
	if (response.data.status != 'OK') {
		// throw Error('Network error, please try again later')
		throw new Error('refresh token error');
	}
	// const token = response.data.data;
	// if (!token) {
	// 	throw new Error('Password Error');
	// }

	return response.data.data;
};

export const reset_password = async (
	baseURL: string,
	localName: string,
	current_password: string,
	newPassword: string,
	access_token: string,
	olaresVersion?: string
) => {
	const instanceProxy = axiosInstanceProxy({
		baseURL: baseURL,
		timeout: 1000 * 10,
		headers: {
			'Content-Type': 'application/json',
			'X-Authorization': access_token
		}
	});
	const response = await instanceProxy.put(
		'/bfl/iam/v1alpha1/users/' + localName + '/password',
		{
			current_password: passwordAddSort(current_password, olaresVersion),
			password: passwordAddSort(newPassword, olaresVersion)
		}
	);

	if (!response || response.status != 200 || !response.data) {
		throw Error(i18n.global.t('errors.network_error_please_try_again_later'));
	}

	if (response.data.code != 0) {
		if (response.data.message) {
			throw Error(response.data.message);
		}
		throw Error('Network error, please try again later');
	}

	return response.data;
};

export const getTerminusInfo = async (
	baseURL: string,
	timeout = 5000,
	params?: any
) => {
	if (baseURL.endsWith('/server')) {
		baseURL = baseURL.substring(0, baseURL.length - 7);
	}
	const instance = axiosInstanceProxy({
		headers: {
			'Content-Type': 'application/json'
		},
		timeout,
		params: params
	});

	const res = await instance.get(
		(baseURL.endsWith('/') ? baseURL : baseURL + '/') +
			'bfl/info/v1/terminus-info'
	);
	const terminusInfo: OlaresInfo = res.data.data;
	return terminusInfo;
};

export const getOlaresInfo = async (
	baseURL: string,
	timeout = 5000,
	params?: any
) => {
	try {
		if (baseURL.endsWith('/server')) {
			baseURL = baseURL.substring(0, baseURL.length - 7);
		}
		const instance = axiosInstanceProxy({
			headers: {
				'Content-Type': 'application/json'
			},
			timeout,
			params: params
		});

		const res = await instance.get(
			(baseURL.endsWith('/') ? baseURL : baseURL + '/') +
				'bfl/info/v1/olares-info'
		);
		const terminusInfo: OlaresInfo = res.data.data;
		return terminusInfo;
	} catch (error) {
		return await getTerminusInfo(baseURL, timeout, params);
	}
};

export const passwordAddSort = (password: string, olaresVersion?: string) => {
	return saltedMD5(password, {
		osVersion: olaresVersion
	});
};
