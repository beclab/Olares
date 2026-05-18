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

	function resetSlideForSearch() {
		if (!slideRef || slideRef.value === 0) return;
		skipBlurOnCarouselSlideChange = true;
		carouselAnimated.value = false;
		slideRef.value = 0;
		void nextTick(() => {
			carouselAnimated.value = true;
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
		isFocus.value = true;
	}

	function blurSearch() {
		isFocus.value = false;
	}

	function updateSearch(val: string | number | null) {
		appStore.update_search_result(String(val ?? ''));
		resetSlideForSearch();
	}

	function cleanSearchVal() {
		searchVal.value = '';
		appStore.dismiss_search_result();
		resetSlideForSearch();
	}

	return {
		searchVal,
		isFocus,
		carouselAnimated,
		focusSearch,
		blurSearch,
		updateSearch,
		cleanSearchVal
	};
}
