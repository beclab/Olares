import { defineStore } from 'pinia';
import {
	DeviceType,
	getInitialDeviceScreenState,
	onDeviceChange,
	useDevice
} from '@bytetrade/core';
import { useBackgroundStore } from 'src/stores/settings/background';

export type DeviceStoreState = {
	deviceInfo: {
		device: DeviceType;
		isVerticalScreen: boolean;
	};
};

export const useDeviceStore = defineStore('deviceStore', {
	state: () => {
		return {
			deviceInfo: getInitialDeviceScreenState()
		} as DeviceStoreState;
	},

	getters: {
		isMobile(): boolean {
			return this.deviceInfo.device === DeviceType.MOBILE;
		},
		platform(): string {
			return this.isMobile ? 'mobile' : 'web';
		}
	},

	actions: {
		init(
			callback?: (state: {
				device: DeviceType;
				isVerticalScreen: boolean;
			}) => void
		) {
			const { state } = useDevice();
			this.deviceInfo = state;
			onDeviceChange(
				(state: { device: DeviceType; isVerticalScreen: boolean }) => {
					if (callback) {
						callback(state);
					}
					this.deviceInfo = state;
				}
			);
		}
	}
});
