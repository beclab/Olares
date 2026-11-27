import { defineStore } from 'pinia';
import { i18n } from 'src/boot/i18n';
import { APP_STATUS } from 'src/constant/constants';
import { getRawAppTitle } from 'src/constant/config';
import { useAppStore } from 'src/stores/market/appStore';
import { notifyFailed, notifySuccess } from 'src/utils/settings/btNotify';
import {
	getComputeResources,
	isErrorEnvelope,
	releaseAppBindings,
	switchDeviceSupportType,
	unwrap
} from 'src/api/settings/accelerator';
import {
	AppComputeScope,
	ComputeResourceDevice,
	ComputeResourceNode,
	ComputeSupportTypeResult,
	DeviceSupportType
} from 'src/constant/compute';

export interface AcceleratorAppView {
	appName: string;
	title: string;
	icon: string;
	state: string;
	memory: number;
	multiCard: boolean;
	mode: string;
	/** Capability from backend spec; undefined => unknown => no scope badge. */
	scope?: AppComputeScope;
}

export type AcceleratorState = {
	nodes: ComputeResourceNode[];
	loading: boolean;
};

export const useAcceleratorStore = defineStore('accelerator', {
	state: (): AcceleratorState => ({
		nodes: [],
		loading: false
	}),

	getters: {
		getNode:
			(state) =>
			(nodeName: string): ComputeResourceNode | undefined =>
				state.nodes.find((n) => n.nodeName === nodeName),

		/** Resolve an app's title / icon / state by name via the market appStore. */
		resolveApp: () => (appName: string) => {
			const appStore = useAppStore();
			const locale = i18n.global.locale.value;
			const found = appStore.findAppByName(appName);
			const aggregation = found
				? appStore.getAppAggregationInfo(found.appId, found.sourceId)
				: null;
			const statusInfo = aggregation?.app_status_latest?.status;
			const simpleInfo = aggregation?.app_simple_latest?.app_simple_info;
			return {
				title: getRawAppTitle(locale, statusInfo, simpleInfo) || appName,
				icon: simpleInfo?.app_icon ?? '',
				state: statusInfo?.state ?? APP_STATUS.STOP.COMPLETED
			};
		},

		/** Apps bound to a device, enriched for display. */
		deviceApps() {
			return (device: ComputeResourceDevice): AcceleratorAppView[] => {
				return (device.bindings ?? []).map((binding) => {
					const info = this.resolveApp(binding.appName);
					const spec = binding.spec;
					return {
						appName: binding.appName,
						title: info.title,
						icon: info.icon,
						state: info.state,
						memory: binding.memory ?? 0,
						multiCard: spec?.supportMultiCards ?? false,
						mode: String(binding.mode ?? device.mode ?? ''),
						scope: spec
							? {
									supportMultiCards: !!spec.supportMultiCards,
									supportMultiNodes: !!spec.supportMultiNodes
							  }
							: undefined
					};
				});
			};
		}
	},

	actions: {
		async fetch() {
			this.loading = true;
			try {
				const nodes = await getComputeResources();
				this.nodes = Array.isArray(nodes) ? nodes : [];
			} catch (error: any) {
				notifyFailed(error?.message ?? 'Failed');
			} finally {
				this.loading = false;
			}
		},

		/**
		 * Switch a device support type. Returns the result so the caller can react
		 * to a `bound-apps-stop-blocked` outcome (HTTP 200 + business code 4603).
		 */
		async switchSupportType(
			nodeName: string,
			deviceId: string,
			supportType: DeviceSupportType | string
		): Promise<ComputeSupportTypeResult | undefined> {
			try {
				const resp: any = await switchDeviceSupportType(
					nodeName,
					deviceId,
					supportType
				);
				const outer = unwrap<any>(resp);
				// Blocked outcome arrives as code:0 with a discriminated envelope:
				//   data: { type: 'computeDeviceSwitchBlocked', Data: { status, blockedApps, ... } }
				const result = (
					outer?.type === 'computeDeviceSwitchBlocked' && outer?.Data
						? outer.Data
						: outer
				) as ComputeSupportTypeResult;
				if (result?.status === 'bound-apps-stop-blocked') {
					return result;
				}
				if (isErrorEnvelope(resp)) {
					throw new Error(resp.message || 'Failed');
				}
				notifySuccess(i18n.global.t('successful'));
				await this.fetch();
				return result;
			} catch (error: any) {
				notifyFailed(error?.message ?? 'Failed');
				throw error;
			}
		},

		async releaseApp(appName: string) {
			try {
				await releaseAppBindings(appName);
				notifySuccess(i18n.global.t('successful'));
				await this.fetch();
			} catch (error: any) {
				notifyFailed(error?.message ?? 'Failed');
				throw error;
			}
		}
	}
});
