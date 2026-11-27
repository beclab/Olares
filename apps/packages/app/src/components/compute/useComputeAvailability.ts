import {
	ComputeAvailabilityDevice,
	ComputeAvailabilityRefetch,
	ComputeAvailabilityResponse,
	ComputeBindingValidation,
	applyExclusiveBoundAppMemory,
	computeTypeLabel,
	normalizeComputeAvailabilityResponse,
	parseValidationCode
} from 'src/constant/compute';
import { AcceleratorAppView } from 'src/stores/settings/accelerator';
import { ComputeNodeView } from './types';
import { computed, ref } from 'vue';

export type ComputeAppResolver = (
	appName: string
) => Pick<AcceleratorAppView, 'title' | 'icon' | 'state'>;

export interface UseComputeAvailabilityOptions {
	availability: ComputeAvailabilityResponse;
	validation?: ComputeBindingValidation;
	refetch?: ComputeAvailabilityRefetch;
	resolveApp: ComputeAppResolver;
}

/** Display order: available → not_available → not_enough (bottom). */
const NODE_STATUS_ORDER: Record<string, number> = {
	available: 0,
	not_available: 1,
	not_enough: 2
};

/**
 * Owns the resume single-source data: the availability + validation snapshot
 * from the failed-operation prompt, the validate-only `refresh`, and the view
 * model (`nodeViews`) the dialog renders. No client recompute — every field is
 * backend-authoritative.
 */
export function useComputeAvailability(opts: UseComputeAvailabilityOptions) {
	const { resolveApp } = opts;

	const availability = ref<ComputeAvailabilityResponse>(
		normalizeComputeAvailabilityResponse(opts.availability)
	);
	const validation = ref<ComputeBindingValidation | undefined>(opts.validation);
	const refreshing = ref(false);

	const requirement = computed(() => availability.value.requirement);

	/** Re-run the validate-only check so freed resources reflect immediately. */
	const refresh = async () => {
		if (!opts.refetch || refreshing.value) return;
		refreshing.value = true;
		try {
			const next = await opts.refetch();
			if (next) {
				availability.value = normalizeComputeAvailabilityResponse(
					next.availability
				);
				validation.value = next.validation;
			}
		} finally {
			refreshing.value = false;
		}
	};

	/* ---- validation ---- */
	// Only a failed validation surfaces in the UI; `ok` snapshots stay silent.
	const parsedValidation = computed(() =>
		validation.value && !validation.value.ok
			? parseValidationCode(validation.value.code)
			: null
	);

	/* ---- view model ---- */
	// An app spanning more than one card across the prompt is "multi-card".
	const isMultiCardOnPrompt = (appName: string): boolean => {
		let count = 0;
		availability.value.nodes.forEach((n) =>
			(n.devices ?? []).forEach((d) => {
				if ((d.bindings ?? []).some((b) => b.appName === appName)) count++;
			})
		);
		return count > 1;
	};

	// App titles / icons are resolved from the catalog (not a GPU concern).
	const bindingApps = (
		dev: ComputeAvailabilityDevice,
		gpuType: string
	): AcceleratorAppView[] => {
		const apps = (dev.bindings ?? []).map((b) => {
			const info = resolveApp(b.appName);
			const spec = b.spec;
			return {
				appName: b.appName,
				title: info.title,
				icon: info.icon,
				state: info.state,
				memory: b.memory ?? 0,
				multiCard: isMultiCardOnPrompt(b.appName),
				mode: String(b.mode ?? gpuType ?? ''),
				scope: spec
					? {
							supportMultiCards: !!spec.supportMultiCards,
							supportMultiNodes: !!spec.supportMultiNodes
					  }
					: undefined
			};
		});
		return applyExclusiveBoundAppMemory(apps, dev.supportType, dev.capacity);
	};

	const nodeViews = computed<ComputeNodeView[]>(() =>
		availability.value.nodes
			.filter((node) => node.status !== 'not_match')
			.map((node) => {
				const devices = node.devices ?? [];
				return {
					nodeName: node.nodeName,
					gpuType: node.gpuType,
					typeTags: nodeTypeTags(node.gpuType),
					status: node.status,
					raw: node,
					devices: devices.map((dev) => ({
						deviceId: dev.deviceId,
						name:
							dev.cardModel || computeTypeLabel(node.gpuType) || dev.deviceId,
						supportType: dev.supportType,
						capacity: dev.capacity,
						available: dev.available,
						used: Math.max(dev.capacity - dev.available, 0),
						operable: dev.operable,
						fitLevel: dev.fitLevel,
						boundApps: bindingApps(dev, node.gpuType),
						raw: dev
					}))
				};
			})
			.sort(
				(a, b) =>
					(NODE_STATUS_ORDER[a.status] ?? 99) -
					(NODE_STATUS_ORDER[b.status] ?? 99)
			)
	);

	return {
		availability,
		validation,
		refreshing,
		refresh,
		requirement,
		parsedValidation,
		nodeViews
	};
}

const nodeTypeTags = (gpuType?: string): string[] => {
	const label = computeTypeLabel(gpuType);
	return label ? [label] : [];
};
