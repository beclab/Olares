/** @jest-environment jsdom */
import { createPinia, setActivePinia } from 'pinia';
import { mount } from '@vue/test-utils';
import { nextTick } from 'vue';
import { DeviceType, onDeviceChange } from '../../../core/src/resize';
import { useDeviceStore } from '../../src/stores/settings/device';
import AdaptiveLayout from '../../src/components/settings/AdaptiveLayout.vue';
import adaptiveBodyClass from '../../src/boot/adaptive-body-class';

const mockDisposers: Array<() => void> = [];
jest.mock('@bytetrade/core', () => {
	const resize = jest.requireActual('../../../core/src/resize');
	return {
		...resize,
		useDevice: () => {
			const session = resize.useDevice();
			mockDisposers.push(session.cleanup);
			return session;
		}
	};
});
jest.mock('quasar/wrappers', () => ({ boot: (callback: unknown) => callback }));

// Captured from the same 720 CSS-pixel tablet in Chrome and LarePass.
const CHROME_TABLET_UA =
	'Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/156.0.0.0 Safari/537.36';
const WEBVIEW_TABLET_UA =
	'Mozilla/5.0 (Linux; Android 14; 23046RP50C Build/UKQ1.230804.001; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/155.0.8059.43 Safari/537.36';
const FOLDED_PHONE_UA =
	'Mozilla/5.0 (Linux; Android 10; K) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/154.0.0.0 Safari/537.36';

let orientation: EventTarget;
let width: number;
let portrait: boolean;
const originals = [
	[window, 'innerWidth'],
	[window, 'matchMedia'],
	[navigator, 'userAgent'],
	[navigator, 'maxTouchPoints']
].map(([target, key]) => ({
	target: target as object,
	key: key as string,
	descriptor: Object.getOwnPropertyDescriptor(target, key as string)
}));

function browser(ua: string, viewportWidth: number, touch = 5) {
	width = viewportWidth;
	Object.defineProperty(navigator, 'userAgent', {
		configurable: true,
		value: ua
	});
	Object.defineProperty(navigator, 'maxTouchPoints', {
		configurable: true,
		value: touch
	});
}

beforeEach(() => {
	setActivePinia(createPinia());
	width = 720;
	portrait = true;
	orientation = new EventTarget();
	Object.defineProperty(window, 'innerWidth', {
		configurable: true,
		get: () => width
	});
	Object.defineProperty(window, 'matchMedia', {
		configurable: true,
		value: (query: string) => ({
			get matches() {
				return query === '(orientation: portrait)' ? portrait : !portrait;
			},
			addEventListener: orientation.addEventListener.bind(orientation),
			removeEventListener: orientation.removeEventListener.bind(orientation)
		})
	});
});

afterEach(() => {
	for (const stop of mockDisposers.splice(0)) stop();
	for (const { target, key, descriptor } of originals) {
		if (descriptor) Object.defineProperty(target, key, descriptor);
		else Reflect.deleteProperty(target, key);
	}
	document.body.className = '';
});

it.each([
	['Chrome tablet', CHROME_TABLET_UA, 5],
	['LarePass tablet WebView', WEBVIEW_TABLET_UA, 5],
	['folded phone without Mobile in its UA', FOLDED_PHONE_UA, 5],
	[
		'iPad desktop UA',
		'Mozilla/5.0 (Macintosh; Intel Mac OS X) Safari/605.1.15',
		5
	],
	['phone with Mobile in its UA', 'Mozilla/5.0 (Linux; Android 14) Mobile', 5],
	['desktop browser', CHROME_TABLET_UA, 0]
])(
	'uses viewport width for %s, including the breakpoint boundary',
	(_name, ua, touch) => {
		for (const [viewportWidth, compact] of [
			[411, true],
			[720, true],
			[767, true],
			[768, false],
			[1024, false]
		] as const) {
			browser(ua, viewportWidth, touch);
			const store = useDeviceStore(createPinia());
			expect(store.isCompactLayout).toBe(compact);
			expect(store.isMobile).toBe(compact);
			expect(store.platform).toBe(compact ? 'mobile' : 'web');
		}
	}
);

it('keeps the tablet device identity while rendering the compact slot on the first frame', () => {
	browser(WEBVIEW_TABLET_UA, 720);
	const store = useDeviceStore();
	expect(store.deviceInfo.device).toBe(DeviceType.TABLET);
	const wrapper = mount(AdaptiveLayout, {
		slots: { mobile: 'compact contents', pc: 'wide contents' }
	});
	expect(wrapper.text()).toBe('compact contents');
	wrapper.unmount();
});

it('updates the shell, compatibility getter, body classes and callback when only layout crosses the breakpoint', async () => {
	browser(WEBVIEW_TABLET_UA, 720);
	const store = useDeviceStore();
	const change = jest.fn();
	store.init(change);
	(adaptiveBodyClass as unknown as () => void)();
	const wrapper = mount(AdaptiveLayout, {
		slots: { mobile: 'compact contents', pc: 'wide contents' }
	});
	expect(document.body.classList.contains('adaptive-is-mobile')).toBe(true);
	change.mockClear();

	width = 768;
	window.dispatchEvent(new Event('resize'));
	await nextTick();
	expect(change).toHaveBeenCalledTimes(1);
	expect(store.deviceInfo.device).toBe(DeviceType.TABLET);
	expect(store.deviceInfo.isVerticalScreen).toBe(true);
	expect(store.isMobile).toBe(false);
	expect(wrapper.text()).toBe('wide contents');
	expect(document.body.classList.contains('adaptive-pc')).toBe(true);
	expect(document.body.classList.contains('adaptive-is-mobile')).toBe(false);

	// Expanding and returning to split-screen need no UA/orientation change.
	width = 720;
	window.dispatchEvent(new Event('resize'));
	await nextTick();
	expect(change).toHaveBeenCalledTimes(2);
	expect(wrapper.text()).toBe('compact contents');
	expect(document.body.classList.contains('adaptive-pc')).toBe(false);
	wrapper.unmount();
});

it('updates layout after rotation without changing the phone device identity', () => {
	browser('Mozilla/5.0 (Linux; Android 14) Mobile', 411);
	const store = useDeviceStore();
	store.init();
	width = 900;
	portrait = false;
	orientation.dispatchEvent(new Event('change'));
	expect(store.deviceInfo).toEqual({
		device: DeviceType.MOBILE,
		isVerticalScreen: false,
		isCompactLayout: false
	});
});

it('does not notify on height-only resize or width changes within the same layout', () => {
	browser(WEBVIEW_TABLET_UA, 720);
	useDeviceStore().init();
	const change = jest.fn();
	onDeviceChange(change);
	change.mockClear();
	window.dispatchEvent(new Event('resize'));
	width = 721;
	window.dispatchEvent(new Event('resize'));
	expect(change).not.toHaveBeenCalled();
});
