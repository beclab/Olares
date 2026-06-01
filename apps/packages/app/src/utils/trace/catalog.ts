export const TRACE_CATALOG = {
	desktopWindow: {
		namespace: 'desktop',
		tags: {
			APP_CLICK_START: 'app-click-start',
			APP_URL_RESOLVED: 'app-url-resolved',
			OPEN_EXTERNAL_MOBILE: 'open-external-mobile',
			OPEN_EXTERNAL_WINDOW: 'open-external-window',
			REUSE_WINDOW: 'reuse-window',
			CREATE_WINDOW: 'create-window',
			APP_NOT_FOUND: 'app-not-found',
			IFRAME_MOUNT: 'iframe-mount',
			IFRAME_LOAD_TIMEOUT: 'iframe-load-timeout',
			IFRAME_LOAD_SUCCESS: 'iframe-load-success',
			IFRAME_LOAD_ERROR: 'iframe-load-error'
		},
		scopes: {
			INDEX: 'desktop-index',
			window: (windowId: string) => `window:${windowId}` as const
		}
	},
	websocket: {
		namespace: 'websocket',
		tags: {
			WORKER_MESSAGE: 'worker-message'
		}
	}
} as const;
