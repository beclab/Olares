import { ref, onMounted, onBeforeUnmount } from 'vue';

const KEYBOARD_THRESHOLD = 150;

type Orientation = 'portrait' | 'landscape';

export function useStableViewportHeight() {
	const stableHeight = ref(0);
	const keyboardOpen = ref(false);

	let orientationTimer: ReturnType<typeof setTimeout> | null = null;
	let lastOrientation: Orientation | '' = '';

	function currentViewportHeight(): number {
		return window.visualViewport?.height ?? window.innerHeight;
	}

	function getOrientation(): Orientation {
		return window.matchMedia('(orientation: portrait)').matches
			? 'portrait'
			: 'landscape';
	}

	function writeCssVar() {
		document.documentElement.style.setProperty(
			'--stable-vh',
			`${stableHeight.value}px`
		);
	}

	function resyncForOrientation() {
		const h = currentViewportHeight();
		stableHeight.value = h;
		keyboardOpen.value = false;
		writeCssVar();
	}

	function updateStableHeight() {
		const h = currentViewportHeight();
		if (h > stableHeight.value || stableHeight.value === 0) {
			stableHeight.value = h;
			writeCssVar();
		}
	}

	function checkKeyboard() {
		if (stableHeight.value <= 0) {
			keyboardOpen.value = false;
			return;
		}
		const orientation = getOrientation();
		if (orientation !== lastOrientation) {
			lastOrientation = orientation;
			resyncForOrientation();
			return;
		}
		keyboardOpen.value =
			stableHeight.value - currentViewportHeight() > KEYBOARD_THRESHOLD;
	}

	function onViewportResize() {
		checkKeyboard();
		if (!keyboardOpen.value) {
			updateStableHeight();
		}
	}

	function onOrientationChange() {
		if (orientationTimer) clearTimeout(orientationTimer);
		orientationTimer = setTimeout(() => {
			orientationTimer = null;
			const orientation = getOrientation();
			if (orientation !== lastOrientation) {
				lastOrientation = orientation;
				resyncForOrientation();
			} else {
				stableHeight.value = 0;
				updateStableHeight();
			}
		}, 200);
	}

	onMounted(() => {
		lastOrientation = getOrientation();
		stableHeight.value = currentViewportHeight();
		writeCssVar();
		window.visualViewport?.addEventListener('resize', onViewportResize);
		window.addEventListener('orientationchange', onOrientationChange);
	});

	onBeforeUnmount(() => {
		window.visualViewport?.removeEventListener('resize', onViewportResize);
		window.removeEventListener('orientationchange', onOrientationChange);
		if (orientationTimer) clearTimeout(orientationTimer);
	});

	return { stableHeight, keyboardOpen };
}
