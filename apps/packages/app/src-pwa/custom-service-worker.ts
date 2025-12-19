/*
 * This file (which will be your service worker)
 * is picked up by the build system ONLY if
 * quasar.config.js > pwa > workboxPluginMode is set to "InjectManifest"
 */

declare const self: ServiceWorkerGlobalScope &
	typeof globalThis & { skipWaiting: () => void };

import { precacheAndRoute } from 'workbox-precaching';
// import { clientsClaim } from 'workbox-core';
import { registerRoute } from 'workbox-routing';
import { NetworkOnly } from 'workbox-strategies';

precacheAndRoute(
	self.__WB_MANIFEST.filter((entry) => {
		if (typeof entry == 'string') {
			return true;
		}
		return !entry.url.endsWith('.html');
	})
);

self.addEventListener('install', () => {
	// console.log('Service Worker installing.');
});

self.addEventListener('activate', () => {
	// console.log('Service Worker activating.');
});

self.addEventListener('fetch', (event) => {
	// console.log('Fetching:', event);
});

registerRoute(
	({ request }) => {
		return request.mode === 'navigate';
	},
	new NetworkOnly({
		plugins: []
	})
);
