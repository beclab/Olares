/**
 * Reading Progress Composable
 *
 * Manages reading progress tracking with IntersectionObserver
 * Tracks visible elements and updates reading indicator
 */

import { ref, onBeforeUnmount, Ref } from 'vue';

export interface ReadingIndicatorStyle {
	transform: string;
	height: string;
}

export interface UseReadingProgressOptions {
	/** Debounce delay for indicator updates (ms) */
	debounceDelay?: number;
	/** IntersectionObserver threshold */
	threshold?: number;
	/** Selectors for elements to track */
	selectors?: string;
}

const DEFAULT_OPTIONS: Required<UseReadingProgressOptions> = {
	debounceDelay: 100,
	threshold: 0.1,
	selectors:
		'section, p, h1, h2, h3, h4, h5, h6, img, ul, ol, blockquote, pre, table'
};

export function useReadingProgress(
	containerRef: Ref<HTMLElement | null>,
	options: UseReadingProgressOptions = {}
) {
	const opts = { ...DEFAULT_OPTIONS, ...options };

	// Reading indicator style
	const indicatorStyle = ref<ReadingIndicatorStyle>({
		transform: 'translateY(0px)',
		height: '0px'
	});

	// Currently selected element
	const selectedElement = ref<HTMLElement | null>(null);
	const clickedElement = ref<HTMLElement | null>(null);

	// IntersectionObserver instance
	let observer: IntersectionObserver | null = null;
	let debounceTimer: ReturnType<typeof setTimeout> | null = null;

	/**
	 * Update indicator position based on element
	 */
	function updateIndicator(element: HTMLElement) {
		if (debounceTimer) {
			clearTimeout(debounceTimer);
		}

		debounceTimer = setTimeout(() => {
			if (element.offsetTop === 0 && element.offsetHeight === 0) {
				return;
			}

			// Add offset for header (if present)
			const additionalOffset = element.offsetTop === 0 ? 0 : 101 + 20;

			indicatorStyle.value = {
				transform: `translateY(${element.offsetTop + additionalOffset}px)`,
				height: `${element.offsetHeight}px`
			};
		}, opts.debounceDelay);
	}

	/**
	 * Handle element click
	 */
	function handleElementClick(e: MouseEvent) {
		clickedElement.value = e.target as HTMLElement;
		updateIndicator(clickedElement.value);
	}

	/**
	 * Start observing elements for visibility
	 */
	function startObserving() {
		if (!containerRef.value) return;

		// Cleanup existing observer
		stopObserving();

		// Create new observer
		observer = new IntersectionObserver(
			(entries) => {
				const visibleEntries = entries
					.filter((entry) => entry.isIntersecting)
					.sort((a, b) => a.boundingClientRect.top - b.boundingClientRect.top);

				if (visibleEntries.length > 0) {
					selectedElement.value = visibleEntries[0].target as HTMLElement;

					if (clickedElement.value) {
						clickedElement.value = null;
					}

					updateIndicator(selectedElement.value);
				}
			},
			{ threshold: opts.threshold }
		);

		// Find and observe target elements
		const targetElements = containerRef.value.querySelectorAll(opts.selectors);
		targetElements.forEach((element) => {
			observer?.observe(element);
		});
	}

	/**
	 * Stop observing
	 */
	function stopObserving() {
		if (observer) {
			observer.disconnect();
			observer = null;
		}

		if (debounceTimer) {
			clearTimeout(debounceTimer);
			debounceTimer = null;
		}
	}

	/**
	 * Reset indicator
	 */
	function resetIndicator() {
		indicatorStyle.value = {
			transform: 'translateY(0px)',
			height: '0px'
		};
		selectedElement.value = null;
		clickedElement.value = null;
	}

	// Lifecycle hooks
	onBeforeUnmount(() => {
		stopObserving();
	});

	return {
		/** Reading indicator style */
		indicatorStyle,
		/** Currently selected element */
		selectedElement,
		/** Update indicator for specific element */
		updateIndicator,
		/** Handle element click */
		handleElementClick,
		/** Start observing elements */
		startObserving,
		/** Stop observing */
		stopObserving,
		/** Reset indicator */
		resetIndicator
	};
}

/**
 * Title Observer Composable
 *
 * Tracks visible headings for table of contents highlighting
 */
export function useTitleObserver(
	containerRef: Ref<HTMLElement | null>,
	onTitleVisible: (id: string) => void
) {
	let observer: IntersectionObserver | null = null;

	/**
	 * Start observing title elements
	 */
	function startObserving() {
		if (!containerRef.value) return;

		stopObserving();

		observer = new IntersectionObserver(
			(entries) => {
				entries.forEach((entry) => {
					if (entry.isIntersecting) {
						onTitleVisible(entry.target.id);
					}
				});
			},
			{
				rootMargin: '0px 0px -90% 0px',
				threshold: 0
			}
		);

		const headings = containerRef.value.querySelectorAll(
			'h1, h2, h3, h4, h5, h6'
		);
		headings.forEach((heading) => {
			if (heading.id) {
				observer?.observe(heading);
			}
		});
	}

	/**
	 * Stop observing
	 */
	function stopObserving() {
		if (observer) {
			observer.disconnect();
			observer = null;
		}
	}

	onBeforeUnmount(() => {
		stopObserving();
	});

	return {
		startObserving,
		stopObserving
	};
}
