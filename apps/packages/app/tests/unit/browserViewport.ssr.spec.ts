/** @jest-environment node */
import {
	getBrowserViewportState,
	getMobileScreenDirection,
	isMobileBrowser,
	observeBrowserViewport,
	readScreenDirection,
	readViewportHeight
} from '../../../core/src/browserViewport';
import { getInitialDeviceScreenState } from '../../../core/src/resize';

it('can be imported and read without browser globals or framework dependencies', () => {
	expect(typeof window).toBe('undefined');
	expect(typeof document).toBe('undefined');
	expect(readScreenDirection()).toBeUndefined();
	expect(getMobileScreenDirection()).toBeUndefined();
	expect(isMobileBrowser()).toBe(false);
	expect(readViewportHeight()).toBe(0);
	expect(getInitialDeviceScreenState()).toEqual({
		device: 'desktop',
		isVerticalScreen: false,
		isCompactLayout: false
	});
	expect(getBrowserViewportState()).toEqual({
		isMobileBrowser: false,
		screenDirection: undefined,
		viewportDirection: undefined,
		layoutWidth: 0,
		layoutHeight: 0,
		availableHeight: 0,
		keyboardOpen: false
	});
});

it('returns an independent synchronous empty snapshot and a safe disposer for each SSR subscriber', () => {
	const first = jest.fn();
	const second = jest.fn();
	const stopFirst = observeBrowserViewport(first);
	const stopSecond = observeBrowserViewport(second);
	expect(first).toHaveBeenCalledTimes(1);
	expect(second).toHaveBeenCalledTimes(1);
	expect(first.mock.calls[0][1]).toBe('initial');
	expect(second.mock.calls[0][1]).toBe('initial');
	expect(first.mock.calls[0][0]).toEqual(getBrowserViewportState());
	expect(() => {
		stopFirst();
		stopFirst();
		stopSecond();
	}).not.toThrow();
});
