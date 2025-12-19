import { i18n } from 'src/boot/i18n';
import { getRequireImage } from '../../../utils/imageUtils';
import { useMDNSStore } from 'src/stores/mdns';
import { VersionInfo } from 'src/constant';

export enum TerminusStatusEnum {
	NotInstalled = 'not-installed',
	Installing = 'installing',
	Uninitialized = 'uninitialized',
	Initializing = 'initializing',
	TerminusRunning = 'terminus-running',
	InvalidIpAddress = 'invalid-ip-address',
	SystemError = 'system-error',
	SelfRepairing = 'self-repairing',
	IPChanging = 'ip-changing',
	AddingNode = 'adding-node',
	RemovingNode = 'removing-node',
	Uninstalling = 'uninstalling',
	DiskModifing = 'disk-modifing',
	Shutdown = 'shutdown',
	InstallFailed = 'install-failed',
	Restarting = 'restarting',
	Checking = 'checking',
	networkNotReady = 'network-not-ready',
	Upgrading = 'upgrading'
}

export const canUnInstall = (terminusStatus: TerminusStatus) => {
	const terminusState = terminusStatus.terminusState;
	return (
		terminusStatus.installingState == InstallingState.Failed ||
		terminusState == TerminusStatusEnum.InstallFailed ||
		terminusState == TerminusStatusEnum.SystemError
	);
};

export const disableOperate = (terminusStatus: TerminusStatus) => {
	const terminusState = terminusStatus.terminusState;
	return terminusState == TerminusStatusEnum.IPChanging;
};

export const isInstalling = (terminusStatus: TerminusStatus) => {
	const terminusState = terminusStatus.terminusState;
	return (
		(terminusState == TerminusStatusEnum.Installing &&
			terminusStatus.installingState != InstallingState.Failed) ||
		(terminusState == TerminusStatusEnum.NotInstalled &&
			terminusStatus.installingProgress != '')
	);
};

export const canInstall = (terminusStatus: TerminusStatus) => {
	const terminusState = terminusStatus.terminusState;
	return (
		terminusState == TerminusStatusEnum.NotInstalled &&
		(terminusStatus.installingProgress == '' ||
			terminusStatus.installingProgress == undefined) &&
		(terminusStatus.installingState == undefined ||
			terminusStatus.installingState != InstallingState.Failed)
	);
};

export const isChecking = (terminusStatus: TerminusStatus) => {
	// return terminusStatus.terminusState ==
	const terminusState = terminusStatus.terminusState;
	return (
		terminusState == TerminusStatusEnum.Checking ||
		terminusState == TerminusStatusEnum.networkNotReady
	);
};

export const canActive = (terminusStatus: TerminusStatus) => {
	const terminusState = terminusStatus.terminusState;
	return (
		terminusState == TerminusStatusEnum.Uninitialized &&
		terminusStatus.installingState != InstallingState.Failed
	);
};
export const displayStatus = (terminusStatus: TerminusStatus) => {
	const mdnsStore = useMDNSStore();
	if (mdnsStore.isIpChanging()) {
		return {
			color: '',
			status: i18n.global.t('ip-changing')
		};
	}
	if (terminusStatus.terminusState == TerminusStatusEnum.TerminusRunning) {
		return {
			color: 'positive',
			status: i18n.global.t('olares-running')
		};
	}
	if (terminusStatus.terminusState == TerminusStatusEnum.InvalidIpAddress) {
		return {
			color: 'negative',
			status: i18n.global.t('invalid-ip-address')
		};
	}
	if (terminusStatus.terminusState == TerminusStatusEnum.SystemError) {
		return {
			color: 'negative',
			status: i18n.global.t('system-error')
		};
	}

	if (terminusStatus.terminusState == TerminusStatusEnum.InstallFailed) {
		return {
			color: 'negative',
			status: 'InstallFailed'
		};
	}
	if (terminusStatus.terminusState == TerminusStatusEnum.SelfRepairing) {
		return {
			color: 'warning',
			status: i18n.global.t('self-repairing')
		};
	}
	if (terminusStatus.terminusState == TerminusStatusEnum.Upgrading) {
		return {
			color: 'negative',
			status: i18n.global.t('upgrading')
		};
	}

	if (terminusStatus.terminusState == TerminusStatusEnum.IPChanging) {
		return {
			color: '',
			status: i18n.global.t('ip-changing')
		};
	}
	if (terminusStatus.terminusState == TerminusStatusEnum.Shutdown) {
		return {
			color: '',
			status: i18n.global.t('Shutdown')
		};
	}

	if (terminusStatus.terminusState == TerminusStatusEnum.Restarting) {
		return {
			color: '',
			status: i18n.global.t('Restarting')
		};
	}

	if (isInstalling(terminusStatus) || canInstall(terminusStatus)) {
		return {
			color: 'negative',
			status: i18n.global.t('Not Installed')
		};
	}

	if (canActive(terminusStatus)) {
		return {
			color: 'negative',
			status: i18n.global.t('Not activated')
		};
	}

	return {
		color: 'negative',
		status: terminusStatus.terminusState
	};
};

export const installDisplayStatus = (terminusStatus: TerminusStatus) => {
	if (
		terminusStatus.installingState == InstallingState.Failed ||
		terminusStatus.terminusState == TerminusStatusEnum.InstallFailed
	) {
		return {
			textClass: 'text-negative',
			icon: 'sym_r_info',
			status: 'Install error, please Uninstall'
		};
	}

	if (isInstalling(terminusStatus) || canInstall(terminusStatus)) {
		return {
			textClass: 'text-negative',
			icon: 'sym_r_info',
			status: i18n.global.t('Not Installed')
		};
	}

	if (canActive(terminusStatus)) {
		return {
			textClass: 'text-negative',
			icon: 'sym_r_info',
			status: i18n.global.t('Not activated')
		};
	}

	if (terminusStatus.terminusState == TerminusStatusEnum.Checking) {
		return {
			icon: 'sym_r_check_circle',
			textClass: 'sym_r_info',
			status: i18n.global.t('Checking')
		};
	}

	if (terminusStatus.terminusState == TerminusStatusEnum.networkNotReady) {
		return {
			icon: 'sym_r_check_circle',
			textClass: 'sym_r_info',
			status: i18n.global.t('Network Not Ready')
		};
	}

	if (terminusStatus.terminusState == TerminusStatusEnum.TerminusRunning) {
		return {
			icon: 'sym_r_check_circle',
			textClass: 'text-positive',
			status: i18n.global.t('Activated')
		};
	}

	return {
		icon: 'sym_r_info',
		textClass: 'text-negative',
		status: terminusStatus.terminusState
	};
};

export const systemInfo = (status: TerminusStatus) => {
	return status.os_type + '_' + status.os_arch + '_' + status.os_version;
};

export enum InstallingState {
	Completed = 'completed',
	Failed = 'failed',
	InProgress = 'in-progress'
}
// "os_type": "linux",
// "os_arch": "amd64",
// "os_info": "",
// "os_version": "22.04"
export interface TerminusStatus {
	installingProgress: string;
	installingState: InstallingState;
	terminusState: TerminusStatusEnum;
	terminusdState: string;
	updateTime: number;
	uninstallingProgress: string;
	os_type: string;
	os_arch: string;
	os_info: string;
	os_version: string;
	terminusName: string;
	terminusVersion: string;
	device_name: string;
	host_name: string;
	cpu_info: string;
	gpu_info?: string;
	memory: string;
	disk: string;
	wifiConnected: boolean;
	wifiSSID: string;
	wiredConnected: boolean;
	hostIp: string;
	externalIp: string;
	installedTime: number;
	initializedTime: number;
	defaultFrpServer: string;
	frpEnable: string;
	containerMode?: 'ioc';
	upgradingDownloadState?: string;
	upgradingTarget?: string;
	upgradingDownloadProgress?: string;
	upgradingDownloadStep?: string;
	upgradingDownloadError?: string;
	upgradingProgress?: string;
	upgradingState?: string; //WaitingForUserConfirm | 'error'
	upgradingError?: string;
	mode?: string;
}

export const deviceLogo = (status?: TerminusStatus) => {
	if (!status) {
		return getRequireImage('wizard/devices/pve.png');
	}
	if (status.device_name && status.device_name.startsWith('Olares')) {
		if (status.device_name == 'Olares One') {
			return getRequireImage('wizard/devices/olares-one.png');
		}
		return getRequireImage('wizard/devices/terminus-zero.png');
	}
	if (status.os_type == 'mac') {
		return getRequireImage('wizard/devices/macos.png');
	}
	if (status.os_type == 'windows') {
		return getRequireImage('wizard/devices/windows.png');
	}
	if (status.os_type.startsWith('Raspberry Pi')) {
		return getRequireImage('wizard/devices/raspberrypi.png');
	}
	return getRequireImage('wizard/devices/pve.png');
};

export enum OlaresSourceType {
	MAIN = 'main',
	ONE = 'one'
}

export const getSourceId = (status?: TerminusStatus) => {
	if (status && status.device_name && status.device_name == 'Olares One') {
		return OlaresSourceType.ONE;
	}
	return OlaresSourceType.MAIN;
};

export const isOlaresGlobalDevice = (status?: TerminusStatus) => {
	if (!status) {
		return false;
	}
	return !!status.device_name && status.device_name.startsWith('Olares');
};

export enum OneWorkerMode {
	Quite = 0,
	Balance = 1,
	Gaming = 2,
	Insane = 3
}

export const oneWorkerModeOptions = () => {
	return [
		{
			label: i18n.global.t('Silent mode'),
			value: OneWorkerMode.Quite
		},
		// {
		// 	label: i18n.global.t('Balance mode'),
		// 	value: OneWorkerMode.Balance
		// },
		// {
		// 	label: i18n.global.t('Gaming mode'),
		// 	value: OneWorkerMode.Gaming
		// },
		{
			label: i18n.global.t('Performance Mode'),
			value: OneWorkerMode.Insane
		}
	];
};

export interface TerminusServiceInfo {
	serviceName: string;
	host: string;
	port: number;
	status?: TerminusStatus;
	password?: string;
	deviceId?: string;
	version?: VersionInfo;
	errorCount?: number;
	isSettingsServer: boolean;
}

export enum MdnsApiEmum {
	UNINSTALL = 'uninstall',
	STATUS = 'status',
	INSTALL = 'install',
	REBOOT = 'reboot',
	SHUTDOWN = 'shutdown',
	UPGRADE = 'upgrade',
	UPGRADE_CONFIRM = 'upgradeComfirm',
	CONNECT_WIFI = 'connectWifi',
	IFS = 'ifs',
	CHANGE_HOST = 'changeHost',
	CHECK_SSH_PASSWORD = 'checkSSHPassword',
	RESET_SSH_PASSWORD = 'resetSSHPassword',

	DID_USER_NAME = 'didUserName',

	//One
	ONE_GPU_MODE = 'oneGpuMode',
	ONE_CPU_GPU = 'oneCpuGpu'
}

const mdnsApis: Record<MdnsApiEmum, string> = {
	[MdnsApiEmum.UNINSTALL]: '/command/uninstall',
	status: '/system/status',
	install: '/command/install',
	reboot: '/command/reboot',
	shutdown: '/command/shutdown',
	upgrade: '/command/upgrade',
	upgradeComfirm: '/command/upgrade/confirm',
	connectWifi: '/command/connect-wifi',
	ifs: '/system/ifs',
	changeHost: '/command/change-host',
	checkSSHPassword: '/system/check-ssh-password',
	[MdnsApiEmum.RESET_SSH_PASSWORD]: '/command/ssh-password',
	[MdnsApiEmum.ONE_GPU_MODE]: '/olares-one/gpu-mode',
	[MdnsApiEmum.ONE_CPU_GPU]: '/olares-one/cpu-gpu',
	[MdnsApiEmum.DID_USER_NAME]: '/system/1.0/name'
};

export const getMdnsRequestApi = (
	api: MdnsApiEmum,
	serviceInfo?: TerminusServiceInfo
) => {
	if (!serviceInfo || !serviceInfo.isSettingsServer) {
		return mdnsApis[api] || api;
	}
	return getSettingsServerMdnsRequestApi(api);
};

export const getSettingsServerMdnsRequestApi = (api: MdnsApiEmum) => {
	return '/api/mdns' + mdnsApis[api];
};

export enum UpgradeLevel {
	ALPHA = 'alpha',
	BETA = 'beta',
	RC = 'rc', //default
	NONE = ''
}
