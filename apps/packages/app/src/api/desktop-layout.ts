import axios from 'axios';
import { useTokenStore } from 'src/stores/desktop/token';
import {
	DockLayout,
	LaunchLayout,
	LayoutEnvelope
} from 'src/types/desktop-layout';

function baseUrl(): string {
	const tokenStore = useTokenStore();
	return tokenStore.url || '';
}

export interface LaunchPutResult {
	current: LaunchLayout;
}

export interface DockPutResult {
	current: DockLayout;
}

export interface OrphanCleanupResult {
	changedSurfaces: string[];
}

export async function getLaunchLayout(): Promise<LayoutEnvelope<LaunchLayout>> {
	return await axios.get(`${baseUrl()}/api/desktop/launch-layout`);
}

export interface PutLaunchPayload {
	expectedVersion: number;
	layout: Omit<LaunchLayout, 'version' | 'updatedAt' | 'updatedBy'>;
	deviceId: string;
	clientOpId: string;
}

export interface PutDockPayload {
	layout: Omit<DockLayout, 'version' | 'updatedAt' | 'updatedBy'>;
	deviceId: string;
	clientOpId: string;
}

// Returns the persisted layout. On version conflict the caller receives a
// 409 with body { code: 'VERSION_MISMATCH', current: LaunchLayout } via
// axios error.response.data; we let the caller catch and decide.
export async function putLaunchLayout(
	body: PutLaunchPayload
): Promise<LaunchPutResult> {
	return await axios.put(`${baseUrl()}/api/desktop/launch-layout`, body);
}

export async function getDockLayout(): Promise<LayoutEnvelope<DockLayout>> {
	return await axios.get(`${baseUrl()}/api/desktop/dock-layout`);
}

export async function putDockLayout(
	body: PutDockPayload
): Promise<DockPutResult> {
	return await axios.put(`${baseUrl()}/api/desktop/dock-layout`, body);
}

export async function orphanCleanup(
	keys: string[]
): Promise<OrphanCleanupResult> {
	return await axios.post(`${baseUrl()}/api/desktop/layout/orphan-cleanup`, {
		keys
	});
}

export async function resetLayout(): Promise<OrphanCleanupResult> {
	return await axios.post(`${baseUrl()}/api/desktop/layout/reset`, {});
}
