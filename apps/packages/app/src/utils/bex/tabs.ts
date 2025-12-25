import { Platform } from 'quasar';
import { getApplication } from 'src/application/base';
import { browser } from 'src/platform/interface/bex/browser/target';

interface TabInfo {
	title: string;
	url: string;
	favIconUrl: string;
}

export const openUrl = (url: string, target: '_blank' | '_self' = '_blank') => {
	const fullUrl = url.startsWith('http') ? url : `https://${url}`;
	getApplication().openUrl(fullUrl, target);
};

const getWebTabInfo = (): TabInfo => ({
	title: document.title,
	url: window.location.href,
	favIconUrl: getFaviconUrl()
});

export const getCurrentTabInfo = (): Promise<TabInfo> => {
	if (!Platform.is.bex) {
		return Promise.resolve(getWebTabInfo());
	}

	return new Promise((resolve) => {
		chrome.tabs.query({ active: true, currentWindow: true }, (tabs) => {
			if (!tabs || !tabs[0]) {
				resolve(getWebTabInfo());
				return;
			}
			const tab = tabs[0];
			resolve({
				title: tab.title || document.title,
				url: tab.url || window.location.href,
				favIconUrl: tab.favIconUrl || getFaviconUrl()
			});
		});
	});
};

export const listenToTabChanges = (
	callback: (info: TabInfo) => void
): (() => void) => {
	if (Platform.is.bex) {
		const chromeListener = (
			tabId: number,
			changeInfo: chrome.tabs.TabChangeInfo,
			tab: chrome.tabs.Tab
		) => {
			if (changeInfo.status === 'complete' && tab.active) {
				callback({
					title: tab.title || document.title,
					url: tab.url || window.location.href,
					favIconUrl: tab.favIconUrl || getFaviconUrl()
				});
			}
		};

		const activatedListener = (activeInfo: chrome.tabs.TabActiveInfo) => {
			chrome.tabs.get(activeInfo.tabId, (tab) => {
				callback({
					title: tab.title || document.title,
					url: tab.url || window.location.href,
					favIconUrl: tab.favIconUrl || getFaviconUrl()
				});
			});
		};

		chrome.tabs.onUpdated.addListener(chromeListener);
		chrome.tabs.onActivated.addListener(activatedListener);

		return () => {
			chrome.tabs.onUpdated.removeListener(chromeListener);
			chrome.tabs.onActivated.removeListener(activatedListener);
		};
	} else {
		const titleObserver = new MutationObserver(() => {
			callback(getWebTabInfo());
		});

		titleObserver.observe(document.querySelector('title') || document.head, {
			subtree: true,
			characterData: true,
			childList: true
		});

		const handleUrlChange = () => callback(getWebTabInfo());
		window.addEventListener('popstate', handleUrlChange);
		window.addEventListener('hashchange', handleUrlChange);

		return () => {
			titleObserver.disconnect();
			window.removeEventListener('popstate', handleUrlChange);
			window.removeEventListener('hashchange', handleUrlChange);
		};
	}
};

function getFaviconUrl(): string {
	const links = document.getElementsByTagName('link');
	for (const link of links) {
		if (link.rel.includes('icon')) return link.href;
	}
	return `${window.location.origin}/favicon.ico`;
}

export const getTabUrl = (debugUrl?: string): Promise<string> => {
	if (!Platform.is.bex) {
		return Promise.resolve(debugUrl || window.location.href);
	}

	return new Promise((resolve) => {
		chrome.tabs.query({ active: true, currentWindow: true }, (tabs) => {
			if (!tabs || !tabs[0]) {
				resolve(window.location.href);
				return;
			}
			resolve(tabs[0].url || window.location.href);
		});
	});
};

export const isExtensionUrl = (url: string) => {
	return (
		url.startsWith('chrome-extension://') || url.startsWith('moz-extension://')
	);
};

export const createTabChangeListenerInCurrentWindow = (handleActivated) => {
	let currentWindowId;
	let activatedListener;
	let updatedListener;

	const initPromise = browser.windows.getCurrent().then((window) => {
		currentWindowId = window.id;

		activatedListener = (activeInfo) => {
			if (activeInfo.windowId === currentWindowId) {
				handleActivated(activeInfo);
			}
		};

		updatedListener = (tabId, changeInfo, tab) => {
			if (tab.windowId === currentWindowId && changeInfo.url) {
				handleActivated({
					tabId,
					windowId: tab.windowId,
					changeInfo
				});
			}
		};

		browser.tabs.onActivated.addListener(activatedListener);
		browser.tabs.onUpdated.addListener(updatedListener);
	});

	return {
		remove: () =>
			initPromise.then(() => {
				if (activatedListener) {
					browser.tabs.onActivated.removeListener(activatedListener);
				}
				if (updatedListener) {
					browser.tabs.onUpdated.removeListener(updatedListener);
				}
				activatedListener = null;
				updatedListener = null;
			}),

		initPromise
	};
};
