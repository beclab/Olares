import {
	APP_STATUS,
	AppStatusInfo,
	AppStatusLatest,
	PAYMENT_STATUS
} from 'src/constant/constants';
import { TerminusEntrance } from '@bytetrade/core';
import { Version } from 'src/utils/version';
import { useCenterStore } from 'src/stores/market/center';

export enum CFG_TYPE {
	APPLICATION = 'app',
	MIDDLEWARE = 'middleware'
}

export function showIcon(cfgType: string): boolean {
	return cfgType === CFG_TYPE.MIDDLEWARE || cfgType === CFG_TYPE.APPLICATION;
}

export function canOpen(app: AppStatusInfo): TerminusEntrance | undefined {
	let entrance: TerminusEntrance | undefined = undefined;
	if (app.entranceStatuses && app.entranceStatuses.length > 0) {
		entrance = app.entranceStatuses.find((item) => !item.invisible);
	}
	return entrance;
}

export function canStop(app: AppStatusInfo): boolean {
	return (
		app &&
		(app.state === APP_STATUS.RUNNING || app.state === APP_STATUS.STOP.FAILED)
	);
}

export function canStopState(state: string): boolean {
	return state === APP_STATUS.RUNNING || state === APP_STATUS.STOP.FAILED;
}

export function canResume(app: AppStatusInfo): boolean {
	return (
		app &&
		(app.state === APP_STATUS.STOP.COMPLETED ||
			app.state === APP_STATUS.RESUME.FAILED)
	);
}

export function canResumeState(state: string): boolean {
	return (
		state === APP_STATUS.STOP.COMPLETED || state === APP_STATUS.RESUME.FAILED
	);
}

export function showDownloadProgress(app: AppStatusInfo): boolean {
	return (
		app.state === APP_STATUS.DOWNLOAD.DEFAULT ||
		app.state === APP_STATUS.UPGRADE.DEFAULT
	);
}

const showAppTextStatus = new Set<string>([
	APP_STATUS.PENDING.CANCELING,
	APP_STATUS.DOWNLOAD.CANCELING,
	APP_STATUS.INSTALL.CANCELING,
	APP_STATUS.INITIALIZE.CANCELING,
	APP_STATUS.UPGRADE.CANCELING,
	APP_STATUS.RESUME.DEFAULT,
	APP_STATUS.RESUME.CANCELING,
	APP_STATUS.STOP.DEFAULT,
	APP_STATUS.PENDING.DEFAULT,
	APP_STATUS.UNINSTALL.DEFAULT,
	APP_STATUS.UPGRADE.DEFAULT,
	APP_STATUS.INITIALIZE.DEFAULT,
	APP_STATUS.INSTALL.DEFAULT,
	APP_STATUS.PENDING.CANCEL_FAILED,
	APP_STATUS.DOWNLOAD.CANCEL_FAILED,
	APP_STATUS.DOWNLOAD.FAILED,
	APP_STATUS.INSTALL.CANCEL_FAILED,
	APP_STATUS.INSTALL.FAILED,
	APP_STATUS.UNINSTALL.FAILED,
	APP_STATUS.UPGRADE.FAILED,
	APP_STATUS.ENV.APPLYING,
	APP_STATUS.ENV.CANCELING,
	APP_STATUS.ENV.APPLY_FAILED,
	APP_STATUS.ENV.CANCEL_FAILED,
	APP_STATUS.RESUME.FAILED,
	APP_STATUS.STOP.FAILED
]);

const canInstallingCancelAppStates = new Set<string>([
	APP_STATUS.PENDING.DEFAULT,
	APP_STATUS.DOWNLOAD.DEFAULT,
	APP_STATUS.INSTALL.DEFAULT,
	APP_STATUS.INITIALIZE.DEFAULT
]);

export function canInstallingCancel(app: AppStatusInfo): boolean {
	return app && canInstallingCancelAppStates.has(app.state);
}

export function showAppStatus(app: AppStatusInfo): boolean {
	return app && showAppTextStatus.has(app.state);
}

const isCancelingAppStates = new Set<string>([
	APP_STATUS.PENDING.CANCELING,
	APP_STATUS.DOWNLOAD.CANCELING,
	APP_STATUS.INSTALL.CANCELING,
	APP_STATUS.INITIALIZE.CANCELING,
	APP_STATUS.ENV.CANCELING,
	APP_STATUS.UPGRADE.CANCELING,
	APP_STATUS.RESUME.CANCELING
]);

export function isCanceling(app: AppStatusInfo): boolean {
	return app && isCancelingAppStates.has(app.state);
}

const isDoingAppStates = new Set<string>([
	APP_STATUS.STOP.DEFAULT,
	APP_STATUS.PENDING.DEFAULT,
	APP_STATUS.UNINSTALL.DEFAULT,
	APP_STATUS.DOWNLOAD.DEFAULT,
	APP_STATUS.INSTALL.DEFAULT,
	APP_STATUS.ENV.APPLYING,
	APP_STATUS.INITIALIZE.DEFAULT,
	APP_STATUS.UPGRADE.DEFAULT,
	APP_STATUS.RESUME.DEFAULT
]);

export function isDoing(app: AppStatusInfo): boolean {
	return app && isDoingState(app.state);
}

export function isDoingState(state: string): boolean {
	return isDoingAppStates.has(state);
}

const isUpgradableAppStates = new Set<string>([
	APP_STATUS.RUNNING,
	APP_STATUS.STOP.COMPLETED,
	APP_STATUS.STOP.FAILED,
	APP_STATUS.UPGRADE.FAILED,
	APP_STATUS.ENV.APPLY_FAILED
]);

export function isUpgradable(app: AppStatusInfo): boolean {
	return app && isUpgradableAppStates.has(app.state);
}

const uninstallableAppStates = new Set<string>([
	APP_STATUS.RUNNING,
	APP_STATUS.STOP.COMPLETED,
	APP_STATUS.STOP.FAILED,
	APP_STATUS.UPGRADE.FAILED,
	APP_STATUS.RESUME.FAILED,
	APP_STATUS.MODEL.INSTALLED,
	APP_STATUS.ENV.APPLY_FAILED
]);

export function isUninstallable(app: AppStatusInfo): boolean {
	return app && uninstallableAppStates.has(app.state);
}

export function isUninstallableState(state: string): boolean {
	return uninstallableAppStates.has(state);
}

const uninstalledAppStates = new Set<string>([
	APP_STATUS.PENDING.CANCELED,
	APP_STATUS.DOWNLOAD.CANCELED,
	APP_STATUS.DOWNLOAD.FAILED,
	APP_STATUS.INSTALL.FAILED,
	APP_STATUS.INSTALL.CANCELED,
	APP_STATUS.UNINSTALL.COMPLETED
]);

export function uninstalledApp(app: AppStatusInfo): boolean {
	return app && uninstalledAppStates.has(app.state);
}

export function isCloneApp(app: AppStatusInfo): boolean {
	return app && !!app.rawAppName && app.rawAppName !== app.name;
}

export function uninstalledAppState(state: string): boolean {
	return uninstalledAppStates.has(state);
}

const failedAppStates = new Set<string>([
	APP_STATUS.PENDING.CANCEL_FAILED,
	APP_STATUS.DOWNLOAD.CANCEL_FAILED,
	APP_STATUS.DOWNLOAD.FAILED,
	APP_STATUS.INSTALL.CANCEL_FAILED,
	APP_STATUS.INSTALL.FAILED,
	APP_STATUS.UNINSTALL.FAILED,
	APP_STATUS.ENV.APPLY_FAILED,
	APP_STATUS.ENV.CANCEL_FAILED,
	APP_STATUS.UPGRADE.FAILED,
	APP_STATUS.RESUME.FAILED,
	APP_STATUS.STOP.FAILED
]);

export function isFailed(app: AppStatusInfo): boolean {
	return app && failedAppStates.has(app.state);
}

export function getBaseTextByState(state: string): string {
	const textMap = {
		[APP_STATUS.UNINSTALL.DEFAULT]: 'app.uninstalling',
		[APP_STATUS.STOP.DEFAULT]: 'app.stopping',
		[APP_STATUS.PENDING.DEFAULT]: 'app.pending',
		[APP_STATUS.DOWNLOAD.DEFAULT]: 'app.downloading',
		[APP_STATUS.INSTALL.DEFAULT]: 'app.installing',
		[APP_STATUS.INITIALIZE.DEFAULT]: 'app.initializing',
		[APP_STATUS.UPGRADE.DEFAULT]: 'app.updating',
		[APP_STATUS.RESUME.DEFAULT]: 'app.resuming',
		[APP_STATUS.ENV.APPLYING]: 'app.applying'
	};
	return textMap[state] || 'app.unknown';
}

export function getErrorTextByState(state: string): string {
	const errorTextMap = {
		[APP_STATUS.PENDING.CANCEL_FAILED]: 'app.pendingCancelFailed',
		[APP_STATUS.DOWNLOAD.CANCEL_FAILED]: 'app.downloadCancelFailed',
		[APP_STATUS.DOWNLOAD.FAILED]: 'app.downloadFailed',
		[APP_STATUS.INSTALL.CANCEL_FAILED]: 'app.installCancelFailed',
		[APP_STATUS.INSTALL.FAILED]: 'app.installFailed',
		[APP_STATUS.UNINSTALL.FAILED]: 'app.uninstallFailed',
		[APP_STATUS.UPGRADE.FAILED]: 'app.upgradeFailed',
		[APP_STATUS.RESUME.FAILED]: 'app.resumeFailed',
		[APP_STATUS.STOP.FAILED]: 'app.stopFailed',
		[APP_STATUS.ENV.APPLY_FAILED]: 'app.envApplyFailed',
		[APP_STATUS.ENV.CANCEL_FAILED]: 'app.envCancelFailed'
	};
	return errorTextMap[state] ?? '';
}

export function getPaymentTextByState(state: string): string {
	const paymentTextMap = {
		[PAYMENT_STATUS.NOT_BUY]: 'app.buy',
		[PAYMENT_STATUS.SIGNATURE_NEED_RESIGN]: 'app.sign',
		[PAYMENT_STATUS.SIGNATURE_REQUIRED]: 'app.sign',
		[PAYMENT_STATUS.PAYMENT_REQUIRED]: 'app.pay',
		[PAYMENT_STATUS.WAITING_DEVELOPER_CONFIRMATION]: 'app.reviewing',
		[PAYMENT_STATUS.PAYMENT_RETRY_REQUIRED]: 'app.view'
	};
	return paymentTextMap[state] ?? 'app.failed';
}

export function canUpgrade(
	statusLatest: AppStatusLatest,
	appId: string,
	sourceId: string
) {
	if (isUpgradable(statusLatest.status)) {
		const centerStore = useCenterStore();
		const simpleLatest = isCloneApp(statusLatest.status)
			? centerStore.getAppSimpleInfo(statusLatest.status.rawAppName, sourceId)
			: centerStore.getAppSimpleInfo(appId, sourceId);
		if (
			simpleLatest &&
			simpleLatest.app_simple_info.app_version &&
			statusLatest.version &&
			simpleLatest.app_simple_info.app_version !== statusLatest.version
		) {
			try {
				const latestVersion = new Version(
					simpleLatest.app_simple_info.app_version
				);
				const myVersion = new Version(statusLatest.version);
				if (latestVersion.isGreater(myVersion)) {
					return true;
				}
			} catch (e) {
				return false;
			}
		}
	}

	return false;
}
