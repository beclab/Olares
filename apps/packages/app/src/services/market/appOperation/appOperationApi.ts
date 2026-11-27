import { gateAppOp } from './appOperationPolicy';
import { OPERATION_TYPE } from './operationTypes';
import { useAppStore } from 'src/stores/market/appStore';
import { bus, BUS_EVENT } from 'src/utils/bus';
import { i18n } from 'src/boot/i18n';
import {
	APP_STATUS,
	AppStatusInfo,
	STATUS_OPERATE_TYPE
} from 'src/constant/constants';
import {
	APP_OPERATION_INPUT_TYPE,
	AppUserInputError,
	normalizeComputeBindingPromptData
} from 'src/constant/appOperationError';
import {
	CloneRequest,
	OperationRequest,
	Resume2RestartRequest,
	StopRequest,
	UninstallRequest,
	cancelInstalling,
	cloneApp,
	installApp,
	removeLocalPackage,
	restartApp,
	resumeApp,
	stopApp,
	uninstallApp,
	upgradeApp,
	validateResumeCompute
} from 'src/api/market/private/operations';
import { ComputeAvailabilitySnapshot } from 'src/constant/compute';

// i18n-ignore-start -- logMessage values are console.log debug strings, not user-facing
const operationConfig = {
	[OPERATION_TYPE.INSTALL]: {
		action: (request: OperationRequest) =>
			installApp({ ...request, sync: true }),
		getStatus: () => APP_STATUS.PENDING.DEFAULT,
		logMessage: 'install app start'
	},
	[OPERATION_TYPE.CLONE]: {
		action: (request: CloneRequest) => cloneApp({ ...request, sync: true }),
		// Clone creates a NEW instance and never mutates the body (app_name).
		// An empty optimistic state tells processAppOperation to skip the
		// optimistic write/revert, so a template body never gets a transient
		// RUNNING status leaking into the installed list.
		getStatus: () => '',
		logMessage: 'clone app start'
	},
	[OPERATION_TYPE.UNINSTALL]: {
		action: (request: UninstallRequest) =>
			uninstallApp({ ...request, sync: true }),
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
		action: (request: Resume2RestartRequest) => resumeApp(request),
		getStatus: () => APP_STATUS.RESUME.DEFAULT,
		logMessage: 'app resuming'
	},
	[OPERATION_TYPE.STOP]: {
		action: (request: StopRequest) => stopApp(request),
		getStatus: () => APP_STATUS.STOP.DEFAULT,
		logMessage: 'app stopping'
	},
	[OPERATION_TYPE.RESTART]: {
		action: (request: Resume2RestartRequest) => restartApp(request),
		getStatus: () => APP_STATUS.STOP.DEFAULT,
		logMessage: 'app restarting'
	}
};
// i18n-ignore-end

// axios non-2xx: e.response.data.data.backend_response
interface BackendResponseEnvelope {
	code: number;
	message?: string;
	data?: any;
	validation?: any;
	[key: string]: any;
}

interface BackendCodeHandler {
	code: number;
	operations: OPERATION_TYPE[];
	toError: (envelope: BackendResponseEnvelope) => AppUserInputError;
}

const BACKEND_CODE_HANDLERS: BackendCodeHandler[] = [
	{
		code: 422,
		operations: [
			OPERATION_TYPE.INSTALL,
			OPERATION_TYPE.CLONE,
			OPERATION_TYPE.RESUME,
			OPERATION_TYPE.RESTART
		],
		toError: (env) => {
			const type = env.data?.type;
			const rawData = env.data?.Data;

			if (
				type === APP_OPERATION_INPUT_TYPE.COMPUTE_BINDING_REQUIRED ||
				type === APP_OPERATION_INPUT_TYPE.COMPUTE_BINDING_UNAVAILABLE
			) {
				const { availability, validation } =
					normalizeComputeBindingPromptData(rawData);
				let mergedValidation = validation ?? env.validation;
				if (
					type === APP_OPERATION_INPUT_TYPE.COMPUTE_BINDING_UNAVAILABLE &&
					!mergedValidation
				) {
					mergedValidation = {
						ok: false,
						code: 'binding_unavailable',
						reason: ''
					};
				}
				return {
					type,
					data: availability,
					...(mergedValidation ? { validation: mergedValidation } : {})
				};
			}

			return {
				type,
				data: rawData,
				...(env.validation ? { validation: env.validation } : {})
			};
		}
	}
];

const matchBackendCodeHandler = (
	operationType: OPERATION_TYPE,
	envelope: BackendResponseEnvelope | undefined
): BackendCodeHandler | undefined => {
	if (!envelope || typeof envelope.code !== 'number') return undefined;
	return BACKEND_CODE_HANDLERS.find(
		(h) => h.code === envelope.code && h.operations.includes(operationType)
	);
};

export async function processAppOperation(
	app: AppStatusInfo,
	operationType: OPERATION_TYPE,
	...args: any[]
): Promise<any> {
	const config = operationConfig[operationType];
	console.log(config.logMessage);

	const allowed = await gateAppOp({
		op: operationType,
		appName: args[0]?.app_name,
		sourceId: args[0]?.source
	});
	if (!allowed) {
		return Promise.reject(new Error('app-policy: operation blocked'));
	}

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
	const optimisticState = config.getStatus();
	if (optimisticState) {
		appStore.updateAppStatus(
			args[0].app_name,
			args[0].source,
			optimisticState,
			opType
		);
	}

	return new Promise((resolve, reject) => {
		config
			// @ts-ignore-next-line: A spread argument must either have a tuple type or be passed to a rest parameter
			.action(...args)
			.then((response) => {
				console.log(response);
				resolve(response);
			})
			.catch((e) => {
				if (optimisticState) {
					appStore.updateAppStatus(args[0].app_name, args[0].source, oldState);
				}
				console.error(e);

				const envelope: BackendResponseEnvelope | undefined =
					e?.response?.data?.data?.backend_response;

				const handler = matchBackendCodeHandler(operationType, envelope);
				if (handler && envelope) {
					reject(handler.toError(envelope));
					return;
				}

				bus.emit(
					BUS_EVENT.APP_BACKEND_ERROR,
					envelope?.message ||
						e?.response?.data?.message ||
						i18n.global.t('error.operation_preform_failure')
				);

				reject(e);
			});
	});
}

/**
 * Pull a {availability, validation} snapshot out of an arbitrary response or
 * 422 envelope. Accepts the resume-consistent shape ({ type, Data, validation })
 * as well as a bare availability response, so it works whatever the (pending)
 * validate endpoint ends up returning.
 */
const extractComputeSnapshot = (
	raw: unknown
): ComputeAvailabilitySnapshot | null => {
	if (!raw || typeof raw !== 'object') return null;
	const obj = raw as Record<string, any>;
	const payload = obj.Data ?? obj.data?.Data ?? obj.data ?? obj;
	const { availability, validation } =
		normalizeComputeBindingPromptData(payload);
	if (!availability || !Array.isArray(availability.nodes)) return null;
	return { availability, validation: validation ?? availability.validation };
};

/**
 * Dedicated validate-only check for resume's accelerator selection — does NOT
 * resume the app. Used by the binding dialog's refresh so the usage bars and
 * warnings reflect freshly-freed resources. Stays consistent with resume: if
 * the endpoint still reports the prompt via a 422 envelope, it's parsed the
 * same way instead of surfacing an error toast.
 */
export async function validateComputeBinding(
	request: Resume2RestartRequest
): Promise<ComputeAvailabilitySnapshot | null> {
	try {
		const res: any = await validateResumeCompute(request);
		return extractComputeSnapshot(res?.data ?? res);
	} catch (e: any) {
		const env = e?.response?.data?.data?.backend_response;
		const snap = env ? extractComputeSnapshot(env.data ?? env) : null;
		// Align with the resume/restart 422 handler: the envelope may carry the
		// validation at its top level rather than nested under `data`.
		if (snap) return { ...snap, validation: snap.validation ?? env.validation };
		console.error('validateComputeBinding failed:', e);
		return null;
	}
}
