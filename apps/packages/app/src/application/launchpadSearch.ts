import { ref, nextTick, watch, type Ref } from 'vue';
import { useApplicationStore } from 'src/stores/desktop/app';

type BlurrableInput = { blur?: () => void } | null | undefined;

export function useLaunchpadSearch(
	slideRef?: Ref<number>,
	searchInputRef?: Ref<BlurrableInput>
) {
	const appStore = useApplicationStore();
	const searchVal = ref('');
	const isFocus = ref(false);
	const carouselAnimated = ref(true);
	let skipBlurOnCarouselSlideChange = false;
	let slideBeforeSearch = 0;
	let pendingSlideAfterSearchDrag: number | null = null;
	let carouselPausedForSearchDrag = false;
	let suppressSlideBeforeSearchCapture = false;

	function captureSlideBeforeSearch() {
		if (!slideRef) return;
		slideBeforeSearch = slideRef.value;
	}

	function resetSlideForSearch() {
		if (!slideRef) return;
		if (slideRef.value !== 0) {
			captureSlideBeforeSearch();
			skipBlurOnCarouselSlideChange = true;
			carouselAnimated.value = false;
			slideRef.value = 0;
			void nextTick(() => {
				carouselAnimated.value = true;
			});
		}
	}

	if (slideRef) {
		watch(slideRef, (v) => {
			if (suppressSlideBeforeSearchCapture) return;
			if (appStore.launchpadSearchQuery.length === 0) {
				slideBeforeSearch = v;
			}
		});
	}

	if (slideRef && searchInputRef) {
		watch(slideRef, () => {
			if (skipBlurOnCarouselSlideChange) {
				skipBlurOnCarouselSlideChange = false;
				return;
			}
			void nextTick(() => {
				requestAnimationFrame(() => {
					searchInputRef.value?.blur?.();
				});
			});
		});
	}

	function focusSearch() {
		if (appStore.launchpadSearchQuery.length === 0) {
			captureSlideBeforeSearch();
		}
		isFocus.value = true;
	}

	function blurSearch() {
		isFocus.value = false;
	}

	function updateSearch(val: string | number | null) {
		const next = String(val ?? '');
		if (next.length > 0 && appStore.launchpadSearchQuery.length === 0) {
			captureSlideBeforeSearch();
		}
		appStore.update_search_result(next);
		resetSlideForSearch();
	}

	function cleanSearchVal() {
		searchVal.value = '';
		appStore.dismiss_search_result();
		resetSlideForSearch();
	}

	function exitSearchForDrag(_draggedAppId?: string) {
		if (!appStore.launchpadSearchQuery.length) return;
		const viewerSlideBeforeSearch = slideBeforeSearch;
		searchVal.value = '';
		suppressSlideBeforeSearchCapture = true;
		appStore.dismiss_search_result();
		suppressSlideBeforeSearchCapture = false;
		if (!slideRef) return;

		const maxSlide = Math.max(0, appStore.launchPadApps.length - 1);
		const target = Math.min(Math.max(0, viewerSlideBeforeSearch), maxSlide);
		pendingSlideAfterSearchDrag = target;
	}

	function applyPendingSlideRestoreAfterDrag(): number | null {
		if (!slideRef || pendingSlideAfterSearchDrag === null) return null;
		const target = pendingSlideAfterSearchDrag;
		pendingSlideAfterSearchDrag = null;
		skipBlurOnCarouselSlideChange = true;
		slideRef.value = target;
		carouselPausedForSearchDrag = true;
		return target;
	}

	function clearPendingSlideAfterSearchDrag() {
		pendingSlideAfterSearchDrag = null;
	}

	function resumeCarouselAfterSearchDrag() {
		if (!carouselPausedForSearchDrag) return;
		carouselPausedForSearchDrag = false;
		void nextTick(() => {
			carouselAnimated.value = true;
		});
	}

	return {
		searchVal,
		isFocus,
		carouselAnimated,
		focusSearch,
		blurSearch,
		updateSearch,
		cleanSearchVal,
		exitSearchForDrag,
		applyPendingSlideRestoreAfterDrag,
		clearPendingSlideAfterSearchDrag,
		resumeCarouselAfterSearchDrag
	};
}
