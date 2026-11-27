// Compute domain types + dialog visibility/operability rules.
// 422 error shapes live in appOperationError.ts.

export type ComputeType = string;

export const COMPUTE_TYPE = {
	CPU: 'cpu',
	NVIDIA: 'nvidia',
	NVIDIA_GB10: 'nvidia-gb10',
	APPLE_M: 'apple-m',
	INTEL: 'intel',
	AMD: 'amd',
	INTEL_GPU: 'intel-gpu',
	AMD_GPU: 'amd-gpu',
	MOORE_SOC: 'moore-soc'
} as const;

/**
 * Per-device support type reported by the cluster.
 * Nvidia: Exclusive | MemorySlice | TimeSlice.
 * Integrated / shared-memory accelerators may also report MemoryShared.
 */
export type DeviceSupportType =
	| 'Exclusive'
	| 'MemorySlice'
	| 'TimeSlice'
	| 'MemoryShared';

/**
 * i18n keys for {@link DeviceSupportType} — same wording as settings GPU
 * (`VRAMModeLabel`: App exclusive / Memory slicing / Time slicing).
 */
export const DEVICE_SUPPORT_TYPE_LABEL_KEY: Record<DeviceSupportType, string> =
	{
		Exclusive: 'App exclusive',
		MemorySlice: 'Memory slicing',
		TimeSlice: 'Time slicing',
		MemoryShared: 'Memory shared'
	};

export function deviceSupportTypeLabelKey(
	type: DeviceSupportType | string
): string {
	return (
		DEVICE_SUPPORT_TYPE_LABEL_KEY[type as DeviceSupportType] ?? String(type)
	);
}

/** Status of a single compute mode for the current cluster (computeModeSelect payload). */
export type ComputeModeStatus =
	| 'installable'
	| 'insufficient-resources'
	| 'no-matching-node';

export interface ComputeModeOption {
	computeType: ComputeType;
	status: ComputeModeStatus;
}

/**
 * Status of a node for a given app launch (computeBindingRequired payload):
 *   not_match     — gpuType doesn't match the app's mode (hidden in the dialog)
 *   not_enough    — arch matches but hardware capacity can't fit the app
 *   not_available — hardware fits but currently occupied; free space insufficient
 *   available     — fits and can be bound right now
 */
export type ComputeNodeStatus =
	| 'not_match'
	| 'not_enough'
	| 'not_available'
	| 'available';

export interface ComputeAvailabilityDevice {
	deviceId: string;
	supportType: DeviceSupportType;
	capacity: number;
	available: number;
	fitLevel: string;
	health: string;
	operable: boolean;
	/** Card display name, mirrored from /compute-resources into the prompt. */
	cardModel?: string;
	/** Apps currently bound to this card, mirrored from /compute-resources. */
	bindings?: ComputeResourceBinding[];
}

export interface ComputeAvailabilityNode {
	nodeName: string;
	gpuType: ComputeType;
	status: ComputeNodeStatus;
	devices: ComputeAvailabilityDevice[];
	/** Node-level CPU / memory usage (optional, from backend). */
	cpu?: ComputeResourceUsage;
	memory?: ComputeResourceUsage;
}

export interface ComputeRequirement {
	mode: ComputeType;
	requiredGpu: number;
	/** Required CPU cores (CPU compute mode). */
	requiredCpu?: number;
	requiredMemory?: number;
	supportMultiCards?: boolean;
	supportMultiNodes?: boolean;
	[key: string]: any;
}

/**
 * Selection scope returned by the backend (after {@link normalizeComputeBindingScope}):
 *   card              — pick exactly one card
 *   single-node-cards — pick several cards on the same node
 *   cross-node-cards  — pick cards across nodes
 *
 * Legacy wire value `'node'` is normalized to `'card'` at parse time.
 */
export type ComputeBindingScope =
	| 'card'
	| 'single-node-cards'
	| 'cross-node-cards';

/** Legacy backend scope; map to `'card'` via {@link normalizeComputeBindingScope}. */
const LEGACY_NODE_SCOPE = 'node';

export interface ComputeBindingValidation {
	ok: boolean;
	code: string;
	reason: string;
	nodePressure?: ComputeNodePressure;
}

/** One pressured (or measured) resource dimension on a node. */
export interface ComputeNodePressureDimension {
	/** Resource kind: 'memory' | 'cpu' | 'disk'. */
	resource: string;
	capacity: number;
	used: number;
	available: number;
	required: number;
	pressured: boolean;
}

/** Node-level scheduling pressure breakdown returned with `node-pressure`. */
export interface ComputeNodePressure {
	nodeName: string;
	dimensions: ComputeNodePressureDimension[];
}

/**
 * Validation code prefixes returned by the backend on a failed compute binding
 * (e.g. resume). The full code is `<prefix>:<targetId>`, where `targetId` is a
 * deviceId for the device-scoped prefixes and a nodeName for `node-pressure`.
 */
export const COMPUTE_VALIDATION_PREFIX = {
	/** Multi-card: aggregate VRAM across the selected cards is insufficient. */
	AGGREGATE_VRAM_INSUFFICIENT: 'aggregate-vram-insufficient',
	/** Single card: the card's VRAM is insufficient. */
	DEVICE_VRAM_INSUFFICIENT: 'device-vram-insufficient',
	/** Single card: the card's (system) memory is insufficient. */
	DEVICE_MEMORY_INSUFFICIENT: 'device-memory-insufficient',
	/** Node-level scheduling pressure. */
	NODE_PRESSURE: 'node-pressure'
} as const;

export type ComputeValidationPrefix =
	(typeof COMPUTE_VALIDATION_PREFIX)[keyof typeof COMPUTE_VALIDATION_PREFIX];

/** i18n key for each validation prefix (used for banner + per-target message). */
export const COMPUTE_VALIDATION_I18N_KEY: Record<string, string> = {
	[COMPUTE_VALIDATION_PREFIX.AGGREGATE_VRAM_INSUFFICIENT]:
		'compute.validation_code.aggregate_vram_insufficient',
	[COMPUTE_VALIDATION_PREFIX.DEVICE_VRAM_INSUFFICIENT]:
		'compute.validation_code.device_vram_insufficient',
	[COMPUTE_VALIDATION_PREFIX.DEVICE_MEMORY_INSUFFICIENT]:
		'compute.validation_code.device_memory_insufficient',
	[COMPUTE_VALIDATION_PREFIX.NODE_PRESSURE]:
		'compute.validation_code.node_pressure'
};

export interface ParsedValidationCode {
	/** The prefix part, or the whole string when there is no colon. */
	prefix: string;
	/** Everything after the first colon — a deviceId or nodeName ('' if absent). */
	targetId: string;
	/** True when the prefix targets a single device (deviceId) rather than a node. */
	isDeviceScoped: boolean;
	/** i18n key for the message, or '' for an unknown prefix. */
	i18nKey: string;
}

/**
 * Split a backend validation code `<prefix>:<targetId>` into its parts. The
 * targetId is everything after the FIRST colon (deviceIds contain hyphens but
 * no colons; nodeNames are plain). Returns prefix='' for an empty code.
 */
export function parseValidationCode(code?: string): ParsedValidationCode {
	const raw = (code ?? '').trim();
	const idx = raw.indexOf(':');
	const prefix = idx >= 0 ? raw.slice(0, idx) : raw;
	const targetId = idx >= 0 ? raw.slice(idx + 1) : '';
	const isDeviceScoped =
		prefix === COMPUTE_VALIDATION_PREFIX.DEVICE_VRAM_INSUFFICIENT ||
		prefix === COMPUTE_VALIDATION_PREFIX.DEVICE_MEMORY_INSUFFICIENT;
	return {
		prefix,
		targetId,
		isDeviceScoped,
		i18nKey: COMPUTE_VALIDATION_I18N_KEY[prefix] ?? ''
	};
}

export interface ComputeAvailabilityResponse {
	schedulable: boolean;
	requirement: ComputeRequirement;
	scope: ComputeBindingScope;
	nodes: ComputeAvailabilityNode[];
	validation?: ComputeBindingValidation;
}

/** Fresh availability + validation, re-fetched after the user frees resources. */
export interface ComputeAvailabilitySnapshot {
	availability: ComputeAvailabilityResponse;
	validation?: ComputeBindingValidation;
}

/**
 * Dry-run re-fetch of the compute binding prompt. Re-runs the operation (resume)
 * without a binding so the backend returns a fresh 422 availability + validation
 * without committing anything. Returns null when no prompt comes back.
 */
export type ComputeAvailabilityRefetch =
	() => Promise<ComputeAvailabilitySnapshot | null>;

/** Payload posted back on resume after the user picks node + card(s). */
export interface ComputeBindingItem {
	nodeName: string;
	deviceId: string;
	memory?: number;
}

/* ---------- Requirement resource helpers ---------- */

/** Only plain nvidia uses VRAM (`requiredGpu`); gb10 and others use `requiredMemory`. */
export function isNvidiaComputeRequirement(req: ComputeRequirement): boolean {
	return req.mode === COMPUTE_TYPE.NVIDIA;
}

/** Bytes threshold for summary display and MemorySlice validation. */
export function getComputeRequirementResourceBytes(
	req: ComputeRequirement
): number {
	if (isNvidiaComputeRequirement(req)) {
		return req.requiredGpu ?? 0;
	}
	return req.requiredMemory ?? 0;
}

/**
 * Normalize a backend scope string. Legacy `'node'` is treated as single-card
 * selection (`'card'`). Unknown values fall back to `'card'`.
 */
export function normalizeComputeBindingScope(
	raw: string | undefined | null
): ComputeBindingScope {
	if (raw === LEGACY_NODE_SCOPE || raw === 'card') return 'card';
	if (raw === 'single-node-cards' || raw === 'cross-node-cards') return raw;
	return 'card';
}

/** Apply {@link normalizeComputeBindingScope} to an availability snapshot. */
export function normalizeComputeAvailabilityResponse(
	availability: ComputeAvailabilityResponse
): ComputeAvailabilityResponse {
	return {
		...availability,
		scope: normalizeComputeBindingScope(availability.scope)
	};
}

/* ---------- Visibility / operability rules ---------- */

/**
 * Whether a node lets the user interact with it. The backend already encodes
 * the visibility/operability rules into each device's `operable` flag
 * (`markOperable`), so a node is interactive iff it has at least one operable
 * device. `not_enough` nodes carry no operable devices.
 */
export function isNodeOperable(node: ComputeAvailabilityNode): boolean {
	return (node.devices ?? []).some((d) => d.operable);
}

/**
 * Whether an individual device (card) can be ticked by the user — the backend
 * `operable` flag is authoritative.
 */
export function isDeviceOperable(device: ComputeAvailabilityDevice): boolean {
	return !!device.operable;
}

/**
 * Cap on how many cards the user can tick, derived from the backend scope.
 *   - single-node-cards | cross-node-cards => unbounded
 *   - card                                 => 1
 */
export function maxSelectableCards(scope: ComputeBindingScope): number {
	if (scope === 'single-node-cards' || scope === 'cross-node-cards') {
		return Number.POSITIVE_INFINITY;
	}
	return 1;
}

/**
 * Whether the user must stay on the same node when selecting multiple cards.
 * True only for the "single-node multi-card" scope.
 */
export function mustStayOnSameNode(scope: ComputeBindingScope): boolean {
	return scope === 'single-node-cards';
}

/* ---------- Mode picker helpers (install computeModeSelect) ---------- */

const MODE_STATUS_ORDER: Record<ComputeModeStatus, number> = {
	installable: 0,
	'insufficient-resources': 1,
	'no-matching-node': 2
};

/** Stable ordering for the radio list: installable first, CPU last. */
export function sortModeOptions(
	options: ComputeModeOption[]
): ComputeModeOption[] {
	return [...options].sort((a, b) => {
		const sa = MODE_STATUS_ORDER[a.status] ?? 99;
		const sb = MODE_STATUS_ORDER[b.status] ?? 99;
		if (sa !== sb) return sa - sb;
		if (a.computeType === COMPUTE_TYPE.CPU) return 1;
		if (b.computeType === COMPUTE_TYPE.CPU) return -1;
		return String(a.computeType).localeCompare(String(b.computeType));
	});
}

/** Default selection for the radio list — first installable, else first row. */
export function pickDefaultMode(
	options: ComputeModeOption[]
): ComputeType | undefined {
	const sorted = sortModeOptions(options);
	const installable = sorted.find((o) => o.status === 'installable');
	return (installable ?? sorted[0])?.computeType;
}

/* ===================================================================== */
/* compute-resources (Settings Accelerator page, Olares PR #3118)        */
/* ===================================================================== */

/** A single app bound to a device, as returned by /compute-resources. */
export interface ComputeResourceBinding {
	appId?: string;
	appName: string;
	owner?: string;
	mode?: ComputeType;
	nodeName?: string;
	deviceId: string;
	/** Allocated memory in bytes. */
	memory?: number;
	/** Per-binding requirement/capability spec from /compute-resources. */
	spec?: ComputeRequirement;
}

/** A GPU/accelerator card on a node. */
export interface ComputeResourceDevice {
	id: string;
	nodeName: string;
	/** Compute mode this device belongs to (e.g. 'nvidia', 'intel'). */
	mode?: ComputeType | string;
	/** Card display name; absent for integrated accelerators. */
	cardModel?: string;
	/** Capacity in bytes. */
	memory: number;
	health?: string;
	supportType: DeviceSupportType | string;
	availableSupportTypes?: (DeviceSupportType | string)[];
	bindings?: ComputeResourceBinding[];
	[key: string]: any;
}

/** used/total pair for node-level CPU / memory / aggregated VRAM summaries. */
export interface ComputeResourceUsage {
	used: number;
	total: number;
}

/** A node and all of its accelerator devices. */
export interface ComputeResourceNode {
	nodeName: string;
	/** Compute modes the node exposes, as returned by the backend (e.g. ['nvidia', 'intel']). */
	gpuTypes?: (ComputeType | string)[];
	health?: string;
	devices?: ComputeResourceDevice[];
	[key: string]: any;
}

/** App reference inside a support-type switch result. */
export interface ComputeSupportTypeAppRef {
	appName: string;
	owner?: string;
	state?: string;
	reason?: string;
}

/**
 * Result of `PUT .../support-type`. `status` discriminates the outcome
 * regardless of the upstream business code (200 switched/unchanged, 4603 blocked).
 */
export interface ComputeSupportTypeResult {
	status: 'switched' | 'unchanged' | 'bound-apps-stop-blocked' | string;
	device?: ComputeResourceDevice;
	stoppedApps?: ComputeSupportTypeAppRef[];
	blockedApps?: ComputeSupportTypeAppRef[];
}

/**
 * supportType label keys for the Accelerator page — wording matches the Figma
 * design ("Exclusive" / "Memory Slicing" / "Time Slicing").
 */
export const ACCELERATOR_SUPPORT_TYPE_LABEL_KEY: Record<string, string> = {
	Exclusive: 'Exclusive',
	MemorySlice: 'Memory Slicing',
	TimeSlice: 'Time Slicing',
	MemoryShared: 'Memory Shared'
};

export function acceleratorSupportTypeLabelKey(type: string): string {
	return ACCELERATOR_SUPPORT_TYPE_LABEL_KEY[type] ?? String(type);
}

/**
 * Badge tone for a device support type, matching the Figma color language:
 * Time Slicing -> green, Memory Slicing -> blue, Exclusive -> red.
 */
export function acceleratorSupportTypeTone(type: string): ComputeTone {
	if (type === 'TimeSlice') return 'success';
	if (type === 'Exclusive') return 'bad';
	if (type === 'MemorySlice' || type === 'MemoryShared') return 'info';
	return 'neutral';
}

/**
 * Sum of allocated memory (bytes) across a device's bindings, capped at the
 * device capacity so the displayed "used / total" never exceeds total (e.g. an
 * Exclusive binding rounding up could otherwise show 13G / 12G).
 */
export function deviceUsedMemory(device: ComputeResourceDevice): number {
	const used = (device.bindings ?? []).reduce(
		(acc, b) => acc + (b.memory ?? 0),
		0
	);
	const capacity = device.memory ?? 0;
	return capacity > 0 ? Math.min(used, capacity) : used;
}

/** Exclusive = the whole card is dedicated to a single app. */
export function isExclusiveSupportType(type?: string | null): boolean {
	return type === 'Exclusive';
}

/**
 * Time slicing shares the whole card by time, so per-app memory usage is not
 * meaningful — the UI shows only the card capacity and hides the usage bar.
 */
export function isTimeSliceSupportType(type?: string | null): boolean {
	return type === 'TimeSlice';
}

/**
 * VRAM "used" for display. An Exclusive device dedicates the entire card to its
 * bound app, so once an app is bound it reads as fully used (overriding the
 * backend per-binding usage). With no bound app the card is idle, so it falls
 * back to the summed binding usage instead of forcing a full bar; other modes
 * always use the summed binding usage.
 */
export function effectiveDeviceUsedMemory(
	device: ComputeResourceDevice
): number {
	const hasBinding = (device.bindings ?? []).length > 0;
	return isExclusiveSupportType(device.supportType) && hasBinding
		? device.memory ?? 0
		: deviceUsedMemory(device);
}

/**
 * Display rule shared by settings + market dialog: an Exclusive device
 * dedicates the whole card to its bound app, so each bound app reads as
 * occupying the full card `capacity` (overriding the per-binding memory).
 * Non-exclusive support types are returned unchanged. Generic over the view
 * type so both `ComputeResourceDevice` (capacity = `memory`) and
 * `ComputeAvailabilityDevice` (capacity = `capacity`) callers can reuse it.
 */
export function applyExclusiveBoundAppMemory<T extends { memory: number }>(
	apps: T[],
	supportType: string | null | undefined,
	capacity: number
): T[] {
	if (!isExclusiveSupportType(supportType)) return apps;
	return apps.map((app) => ({ ...app, memory: capacity }));
}

/* ===================================================================== */
/* Shared UI helpers (compute dialogs + Settings accelerator pages)      */
/* ===================================================================== */

/** used / total pair consumed by the shared usage bar. */
export interface ComputeUsageValue {
	used: number;
	total: number;
}

/** Tone keyword used by the shared status / badge styling. */
export type ComputeTone =
	| 'ok'
	| 'success'
	| 'warn'
	| 'bad'
	| 'info'
	| 'brand'
	| 'neutral';

// Tolerance to absorb byte<->Gi float drift before floor/ceil at 2 decimals.
const GI_ROUND_EPSILON = 1e-6;

// Numeric Gi value (no unit), floored to 2 decimals so it never overstates capacity.
export function formatComputeMemoryValue(bytes: number): string {
	if (!bytes || bytes <= 0) return '0';
	const g = bytes / 1024 ** 3;
	const floored = Math.floor(g * 100 + GI_ROUND_EPSILON) / 100;
	return Number.isInteger(floored) ? String(floored) : floored.toFixed(2);
}

export function formatComputeMemory(bytes: number, splite = ' '): string {
	return `${formatComputeMemoryValue(bytes)}${splite}Gi`;
}

// Ceiled variant for requirement thresholds, so "needed" never understates the real value.
export function formatComputeMemoryCeil(bytes: number): string {
	if (!bytes || bytes <= 0) return '0 Gi';
	const g = bytes / 1024 ** 3;
	const ceiled = Math.ceil(g * 100 - GI_ROUND_EPSILON) / 100;
	return `${Number.isInteger(ceiled) ? ceiled : ceiled.toFixed(2)} Gi`;
}

/** i18n key for each node-pressure resource dimension label. */
export const COMPUTE_PRESSURE_RESOURCE_LABEL_KEY: Record<string, string> = {
	memory: 'compute.pressure_resource.memory',
	cpu: 'compute.pressure_resource.cpu',
	disk: 'compute.pressure_resource.disk'
};

/** Resolve the label i18n key for a pressure resource (falls back to the raw value). */
export function computePressureResourceLabelKey(resource: string): string {
	return (
		COMPUTE_PRESSURE_RESOURCE_LABEL_KEY[resource] ?? String(resource ?? '')
	);
}

// CPU values arrive in millicores (e.g. 150 = 0.15 core); show whole cores,
// trimming to at most 2 decimals and dropping trailing zeros.
export function formatComputeCpuCores(milli: number): string {
	if (!milli || milli <= 0) return '0 cores';
	const cores = milli / 1000;
	const rounded = Math.round(cores * 100) / 100;
	const value = Number.isInteger(rounded)
		? String(rounded)
		: rounded.toFixed(2);
	return `${value} ${rounded === 1 ? 'core' : 'cores'}`;
}

/** Format a node-pressure amount by resource: cpu → cores, memory/disk → Gi. */
export function formatComputePressureAmount(
	resource: string,
	value: number
): string {
	return resource === 'cpu'
		? formatComputeCpuCores(value)
		: formatComputeMemory(value);
}

/** The pressured dimensions of a node-pressure breakdown (empty when none). */
export function pressuredDimensions(
	np?: ComputeNodePressure
): ComputeNodePressureDimension[] {
	return (np?.dimensions ?? []).filter((d) => d.pressured);
}

/** Clamp a used/total pair to a 0..1 progress ratio. */
export function computeUsageRatio(usage: ComputeUsageValue): number {
	if (!usage.total || usage.total <= 0) return 0;
	return Math.min(Math.max(usage.used / usage.total, 0), 1);
}

/** Status badge tone for a node row. */
export function nodeStatusTone(
	status: ComputeNodeStatus | string
): ComputeTone {
	if (status === 'available') return 'ok';
	if (status === 'not_available') return 'warn';
	if (status === 'not_enough') return 'bad';
	return 'neutral';
}

/**
 * Display titles for compute modes. Brand names — shown identically in every
 * locale. Keyed by the canonical backend ResourceMode values; unknown values
 * fall back to the raw string via {@link computeModeTitle}.
 */
// i18n-ignore-start
export const COMPUTE_MODE_TITLE: Record<string, string> = {
	cpu: 'CPU',
	nvidia: 'NVIDIA GPU',
	'nvidia-gb10': 'NVIDIA GB10',
	'apple-m': 'Apple Silicon-M',
	intel: 'Intel',
	amd: 'AMD',
	'intel-gpu': 'Intel GPU',
	'amd-gpu': 'AMD GPU',
	'moore-soc': 'Moore Threads'
};
// i18n-ignore-end

/** Normalize a compute mode value so spelling variants (case, `_` vs `-`) match. */
export function normalizeComputeMode(mode?: string | null): string {
	return String(mode ?? '')
		.trim()
		.toLowerCase()
		.replace(/_/g, '-');
}

/** Resolve the display title for a compute mode value (falls back to the raw value). */
export function computeModeTitle(type: string): string {
	return COMPUTE_MODE_TITLE[normalizeComputeMode(type)] ?? String(type ?? '');
}

/** Compute type label for chips / bars — empty only when the type is missing. */
export function computeTypeLabel(type?: string): string {
	if (!type) return '';
	return computeModeTitle(type);
}

/**
 * Modes whose device `memory` is dedicated VRAM: discrete GPUs only. Integrated
 * accelerators (intel / amd / apple-m / nvidia-gb10 / moore-soc) share system
 * memory, so their `memory` counts towards the node's Memory total instead.
 */
export const VRAM_COMPUTE_MODES: ReadonlySet<string> = new Set([
	COMPUTE_TYPE.NVIDIA,
	COMPUTE_TYPE.INTEL_GPU,
	COMPUTE_TYPE.AMD_GPU
]);

/** Whether a device of this mode reports dedicated VRAM (vs shared memory). */
export function isVramComputeMode(mode?: string | null): boolean {
	return VRAM_COMPUTE_MODES.has(normalizeComputeMode(mode));
}

export interface AppComputeBadge {
	/** A resolved label (brand name) or an i18n key — see `i18nKey`. */
	label: string;
	/** True when `label` is an i18n key to resolve via t(). */
	i18nKey: boolean;
}

/** App scope capability from backend spec; absence means "unknown". */
export interface AppComputeScope {
	supportMultiCards: boolean;
	supportMultiNodes: boolean;
}

export function appComputeBadges(
	mode: string | undefined | null,
	scope?: AppComputeScope | null
): AppComputeBadge[] {
	const badges: AppComputeBadge[] = [];
	const typeLabel = computeModeTitle(String(mode ?? ''));
	if (typeLabel) {
		badges.push({ label: typeLabel, i18nKey: false });
	}
	// Scope tag only when capability is known (spec present), NVIDIA only.
	// multi-node takes precedence over the single/multi-card tag.
	if (scope && normalizeComputeMode(mode) === COMPUTE_TYPE.NVIDIA) {
		const label = scope.supportMultiNodes
			? 'compute.multi_node_chip'
			: scope.supportMultiCards
			? 'compute.multi_card_chip'
			: 'compute.single_card_chip';
		badges.push({ label, i18nKey: true });
	}
	return badges;
}
