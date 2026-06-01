interface MessageData {
	message?: string;
	type: string;
}

export async function postMessage(params: MessageData, origin = '*') {
	window.parent.postMessage(params, origin);
}
export class IframePostMessage {
	oldHref: string;
	observer: any;

	childPostMessage(params: MessageData, origin = '*') {
		this.oldHref = document.location.href;

		const notifyIfChanged = () => {
			if (this.oldHref !== document.location.href) {
				this.oldHref = document.location.href;
				params.message = document.location.href;
				postMessage(params, origin);
			}
		};

		// Intercept history.pushState / replaceState for SPA navigation (including keep-alive routes)
		const originalPushState = history.pushState.bind(history);
		const originalReplaceState = history.replaceState.bind(history);
		history.pushState = function (
			...args: Parameters<typeof history.pushState>
		) {
			originalPushState(...args);
			notifyIfChanged();
		};
		history.replaceState = function (
			...args: Parameters<typeof history.replaceState>
		) {
			originalReplaceState(...args);
			notifyIfChanged();
		};

		// Listen to browser back/forward navigation
		window.addEventListener('popstate', notifyIfChanged);

		// Keep MutationObserver as fallback for hash-based or non-standard routers
		const body = document.querySelector('body') as HTMLBodyElement;
		this.observer = new MutationObserver((mutations: any[]) => {
			mutations.forEach(() => {
				notifyIfChanged();
			});
		});
		this.observer.observe(body, { childList: true, subtree: true });
	}

	childDisconnect() {
		this.observer.disconnect();
	}

	parentEventListener(callBack) {
		window.addEventListener(
			'message',
			function (e: any) {
				callBack(e);
			},
			false
		);
	}
}
