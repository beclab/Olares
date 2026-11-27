import { OlaresInfo } from '@bytetrade/core';
import type { UserItem } from 'src/account/localUserVault';
// import axios from 'axios';
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
import { useLarepassWebsocketManagerStore } from 'src/stores/larepassWebsocketManager';

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
	SsoTokenInvalid,
	RefreshTokenInvalid,
	TokenNoNeedRefresh,
	TokenRefreshed,
	Completed
}

export interface TermiPassStateInterface {
	status: TermiPassStatus;
}

type TermiPassStateCacheInfo = 'termimusInfo' | 'vpnStateInfo';

type TermiPassCheckItem = 'termimusInfo' | 'refreshToken';

interface CacheInfo<T> {
	cacheDate?: Date;
	info: T;
}
interface CheckHistoryResult {
	before?:
		| {
				reactivation: boolean;
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

// A single forced token refresh rejection (e.g. right after activation while
// VPN/session is still settling) must not immediately surface "need login".
// Require this many consecutive hard refresh failures before escalating to
// RefreshTokenInvalid.
const RefreshTokenInvalidConfirmCount = 3;

// Persistent-unreachable heuristic: when the olares has been uninstalled for a
// long time, requests fail as plain network errors instead of the inactive
// status. After this many consecutive hard failures spanning at
// least this duration, and only if the phone clearly has public internet, we
// escalate to `reactivation` so the "Olares inactive" entry can appear.
const OlaresUnreachableCountMax = 3;

const OlaresUnreachableDurationMs = 10 * 60 * 1000;

const PublicInternetProbeUrls = [
	'https://www.gstatic.com/generate_204',
	'https://cloudflare.com/cdn-cgi/trace'
];

const PublicInternetProbeTimeoutMs = 1500;

export class TermiPassState {
	private getVPNHostPeerInfoTimer: NodeJS.Timer | undefined;

	private tokenRefresh = false;

	private tokenRefreshIng = false;

	private tokenForceRefresh = false;

	// Consecutive hard refresh failures; only escalate to RefreshTokenInvalid
	// once it reaches RefreshTokenInvalidConfirmCount.
	private refreshTokenInvalidStreak = 0;

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

	private olaresUnreachable: Record<
		string,
		{ count: number; firstAt: number }
	> = {};

	// Sticky per-user flag: once we have definitively surfaced reactivation
	// (persistent-unreachable escalation, remote "gone" answer, or a special gone
	// HTTP status), keep it set so the VPN-off suppression branch cannot silently
	// flip reactivation back off. Cleared only when the olares becomes reachable
	// again (or the remote answer arrives).
	private olaresReactivationConfirmed: Record<string, boolean> = {};

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
				termipassStore.ssoInvalid = false;
				termipassStore.reactivation = false;
				this.refreshTokenInvalidStreak = 0;
			}),
			goto: BuildTransition<TermipassActionStatus>(
				'*',
				(state) => state,
				(_from, to) => {
					const termipassStore = useTermipassStore();
					if (to == TermipassActionStatus.SsoTokenInvalid) {
						this.tokenForceRefresh = true;
					} else if (to == TermipassActionStatus.RefreshTokenInvalid) {
						termipassStore.ssoInvalid = true;
						termipassStore.reactivation = false;
					} else if (
						to == TermipassActionStatus.TokenRefreshed ||
						to == TermipassActionStatus.TokenNoNeedRefresh
					) {
						if (to == TermipassActionStatus.TokenRefreshed) {
							termipassStore.ssoInvalid = false;
						}
						setTimeout(() => {
							if (this.stateMachine.state() == to) {
								this.stateMachine
									.transition()
									.goto(TermipassActionStatus.Completed);
							}
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
				if (now - this.lastErrorCheckNetworkTimer > 10 * 1000) {
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
					this.publicActions.startTokenRefresh();
					this.publicActions.startTerminusInfoRefresh();
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
			this.publicActions.startTokenRefresh();
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
				mode == NetworkUpdateMode.vpnStop ||
				(!userStore.current_user?.isLocal && mode == NetworkUpdateMode.vpnStart)
			) {
				await this.actions.init();
				this.addCheckHistory(this.currentUser.id, {
					date: new Date(),
					type: 'reason',
					reasonDesc: 'network_update:'
				});
				await this.actions.ping();
				this.publicActions.startTokenRefresh();
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
			this.publicActions.startTokenRefresh();
		});

		busOn('terminus_update', async () => {
			setTimeout(async () => {
				await this.actions.init();
				await this.actions.ping();
				this.publicActions.startTokenRefresh();
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
		},
		updateIsLocal: (isLocal: boolean) => {
			if (isLocal != this.currentUser!.isLocal) {
				this.currentUser!.isLocal = isLocal;
				const websocketStore = useLarepassWebsocketManagerStore();
				websocketStore.dispose();
				websocketStore.restart();
				if (!isLocal) {
					setTimeout(() => {
						this.tokenForceRefresh = true;
					}, 1000);
				}
				busEmit('userIsLocalUpdate', isLocal);
			}
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

			if (
				userStore.current_user?.isLargeVersion12 &&
				!userStore.current_user?.isLargeVersion12_4
			) {
				await this.actions.getTerminusInfo(false, false);
			} else {
				const isLocal =
					(await this.actions.getTerminusInfo(false, true)) != undefined;

				if (!userStore.current_user?.os_version) {
					this.terminusInfoRefresh = true;
				}
				this.publicActions.updateIsLocal(isLocal);
			}

			this.stateMachine.transition().goto(TermipassActionStatus.TerminusPinged);
		},
		// isPing on used in Olares 1.11
		getTerminusInfo: async (
			addHistory = false,
			isPing = false,
			is_local = true
		) => {
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
					ssoInvalid: termipassStore.ssoInvalid
				},
				description: ''
			};

			let saveLastTerminusInfo = true;

			try {
				const baseUrl = isPing
					? userStore.pingTerminusInfo
					: this.currentUser.get_module_url('', 'https', '', is_local);

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

					this.publicActions.updateIsLocal(false);

					termipassStore.reactivation = true;
					// Definitive remote answer: clear the unreachable streak and
					// keep reactivation sticky against later VPN-off suppression.
					this.resetOlaresUnreachable(this.currentUser.id);
					this.setReactivationConfirmed(this.currentUser.id, true);
					await userStore.setUserTerminusInfo(this.currentUser.id, undefined);
					checkResult.description = data.data;
				} else {
					const terminusInfo: OlaresInfo = data.data.data;

					termipassStore.reactivation = false;
					// Reached the olares successfully: reset the unreachable streak and
					// clear the sticky reactivation confirmation.
					this.resetOlaresUnreachable(this.currentUser.id);
					this.setReactivationConfirmed(this.currentUser.id, false);
					termipassStore.vpnErrorCount = 0;
					if (is_local) {
						await userStore.setUserTerminusInfo(
							this.currentUser.id,
							terminusInfo
						);
					}

					this.currentUser.tailscale_activated = terminusInfo.tailScaleEnable;

					checkResult.description = JSON.stringify(terminusInfo);
					return terminusInfo;
				}
			} catch (e) {
				const termipassStore = useTermipassStore();
				checkResult.description = e.message;
				if (!isPing) {
					termipassStore.updateVpn();
				}

				if (!isPing) {
					const isSpecialGoneStatus =
						process.env.PLATFORM == 'BEX' ||
						(!!e.response &&
							(e.response.status == 525 ||
								e.response.status == 522 ||
								e.response.status == 530 ||
								e.response.status > 1000));

					if (isSpecialGoneStatus) {
						if (this.currentUser.tailscale_activated) {
							const scaleStore = useScaleStore();
							// A VPN-activated machine with VPN off normally just needs the
							// VPN, so don't flag reactivation yet. But once it has been
							// unreachable persistently, stop suppressing and surface it.
							if (
								!scaleStore.isOn &&
								!this.reachedReactivationThreshold(this.currentUser.id) &&
								!this.isReactivationConfirmed(this.currentUser.id)
							) {
								termipassStore.reactivation = false;
								termipassStore.ssoInvalid = false;
								return;
							}
						}
						termipassStore.reactivation = true;
						// Sticky: a special gone status is a definitive "olares gone"
						// signal; keep it surfaced against later VPN-off suppression.
						this.setReactivationConfirmed(this.currentUser.id, true);
						this.resetOlaresUnreachable(this.currentUser.id);
					} else if (!e.response) {
						// Hard unreachable: no HTTP response at all (DNS failure, timeout,
						// connection refused). After a long-uninstalled olares the inactive
						// status never arrives, so only this
						// shape feeds the persistent-unreachable heuristic. Escalate once
						// we've failed persistently and the phone clearly has public
						// internet.
						if (this.currentUser.tailscale_activated) {
							const scaleStore = useScaleStore();
							// VPN-activated machine with VPN off: being unreachable is
							// expected and only means "turn on the VPN", NOT that Olares is
							// gone. Don't feed the persistent-unreachable heuristic — doing
							// so would eventually escalate to `reactivation`, which
							// `totalStatus` ranks above Requires-VPN and would hide the
							// "turn on VPN" hint behind a false "Olares inactive". Reset the
							// streak so VPN-off failures never accumulate toward escalation.
							// (If reactivation was already confirmed, it stays surfaced.)
							if (!scaleStore.isOn) {
								this.resetOlaresUnreachable(this.currentUser.id);
								return;
							}
						}
						this.recordOlaresUnreachable(this.currentUser.id);
						await this.maybeEscalateReactivation(this.currentUser.id);
					} else {
						// Reachable but erroring (generic 4xx/5xx with a response): we
						// got *some* response, so this isn't a hard-unreachable data
						// point — clear the unreachable streak. But a generic gateway
						// error is NOT proof the olares is back (it can be a transient
						// 5xx), so do NOT clear the sticky reactivation confirmation
						// here: doing so while `reactivation` stays true would let a
						// later special-gone status with VPN off re-enter the
						// suppression branch and hide the inactive Olares entry. The
						// sticky flag is cleared only on genuine reachability (the
						// success branch above, which received terminusInfo).
						this.resetOlaresUnreachable(this.currentUser.id);
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
						ssoInvalid: termipassStore.ssoInvalid
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
				this.tokenForceRefresh
			);
			if (result.status) {
				this.refreshTokenInvalidStreak = 0;
				this.stateMachine
					.transition()
					.goto(TermipassActionStatus.TokenRefreshed);
				this.tokenForceRefresh = false;
			} else if (!result.status && result.refreshError) {
				// Require consecutive confirmations before surfacing "need login",
				// so a transient post-activation rejection does not falsely log the
				// user out. Until the threshold, treat it as no-need (will retry).
				this.refreshTokenInvalidStreak += 1;
				if (this.refreshTokenInvalidStreak >= RefreshTokenInvalidConfirmCount) {
					this.stateMachine
						.transition()
						.goto(TermipassActionStatus.RefreshTokenInvalid);
					this.tokenForceRefresh = false;
				} else {
					this.stateMachine
						.transition()
						.goto(TermipassActionStatus.TokenNoNeedRefresh);
				}
			} else if (!result.status) {
				this.refreshTokenInvalidStreak = 0;
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

	private recordOlaresUnreachable(userId: string) {
		const cur = this.olaresUnreachable[userId];
		if (!cur || cur.count === 0) {
			this.olaresUnreachable[userId] = { count: 1, firstAt: Date.now() };
		} else {
			cur.count += 1;
		}
	}

	private resetOlaresUnreachable(userId: string) {
		this.olaresUnreachable[userId] = { count: 0, firstAt: 0 };
	}

	private setReactivationConfirmed(userId: string, confirmed: boolean) {
		this.olaresReactivationConfirmed[userId] = confirmed;
	}

	private isReactivationConfirmed(userId: string) {
		return this.olaresReactivationConfirmed[userId] === true;
	}

	private reachedReactivationThreshold(userId: string) {
		const info = this.olaresUnreachable[userId];
		if (!info || info.count === 0) {
			return false;
		}
		return (
			info.count >= OlaresUnreachableCountMax &&
			Date.now() - info.firstAt >= OlaresUnreachableDurationMs
		);
	}

	// Lightweight public-internet probe used to avoid false positives: if the
	// phone itself has no internet, olares being unreachable means nothing.
	private async hasPublicInternet(): Promise<boolean> {
		for (const probeUrl of PublicInternetProbeUrls) {
			const controller = new AbortController();
			const timer = setTimeout(
				() => controller.abort(),
				PublicInternetProbeTimeoutMs
			);
			try {
				await fetch(probeUrl, {
					method: 'GET',
					mode: 'no-cors',
					cache: 'no-store',
					signal: controller.signal
				});
				clearTimeout(timer);
				return true;
			} catch (e) {
				clearTimeout(timer);
			}
		}
		return false;
	}

	private async maybeEscalateReactivation(userId: string): Promise<boolean> {
		if (!this.reachedReactivationThreshold(userId)) {
			return false;
		}
		const deviceStore = useDeviceStore();
		if (!deviceStore.networkOnLine) {
			return false;
		}
		if (!(await this.hasPublicInternet())) {
			return false;
		}
		const termipassStore = useTermipassStore();
		termipassStore.reactivation = true;
		termipassStore.ssoInvalid = false;
		// Mirror the remote "gone" path: clear the cached Olares info so account
		// UI and flows reading `getUserTerminusInfo` don't keep showing stale data
		// while reactivation is surfaced.
		const userStore = useUserStore();
		await userStore.setUserTerminusInfo(userId, undefined);
		// Sticky: keep reactivation surfaced so a later special-gone status while
		// the VPN is off does not re-enter the suppression branch and hide it.
		this.setReactivationConfirmed(userId, true);
		this.addCheckHistory(userId, {
			type: 'reason',
			date: new Date(),
			reasonDesc: 'olares persistently unreachable -> reactivation' // i18n-ignore
		});
		// Definitive escalation: clear the streak so subsequent failing polls
		// don't keep re-triggering and spamming the check history.
		this.resetOlaresUnreachable(userId);
		return true;
	}

	private resetCheckIntervalStatus() {
		if (this.terminusCheckingRunLoopTimer) {
			return;
		}
		const ms = 1000;
		this.terminusCheckingRunLoopTimer = setInterval(() => {
			this.actions.runloopTasks(ms);

			if (!this.checkEnable || !this.needChecking() || !this.currentUser) {
				return;
			}
			if (
				!this.tokenRefreshIng &&
				(this.tokenRefresh || this.tokenForceRefresh)
			) {
				this.tokenRefresh = false;
				this.actions.refreshCurrentToken();
			}

			const date = new Date();
			const cacheTerminusInfo = this.getTermiPassStateUserLastCheckCacheInfo(
				this.currentUser.id,
				'termimusInfo'
			);
			const termipassStore = useTermipassStore();
			if (
				this.terminusInfoRefresh ||
				cacheTerminusInfo == undefined ||
				(cacheTerminusInfo !== undefined &&
					(cacheTerminusInfo.cacheDate == undefined ||
						(cacheTerminusInfo.cacheDate &&
							cacheTerminusInfo.cacheDate.getTime() / 1000 +
								(termipassStore.isVPNOn
									? CheckVPNStatusInfoTimeInterval
									: CheckTerminusInfoTimeInterval) <
								date.getTime() / 1000)))
			) {
				this.terminusInfoRefresh = false;
				this.actions
					.getTerminusInfo(true, undefined, true)
					.then((olaresInfo) => {
						if (!olaresInfo) {
							this.publicActions.updateIsLocal(false);
						}
						if (this.currentUser!.isLocal) {
							this.actions.getTerminusInfo(true, undefined, false);
						}
					});
			}
		}, 1000);
	}
}
