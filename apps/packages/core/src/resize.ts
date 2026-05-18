export enum DeviceType {
	MOBILE = 'mobile',
	TABLET = 'tablet',
	DESKTOP = 'desktop'
}

let deviceChangeCallbacks: Array<
	(state: { device: DeviceType; isVerticalScreen: boolean }) => void
> = [];

const MOBILE_KEYWORDS = [
	'phone',
	'pad',
	'pod',
	'iphone',
	'ipod',
	'ios',
	'ipad',
	'android',
	'mobile',
	'blackberry',
	'iemobile',
	'mqqbrowser',
	'juc',
	'fennec',
	'wosbrowser',
	'browserng',
	'webos',
	'symbian',
	'windows phone'
];

const TABLET_KEYWORDS = [
	'ipad',
	'tablet',
	'kindle',
	'galaxy tab',
	'surface',
	'nexus 7',
	'nexus 10'
];

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
	const device = detectDevice();
	const isVerticalScreen = window.matchMedia('(orientation: portrait)').matches;
	return { device, isVerticalScreen };
}

export function getInitialDeviceScreenState(): {
	device: DeviceType;
	isVerticalScreen: boolean;
} {
	if (typeof window === 'undefined') {
		return { device: DeviceType.DESKTOP, isVerticalScreen: false };
	}
	return getScreenState();
}

function detectDevice(): DeviceType {
	const userAgent = navigator.userAgent.toLowerCase();
	const screenWidth = window.screen.width;
	const viewportWidth = window.innerWidth;

	if (TABLET_KEYWORDS.some((keyword) => userAgent.includes(keyword))) {
		return DeviceType.TABLET;
	}

	if (MOBILE_KEYWORDS.some((keyword) => userAgent.includes(keyword))) {
		return screenWidth > 768 ? DeviceType.TABLET : DeviceType.MOBILE;
	}

	const isTouchDevice =
		'ontouchstart' in window || navigator.maxTouchPoints > 0;

	if (viewportWidth < 768) {
		return DeviceType.MOBILE;
	}

	if (isTouchDevice && viewportWidth >= 768 && viewportWidth <= 1280) {
		return DeviceType.TABLET;
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
