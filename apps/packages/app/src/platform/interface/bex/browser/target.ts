const mockBrowser = {
	runtime: {
		sendMessage: async (...args: any[]) => {
			console.warn('Mock browser.runtime.sendMessage called:', args);
			return undefined;
		},
		connect: (...args: any[]) => {
			console.warn('Mock browser.runtime.connect called:', args);
			return {
				postMessage: () => {},
				onMessage: {
					addListener: () => {},
					removeListener: () => {}
				},
				disconnect: () => {}
			};
		},
		onMessage: {
			addListener: () => {},
			removeListener: () => {}
		},
		getManifest: () => ({
			version: '1.0.0',
			name: 'Mock Extension'
		})
	},
	storage: {
		local: {
			get: async (keys?: string | string[] | object) => ({}),
			set: async (items: object) => {},
			remove: async (keys: string | string[]) => {},
			clear: async () => {}
		},
		sync: {
			get: async (keys?: string | string[] | object) => ({}),
			set: async (items: object) => {},
			remove: async (keys: string | string[]) => {},
			clear: async () => {}
		}
	},
	tabs: {
		query: async (queryInfo: object) => [],
		create: async (createProperties: object) => ({ id: -1 }),
		sendMessage: async (tabId: number, message: any) => {},
		update: async (tabId: number, updateProperties: object) => {},
		remove: async (tabIds: number | number[]) => {},
		onUpdated: {
			addListener: () => {},
			removeListener: () => {}
		}
	},
	windows: {
		create: async (createData: object) => ({ id: -1 }),
		update: async (windowId: number, updateInfo: object) => {},
		getAll: async () => [],
		getCurrent: async () => ({ id: -1 })
	},
	contextMenus: {
		create: (createProperties: object) => {},
		update: (id: string, updateProperties: object) => {},
		remove: (menuItemId: string) => {},
		removeAll: () => {}
	},
	notifications: {
		create: async (notificationId: string, options: object) => {},
		clear: async (notificationId: string) => {},
		onClicked: {
			addListener: () => {},
			removeListener: () => {}
		}
	},
	permissions: {
		request: async (permissions: object) => false,
		contains: async (permissions: object) => false,
		remove: async (permissions: object) => false
	},
	webRequest: {
		onBeforeRequest: {
			addListener: () => {},
			removeListener: () => {}
		},
		onCompleted: {
			addListener: () => {},
			removeListener: () => {}
		}
	},
	cookies: {
		get: async (details: object) => null,
		getAll: async (details: object) => [],
		set: async (details: object) => {},
		remove: async (details: object) => {}
	},
	i18n: {
		getMessage: (messageName: string, substitutions?: string | string[]) => ''
	}
};

function webextension() {
	if (process.env.DEV_PLATFORM_BEX) {
		return mockBrowser;
	} else {
		try {
			return require('webextension-polyfill-ts');
		} catch (err) {
			console.warn(
				'webextension-polyfill-ts not available, using mock browser API'
			);
		}
	}
}

const webextensionTarget = webextension();
export const browser = webextensionTarget?.browser;

export function isExtensionEnvironment(): boolean {
	return (
		typeof chrome !== 'undefined' && !!chrome.runtime && !!chrome.runtime.id
	);
}
