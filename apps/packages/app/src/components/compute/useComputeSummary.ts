import {
	COMPUTE_TYPE,
	ComputeRequirement,
	computeTypeLabel,
	formatComputeMemory,
	formatComputeMemoryCeil,
	getComputeRequirementResourceBytes,
	normalizeComputeMode
} from 'src/constant/compute';
import { ComputeNodeView } from './types';
import { ComputeBinding } from './useComputeBinding';
import { computed, Ref } from 'vue';
import { useI18n } from 'vue-i18n';

export interface UseComputeSummaryOptions {
	binding: ComputeBinding;
	nodeViews: Ref<ComputeNodeView[]>;
	requirement: Ref<ComputeRequirement>;
}

/**
 * Header summary + submit gating. Keeps the dialog template declarative.
 */
export function useComputeSummary(opts: UseComputeSummaryOptions) {
	const { binding, nodeViews, requirement } = opts;
	const { t } = useI18n();

	const requirementTags = computed(() => {
		const tags: string[] = [];
		const label = computeTypeLabel(requirement.value.mode);
		if (label) tags.push(label);
		const isNvidia =
			normalizeComputeMode(requirement.value.mode) === COMPUTE_TYPE.NVIDIA;
		if (isNvidia && requirement.value.supportMultiNodes) {
			tags.push(t('compute.multi_node_chip'));
		} else {
			tags.push(
				requirement.value.supportMultiCards
					? t('compute.multi_card_chip')
					: t('compute.single_card_chip')
			);
		}
		return tags;
	});

	const requiredBytes = computed(() =>
		getComputeRequirementResourceBytes(requirement.value)
	);

	const allocatedBytes = computed(() => {
		const selected = new Set(binding.selectedKeys.value);
		let sum = 0;
		nodeViews.value.forEach((node) => {
			node.devices.forEach((dev) => {
				const key = binding.deviceKey(node.raw, dev.raw);
				if (!selected.has(key)) return;
				sum += binding.cardContributionBytes(node.raw, dev.raw);
			});
		});
		return sum;
	});

	const requirementMet = computed(() => {
		if (!binding.canSubmit.value) return false;
		// A card whose inline quota input holds an invalid value must block
		// submission even if the aggregate (seeded) allocation looks sufficient.
		if (!binding.allAllocationsValid.value) return false;
		return allocatedBytes.value >= requiredBytes.value;
	});

	const hasSelection = computed(() => binding.selectedKeys.value.length > 0);

	const usageState = computed<'none' | 'sufficient' | 'insufficient'>(() => {
		if (!hasSelection.value) return 'none';
		return requirementMet.value ? 'sufficient' : 'insufficient';
	});

	const allocatedText = computed(() =>
		formatComputeMemory(allocatedBytes.value)
	);
	const requiredText = computed(
		() => `/ ${formatComputeMemoryCeil(requiredBytes.value)}`
	);
	const vramSummary = computed(
		() => `${allocatedText.value} ${requiredText.value}`
	);

	const usageValueClass = computed(() => {
		switch (usageState.value) {
			case 'sufficient':
				return 'text-green-default';
			case 'insufficient':
				return 'text-orange-default';
			default:
				return 'text-ink-1';
		}
	});
	const usageLabelClass = computed(() => {
		switch (usageState.value) {
			case 'sufficient':
				return 'text-green-default';
			case 'insufficient':
				return 'text-orange-default';
			default:
				return 'text-ink-3';
		}
	});
	const usageLabel = computed(() => {
		switch (usageState.value) {
			case 'sufficient':
				return t('compute.sufficient_vram');
			case 'insufficient':
				return t('compute.insufficient_vram');
			default:
				return t('compute.vram_allocation');
		}
	});

	const sectionTitle = computed(() => t('compute.select_compute_cards'));
	const hasUnavailable = computed(() =>
		nodeViews.value.some((n) => n.status !== 'available')
	);
	const okLabel = computed(() =>
		hasUnavailable.value ? t('compute.confirm_launch') : t('compute.launch')
	);

	return {
		requirementTags,
		requiredBytes,
		allocatedBytes,
		requirementMet,
		usageState,
		allocatedText,
		requiredText,
		vramSummary,
		usageValueClass,
		usageLabelClass,
		usageLabel,
		sectionTitle,
		okLabel
	};
}
