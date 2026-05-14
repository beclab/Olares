import {
	createTraceClient,
	createTraceConfig,
	registerTrace
} from 'src/utils/trace';
import { TRACE_CATALOG } from 'src/utils/trace/catalog';

type ValueOf<T> = T[keyof T];

const WEBSOCKET_TRACE = TRACE_CATALOG.websocket;

export const WEBSOCKET_TRACE_TAG = WEBSOCKET_TRACE.tags;
type WebsocketTraceTag = ValueOf<typeof WEBSOCKET_TRACE_TAG>;

const websocketTraceClient = createTraceClient(
	createTraceConfig(WEBSOCKET_TRACE.namespace)
);
registerTrace(WEBSOCKET_TRACE.namespace, websocketTraceClient.publicAPI);

export const traceWebsocket = (tag: WebsocketTraceTag, payload: unknown) => {
	websocketTraceClient.trace(tag, payload);
};
