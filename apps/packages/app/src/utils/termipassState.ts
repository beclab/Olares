import { OlaresInfo, TerminusInfo } from '@bytetrade/core';
import { ErrorCode, UserItem } from '@didvault/sdk/src/core';
// import axios from 'axios';
import { app, getSenderUrl, setSenderUrl } from 'src/globals';
import { useUserStore } from 'src/stores/user';
import {
	NetworkUpdateMode,
	NetworkErrorMode,
	busOn,
	networkErrorModeString,
	busEmit
} from './bus';
import { useDeviceStore } from 'src/stores/device';
import { useScaleStore } from 'src/stores/scale';
import { useTermipassStore } from 'src/stores/termipass';
import { BuildTransition, StateMachine } from './stateMachine';
import { axiosInstanceProxy } from 'src/platform/httpProxy';
import { commonInterceptValue } from './response';
import { getAppPlatform } from 'src/application/platform';

export enum TermiPassStatus {
	INIT = 0,
	OfflineMode = 1,
	Offline,
	VPNConnecting,
	VPNDisconnecting,
	NetworkOffline,
	RequiresVpn,
	Reactivation,
	TokenInvalid,
	VaultTokenInvalid,
	P2P,
	DERP,
	Intranet,
	Internet
}

enum TermipassActionStatus {
	None = 0,
	UserSetupFinished,
	TerminusPinged,
	NeedReactive,
	SrpInvalid,
	SsoTokenInvalid,
	SrpValid,
	RefreshTokenInvalid,
	TokenNoNeedRefresh,
	TokenRefreshed,
	Completed
}

export interface TermiPassStateInterface {
	status: TermiPassStatus;
}

type TermiPassStateCacheInfo = 'termimusInfo' | 'vpnStateInfo';

type TermiPassCheckItem = 'termimusInfo' | 'srpToken' | 'refreshToken';

interface CacheInfo<T> {
	cacheDate?: Date;
	info: T;
}
// termipassStore.srpInvalid = false;
// 				termipassStore.ssoInvalid = false;
interface CheckHistoryResult {
	before?:
		| {
				reactivation: boolean;
				srpInvalid: boolean;
				ssoInvalid: boolean;
		  }
		| {
				access_token: string;
				refresh_token: string;
				session_id: string;
		  };
	after?:
		| {
				reactivation: boolean;
				srpInvalid: boolean;
				ssoInvalid: boolean;
		  }
		| {
				access_token: string;
				refresh_token: string;
				session_id: string;
		  };
	description: string;
}

type CheckLogHistoryType = 'reason' | 'check';

interface CheckLogHistoryInterface {
	type: CheckLogHistoryType;
	date: Date;
	result?: CheckHistoryResult;
	checkItem?: TermiPassCheckItem;
	reasonDesc?: string;
}

const GetVPNHostPeerInfoCountMax = 6;

const CheckTerminusInfoTimeInterval = 60 * 2;

const CheckVPNStatusInfoTimeInterval = 30;

const UserCheckHistoryMaxLength = 100;

export class TermiPassState {
	private getVPNHostPeerInfoTimer: NodeJS.Timer | undefined;

	private tokenRefresh = false;

	private tokenRefreshIng = false;

	private terminusInfoRefresh = false;

	private terminusInfoRefreshIng = false;

	private currentUser: UserItem;

	private termiPassStateUserLastCheckCacheInfo: Record<
		string,
		Record<
			TermiPassStateCacheInfo,
			CacheInfo<TermiPassStateCacheInfo> | undefined
		>
	> = {};

	private termiPassStateCheckHistory: Record<
		string,
		CheckLogHistoryInterface[]
	> = {};

	private srpTokenCheck = false;

	private srpTokenChecking = false;

	private appIsActive = true;

	private getVPNHostPeerInfoCount = 0;

	private lastErrorCheckNetworkTimer = 0;

	private checkEnable = false;

	private terminusCheckingRunLoopTimer: NodeJS.Timer | undefined;

	private termipassActionStatusOptions = {
		init: TermipassActionStatus.None,
		onBefore: (from: TermipassActionStatus, to: TermipassActionStatus) => {
			console.log('before from ===>', from);
			console.log('before to ===>', to);
		},
		onAfter: (from: TermipassActionStatus, to: TermipassActionStatus) => {
			console.log('after from ===>', from);
			console.log('after to ===>', to);
		},
		transitions: {
			step: [],
			reset: BuildTransition('*', TermipassActionStatus.None, () => {
				const termipassStore = useTermipassStore();
				termipassStore.srpInvalid = false;
				termipassStore.ssoInvalid = false;
				termipassStore.reactivation = false;
			}),
			goto: BuildTransition<TermipassActionStatus>(
				'*',
				(state) => state,
				(from, to) => {
					const termipassStore = useTermipassStore();
					if (to == TermipassActionStatus.SrpValid) {
						termipassStore.srpInvalid = false;
						termipassStore.ssoInvalid = false;
						termipassStore.reactivation = false;
						this.publicActions.startTokenRefresh();
					} else if (to == TermipassActionStatus.TokenRefreshed) {
						if (from == TermipassActionStatus.SsoTokenInvalid) {
							this.srpTokenCheck = true;
						} else if (from == TermipassActionStatus.SrpValid) {
							setTimeout(() => {
								this.stateMachine
									.transition()
									.goto(TermipassActionStatus.Completed);
							}, 100);
						}
					} else if (to == TermipassActionStatus.SrpInvalid) {
						termipassStore.srpInvalid = true;
						termipassStore.ssoInvalid = false;
						termipassStore.reactivation = false;
					} else if (to == TermipassActionStatus.SsoTokenInvalid) {
						this.publicActions.startTokenRefresh();
					} else if (to == TermipassActionStatus.NeedReactive) {
						termipassStore.srpInvalid = false;
						termipassStore.ssoInvalid = false;
						termipassStore.reactivation = true;
					} else if (to == TermipassActionStatus.RefreshTokenInvalid) {
						if (from !== TermipassActionStatus.SrpValid) {
							termipassStore.srpInvalid = false;
							termipassStore.ssoInvalid = true;
							termipassStore.reactivation = false;
						}
					} else if (to == TermipassActionStatus.TokenNoNeedRefresh) {
						setTimeout(() => {
							this.stateMachine
								.transition()
								.goto(TermipassActionStatus.Completed);
						}, 100);
					}
				}
			)
		}
	};

	private stateMachine = new StateMachine(this.termipassActionStatusOptions);

	constructor() {
		busOn(
			'network_error',
			async (info: { type: NetworkErrorMode; error: any }) => {
				const now = new Date().getTime();
				if (now - this.lastErrorCheckNetworkTimer > 30 * 1000) {
					if (!this.needChecking()) {
						return;
					}
					this.lastErrorCheckNetworkTimer = now;
					if (this.currentUser) {
						this.addCheckHistory(this.currentUser.id, {
							date: new Date(),
							type: 'reason',
							reasonDesc:
								'network_error:' +
								' type =>' +
								networkErrorModeString(info.type) +
								' error:' +
								info.error
						});
					}

					this.srpTokenCheck = true;
				}
			}
		);

		busOn('account_update', async () => {
			await this.actions.init();
			if (!this.needChecking()) {
				return;
			}
			this.addCheckHistory(this.currentUser.id, {
				date: new Date(),
				type: 'reason',
				reasonDesc: 'account_update'
			});
			await this.actions.ping();
			this.srpTokenCheck = true;
		});

		busOn('network_update', async (mode: NetworkUpdateMode) => {
			if (!this.needChecking()) {
				return;
			}
			this.getVPNHostPeerInfoCount = 0;

			this.actions.getVPNHostPeerInfo();
			const userStore = useUserStore();
			if (
				mode == NetworkUpdateMode.update ||
				(userStore.current_user?.isLocal &&
					mode == NetworkUpdateMode.vpnStop) ||
				(!userStore.current_user?.isLocal && mode == NetworkUpdateMode.vpnStart)
			) {
				await this.actions.init();
				this.addCheckHistory(this.currentUser.id, {
					date: new Date(),
					type: 'reason',
					reasonDesc: 'network_update:'
				});
				await this.actions.ping();
				this.srpTokenCheck = true;
			}
		});

		busOn('appStateChange', async (state: { isActive: boolean }) => {
			this.appIsActive = state.isActive;
			if (!this.needChecking()) {
				return;
			}
			this.addCheckHistory(this.currentUser.id, {
				date: new Date(),
				type: 'reason',
				reasonDesc: 'appStateChange:' + state.isActive
			});
			await this.actions.init();
			await this.actions.ping();
			this.srpTokenCheck = true;
		});

		busOn('terminus_update', async () => {
			setTimeout(async () => {
				await this.actions.init();
				await this.actions.ping();
				this.srpTokenCheck = true;
			}, 5000);
		});

		this.resetCheckIntervalStatus();
	}

	publicActions = {
		startTerminusInfoRefresh: () => {
			this.terminusInfoRefresh = true;
		},
		startTokenRefresh: () => {
			this.tokenRefresh = true;
		},
		resetCheckEnable: (checkEnable: boolean) => {
			this.checkEnable = checkEnable;
		},
		startSrpTokenCheck: () => {
			this.srpTokenCheck = true;
		},
		getCheckHistory: () => {
			return this.termiPassStateCheckHistory[this.currentUser.id];
		},
		setSSOTokenInvalid: () => {
			if (
				this.stateMachine.state() == TermipassActionStatus.RefreshTokenInvalid
			) {
				return;
			}
			this.stateMachine
				.transition()
				.goto(TermipassActionStatus.SsoTokenInvalid);
		}
	};

	private actions = {
		init: async () => {
			const userStore = useUserStore();
			const user = userStore.current_user;

			this.stateMachine.transition().reset();
			if (!user || !user.setup_finished) {
				return;
			}
			this.currentUser = user;
			this.stateMachine
				.transition()
				.goto(TermipassActionStatus.UserSetupFinished);
		},
		ping: async () => {
			if (this.stateMachine.state() < TermipassActionStatus.UserSetupFinished) {
				return;
			}

			if (!getAppPlatform().hookServerHttp) {
				return;
			}

			const userStore = useUserStore();

			if (userStore.current_user?.isLargeVersion12) {
				await this.actions.getTerminusInfo(false, false);
			} else {
				const isLocal =
					(await this.actions.getTerminusInfo(false, true)) != undefined;

				if (!userStore.current_user?.os_version) {
					this.terminusInfoRefresh = true;
				}

				if (isLocal != this.currentUser!.isLocal) {
					this.currentUser!.isLocal = isLocal;
					busEmit('userIsLocalUpdate', isLocal);
				}
				this.actions.resetSenderUrl();
			}

			this.stateMachine.transition().goto(TermipassActionStatus.TerminusPinged);
		},
		resetSenderUrl: async () => {
			if (this.stateMachine.state() < TermipassActionStatus.UserSetupFinished) {
				return;
			}
			if (!getAppPlatform().hookServerHttp) {
				return;
			}
			if (getSenderUrl() != this.currentUser!.vault_url) {
				setSenderUrl({
					url: this.currentUser!.vault_url
				});
			}
		},
		checkSRPValid: async () => {
			if (this.tokenRefreshIng) {
				return;
			}

			if (this.srpTokenChecking) {
				return;
			}

			this.srpTokenCheck = false;

			if (!getAppPlatform().hookServerHttp) {
				return;
			}

			this.srpTokenChecking = true;

			const termipassStore = useTermipassStore();

			const checkResult: CheckHistoryResult = {
				before: {
					reactivation: termipassStore.reactivation,
					ssoInvalid: termipassStore.ssoInvalid,
					srpInvalid: termipassStore.srpInvalid
				},
				description: ''
			};

			if (termipassStore.reactivation) {
				await this.actions.getTerminusInfo(true, false);
				if (termipassStore.reactivation) {
					checkResult.after = {
						reactivation: termipassStore.reactivation,
						ssoInvalid: termipassStore.ssoInvalid,
						srpInvalid: termipassStore.srpInvalid
					};
					checkResult.description = 'No need to check again app.simpleSync';
					this.addCheckHistory(this.currentUser.id, {
						type: 'check',
						date: new Date(),
						result: checkResult,
						checkItem: 'srpToken'
					});
					this.srpTokenChecking = false;
					return;
				}
			}

			const result = await app.simpleSync();
			console.log('result ===>', result);

			if (result) {
				checkResult.description = result;
				if (result == ErrorCode.INVALID_SESSION) {
					this.stateMachine.transition().goto(TermipassActionStatus.SrpInvalid);
				} else {
					if (result == ErrorCode.TOKE_INVILID) {
						// 400
						const terminusInfo = await this.actions.getTerminusInfo(false);
						if (
							terminusInfo &&
							((terminusInfo.id &&
								terminusInfo.id == this.currentUser.olares_device_id) ||
								(terminusInfo.terminusId &&
									terminusInfo.terminusId == this.currentUser.olares_device_id))
						) {
							this.stateMachine
								.transition()
								.goto(TermipassActionStatus.SsoTokenInvalid);
						} else {
							this.stateMachine
								.transition()
								.goto(TermipassActionStatus.SrpInvalid);
						}
					} else {
						//525
						if (result == ErrorCode.SERVER_NOT_EXIST) {
							this.stateMachine
								.transition()
								.goto(TermipassActionStatus.NeedReactive);
							await this.actions.getTerminusInfo(false);
						} else if (result == ErrorCode.SERVER_ERROR) {
							if (this.currentUser.isLocal) {
								await this.actions.ping();
							}
							if (!this.currentUser.isLocal) {
								await this.actions.getTerminusInfo(false);
							}
						}
					}
				}
			} else {
				this.stateMachine.transition().goto(TermipassActionStatus.SrpValid);
			}

			checkResult.after = {
				reactivation: termipassStore.reactivation,
				ssoInvalid: termipassStore.ssoInvalid,
				srpInvalid: termipassStore.srpInvalid
			};

			this.addCheckHistory(this.currentUser.id, {
				type: 'check',
				date: new Date(),
				result: checkResult,
				checkItem: 'srpToken'
			});
			this.srpTokenChecking = false;
		},
		// isPing on used in Olares 1.11
		getTerminusInfo: async (addHistory = false, isPing = false) => {
			if (this.stateMachine.state() < TermipassActionStatus.UserSetupFinished) {
				return;
			}
			const userStore = useUserStore();

			if (this.terminusInfoRefreshIng) {
				return;
			}

			this.terminusInfoRefreshIng = true;

			const termipassStore = useTermipassStore();

			const checkUserId = this.currentUser.id;

			const checkResult: CheckHistoryResult = {
				before: {
					reactivation: termipassStore.reactivation,
					ssoInvalid: termipassStore.ssoInvalid,
					srpInvalid: termipassStore.srpInvalid
				},
				description: ''
			};

			let saveLastTerminusInfo = true;

			try {
				const baseUrl = isPing
					? userStore.pingTerminusInfo
					: this.currentUser.terminus_url;

				const instance = axiosInstanceProxy({
					baseURL: baseUrl,
					headers: {
						'Content-Type': 'application/json'
					},
					timeout: 10000
				});
				const data = await instance.get(
					baseUrl +
						(this.currentUser.isLargeVersion12_2
							? '/api/olares-info'
							: '/api/terminus-info'),
					{}
				);
				if (commonInterceptValue.includes(data.data)) {
					if (isPing) {
						saveLastTerminusInfo = false;
						return;
					}

					this.currentUser.isLocal = false;
					this.actions.resetSenderUrl();

					termipassStore.reactivation = true;
					await userStore.setUserTerminusInfo(this.currentUser.id, undefined);
					checkResult.description = data.data;
				} else {
					const terminusInfo: OlaresInfo = data.data.data;

					termipassStore.reactivation = false;
					termipassStore.vpnErrorCount = 0;

					await userStore.setUserTerminusInfo(
						this.currentUser.id,
						terminusInfo
					);
					this.currentUser.tailscale_activated = terminusInfo.tailScaleEnable;

					if (terminusInfo.tailScaleEnable) {
						this.currentUser.isLocal = true;
						this.actions.resetSenderUrl();
					}
					checkResult.description = JSON.stringify(terminusInfo);
					return terminusInfo;
				}
			} catch (e) {
				const termipassStore = useTermipassStore();
				checkResult.description = e.message;
				if (!isPing) {
					termipassStore.updateVpn();
				}

				if (!isPing && (e.response || process.env.PLATFORM == 'BEX')) {
					if (
						process.env.PLATFORM == 'BEX' ||
						e.response.status == 525 ||
						e.response.status == 522 ||
						e.response.status == 530 ||
						e.response.status > 1000
					) {
						if (this.currentUser.tailscale_activated) {
							const scaleStore = useScaleStore();
							if (!scaleStore.isOn) {
								termipassStore.reactivation = false;
								termipassStore.srpInvalid = false;
								termipassStore.ssoInvalid = false;
								return;
							}
						}
						termipassStore.reactivation = true;
					}
				}
			} finally {
				this.terminusInfoRefreshIng = false;

				if (saveLastTerminusInfo) {
					this.setTermiPassStateUserLastCheckCacheInfo(
						checkUserId,
						'termimusInfo',
						{
							cacheDate: new Date(),
							info: 'termimusInfo'
						}
					);
				}

				if (addHistory) {
					checkResult.after = {
						reactivation: termipassStore.reactivation,
						ssoInvalid: termipassStore.ssoInvalid,
						srpInvalid: termipassStore.srpInvalid
					};
					this.addCheckHistory(this.currentUser.id, {
						type: 'check',
						date: new Date(),
						result: checkResult,
						checkItem: 'termimusInfo'
					});
				}
			}
		},
		getVPNHostPeerInfo: async () => {
			const scaleStore = useScaleStore();
			if (!scaleStore.isOn) {
				scaleStore.hostPeerInfo = undefined;
				this.getVPNHostPeerInfoTimer = undefined;
				return;
			}
			if (this.getVPNHostPeerInfoTimer) {
				return;
			}
			await scaleStore.configHostPeerInfo();
			this.getVPNHostPeerInfoTimer = setTimeout(() => {
				this.getVPNHostPeerInfoTimer = undefined;
				this.getVPNHostPeerInfoCount += 1;
				if (this.getVPNHostPeerInfoCount < GetVPNHostPeerInfoCountMax) {
					this.actions.getVPNHostPeerInfo();
				}
			}, 10 * 1000);
		},
		refreshCurrentToken: async () => {
			const userStore = useUserStore();
			if (this.tokenRefreshIng) {
				return;
			}
			this.tokenRefreshIng = true;
			const result = await userStore.currentUserRefreshToken(
				this.stateMachine.state() == TermipassActionStatus.SsoTokenInvalid
			);
			if (result.status) {
				this.stateMachine
					.transition()
					.goto(TermipassActionStatus.TokenRefreshed);
			} else if (!result.status && result.refreshError) {
				this.stateMachine
					.transition()
					.goto(TermipassActionStatus.RefreshTokenInvalid);
			} else if (!result.status) {
				this.stateMachine
					.transition()
					.goto(TermipassActionStatus.TokenNoNeedRefresh);
			}

			this.tokenRefreshIng = false;

			this.addCheckHistory(this.currentUser.id, {
				type: 'check',
				date: new Date(),
				result: {
					before: result.oldToken,
					after: result.newToken,
					description:
						'status:' + result.status + ' ' + 'message:' + result.message
				},
				checkItem: 'refreshToken'
			});
		},
		checkVPNStatusTask: async () => {
			const scaleStore = useScaleStore();
			if (!scaleStore.isOn) {
				return;
			}
			if (!this.currentUser.id) {
				return;
			}
			const date = new Date();
			const cacheVPNCheckInfo = this.getTermiPassStateUserLastCheckCacheInfo(
				this.currentUser.id,
				'vpnStateInfo'
			);
			if (
				cacheVPNCheckInfo == undefined ||
				!cacheVPNCheckInfo.cacheDate ||
				cacheVPNCheckInfo.cacheDate.getTime() / 1000 +
					CheckVPNStatusInfoTimeInterval <
					date.getTime() / 1000
			) {
				await this.actions.getVPNHostPeerInfo();
				this.setTermiPassStateUserLastCheckCacheInfo(
					this.currentUser.id,
					'vpnStateInfo',
					{
						cacheDate: date,
						info: 'vpnStateInfo'
					}
				);
			}
		},
		runloopTasks: async (ms: number) => {
			this.actions.checkVPNStatusTask();
			busEmit('runTask', ms);
		}
	};

	private needChecking() {
		const userStore = useUserStore();
		const deviceStore = useDeviceStore();
		return (
			userStore.current_user?.offline_mode == false &&
			deviceStore.networkOnLine &&
			userStore.current_user?.setup_finished == true &&
			this.appIsActive
		);
	}

	private getTermiPassStateUserLastCheckCacheInfo(
		userId: string,
		type: TermiPassStateCacheInfo
	) {
		try {
			this.configUserStatus(userId);
			return this.termiPassStateUserLastCheckCacheInfo[userId][type];
		} catch (error) {
			return undefined;
		}
	}

	private configUserStatus(userId: string) {
		if (!this.termiPassStateUserLastCheckCacheInfo[userId]) {
			this.termiPassStateUserLastCheckCacheInfo[userId] = {
				termimusInfo: {
					cacheDate: undefined,
					info: 'termimusInfo'
				},
				vpnStateInfo: {
					cacheDate: undefined,
					info: 'vpnStateInfo'
				}
			};
		}
	}

	private setTermiPassStateUserLastCheckCacheInfo(
		userId: string,
		type: TermiPassStateCacheInfo,
		cache?: CacheInfo<TermiPassStateCacheInfo>
	) {
		if (!this.termiPassStateUserLastCheckCacheInfo[userId]) {
			this.configUserStatus(userId);
		}
		this.termiPassStateUserLastCheckCacheInfo[userId][type] = cache;
	}

	private addCheckHistory(userid: string, history: CheckLogHistoryInterface) {
		if (!this.termiPassStateCheckHistory[userid]) {
			this.termiPassStateCheckHistory[userid] = [];
		}
		if (
			UserCheckHistoryMaxLength > 0 &&
			this.termiPassStateCheckHistory[userid].length >=
				UserCheckHistoryMaxLength
		) {
			this.termiPassStateCheckHistory[userid].splice(
				UserCheckHistoryMaxLength - 1,
				this.termiPassStateCheckHistory[userid].length -
					UserCheckHistoryMaxLength +
					1
			);
		}
		this.termiPassStateCheckHistory[userid] = [
			history,
			...this.termiPassStateCheckHistory[userid]
		];
	}

	private resetCheckIntervalStatus() {
		if (this.terminusCheckingRunLoopTimer) {
			return;
		}
		const ms = 1000;
		this.terminusCheckingRunLoopTimer = setInterval(() => {
			this.actions.runloopTasks(ms);
			if (!this.checkEnable || !this.needChecking()) {
				return;
			}

			if (this.tokenRefresh) {
				this.tokenRefresh = false;
				this.actions.refreshCurrentToken();
			}

			if (this.srpTokenCheck) {
				this.actions.checkSRPValid();
			} else {
				const date = new Date();

				if (!this.currentUser) {
					return;
				}
				const cacheTerminusInfo = this.getTermiPassStateUserLastCheckCacheInfo(
					this.currentUser.id,
					'termimusInfo'
				);

				if (
					this.terminusInfoRefresh ||
					cacheTerminusInfo == undefined ||
					(cacheTerminusInfo !== undefined &&
						(cacheTerminusInfo.cacheDate == undefined ||
							(cacheTerminusInfo.cacheDate &&
								cacheTerminusInfo.cacheDate.getTime() / 1000 +
									CheckTerminusInfoTimeInterval <
									date.getTime() / 1000)))
				) {
					this.terminusInfoRefresh = false;
					this.actions.getTerminusInfo(true);
					return;
				}
			}
		}, 1000);
	}
}
