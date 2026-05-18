import { boot } from 'quasar/wrappers';
import {
	DeviceType,
	getInitialDeviceScreenState,
	onDeviceChange
} from '@bytetrade/core';

function syncAdaptiveBodyClass(device: DeviceType) {
	document.body.classList.toggle('adaptive-pc', device !== DeviceType.MOBILE);
	document.body.classList.toggle(
		'adaptive-is-mobile',
		device === DeviceType.MOBILE
	);
}

export default boot(() => {
	if (typeof document === 'undefined') return;
	syncAdaptiveBodyClass(getInitialDeviceScreenState().device);
	onDeviceChange((state) => {
		syncAdaptiveBodyClass(state.device);
	});
});
