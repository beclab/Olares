import axios from 'axios';
import { i18n } from 'src/boot/i18n';
import { useTokenStore } from 'src/stores/settings/token';
import {
	ComputeResourceBinding,
	ComputeResourceNode,
	ComputeSupportTypeResult,
	DeviceSupportType
} from 'src/constant/compute';

/**
 * Accelerator (compute-resources) API.
 *
 * Both builds now resolve the raw `{ code, message, data }` envelope for these
 * endpoints (the Settings interceptor has a `compute-resources` passthrough
 * that mirrors the Market interceptor), so the helpers below give every caller
 * the same flow: `unwrap` peels the envelope, `isErrorEnvelope` detects a
 * business error, and `unwrapOrThrow` turns such an error into a rejection.
 */
export function unwrap<T>(resp: any): T {
	if (
		resp &&
		!Array.isArray(resp) &&
		typeof resp === 'object' &&
		'code' in resp &&
		'data' in resp
	) {
		return resp.data as T;
	}
	return resp as T;
}

/** True when a resolved response is actually a business-error envelope. */
export function isErrorEnvelope(resp: any): boolean {
	return (
		resp &&
		!Array.isArray(resp) &&
		typeof resp === 'object' &&
		'code' in resp &&
		resp.code !== 0 &&
		resp.code !== 200
	);
}

/** Reject on a business-error envelope, otherwise return the inner payload. */
function unwrapOrThrow<T>(resp: any): T {
	if (isErrorEnvelope(resp)) {
		throw new Error(resp.message || i18n.global.t('Failed'));
	}
	return unwrap<T>(resp);
}

export async function getComputeResources(): Promise<ComputeResourceNode[]> {
	const tokenStore = useTokenStore();
	const data: any = await axios.get(`${tokenStore.url}/api/compute-resources`);
	const nodes = unwrapOrThrow<ComputeResourceNode[]>(data);
	return Array.isArray(nodes) ? nodes : [];
}

export async function getAppBindings(
	appName: string
): Promise<ComputeResourceBinding[]> {
	const tokenStore = useTokenStore();
	const data: any = await axios.get(
		`${tokenStore.url}/api/apps/${encodeURIComponent(
			appName
		)}/compute-resources/bindings`
	);
	const list = unwrapOrThrow<ComputeResourceBinding[]>(data);
	return Array.isArray(list) ? list : [];
}

export async function releaseAppBindings(appName: string): Promise<void> {
	const tokenStore = useTokenStore();
	const data: any = await axios.delete(
		`${tokenStore.url}/api/apps/${encodeURIComponent(
			appName
		)}/compute-resources/bindings`
	);
	unwrapOrThrow<unknown>(data);
}

export async function switchDeviceSupportType(
	nodeName: string,
	deviceId: string,
	supportType: DeviceSupportType | string
): Promise<any> {
	const tokenStore = useTokenStore();
	// Returns the raw envelope: the caller distinguishes the special
	// `bound-apps-stop-blocked` outcome from a genuine error before unwrapping.
	return axios.put(
		`${tokenStore.url}/api/compute-resources/nodes/${encodeURIComponent(
			nodeName
		)}/devices/${encodeURIComponent(deviceId)}/support-type`,
		{ supportType }
	);
}
