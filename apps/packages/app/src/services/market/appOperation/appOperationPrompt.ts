import UninstallAppDialog from 'src/components/appcard/UninstallAppDialog.vue';
import StopDialog from 'src/components/appcard/StopDialog.vue';
import { AppStatusInfo, isCSV2 } from 'src/constant/constants';
import { getRawAppTitle } from 'src/constant/config';
import { useUsersStore } from 'src/stores/settings/users';
import { useAppStore } from 'src/stores/market/appStore';
import { useMeStore } from 'src/stores/settings/me';
import { QVueGlobals } from 'quasar';
import { i18n } from 'src/boot/i18n';
import { Component } from 'vue';
import {
	StopRequest,
	UninstallRequest
} from 'src/api/market/private/operations';

export interface PrePromptCtx<TReq> {
	request: TReq;
	app: AppStatusInfo;
	q: QVueGlobals;
}

export interface PrePrompt<TReq, TOut = TReq> {
	apply: (ctx: PrePromptCtx<TReq>) => Promise<TOut | null>;
}

const awaitDialog = <T>(
	q: QVueGlobals,
	component: Component,
	componentProps: Record<string, unknown>
): Promise<T | null> =>
	new Promise((resolve) => {
		q.dialog({ component, componentProps })
			.onOk((result: T) => resolve(result))
			.onCancel(() => resolve(null));
	});

const resolveAppContext = (
	appName: string,
	sourceId: string
): { displayName: string; isCsV2: boolean } => {
	const aggregation = useAppStore().getAppAggregationInfo(appName, sourceId);
	const displayName =
		getRawAppTitle(
			i18n.global.locale.value,
			aggregation?.app_status_latest?.status,
			aggregation?.app_simple_latest?.app_simple_info
		) || appName;
	return {
		displayName,
		isCsV2: !!aggregation && isCSV2(aggregation.app_simple_latest)
	};
};

export const uninstallChoicePrompt = (): PrePrompt<UninstallRequest> => ({
	async apply({ request, q }) {
		const { displayName, isCsV2 } = resolveAppContext(
			request.app_name,
			request.source
		);
		const isAdmin = useMeStore().isAdmin;
		const users = useUsersStore().accounts.length;

		const result = await awaitDialog<{ all: boolean; clearData: boolean }>(
			q,
			UninstallAppDialog,
			{
				appName: displayName,
				showCSAppCheckbox: isCsV2 && isAdmin && users > 1,
				all: isCsV2 && isAdmin,
				allDisabled: isCsV2,
				clearData: true
			}
		);
		if (!result) return null;
		return { ...request, all: result.all, deleteData: result.clearData };
	}
});

export const stopChoicePrompt = (): PrePrompt<StopRequest> => ({
	async apply({ request, q }) {
		const { displayName, isCsV2 } = resolveAppContext(
			request.app_name,
			request.source
		);
		const isAdmin = useMeStore().isAdmin;
		const users = useUsersStore().accounts.length;

		// CS v2 admin + multi-user: ask whether to stop the shared server.
		// Otherwise: non-CS/non-admin → all=false; single user → all=true.
		if (!isCsV2 || !isAdmin) return { ...request, all: false };
		if (users <= 1) return { ...request, all: true };

		const result = await awaitDialog<{ all: boolean }>(q, StopDialog, {
			appName: displayName,
			all: true,
			allDisabled: true
		});
		if (!result) return null;
		return { ...request, all: result.all };
	}
});
