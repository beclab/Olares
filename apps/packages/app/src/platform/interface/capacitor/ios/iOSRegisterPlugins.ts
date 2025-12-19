import { registerPlugin } from '@capacitor/core';

export interface StorageiOSPlugin {
	set(options: { key: string; value: any }): Promise<void>;
	get(options: { key: string }): Promise<{
		value: string;
	}>;
	delete(options: { key: string }): Promise<void>;
	clear(): Promise<void>;
}

export interface ExtensionWebviewPlugin {
	open(options: { url: string }): Promise<void>;
}

export interface AutofilliOSPlugin {
	getAutofillList(): Promise<{ list: string[] }>;
	clearAutofillList(): Promise<void>;
	replaceAllIdentities(options: {
		identities: { id: string; url: string; userName: string }[];
	}): Promise<void>;
}

export interface AppSettingsPlugin {
	showAllowCrossSiteTracking(): Promise<{
		value: boolean;
	}>;
	jumpToAppSettings(): Promise<void>;
}

export interface IOSUniversalPlugin {
	getDeviceId(): Promise<{
		deviceId: string;
	}>;
	getWifiSSID(): Promise<{
		ssid: string;
	}>;

	finished(): Promise<void>;

	showHomeIndicator(): void;

	hideHomeIndicator(): void;

	getDiskSpace(): Promise<{
		freeSpace: number;
		totalSpace: number;
	}>;
}

const iOSStoragePlugin = registerPlugin<StorageiOSPlugin>('StorageiOS');

const iOSStorageUserPlugin = registerPlugin<StorageiOSPlugin>('StorageUseriOS');

const iOSExtensionWebviewPlugin =
	registerPlugin<ExtensionWebviewPlugin>('ExtensionWebview');

const iOSAutofillPlugin = registerPlugin<AutofilliOSPlugin>('AutofilliOS');

const iOSAppSettingsPlugin =
	registerPlugin<AppSettingsPlugin>('AppSettingsPlugin');

const IOSUniversal = registerPlugin<IOSUniversalPlugin>('IOSUniversalPlugin');

export default {
	iOSStoragePlugin,
	iOSStorageUserPlugin,
	iOSExtensionWebviewPlugin,
	iOSAutofillPlugin,
	iOSAppSettingsPlugin,
	IOSUniversal
};
