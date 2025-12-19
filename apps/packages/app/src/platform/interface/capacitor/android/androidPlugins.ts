import { PluginListenerHandle, registerPlugin } from '@capacitor/core';

export declare type onAutofillListener = ({ uri }) => void;

export declare type onAutofillSaveListener = ({ uri, dataType, data }) => void;

export interface AccessibilityPlugin {
	addListener(
		eventName: 'onAutofillAccessibility',
		listenerFunc: onAutofillListener
	): Promise<PluginListenerHandle> & PluginListenerHandle;

	openAccessibilitySettingPage(): void;

	openOverlaySettingPage(): void;

	isAccessibilityEnable(): Promise<{ isEnable: boolean }>;

	isOverlayEnable(): Promise<{ isEnable: boolean }>;

	setResult(options: { uri: string; userName: string; password: string }): void;
}

export interface AutofillFrameworkPlugin {
	addListener(
		eventName: 'onAutofillFramework',
		listenerFunc: onAutofillListener
	): Promise<PluginListenerHandle> & PluginListenerHandle;

	addListener(
		eventName: 'onAutofillFrameworkSave',
		listenerFunc: onAutofillSaveListener
	): Promise<PluginListenerHandle> & PluginListenerHandle;

	openSettingPage(): void;

	isSupport(): Promise<{ isSupport: boolean }>;

	isEnable(): Promise<{ isEnable: boolean }>;

	closeService(): void;

	setResult(options: { uri: string; userName: string; password: string }): void;
}

export interface EdgeToEdgePlugin {
	enable(): Promise<void>;

	disable(): Promise<void>;

	getInsets(): Promise<{
		bottom: number;
		left: number;
		right: number;
		top: number;
	}>;

	setBackgroundColor(options: { color: string }): Promise<void>;
}

export interface AndroidUniversalPlugin {
	finish(): void;

	shareLogDir(): void;

	isDebugApp(): Promise<{ app: boolean }>;

	moveTaskToBack(): void;

	getAppIconByPackageName(options: {
		packageName: string;
	}): Promise<{ base64: string }>;

	getDeviceId(): Promise<{
		deviceId: string;
		androidId: string;
		deviceUUID: string;
	}>;
	getWifiSSID(): Promise<{ ssid: string }>;

	showNavigationBar(): void;
	hideNavigationBar(): void;
	setAcceptedPrivacyTerms(options: { accepted: boolean }): void;

	getDiskSpace(): Promise<{
		freeSpace: number;
		totalSpace: number;
	}>;

	openLocationSettings(): void;
	openBluetoothSettings(): void;
	openAppSettings(): void;
}

const AndroidUniversal = registerPlugin<AndroidUniversalPlugin>(
	'AndroidUniversalPlugin'
);
const Accessibility = registerPlugin<AccessibilityPlugin>(
	'AccessibilityPlugin'
);
const AutofillFramework = registerPlugin<AutofillFrameworkPlugin>(
	'AutofillFrameworkPlugin'
);

const EdgeToEdge = registerPlugin<EdgeToEdgePlugin>('EdgeToEdgePlugin');

export default {
	Accessibility,
	AutofillFramework,
	AndroidUniversal,
	EdgeToEdge
};
