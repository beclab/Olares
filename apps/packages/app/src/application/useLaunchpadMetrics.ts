import { ref, computed, onMounted, nextTick, type Ref } from 'vue';
import { debounce } from 'lodash';
import { useApplicationStore } from 'src/stores/desktop/app';

export function useLaunchpadMetrics(
	launchpadPage: Ref<HTMLElement | undefined>,
	appStore: ReturnType<typeof useApplicationStore>,
	carouselAnimated: Ref<boolean>
) {
	function applyLaunchpadMetrics() {
		const el = launchpadPage.value;
		if (!el) return;
		const offsetWidth = el.offsetWidth;
		const offsetHeight = el.offsetHeight;
		if (offsetWidth < 8 || offsetHeight < 8) return;

		let grid_x: number = Math.floor(offsetWidth / 20);
		const grid_y: number = Math.floor(offsetHeight / 11.5);
		grid_x = grid_x < grid_y ? grid_x : grid_y;

		appStore.DESKTOP_APP_SIZE = grid_x;
		appStore.DESKTOP_APP_X_GAP = grid_x * 1.2;
		appStore.DESKTOP_APP_Y_GAP = grid_x * 1 + 16;

		appStore.DOCKER_APP_START_Y_GAP =
			Math.floor(
				offsetHeight -
					appStore.DESKTOP_APP_Y_NUM * grid_x -
					(appStore.DESKTOP_APP_Y_NUM - 1) * appStore.DESKTOP_APP_Y_GAP
			) * 0.15;

		appStore.DOCKER_APP_START_X_GAP =
			Math.floor(
				offsetWidth -
					appStore.DESKTOP_APP_X_NUM * grid_x -
					(appStore.DESKTOP_APP_X_NUM - 1) * appStore.DESKTOP_APP_X_GAP
			) * 0.5;
		appStore.DOCKER_APP_END_X_GAP = appStore.DOCKER_APP_START_X_GAP;

		const search = appStore.launchpadSearchQuery;
		if (search.length > 0) {
			void appStore.update_search_result(search);
		} else {
			appStore.resize();
		}
	}

	const debouncedApplyLaunchpadMetrics = debounce(applyLaunchpadMetrics, 120);

	const LP_REORDER_MS = 180;
	const iconPosTransitionEnabled = ref(false);
	const iconPosTransitionMs = computed(() =>
		iconPosTransitionEnabled.value ? LP_REORDER_MS : 0
	);

	onMounted(() => {
		applyLaunchpadMetrics();
		window.addEventListener('resize', debouncedApplyLaunchpadMetrics);
		void nextTick(() => {
			requestAnimationFrame(() => {
				iconPosTransitionEnabled.value = true;
				carouselAnimated.value = true;
			});
		});
	});

	function dispose() {
		debouncedApplyLaunchpadMetrics.cancel();
		window.removeEventListener('resize', debouncedApplyLaunchpadMetrics);
	}

	return { iconPosTransitionMs, dispose };
}
