export enum DeviceType {
	MOBILE = 'mobile',
	TABLET = 'tablet',
	DESKTOP = 'desktop'
}

let deviceChangeCallbacks: Array<
	(state: { device: DeviceType; isVerticalScreen: boolean }) => void
> = [];

const MOBILE_MAX_WIDTH = 768;
const TABLET_MAX_WIDTH = 1280;

// Word-boundary regexes avoid false positives from broad substrings such as
// `pad` (notepad/gamepad) or `phone`. `surface` is intentionally excluded —
// Surface Laptop/Pro are mostly used as desktops.
// Vendor mobile browsers (UC/QQ/Samsung/Huawei/MIUI/Vivo/OPPO/WeChat etc.)
// already carry `Mobi` or `Mobile` in their UA, so `\bmobi\b` covers them.
const TABLET_UA_REGEX =
	/\b(ipad|tablet|kindle|silk|playbook|nexus 7|nexus 9|nexus 10|galaxy tab|sm-[tp]\d|sch-i800|pixel c|pixel tablet|lenovo (tb|yt)-|matepad)\b|\bharmonyos\b(?!.*\bmobile\b)/;

const MOBILE_UA_REGEX =
	/\b(iphone|ipod|windows phone|iemobile|blackberry|bb10|webos|opera mini|opera mobi|fennec|mqqbrowser|symbian|kaios|mobi)\b|\btizen\b(?!.*\btv\b)/;

export function useDevice(): {
	state: { device: DeviceType; isVerticalScreen: boolean };
	cleanup: () => void;
} {
	let state = getScreenState();
	const mediaOrientation = window.matchMedia('(orientation: landscape)');

	const resizeHandler = () => {
		const newState = getScreenState();
		if (
			newState.device !== state.device ||
			newState.isVerticalScreen !== state.isVerticalScreen
		) {
			state = newState;
			deviceChangeCallbacks.forEach((callback) => {
				try {
					callback(state);
				} catch (error) {
					console.error('Device change callback error:', error);
				}
			});
		}
	};

	window.addEventListener('resize', resizeHandler);
	mediaOrientation.addEventListener('change', resizeHandler);

	const cleanup = () => {
		window.removeEventListener('resize', resizeHandler);
		mediaOrientation.removeEventListener('change', resizeHandler);
		deviceChangeCallbacks = [];
	};

	return { state, cleanup };
}

function getScreenState(): { device: DeviceType; isVerticalScreen: boolean } {
	if (typeof window === 'undefined') {
		return { device: DeviceType.DESKTOP, isVerticalScreen: false };
	}
	const device = detectDevice();
	const isVerticalScreen = window.matchMedia('(orientation: portrait)').matches;
	return { device, isVerticalScreen };
}

export function getInitialDeviceScreenState(): {
	device: DeviceType;
	isVerticalScreen: boolean;
} {
	return getScreenState();
}

function detectDevice(): DeviceType {
	if (typeof window === 'undefined' || typeof navigator === 'undefined') {
		return DeviceType.DESKTOP;
	}

	const ua = navigator.userAgent.toLowerCase();
	const viewportWidth = window.innerWidth;
	const maxTouchPoints = navigator.maxTouchPoints ?? 0;
	const isTouchDevice = 'ontouchstart' in window || maxTouchPoints > 0;

	// iPadOS 13+ ships a desktop Mac UA (`Macintosh`); multi-touch on the Mac
	// platform is the only reliable tell. iPhone/iPod UAs also contain
	// `like Mac OS X`, so match `macintosh` only and exclude them explicitly.
	const isIPadOS =
		/macintosh/.test(ua) && !/iphone|ipod/.test(ua) && maxTouchPoints > 1;

	if (isIPadOS || TABLET_UA_REGEX.test(ua)) {
		return DeviceType.TABLET;
	}

	// Android UA convention: phones include `Mobile`, tablets do not.
	if (/android/.test(ua)) {
		return /mobile/.test(ua) ? DeviceType.MOBILE : DeviceType.TABLET;
	}

	if (MOBILE_UA_REGEX.test(ua)) {
		return DeviceType.MOBILE;
	}

	// Unknown UA but genuine multi-touch device → infer from viewport.
	// Single-point touch (e.g. stylus on a touch laptop) is intentionally excluded.
	if (isTouchDevice && maxTouchPoints > 1) {
		if (viewportWidth < MOBILE_MAX_WIDTH) return DeviceType.MOBILE;
		if (viewportWidth <= TABLET_MAX_WIDTH) return DeviceType.TABLET;
	}

	// Responsive fallback: a desktop browser shrunk below the mobile breakpoint
	// still gets the mobile layout (relied on by `adaptive-body-class`).
	if (viewportWidth < MOBILE_MAX_WIDTH) {
		return DeviceType.MOBILE;
	}

	return DeviceType.DESKTOP;
}

export function onDeviceChange(
	callback: (state: { device: DeviceType; isVerticalScreen: boolean }) => void
) {
	deviceChangeCallbacks.push(callback);
	callback(getScreenState());

	return () => {
		deviceChangeCallbacks = deviceChangeCallbacks.filter(
			(cb) => cb !== callback
		);
	};
}
