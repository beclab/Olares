import {
	ComputeAvailabilityDevice,
	ComputeAvailabilityNode,
	ComputeAvailabilityResponse,
	ComputeBindingItem,
	getComputeRequirementResourceBytes,
	isDeviceOperable,
	isExclusiveSupportType,
	isTimeSliceSupportType,
	maxSelectableCards,
	mustStayOnSameNode
} from 'src/constant/compute';
import { computed, reactive, ref, Ref } from 'vue';

const GiB = 1024 ** 3;

export interface VramAllocation {
	memory: number;
	exclusive: boolean;
}

/**
 * Selection state machine shared by the compute binding dialog.
 * Handles card selection (checkbox / radio), the single-node multi-card
 * constraint, per-card VRAM allocations, and building the payload.
 */
export function useComputeBinding(
	availability: Ref<ComputeAvailabilityResponse>
) {
	const requirement = computed(() => availability.value.requirement);
	const requiredPerCard = computed(() =>
		getComputeRequirementResourceBytes(requirement.value)
	);

	const scope = computed(() => availability.value.scope);
	const allowMultiCard = computed(() => maxSelectableCards(scope.value) > 1);
	const mustSameNode = computed(() => mustStayOnSameNode(scope.value));

	const deviceKey = (
		node: ComputeAvailabilityNode,
		dev: ComputeAvailabilityDevice
	) => `${node.nodeName}::${dev.deviceId}`;

	const selectedKeys = ref<string[]>([]);
	const allocations = reactive<Record<string, VramAllocation>>({});
	// Per-card input validity, reported by the inline quota editor. Absence is
	// treated as valid (Exclusive has no input; TimeSlice is read-only), so only
	// an explicit `false` from an editable MemorySlice card gates submission.
	const allocationValid = reactive<Record<string, boolean>>({});

	const singleSelectedKey = computed(() =>
		allowMultiCard.value ? '' : selectedKeys.value[0] ?? ''
	);

	const isDeviceChecked = (
		node: ComputeAvailabilityNode,
		dev: ComputeAvailabilityDevice
	) => selectedKeys.value.includes(deviceKey(node, dev));

	const isDeviceSelectable = (
		dev: ComputeAvailabilityDevice,
		node: ComputeAvailabilityNode
	): boolean => {
		if (!isDeviceOperable(dev)) return false;
		if (mustSameNode.value && selectedKeys.value.length > 0) {
			const lockedNode = selectedKeys.value[0].split('::')[0];
			if (lockedNode !== node.nodeName) return false;
		}
		return true;
	};

	const defaultAllocation = (
		dev: ComputeAvailabilityDevice
	): VramAllocation => ({
		// Seed with the app's actual per-card requirement (no 1 GB floor). Fall
		// back to 1 GB only when there is no per-card requirement, so a freshly
		// checked card never seeds an invalid 0-byte allocation.
		memory: Math.min(dev.available, requiredPerCard.value || GiB),
		exclusive: false
	});

	const addAllocation = (key: string, dev: ComputeAvailabilityDevice) => {
		if (dev.supportType === 'MemorySlice' && !allocations[key]) {
			allocations[key] = defaultAllocation(dev);
		}
	};

	const clearAllocations = () => {
		Object.keys(allocations).forEach((k) => delete allocations[k]);
		Object.keys(allocationValid).forEach((k) => delete allocationValid[k]);
	};

	const toggleDevice = (
		node: ComputeAvailabilityNode,
		dev: ComputeAvailabilityDevice
	) => {
		const key = deviceKey(node, dev);
		const idx = selectedKeys.value.indexOf(key);
		if (idx >= 0) {
			selectedKeys.value.splice(idx, 1);
			delete allocations[key];
			delete allocationValid[key];
			return;
		}
		if (!isDeviceSelectable(dev, node)) return;
		selectedKeys.value.push(key);
		addAllocation(key, dev);
	};

	const pickSingleDevice = (
		node: ComputeAvailabilityNode,
		dev: ComputeAvailabilityDevice
	) => {
		const key = deviceKey(node, dev);
		// Re-clicking the currently selected card deselects it (single-select can
		// be cleared), so the user can back out without picking another card.
		if (selectedKeys.value.length === 1 && selectedKeys.value[0] === key) {
			selectedKeys.value = [];
			clearAllocations();
			return;
		}
		if (!isDeviceSelectable(dev, node)) return;
		selectedKeys.value = [key];
		clearAllocations();
		addAllocation(key, dev);
	};

	/**
	 * Restore a previously-attempted binding (the parameters passed last time)
	 * so the user sees what they had. Only devices that are still selectable
	 * are re-checked — the card that the backend rejected (e.g. the
	 * validation-code target) is intentionally left unchecked, with its row
	 * showing the reason it can no longer be used.
	 */
	const initSelection = (items?: ComputeBindingItem[]) => {
		if (!items || items.length === 0) return;
		items.forEach((item) => {
			const node = availability.value.nodes.find(
				(n) => n.nodeName === item.nodeName
			);
			const dev = node?.devices.find((d) => d.deviceId === item.deviceId);
			if (!node || !dev) return;
			if (!isDeviceSelectable(dev, node)) return;
			const key = deviceKey(node, dev);
			if (selectedKeys.value.includes(key)) return;
			if (allowMultiCard.value) {
				selectedKeys.value.push(key);
			} else {
				selectedKeys.value = [key];
			}
			addAllocation(key, dev);
			if (item.memory && allocations[key]) {
				allocations[key].memory = item.memory;
			}
		});
	};

	/**
	 * Bytes a single selected card contributes to the aggregate VRAM shown in
	 * the dialog (header summary + per-card usage bar). Kept here so the summary
	 * and the device list never drift:
	 *   - Exclusive  → the card's free VRAM (at least the requirement).
	 *   - TimeSlice  → the whole card capacity (time-slicing shares the entire
	 *                  card; per-app memory is not meaningful).
	 *   - MemorySlice→ the user's allocation, falling back to the requirement.
	 */
	const cardContributionBytes = (
		node: ComputeAvailabilityNode,
		dev: ComputeAvailabilityDevice
	): number => {
		if (isExclusiveSupportType(dev.supportType)) {
			return Math.max(dev.available, requiredPerCard.value);
		}
		if (isTimeSliceSupportType(dev.supportType)) {
			return dev.capacity;
		}
		const alloc = allocations[deviceKey(node, dev)];
		return alloc ? alloc.memory : requiredPerCard.value;
	};

	/**
	 * Write a MemorySlice card's per-app allocation. Kept on the composable so
	 * the inline quota editor mutates selection state through a method instead of
	 * touching the reactive `allocations` map across a prop boundary.
	 */
	const setAllocationMemory = (
		node: ComputeAvailabilityNode,
		dev: ComputeAvailabilityDevice,
		bytes: number
	) => {
		const key = deviceKey(node, dev);
		const existing = allocations[key];
		if (existing) {
			existing.memory = bytes;
		} else {
			allocations[key] = { memory: bytes, exclusive: false };
		}
	};

	/** Record whether a card's inline quota input currently holds a valid value. */
	const setAllocationValid = (
		node: ComputeAvailabilityNode,
		dev: ComputeAvailabilityDevice,
		valid: boolean
	) => {
		allocationValid[deviceKey(node, dev)] = valid;
	};

	// Submission gate for the inline quota inputs: every selected card must not
	// be explicitly invalid (absence = valid, so Exclusive / TimeSlice pass).
	const allAllocationsValid = computed(() =>
		selectedKeys.value.every((key) => allocationValid[key] !== false)
	);

	const canSubmit = computed(() => selectedKeys.value.length > 0);

	const buildBinding = (): ComputeBindingItem[] =>
		selectedKeys.value.map((key) => {
			const [nodeName, deviceId] = key.split('::');
			const item: ComputeBindingItem = { nodeName, deviceId };
			const alloc = allocations[key];
			if (alloc) item.memory = alloc.memory;
			return item;
		});

	return {
		requirement,
		requiredPerCard,
		allowMultiCard,
		mustSameNode,
		selectedKeys,
		allocations,
		singleSelectedKey,
		deviceKey,
		isDeviceChecked,
		isDeviceSelectable,
		toggleDevice,
		pickSingleDevice,
		cardContributionBytes,
		setAllocationMemory,
		setAllocationValid,
		allAllocationsValid,
		canSubmit,
		buildBinding,
		initSelection
	};
}

/** Selection/submit state machine API, shared by the dialog and its panels. */
export type ComputeBinding = ReturnType<typeof useComputeBinding>;
