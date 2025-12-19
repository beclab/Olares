import { defineStore } from 'pinia';
import {
	TerminusServiceInfo,
	TerminusStatus,
	canInstall,
	canActive,
	isInstalling,
	canUnInstall,
	disableOperate,
	isChecking,
	TerminusStatusEnum,
	getSourceId,
	getMdnsRequestApi,
	MdnsApiEmum,
	UpgradeLevel
	// InstallingState,
	// TerminusStatusEnum
} from '../services/abstractions/mdns/service';
import axios from 'axios';
import {
	notifyRequestMessageError,
	notifySuccess,
	notifyFailed
} from 'src/utils/notifyRedefinedUtil';
import { useUserStore } from './user';
import { generatePasword } from 'src/utils/utils';
import TerminusTipDialog from '../components/dialog/TerminusTipDialog.vue';
import OlaresUpgradeReminderDialog from '../components/dialog/OlaresUpgradeReminderDialog.vue';
import UserStatusCommonDialog from '../components/userStatusDialog/UserStatusCommonDialog.vue';
import { i18n } from 'src/boot/i18n';
import { busEmit, busOff, busOn } from 'src/utils/bus';
import { passwordAddSort } from 'src/utils/account';
import { getDID, getPrivateJWK } from 'src/did/did-key';
import { compareOlaresVersion, GolbalHost, PrivateJwk } from '@bytetrade/core';
import { signJWS } from 'src/layouts/dialog/sign';
import { axiosInstanceProxy } from 'src/platform/httpProxy';
import { UpgradeMode } from 'src/constant/larepass';
import { Platform } from 'quasar';
import { userModeSetItem, userModeGetItem } from './userStorageAction';
import { OlaresTunneV2Interface } from 'src/utils/interface/frp';
import { getNativeAppPlatform } from 'src/application/platform';
import { DNSService } from 'src/platform/interface/capacitor/plugins/dns';

export type MDNSStoreState = {
	mdnsMachines: TerminusServiceInfo[];
	mdnsMachine: TerminusServiceInfo | undefined;
	apiMachine: TerminusServiceInfo | undefined;
	mdnsUsed: boolean;
	// olaresInfo: TerminusStatus | undefined;
	refreshTimer: any;
	rebootMachines: Record<string, { rebootTime: number; status: boolean }>;
	shutdownMachines: Record<string, { shutdownTime: number; status: boolean }>;
	ipChangingsMachines: Record<
		string,
		{ ipChangingTime: number; status: boolean }
	>;

	serverRequestMachines: Record<
		string,
		{ errorStartTime: number; isError: boolean }
	>;

	bluetoothMachines: TerminusServiceInfo[];
	blueConfigNetworkMachine: TerminusServiceInfo | undefined;
	includeRC: boolean;
	frpMap: Record<string, OlaresTunneV2Interface[]>;
};

export interface SystemIPS {
	iface: string;
	ip: string;
	isHostIp: boolean;
	isWifi?: boolean;
	ssid?: string;
	strength?: number;
}

const terminusMachineInfo: Record<string, TerminusStatus | undefined> = {};

const shutdownOrRebootLimitTime = 30000;

const commandDefaultTimeout = 60000;

const serverApiErrorLimitTime = 600000;

const apiMachineMineVersion = '1.12.1-0';

const newActivedMineVersion = '1.12.2-0';

export const useMDNSStore = defineStore('mdnsStore', {
	state: () => {
		return {
			mdnsMachines: [],
			mdnsMachine: undefined,
			apiMachine: undefined,
			refreshTimer: undefined,
			rebootMachines: {},
			shutdownMachines: {},
			ipChangingsMachines: {},
			bluetoothMachines: [],
			blueConfigNetworkMachine: undefined,
			mdnsUsed: true,
			includeRC: false,
			frpMap: {},
			serverRequestMachines: {}
		} as MDNSStoreState;
	},

	getters: {
		activedMachine(): TerminusServiceInfo | undefined {
			if (this.mdnsUsed || !this.apiMachine || !this.apiMachine.status) {
				return this.mdnsMachine;
			}
			if (
				compareOlaresVersion(
					this.apiMachine.status.terminusVersion,
					apiMachineMineVersion
				).compare < 0
			) {
				return this.mdnsMachine;
			}
			return this.apiMachine;
		},

		statusRequestFinishedMachines(state) {
			return state.mdnsMachines.filter((e) => e.status != undefined);
		},
		discoverPageDisplayMachines(state) {
			const userStore = useUserStore();
			const item = state.mdnsMachines.find(
				(e) => e.status && e.status.terminusName == userStore.current_user?.name
			);
			if (item) {
				return [item];
			}

			const isCurrent = (item: TerminusServiceInfo, notFound = false) => {
				if (!state.mdnsMachine) {
					return notFound;
				}
				return state.mdnsMachine && state.mdnsMachine.host == item.host;
			};

			return state.mdnsMachines.filter((e) => {
				return (
					e.status != undefined &&
					((canInstall(e.status) && isCurrent(e, true)) ||
						(canActive(e.status) && isCurrent(e)) ||
						(isInstalling(e.status) && isCurrent(e)) ||
						canUnInstall(e.status))
				);
			});
		},

		bluetoothPageDisplayMachines(state) {
			const userStore = useUserStore();
			const item = state.bluetoothMachines.find(
				(e) => e.status && e.status.terminusName == userStore.current_user?.name
			);
			if (item) {
				return [item];
			}

			const isCurrent = (item: TerminusServiceInfo, notFound = false) => {
				if (!state.mdnsMachine) {
					return notFound;
				}
				return state.mdnsMachine && state.mdnsMachine.host == item.host;
			};

			const items = state.bluetoothMachines.filter((e) => {
				if (!e.status) {
					return false;
				}
				return (
					e.status != undefined &&
					((canInstall(e.status) && isCurrent(e, true)) ||
						(canActive(e.status) && isCurrent(e)) ||
						(isInstalling(e.status) && isCurrent(e)) ||
						canUnInstall(e.status) ||
						disableOperate(e.status) ||
						isChecking(e.status))
				);
			});
			return items;
		},

		isShutdown(state) {
			const userStore = useUserStore();
			if (!userStore.current_id) {
				return false;
			}
			return state.shutdownMachines[userStore.current_id]
				? state.shutdownMachines[userStore.current_id].status
				: false;
		},
		isReboot(state) {
			const userStore = useUserStore();
			if (!userStore.current_id) {
				return false;
			}
			return state.rebootMachines[userStore.current_id]
				? state.rebootMachines[userStore.current_id].status
				: false;
		},

		isInstalled() {
			const userStore = useUserStore();
			if (!userStore.current_user) {
				return false;
			}
			return userStore.current_user.localMachine.length > 0;
		},

		isUpgrading(): boolean {
			const activeM = this.activedMachine;

			if (!activeM || !activeM.status) {
				return false;
			}
			return (
				activeM.status.terminusState == TerminusStatusEnum.Upgrading ||
				(activeM.status.upgradingTarget != undefined &&
					activeM.status.upgradingTarget.length > 0) ||
				(activeM.status.upgradingDownloadState != undefined &&
					activeM.status.upgradingDownloadState.length > 0) ||
				(activeM.status.upgradingDownloadProgress != undefined &&
					activeM.status.upgradingDownloadProgress.length > 0)
			);
		},

		isNewVersionUpgrading(): boolean {
			const activeM = this.activedMachine;

			if (!activeM || !activeM.status) {
				return false;
			}
			return (
				activeM.status.terminusState == TerminusStatusEnum.Upgrading ||
				activeM.status.upgradingState == 'in-progress'
			);
		},

		isNewVersionDowaloading(): boolean {
			const activeM = this.activedMachine;

			if (!activeM || !activeM.status) {
				return false;
			}
			if (activeM.status.terminusState == TerminusStatusEnum.Upgrading) {
				return false;
			}
			return (
				(activeM.status.upgradingDownloadState != undefined &&
					activeM.status.upgradingDownloadState.length > 0) ||
				(activeM.status.upgradingDownloadProgress != undefined &&
					activeM.status.upgradingDownloadProgress.length > 0)
			);
		},

		upgradingDownloadCompleted(): boolean {
			const activeM = this.activedMachine;
			if (!activeM || !activeM.status) {
				false;
			}

			if (activeM!.status!.terminusState == TerminusStatusEnum.Upgrading) {
				return false;
			}
			if (!activeM!.status!.upgradingState) {
				return false;
			}

			return (
				activeM!.status!.upgradingState.length > 0 &&
				activeM!.status!.upgradingState == 'WaitingForUserConfirm'
			);
		},
		upgradingInfo(): string {
			const activeM = this.activedMachine;
			const userStore = useUserStore();
			if (
				!activeM ||
				!activeM.status ||
				!userStore.current_user?.isLargeVersion12
			) {
				return '';
			}
			if (activeM.status.terminusState == TerminusStatusEnum.Upgrading) {
				return `${activeM.status.upgradingProgress}`;
			}
			return `${activeM.status.upgradingDownloadStep} ===> ${activeM.status.upgradingDownloadState} ===> ${activeM.status.upgradingDownloadProgress}`;
		},
		upgradingError(): boolean {
			const activeM = this.activedMachine;
			if (!activeM || !activeM.status) {
				return false;
			}
			return (
				(activeM.status.upgradingDownloadError != undefined &&
					activeM.status.upgradingDownloadError.length > 0) ||
				(activeM.status.upgradingState != undefined &&
					activeM.status.upgradingState == 'failed') ||
				(activeM.status.upgradingError != undefined &&
					activeM.status.upgradingError.length > 0) ||
				this.isServerApiError
			);
		},
		isServerApiError(): boolean {
			const userStore = useUserStore();
			if (!userStore.current_id) {
				return false;
			}
			const data = this.serverRequestMachines[userStore.current_id];
			if (!data) {
				return false;
			}
			return data.isError;
		},
		upgradeEnable(): boolean {
			const activeM = this.activedMachine;
			const userStore = useUserStore();
			if (!userStore.current_user?.isLargeVersion12) {
				return false;
			}

			if (!activeM || !activeM.status) {
				return false;
			}

			if (activeM.status.terminusState == TerminusStatusEnum.Upgrading) {
				return false;
			}

			if (
				// this.activedMachine.status.upgradingTarget.length > 0 ||
				(activeM.status.upgradingDownloadState != undefined &&
					activeM.status.upgradingDownloadState.length > 0) ||
				(activeM.status.upgradingDownloadProgress != undefined &&
					activeM.status.upgradingDownloadProgress.length > 0) ||
				(activeM.status.upgradingDownloadError != undefined &&
					activeM.status.upgradingDownloadError.length > 0) ||
				(activeM.status.upgradingState &&
					activeM.status.upgradingState.length > 0)
			) {
				return false;
			}

			if (!activeM.version) {
				return false;
			}

			return (
				compareOlaresVersion(
					activeM.status.terminusVersion,
					activeM.version.upgradeableVersion
				).compare < 0
			);
		}
	},

	actions: {
		async init() {
			const includeRC = await userModeGetItem('upgradeIncludeRC');
			this.includeRC = includeRC != undefined && includeRC == true;
		},

		async updateIncludeRc(includeRc: boolean) {
			this.includeRC = includeRc;
			await userModeSetItem('upgradeIncludeRC', this.includeRC);
		},

		async initTerminusMachinesInfo(mdnsMachines: TerminusServiceInfo[]) {
			const promises: any = [];

			for (let index = 0; index < mdnsMachines.length; index++) {
				const machine = mdnsMachines[index];
				const info = terminusMachineInfo[machine.host];
				const time = new Date().getTime();

				if (info && info.updateTime + 2000 > time) {
					if (!machine.status) {
						machine.status = info;
					}
					continue;
				}
				const promise = new Promise((resolve) => {
					this.initTerminusInfo(machine).then(() => {
						resolve(machine);
					});
				});
				promises.push(promise);
			}
			await Promise.all(promises);
		},
		async initTerminusInfo(machine: TerminusServiceInfo) {
			try {
				const instance = await this.createAxiosInstanceProxy(
					machine,
					undefined,
					machine.isSettingsServer ? 15000 : 2000
				);
				const result = await instance.get(
					getMdnsRequestApi(MdnsApiEmum.STATUS, machine)
				);
				if (result && result.status == 200 && result.data) {
					terminusMachineInfo[machine.host] = {
						...result.data.data,
						updateTime: new Date().getTime()
					};
					machine.status = result.data.data;
					if (machine.status?.terminusName == '@') {
						const userstore = useUserStore();
						machine.status.terminusName = userstore.current_user?.name || '';
					}
					machine.errorCount = 0;
				} else {
					terminusMachineInfo[machine.host] = undefined;
				}
			} catch (error) {
				if (machine.errorCount) {
					machine.errorCount += 1;
				} else {
					machine.errorCount = 1;
				}
				if (!machine.isSettingsServer && machine.errorCount > 3) {
					terminusMachineInfo[machine.host] = undefined;
					machine.status = undefined;
				}
			}
		},
		async installMachineTerminus(machine: TerminusServiceInfo) {
			try {
				if (!this.precheckAccountEnable(machine)) {
					return;
				}

				const password = generatePasword();
				machine.password = password;
				const userstore = useUserStore();

				const data = {
					username: userstore.current_user?.local_name,
					password: passwordAddSort(password, machine.status?.terminusVersion),
					email: '',
					domain: userstore.current_user?.domain_name
				};

				const instance = await this.createAxiosInstanceProxy(
					machine,
					data,
					commandDefaultTimeout
				);

				const result = await instance.post(
					getMdnsRequestApi(MdnsApiEmum.INSTALL, machine),
					data
				);

				if (result && result.status == 200 && result.data.code == 200) {
					const body = {
						host: machine.host,
						password: password,
						port: machine.port,
						serviceName: machine.serviceName
					};
					const localMechainInfo = JSON.stringify(body);
					const new_user = userstore.users!.items.get(userstore.current_id!)!;
					new_user.localMachine = localMechainInfo;
					await userstore.users!.items.update(new_user);
					await userstore.save();
					notifySuccess(i18n.global.t('Start installing'));
					return true;
				}
				return false;
			} catch (error) {
				notifyRequestMessageError(error);
				return false;
			}
		},

		precheckAccountEnable(machine: TerminusServiceInfo) {
			if (!machine.status) {
				return true;
			}
			if (!machine.status.frpEnable || machine.status.frpEnable != '1') {
				return true;
			}

			const userstore = useUserStore();
			const name = userstore.current_user?.name;
			if (!name) {
				return false;
			}
			if (
				machine.status.defaultFrpServer.endsWith('jointerminus.cn') &&
				!name.endsWith('olares.cn')
			) {
				getNativeAppPlatform()
					.getQuasar()
					?.dialog({
						component: UserStatusCommonDialog,
						componentProps: {
							title: i18n.global.t('Domain resolution error'),
							message: i18n.global.t(
								'Unable to resolve the Olares domain ({domain}) due to an FRP configuration conflict. To fix this issue, create a new Olares ID with the {domain2} domain and reactivate your device:<br>1. Tap the profile icon, select Add a new account, and create a new Olares ID.<br>2. When prompted to confirm the domain, tap Change default domain and select {domain2}.<br>3. Complete the Olares ID creation.<br>4. Reactivate your device using the new Olares ID.',
								{
									domain: userstore.current_user?.domain_name,
									domain2: 'olares.cn'
								}
							)
						}
					});
				return false;
			}
			return true;
		},

		async uninstallMachineTerminus(machine: TerminusServiceInfo) {
			try {
				const userstore = useUserStore();
				const instance = await this.createAxiosInstanceProxy(
					machine,
					{
						username: userstore.current_user?.local_name
					},
					commandDefaultTimeout
				);
				const result = await instance.post(
					getMdnsRequestApi(MdnsApiEmum.UNINSTALL, machine),
					{
						username: userstore.current_user?.local_name
					}
				);
				if (result && result.status == 200 && result.data) {
					const userstore = useUserStore();
					const new_user = userstore.users!.items.get(userstore.current_id!)!;
					new_user.localMachine = '';
					await userstore.users!.items.update(new_user);
					await userstore.save();
					notifySuccess(i18n.global.t('Start uninstalling'));
					return true;
				}
				return false;
			} catch (error) {
				notifyRequestMessageError(error);
				return false;
			}
		},
		async rebootMachineTerminus(machine: TerminusServiceInfo) {
			const userstore = useUserStore();
			try {
				const instance = await this.createAxiosInstanceProxy(
					machine,
					{
						username: userstore.current_user?.local_name
					},
					commandDefaultTimeout
				);
				const result = await instance.post(
					getMdnsRequestApi(MdnsApiEmum.REBOOT, machine),
					{
						username: userstore.current_user?.local_name
					}
				);

				if (result && result.status == 200 && result.data) {
					if (result && result.data && result.data.message) {
						notifySuccess(i18n.global.t('Restarting'));
					}
					setTimeout(() => {
						this.rebootMachines[userstore.current_id!] = {
							rebootTime: new Date().getTime(),
							status: true
						};
						this.setActivedMachine(undefined);
					}, 1000);
					busEmit('terminus_update');
					return true;
				}
				return false;
			} catch (error) {
				this.rebootMachines[userstore.current_id!] = {
					rebootTime: -1,
					status: false
				};
				notifyRequestMessageError(error);
				return false;
			}
		},
		async shutdownMachineTerminus(machine: TerminusServiceInfo) {
			const userstore = useUserStore();
			try {
				const instance = await this.createAxiosInstanceProxy(
					machine,
					{
						username: userstore.current_user?.local_name
					},
					commandDefaultTimeout
				);
				const result = await instance.post(
					getMdnsRequestApi(MdnsApiEmum.SHUTDOWN, machine),
					{
						username: userstore.current_user?.local_name
					}
				);
				if (result && result.status == 200 && result.data) {
					if (result && result.data && result.data.message) {
						notifySuccess(i18n.global.t('Shutting down'));
					}
					setTimeout(() => {
						this.shutdownMachines[userstore.current_id!] = {
							shutdownTime: new Date().getTime(),
							status: true
						};
						this.setActivedMachine(undefined);
					}, 1000);
					busEmit('terminus_update');
					return true;
				}
				return false;
			} catch (error) {
				this.shutdownMachines[userstore.current_id!] = {
					shutdownTime: -1,
					status: false
				};
				notifyRequestMessageError(error);
				return false;
			}
		},

		async upgradeNeedRestartPrecheck(
			machine: TerminusServiceInfo,
			ok: () => void
		) {
			if (machine.version?.needRestart == true) {
				getNativeAppPlatform()
					.getQuasar()
					?.dialog({
						component: TerminusTipDialog,
						componentProps: {
							title: i18n.global.t('Restart required for upgrade'),
							message: i18n.global.t(
								'The system will restart automatically after the update is installed. During this process, Olares services will be unavailable for several minutes.<br>Please complete any ongoing tasks before proceeding.'
							),
							position: i18n.global.t('continue'),
							navigation: i18n.global.t('Not now'),
							descAlign: 'left'
						}
					})
					.onOk(async () => {
						ok();
					});
				return;
			}
			ok();
		},

		async upgradeOlares(machine: TerminusServiceInfo) {
			// const userstore = useUserStore();
			getNativeAppPlatform()
				.getQuasar()
				?.dialog({
					component: OlaresUpgradeReminderDialog,
					componentProps: {
						title: i18n.global.t('Select upgrade option')
					}
				})
				.onOk(async (upgradeMode: UpgradeMode) => {
					const onRequestUpgrade = async (upgradeMode: UpgradeMode) => {
						try {
							const upgradeInfo = {
								version: machine.version?.upgradeableVersion,
								wizardURL: machine.version?.wizardUrl,
								cliUrl: machine.version?.cliUrl,
								downloadOnly: upgradeMode === UpgradeMode.DOWNLOAD_ONLY
							};
							const instance = await this.createAxiosInstanceProxy(
								machine,
								upgradeInfo,
								commandDefaultTimeout
							);
							const result = await instance.post(
								getMdnsRequestApi(MdnsApiEmum.UPGRADE, machine),
								upgradeInfo
							);
							if (result && result.status == 200 && result.data) {
								this.initTerminusInfo(machine);
								return true;
							}
							return false;
						} catch (error) {
							notifyRequestMessageError(error);
							return false;
						}
					};
					if (upgradeMode == UpgradeMode.DOWNLOAD_AND_UPGRADE) {
						this.upgradeNeedRestartPrecheck(machine, () => {
							onRequestUpgrade(upgradeMode);
						});
					} else {
						onRequestUpgrade(upgradeMode);
					}
				});
		},
		async cancelUpgradeOlares(machine: TerminusServiceInfo) {
			try {
				const instance = await this.createAxiosInstanceProxy(
					machine,
					{
						version: machine.version?.upgradeableVersion
					},
					commandDefaultTimeout
				);
				const result = await instance.delete(
					getMdnsRequestApi(MdnsApiEmum.UPGRADE, machine)
				);
				if (result && result.status == 200 && result.data) {
					this.initTerminusInfo(machine);
					const activeM = this.activedMachine;
					if (activeM)
						this.checkLastOsVersion(activeM).then((version: any) => {
							if (version) {
								activeM.version = version;
							}
						});
					return true;
				}
				return false;
			} catch (error) {
				notifyRequestMessageError(error);
				return false;
			}
		},
		async confirmUpgradeOlares(machine: TerminusServiceInfo) {
			const onOk = async () => {
				try {
					const instance = await this.createAxiosInstanceProxy(
						machine,
						{
							version: machine.status?.upgradingTarget
						},
						commandDefaultTimeout
					);
					const result = await instance.post(
						getMdnsRequestApi(MdnsApiEmum.UPGRADE_CONFIRM, machine)
					);
					if (result && result.status == 200 && result.data) {
						this.initTerminusInfo(machine);
						const activeM = this.activedMachine;
						if (activeM)
							this.checkLastOsVersion(activeM).then((version: any) => {
								if (version) {
									activeM.version = version;
								}
							});
						return true;
					}
					return false;
				} catch (error) {
					notifyRequestMessageError(error);
					return false;
				}
			};

			if (machine.version?.needRestart == true) {
				this.upgradeNeedRestartPrecheck(machine, onOk);
			} else {
				getNativeAppPlatform()
					.getQuasar()
					?.dialog({
						component: TerminusTipDialog,
						componentProps: {
							title: i18n.global.t('Upgrade now'),
							message: i18n.global.t(
								'Upgrade to the new version of Olares {version}?',
								{
									version: machine.status?.upgradingTarget
								}
							)
						}
					})
					.onOk(onOk);
			}
		},

		async connectWifiMachineTerminus(
			machine: TerminusServiceInfo,
			wifi: { ssid: string; password: string }
		) {
			try {
				const instance = await this.createAxiosInstanceProxy(
					machine,
					wifi,
					commandDefaultTimeout
				);
				const result = await instance.post(
					getMdnsRequestApi(MdnsApiEmum.CONNECT_WIFI, machine),
					{
						...wifi
					}
				);
				if (result && result.status == 200 && result.data) {
					return true;
				}
				return false;
			} catch (error) {
				notifyRequestMessageError(error);
				return false;
			}
		},

		async getSystemIfs(machine: TerminusServiceInfo): Promise<SystemIPS[]> {
			try {
				const instance = await this.createAxiosInstanceProxy(
					machine,
					undefined
				);
				const result = await instance.get(
					getMdnsRequestApi(MdnsApiEmum.IFS, machine)
				);

				if (result && result.status == 200 && result.data) {
					return result.data.data;
				}
				return [];
			} catch (error) {
				return [];
			}
		},

		async getBluetoothMachineWifi(machine?: TerminusServiceInfo) {
			if (!machine || !machine.deviceId) {
				return [];
			}
			return getNativeAppPlatform().bluetoothsGetWifiList(machine.deviceId);
		},

		async bluetoothConnectWifi(
			machine: TerminusServiceInfo,
			ssid: string,
			password: string
		) {
			if (!machine.deviceId) {
				return {
					status: false,
					message: ''
				};
			}
			return getNativeAppPlatform().bluetoothConnectWifi(
				machine.deviceId!,
				ssid,
				password
			);
		},

		async changeHostMachineTerminus(
			machine: TerminusServiceInfo,
			info: { ip: string }
		) {
			try {
				const instance = await this.createAxiosInstanceProxy(
					machine,
					info,
					commandDefaultTimeout
				);
				const result = await instance.post(
					getMdnsRequestApi(MdnsApiEmum.CHANGE_HOST, machine),
					info
				);
				if (result && result.status == 200 && result.data) {
					notifySuccess(i18n.global.t('success to change host ip'));
					const userstore = useUserStore();
					this.ipChangingsMachines[userstore.current_id!] = {
						ipChangingTime: new Date().getTime(),
						status: true
					};
					return true;
				}
				return false;
			} catch (error) {
				notifyRequestMessageError(error);
				return false;
			}
		},

		async createAxiosInstanceProxy(
			info: { host: string; port: number; isSettingsServer: boolean },
			body: any,
			timeout = 10000
		) {
			if (info.isSettingsServer) {
				return await this.getOlaresStatusInfoInstance(body);
			}
			return await this.createMndsAxiosInstanceProxy(info, body, timeout);
		},
		async createMndsAxiosInstanceProxy(
			info: {
				host: string;
				port: number;
				isSettingsServer: boolean;
			},
			body: any,
			timeout = 10000
		) {
			const userStore = useUserStore();
			let jws: string | undefined = userStore.current_user?.isLargeVersion12
				? undefined
				: userStore.current_user?.name;
			const baseURL = `http://${info.host}:${info.port}`;
			if (body) {
				jws = await this.getJws(baseURL, body);
			}
			const instance = axios.create({
				baseURL: baseURL,
				timeout: timeout,
				headers: {
					'Content-Type': 'application/json',
					'X-Signature': jws as string
				}
			});
			return instance;
		},
		startSearchMdnsService() {
			this.startRefreshActivedMachine();
			const activeM = this.activedMachine;
			if (!Platform.is.nativeMobile || (!this.mdnsUsed && activeM)) {
				return;
			}
			DNSService.start().catch((e) => {
				if (e.message == 'PolicyDenied') {
					getNativeAppPlatform()
						.getQuasar()
						?.dialog({
							component: TerminusTipDialog,
							componentProps: {
								title: i18n.global.t('Local network permissions'),
								message: i18n.global.t(
									'You do not have permission to grant LarePass access to local network. Do you want to open it?'
								)
							}
						})
						.onOk(() => {
							DNSService.stop();
							getNativeAppPlatform().openAppSettings();
						})
						.onCancel(() => {
							getNativeAppPlatform().getRouter()?.back();
						});
				} else {
					notifyRequestMessageError(e);
				}
			});
			busOn('appStateChange', appStateChange);
		},

		async startSearchBluetoothDevice() {
			getNativeAppPlatform().startSearchBluethooth();
		},

		stopSearchMdnsService() {
			this.endRefreshActivedMachine();
			if (!Platform.is.nativeMobile) {
				return;
			}
			DNSService.stop();
			this.mdnsMachines = [];
			busOff('appStateChange', appStateChange);
		},

		async stopSearchBluetoothDevice() {
			getNativeAppPlatform().stopSearchBluetooth();
		},

		setActivedMachine(
			activedMachine?: TerminusServiceInfo,
			ms?: number,
			autoRefresh = true
		) {
			if (activedMachine) {
				const lastRebootTime = this.lastRebootTime();
				const lasShutdownTime = this.lastShutdownTime();
				if (
					(lastRebootTime > 0 && lastRebootTime < shutdownOrRebootLimitTime) ||
					(lasShutdownTime > 0 && lasShutdownTime < shutdownOrRebootLimitTime)
				) {
					return;
				}
			}
			this.resetCurrentTerminusStatus();
			const activeMachine = this.activedMachine;
			if (
				this.mdnsUsed ||
				!activeMachine ||
				activeMachine !== this.apiMachine
			) {
				this.mdnsMachine = activedMachine;
			} else {
				this.apiMachine = activedMachine;
			}
			if (activedMachine && autoRefresh) {
				this.startRefreshActivedMachine(ms);
			} else {
				if (!activedMachine) {
					this.endRefreshActivedMachine();
				}
			}
		},

		startRefreshActivedMachine(ms = 5000) {
			if (this.refreshTimer) {
				return;
			}
			const activedM = this.activedMachine;

			if (activedM) {
				this.initTerminusInfo(activedM);
			}
			this.refreshTimer = setInterval(async () => {
				const activedM = this.activedMachine;
				if (!activedM) {
					clearInterval(this.refreshTimer);
					this.refreshTimer = undefined;
					return;
				}
				await this.initTerminusInfo(activedM);
				this.resetCurrentTerminusStatus();
			}, ms);
		},

		endRefreshActivedMachine() {
			const activeMachine = this.activedMachine;
			if (this.mdnsUsed || !activeMachine || activeMachine != this.apiMachine) {
				this.mdnsMachine = undefined;
			}

			if (this.refreshTimer) {
				clearInterval(this.refreshTimer);
				this.refreshTimer = undefined;
			}
		},
		resetCurrentTerminusStatus() {
			const userStore = useUserStore();
			if (!userStore.current_id) {
				return false;
			}
			const lastRebootTime = this.lastRebootTime();
			const lasShutdownTime = this.lastShutdownTime();
			if (
				(lastRebootTime > 0 && lastRebootTime >= shutdownOrRebootLimitTime) ||
				(lasShutdownTime > 0 && lasShutdownTime >= shutdownOrRebootLimitTime)
			) {
				this.rebootMachines[userStore.current_id] = {
					status: false,
					rebootTime: -1
				};
				this.shutdownMachines[userStore.current_id] = {
					status: false,
					shutdownTime: -1
				};
			}
			const lastIpChangingTime = this.lastIpChangingTime();
			if (
				lastIpChangingTime > 0 &&
				lastIpChangingTime >= shutdownOrRebootLimitTime
			) {
				this.ipChangingsMachines[userStore.current_id] = {
					status: false,
					ipChangingTime: -1
				};
			}

			if (this.activedMachine && this.activedMachine.isSettingsServer) {
				console.log(
					'this.activedMachine.errorCount ===>',
					this.activedMachine.errorCount
				);

				if (
					this.activedMachine.errorCount != undefined &&
					this.activedMachine.errorCount > 0
				) {
					if (this.serverRequestMachines[userStore.current_id] == undefined) {
						this.serverRequestMachines[userStore.current_id] = {
							errorStartTime: new Date().getTime(),
							isError: false
						};
					} else {
						this.serverRequestMachines[userStore.current_id].isError =
							new Date().getTime() -
								this.serverRequestMachines[userStore.current_id]
									.errorStartTime >
							serverApiErrorLimitTime;
					}
				} else {
					delete this.serverRequestMachines[userStore.current_id];
				}
			}
		},
		lastRebootTime() {
			const userStore = useUserStore();
			if (!userStore.current_id) {
				return -1;
			}
			return this.rebootMachines[userStore.current_id] &&
				this.rebootMachines[userStore.current_id].rebootTime > 0
				? new Date().getTime() -
						this.rebootMachines[userStore.current_id].rebootTime
				: -1;
		},
		lastShutdownTime() {
			const userStore = useUserStore();
			if (!userStore.current_id) {
				return -1;
			}
			return this.shutdownMachines[userStore.current_id] &&
				this.shutdownMachines[userStore.current_id].shutdownTime > 0
				? new Date().getTime() -
						this.shutdownMachines[userStore.current_id].shutdownTime
				: -1;
		},
		lastIpChangingTime() {
			const userStore = useUserStore();
			if (!userStore.current_id) {
				return -1;
			}
			return this.ipChangingsMachines[userStore.current_id] &&
				this.ipChangingsMachines[userStore.current_id].ipChangingTime > 0
				? new Date().getTime() -
						this.ipChangingsMachines[userStore.current_id].ipChangingTime
				: -1;
		},
		isIpChanging() {
			const userStore = useUserStore();
			if (!userStore.current_id) {
				return false;
			}
			const status = this.ipChangingsMachines[userStore.current_id]
				? this.ipChangingsMachines[userStore.current_id].status
				: false;
			return status;
		},

		checkContainerMode() {
			const activedM = this.activedMachine;
			if (!activedM || !activedM.status) {
				return false;
			}
			const containMode = activedM.status.containerMode != undefined;
			if (containMode) {
				notifyFailed(
					i18n.global.t(
						'Operation is not available for the current Olares environment.'
					)
				);
			}
			return containMode;
		},
		async checkLastOsVersion(
			machine: TerminusServiceInfo,
			testLevel = UpgradeLevel.NONE
		) {
			try {
				const userStore = useUserStore();
				const version = machine.status?.terminusVersion;
				const deviceId = userStore.current_user?.olares_device_id;
				const sourceId = getSourceId(machine.status);

				const data = {
					olaresId: userStore.current_user?.name,
					version,
					deviceId,
					sourceId,
					testLevel,
					osType: machine.status?.os_type,
					osArch: machine.status?.os_arch,
					appVersion: (await getNativeAppPlatform().getDeviceInfo()).appVersion
				};

				const axiosP = axiosInstanceProxy({
					baseURL: this.versionRequestBaseUrl(),
					timeout: 1000 * 30,
					headers: {
						'Content-Type': 'application/json'
					}
				});

				const result = await axiosP.post('/getUpgradeableVersion', data);
				if (result && result.status == 200 && result.data.code == 200) {
					return result.data.data;
				}
				return undefined;
			} catch (error) {
				console.log(error);
				return undefined;
			}
		},
		versionRequestBaseUrl() {
			if (process.env.NODE_ENV == 'development') {
				return '';
			}
			const userStore = useUserStore();
			if (userStore.defaultDomain == 'cn') {
				return 'https://api.olares.cn/upgrade';
			}
			return 'https://api.olares.com/upgrade';
		},

		async getOlaresInfo() {
			try {
				const res = await (
					await this.getOlaresStatusInfoInstance()
				).get('/api/system/status');

				if (res && res.data && res.data.data) {
					this.apiMachine = {
						serviceName: '',
						host: '',
						port: 0,
						status: res.data.data,
						isSettingsServer: true
					};
				} else {
					if (
						res &&
						res.data &&
						res.data.message &&
						res.data.message.startsWith('connect ECONNREFUSED')
					) {
						this.apiMachine = {
							serviceName: '',
							host: '',
							port: 0,
							isSettingsServer: true
						};
					} else {
						this.apiMachine = undefined;
					}
				}
			} catch (error) {
				/* empty */
				console.log('error ===>', error);
				this.apiMachine = undefined;
			}
		},
		async getOlaresStatusInfoInstance(body?: any) {
			let baseURL = '';
			const userStore = useUserStore();
			if (process.env.NODE_ENV !== 'development') {
				baseURL = userStore.getModuleSever('settings');
			}
			let jws: string | undefined = userStore.current_user?.isLargeVersion12
				? undefined
				: userStore.current_user?.name;
			if (body) {
				jws = await this.getJws(baseURL, body);
			}

			const instance = axiosInstanceProxy({
				baseURL: baseURL,
				timeout: 1000 * 30,
				headers: {
					'Content-Type': 'application/json',
					'X-Signature': jws as any
				}
			});
			return instance;
		},
		async getJws(baseUrl: string, body: any) {
			const userStore = useUserStore();

			const mneminicItem = userStore.current_mnemonic;
			const did = await getDID(mneminicItem?.mnemonic);
			const privateJWK: PrivateJwk | undefined = await getPrivateJWK(
				mneminicItem?.mnemonic
			);

			let jws = userStore.current_user?.name;
			if (body) {
				jws = await signJWS(
					did,
					{
						did: did,
						name: userStore.current_user?.name,
						time: `${new Date().getTime()}`,
						domain: baseUrl,
						challenge: 'challenge',
						body: body
					},
					privateJWK
				);
			}
			return jws;
		},
		async getFrpList(username: string) {
			try {
				if (this.frpMap[username]) {
					return this.frpMap[username];
				}

				const frpBaseUrl =
					GolbalHost.FRP_LIST_URL[GolbalHost.userNameToEnvironment(username)];

				const instance = axios.create({
					baseURL: frpBaseUrl,
					timeout: 1000 * 10,
					headers: {}
				});
				const response = await instance.post('/v2/servers', {
					name: username
				});
				if (response.status == 200) {
					if (response.data.length > 0) {
						this.frpMap[username] = response.data;
					}
					return response.data as OlaresTunneV2Interface[];
				}
				return undefined;
			} catch (error) {
				return undefined;
			}
		},
		isNewActivedMineVersion(version: string) {
			return compareOlaresVersion(version, newActivedMineVersion).compare >= 0;
		}
	}
});

const appStateChange = async (state: { isActive: boolean }) => {
	const mdnsStore = useMDNSStore();
	if (state.isActive) {
		mdnsStore.startSearchMdnsService();
	} else {
		DNSService.stop();
	}
};
