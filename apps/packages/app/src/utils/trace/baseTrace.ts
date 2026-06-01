import {
	TraceClient,
	TraceConfig,
	TracePublicAPI,
	TraceScopeResolver
} from './types';

function resolveScope(scope: TraceScopeResolver): string {
	return typeof scope === 'function' ? scope() : scope;
}

export function createTraceConfig(namespace: string): TraceConfig {
	return {
		key: `${namespace}:trace`,
		prefix: namespace
	};
}

export function createTraceClient(config: TraceConfig): TraceClient {
	let enabled = localStorage.getItem(config.key) === '1';

	const syncEnabledFromStorage = () => {
		enabled = localStorage.getItem(config.key) === '1';
	};
	window.addEventListener('storage', (event: StorageEvent) => {
		if (event.key === config.key) {
			syncEnabledFromStorage();
		}
	});

	const publicAPI: TracePublicAPI = {
		enable(): string {
			enabled = true;
			localStorage.setItem(config.key, '1');
			return 'Trace enabled.';
		},
		disable(): string {
			enabled = false;
			localStorage.removeItem(config.key);
			return 'Trace disabled.';
		},
		isEnabled(): boolean {
			return enabled;
		}
	};

	return {
		...publicAPI,
		trace(tag: string, payload: unknown, scope?: string): void {
			if (!enabled) {
				return;
			}
			const logLabel = scope
				? `[${config.prefix}] ${tag} (${scope})`
				: `[${config.prefix}] ${tag}`;
			console.info(logLabel, payload);
		},
		scoped(scope: TraceScopeResolver) {
			return (tag: string, payload: unknown) => {
				const currentScope = resolveScope(scope);
				if (!enabled) {
					return;
				}
				const logLabel = currentScope
					? `[${config.prefix}] ${tag} (${currentScope})`
					: `[${config.prefix}] ${tag}`;
				console.info(logLabel, payload);
			};
		},
		publicAPI
	};
}

export function registerTrace(
	namespace: string,
	publicAPI: TracePublicAPI
): void {
	window.__trace ??= {};
	window.__trace[namespace] = publicAPI;
}
