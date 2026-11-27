/** @jest-environment jsdom */
import {
	getBrowserViewportState,
	observeBrowserViewport,
	readScreenDirection,
	type BrowserViewportListener,
	type BrowserViewportState
} from '../../../core/src/browserViewport';
import {
	DeviceType,
	getInitialDeviceScreenState
} from '../../../core/src/resize';

let visual: EventTarget & { height: number; scale: number };
let orientation: EventTarget & { type: string };
const subscriptions = new Set<() => void>();
const originalDescriptors = [
	[window, 'innerWidth'],
	[window, 'innerHeight'],
	[window, 'visualViewport'],
	[window, 'orientation'],
	[window, 'matchMedia'],
	[window.screen, 'orientation'],
	[navigator, 'userAgent'],
	[navigator, 'maxTouchPoints']
].map(([target, name]) => ({
	target: target as object,
	name: name as string,
	descriptor: Object.getOwnPropertyDescriptor(target, name as string)
}));

function set(target: object, key: string, value: unknown) {
	Object.defineProperty(target, key, { configurable: true, value });
}
function geometry(
	width: number,
	layoutHeight: number,
	visualHeight = layoutHeight
) {
	set(window, 'innerWidth', width);
	set(window, 'innerHeight', layoutHeight);
	visual.height = visualHeight;
}
function resize(
	width: number,
	layoutHeight: number,
	visualHeight = layoutHeight
) {
	geometry(width, layoutHeight, visualHeight);
	window.dispatchEvent(new Event('resize'));
}
function subscribe(listener: BrowserViewportListener) {
	const unsubscribe = observeBrowserViewport(listener);
	subscriptions.add(unsubscribe);
	return () => {
		unsubscribe();
		subscriptions.delete(unsubscribe);
	};
}
function focusInput() {
	(document.querySelector('input') as HTMLInputElement).focus();
}

beforeEach(() => {
	visual = Object.assign(new EventTarget(), { height: 850, scale: 1 });
	orientation = Object.assign(new EventTarget(), { type: 'portrait-primary' });
	set(window, 'visualViewport', visual);
	set(window.screen, 'orientation', orientation);
	set(window, 'orientation', undefined);
	set(navigator, 'maxTouchPoints', 0);
	set(
		navigator,
		'userAgent',
		'Mozilla/5.0 (iPhone; CPU iPhone OS 18_0 like Mac OS X) Mobile'
	);
	set(window, 'matchMedia', (query: string) => ({
		matches: query.includes('portrait')
			? window.innerHeight >= window.innerWidth
			: window.innerWidth > window.innerHeight,
		media: query
	}));
	geometry(390, 850);
	document.body.innerHTML = '<input type="password"><button>Other</button>';
});
afterEach(() => {
	subscriptions.forEach((unsubscribe) => unsubscribe());
	subscriptions.clear();
	document.body.innerHTML = '';
	jest.restoreAllMocks();
});
afterAll(() => {
	originalDescriptors.forEach(({ target, name, descriptor }) => {
		if (descriptor) Object.defineProperty(target, name, descriptor);
		else Reflect.deleteProperty(target, name);
	});
});

it('reads a snapshot without installing listeners and separates screen from viewport direction', () => {
	geometry(390, 320);
	const addWindow = jest.spyOn(window, 'addEventListener');
	const addDocument = jest.spyOn(document, 'addEventListener');
	expect(getBrowserViewportState()).toMatchObject({
		isMobileBrowser: true,
		screenDirection: 'portrait',
		viewportDirection: 'landscape',
		layoutWidth: 390,
		layoutHeight: 320,
		availableHeight: 320,
		keyboardOpen: false
	});
	expect(addWindow).not.toHaveBeenCalled();
	expect(addDocument).not.toHaveBeenCalled();
});

it('preserves legacy device classification and viewport direction for a narrow desktop window', () => {
	set(
		navigator,
		'userAgent',
		'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)'
	);
	orientation.type = 'landscape-primary';
	geometry(390, 500);
	expect(getInitialDeviceScreenState()).toEqual({
		device: DeviceType.MOBILE,
		isVerticalScreen: true,
		isCompactLayout: true
	});
	expect(getBrowserViewportState()).toMatchObject({
		isMobileBrowser: false,
		screenDirection: 'landscape',
		viewportDirection: 'portrait'
	});
});

it('shares keyboard history and one listener set until the last subscriber leaves', () => {
	const addWindow = jest.spyOn(window, 'addEventListener');
	const removeWindow = jest.spyOn(window, 'removeEventListener');
	const addDocument = jest.spyOn(document, 'addEventListener');
	const removeDocument = jest.spyOn(document, 'removeEventListener');
	const addVisual = jest.spyOn(visual, 'addEventListener');
	const removeVisual = jest.spyOn(visual, 'removeEventListener');
	const addOrientation = jest.spyOn(orientation, 'addEventListener');
	const removeOrientation = jest.spyOn(orientation, 'removeEventListener');
	const first = jest.fn();
	const stopFirst = subscribe(first);
	focusInput();
	resize(390, 850, 320);
	expect(getBrowserViewportState().keyboardOpen).toBe(true);
	const second = jest.fn();
	const stopSecond = subscribe(second);
	expect(second).toHaveBeenLastCalledWith(
		expect.objectContaining({ keyboardOpen: true }),
		'initial'
	);
	expect(
		addWindow.mock.calls.filter(([name]) => name === 'resize')
	).toHaveLength(1);
	expect(
		addDocument.mock.calls.filter(([name]) => name === 'focusin')
	).toHaveLength(1);
	expect(addVisual).toHaveBeenCalledTimes(1);
	expect(addOrientation).toHaveBeenCalledTimes(1);
	stopFirst();
	stopFirst();
	expect(removeWindow).not.toHaveBeenCalled();
	expect(removeDocument).not.toHaveBeenCalled();
	resize(390, 850);
	expect(second).toHaveBeenLastCalledWith(
		expect.objectContaining({ keyboardOpen: false, availableHeight: 850 }),
		'resize'
	);
	stopSecond();
	expect(
		removeWindow.mock.calls.filter(([name]) => name === 'resize')
	).toHaveLength(1);
	expect(
		removeDocument.mock.calls.filter(([name]) => name === 'focusin')
	).toHaveLength(1);
	expect(removeVisual).toHaveBeenCalledTimes(1);
	expect(removeOrientation).toHaveBeenCalledTimes(1);
});

it('gives repeated registrations of the same callback independent subscriptions', () => {
	const listener = jest.fn();
	const stopFirst = subscribe(listener);
	const stopSecond = subscribe(listener);
	expect(listener).toHaveBeenCalledTimes(2);
	resize(390, 800);
	expect(listener).toHaveBeenCalledTimes(4);
	stopFirst();
	resize(390, 750);
	expect(listener).toHaveBeenCalledTimes(5);
	stopSecond();
});

it('publishes keyboard and geometry together without mutating previous snapshots', () => {
	const listener = jest.fn();
	subscribe(listener);
	const initial = listener.mock.calls[0][0] as BrowserViewportState;
	focusInput();
	resize(391, 320);
	const [current, reason] = listener.mock.calls[listener.mock.calls.length - 1];
	expect(reason).toBe('resize');
	expect(current).toMatchObject({
		keyboardOpen: true,
		availableHeight: 320,
		layoutHeight: 320,
		screenDirection: 'portrait',
		viewportDirection: 'landscape'
	});
	expect(current).not.toBe(initial);
	expect(initial).toMatchObject({ availableHeight: 850, keyboardOpen: false });
});

it('publishes physical direction changes without rebasing a keyboard on stale dimensions', () => {
	const listener = jest.fn();
	subscribe(listener);
	focusInput();
	orientation.type = 'landscape-primary';
	orientation.dispatchEvent(new Event('change'));
	expect(listener).toHaveBeenLastCalledWith(
		expect.objectContaining({
			screenDirection: 'landscape',
			layoutWidth: 390,
			layoutHeight: 850,
			keyboardOpen: false
		}),
		'orientation'
	);
	resize(850, 390);
	expect(getBrowserViewportState().keyboardOpen).toBe(false);
});

it('reports late focus without another geometry event so layout policies can recover height', () => {
	geometry(390, 850, 320);
	const listener = jest.fn();
	subscribe(listener);
	expect(getBrowserViewportState().keyboardOpen).toBe(false);
	focusInput();
	expect(listener).toHaveBeenLastCalledWith(
		expect.objectContaining({
			layoutHeight: 850,
			availableHeight: 320,
			keyboardOpen: true
		}),
		'focus'
	);
});

it('forgets the last session when no observers remain and starts fresh on remount', () => {
	const oldListener = jest.fn();
	const stop = subscribe(oldListener);
	focusInput();
	resize(390, 320);
	expect(getBrowserViewportState().keyboardOpen).toBe(true);
	stop();
	oldListener.mockClear();
	(document.activeElement as HTMLInputElement).blur();
	resize(390, 850);
	expect(oldListener).not.toHaveBeenCalled();
	const fresh = jest.fn();
	subscribe(fresh);
	expect(fresh).toHaveBeenLastCalledWith(
		expect.objectContaining({ keyboardOpen: false, availableHeight: 850 }),
		'initial'
	);
});

it('does not deliver the current event twice to a subscriber added during notification', () => {
	const added = jest.fn();
	subscribe((_, reason) => {
		if (reason === 'resize' && added.mock.calls.length === 0) subscribe(added);
	});
	resize(390, 800);
	expect(added).toHaveBeenCalledTimes(1);
	expect(added.mock.calls[0][1]).toBe('initial');
	resize(390, 750);
	expect(added).toHaveBeenCalledTimes(2);
});

it('does not notify a sibling disposed before its turn in the current event', () => {
	let stopSibling = () => {};
	subscribe((_, reason) => {
		if (reason === 'resize') stopSibling();
	});
	const sibling = jest.fn();
	stopSibling = subscribe(sibling);
	sibling.mockClear();
	resize(390, 800);
	expect(sibling).not.toHaveBeenCalled();
});

it('keeps subscriber snapshots ordered when another callback synchronously resizes again', () => {
	subscribe((state, reason) => {
		if (reason === 'resize' && state.layoutHeight === 800) {
			resize(390, 750);
		}
	});
	const other = jest.fn();
	subscribe(other);
	other.mockClear();
	resize(390, 800);
	expect(other.mock.calls.map(([state]) => state.layoutHeight)).toEqual([
		800, 750
	]);
	expect(other.mock.calls[other.mock.calls.length - 1][0]).toEqual(
		getBrowserViewportState()
	);
});

it('prevents a consumer from corrupting the snapshot delivered to other subscribers', () => {
	subscribe((state, reason) => {
		if (reason === 'resize') Reflect.set(state, 'layoutHeight', 1);
	});
	const other = jest.fn();
	subscribe(other);
	resize(390, 800);
	expect(other).toHaveBeenLastCalledWith(
		expect.objectContaining({ layoutHeight: 800, availableHeight: 800 }),
		'resize'
	);
	expect(getBrowserViewportState().layoutHeight).toBe(800);
});

it('discards queued events from a disposed observer when a callback remounts the consumer', () => {
	let stopFirst = () => {};
	const replacement = jest.fn();
	stopFirst = subscribe((state, reason) => {
		if (reason !== 'resize' || state.layoutHeight !== 800) return;
		resize(390, 750);
		stopFirst();
		subscribe(replacement);
	});
	resize(390, 800);
	expect(replacement).toHaveBeenCalledTimes(1);
	expect(replacement).toHaveBeenLastCalledWith(
		expect.objectContaining({ layoutHeight: 750 }),
		'initial'
	);
	resize(390, 700);
	expect(replacement).toHaveBeenCalledTimes(2);
	expect(replacement).toHaveBeenLastCalledWith(
		expect.objectContaining({ layoutHeight: 700 }),
		'resize'
	);
});

it('isolates callback errors so another subscriber still receives the update', () => {
	jest.spyOn(console, 'error').mockImplementation(() => {});
	const broken: BrowserViewportListener = () => {
		throw new Error('consumer failed');
	};
	expect(() => subscribe(broken)).not.toThrow();
	const other = jest.fn();
	subscribe(other);
	resize(390, 800);
	expect(other).toHaveBeenLastCalledWith(
		expect.objectContaining({ availableHeight: 800 }),
		'resize'
	);
});

it('uses legacy iOS direction events while retaining an already open keyboard', () => {
	set(window.screen, 'orientation', undefined);
	set(window, 'orientation', 0);
	const listener = jest.fn();
	subscribe(listener);
	focusInput();
	resize(390, 320);
	set(window, 'orientation', -90);
	window.dispatchEvent(new Event('orientationchange'));
	expect(listener).toHaveBeenLastCalledWith(
		expect.objectContaining({
			screenDirection: 'landscape',
			keyboardOpen: true
		}),
		'orientation'
	);
	resize(850, 210);
	resize(850, 390);
	expect(getBrowserViewportState().keyboardOpen).toBe(false);
});

it.each([
	['portrait', 0, 'landscape-primary'],
	['portrait', 180, 'landscape-secondary'],
	['landscape', 90, 'portrait-primary'],
	['landscape', -90, 'portrait-secondary']
])(
	'uses the current iOS %s angle when the modern direction is stale',
	(direction, angle, staleType) => {
		set(window, 'orientation', angle);
		orientation.type = staleType;
		expect(readScreenDirection()).toBe(direction);
	}
);

it('observes iOS rotation immediately when modern orientation changes arrive late', () => {
	set(window, 'orientation', 0);
	subscribe(jest.fn());
	set(window, 'orientation', 90);
	window.dispatchEvent(new Event('orientationchange'));
	expect(getBrowserViewportState().screenDirection).toBe('landscape');
	resize(850, 390);
	orientation.type = 'landscape-primary';
	orientation.dispatchEvent(new Event('change'));
	set(window, 'orientation', 0);
	window.dispatchEvent(new Event('orientationchange'));
	expect(getBrowserViewportState().screenDirection).toBe('portrait');
	resize(390, 850);
	focusInput();
	resize(390, 320);
	expect(getBrowserViewportState()).toMatchObject({
		screenDirection: 'portrait',
		keyboardOpen: true
	});
});

it('keeps modern orientation authoritative on Android', () => {
	set(navigator, 'userAgent', 'Mozilla/5.0 (Linux; Android 14) Mobile');
	set(window, 'orientation', 0);
	orientation.type = 'landscape-primary';
	expect(readScreenDirection()).toBe('landscape');
});
