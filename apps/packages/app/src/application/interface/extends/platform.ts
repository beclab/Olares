import { PushNotificationSchema } from '@capacitor/push-notifications';
import { AppPlatform } from '../platform';
import { BiometricKeyStore } from '@didvault/sdk/src/core';
import { AvailableResult } from '@capgo/capacitor-native-biometric';
import { ConnectionStatus } from '@capacitor/network';
import { NativeScanQRProtocol } from './protocols';
import { OrientationLockType } from '@capacitor/screen-orientation';
import { DriveType } from 'src/utils/interface/files';
import { BluetoothWifiInfo } from 'src/utils/interface/bluetooth';
import { QVueGlobals } from 'quasar';
import { Router } from 'vue-router';

export interface NativeAppBiometricKeyStore extends BiometricKeyStore {
	deleteKey(id: string): Promise<void>;
	isSupportedWithData(): Promise<AvailableResult>;
}

export interface NativeAppPlatform extends AppPlatform {
	getFCMToken(token: { value: string }): Promise<{ token: string }>;
	pushNotificationReceived(notification: PushNotificationSchema): Promise<void>;

	biometricKeyStore: NativeAppBiometricKeyStore;

	openBiometric(): Promise<{
		status: boolean;
		message: string;
	}>;

	closeBiometric(): Promise<{
		status: boolean;
		message: string;
	}>;

	unlockByBiometric(): Promise<string>;

	scanQRDidUserGrantPermission(): Promise<boolean>;

	scanQrCheckPermission(): Promise<void>;

	getQRCodeImageFromPhotoAlbum(): Promise<string>;

	hookBackAction(): void;

	scanQRProtocolList: NativeScanQRProtocol[];

	defaultOrientationLockType: OrientationLockType;

	resetOrientationLockType(): Promise<void>;

	isLandscape(): boolean;

	getDeviceId(): Promise<string>;

	getWifiSSID(): Promise<string>;

	getDiskSpace(): Promise<{
		freeSpace: number;
		totalSpace: number;
	}>;

	netConnectionStatus(): Promise<ConnectionStatus>;

	openLocationSettings(): void;
	openBluetoothSettings(): void;
	openAppSettings(): void;

	finished(): void;

	selectUploadFiles(
		driveType: DriveType,
		path: string,
		params: any,
		isImage?: boolean
	): void;

	addUploadTasks(
		files: {
			path: string;
			name: string;
			size: number;
			mimeType: string;
		}[],
		driveType: DriveType,
		path: string,
		params: any
	): Promise<void>;

	startSearchBluethooth(): Promise<boolean>;

	stopSearchBluetooth(): Promise<void>;

	bluetoothsGetWifiList(deviceId: string): Promise<BluetoothWifiInfo[]>;

	bluetoothConnectWifi(
		deviceId: string,
		ssid: string,
		password: string
	): Promise<{
		status: boolean;
		message: string;
	}>;

	shareLogDir(): Promise<void>;

	isDebugApp(): Promise<boolean>;

	getQuasar(): QVueGlobals | undefined;

	getRouter(): Router | undefined;

	
}
