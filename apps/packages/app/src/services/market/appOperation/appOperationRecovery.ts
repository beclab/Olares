import SetEnvironmentDialog from 'src/pages/market/application/SetEnvironmentDialog.vue';
import SetCloneInfoDialog from 'src/pages/market/application/SetCloneInfoDialog.vue';
import {
	pickComputeBinding,
	pickGpuMode
} from 'src/pages/market/application/compute/computePickers';
import {
	ComputeAvailabilityRefetch,
	ComputeBindingItem,
	ComputeType
} from 'src/constant/compute';
import {
	isAppEnvError,
	isAppEntranceError,
	isComputeBindingPromptError,
	isComputeModeSelectError
} from 'src/constant/appOperationError';
import { CloneEntrance, UpdateEnvBody } from 'src/constant';
import {
	BaseOperationRequest,
	CloneRequest,
	OperationRequest,
	Resume2RestartRequest
} from 'src/api/market/private/operations';
import { AppStatusInfo } from 'src/constant/constants';
import { getRawAppTitle } from 'src/constant/config';
import { OPERATION_TYPE } from './operationTypes';
import { validateComputeBinding } from './appOperationApi';
import { useAppStore } from 'src/stores/market/appStore';
import { i18n } from 'src/boot/i18n';
import { QVueGlobals } from 'quasar';

export interface RecoveryCtx<TReq> {
	request: TReq;
	appTitle: string;
	appIcon: string;
	error: any;
	q: QVueGlobals;
	/** Re-fetch a fresh compute binding prompt (used by the binding picker). */
	refetch?: ComputeAvailabilityRefetch;
}

export interface Recovery<TReq> {
	match: (error: unknown) => boolean;
	apply: (ctx: RecoveryCtx<TReq>) => Promise<TReq | null>;
}

const mergeEnvs = (
	originalEnvs: UpdateEnvBody = [],
	newEnvs: UpdateEnvBody = []
): UpdateEnvBody => {
	const map = new Map(originalEnvs.map((item) => [item.envName, item]));
	newEnvs.forEach((item) => map.set(item.envName, item));
	return Array.from(map.values());
};

const mergeEntrances = (
	originalEntrances: CloneEntrance[] = [],
	newEntrances: CloneEntrance[] = []
): CloneEntrance[] => {
	const map = new Map(originalEntrances.map((item) => [item.name, item]));
	newEntrances.forEach((item) => map.set(item.name, item));
	return Array.from(map.values());
};

const resolveAppTitleFor = (appName: string, sourceId: string): string => {
	// Use the aggregation so clone apps resolve to their own `status.title`
	// instead of the catalog title of the source app — same rule as appOperationPrompt.
	const aggregation = useAppStore().getAppAggregationInfo(appName, sourceId);
	return (
		getRawAppTitle(
			i18n.global.locale.value,
			aggregation?.app_status_latest?.status,
			aggregation?.app_simple_latest?.app_simple_info
		) || appName
	);
};

const resolveAppIconFor = (appName: string, sourceId: string): string => {
	const simpleLatest = useAppStore().getAppSimpleInfo(appName, sourceId);
	return simpleLatest?.app_simple_info.app_icon ?? '';
};

export const appEnvRecovery = <
	T extends BaseOperationRequest & { envs?: UpdateEnvBody; title?: string }
>(): Recovery<T> => ({
	match: isAppEnvError,
	apply({ request, appTitle, error, q }) {
		return new Promise<T | null>((resolve) => {
			q.dialog({
				component: SetEnvironmentDialog,
				componentProps: { data: error.data, appTitle }
			})
				.onOk((data: { envs?: UpdateEnvBody } | undefined) => {
					if (data && data.envs) {
						resolve({
							...request,
							envs: mergeEnvs(request.envs, data.envs)
						});
					} else {
						resolve(null);
					}
				})
				.onCancel(() => resolve(null));
		});
	}
});

export const appEntranceRecovery = <
	T extends BaseOperationRequest & {
		title?: string;
		entrances?: CloneEntrance[];
		templateClone?: boolean;
	}
>(): Recovery<T> => ({
	match: isAppEntranceError,
	apply({ request, error, q }) {
		return new Promise<T | null>((resolve) => {
			q.dialog({
				component: SetCloneInfoDialog,
				componentProps: {
					data: error.data,
					templateClone: request.templateClone
				}
			})
				.onOk(
					(data: { title: string; entrances: CloneEntrance[] } | undefined) => {
						if (data) {
							resolve({
								...request,
								title: data.title,
								entrances: mergeEntrances(request.entrances, data.entrances)
							});
						} else {
							resolve(null);
						}
					}
				)
				.onCancel(() => resolve(null));
		});
	}
});

export const computeModeRecovery = <
	T extends BaseOperationRequest & { selectedGpuType?: ComputeType }
>(): Recovery<T> => ({
	match: isComputeModeSelectError,
	async apply({ request, appTitle, error, q }) {
		try {
			const selectedGpuType = await pickGpuMode(q, appTitle, error.data);
			return { ...request, selectedGpuType };
		} catch (cancel) {
			console.log('compute mode picker canceled:', cancel);
			return null;
		}
	}
});

export const computeBindingRecovery = <
	T extends BaseOperationRequest & { computeBinding?: ComputeBindingItem[] }
>(): Recovery<T> => ({
	match: isComputeBindingPromptError,
	async apply({ request, appTitle, appIcon, error, q, refetch }) {
		try {
			const computeBinding = await pickComputeBinding(
				q,
				appTitle,
				error.data,
				error.validation,
				appIcon,
				// Bring back the parameters passed last time so the dialog can
				// restore the prior selection and explain why it failed.
				request.computeBinding,
				refetch
			);
			return { ...request, computeBinding };
		} catch (cancel) {
			console.log('compute binding picker canceled:', cancel);
			return null;
		}
	}
});

export const INSTALL_RECOVERIES: Recovery<OperationRequest>[] = [
	computeModeRecovery<OperationRequest>(),
	appEnvRecovery<OperationRequest>()
];

export const CLONE_RECOVERIES: Recovery<CloneRequest>[] = [
	computeModeRecovery<CloneRequest>(),
	appEnvRecovery<CloneRequest>(),
	appEntranceRecovery<CloneRequest>()
];

export const RESUME2RESTART_RECOVERIES: Recovery<Resume2RestartRequest>[] = [
	computeBindingRecovery<Resume2RestartRequest>()
];

// `process` is injected to avoid an import cycle with appService.
export async function runWithRecovery<
	TReq extends BaseOperationRequest & { title?: string }
>(
	op: OPERATION_TYPE,
	app: AppStatusInfo,
	request: TReq,
	recoveries: Recovery<TReq>[],
	q: QVueGlobals,
	process: (
		app: AppStatusInfo,
		op: OPERATION_TYPE,
		req: TReq
	) => Promise<unknown>
): Promise<void> {
	try {
		await process(app, op, request);
	} catch (error) {
		const recovery = recoveries.find((r) => r.match(error));
		if (!recovery) {
			throw error;
		}
		// Refresh uses the dedicated validate-only endpoint (no submit) so the
		// dialog gets fresh, backend-authoritative availability + validation after
		// the user frees resources. Only resume/restart expose this check.
		const supportsValidate =
			op === OPERATION_TYPE.RESUME || op === OPERATION_TYPE.RESTART;
		const refetch: ComputeAvailabilityRefetch | undefined = supportsValidate
			? () =>
					validateComputeBinding({
						app_name: request.app_name,
						source: request.source
					})
			: undefined;
		const next = await recovery.apply({
			request,
			appTitle:
				request.title ?? resolveAppTitleFor(request.app_name, request.source),
			appIcon: resolveAppIconFor(request.app_name, request.source),
			error,
			q,
			refetch
		});
		if (!next) return;
		await runWithRecovery(op, app, next, recoveries, q, process);
	}
}
