import { onUnmounted, ref, Ref } from 'vue';

/**
 * Inline style for the marquee (rubber-band) overlay, positioned relative to
 * the container element passed to `useMarqueeSelection`.
 */
export interface MarqueeRect {
	visible: boolean;
	left: number;
	top: number;
	width: number;
	height: number;
}

export interface UseMarqueeSelectionOptions {
	/**
	 * Element that holds the selectable `.item` nodes. It is also the
	 * positioning context for the overlay, so it must be `position: relative`.
	 */
	container: Ref<HTMLElement | null | undefined>;
	/** Whether marquee selection is currently allowed (e.g. mosaic view on desktop). */
	isEnabled: () => boolean;
	/** Resolve a selectable index from an `.item` element (return null to skip, e.g. headers). */
	resolveIndex: (el: HTMLElement) => number | null;
	/** Apply the resolved indices (already merged with the base selection). */
	applySelection: (indices: number[]) => void;
	/** Selection that exists when the drag starts, kept when extending with a modifier key. */
	getBaseSelection: () => number[];
}

// Pointer travel (px) required before a press is treated as a drag instead of a
// click, so plain clicks keep working.
const DRAG_THRESHOLD = 6;

const emptyRect = (): MarqueeRect => ({
	visible: false,
	left: 0,
	top: 0,
	width: 0,
	height: 0
});

export function useMarqueeSelection(options: UseMarqueeSelectionOptions) {
	const {
		container,
		isEnabled,
		resolveIndex,
		applySelection,
		getBaseSelection
	} = options;

	const marquee = ref<MarqueeRect>(emptyRect());

	let startClientX = 0;
	let startClientY = 0;
	let dragging = false;
	let additive = false;
	let baseSelection: number[] = [];
	// A real drag ends with a synthetic `click`; this flag lets the caller skip
	// the click handler once so it does not clear the freshly made selection.
	let suppressClickOnce = false;

	const updateSelection = (event: MouseEvent) => {
		const el = container.value;
		if (!el) {
			return;
		}
		const bounds = el.getBoundingClientRect();

		const left = Math.min(startClientX, event.clientX);
		const top = Math.min(startClientY, event.clientY);
		const right = Math.max(startClientX, event.clientX);
		const bottom = Math.max(startClientY, event.clientY);

		marquee.value = {
			visible: true,
			left: left - bounds.left,
			top: top - bounds.top,
			width: right - left,
			height: bottom - top
		};

		const indices = new Set<number>(baseSelection);
		el.querySelectorAll<HTMLElement>('.item').forEach((itemEl) => {
			const rect = itemEl.getBoundingClientRect();
			const intersects =
				rect.left < right &&
				rect.right > left &&
				rect.top < bottom &&
				rect.bottom > top;
			if (!intersects) {
				return;
			}
			const index = resolveIndex(itemEl);
			if (index !== null) {
				indices.add(index);
			}
		});

		applySelection([...indices]);
	};

	const onMouseMove = (event: MouseEvent) => {
		if (!dragging) {
			if (
				Math.abs(event.clientX - startClientX) < DRAG_THRESHOLD &&
				Math.abs(event.clientY - startClientY) < DRAG_THRESHOLD
			) {
				return;
			}
			dragging = true;
		}
		// Avoid selecting page text while drawing the marquee.
		event.preventDefault();
		updateSelection(event);
	};

	const onMouseUp = () => {
		window.removeEventListener('mousemove', onMouseMove);
		window.removeEventListener('mouseup', onMouseUp);
		if (dragging) {
			suppressClickOnce = true;
		}
		dragging = false;
		marquee.value = emptyRect();
	};

	const onMouseDown = (event: MouseEvent) => {
		// Every click is preceded by a mousedown, so clearing here also discards
		// any stale suppression left by a drag that ended outside the container.
		suppressClickOnce = false;
		if (!isEnabled() || event.button !== 0) {
			return;
		}
		// Starting on an item is a click/drag-move on that item, not a marquee.
		const target = event.target as HTMLElement | null;
		if (target && target.closest('.item')) {
			return;
		}
		if (!container.value) {
			return;
		}

		startClientX = event.clientX;
		startClientY = event.clientY;
		additive = event.ctrlKey || event.metaKey || event.shiftKey;
		baseSelection = additive ? [...getBaseSelection()] : [];
		dragging = false;

		window.addEventListener('mousemove', onMouseMove);
		window.addEventListener('mouseup', onMouseUp);
	};

	/** Returns true once right after a drag, so the caller can ignore that click. */
	const consumeClickSuppression = () => {
		const suppress = suppressClickOnce;
		suppressClickOnce = false;
		return suppress;
	};

	// If the host component unmounts mid-drag (e.g. route change while drawing
	// the marquee), the window listeners would otherwise stay attached and keep
	// firing on a stale closure / null container until the next mouseup.
	onUnmounted(() => {
		window.removeEventListener('mousemove', onMouseMove);
		window.removeEventListener('mouseup', onMouseUp);
	});

	return { marquee, onMouseDown, consumeClickSuppression };
}
