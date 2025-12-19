import { PluginListenerHandle } from '@capacitor/core';
import { TerminusServiceInfo } from 'src/services/abstractions/mdns/service';

export declare type onServiceUpdated = (data: {
	data: TerminusServiceInfo[];
}) => void;
export interface DNSServicePlugin {
	start(): Promise<void>;
	stop(): Promise<void>;
	addListener(
		eventName: 'onServiceUpdated',
		listenerFunc: onServiceUpdated
	): Promise<PluginListenerHandle> & PluginListenerHandle;
}
