export interface TraceConfig {
	key: string;
	prefix: string;
}

export type TraceScopeResolver = string | (() => string);

export interface TracePublicAPI {
	enable(): string;
	disable(): string;
	isEnabled(): boolean;
}

export interface TraceClient {
	enable(): string;
	disable(): string;
	isEnabled(): boolean;
	trace(tag: string, payload: unknown, scope?: string): void;
	scoped(scope: TraceScopeResolver): (tag: string, payload: unknown) => void;
	publicAPI: TracePublicAPI;
}

declare global {
	interface Window {
		__trace?: Record<string, TracePublicAPI>;
	}
}
