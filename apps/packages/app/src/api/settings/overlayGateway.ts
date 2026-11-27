import axios from 'axios';
import { useTokenStore } from 'src/stores/settings/token';

export type OverlayGatewayStatusValue =
	| 'on'
	| 'off'
	| 'activating'
	| 'deactivating';

export type UnderlayPortProtocol = 'tcp' | 'udp' | '';

export interface UnderlayPort {
	title: string;
	port: number;
	workload: string;
	description?: string;
	protocol?: string;
}

export interface UnderlayNetwork {
	ip: string;
	ports: UnderlayPort[];
}

export interface OverlayGatewaySupportedApp {
	app_name: string;
	app_id: string;
	enabled: boolean;
	shared_app: boolean;
	underlay_networks: UnderlayNetwork[];
}

export function normalizeUnderlayPortProtocol(
	value: string | undefined | null
): UnderlayPortProtocol {
	if (value == null) return '';
	const v = String(value).trim().toLowerCase();
	return v === 'tcp' || v === 'udp' ? v : '';
}

export interface OverlayGatewayStatus {
	status: OverlayGatewayStatusValue;
	disable: boolean;
	disable_reason: string;
	supported_apps: OverlayGatewaySupportedApp[];
	error_message: string;
}

export interface AppOverlayGatewayCommand {
	app_id: string;
	user: string;
}

export async function getOverlayGatewayStatus(
	user: string
): Promise<OverlayGatewayStatus> {
	const tokenStore = useTokenStore();
	return axios.get(
		`${tokenStore.url}/api/system/overlay-gateway-status/${encodeURIComponent(
			user
		)}`
	);
}

export async function enableOverlayGateway(): Promise<any> {
	const tokenStore = useTokenStore();
	return axios.post(
		`${tokenStore.url}/api/command/enable-overlay-gateway`,
		null
	);
}

export async function disableOverlayGateway(): Promise<any> {
	const tokenStore = useTokenStore();
	return axios.post(
		`${tokenStore.url}/api/command/disable-overlay-gateway`,
		null
	);
}

export async function enableAppOverlayGateway(
	body: AppOverlayGatewayCommand
): Promise<any> {
	const tokenStore = useTokenStore();
	return axios.post(
		`${tokenStore.url}/api/command/enable-app-overlay-gateway`,
		body
	);
}

export async function disableAppOverlayGateway(
	body: AppOverlayGatewayCommand
): Promise<any> {
	const tokenStore = useTokenStore();
	return axios.post(
		`${tokenStore.url}/api/command/disable-app-overlay-gateway`,
		body
	);
}
