import { defineStore } from 'pinia';
import {
	PrivateJwk,
	OlaresInfo,
	Token,
	DefaultOlaresInfo,
	TerminusDefaultDomain
} from '@bytetrade/core';
import {
	LocalUserVault,
	UserItem,
	MnemonicItem,
	base64ToString,
	uuid
} from '@didvault/sdk/src/core';
import { getDID, getPrivateJWK } from '../did/did-key';
import { GeneralJwsSigner } from '../jose/jws/general/signer';
import { i18n } from '../boot/i18n';
import { SupportLanguageType } from '../i18n';
import { app, setSenderUrl } from '../globals';
import { axiosInstanceProxy } from '../platform/httpProxy';
import { useScaleStore } from './scale';
import {
	current_user_bind_status,
	BIND_STATUS
} from '../utils/terminusBindUtils';
import { refresh_token, SSOTokenRaw } from '../utils/account';
import { NetworkUpdateMode, busEmit } from 'src/utils/bus';
import { useMonitorStore } from './monitor';
import { unlockUserFirstBusiness } from 'src/utils/BindTerminusBusiness';
import { useIntegrationStore } from './integration';
import { unlockByPwd } from '../utils/UnlockBusiness';
import { useFilesStore, FilesIdType } from './files';
import { DefaultDomainValueType } from '../utils/contact';
import { useLarepassWebsocketManagerStore } from './larepassWebsocketManager';
import { useAppsStore } from 'src/stores/bex/apps';
import { useMDNSStore } from './mdns';
import { useDeviceStore } from './device';
import {
	userModeGetItem,
	userModeRemoveItem,
	userModeSetItem
} from './userStorageAction';
import { signJWS } from 'src/layouts/dialog/sign';
import { TermiPassVpnStatus } from 'src/platform/terminusCommon/terminusCommonInterface';
import { notifyFailed } from 'src/utils/notifyRedefinedUtil';
import { useCloudStore } from './cloud';

export const defaultPassword = 'Terminus_p_d_abcd1234.';

export const userMaxCount = 20;

export const UserTemporaryLockTimer = 15 * 60 * 1000;

export const UserTemporaryLockErrorMaxCount = 6;

export interface UserSate {
	users: LocalUserVault | undefined;
	id: string | undefined;
	current_id: string | undefined;
	temp_url: string | undefined;
	temp_import_data: {
		token: Token | undefined;
		terminusName: string | undefined;
		mnemonic: string | undefined;
		osName: string | undefined;
		localServer: boolean;
	};
	openBiometric: boolean;

	transferOnlyWifi: boolean;

	password: string | undefined;

	userUpdating: boolean;

	locale: SupportLanguageType;

	backupList: string[];

	checkRequestError: boolean;

	checkRequestCount: number;

	launchCounts: number;

	passwordReseted: boolean;

	isNewCreateUser: boolean;

	defaultDomain: DefaultDomainValueType;

	temporaryLockUsers: {
		userId: string;
		timer: number;
		errorCount: number;
	}[];

	usetTerminusInfoCache: Record<string, OlaresInfo | undefined>;
}

// export let UsersTerminusInfo: Record<string, OlaresInfo | undefined> = {};

export const useUserStore = defineStore('user', {
	state: () => {
		return {
			users: undefined,
			previousUsers: undefined,
			id: undefined,
			current_id: undefined,
			temp_url: undefined,
			temp_import_data: {
				token: undefined,
				terminusName: undefined,
				mnemonic: undefined,
				osName: undefined,
				localServer: false
			},
			openBiometric: false,
			transferOnlyWifi: true,
			password: undefined,
			userUpdating: false,
			locale: undefined,
			backupList: [],
			reactivation: false,
			SSOInvalid: false,
			srpInvalid: false,
			isLocal: false,
			checkRequestError: false,
			checkRequestCount: 0,
			launchCounts: 0,
			passwordReseted: false,
			isNewCreateUser: false,
			defaultDomain: 'global',
			temporaryLockUsers: [],
			usetTerminusInfoCache: {}
		} as UserSate;
	},
	getters: {
		isBooted(): boolean {
			return this.id != undefined;
		},
		isUnlocked(): boolean {
			if (!this.users) {
				return false;
			}
			return !this.users.locked && this.password != undefined;
		},
		connected(): boolean {
			if (!this.current_id) {
				return false;
			}
			const user: UserItem = this.users!.items.get(this.current_id!)!;
			return user.setup_finished;
		},
		current_user(): UserItem | null {
			if (!this.current_id) {
				return null;
			}
			if (!this.users) {
				return null;
			}

			return this.users.items.get(this.current_id);
		},
		current_mnemonic(): MnemonicItem | null {
			if (!this.current_id) {
				return null;
			}
			if (!this.users) {
				return null;
			}

			return this.users.mnemonics.get(this.current_id);
		},
		user_name() {
			return this.current_user ? this.current_user.name.split('@')[0] : '';
		},
		async current_user_private_key(): Promise<PrivateJwk | null> {
			if (!this.current_user) {
				return null;
			}
			return await getPrivateJWK(this.current_mnemonic?.mnemonic);
		},
		currentUserBackup(): boolean {
			return this.backupList.find((e) => e == this.current_id) != undefined;
		},
		pingTerminusInfo() {
			if (!this.current_user) {
				return '';
			}
			return (this.current_user as any).local_terminus_url;
		},
		currentUserTokenExpired(): boolean {
			if (!this.current_user) {
				return false;
			}
			const access_token = this.current_user.access_token.split('.')[1];

			const ssoToken: SSOTokenRaw = JSON.parse(base64ToString(access_token));

			const exp = ssoToken.exp;

			return exp < new Date().getTime() / 1000;
		},
		currentUserOlaresInfo(): OlaresInfo {
			if (!this.current_id) {
				return {
					...DefaultOlaresInfo
				};
			}
			return (
				this.usetTerminusInfoCache[this.current_id || ''] || {
					...DefaultOlaresInfo,
					olaresId: this.users?.items.get(this.current_id)?.name || ''
				}
			);
		},
		currentOlaresdAvailable(): boolean {
			return (
				this.currentUserOlaresInfo.olaresd == '1' ||
				this.currentUserOlaresInfo.terminusd == '1'
			);
		}
	},
	actions: {
		async load() {
			this.locale = (await userModeGetItem('locale')) || undefined;
			if (this.locale) {
				i18n.global.locale.value = this.locale;
			}

			const defaultD = await userModeGetItem('defaultDomain');
			if (defaultD) {
				this.defaultDomain = defaultD;
			} else {
				this.setDefaultDomain(
					i18n.global.locale.value == 'zh-CN' ? 'cn' : 'global'
				);
			}

			const passwordReseted = await userModeGetItem('passwordReseted');
			this.passwordReseted =
				passwordReseted == undefined ? true : passwordReseted;

			if (!this.passwordReseted) {
				this.launchCounts = (await userModeGetItem('launchCounts')) || 0;
				this.addLaunchCount();
			}

			const backupListString = await userModeGetItem('backupList');
			this.backupList =
				backupListString != undefined
					? typeof backupListString == 'string'
						? JSON.parse(backupListString)
						: backupListString
					: [];
			const terminusInfos = await userModeGetItem('terminusInfos');
			this.usetTerminusInfoCache =
				terminusInfos != undefined
					? typeof terminusInfos == 'string'
						? JSON.parse(terminusInfos)
						: terminusInfos
					: {};
			this.id = (await userModeGetItem('local-user-id')) || undefined;
			this.current_id = (await userModeGetItem('current-user-id')) || undefined;

			const temporaryLockUsersData = await userModeGetItem(
				'temporaryLockUsers'
			);

			this.temporaryLockUsers =
				temporaryLockUsersData != undefined
					? typeof temporaryLockUsersData == 'string'
						? JSON.parse(temporaryLockUsersData)
						: temporaryLockUsersData
					: [];

			// if ()
			this.temporaryLockUsers = this.temporaryLockUsers.filter(
				(e) => e.timer > new Date().getTime() - UserTemporaryLockTimer
			);

			userModeSetItem(
				'temporaryLockUsers',
				JSON.stringify(this.temporaryLockUsers)
			);

			this.openBiometric = (await userModeGetItem('openBiometric')) || false;

			const transferOnlyWifi = await userModeGetItem('transferOnlyWifi');

			this.transferOnlyWifi =
				transferOnlyWifi == undefined ? true : transferOnlyWifi;
			if (this.id) {
				this.users = new LocalUserVault();
				const res = await userModeGetItem('users');
				if (!res) {
					return;
				}

				this.users.fromRaw(res);

				if (this.current_id) {
					await app.load(this.current_id);
					this.current_user!.isLocal = false;
					if (this.current_user?.setup_finished) {
						let isWebPlatform = false;
						if (process.env.APPLICATION == 'VAULT') {
							isWebPlatform = true;
						}
						let baseUrl = this.current_user.vault_url;
						if (isWebPlatform) {
							if (process.env.NODE_ENV === 'development') {
								baseUrl = '/server';
							} else {
								baseUrl =
									process.env.PL_SERVER_URL ||
									window.location.origin + '/server';
							}
						}
						setSenderUrl({
							url: baseUrl
						});
					}
				}
			}
		},
		getModuleSever(
			module: string,
			protocol = 'https:',
			suffix = '',
			useLocal = true
		) {
			return this.getSelectUserModuleServer(module, protocol, suffix, useLocal);
		},
		getSelectUserModuleServer(
			module: string,
			protocol = 'https:',
			suffix = '',
			useLocal = true,
			userID = ''
		) {
			if (!this.current_user) {
				return '';
			}

			const user =
				userID.length > 0 ? this.users?.items.get(userID) : this.current_user;

			if (!user) {
				return '';
			}

			const server = user.get_module_url(module, protocol, suffix, useLocal);

			return server;
		},

		async create(password: string, openBiometric = false) {
			if (this.id) {
				return;
			}

			this.users = new LocalUserVault();
			this.users.id = await uuid();
			this.users.name = 'LocalUserVault';
			this.users.created = new Date();
			this.users.updated = new Date();
			this.id = this.users.id;
			await this.users.setPassword(password);
			this.password = password;

			this.openBiometric = openBiometric;

			await userModeSetItem('local-user-id', this.id);

			await this.save();

			await userModeSetItem('openBiometric', this.openBiometric);
		},

		async updateOpenBiometricStatus(openBiometric: boolean) {
			this.openBiometric = openBiometric;
			await userModeSetItem('openBiometric', this.openBiometric);
		},

		async save() {
			if (!this.users) {
				console.error('save error ' + JSON.stringify(this.users));
				return;
			}
			if (!this.isUnlocked) {
				return;
			}
			await this.users.commit();
			await userModeSetItem('users', this.users.toRaw());
		},

		async clear() {
			await userModeRemoveItem('users');
			await userModeRemoveItem('local-user-id');
			await userModeRemoveItem('current-user-id');
			await userModeRemoveItem('openBiometric');
			await userModeRemoveItem('backupList');
			await userModeRemoveItem('terminusInfos');
			await userModeRemoveItem('launchCounts');
			await userModeRemoveItem('temporaryLockUsers');

			this.id = undefined;
			this.users = undefined;
		},

		async setCurrentID(id: string) {
			this.current_id = id;
			if (this.current_user) {
				this.current_user.isLocal = false;
			}
			await userModeSetItem('current-user-id', id);
			this.resetCurrentUserData();
		},
		async removeCurrentID() {
			this.current_id = undefined;
			await userModeRemoveItem('current-user-id');
		},
		clearTempData() {
			this.temp_import_data = {
				token: undefined,
				terminusName: undefined,
				mnemonic: undefined,
				osName: undefined,
				localServer: false
			};
		},
		async signJWS(payload: any): Promise<string | null> {
			if (!this.current_user) {
				return null;
			}
			const privateKey = await this.current_user_private_key;
			if (!privateKey) {
				return null;
			}
			const signer = await GeneralJwsSigner.create(
				new TextEncoder().encode(JSON.stringify(payload)),
				[
					{
						privateJwk: privateKey,
						protectedHeader: {
							alg: 'EdDSA',
							kid: this.current_user.id
						}
					}
				]
			);
			const jws = signer.getJws();
			return jws;
		},
		async removeUser(id: string) {
			if (!this.users) {
				return;
			}
			const u = this.users.items.get(id);
			const mnemonic = this.users.mnemonics.get(id);
			if (!u || !mnemonic) {
				return;
			}

			this.users.items.remove(u);
			this.users.mnemonics.remove(mnemonic);

			await this.save();
			await app.removeState(id);
			await this.removeBackupByUserId(id);

			const socketStore = useLarepassWebsocketManagerStore();
			socketStore.dispose();

			if (this.users.items.size > 0) {
				for (const user of this.users.items) {
					// this.current_id = user.id;
					if (user.id) {
						await this.setCurrentID(user.id);
					} else {
						console.error(' remove User current id is null ');
					}
				}
			} else {
				this.current_id = undefined;
				await userModeRemoveItem('current-user-id');
			}
		},
		async removeUsers(ids: string[]) {
			if (!this.users) {
				return;
			}
			// ids.filter(e => this.users!.items.get(e) == null).map(e => this.users!.items.get(e))
			const filterUsers: UserItem[] = ids.reduce((users, id) => {
				const user = this.users?.items.get(id);
				if (user) {
					users.push(user);
				}
				return users;
			}, [] as UserItem[]);

			const filterMnemonics: MnemonicItem[] = ids.reduce((mneminics, id) => {
				const mnemonic = this.users?.mnemonics.get(id);
				if (mnemonic) {
					mneminics.push(mnemonic);
				}
				return mneminics;
			}, [] as MnemonicItem[]);

			if (filterUsers.length == 0 || filterMnemonics.length == 0) {
				return;
			}

			this.users.items.remove(...filterUsers);
			this.users.mnemonics.remove(...filterMnemonics);

			await this.save();

			if (this.current_id && ids.find((e) => e == this.current_id)) {
				await app.removeState(this.current_id);
				await this.removeBackupByUserId(this.current_id);

				if (this.users.items.size > 0) {
					for (const user of this.users.items) {
						// this.current_id = user.id;
						if (user.id) {
							await this.setCurrentID(user.id);
							break;
						} else {
							console.error(' remove User current id is null ');
						}
					}
				} else {
					this.current_id = undefined;
					await userModeRemoveItem('current-user-id');
				}
			}
		},
		// async temporaryCreateUser(did: string, name: string, mnemonic: string) {
		// 	const user1 = new UserItem();
		// 	user1.name = name;
		// 	user1.id = did;
		// 	user1.mnemonic = mnemonic;

		// 	return user1;
		// },
		async importTemporaryUser(user: UserItem) {
			const unlocked = await this.unlockFirst();
			if (!unlocked) {
				return;
			}
			this.users!.items.update(user);
			await this.save();
			return user;
		},
		async importUserPrecheck() {
			if (this.users && this.users.items.size >= userMaxCount) {
				throw Error(
					i18n.global.t('The number of accounts has reached the upper limit')
				);
			}
			return true;
		},
		async importUser(
			did: string,
			name: string,
			mnemonic: string
		): Promise<UserItem | null> {
			const unlocked = await this.unlockFirst();
			if (!unlocked) {
				return null;
			}

			if (this.users!.items.get(did)) {
				return this.users!.items.get(did);
			}

			const user1 = new UserItem();
			user1.name = name;
			user1.id = did;
			//user1.mnemonic = mnemonic;

			const m = new MnemonicItem();
			m.id = did;
			m.mnemonic = mnemonic;

			this.users!.items.update(user1);
			this.users!.mnemonics.update(m);
			await this.save();
			return user1;
		},
		async updateUserPassword(oldPassword: string, newPassword: string) {
			if (!this.users) {
				return {
					status: false,
					message: 'Empty users'
				};
			}
			if (!(await this.unlockFirst())) {
				return {
					status: false,
					message: i18n.global.t('please_unlock_first')
				};
			}

			// Verify the old password without touching the currently active
			// `_key` or `mnemonics`. After `unlockFirst()` the vault is already
			// unlocked and `mnemonics` is populated; we want to keep it that
			// way until we have successfully built the new vault below.
			const oldPasswordValid = await this.users.verifyPassword(oldPassword);
			if (!oldPasswordValid) {
				return {
					status: false,
					message: i18n.global.t('wrong_password_please_try_again')
				};
			}

			const newUsers = new LocalUserVault();
			newUsers.id = this.users.id;
			newUsers.name = this.users.name;
			newUsers.created = this.users.created;
			newUsers.updated = new Date();

			const items = this.users.items;
			const mnemonics = this.users.mnemonics;

			newUsers.items.update(...items);
			newUsers.mnemonics.update(...mnemonics);

			await newUsers.setPassword(newPassword);

			this.password = newPassword;

			this.users = newUsers;
			await this.save();
			this.setPasswordResetedValue(true);

			return {
				status: true,
				message: ''
			};
		},

		async updateLanguageLocale(locale: SupportLanguageType) {
			const deviceStore = useDeviceStore();
			this.locale = locale;
			if (locale) {
				i18n.global.locale.value = locale;
				try {
					app.state.device.locale = locale.split('-')[0].toLowerCase() || 'en';
					if (app.state.device.locale == 'en') {
						// locale.split('-')[0].toLowerCase() || 'en';
					}
					await app.save();
					// await loadLanguage(app.state.device.locale);
				} catch (error) {
					console.log(error);
				}
			}
			deviceStore.setLanguage(this.locale);
		},

		async updateDeviceInfo(data: any): Promise<boolean> {
			if (!this.current_user || !this.current_user.setup_finished) {
				return false;
			}

			try {
				const baseURL = this.getModuleSever('settings');

				if (!baseURL) {
					return false;
				}
				const instance = axiosInstanceProxy({
					baseURL: baseURL,
					timeout: 1000 * 10,
					headers: {
						'Content-Type': 'application/json',
						'X-Authorization': this.current_user.access_token
					}
				});
				await instance.post('/api/device', data);

				return true;
			} catch (e) {
				return false;
			}
		},

		async listUsers() {
			const userStore = useUserStore();
			if (
				!userStore.current_user ||
				!userStore.current_user.auth_url ||
				!userStore.current_user.name
			) {
				return;
			}
			const baseURL = userStore.current_user.auth_url.replace('/server', '/');
			const instance = axiosInstanceProxy({
				baseURL: baseURL,
				timeout: 1000 * 10,
				headers: {
					'Content-Type': 'application/json',
					'X-Authorization': userStore.current_user.access_token
				}
			});

			const response = await instance.get('/api/users');
			if (
				!response ||
				response.status != 200 ||
				!response.data ||
				response.data.code != 0
			) {
				throw Error('Network error, please try again later');
			}
		},
		resetCurrentUserData() {
			const scale = useScaleStore();
			scale.reset();
			const monitor = useMonitorStore();
			monitor.clear();
			const mdnsStore = useMDNSStore();
			mdnsStore.apiMachine = undefined;

			const integrationStore = useIntegrationStore();
			integrationStore.accounts = [];

			const fileStore = useFilesStore();
			fileStore.backStack[FilesIdType.PAGEID] = [];
			fileStore.previousStack[FilesIdType.PAGEID] = [];
			fileStore.mobileRepo = undefined;
			fileStore.nodes = [];
			fileStore.currentNode = {};

			const socketStore = useLarepassWebsocketManagerStore();
			socketStore.dispose();

			const appsStore = useAppsStore();
			appsStore.resetAppList();

			const spaceStore = useCloudStore();
			spaceStore.reset();
		},

		async backupCurrentUser() {
			if (
				!this.current_id ||
				this.backupList.find((e) => e == this.current_id)
			) {
				return;
			}

			this.backupList.push(this.current_id);
			await userModeSetItem('backupList', JSON.stringify(this.backupList));
		},

		async removeBackupByUserId(id: string) {
			const index = this.backupList.findIndex((e) => e == id);
			if (index < 0) {
				return;
			}
			this.backupList.splice(index, 1);
			await userModeSetItem('backupList', JSON.stringify(this.backupList));
		},
		async currentUserRefreshToken(forceRefresh = false) {
			console.log('forceRefresh --->', forceRefresh);

			if (!this.current_user) {
				return {
					status: false,
					refreshError: false,
					oldToken: {
						access_token: '',
						refresh_token: '',
						session_id: ''
					},
					newToken: undefined,
					message: 'no has user'
				};
			}
			if (current_user_bind_status() != BIND_STATUS.BIND_OK) {
				return {
					status: false,
					refreshError: false,
					oldToken: {
						access_token: '',
						refresh_token: '',
						session_id: ''
					},
					newToken: undefined,
					message: 'user not setup_finished'
				};
			}
			const user = this.current_user;

			if (user.access_token.length == 0) {
				return {
					status: false,
					refreshError: false,
					oldToken: {
						access_token: '',
						refresh_token: '',
						session_id: ''
					},
					newToken: undefined,
					message: 'user not has access_token'
				};
			}
			try {
				const access_token = user.access_token.split('.')[1];

				const ssoToken: SSOTokenRaw = JSON.parse(base64ToString(access_token));

				const exp = ssoToken.exp;

				const refreshTime = new Date().getTime() / 1000 + 3600 * 23;

				if (!forceRefresh && exp > refreshTime) {
					return {
						status: false,
						refreshError: false,
						oldToken: {
							access_token: '',
							refresh_token: '',
							session_id: ''
						},
						newToken: undefined,
						message: 'access_token not expired'
					};
				}

				let refreshTokenBaseUrl = this.getModuleSever('auth');
				if (user.tailscale_activated && user.isLargeVersion12_3) {
					refreshTokenBaseUrl = this.getModuleSever('headscale');
				}

				const token: Token = await refresh_token(
					refreshTokenBaseUrl,
					user.refresh_token,
					user.access_token
				);

				const oldrefreshToken = user.refresh_token;
				const oldresessionId = user.session_id;

				user.access_token = token.access_token!;
				if (token && token.refresh_token && token.refresh_token.length > 0) {
					user.refresh_token = token.refresh_token!;
				}
				if (this.isUnlocked) {
					user.session_id = token.session_id!;
					this.users!.items.update(user);
					await this.save();
				} else {
					await this.updateLockUserToken(
						token.access_token,
						token.session_id,
						user.id
					);
				}

				busEmit('account_token_refreshed');

				const scale = useScaleStore();
				if (scale.vpnStatus == TermiPassVpnStatus.on) {
					scale.resendCache();
				}

				return {
					status: true,
					refreshError: false,
					oldToken: {
						access_token: access_token,
						refresh_token: oldrefreshToken,
						session_id: oldresessionId
					},
					newToken: {
						access_token: user.access_token,
						refresh_token: user.refresh_token,
						session_id: user.session_id
					},
					message: 'access_token refresh success'
				};
			} catch (error) {
				return {
					status: false,
					refreshError: true,
					oldToken: {
						access_token: user.access_token,
						refresh_token: user.refresh_token,
						session_id: user.session_id
					},
					newToken: undefined,
					message: 'access_token refresh failed:' + error.message
				};
			}
		},

		updateOfflineMode(offlineMode: boolean) {
			if (this.current_user) {
				this.current_user.offline_mode = offlineMode;
			}
			if (!offlineMode) {
				busEmit('network_update', NetworkUpdateMode.update);
			}

			if (offlineMode) {
				const scale = useScaleStore();
				scale.stop();
			}
		},

		currentUserSaveTerminusInfo(terminusInfo: OlaresInfo | undefined) {
			this.usetTerminusInfoCache[this.current_id || ''] = terminusInfo;
		},

		async setUserTerminusInfo(
			userId: string,
			terminusInfo: OlaresInfo | undefined
		) {
			if (terminusInfo) {
				this.usetTerminusInfoCache[userId] = terminusInfo;
			} else {
				delete this.usetTerminusInfoCache[userId];
			}
			const user = this.current_user;
			if (
				terminusInfo &&
				user &&
				user.id == userId &&
				user.os_version != terminusInfo.osVersion
			) {
				user.os_version = terminusInfo.osVersion;
				this.users!.items.update(user);
				await this.save();
			}

			userModeSetItem(
				'terminusInfos',
				JSON.stringify(this.usetTerminusInfoCache)
			);
		},

		getUserTerminusInfo(userId: string) {
			return (
				this.usetTerminusInfoCache[userId] || {
					...DefaultOlaresInfo,
					olaresId: this.users?.items.get(userId)?.name || ''
				}
			);
		},

		async removeTerminusInfoByUserId(id: string) {
			this.usetTerminusInfoCache[id] = undefined;
			await userModeSetItem(
				'backupList',
				JSON.stringify(this.usetTerminusInfoCache)
			);
		},

		terminusInfo() {
			const olaresInfo = this.usetTerminusInfoCache[this.current_id || ''] || {
				...DefaultOlaresInfo,
				olaresId: this.current_user?.name || ''
			};
			return olaresInfo;
		},

		getCurrentDomain() {
			const current_user = this.current_user;
			if (current_user && current_user.name.indexOf('@')) {
				return current_user.name.split('@')[1];
			} else {
				return TerminusDefaultDomain;
			}
		},
		async unlockFirst(next?: () => void, props?: any) {
			if (this.users && !this.users.locked) {
				if (next) {
					next();
				}
				return true;
			}

			return new Promise<boolean>(async (resolve) => {
				if (this.passwordReseted) {
					const lockStatus = this.currentUserIsTemporaryLocked();
					if (lockStatus.locked) {
						notifyFailed(
							i18n.global.t(
								'Password entered incorrectly multiple times. Account locked for {minutes} minutes. Please try again later.',
								{
									minutes: lockStatus.leftTimes
								}
							)
						);
						resolve(false);
						return;
					}

					const unclocked = await unlockUserFirstBusiness(props);
					if (unclocked) {
						this.removeCurrentUserTemporaryLocked();
						if (next) {
							next();
						}
					}
					resolve(unclocked);
					return;
				}
				unlockByPwd(defaultPassword, {
					async onSuccess() {
						const userStore = useUserStore();
						const hideNotify = props && props.hide;

						if (!hideNotify) {
							const notify = userStore.sendSetPassworNotify();
							if (notify) {
								resolve(false);
								return;
							}
						}
						if (next) {
							next();
						}
						resolve(true);
					},
					async onFailure() {
						const unclocked = await unlockUserFirstBusiness(props);
						if (unclocked) {
							if (next) {
								next();
							}
						}
						resolve(unclocked);
					}
				});
			});
		},
		async unlock(password: string) {
			await this.users?.unlock(password);
			this.password = password;
			await this.save();
		},
		addLaunchCount() {
			this.launchCounts = this.launchCounts + 1;
			userModeSetItem('launchCounts', this.launchCounts);
		},
		sendSetPassworNotify() {
			if (!this.connected) {
				return false;
			}
			if (this.passwordReseted) {
				return false;
			}
			if (
				this.password &&
				this.password == defaultPassword &&
				this.launchCounts % 5 == 2
			) {
				busEmit('configPassword');
				return true;
			}
			return false;
		},
		setPasswordResetedValue(value: boolean) {
			this.passwordReseted = value;
			userModeSetItem('passwordReseted', value);
		},
		setDefaultDomain(domainType: DefaultDomainValueType) {
			this.defaultDomain = domainType;
			userModeSetItem('defaultDomain', this.defaultDomain);
		},
		userIsBackup(id: string): boolean {
			return this.backupList.find((e) => e == id) != undefined;
		},
		async updateTransferOnlyWifiStatus(transferOnlyWifi: boolean) {
			this.transferOnlyWifi = transferOnlyWifi;
			await userModeSetItem('transferOnlyWifi', this.transferOnlyWifi);

			busEmit('appTransferTypeChanged');
		},

		async activeOlaresJws(body: any) {
			if (!body) {
				return '';
			}

			const userStore = useUserStore();

			const mneminicItem = userStore.current_mnemonic;
			const did = await getDID(mneminicItem?.mnemonic);
			const privateJWK: PrivateJwk | undefined = await getPrivateJWK(
				mneminicItem?.mnemonic
			);

			return await signJWS(
				did,
				{
					did: did,
					name: userStore.current_user?.name,
					time: `${new Date().getTime()}`,
					...body,
					challenge: 'challenge',
					body: body
				},
				privateJWK
			);
		},

		async updateLockUserToken(
			access_token: string,
			session_id: string,
			userID: string
		) {
			const users:
				| {
						items: {
							items: [{ id: string; access_token: string; session_id: string }];
						};
				  }
				| undefined = await userModeGetItem('users');
			if (!users) {
				return;
			}

			const item = users.items.items.find((e) => e.id == userID);

			if (!item) {
				return;
			}

			item.access_token = access_token;
			item.session_id = session_id;
			await userModeSetItem('users', users);
		},
		async tryLockCurrentUser() {
			if (!this.current_id) {
				return;
			}
			const lockUserIndex = this.temporaryLockUsers.findIndex(
				(e) => e.userId == this.current_id
			);

			let lockUser = {
				errorCount: 1,
				userId: this.current_id,
				timer: new Date().getTime()
			};

			if (lockUserIndex < 0) {
				this.temporaryLockUsers.push(lockUser);
			} else {
				lockUser = this.temporaryLockUsers[lockUserIndex];
				if (lockUser.errorCount < UserTemporaryLockErrorMaxCount) {
					lockUser.errorCount = lockUser.errorCount + 1;
					lockUser.timer = new Date().getTime();
				}
			}
			console.log('this.temporaryLockUsers', this.temporaryLockUsers);

			await userModeSetItem(
				'temporaryLockUsers',
				JSON.stringify(this.temporaryLockUsers)
			);
		},
		removeCurrentUserTemporaryLocked() {
			if (!this.current_id) {
				return;
			}
			const lockUserIndex = this.temporaryLockUsers.findIndex(
				(e) => e.userId == this.current_id
			);
			if (lockUserIndex < 0) {
				return;
			}
			this.temporaryLockUsers.splice(lockUserIndex, 1);
			userModeSetItem(
				'temporaryLockUsers',
				JSON.stringify(this.temporaryLockUsers)
			);
		},
		currentUserIsTemporaryLocked() {
			if (!this.current_id) {
				return {
					locked: false,
					leftTimes: 0
				};
			}
			const tLockedUser = this.temporaryLockUsers.find(
				(e) => e.userId == this.current_id
			);
			if (!tLockedUser) {
				return {
					locked: false,
					leftTimes: 0
				};
			}
			if (tLockedUser.timer <= new Date().getTime() - UserTemporaryLockTimer) {
				this.removeCurrentUserTemporaryLocked();
			}

			return {
				locked:
					tLockedUser.errorCount >= UserTemporaryLockErrorMaxCount &&
					tLockedUser.timer > new Date().getTime() - UserTemporaryLockTimer,
				leftTimes: Math.ceil(
					(tLockedUser.timer + UserTemporaryLockTimer - new Date().getTime()) /
						60000
				)
			};
		}
	}
});
