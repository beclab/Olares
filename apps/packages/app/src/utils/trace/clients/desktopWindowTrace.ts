import {
	createTraceClient,
	createTraceConfig,
	registerTrace
} from 'src/utils/trace';
import { TRACE_CATALOG } from 'src/utils/trace/catalog';

type ValueOf<T> = T[keyof T];

const DESKTOP_WINDOW_TRACE = TRACE_CATALOG.desktopWindow;

export const DESKTOP_WINDOW_TRACE_SCOPE = DESKTOP_WINDOW_TRACE.scopes;
export const DESKTOP_WINDOW_TRACE_TAG = DESKTOP_WINDOW_TRACE.tags;

type DesktopWindowTraceTag = ValueOf<typeof DESKTOP_WINDOW_TRACE_TAG>;
type DesktopWindowTraceScope =
	| typeof DESKTOP_WINDOW_TRACE_SCOPE.INDEX
	| ReturnType<typeof DESKTOP_WINDOW_TRACE_SCOPE.window>;

const desktopWindowTraceClient = createTraceClient(
	createTraceConfig(DESKTOP_WINDOW_TRACE.namespace)
);
registerTrace(
	DESKTOP_WINDOW_TRACE.namespace,
	desktopWindowTraceClient.publicAPI
);

export const isDesktopWindowTraceEnabled = () =>
	desktopWindowTraceClient.isEnabled();

export const traceDesktopWindow = (
	tag: DesktopWindowTraceTag,
	payload: unknown,
	scope?: DesktopWindowTraceScope
) => desktopWindowTraceClient.trace(tag, payload, scope);

export const createDesktopWindowTracer = (
	scope: DesktopWindowTraceScope | (() => DesktopWindowTraceScope)
) => desktopWindowTraceClient.scoped(scope);
