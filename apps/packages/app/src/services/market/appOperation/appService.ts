import DependencyDialog from 'src/components/appintro/DependencyDialog.vue';
import { gateAppOp } from './appOperationPolicy';
import { OPERATION_TYPE } from './operationTypes';
import { Action, Category, TerminusEntrance } from '@bytetrade/core';
import { openApplication } from 'src/api/market/private/server';
import { usePreCheckStore } from 'src/stores/market/preCheck';
import { useAppStore } from 'src/stores/market/appStore';
import { suspendApp } from 'src/constant/config';
import { bus, BUS_EVENT } from 'src/utils/bus';
import { useMeStore } from 'src/stores/settings/me';
import { i18n } from 'src/boot/i18n';
import { QVueGlobals } from 'quasar';
import { AppStatusInfo } from 'src/constant/constants';
import {
	BaseOperationRequest,
	CloneRequest,
	OperationRequest,
	Resume2RestartRequest,
	StopRequest,
	UninstallRequest
} from 'src/api/market/private/operations';
import { processAppOperation } from './appOperationApi';
import {
	CLONE_RECOVERIES,
	INSTALL_RECOVERIES,
	RESUME2RESTART_RECOVERIES,
	runWithRecovery
} from './appOperationRecovery';
import { stopChoicePrompt, uninstallChoicePrompt } from './appOperationPrompt';

export { OPERATION_TYPE };
export { processAppOperation };

const maybeShowDependencyDialog = (
	appName: string,
	sourceId: string,
	q: QVueGlobals
): void => {
	const fullLatest = useAppStore().getAppFullInfo(appName, sourceId);
	if (
		fullLatest &&
		usePreCheckStore().hasUninstallDependencies(fullLatest.app_info.app_entry)
	) {
		q.dialog({
			component: DependencyDialog,
			componentProps: {
				appName,
				sourceId
			}
		});
	}
};

export async function handleWindowOperation(entrance: {
	id: string;
	url: string;
}) {
	if (window.top === window) {
		// On a local network, rebuild the host from the app id so it follows the
		// current .local domain.
		const meStore = useMeStore();
		let url = entrance.url;
		if (meStore.islocal) {
			url = meStore.getAppLocalUrl(entrance.id);
		}
		window.open(`//${url}`, '_blank');
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

const checkGate = (op: OPERATION_TYPE, req: BaseOperationRequest) =>
	gateAppOp({ op, appName: req.app_name, sourceId: req.source });

export const AppService = {
	installApp: async (
		app: AppStatusInfo,
		request: OperationRequest,
		q: QVueGlobals
	) => {
		if (!(await checkGate(OPERATION_TYPE.INSTALL, request))) return;
		await runWithRecovery(
			OPERATION_TYPE.INSTALL,
			app,
			request,
			INSTALL_RECOVERIES,
			q,
			processAppOperation
		);
		maybeShowDependencyDialog(request.app_name, request.source, q);
	},

	cloneApp: async (
		app: AppStatusInfo,
		request: CloneRequest,
		q: QVueGlobals
	) => {
		await runWithRecovery(
			OPERATION_TYPE.CLONE,
			app,
			request,
			CLONE_RECOVERIES,
			q,
			processAppOperation
		);
		maybeShowDependencyDialog(request.app_name, request.source, q);
	},

	uninstallApp: async (
		app: AppStatusInfo,
		request: UninstallRequest,
		q: QVueGlobals
	): Promise<boolean> => {
		if (!(await checkGate(OPERATION_TYPE.UNINSTALL, request))) return false;
		const finalReq = await uninstallChoicePrompt().apply({ request, app, q });
		if (!finalReq) return false;
		try {
			await processAppOperation(app, OPERATION_TYPE.UNINSTALL, finalReq);
			return true;
		} catch {
			return false;
		}
	},

	upgradeApp: async (app: AppStatusInfo, request: OperationRequest) => {
		if (!(await checkGate(OPERATION_TYPE.UPGRADE, request))) return;
		const simpleLatest = useAppStore().getAppSimpleInfo(
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

	resumeApp: async (
		app: AppStatusInfo,
		request: Resume2RestartRequest,
		q?: QVueGlobals
	) => {
		if (!q) {
			console.error(
				'AppService.resumeApp: QVueGlobals required to open compute binding picker'
			);
			return;
		}
		if (!(await checkGate(OPERATION_TYPE.RESUME, request))) return;
		await runWithRecovery(
			OPERATION_TYPE.RESUME,
			app,
			request,
			RESUME2RESTART_RECOVERIES,
			q,
			processAppOperation
		);
	},

	stopApp: async (app: AppStatusInfo, request: StopRequest, q: QVueGlobals) => {
		if (!(await checkGate(OPERATION_TYPE.STOP, request))) return;
		const finalReq = await stopChoicePrompt().apply({ request, app, q });
		if (!finalReq) return;
		await processAppOperation(app, OPERATION_TYPE.STOP, finalReq);
	},

	restartApp: async (
		app: AppStatusInfo,
		request: Resume2RestartRequest,
		q?: QVueGlobals
	) => {
		if (!q) {
			console.error(
				'AppService.restartApp: QVueGlobals required to open compute binding picker'
			);
			return;
		}
		if (!(await checkGate(OPERATION_TYPE.RESTART, request))) return;
		await runWithRecovery(
			OPERATION_TYPE.RESTART,
			app,
			request,
			RESUME2RESTART_RECOVERIES,
			q,
			processAppOperation
		);
	},

	openApp: (entrance: TerminusEntrance) =>
		handleWindowOperation({ id: entrance.id, url: entrance.url })
};
