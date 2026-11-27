import { CPUResource, MemoryResource } from '@icebergtsn/k8s-resources';
import {
	AppAcceleratorRequirement,
	AppAcceleratorStatus,
	AppEntry,
	Resource,
	TerminusResource,
	UserResource
} from 'src/constant/constants';
import { COMPUTE_TYPE, normalizeComputeMode } from 'src/constant/compute';
import type { ComputeResourceNode } from 'src/constant/compute';
import { intersection } from 'src/utils/utils';

export type AppResourceMode = {
	mode: string;
	requiredCpu?: string;
	requiredMemory?: string;
	requiredDisk?: string;
	requiredGPUMemory?: string;
	isLegacy: boolean;
};

/**
 * Which cluster resource a mode falls short on. The app detail card only checks
 * cluster-side capacity (display hint only); user quota is not considered here.
 */
export type ResourceShortfall =
	| 'cluster-cpu'
	| 'cluster-memory'
	| 'cluster-disk'
	| 'cluster-gpu';

export type EvaluatedResourceMode = AppResourceMode & {
	status: AppAcceleratorStatus;
	shortfalls: ResourceShortfall[];
};

export type NormalizedSystemResources = {
	user: UserResource | null;
	cluster: TerminusResource | null;
	/**
	 * Per-node compute capabilities from /compute-resources. When present and
	 * non-empty, a GPU mode is only installable if some node exposes it; when
	 * absent, the evaluator falls back to the aggregate-GPU heuristic.
	 */
	computeNodes?: ComputeResourceNode[] | null;
};

/**
 * Set of compute modes exposed by any node, drawn from each node's `gpuTypes`
 * list and its devices' `mode`, normalized so spelling variants match.
 */
export function getSupportedModes(
	computeNodes?: ComputeResourceNode[] | null
): Set<string> {
	const supported = new Set<string>();
	const add = (mode: string | undefined | null) => {
		const key = normalizeComputeMode(mode);
		if (key) supported.add(key);
	};
	(computeNodes ?? []).forEach((node) => {
		(node.gpuTypes ?? []).forEach((mode) => add(mode as string));
		(node.devices ?? []).forEach((device) => add(device.mode as string));
	});
	return supported;
}

export function parseCpuResource(
	value: string | number | undefined | null
): CPUResource | null {
	if (value === undefined || value === null) return null;

	try {
		if (typeof value === 'number') {
			if (!Number.isFinite(value) || value < 0) return null;
			return CPUResource.fromMillicores(Math.round(value * 1000));
		}

		const str = String(value).trim();
		if (!str || str.startsWith('-')) return null;

		if (/^\d+(\.\d+)?$/.test(str)) {
			return CPUResource.fromMillicores(Math.round(parseFloat(str) * 1000));
		}

		return new CPUResource(str);
	} catch (e) {
		console.error('parseCpuResource error', e);
		return null;
	}
}

export function parseMemoryResource(
	value: string | number | undefined | null
): MemoryResource | null {
	if (value === undefined || value === null) return null;

	try {
		if (typeof value === 'number') {
			if (!Number.isFinite(value) || value < 0) return null;
			return MemoryResource.fromBytes(Math.round(value));
		}

		const str = String(value).trim();
		if (!str || str.startsWith('-')) return null;

		if (/^\d+(\.\d+)?$/.test(str)) {
			return MemoryResource.fromBytes(Math.round(Number(str)));
		}

		return new MemoryResource(str);
	} catch (e) {
		console.error('parseMemoryResource error', e);
		return null;
	}
}

function splitK8sResource(str: string): { value: string; unit: string } {
	const trimmed = str.trim();
	if (!trimmed) {
		return { value: '-', unit: '-' };
	}

	const match = trimmed.match(/^(\d+(?:\.\d+)?)([a-zA-Z]+)?$/);
	if (!match) {
		return { value: trimmed, unit: '' };
	}

	let value = match[1];
	const unit = match[2] || '';

	const num = Number(value);
	if (Number.isFinite(num)) {
		value = Number(num.toFixed(2)).toString();
	}

	return { value, unit };
}

function formatNumberForDisplay(value: number): string {
	if (!Number.isFinite(value)) return '-';
	return Number(value.toFixed(value >= 10 ? 0 : 1)).toString();
}

function formatBytesForDisplay(bytes: number): { value: string; unit: string } {
	if (!Number.isFinite(bytes) || bytes < 0) return { value: '-', unit: '-' };

	const Gi = 1024 ** 3;
	const Mi = 1024 ** 2;
	const Ki = 1024;

	if (bytes >= Gi) {
		return { value: formatNumberForDisplay(bytes / Gi), unit: 'Gi' };
	}
	if (bytes >= Mi) {
		return { value: formatNumberForDisplay(bytes / Mi), unit: 'Mi' };
	}
	if (bytes >= Ki) {
		return { value: formatNumberForDisplay(bytes / Ki), unit: 'Ki' };
	}
	return { value: String(bytes), unit: 'B' };
}

export function formatCpuForDisplay(raw: string | number | undefined | null): {
	value: string;
	unit: string;
} {
	if (raw === undefined || raw === null || raw === '') {
		return { value: '-', unit: '-' };
	}

	try {
		const res = parseCpuResource(raw);
		if (!res) return { value: '-', unit: '-' };

		const { value, unit } = splitK8sResource(res.toString());
		if (unit === 'm') return { value, unit: 'm' };
		const cores = Number(value);
		return {
			value,
			unit: Number.isNaN(cores) || cores === 1 ? 'core' : 'cores'
		};
	} catch (e) {
		console.error('formatCpuForDisplay error', e);
		return { value: '-', unit: '-' };
	}
}

export function formatMemoryForDisplay(
	raw: string | number | undefined | null
): { value: string; unit: string } {
	if (raw === undefined || raw === null || raw === '') {
		return { value: '-', unit: '-' };
	}

	try {
		const res = parseMemoryResource(raw);
		if (!res) return { value: '-', unit: '-' };

		return formatBytesForDisplay(res.valueOf());
	} catch (e) {
		console.error('formatMemoryForDisplay error', e);
		return { value: '-', unit: '-' };
	}
}

/** Cluster GPU metrics without unit are treated as GiB capacity. */
export function parseClusterGpuAvailable(
	resource: Resource
): MemoryResource | null {
	const available = Math.max((resource.total ?? 0) - (resource.usage ?? 0), 0);
	const unit = String(resource.unit ?? '').trim();

	if (unit) {
		return parseMemoryResource(`${available}${unit}`);
	}

	return parseMemoryResource(`${available}Gi`);
}

export function hasAcceleratorRequirements(
	appEntry?: AppEntry | null
): boolean {
	return !!appEntry?.accelerator && appEntry.accelerator.length > 0;
}

export function getAppResourceModes(appEntry: AppEntry): AppResourceMode[] {
	if (hasAcceleratorRequirements(appEntry)) {
		return appEntry.accelerator!.map((item) => ({
			mode: item.mode,
			requiredCpu: item.requiredCpu,
			requiredMemory: item.requiredMemory,
			requiredDisk: item.requiredDisk,
			requiredGPUMemory: item.requiredGPUMemory,
			isLegacy: false
		}));
	}

	return [
		{
			mode: 'legacy',
			requiredCpu: appEntry.requiredCPU,
			requiredMemory: appEntry.requiredMemory,
			requiredDisk: appEntry.requiredDisk,
			requiredGPUMemory: appEntry.requiredGPU,
			isLegacy: true
		}
	];
}

export function isGpuComputeMode(mode: string): boolean {
	const normalized = mode.replace(/_/g, '-');
	return (
		normalized !== COMPUTE_TYPE.CPU &&
		normalized !== 'cpu-only' &&
		normalized !== 'legacy'
	);
}

function isCpuInsufficient(
	available: CPUResource | null,
	required: string | number | undefined | null
): boolean {
	if (required === undefined || required === null || required === '') {
		return false;
	}
	const requiredRes = parseCpuResource(required);
	if (!requiredRes) return false;
	if (!available) return true;
	return requiredRes.isGreaterThan(available);
}

function isMemoryInsufficient(
	available: MemoryResource | null,
	required: string | number | undefined | null
): boolean {
	if (required === undefined || required === null || required === '') {
		return false;
	}
	const requiredRes = parseMemoryResource(required);
	if (!requiredRes) return false;
	if (!available) return true;
	return requiredRes.isGreaterThan(available);
}

export function hasArchSupport(
	appEntry: AppEntry,
	cluster: TerminusResource | null | undefined
): boolean {
	if (!appEntry.supportArch || appEntry.supportArch.length === 0) {
		return false;
	}
	if (!cluster?.nodes || cluster.nodes.length === 0) {
		return false;
	}
	return intersection(cluster.nodes, appEntry.supportArch).length > 0;
}

export function evaluateResourceMode(
	appEntry: AppEntry,
	mode: AppResourceMode,
	resources: NormalizedSystemResources
): { status: AppAcceleratorStatus; shortfalls: ResourceShortfall[] } {
	const { cluster, computeNodes } = resources;
	const shortfalls: ResourceShortfall[] = [];

	if (!cluster?.metrics) {
		return { status: 'no-matching-node', shortfalls };
	}

	if (!hasArchSupport(appEntry, cluster)) {
		return { status: 'no-matching-node', shortfalls };
	}

	const gpuMode = isGpuComputeMode(mode.mode);

	if (gpuMode) {
		const supportedModes = getSupportedModes(computeNodes);
		if (supportedModes.size > 0) {
			// Accurate path: a GPU mode needs a node that actually exposes it.
			if (!supportedModes.has(normalizeComputeMode(mode.mode))) {
				return { status: 'no-matching-node', shortfalls };
			}
		} else {
			// Fallback when compute-resources is unavailable: only the
			// aggregate cluster GPU total can be checked.
			const clusterGpuTotal = cluster.metrics.gpu?.total ?? 0;
			if (clusterGpuTotal <= 0) {
				return { status: 'no-matching-node', shortfalls };
			}
		}
	}

	const availableClusterCpu = parseCpuResource(
		Math.max(cluster.metrics.cpu.total - cluster.metrics.cpu.usage, 0)
	);
	const availableClusterMemory = parseMemoryResource(
		Math.max(cluster.metrics.memory.total - cluster.metrics.memory.usage, 0)
	);
	const availableClusterDisk = parseMemoryResource(
		Math.max(cluster.metrics.disk.total - cluster.metrics.disk.usage, 0)
	);

	if (isCpuInsufficient(availableClusterCpu, mode.requiredCpu)) {
		shortfalls.push('cluster-cpu');
	}

	if (isMemoryInsufficient(availableClusterMemory, mode.requiredMemory)) {
		shortfalls.push('cluster-memory');
	}

	if (isMemoryInsufficient(availableClusterDisk, mode.requiredDisk)) {
		shortfalls.push('cluster-disk');
	}

	if (gpuMode && mode.requiredGPUMemory) {
		const availableGpu = parseClusterGpuAvailable(cluster.metrics.gpu);
		const requiredGpu = parseMemoryResource(mode.requiredGPUMemory);
		if (
			requiredGpu &&
			availableGpu &&
			requiredGpu.isGreaterThan(availableGpu)
		) {
			shortfalls.push('cluster-gpu');
		} else if (requiredGpu && !availableGpu) {
			shortfalls.push('cluster-gpu');
		}
	}

	return {
		status: shortfalls.length > 0 ? 'insufficient-resources' : 'installable',
		shortfalls
	};
}

export function evaluateAllResourceModes(
	appEntry: AppEntry,
	resources: NormalizedSystemResources
): EvaluatedResourceMode[] {
	return getAppResourceModes(appEntry).map((mode) => {
		const { status, shortfalls } = evaluateResourceMode(
			appEntry,
			mode,
			resources
		);
		return { ...mode, status, shortfalls };
	});
}

export function unwrapSystemStatusPayload(payload: any): {
	user_info?: { role: string; username: string };
	cur_user_resource?: UserResource;
	cluster_resource?: TerminusResource;
	terminus_version?: { version: string };
} | null {
	if (!payload) return null;
	if (payload.data && (payload.success !== undefined || payload.message)) {
		return payload.data;
	}
	return payload;
}

export function acceleratorSummaryLabel(modes: AppAcceleratorRequirement[]): {
	primary: string;
	extraCount: number;
} {
	return {
		primary: modes[0]?.mode ?? '',
		extraCount: Math.max(modes.length - 1, 0)
	};
}
