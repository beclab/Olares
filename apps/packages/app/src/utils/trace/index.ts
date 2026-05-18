/**
 * Console example (desktop):
 * - window.__trace?.desktop?.enable() // "Trace enabled."
 * - window.__trace?.desktop?.disable() // "Trace disabled."
 * - window.__trace?.desktop?.isEnabled() // true | false
 */
export {
	createTraceClient,
	createTraceConfig,
	registerTrace
} from './baseTrace';
export { TRACE_CATALOG } from './catalog';
export type {
	TraceClient,
	TraceConfig,
	TracePublicAPI,
	TraceScopeResolver
} from './types';
