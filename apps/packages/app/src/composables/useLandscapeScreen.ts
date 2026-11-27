import { computed, ComputedRef } from 'vue';
import { useQuasar } from 'quasar';
import { useDeviceStore } from 'src/stores/settings/device';

/**
 * Shared visibility logic for the landscape-orientation block overlay.
 *
 * Only browser builds need to block landscape; native apps lock orientation
 * themselves. `isVerticalScreen` is kept in sync with the orientation media
 * query, so the overlay reacts to live rotation.
 */
export function useLandscapeScreen(): {
	showLandscapeBlock: ComputedRef<boolean>;
} {
	const $q = useQuasar();
	const deviceStore = useDeviceStore();

	const showLandscapeBlock = computed(
		() =>
			!$q.platform.is.nativeMobile &&
			deviceStore.isMobile &&
			!deviceStore.deviceInfo.isVerticalScreen
	);

	return { showLandscapeBlock };
}
