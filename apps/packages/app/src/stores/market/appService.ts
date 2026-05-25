import SetEnvironmentDialog from 'src/pages/market/application/SetEnvironmentDialog.vue';
import SetCloneInfoDialog from 'src/pages/market/application/SetCloneInfoDialog.vue';
import DependencyDialog from 'src/components/appintro/DependencyDialog.vue';
import { csAppStop, csAppUninstall } from 'src/stores/market/csAppOperation';
import { Action, Category, TerminusEntrance } from '@bytetrade/core';
import { openApplication } from 'src/api/market/private/server';
import { usePreCheckStore } from 'src/stores/market/preCheck';
import { useUserStore } from 'src/stores/settings/user';
import { UpdateEnvBody, CloneEntrance } from 'src/constant';
import { useAppStore } from 'src/stores/market/appStore';
import { suspendApp } from 'src/constant/config';
import { bus, BUS_EVENT } from 'src/utils/bus';
import { i18n } from 'src/boot/i18n';
import { QVueGlobals } from 'quasar';
import {
	APP_STATUS,
	AppStatusInfo,
	getI18nValue,
	isCSV2,
	STATUS_OPERATE_TYPE
} from 'src/constant/constants';
import {
	installApp,
	uninstallApp,
	resumeApp,
	stopApp,
	upgradeApp,
	OperationRequest,
	cancelInstalling,
	removeLocalPackage,
	CloneRequest,
	cloneApp,
	CSV2Request,
	BaseOperationRequest
} from 'src/api/market/private/operations';

export enum OPERATION_TYPE {
	INSTALL = 'install',
	CLONE = 'clone',
	UNINSTALL = 'uninstall',
	UPGRADE = 'upgrade',
	CANCEL = 'cancel',
	RESUME = 'resume',
	STOP = 'stop',
	REMOVE = 'remove'
}

const operationConfig = {
	[OPERATION_TYPE.INSTALL]: {
		action: (request: OperationRequest) =>
			installApp({ ...request, sync: true }),
		getStatus: () => APP_STATUS.PENDING.DEFAULT,
		logMessage: 'install app start'
	},
	[OPERATION_TYPE.CLONE]: {
		action: (request: CloneRequest) => cloneApp({ ...request, sync: true }),
		//old app running
		getStatus: () => APP_STATUS.RUNNING,
		logMessage: 'clone app start'
	},
	[OPERATION_TYPE.UNINSTALL]: {
		action: (request: CSV2Request) => uninstallApp({ ...request, sync: true }),
		getStatus: () => APP_STATUS.UNINSTALL.DEFAULT,
		logMessage: 'uninstall app start'
	},
	[OPERATION_TYPE.UPGRADE]: {
		action: (request: OperationRequest) =>
			upgradeApp({ ...request, sync: true }),
		getStatus: () => APP_STATUS.UPGRADE.DEFAULT,
		logMessage: 'upgrade app start'
	},
	[OPERATION_TYPE.CANCEL]: {
		action: (request: OperationRequest) =>
			cancelInstalling({ ...request, sync: true }),
		getStatus: () => APP_STATUS.INSTALL.CANCELING,
		logMessage: 'cancel app start'
	},
	[OPERATION_TYPE.REMOVE]: {
		action: (request: OperationRequest) => removeLocalPackage(request),
		getStatus: () => APP_STATUS.UNINSTALL.DEFAULT,
		logMessage: 'app removing'
	},
	[OPERATION_TYPE.RESUME]: {
		action: (request: BaseOperationRequest) => resumeApp(request.app_name),
		getStatus: () => APP_STATUS.RESUME.DEFAULT,
		logMessage: 'app resuming'
	},
	[OPERATION_TYPE.STOP]: {
		action: (request: BaseOperationRequest, all: boolean) =>
			stopApp(request.app_name, all),
		getStatus: () => APP_STATUS.STOP.DEFAULT,
		logMessage: 'app stopping'
	}
};

export async function processAppOperation(
	app: AppStatusInfo,
	operationType: OPERATION_TYPE,
	...args: any[]
): Promise<any> {
	const config = operationConfig[operationType];
	console.log(config.logMessage);

	const oldState = app.state;
	let opType = '';
	if (
		operationType === OPERATION_TYPE.INSTALL ||
		operationType === OPERATION_TYPE.CLONE
	) {
		opType = STATUS_OPERATE_TYPE.INSTALL;
	} else if (operationType === OPERATION_TYPE.UPGRADE) {
		opType = STATUS_OPERATE_TYPE.UPGRADE;
	}
	const appStore = useAppStore();
	appStore.updateAppStatus(
		args[0].app_name,
		args[0].source,
		config.getStatus(),
		opType
	);

	return new Promise((resolve, reject) => {
		config
			// @ts-ignore-next-line: A spread argument must either have a tuple type or be passed to a rest parameter
			.action(...args)
			.then((response) => {
				console.log(response);
				resolve(response);
			})
			.catch((e) => {
				appStore.updateAppStatus(args[0].app_name, args[0].source, oldState);
				console.error(e);

				//handle install/clone app with env
				const isInstall422 =
					(operationType === OPERATION_TYPE.INSTALL ||
						operationType === OPERATION_TYPE.CLONE) &&
					e.response?.data?.data?.backend_response?.code === 422;

				if (isInstall422) {
					reject({
						type: e.response.data.data.backend_response.data.type,
						data: e.response.data.data.backend_response.data.Data
					});
					return;
				}

				//handle clone app with name
				const isInstall104222 =
					operationType === OPERATION_TYPE.CLONE &&
					e.response?.data?.data?.backend_response?.code === 104222;

				if (isInstall104222) {
					reject({
						type: e.response.data.data.backend_response.data.type,
						data: e.response.data.data.backend_response.data.Data
					});
					return;
				}

				bus.emit(
					BUS_EVENT.APP_BACKEND_ERROR,
					e?.response?.data?.data?.backend_response?.message ||
						e?.response?.data?.message ||
						i18n.global.t('error.operation_preform_failure')
				);

				reject(e);
			});
	});
}

export async function handleWindowOperation(entrance: {
	id: string;
	url: string;
}) {
	if (window.top === window) {
		window.open(`//${entrance.url}`, '_blank');
	} else {
		try {
			await openApplication({
				action: Action.ACTION_VIEW,
				category: Category.CATEGORY_LAUNCHER,
				data: {
					appid: entrance.id
				}
			});
		} catch (e) {
			console.error(e);
			bus.emit(
				BUS_EVENT.APP_BACKEND_ERROR,
				e.response?.data?.message ||
					i18n.global.t('error.operation_preform_failure')
			);
		}
	}
}

const mergeEnvs = (
	originalEnvs: UpdateEnvBody = [],
	newEnvs: UpdateEnvBody = []
): UpdateEnvBody => {
	const envMap = new Map(originalEnvs.map((item) => [item.envName, item]));

	newEnvs.forEach((item) => {
		envMap.set(item.envName, item);
	});
	return Array.from(envMap.values());
};

const mergeEntrances = (
	originalEntrances: CloneEntrance[] = [],
	newEntrances: CloneEntrance[] = []
): CloneEntrance[] => {
	const entrancesMap = new Map(
		originalEntrances.map((item) => [item.name, item])
	);

	newEntrances.forEach((item) => {
		entrancesMap.set(item.name, item);
	});
	return Array.from(entrancesMap.values());
};

export const AppService = {
	installApp: async (
		app: AppStatusInfo,
		request: OperationRequest,
		q: QVueGlobals
	) => {
		return processAppOperation(app, OPERATION_TYPE.INSTALL, request)
			.then(() => {
				const fullLatest = useAppStore().getAppFullInfo(
					request.app_name,
					request.source
				);
				if (
					fullLatest &&
					usePreCheckStore().hasUninstallDependencies(
						fullLatest.app_info.app_entry
					)
				) {
					q.dialog({
						component: DependencyDialog,
						componentProps: {
							appName: request.app_name,
							sourceId: request.source
						}
					});
				}
			})
			.catch((error) => {
				console.log(error);
				if (error?.type === 'appenv') {
					const simpleLatest = useAppStore().getAppSimpleInfo(
						request.app_name,
						request.source
					);
					q.dialog({
						component: SetEnvironmentDialog,
						componentProps: {
							data: error.data,
							appTitle: getI18nValue(
								simpleLatest?.app_simple_info.app_title,
								i18n.global.locale.value
							)
						}
					}).onOk(async (data) => {
						if (data && data.envs) {
							await AppService.installApp(
								app,
								{
									...request,
									envs: mergeEnvs(request.envs, data.envs)
								},
								q
							);
						}
					});
				}
			});
	},

	cloneApp: async (
		app: AppStatusInfo,
		request: CloneRequest,
		q: QVueGlobals
	) => {
		return processAppOperation(app, OPERATION_TYPE.CLONE, request)
			.then(() => {
				const fullLatest = useAppStore().getAppFullInfo(
					request.app_name,
					request.source
				);
				if (
					fullLatest &&
					usePreCheckStore().hasUninstallDependencies(
						fullLatest.app_info.app_entry
					)
				) {
					q.dialog({
						component: DependencyDialog,
						componentProps: {
							appName: request.app_name,
							sourceId: request.source
						}
					});
				}
			})
			.catch((error) => {
				console.log(error);
				if (error?.type === 'appenv') {
					const simpleLatest = useAppStore().getAppSimpleInfo(
						request.app_name,
						request.source
					);
					q.dialog({
						component: SetEnvironmentDialog,
						componentProps: {
							data: error.data,
							appTitle:
								request.title ??
								getI18nValue(
									simpleLatest?.app_simple_info.app_title,
									i18n.global.locale.value
								)
						}
					}).onOk(async (data) => {
						if (data && data.envs) {
							await AppService.cloneApp(
								app,
								{
									...request,
									envs: mergeEnvs(request.envs, data.envs)
								} as CloneRequest,
								q
							);
						}
					});
				} else if (error?.type === 'appEntrance') {
					q.dialog({
						component: SetCloneInfoDialog,
						componentProps: {
							data: error.data
						}
					}).onOk(async (data2) => {
						if (data2) {
							await AppService.cloneApp(
								app,
								{
									...request,
									title: data2.title,
									entrances: mergeEntrances(request.entrances, data2.entrances)
								} as CloneRequest,
								q
							);
						}
					});
				}
			});
	},

	uninstallApp: async (
		app: AppStatusInfo,
		request: OperationRequest,
		q: QVueGlobals
	) => {
		const appStore = useAppStore();
		const preCheckStore = usePreCheckStore();
		const userStore = useUserStore();
		const aggregation = appStore.getAppAggregationInfo(
			request.app_name,
			request.source
		);
		const appName =
			getI18nValue(
				aggregation?.app_simple_latest?.app_simple_info.app_title,
				i18n.global.locale.value
			) ?? request.app_name;

		return csAppUninstall(
			q,
			appName,
			!!aggregation && isCSV2(aggregation.app_full_info),
			preCheckStore.hasAdminPermissions(),
			userStore.accounts.length
		).then((result: { all: boolean; clearData: boolean }) => {
			processAppOperation(app, OPERATION_TYPE.UNINSTALL, {
				...request,
				all: result.all,
				deleteData: result.clearData
			});
		});
	},

	upgradeApp: (app: AppStatusInfo, request: OperationRequest) => {
		const appStore = useAppStore();
		const simpleLatest = appStore.getAppSimpleInfo(
			request.app_name,
			request.source,
			app
		);
		if (suspendApp(simpleLatest)) {
			bus.emit(
				BUS_EVENT.APP_BACKEND_ERROR,
				i18n.global.t('This app has been removed, the operation is rejected.')
			);
			return;
		}
		return processAppOperation(app, OPERATION_TYPE.UPGRADE, request);
	},

	cancelInstallingApp: (app: AppStatusInfo, request: OperationRequest) =>
		processAppOperation(app, OPERATION_TYPE.CANCEL, request),

	removeApp: (app: AppStatusInfo, request: OperationRequest) =>
		processAppOperation(app, OPERATION_TYPE.REMOVE, request),

	resumeApp: async (app: AppStatusInfo, request: BaseOperationRequest) =>
		processAppOperation(app, OPERATION_TYPE.RESUME, request),

	stopApp: async (
		app: AppStatusInfo,
		request: BaseOperationRequest,
		q: QVueGlobals
	) => {
		const appStore = useAppStore();
		const userStore = useUserStore();
		const preCheckStore = usePreCheckStore();
		const aggregation = appStore.getAppAggregationInfo(
			request.app_name,
			request.source
		);
		const displayName =
			getI18nValue(
				aggregation?.app_simple_latest?.app_simple_info?.app_title,
				i18n.global.locale.value
			) ?? request.app_name;

		return csAppStop(
			q,
			displayName,
			!!aggregation && isCSV2(aggregation.app_full_info),
			preCheckStore.hasAdminPermissions(),
			userStore.accounts.length
		).then((result: { all: boolean }) => {
			processAppOperation(app, OPERATION_TYPE.STOP, request, result.all);
		});
	},

	openApp: (entrance: TerminusEntrance) =>
		handleWindowOperation({ id: entrance.id, url: entrance.url })
};
