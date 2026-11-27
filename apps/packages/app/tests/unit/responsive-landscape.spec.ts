/** @jest-environment jsdom */
import { createPinia, setActivePinia } from 'pinia';
import { mount } from '@vue/test-utils';
import { nextTick, reactive } from 'vue';
import DashboardLayout from '../../src/apps/dashboard/layouts/MainLayout.vue';
import { useDeviceStore } from '../../src/stores/settings/device';

const mockDisposers: Array<() => void> = [];
jest.mock('@bytetrade/core', () => {
	const resize = jest.requireActual('../../../core/src/resize');
	return {
		...resize,
		...jest.requireActual('../../../core/src/browserViewport'),
		useDevice: () => {
			const session = resize.useDevice();
			mockDisposers.push(session.cleanup);
			return session;
		}
	};
});
jest.mock('quasar', () => ({ useQuasar: () => mockQuasar }));
jest.mock('../../src/apps/dashboard/layouts/LayoutPc.vue', () => ({
	template: '<div data-layout="wide" />'
}));
jest.mock('../../src/apps/dashboard/layouts/LayoutMobile.vue', () => ({
	template: '<div data-layout="compact" />'
}));
jest.mock('../../src/components/LandscapeScreenIndex.vue', () => ({
	template: '<div data-landscape-overlay />'
}));

const mockQuasar = reactive({
	platform: { is: { mobile: true, nativeMobile: false } },
	screen: { width: 844, height: 390 }
});
let orientation: EventTarget & { type: string };
const originals = [
	[window, 'innerWidth'],
	[window, 'innerHeight'],
	[window, 'matchMedia'],
	[window, 'visualViewport'],
	[window, 'orientation'],
	[window.screen, 'orientation'],
	[navigator, 'userAgent'],
	[navigator, 'maxTouchPoints']
].map(([target, key]) => ({
	target: target as object,
	key: key as string,
	descriptor: Object.getOwnPropertyDescriptor(target, key as string)
}));

function set(target: object, key: string, value: unknown) {
	Object.defineProperty(target, key, { configurable: true, value });
}

function resize(width: number, height: number) {
	set(window, 'innerWidth', width);
	set(window, 'innerHeight', height);
	Object.assign(mockQuasar.screen, { width, height });
	orientation.type = height > width ? 'portrait-primary' : 'landscape-primary';
	window.dispatchEvent(new Event('resize'));
	orientation.dispatchEvent(new Event('change'));
}

beforeEach(() => {
	setActivePinia(createPinia());
	Object.assign(mockQuasar.platform.is, { mobile: true, nativeMobile: false });
	set(navigator, 'userAgent', 'Mozilla/5.0 (Linux; Android 14) Mobile');
	set(navigator, 'maxTouchPoints', 5);
	set(window, 'visualViewport', undefined);
	set(window, 'orientation', undefined);
	orientation = Object.assign(new EventTarget(), { type: 'landscape-primary' });
	set(window.screen, 'orientation', orientation);
	set(window, 'matchMedia', (query: string) => ({
		get matches() {
			return query === '(orientation: portrait)'
				? innerHeight > innerWidth
				: innerWidth >= innerHeight;
		},
		addEventListener: orientation.addEventListener.bind(orientation),
		removeEventListener: orientation.removeEventListener.bind(orientation)
	}));
	resize(844, 390);
});

afterEach(() => {
	for (const stop of mockDisposers.splice(0)) stop();
	for (const { target, key, descriptor } of originals) {
		if (descriptor) Object.defineProperty(target, key, descriptor);
		else Reflect.deleteProperty(target, key);
	}
});

it.each([
	['Android phone', 'Mozilla/5.0 (Linux; Android 14) Mobile'],
	['Android tablet', 'Mozilla/5.0 (Linux; Android 14)'],
	['iPhone', 'Mozilla/5.0 (iPhone; CPU iPhone OS 18_0 like Mac OS X) Mobile']
])(
	'keeps the wide layout usable while rotating or resizing an %s',
	async (_, ua) => {
		set(navigator, 'userAgent', ua);
		const store = useDeviceStore();
		store.init();
		const wrapper = mount(DashboardLayout);
		try {
			for (const [width, height, compact, blocked] of [
				[844, 390, false, false],
				[768, 390, false, false],
				[767, 390, true, true],
				[390, 844, true, false],
				[844, 390, false, false]
			] as const) {
				resize(width, height);
				await nextTick();
				expect(store.isCompactLayout).toBe(compact);
				expect(wrapper.find('[data-layout="wide"]').exists()).toBe(!compact);
				expect(wrapper.find('[data-layout="compact"]').exists()).toBe(compact);
				expect(wrapper.find('[data-landscape-overlay]').exists()).toBe(blocked);
			}
		} finally {
			wrapper.unmount();
		}
	}
);

it.each([
	['expanded foldable', 1024, 735, false, true],
	['native phone', 844, 390, true, true],
	['wide desktop window', 844, 390, false, false]
])(
	'keeps an %s usable without an unnecessary overlay',
	(_name, width, height, native, mobile) => {
		mockQuasar.platform.is.nativeMobile = native;
		mockQuasar.platform.is.mobile = mobile;
		if (!mobile) {
			set(
				navigator,
				'userAgent',
				'Mozilla/5.0 (X11; Linux x86_64) Chrome/156.0.0.0'
			);
			set(navigator, 'maxTouchPoints', 0);
		}
		resize(width, height);
		useDeviceStore().init();
		const wrapper = mount(DashboardLayout);
		expect(wrapper.find('[data-layout="wide"]').exists()).toBe(true);
		expect(wrapper.find('[data-landscape-overlay]').exists()).toBe(false);
		wrapper.unmount();
	}
);
