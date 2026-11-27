import { defineStore } from 'pinia';
import { QVueGlobals } from 'quasar';
import { APP_STATUS } from 'src/constant/constants';
import { getRawAppTitle } from 'src/constant/config';
import { useAppStore } from 'src/stores/market/appStore';
import { AppService } from 'src/services/market/appOperation/appService';
import {
	AppOverlayGatewayCommand,
	disableAppOverlayGateway,
	disableOverlayGateway,
	enableAppOverlayGateway,
	enableOverlayGateway,
	getOverlayGatewayStatus,
	OverlayGatewayStatusValue,
	OverlayGatewaySupportedApp,
	UnderlayNetwork
} from 'src/api/settings/overlayGateway';
import { useMeStore } from './me';
import { i18n } from 'src/boot/i18n';

export interface OverlayApp {
	name: string;
	title: string;
	icon: string;
	status: string;
	overlay: boolean;
	underlay_networks: UnderlayNetwork[];
}

export interface UnavailableReason {
	key: string;
	label: string;
}

export type OverlayGatewayStatusSyncTargets = {
	gateway: 'on' | 'off' | null;
	apps: Record<string, boolean>;
};

export type OverlayGatewayState = {
	isEnabled: boolean;
	isAvailable: boolean;
	gatewayStatus: OverlayGatewayStatusValue;
	errorMessage: string;
	unavailableReasons: UnavailableReason[];
	supportedApps: OverlayGatewaySupportedApp[];
	pendingAppNames: Set<string>;
	statusSyncTargets: OverlayGatewayStatusSyncTargets;
};

export type FetchStatusResult =
	| { ok: true }
	| { ok: false; transient: true }
	| { ok: false; transient: false; message: string };

/** Poll interval while a gateway/app status sync is in progress. */
export const OVERLAY_GATEWAY_POLL_MS = 8000;

// i18n-ignore-start: backend-provided reason strings used as lookup keys, not UI text
const DISABLE_REASON_KEYS: Record<string, string> = {
	'WSL is not supported': 'no_wsl',
	'MacOS is not supported': 'no_macos',
	'Ethernet connection is not active': 'no_ethernet'
};
// i18n-ignore-end

function isOverlayTransientError(error: unknown): boolean {
	const status = (error as { response?: { status?: number } })?.response
		?.status;
	return status === 530;
}

function hasUnderlayIp(app: OverlayGatewaySupportedApp | undefined): boolean {
	return !!app?.underlay_networks?.some((network) => !!network.ip);
}

function mapDisableReason(reason: string): UnavailableReason {
	return {
		key: DISABLE_REASON_KEYS[reason] ?? reason,
		label: reason
	};
}

export const useOverlayGatewayStore = defineStore('overlayGateway', {
	state: (): OverlayGatewayState => ({
		isEnabled: false,
		isAvailable: true,
		gatewayStatus: 'off',
		errorMessage: '',
		unavailableReasons: [],
		supportedApps: [],
		pendingAppNames: new Set<string>(),
		statusSyncTargets: {
			gateway: null,
			apps: {}
		}
	}),

	getters: {
		isGatewayBusy: (state): boolean => {
			return (
				state.gatewayStatus === 'activating' ||
				state.gatewayStatus === 'deactivating'
			);
		},

		isAppPending:
			(state) =>
			(name: string): boolean => {
				return state.pendingAppNames.has(name);
			},

		needsStatusPolling(state): boolean {
			if (state.pendingAppNames.size > 0) {
				return true;
			}
			if (
				state.gatewayStatus === 'activating' ||
				state.gatewayStatus === 'deactivating'
			) {
				return true;
			}
			if (state.statusSyncTargets.gateway) {
				return true;
			}
			return Object.keys(state.statusSyncTargets.apps).length > 0;
		},

		apps(state): OverlayApp[] {
			if (state.gatewayStatus !== 'on') {
				return [];
			}

			const appStore = useAppStore();
			const locale = i18n.global.locale.value;

			return state.supportedApps.map((supported) => {
				const found = appStore.findAppByName(supported.app_name);
				const aggregation = found
					? appStore.getAppAggregationInfo(found.appId, found.sourceId)
					: null;

				const statusInfo = aggregation?.app_status_latest?.status;
				const simpleInfo = aggregation?.app_simple_latest?.app_simple_info;

				const title =
					getRawAppTitle(locale, statusInfo, simpleInfo) || supported.app_name;
				const icon = simpleInfo?.app_icon ?? '';
				const status = statusInfo?.state ?? APP_STATUS.STOP.COMPLETED;

				return {
					name: supported.app_name,
					title,
					icon,
					status,
					overlay: supported.enabled,
					underlay_networks: supported.underlay_networks
						? supported.underlay_networks
						: []
				};
			});
		}
	},

	actions: {
		applyStatusData(data: {
			status: OverlayGatewayStatusValue;
			disable: boolean;
			disable_reason: string;
			supported_apps?: OverlayGatewaySupportedApp[];
			error_message?: string;
		}) {
			this.gatewayStatus = data.status;
			this.isEnabled = data.status === 'on';
			this.isAvailable = !data.disable;
			this.errorMessage = data.error_message ?? '';
			this.unavailableReasons = data.disable
				? [mapDisableReason(data.disable_reason)]
				: [];
			this.supportedApps =
				data.status === 'on' ? data.supported_apps ?? [] : [];
			this.evaluateStatusSyncTargets();
		},

		evaluateStatusSyncTargets() {
			if (
				this.statusSyncTargets.gateway &&
				this.gatewayStatus === this.statusSyncTargets.gateway
			) {
				this.statusSyncTargets.gateway = null;
			}

			for (const [appName, expectedEnabled] of Object.entries(
				this.statusSyncTargets.apps
			)) {
				const supported = this.supportedApps.find(
					(item) => item.app_name === appName
				);

				// Enabling is only "done" once the app reports enabled AND an
				// underlay IP is assigned (the app finished restarting). Disabling
				// is done as soon as it is no longer enabled.
				const reached = expectedEnabled
					? !!supported?.enabled && hasUnderlayIp(supported)
					: !supported || !supported.enabled;

				if (reached) {
					const next = { ...this.statusSyncTargets.apps };
					delete next[appName];
					this.statusSyncTargets.apps = next;
				}
			}
		},

		async fetchStatus(): Promise<FetchStatusResult> {
			try {
				const meStore = useMeStore();
				const data = await getOverlayGatewayStatus(meStore.user.name);
				this.applyStatusData(data);
				return { ok: true };
			} catch (error: unknown) {
				if (isOverlayTransientError(error)) {
					return { ok: false, transient: true };
				}
				const message =
					(error as { message?: string })?.message ??
					(error as { response?: { data?: { message?: string } } })?.response
						?.data?.message ??
					'Failed';
				return { ok: false, transient: false, message };
			}
		},

		async toggleGateway(enable: boolean) {
			const previousStatus = this.gatewayStatus;
			const targetStatus: 'on' | 'off' = enable ? 'on' : 'off';

			this.gatewayStatus = enable ? 'activating' : 'deactivating';
			this.statusSyncTargets.gateway = targetStatus;

			try {
				if (enable) {
					await enableOverlayGateway();
				} else {
					await disableOverlayGateway();
				}
			} catch (error: unknown) {
				if (isOverlayTransientError(error)) {
					return;
				}
				this.gatewayStatus = previousStatus;
				this.statusSyncTargets.gateway = null;
				const message =
					(error as { message?: string })?.message ??
					(error as { response?: { data?: { message?: string } } })?.response
						?.data?.message ??
					'Failed';
				throw error;
			}
		},

		async toggleAppOverlay(name: string, overlay: boolean, q?: QVueGlobals) {
			const supported = this.supportedApps.find(
				(item) => item.app_name === name
			);
			if (!supported) {
				return;
			}

			const meStore = useMeStore();
			const appStore = useAppStore();
			const found = appStore.findAppByName(name);

			const payload: AppOverlayGatewayCommand = {
				app_id: supported.app_id || found?.appId || name,
				user: supported.shared_app ? '' : meStore.user.name
			};

			this.pendingAppNames.add(name);
			this.statusSyncTargets.apps = {
				...this.statusSyncTargets.apps,
				[name]: overlay
			};

			try {
				if (overlay) {
					await enableAppOverlayGateway(payload);
				} else {
					await disableAppOverlayGateway(payload);
				}

				await this.fetchStatus();

				// Stopped apps stay stopped: overlay change is persisted but we don't
				// restart them. Running apps restart so the change takes effect.
				const isStopped = found?.status?.state === APP_STATUS.STOP.COMPLETED;

				if (q && found && !isStopped) {
					try {
						await AppService.restartApp(
							found.status,
							{ app_name: name, source: found.sourceId },
							q
						);
					} catch (error: unknown) {
						if (!isOverlayTransientError(error)) {
							throw error;
						}
					}
				}
			} catch (error: unknown) {
				if (isOverlayTransientError(error)) {
					return;
				}
				const next = { ...this.statusSyncTargets.apps };
				delete next[name];
				this.statusSyncTargets.apps = next;
				const message =
					(error as { message?: string })?.message ??
					(error as { response?: { data?: { message?: string } } })?.response
						?.data?.message ??
					'Failed';
				throw error;
			} finally {
				this.pendingAppNames.delete(name);
			}
		}
	}
});
