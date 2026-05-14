import { PluginListenerHandle } from '@capacitor/core';

export declare type onVpnStatusUpdate = (status: string) => void;

export interface TailScalePluginInterface {
	open(options: {
		authKey: string | undefined;
		server: string;
		acceptDns: boolean;
		authToken: string;
	}): void;
	close(): void;

	addListener(
		eventName: 'vpnStatusUpdate',
		listenerFunc: onVpnStatusUpdate
	): Promise<PluginListenerHandle> & PluginListenerHandle;

	status(): Promise<{
		status: string;
		options: any;
	}>;

	currentNodeId(): Promise<{
		nodeId: string;
	}>;

	peersState(): Promise<{
		isRunning: boolean;
		state: string;
	}>;

	resendCache(options: { server: string; authToken: string }): Promise<void>;
}
