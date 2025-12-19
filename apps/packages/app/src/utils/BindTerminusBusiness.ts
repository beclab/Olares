import {
	AuthPurpose,
	AuthType,
	Err,
	ErrorCode,
	UserItem,
	MnemonicItem
} from '@didvault/sdk/src/core';
import { OlaresInfo, TerminusDefaultDomain, Token } from '@bytetrade/core';
import { onFirstFactor } from './account';
import { app, clearSenderUrl, setSenderUrl } from '../globals';
import { _authenticate } from '@didvault/sdk/src/authenticate';
import queryString from 'query-string';
import { TerminusInfo } from '@bytetrade/core';
import { useSSIStore } from '../stores/ssi';
import { useUserStore } from '../stores/user';
import { useBexStore } from '../stores/bex';
import { getDID } from '../did/did-key';
import { axiosInstanceProxy } from '../platform/httpProxy';
import { BusinessAsyncCallback, BusinessCallback } from '../utils/Callback';
import { base32ToBytes, hotp, VaultType } from '@didvault/sdk/src/core';
import { getVaultsByType } from './terminusBindUtils';
import axios from 'axios';
import { getAppPlatform } from 'src/application/platform';
import { i18n } from '../boot/i18n';
import UnlockAppDialog from '../components/unlock/UnlockAppDialog.vue';
import UserStatusCommonDialog from '../components/userStatusDialog/UserStatusCommonDialog.vue';

export const userBindTerminus = async (
	user: UserItem,
	mnemonic: MnemonicItem,
	host: string,
	osPwd: string,
	osVersion: string,
	callback: BusinessCallback
) => {
	if (!user) {
		throw new Error(i18n.global.t('errors.user_is_empty'));
	}

	if (!user.name) {
		throw new Error(i18n.global.t('errors.olares_id_is_empty'));
	}

	const userStore = useUserStore();
	userStore.temp_url = host;

	try {
		let baseURL = host;
		if (process.env.IS_PC_TEST) {
			baseURL = process.env.BFL_URL!;
		} else {
			if (baseURL.endsWith('/server')) {
				baseURL = baseURL.substring(0, baseURL.length - 7);
			}
		}

		const token: Token = await onFirstFactor(
			baseURL,
			user.name,
			user.local_name,
			osPwd,
			false,
			undefined,
			osVersion
		);

		setSenderUrl({
			url: host
		});

		await app.load(userStore.current_id!);
		await app.unlock(mnemonic.mnemonic, false);

		await app.clearSession();
		app.state._errors = [];

		const authRes = await _authenticate({
			did: user.local_name,
			type: AuthType.SSI,
			purpose: AuthPurpose.Signup,
			caller: 'E001'
		});

		if (authRes == null) {
			throw new Err(
				ErrorCode.AUTHENTICATION_FAILED,
				i18n.global.t('errors.authentication_failed')
			);
		}

		const masterPassword = mnemonic.mnemonic;
		let { url } = queryString.parseUrl(host);
		if (url) {
			if (url.startsWith('http://')) {
				url = url.substring(7);
			} else if (url.startsWith('https://')) {
				url = url.substring(8);
			}
		}
		const domain = url;
		const jws = await userStore.signJWS({
			name: user.name,
			did: user.id,
			domain: domain,
			time: '' + new Date().getTime()
		});
		if (!jws) {
			throw new Err(
				ErrorCode.JWS_AUTH_FAILED,
				i18n.global.t('errors.jws_auth_failed')
			);
		}

		await app.signup({
			did: authRes.did,
			masterPassword: masterPassword,
			name: user.name,
			authToken: authRes.token,
			sessionId: token.session_id,
			bflToken: token.access_token,
			bflUser: user.local_name,
			jws: jws
		});
		await app.synchronize();

		user.access_token = token.access_token;
		user.refresh_token = token.refresh_token;
		user.session_id = token.session_id;
		//user.fa2 = data.fa2;
		user.ip = userStore.temp_url!;
		user.url = userStore.temp_url!;
		user.binding = false;

		setSenderUrl({
			url: user.url
		});

		userStore.users!.items.update(user);
		await userStore.save();
		userStore.resetCurrentUserData();

		app.state._errors = [];
		if (process.env.PLATFORM === 'BEX' && userStore.current_id) {
			const bexStore = useBexStore();
			await bexStore.controller.changeAccount(userStore.current_id);
		}

		callback.onSuccess('');
	} catch (e) {
		if (e && e.code && e.code == ErrorCode.DOMAIN_RESOLUTION_ERROR) {
			callback.onFailure('');
			getAppPlatform()
				.getQuasar()
				?.dialog({
					component: UserStatusCommonDialog,
					componentProps: {
						title: i18n.global.t('Domain resolution error'),
						message: i18n.global.t(
							'Unable to resolve the Olares domain ({domain}) due to an FRP configuration conflict. To fix this issue, create a new Olares ID with the {domain2} domain and reactivate your device:<br>1. Tap the profile icon, select Add a new account, and create a new Olares ID.<br>2. When prompted to confirm the domain, tap Change default domain and select {domain2}.<br>3. Complete the Olares ID creation.<br>4. Reactivate your device using the new Olares ID.',
							{
								domain: userStore.current_user?.domain_name,
								domain2: 'olares.cn'
							}
						)
					}
				});
			return;
		}
		if (e.message) {
			callback.onFailure(e.message);
		} else {
			callback.onFailure(e);
		}
	}
};

export interface BindBusinessAsyncCallback {
	onSuccess(data: any): Promise<void>;
	needCheckFa(): void;
	onFailure(message: string): void;
}

export async function importUserByTerminusToken(
	terminusName: string,
	token: Token,
	mnemonic: string,
	use_local: boolean,
	callback: BusinessAsyncCallback
) {
	const userStore = useUserStore();
	const old_current_id = userStore.current_id;
	const did = await getDID(mnemonic);
	try {
		const user = await userStore.importUser(did!, terminusName, mnemonic);
		if (!user) {
			throw Error(i18n.global.t('errors.create_user_failed'));
		}
		await userStore.setCurrentID(user.id);
		let host = userStore.getModuleSever('vault');
		if (process.env.IS_PC_TEST && use_local) {
			host = process.env.PL_SERVER_URL!;
		}
		setSenderUrl({
			url: host
		});
		user.access_token = token.access_token!;
		user.refresh_token = token.refresh_token!;
		user.session_id = token.session_id!;

		if (!user) {
			throw new Error(i18n.global.t('errors.add_user_failed'));
		}
		await app.load(user.id);

		const authRes = await _authenticate({
			did: terminusName,
			type: AuthType.SSI,
			purpose: AuthPurpose.Login,
			caller: 'E002'
		});
		if (authRes == null) {
			throw new Err(
				ErrorCode.AUTHENTICATION_FAILED,
				i18n.global.t('errors.authentication_failed')
			);
		}
		await app.login({
			did: authRes.did,
			password: mnemonic,
			authToken: authRes.token
		});

		user.url = host;
		user.binding = true;
		user.name = terminusName;
		userStore.users!.items.update(user);
		await userStore.save();

		if (process.env.PLATFORM === 'BEX' && userStore.current_id) {
			const bexStore = useBexStore();
			await bexStore.controller.changeAccount(userStore.current_id);
		}
		userStore.clearTempData();
		await callback.onSuccess('');
	} catch (e) {
		await userStore.removeUser(did);
		if (old_current_id) {
			await userStore.setCurrentID(old_current_id);
		} else {
			await userStore.removeCurrentID();
		}
		callback.onFailure(e.message);
	}
}

export async function importUserSkipBind(
	terminusName: string,
	mnemonic: string,
	callback: BusinessCallback
) {
	try {
		const userStore = useUserStore();
		const did = await getDID(mnemonic);
		const user = await userStore.importUser(did, terminusName, mnemonic);
		if (!user) {
			throw Error(i18n.global.t('errors.create_user_failed'));
		}

		clearSenderUrl();
		user.access_token = '';
		user.refresh_token = '';
		user.session_id = '';

		await userStore.setCurrentID(user.id);
		await app.load(user.id);

		await app.new(user.id, mnemonic);

		user.url = '';
		user.binding = false;
		user.name = terminusName;
		userStore.users!.items.update(user);

		await userStore.save();

		callback.onSuccess('');
	} catch (e) {
		callback.onFailure(e.message);
	}
}

export async function importUserCheckFa(
	terminusName: string,
	token: Token,
	oneTimePasswordMethod: any,
	mnemonic: string,
	use_local: boolean,
	callback: BusinessAsyncCallback
) {
	try {
		if (!token) {
			throw Error(i18n.global.t('errors.token_is_null'));
		}
		let baseURL = 'https://auth.' + terminusName + '.' + TerminusDefaultDomain;
		let headers: any = {
			'Access-Control-Allow-Origin': '*',
			'Access-Control-Allow-Headers':
				'access-control-allow-headers,access-control-allow-methods,access-control-allow-origin,content-type,x-auth,x-unauth-error,x-authorization',
			'Access-Control-Allow-Methods': 'PUT,POST,GET,DELETE,OPTIONS',
			'X-Unauth-Error': 'Non-Redirect',
			'Content-Type': 'application/json'
		};
		let withCredentials = true;
		if (process.env.IS_PC_TEST && use_local) {
			baseURL = process.env.BFL_URL!;
			headers = {};
			withCredentials = false;
		}

		const instance = axiosInstanceProxy({
			baseURL: baseURL,
			timeout: 1000 * 10,
			withCredentials,
			headers
		});

		const response = await instance.post('/api/secondfactor/totp', {
			targetURL:
				'https://vault.' + terminusName + '.' + TerminusDefaultDomain + '/',
			token: oneTimePasswordMethod
		});
		if (!response || response.status != 200 || !response.data) {
			throw Error(i18n.global.t('errors.network_error_please_try_again_later'));
		}

		if (response.data.status != 'OK') {
			throw new Error(i18n.global.t('errors.error_one_time_password'));
		}

		await importUserByTerminusToken(
			terminusName,
			token,
			mnemonic,
			use_local,
			callback
		);
	} catch (e) {
		callback.onFailure(e.message);
	}
}

export async function loginTerminus(
	user: UserItem,
	password: string,
	needTwoFactor: boolean,
	use_local = true
) {
	let baseURL = user.auth_url;
	if (process.env.IS_PC_TEST && use_local) {
		baseURL = process.env.BFL_URL!;
	}
	const userStore = useUserStore();
	const token = await onFirstFactor(
		baseURL,
		user.name,
		user.local_name,
		password,
		true,
		needTwoFactor,
		userStore.getUserTerminusInfo(user.id).osVersion
	);

	user.access_token = token.access_token;
	user.refresh_token = token.refresh_token;
	user.session_id = token.session_id;
	user.passed_fa2 = false;

	if (token.fa2 || needTwoFactor) {
		const itemList = getVaultsByType(VaultType.TerminusTotp);
		let totpFiledRef;
		if (itemList.length > 0) {
			totpFiledRef = itemList[0].fields.find((value) => {
				return value.type === 'totp';
			});
		}
		if (!totpFiledRef) {
			throw new Error(i18n.global.t('errors.no_local_onetime_password'));
		}
		const time = Date.now();
		const interval = 30;
		const tempCounter = Math.floor(time / 1000 / interval);
		let oneTimePasswordMethod = await hotp(
			base32ToBytes(totpFiledRef!.value),
			tempCounter
		);
		if (!oneTimePasswordMethod) {
			throw new Error(i18n.global.t('errors.no_local_onetime_password'));
		}

		if (process.env.IS_PC_TEST && use_local) {
			oneTimePasswordMethod = '123456';
		}

		let headers: any = {
			'Access-Control-Allow-Origin': '*',
			'Access-Control-Allow-Headers':
				'access-control-allow-headers,access-control-allow-methods,access-control-allow-origin,content-type,x-auth,x-unauth-error,x-authorization',
			'Access-Control-Allow-Methods': 'PUT,POST,GET,DELETE,OPTIONS',
			'X-Unauth-Error': 'Non-Redirect',
			'Content-Type': 'application/json'
		};
		let withCredentials = true;
		if (process.env.IS_PC_TEST && use_local) {
			headers = {};
			withCredentials = false;
		}

		const instance = axiosInstanceProxy({
			baseURL: baseURL,
			timeout: 1000 * 10,
			withCredentials,
			headers
		});

		const targetUrl = 'https://desktop.' + user.name.replace('@', '.') + '/';

		const response = await instance.post('/api/secondfactor/totp', {
			targetUrl,
			token: oneTimePasswordMethod
		});
		if (!response || response.status != 200 || !response.data) {
			throw Error(i18n.global.t('errors.network_error_please_try_again_later'));
		}

		if (response.data.status != 'OK') {
			throw new Error(i18n.global.t('errors.error_one_time_password'));
		}
		const token = response.data.data;

		user.access_token = token.access_token;
		user.refresh_token = token.refresh_token;
		user.session_id = token.session_id;

		user.passed_fa2 = true;
	}

	// const userStore = useUserStore();
	await userStore.users!.items.update(user);
	await userStore.save();
}

export async function loginVault(user: UserItem, mnemonic: string) {
	await app.clearSession();
	const authRes = await _authenticate({
		did: user.local_name,
		type: AuthType.SSI,
		purpose: AuthPurpose.Login,
		caller: 'E003'
	});
	if (authRes == null) {
		throw new Err(
			ErrorCode.AUTHENTICATION_FAILED,
			i18n.global.t('errors.authentication_failed')
		);
	}
	await app.login({
		did: authRes.did,
		password: mnemonic,
		authToken: authRes.token
	});

	user.setup_finished = true;
	user.wizard = '';
	user.terminus_activate_status = 'completed';

	const userStore = useUserStore();
	userStore.users!.items.update(user);
	await userStore.save();
}

export async function connectTerminus(
	user: UserItem,
	mnemonic: string,
	password: string,
	use_local = true
) {
	if (!user) {
		throw new Error(i18n.global.t('errors.user_is_empty'));
	}
	let isWebPlatform = false;
	if (process.env.APPLICATION == 'VAULT') {
		isWebPlatform = true;
	}

	let baseUrl = user.vault_url;
	if (isWebPlatform) {
		if (process.env.NODE_ENV === 'development') {
			baseUrl = '/server';
		} else {
			baseUrl = process.env.PL_SERVER_URL || window.location.origin + '/server';
		}
	}

	if (!isWebPlatform) {
		await loginTerminus(user, password, use_local, false);
	}

	setSenderUrl({
		url: baseUrl
	});
	if (isWebPlatform) {
		await app.load(undefined);
	} else {
		await app.load(user.id);
		await app.unlock(mnemonic, false);
	}
	if (user.setup_finished) {
		const code = await app.simpleSync();
		if (code == ErrorCode.INVALID_SESSION) {
			await loginVault(user, mnemonic);
		} else {
			if (code == ErrorCode.TOKE_INVILID) {
				const terminusInfo = await getTerminusInfo(user);
				if (terminusInfo) {
					if (
						(terminusInfo?.id && user.olares_device_id != terminusInfo.id) ||
						(terminusInfo?.terminusId &&
							user.olares_device_id != terminusInfo.terminusId)
					) {
						await loginVault(user, mnemonic);
					}
				}
			}
		}
	} else {
		await loginVault(user, mnemonic);
	}
}

export async function getTerminusInfo(
	user: UserItem,
	timeout = 5000,
	params = {} as any
): Promise<OlaresInfo | null> {
	try {
		let data: OlaresInfo | null = null;
		if (user.os_version && user.isLargeVersion12_2) {
			data = await axios.get(user.terminus_url + '/api/olares-info', {
				timeout: timeout,
				params: params
			});
		} else {
			if (user.os_version) {
				data = await axios.get(user.terminus_url + '/api/terminus-info', {
					timeout: timeout,
					params: params
				});
			} else {
				try {
					data = await axios.get(user.terminus_url + '/api/olares-info', {
						timeout: timeout,
						params: params
					});
				} catch (error) {
					data = await axios.get(user.terminus_url + '/api/terminus-info', {
						timeout: timeout,
						params: params
					});
				}
			}
		}

		if (data && data.osVersion) {
			const userStore = useUserStore();
			await userStore.setUserTerminusInfo(user.id, data);
		}
		return data;
	} catch (e) {
		return null;
	}
}

export async function getBflTerminusInfo(
	user: UserItem,
	timeout = 5000,
	params = {} as any
): Promise<OlaresInfo | null> {
	try {
		let data: OlaresInfo | null = null;
		if (user.os_version && user.isLargeVersion12_2) {
			data = await axios.get(user.auth_url + 'bfl/info/v1/olares-info', {
				timeout: timeout,
				params: params
			});
		} else {
			if (user.os_version) {
				data = await axios.get(user.auth_url + 'bfl/info/v1/terminus-info', {
					timeout: timeout,
					params: params
				});
			} else {
				try {
					data = await axios.get(user.auth_url + 'bfl/info/v1/olares-info', {
						timeout: timeout,
						params: params
					});
				} catch (error) {
					data = await axios.get(user.auth_url + 'bfl/info/v1/terminus-info', {
						timeout: timeout,
						params: params
					});
				}
			}
		}

		if (data && data.osVersion) {
			const userStore = useUserStore();
			await userStore.setUserTerminusInfo(user.id, data);
		}
		return data;
	} catch (e) {
		return null;
	}
}

export async function importUser(
	terminusName: string | null,
	mnemonic: string
): Promise<UserItem | null> {
	try {
		const userStore = useUserStore();
		if (!(await userStore.importUserPrecheck())) {
			return null;
		}
		let isWebPlatform = false;
		if (process.env.APPLICATION == 'VAULT') {
			isWebPlatform = true;
		}

		if (!mnemonic) {
			throw new Error(i18n.global.t('errors.master_password_can_not_be_empty'));
		}

		const array = mnemonic.split(' ');
		if (array.length != 12) {
			throw new Error(i18n.global.t('errors.mnemonics_is_not_valid'));
		}
		// const userStore = useUserStore();
		const did = await getDID(mnemonic);

		if (terminusName) {
			const ssiStore = useSSIStore();
			const name = await ssiStore.get_name_by_did(did);

			if (!name) {
				throw new Error(i18n.global.t('errors.get_name_by_did_error'));
			}

			if (terminusName.replace('@', '.') !== name) {
				throw new Error(i18n.global.t('errors.mnemonic_is_error'));
			}
		}

		const user = await userStore.importUser(
			did,
			terminusName ? terminusName : '',
			mnemonic
		);
		if (!user) {
			throw Error(i18n.global.t('errors.create_user_failed'));
		}

		clearSenderUrl();
		user.access_token = '';
		user.refresh_token = '';
		user.session_id = '';

		let terminusInfo: OlaresInfo | undefined;

		if (isWebPlatform) {
			if (userStore.terminusInfo) {
				terminusInfo = userStore.terminusInfo();
				userStore.currentUserSaveTerminusInfo(undefined);
			}
		}
		await userStore.setCurrentID(user.id);
		if (isWebPlatform) {
			if (terminusInfo) {
				userStore.currentUserSaveTerminusInfo(terminusInfo);
			}
			await app.load(undefined);
		} else {
			await app.load(user.id, getAppPlatform().reconfigAppStateDefaultValue);
			await app.new(user.id, mnemonic);
		}

		user.binding = false;
		user.setup_finished = false;
		if (terminusName) {
			user.name = terminusName;
		} else {
			user.name = '';
		}
		user.wizard = 'import';
		userStore.users!.items.update(user);
		await userStore.save();

		await userStore.backupCurrentUser();

		return user;
	} catch (e) {
		throw new Error(e.message);
	}
}

export const unlockUserFirstBusiness = async (props: any) => {
	if (!getAppPlatform().getQuasar()) {
		return false;
	}
	return new Promise<boolean>((resolve) => {
		getAppPlatform()
			.getQuasar()
			?.dialog({
				component: UnlockAppDialog,
				componentProps: props
			})
			.onOk(() => {
				resolve(true);
			})
			.onCancel(() => {
				resolve(false);
			});
	});
};
